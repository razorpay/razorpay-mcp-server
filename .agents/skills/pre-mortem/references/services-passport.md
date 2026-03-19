# Services: Passport Integration (10 Checks)

Validates Passport authentication/authorization integration patterns. Passport is Razorpay's Edge service for JWT-based auth that validates and extracts user identity. Focuses on correct SDK usage, JWT validation, and secure context handling.

**Total Checks**: 10 (5 Critical, 4 High, 1 Medium)

---

## Check #1: JWKS Host Configuration (Critical)

**Problem**: Incorrect or missing JWKS host causes passport initialization failure, blocking all authenticated requests.

**Detection Strategy**:
```bash
# Find passport initialization
grep -E "passport\.InitHandler|InitHandler" <pr_files>

# Check config files for passport host
grep -E "\[passport\]|\[Passport\]" configs/*.toml

# Verify host configuration
grep -i "edge.*host\|jwks.*host\|passport.*host" configs/*.toml
```

**What to Flag**:
```go
// ❌ BAD: Hardcoded JWKS host (wrong for prod)
ph, err := passport.InitHandler("https://edge-base.dev.razorpay.in")

// ❌ BAD: Wrong environment host in prod code
// In production deployment:
config := Config{
    PassportHost: "https://edge-base.dev.razorpay.in",  // <-- Dev host in prod!
}

// ❌ BAD: No validation of host config
host := os.Getenv("PASSPORT_HOST")  // <-- Could be empty!
ph, _ := passport.InitHandler(host)

// ❌ BAD: Missing passport config in TOML
// configs/prod.toml - no [passport] section!
```

**How to Fix**:
```go
// ✅ GOOD: Load from config with validation
func initPassport(cfg *config.Config) (passport.IPassportHandler, error) {
    if cfg.Passport.Host == "" {
        return nil, errors.New("passport host not configured")
    }

    ph, err := passport.InitHandler(cfg.Passport.Host)
    if err != nil {
        return nil, fmt.Errorf("failed to init passport: %w", err)
    }

    return ph, nil
}

// ✅ GOOD: Config in TOML with environment-specific hosts
// configs/default.toml (dev/stage):
[passport]
    host = "https://edge-base.dev.razorpay.in"

// configs/stage.toml:
[passport]
    host = "https://edge-admin.int.stage.razorpay.in"

// configs/automation-live.toml:
[passport]
    host = "https://edge-admin.int.qa.razorpay.in"

// configs/prod-live.toml:
[passport]
    host = "https://edge-admin-internal.razorpay.com"
```

**Razorpay Standard - JWKS Hosts**:
- **Devstack**: `https://edge-base.dev.razorpay.in`
- **Stage**: `https://edge-admin.int.stage.razorpay.in`
- **Automation**: `https://edge-admin.int.qa.razorpay.in`
- **Production**: `https://edge-admin-internal.razorpay.com`

**Configuration Options**:
```go
// ✅ GOOD: Override HTTP config for faster boot time
ph, err := passport.InitHandler(
    jwksHost,
    passport.WithJwksFetchConnectionTimeout(1*time.Second),  // default: 2s
    passport.WithJwksFetchMaxRetryAttempts(3),               // default: 2
    passport.WithJwksFetchDelayBetweenRetries(100*time.Millisecond), // default: 200ms
)
```

**Rationale**: Wrong JWKS host causes all passport validation to fail, blocking authenticated API calls.

---

## Check #2: JWT Token Extraction (Critical)

**Problem**: Missing or incorrect header key extraction causes auth failures.

**Detection Strategy**:
```bash
# Find JWT token extraction in middleware
grep -E "Header\.Get.*Passport|GetHeader.*Passport" <pr_files>

# Check if correct constant is used
grep -E "passport\.HeaderKeyPassportJWTV1|X-Passport-JWT-V1" <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Hardcoded header name (typo risk)
jwtToken := r.Header.Get("X-Passport-JWT-V1")  // <-- Should use constant!

// ❌ BAD: Wrong header name
jwtToken := r.Header.Get("X-Passport-Token")  // <-- Wrong header!

// ❌ BAD: No empty token check
jwtToken := c.Request.Header.Get(passport.HeaderKeyPassportJWTV1)
p, _ := ph.FromToken(jwtToken)  // <-- Will fail if jwtToken is empty!

// ❌ BAD: Using Authorization header instead
jwtToken := r.Header.Get("Authorization")  // <-- Wrong! Passport uses specific header
```

**How to Fix**:
```go
// ✅ GOOD: Use SDK constant for header key
import "github.com/razorpay/goutils/passport/v4"

func ExtractPassportToken(c *gin.Context) {
    jwtToken := c.Request.Header.Get(passport.HeaderKeyPassportJWTV1)

    if jwtToken == "" {
        logger.Error(c, "PASSPORT_TOKEN_MISSING")
        c.AbortWithStatus(http.StatusUnauthorized)
        return
    }

    p, err := passportHandler.FromToken(jwtToken)
    if err != nil {
        logger.Error(c, "PASSPORT_VALIDATION_FAILED", "error", err)
        c.AbortWithStatus(http.StatusUnauthorized)
        return
    }

    // Add to context
    c.Set("passport", p)
}

// ✅ GOOD: For gorilla/mux
func PassportMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        jwtToken := r.Header.Get(passport.HeaderKeyPassportJWTV1)

        if jwtToken == "" {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }

        pass, err := passportHandler.FromToken(jwtToken)
        if err != nil {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }

        // Add to context
        ctx := context.WithValue(r.Context(), "passport", pass)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Razorpay Standard**:
- **Header constant**: Always use `passport.HeaderKeyPassportJWTV1`
- **Header name**: `"X-Passport-JWT-V1"`
- **Empty check**: Always validate token is non-empty before calling `FromToken`
- **Error response**: Return `401 Unauthorized` on validation failure

**Rationale**: Incorrect header extraction causes all authenticated requests to fail.

---

## Check #3: Passport Validation & Error Handling (Critical)

**Problem**: Ignored validation errors allow unauthenticated requests through.

**Detection Strategy**:
```bash
# Find FromToken calls
grep -E "FromToken|passportHandler\.FromToken" <pr_files>

# Check if error is handled
grep -A 5 "FromToken" <pr_files> | grep "if err"

# Check for nil passport checks
grep -A 5 "FromToken" <pr_files> | grep "if.*== nil\|if.*!= nil"
```

**What to Flag**:
```go
// ❌ BAD: Ignoring validation error
p, _ := passportHandler.FromToken(jwtToken)  // <-- Ignores error!
c.Set("passport", p)  // p could be nil!

// ❌ BAD: No nil check after validation
p, err := passportHandler.FromToken(jwtToken)
if err != nil {
    logger.Error(c, "PASSPORT_ERROR", "error", err)
    // Continues execution without aborting!
}
c.Set("passport", p)  // <-- p could still be nil!

// ❌ BAD: Not checking IsIdentified
p, err := passportHandler.FromToken(jwtToken)
if err != nil {
    c.AbortWithStatus(401)
    return
}
// Missing: check if p.IsIdentified() == true
c.Set("passport", p)
```

**How to Fix**:
```go
// ✅ GOOD: Proper validation with error handling
p, err := passportHandler.FromToken(jwtToken)
if err != nil {
    logger.Error(c, "PASSPORT_VALIDATION_FAILED", "error", err)
    c.AbortWithStatus(http.StatusUnauthorized)
    return
}

if p == nil {
    logger.Error(c, "PASSPORT_IS_NIL")
    c.AbortWithStatus(http.StatusUnauthorized)
    return
}

if !p.IsIdentified() {
    logger.Error(c, "PASSPORT_NOT_IDENTIFIED")
    c.AbortWithStatus(http.StatusUnauthorized)
    return
}

// Now safe to use passport
c.Set("passport", p)

// ✅ GOOD: Combined validation
p, err := passportHandler.FromToken(jwtToken)
if err != nil || p == nil || !p.IsIdentified() {
    logger.Error(c, "PASSPORT_VALIDATION_FAILED", "error", err, "nil", p == nil)
    c.AbortWithStatus(http.StatusUnauthorized)
    return
}
```

**Razorpay Standard**:
- **Always check error** from `FromToken`
- **Always check nil passport** after validation
- **Always check `IsIdentified()`** for authenticated routes
- **Abort request** on any validation failure
- **Log failure** with context for debugging

**Rationale**: Skipping validation allows unauthenticated access to protected resources.

---

## Check #4: Context Storage & Retrieval (High)

**Problem**: Incorrect context storage or retrieval causes runtime panics.

**Detection Strategy**:
```bash
# Find context storage
grep -E "Set\(\"passport\"|WithValue.*passport|context\.WithValue" <pr_files>

# Find context retrieval
grep -E "Get\(\"passport\"|Value\(.*passport" <pr_files>

# Check for type assertions
grep -A 2 "passport.*\)\." <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: No type assertion check
func GetMerchantID(ctx context.Context) string {
    passport := ctx.Value("passport").(passport.IPassport)  // <-- Panic if not found!
    merchantID, _ := passport.GetResourceOwnerID(passport)
    return merchantID
}

// ❌ BAD: Wrong context key
c.Set("passport_token", p)  // <-- Stored with wrong key
// Later:
passport := c.Get("passport")  // <-- Different key, returns nil

// ❌ BAD: Not checking ok boolean
passport, _ := ctx.Value("passport").(passport.IPassport)  // <-- Ignores ok!
ownerID := passport.GetConsumerClaims().ID  // <-- Panic if passport is nil!
```

**How to Fix**:
```go
// ✅ GOOD: Safe type assertion with check (gin)
func GetMerchantID(c *gin.Context) (string, error) {
    passportVal, exists := c.Get("passport")
    if !exists {
        return "", errors.New("passport not found in context")
    }

    passport, ok := passportVal.(passport.IPassport)
    if !ok {
        return "", errors.New("invalid passport type in context")
    }

    merchantID, err := passport.GetResourceOwnerID(passport)
    if err != nil {
        return "", fmt.Errorf("failed to get merchant ID: %w", err)
    }

    return merchantID, nil
}

// ✅ GOOD: Safe type assertion with check (standard context)
func GetPassportFromContext(ctx context.Context) (passport.IPassport, error) {
    passportVal := ctx.Value("passport")
    if passportVal == nil {
        return nil, errors.New("passport not found in context")
    }

    passport, ok := passportVal.(passport.IPassport)
    if !ok {
        return nil, errors.New("invalid passport type in context")
    }

    return passport, nil
}

// Usage:
p, err := GetPassportFromContext(ctx)
if err != nil {
    return err
}
merchantID := p.GetConsumerClaims().ID

// ✅ BETTER: Define typed context key
type passportKeyType struct{}
var PassportKey = passportKeyType{}

// Store:
ctx = context.WithValue(ctx, PassportKey, p)

// Retrieve:
passportVal := ctx.Value(PassportKey)
passport, ok := passportVal.(passport.IPassport)
```

**Razorpay Standard**:
- **Always check type assertion** with `ok` boolean
- **Use consistent context keys** ("passport" is standard)
- **Consider typed keys** for compile-time safety
- **Create helper functions** for retrieval to avoid duplication

**Rationale**: Unchecked type assertions cause panics in request handlers.

---

## Check #5: Resource Owner Extraction (High)

**Problem**: Incorrect merchant/user ID extraction causes authorization bugs.

**Detection Strategy**:
```bash
# Find resource owner extraction
grep -E "GetResourceOwnerID|GetMerchantIdFromPassport|GetResourceOwnerType" <pr_files>

# Check for error handling
grep -A 3 "GetResourceOwnerID" <pr_files> | grep "if err"
```

**What to Flag**:
```go
// ❌ BAD: Ignoring error from GetResourceOwnerID
ownerID, _ := passport.GetResourceOwnerID(p)  // <-- Could return error!
processRequest(ownerID)  // <-- May have empty ownerID!

// ❌ BAD: Not checking owner type before assuming merchant
ownerID, _ := passport.GetResourceOwnerID(p)
// Assumes ownerID is merchant ID, but could be admin, customer, etc.
merchant := getMerchant(ownerID)  // <-- Wrong if ownerID is admin!

// ❌ BAD: Direct consumer ID access without helper
merchantID := p.GetConsumerClaims().ID
// This may not be correct in impersonation scenarios!
```

**How to Fix**:
```go
// ✅ GOOD: Use helper with error handling
ownerID, err := passport.GetResourceOwnerID(p)
if err != nil {
    return fmt.Errorf("failed to get resource owner ID: %w", err)
}

// ✅ GOOD: Check owner type before assuming merchant
ownerType, err := passport.GetResourceOwnerType(p)
if err != nil {
    return fmt.Errorf("failed to get owner type: %w", err)
}

if ownerType != "merchant" {
    return errors.New("only merchant access allowed")
}

ownerID, err := passport.GetResourceOwnerID(p)
if err != nil {
    return err
}

// Now safe to use as merchant ID
merchant := getMerchant(ownerID)

// ✅ GOOD: Dedicated helper for merchant ID
func GetMerchantID(p passport.IPassport) (string, error) {
    ownerType, err := passport.GetResourceOwnerType(p)
    if err != nil {
        return "", err
    }

    if ownerType != "merchant" {
        return "", errors.New("resource owner is not a merchant")
    }

    merchantID, err := passport.GetResourceOwnerID(p)
    if err != nil {
        return "", err
    }

    return merchantID, nil
}
```

**Razorpay Standard - Consumer Types**:
- `merchant` - Merchant account
- `admin` - Razorpay admin
- `customer` - End customer
- `application` - OAuth application

**Helper Functions Available**:
```go
// From passport SDK:
passport.GetResourceOwnerID(p)        // Handles impersonation correctly
passport.GetResourceOwnerType(p)      // Returns consumer type
passport.GetLegacyAuthType(p)         // Returns "direct", "public", "private", etc.

// From passport instance:
p.GetConsumerClaims().ID              // Raw consumer ID (may not be correct for impersonation)
p.GetConsumerClaims().Type            // Consumer type
p.IsAuthenticated()                   // Check if authenticated
p.IsIdentified()                      // Check if identified
p.GetMode()                           // "test" or "live"
p.GetDomain()                         // Domain information
p.GetOrg()                            // Organization ID
p.GetRoles()                          // User roles
```

**Rationale**: Incorrect owner extraction causes authorization bypass or wrong data access.

---

## Check #6: Test Environment Handling (Critical)

**Problem**: Test-only passport handling in production code or vice versa.

**Detection Strategy**:
```bash
# Find test passport handling
grep -E "test.*passport|passport.*test|EnvSlit" <pr_files>

# Check for hardcoded test tokens
grep -E "eyJ.*\.|\".*passport.*:.*authenticated" <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Hardcoded test passport in code (not env-gated)
testPassport := "{\"authenticated\":true,\"consumer\":{\"id\":\"M123\",\"type\":\"merchant\"}}"
c.Set("passport", testPassport)  // <-- Always sets test passport!

// ❌ BAD: Test check in wrong place (should be in middleware, not business logic)
func ProcessPayment(ctx context.Context) {
    if config.Env == "test" {
        // Skip passport check
        merchantID = "test_merchant"
    } else {
        passport := getPassport(ctx)
        merchantID = passport.GetResourceOwnerID()
    }
}

// ❌ BAD: Production code using test mode incorrectly
if bootstrap.Config.Application.Mode == "test" {
    // Skips passport validation
    return  // <-- No fallthrough to real validation!
}
```

**How to Fix**:
```go
// ✅ GOOD: Test passport only in test environments (gin)
func ExtractPassportToken() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Test mode: use hardcoded passport
        if (config.Application.Env == "test" && config.Application.Mode == "test") ||
            config.Application.Env == constants.EnvSlit {
            testPassport := "{\"authenticated\":true,\"identified\":true,\"consumer\":{\"id\":\"test_merchant\",\"type\":\"merchant\"}}"
            c.Set("passport", testPassport)
            return  // <-- Early return for test
        }

        // Production: real validation
        jwtToken := c.Request.Header.Get(passport.HeaderKeyPassportJWTV1)
        p, err := passportHandler.FromToken(jwtToken)
        if err != nil {
            logger.Error(c, "PASSPORT_FAILURE", "error", err)
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }

        c.Set("passport", p)
    }
}

// ✅ GOOD: Separate test handler
func GetPassportHandler(config *Config) gin.HandlerFunc {
    if config.IsTestEnv() {
        return TestPassportMiddleware()
    }
    return ProductionPassportMiddleware()
}
```

**Razorpay Standard - Test Environments**:
- **Test mode**: `config.Application.Env == "test" && config.Application.Mode == "test"`
- **SLIT environment**: `config.Application.Env == constants.EnvSlit`
- **Test passport**: Should have realistic structure matching production
- **Gating**: Test-specific code must be env-gated in middleware only

**Rationale**: Test bypass in production or missing test handling breaks CI/testing.

---

## Check #7: Mode & Domain Validation (High)

**Problem**: Not validating mode ("test" vs "live") allows test data in production.

**Detection Strategy**:
```bash
# Find mode checks
grep -E "GetMode\(\)|passport\.Mode|p\.mode" <pr_files>

# Find payment/transaction processing
grep -E "ProcessPayment|CreateOrder|RefundPayment" <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Not checking mode before processing payment
func ProcessPayment(ctx context.Context, amount int) error {
    passport := getPassportFromContext(ctx)
    merchantID := passport.GetConsumerClaims().ID

    // Missing: check if mode == "live" for real money!
    return processRealPayment(merchantID, amount)
}

// ❌ BAD: Allowing test mode in production
if passport.GetMode() == "test" {
    // Test mode processing in production environment!
    return processTestPayment()
}
```

**How to Fix**:
```go
// ✅ GOOD: Validate mode before processing
func ProcessPayment(ctx context.Context, amount int) error {
    passport := getPassportFromContext(ctx)

    mode := passport.GetMode()
    if mode != "live" {
        return errors.New("payments only allowed in live mode")
    }

    merchantID := passport.GetConsumerClaims().ID
    return processRealPayment(merchantID, amount)
}

// ✅ GOOD: Separate test and live handlers
func ProcessPayment(ctx context.Context, amount int) error {
    passport := getPassportFromContext(ctx)

    switch passport.GetMode() {
    case "live":
        return processLivePayment(ctx, amount)
    case "test":
        return processTestPayment(ctx, amount)
    default:
        return errors.New("unknown mode")
    }
}

// ✅ GOOD: Domain validation for multi-tenant
func GetMerchantData(ctx context.Context) (*Merchant, error) {
    passport := getPassportFromContext(ctx)

    domain := passport.GetDomain()
    if domain != "" && domain != config.AllowedDomain {
        return nil, errors.New("domain not allowed")
    }

    merchantID, _ := passport.GetResourceOwnerID(passport)
    return getMerchant(merchantID)
}
```

**Razorpay Standard**:
- **Mode values**: `"test"` or `"live"`
- **Real money operations**: Always validate `mode == "live"`
- **Test operations**: Explicitly allow `mode == "test"`
- **Domain**: Validate if using multi-tenant setup

**Rationale**: Mode mismatch allows test transactions in production or blocks real transactions.

---

## Check #8: Role-Based Access Control (High)

**Problem**: Not checking roles allows unauthorized admin actions.

**Detection Strategy**:
```bash
# Find admin-only operations
grep -E "DeleteMerchant|UpdateConfig|AdminAction" <pr_files>

# Check for role validation
grep -E "GetRoles\(\)|HasRole|CheckRole" <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Admin action without role check
func DeleteMerchant(ctx context.Context, merchantID string) error {
    passport := getPassportFromContext(ctx)
    // Missing: role check!
    return deleteMerchantFromDB(merchantID)
}

// ❌ BAD: Case-sensitive role check
roles := passport.GetRoles()
if roles[0] == "superadmin" {  // <-- Should be "SuperAdmin"
    allowAdmin()
}
```

**How to Fix**:
```go
// ✅ GOOD: Role validation before admin action
func DeleteMerchant(ctx context.Context, merchantID string) error {
    passport := getPassportFromContext(ctx)

    roles := passport.GetRoles()
    if !hasRole(roles, "SuperAdmin") {
        return errors.New("unauthorized: SuperAdmin role required")
    }

    logger.Info(ctx, "ADMIN_DELETE_MERCHANT",
        "admin_id", passport.GetConsumerClaims().ID,
        "merchant_id", merchantID,
    )

    return deleteMerchantFromDB(merchantID)
}

// Helper: case-insensitive role check
func hasRole(roles []string, requiredRole string) bool {
    for _, role := range roles {
        if strings.EqualFold(role, requiredRole) {
            return true
        }
    }
    return false
}

// ✅ GOOD: Multiple role support
func hasAnyRole(roles []string, allowedRoles []string) bool {
    roleMap := make(map[string]bool)
    for _, role := range roles {
        roleMap[strings.ToLower(role)] = true
    }

    for _, allowed := range allowedRoles {
        if roleMap[strings.ToLower(allowed)] {
            return true
        }
    }
    return false
}

func UpdateConfig(ctx context.Context) error {
    passport := getPassportFromContext(ctx)

    allowedRoles := []string{"SuperAdmin", "sop_qc_admin_role"}
    if !hasAnyRole(passport.GetRoles(), allowedRoles) {
        return errors.New("unauthorized")
    }

    return updateConfigInDB()
}
```

**Razorpay Standard - Common Roles**:
- `SuperAdmin` - Full admin access
- `sop_qc_admin_role` - SOP quality control
- `finance_admin` - Finance operations
- `support_admin` - Support operations

**Rationale**: Missing role checks allow unauthorized admin operations.

---

## Check #9: Passport Initialization Timeout (Medium)

**Problem**: Default timeouts cause slow pod boot times.

**Detection Strategy**:
```bash
# Find InitHandler calls
grep -E "passport\.InitHandler" <pr_files>

# Check for timeout configuration
grep -E "WithJwksFetchConnectionTimeout|WithJwksFetchMaxRetryAttempts" <pr_files>
```

**What to Flag**:
```go
// ⚠️ SUBOPTIMAL: Using default timeouts (slow boot)
ph, err := passport.InitHandler(jwksHost)
// Default: 2s connection timeout, 2 retries, 200ms delay = ~6s worst case

// ❌ BAD: Too aggressive timeouts (may cause failures)
ph, err := passport.InitHandler(
    jwksHost,
    passport.WithJwksFetchConnectionTimeout(100*time.Millisecond),  // Too low!
    passport.WithJwksFetchMaxRetryAttempts(0),  // No retries!
)
```

**How to Fix**:
```go
// ✅ GOOD: Optimized timeouts for production
ph, err := passport.InitHandler(
    jwksHost,
    passport.WithJwksFetchConnectionTimeout(1*time.Second),  // 1s (down from 2s)
    passport.WithJwksFetchMaxRetryAttempts(3),               // 3 retries (up from 2)
    passport.WithJwksFetchDelayBetweenRetries(100*time.Millisecond), // 100ms (down from 200ms)
)
// Worst case: 1s + (3 retries * 1s) + (3 * 100ms) = ~4.3s (vs 6s default)

// ✅ GOOD: Different timeouts for different environments
var timeout time.Duration
var retries int

if config.Env == "production" {
    timeout = 1 * time.Second
    retries = 3
} else {
    timeout = 500 * time.Millisecond  // Faster for dev/test
    retries = 2
}

ph, err := passport.InitHandler(
    jwksHost,
    passport.WithJwksFetchConnectionTimeout(timeout),
    passport.WithJwksFetchMaxRetryAttempts(retries),
)
```

**Razorpay Standard - Recommended Timeouts**:
- **Connection timeout**: `1s` (default: 2s)
- **Max retries**: `3` (default: 2)
- **Retry delay**: `100ms` (default: 200ms)
- **Production**: Higher retries for reliability
- **Dev/Test**: Lower timeouts for faster iteration

**Rationale**: Default timeouts slow down pod boot times (6s vs 4s).

---

## Check #10: Logging & Observability (Critical)

**Problem**: No logging of passport failures makes debugging difficult.

**Detection Strategy**:
```bash
# Find FromToken error handling
grep -A 5 "FromToken" <pr_files> | grep -E "logger\.|log\.|Error|Info"

# Check for metric emission
grep -E "metrics.*passport|passport.*metric" <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Silent failure
p, err := passportHandler.FromToken(jwtToken)
if err != nil {
    c.AbortWithStatus(401)  // <-- No logging!
    return
}

// ❌ BAD: Logging without context
p, err := passportHandler.FromToken(jwtToken)
if err != nil {
    fmt.Println("passport error:", err)  // <-- Lost in logs!
}

// ❌ BAD: Not logging successful passport extraction
p, err := passportHandler.FromToken(jwtToken)
if err != nil {
    logger.Error(ctx, "PASSPORT_ERROR", "error", err)
    return
}
// Missing: log success with merchant ID, roles
```

**How to Fix**:
```go
// ✅ GOOD: Log failures with context
p, err := passportHandler.FromToken(jwtToken)
if err != nil {
    logger.Error(c, "PASSPORT_VALIDATION_FAILED",
        "error", err,
        "token_present", jwtToken != "",
    )
    metrics.Count(c, "passport.validation.failed", 1)
    c.AbortWithStatus(http.StatusUnauthorized)
    return
}

// ✅ GOOD: Log successful extraction with key fields
logger.Info(c, "PASSPORT_EXTRACTED",
    "consumer_id", p.GetConsumerClaims().ID,
    "consumer_type", p.GetConsumerClaims().Type,
    "mode", p.GetMode(),
    "roles", p.GetRoles(),
    "authenticated", p.IsAuthenticated(),
    "identified", p.IsIdentified(),
)

// ✅ GOOD: Emit metrics for monitoring
metrics.Count(c, "passport.validation.success", 1,
    "mode", p.GetMode(),
    "consumer_type", p.GetConsumerClaims().Type,
)

c.Set("passport", p)

// ✅ GOOD: Log resource owner extraction
ownerID, err := passport.GetResourceOwnerID(p)
if err != nil {
    logger.Error(c, "PASSPORT_OWNER_EXTRACTION_FAILED", "error", err)
    return err
}

logger.Info(c, "PASSPORT_OWNER_EXTRACTED",
    "owner_id", ownerID,
    "owner_type", p.GetConsumerClaims().Type,
)
```

**Razorpay Standard - Logging**:
- **Trace codes**: Use constants (e.g., `PASSPORT_VALIDATION_FAILED`)
- **Success logs**: Log consumer ID, type, mode, roles
- **Failure logs**: Log error, token presence
- **Metrics**: Emit success/failure counts with dimensions
- **Context**: Always include request context

**Rationale**: Missing logs make debugging auth failures impossible.

---

## Integration Instructions

### When to Load This File
Load when PR contains any of:
- `import.*passport` or `github.com/razorpay/goutils/passport`
- `passport.InitHandler` or `passport.IPassport`
- `FromToken`, `GetResourceOwnerID`, `GetConsumerClaims`
- Middleware with `X-Passport-JWT-V1` header
- New authenticated endpoint

### Progressive Loading
Only load if Passport-related code changes detected. Defer until actual Passport usage confirmed.

### File Pattern Detection

```bash
# Detect Passport usage
grep -r "passport\.InitHandler\|passport\.IPassport" --include="*.go"
grep -r "github.com/razorpay/goutils/passport" go.mod
grep -r "X-Passport-JWT-V1\|HeaderKeyPassportJWTV1" --include="*.go"

# Check middleware
grep -r "ExtractPassport\|PassportMiddleware" --include="*.go"

# Check config
grep -r "\[passport\]" --include="*.toml"
```

### Common Issue Patterns

| Issue | Checks | Severity |
|-------|--------|----------|
| Wrong JWKS host | #1 | Critical |
| Header extraction error | #2 | Critical |
| Ignored validation error | #3 | Critical |
| Test bypass in prod | #6 | Critical |
| No logging | #10 | Critical |
| Unsafe type assertion | #4 | High |
| Wrong owner extraction | #5 | High |
| No mode check | #7 | High |
| Missing role check | #8 | High |
| Slow init timeouts | #9 | Medium |

### Environment-Specific Hosts

| Environment | JWKS Host |
|-------------|-----------|
| Devstack | `https://edge-base.dev.razorpay.in` |
| Stage | `https://edge-admin.int.stage.razorpay.in` |
| Automation | `https://edge-admin.int.qa.razorpay.in` |
| Production | `https://edge-admin-internal.razorpay.com` |

### Reference Links
- SDK: `github.com/razorpay/goutils/passport/v4`
- Documentation: https://write.razorpay.com/doc/about-edge-passport-mCa579K52t
- Support: #platform_spine_edge on Slack
