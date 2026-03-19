# Router SDK Integration Checks

Router SDK (goutils/router) is used to call pg-router for smart terminal selection in payment flows. This document covers 10 critical checks for router integration.

---

## Check #1: Client Initialization with Mode Validation 🚨 CRITICAL

### What to Check
Verify router client is initialized with proper configuration and mode is set from environment config (test/live), not hardcoded.

### Bad Pattern ❌
```go
// Hardcoded mode
routerClient := router.NewClient(baseClient, &config, "test")

// Missing error handling on config load
routerConfig := config.Get().Router
client := router.NewClient(baseClient, &routerConfig, boot.Mode())
```

### Good Pattern ✅
```go
// Mode from environment config
mode := boot.Mode() // Returns "test" or "live" based on environment
if mode == "" {
    return nil, fmt.Errorf("boot mode not configured")
}

routerConfig := config.Get().Router
if routerConfig.BaseURL == "" {
    return nil, fmt.Errorf("router base URL not configured")
}

routerClient := router.NewClient(baseClient, &routerConfig, mode)
```

### Detection Strategy
1. Search for `router.NewClient` or `router.New` calls
2. Check if third parameter (mode) comes from `boot.Mode()` or config
3. Verify mode is not a hardcoded string literal

### Flag Conditions
- **CRITICAL**: Mode hardcoded to "test" or "live" → Production/test traffic routing wrong environment
- **HIGH**: Missing validation of router config before client initialization

### Severity
🚨 **CRITICAL** - Hardcoded mode can route production payments to test terminals or vice versa, causing payment failures

---

## Check #2: Request Validation Before Router Call 🚨 CRITICAL

### What to Check
Verify payment and merchant entities are validated (not nil) before creating router request.

### Bad Pattern ❌
```go
// No nil checks
terminalReq, _ := routerClient.NewFetchTerminalsRequest(ctx, payment, req, false)
terminals, _ := routerClient.ExecuteFetchTerminalsRequest(ctx, terminalReq)
```

### Good Pattern ✅
```go
if payment == nil || actionReq == nil {
    return nil, apperrors.ErrServer.New(codes.UpsValidationFailure).
        Wrap(fmt.Errorf("invalid input provided to router request"))
}

terminalReq, err := routerClient.NewFetchTerminalsRequest(ctx, payment, actionReq, ccOnUpiOfferApplied)
if err != nil {
    return nil, err
}

if terminalReq == nil {
    return nil, fmt.Errorf("failed to create router request")
}

terminals, err := routerClient.ExecuteFetchTerminalsRequest(ctx, terminalReq)
if err != nil {
    logger.Error(ctx, "ROUTER_FETCH_TERMINALS_FAILED", "error", err)
    return nil, err
}
```

### Detection Strategy
1. Find `NewFetchTerminalsRequest` calls
2. Check for nil validation of payment/merchant before the call
3. Verify error handling exists for both request creation and execution

### Flag Conditions
- **CRITICAL**: Missing nil checks on payment/merchant → Panic risk
- **HIGH**: Ignored errors from request creation or execution
- **MEDIUM**: Missing error logging on router failures

### Severity
🚨 **CRITICAL** - Missing nil checks can cause panic in router SDK (seen in payments-upi router.go:77)

---

## Check #3: Appropriate Request Factory Method ⚠️ HIGH

### What to Check
Use correct request factory method based on payment type (standard vs recurring vs intent).

### Bad Pattern ❌
```go
// Using standard request for recurring payment
if payment.IsRecurring() {
    // ❌ Should use NewFetchTerminalsRequestForRecurring
    terminalReq, err := routerClient.NewFetchTerminalsRequest(ctx, payment, req, false)
}

// Using payment request for intent flow
if intent != nil {
    // ❌ Should use NewFetchTerminalsRequestForIntent
    terminalReq, err := routerClient.NewFetchTerminalsRequest(ctx, intent, req, false)
}
```

### Good Pattern ✅
```go
var terminalReq RouterReq.IRequest
var err error

if payment.IsRecurring() {
    // Use recurring-specific factory that includes mandate frequency
    terminalReq, err = routerClient.NewFetchTerminalsRequestForRecurring(
        ctx, payment, req, ccOnUpiOfferApplied, mandateFrequency,
    )
} else {
    // Standard payment flow
    terminalReq, err = routerClient.NewFetchTerminalsRequest(
        ctx, payment, req, ccOnUpiOfferApplied,
    )
}

// For intent flows (separate from payment)
if isIntentFlow {
    terminalReq, err = routerClient.NewFetchTerminalsRequestForIntent(
        ctx, intent, createReq,
    )
}

if err != nil {
    return nil, err
}
```

### Detection Strategy
1. Search for recurring payment code paths (`payment.IsRecurring()`, `payment.RecurringType`)
2. Check if `NewFetchTerminalsRequestForRecurring` is used instead of `NewFetchTerminalsRequest`
3. Look for intent flows using `NewFetchTerminalsRequestForIntent`

### Flag Conditions
- **HIGH**: Using standard request for recurring → Router won't select recurring-enabled terminals
- **HIGH**: Using payment request for intent → Type mismatch, missing intent-specific fields
- **MEDIUM**: Mandate frequency missing in recurring request

### Severity
⚠️ **HIGH** - Wrong request type leads to incorrect terminal selection, payment failures

---

## Check #4: Mode Consistency Across Request ⚠️ HIGH

### What to Check
Ensure mode is consistently set in router client initialization and request execution (test payments → test terminals).

### Bad Pattern ❌
```go
// Client initialized with "live" but request built with test mode
routerClient := router.NewClient(baseClient, &config, "live")

req := RouterReq.NewFetchTerminals()
req.SetMode("test") // ❌ Inconsistent with client mode
```

### Good Pattern ✅
```go
mode := boot.Mode()
routerClient := router.NewClient(baseClient, &config, mode)

// Request automatically uses same mode via client method
terminalReq, err := routerClient.NewFetchTerminalsRequest(ctx, payment, actionReq, false)
// Internally sets req.SetMode(c.mode) - consistent with client

// If manually building request, use same mode
req := RouterReq.NewFetchTerminals()
req.SetMode(mode) // Same as client
```

### Detection Strategy
1. Find router client initialization mode
2. Check if requests manually call `SetMode()` with different value
3. Verify payment mode matches router mode

### Flag Conditions
- **HIGH**: Client mode differs from request mode → Routing mismatch
- **HIGH**: Payment mode (test/live) doesn't match router mode
- **MEDIUM**: Mode validation missing before router call

### Severity
⚠️ **HIGH** - Mode mismatch routes test payments to live terminals or vice versa

---

## Check #5: Terminal Response Validation 📋 MEDIUM

### What to Check
Validate router response contains terminals before proceeding with payment flow.

### Bad Pattern ❌
```go
terminals, err := routerClient.ExecuteFetchTerminalsRequest(ctx, terminalReq)
if err != nil {
    return nil, err
}

// ❌ No validation - accessing terminals[0] can panic
selectedTerminal := terminals[0]
```

### Good Pattern ✅
```go
terminals, err := routerClient.ExecuteFetchTerminalsRequest(ctx, terminalReq)
if err != nil {
    logger.Error(ctx, "ROUTER_FETCH_FAILED", "error", err)
    return nil, err
}

if terminals == nil || len(terminals) == 0 {
    logger.Error(ctx, "NO_TERMINALS_AVAILABLE",
        "payment_id", payment.ID,
        "method", payment.Method,
        "merchant_id", merchant.ID,
    )
    return nil, apperrors.ErrValidationFailure.New(codes.NoTerminalsAvailable).
        Wrap(fmt.Errorf("no terminals available for routing"))
}

logger.Info(ctx, "ROUTER_TERMINALS_FETCHED",
    "terminal_count", len(terminals),
    "payment_id", payment.ID,
)

selectedTerminal := terminals[0]
```

### Detection Strategy
1. Find `ExecuteFetchTerminalsRequest` calls
2. Check for empty slice validation before accessing terminals
3. Verify appropriate error code when no terminals found

### Flag Conditions
- **HIGH**: Accessing terminals without nil/empty check → Panic risk
- **MEDIUM**: No logging when no terminals available
- **LOW**: Missing terminal count in success logs

### Severity
📋 **MEDIUM** - Missing validation can cause panic on empty response

---

## Check #6: Error Handling with Status Code Mapping ⚠️ HIGH

### What to Check
Handle router errors based on HTTP status codes and map to appropriate application errors.

### Bad Pattern ❌
```go
terminals, err := routerClient.ExecuteFetchTerminalsRequest(ctx, terminalReq)
if err != nil {
    // ❌ Generic error, no differentiation between validation vs server errors
    return nil, fmt.Errorf("router failed: %w", err)
}
```

### Good Pattern ✅
```go
terminals, err := routerClient.ExecuteFetchTerminalsRequest(ctx, terminalReq)
if err != nil {
    logger.Error(ctx, "ROUTER_FETCH_TERMINALS_FAILED",
        "error", err,
        "payment_id", payment.ID,
        "merchant_id", merchant.ID,
    )

    // Router client already maps status codes to appropriate errors
    // 400 → apperrors.ErrValidationFailure (codes.GatewayValidationFailure)
    // 500 → apperrors.ErrServer (codes.ServerError)

    // Propagate the typed error with context
    return nil, err.WithContext(ctx,
        "payment_id", payment.ID,
        "merchant_id", merchant.ID,
    )
}

responseLog := map[string]interface{}{
    "terminal_count": len(terminals),
    "payment_id":     payment.ID,
}

logger.Info(ctx, "ROUTER_RESPONSE_SUCCESS", responseLog)
```

### Detection Strategy
1. Find router error handling blocks
2. Check if errors are logged with trace codes
3. Verify errors include payment/merchant context

### Flag Conditions
- **HIGH**: Router errors not logged → Lost debugging context
- **MEDIUM**: Generic error wrapping loses router error classification
- **MEDIUM**: Missing payment/merchant ID in error context

### Severity
⚠️ **HIGH** - Poor error handling makes production debugging difficult

---

## Check #7: Basic Auth Configuration ⚠️ HIGH

### What to Check
Verify router client has Basic Auth credentials configured (username/password).

### Bad Pattern ❌
```go
routerConfig := &router.Config{
    Config: clients.Config{
        BaseURL: "https://pg-router.razorpay.com",
        Timeout: 30 * time.Second,
    },
    // ❌ Missing Auth credentials
}

client := router.NewClient(baseClient, routerConfig, mode)
```

### Good Pattern ✅
```go
routerConfig := &router.Config{
    Config: clients.Config{
        BaseURL: config.Get().Router.BaseURL,
        Timeout: 30 * time.Second,
    },
    Auth: struct {
        Username string
        Password string
    }{
        Username: config.Get().Router.Auth.Username,
        Password: config.Get().Router.Auth.Password,
    },
}

if routerConfig.Auth.Username == "" || routerConfig.Auth.Password == "" {
    return nil, fmt.Errorf("router auth credentials not configured")
}

client := router.NewClient(baseClient, routerConfig, mode)
```

### Detection Strategy
1. Find router Config struct initialization
2. Check if Auth.Username and Auth.Password are set
3. Verify credentials come from config, not hardcoded

### Flag Conditions
- **CRITICAL**: Missing auth credentials → 401 Unauthorized from pg-router
- **CRITICAL**: Hardcoded credentials in code → Security risk
- **HIGH**: Empty validation missing for credentials

### Severity
⚠️ **HIGH** - Missing auth causes all router calls to fail with 401

---

## Check #8: Context Propagation for Tracing ⚠️ HIGH

### What to Check
Ensure request context contains required headers for distributed tracing (X-Request-ID, etc.).

### Bad Pattern ❌
```go
// Using background context
terminals, err := routerClient.ExecuteFetchTerminalsRequest(
    context.Background(), // ❌ No trace headers
    terminalReq,
)
```

### Good Pattern ✅
```go
// Router client automatically propagates context headers
// Ensure context has required headers before calling router
for _, key := range contextkeys.HeaderKeys() {
    if ctx.Value(key) == nil {
        logger.Warn(ctx, "MISSING_CONTEXT_HEADER", "key", key)
    }
}

// Use request context with trace headers
terminals, err := routerClient.ExecuteFetchTerminalsRequest(
    ctx, // Context from gin.Context or gRPC request
    terminalReq,
)

// Router client internally does:
// req.SetHeaders(headers) - propagates X-Request-ID, X-Task-ID, etc.
// req.SetBasicAuth(username, password)
```

### Detection Strategy
1. Check if router calls use request context (not background context)
2. Verify context headers are set upstream (in HTTP/gRPC handlers)
3. Look for missing trace IDs in router error logs

### Flag Conditions
- **HIGH**: Using context.Background() → Lost distributed tracing
- **MEDIUM**: Missing request ID validation before router call
- **LOW**: No logging of trace headers for debugging

### Severity
⚠️ **HIGH** - Without context, can't trace payment flow across services

---

## Check #9: Timeout Configuration ⚠️ HIGH

### What to Check
Configure appropriate timeout for router calls (default 30s, not indefinite).

### Bad Pattern ❌
```go
routerConfig := &router.Config{
    Config: clients.Config{
        BaseURL: "https://pg-router.razorpay.com",
        // ❌ No timeout - uses HTTP client default (indefinite)
    },
}
```

### Good Pattern ✅
```go
routerConfig := &router.Config{
    Config: clients.Config{
        BaseURL: config.Get().Router.BaseURL,
        Timeout: 30 * time.Second, // Explicit timeout
    },
    Auth: struct {
        Username string
        Password string
    }{
        Username: config.Get().Router.Auth.Username,
        Password: config.Get().Router.Auth.Password,
    },
}

// If using custom HTTP client, ensure timeout is set
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}
```

### Detection Strategy
1. Find router Config initialization
2. Check if Timeout field is explicitly set
3. Verify timeout value is reasonable (10-60s range)

### Flag Conditions
- **HIGH**: Missing timeout → Indefinite hangs possible
- **MEDIUM**: Timeout > 60s → Too long for payment flow
- **MEDIUM**: Timeout < 5s → Too short, frequent failures

### Severity
⚠️ **HIGH** - Missing timeout can cause indefinite request hangs

---

## Check #10: Request Logging for Debugging 📋 MEDIUM

### What to Check
Log router request details before execution and response after (with PII masking).

### Bad Pattern ❌
```go
terminals, err := routerClient.ExecuteFetchTerminalsRequest(ctx, terminalReq)
// ❌ No request logging
if err != nil {
    return nil, err
}
// ❌ No response logging
```

### Good Pattern ✅
```go
// Log request (router client has ToTrace() method for PII masking)
requestLog := map[string]interface{}{
    "URI":     terminalReq.GetURI(), // "route_authn"
    "request": terminalReq.ToTrace(), // Masks card IIN, name
    "mode":    terminalReq.GetMode(),
}
logger.Info(ctx, "ROUTER_REQUEST", requestLog)

terminals, err := routerClient.ExecuteFetchTerminalsRequest(ctx, terminalReq)
if err != nil {
    logger.Error(ctx, "ROUTER_RESPONSE_ERROR",
        "error", err,
        "request": terminalReq.ToTrace(),
    )
    return nil, err
}

// Log response (mask terminal secrets)
responseLog := map[string]interface{}{
    "terminal_count": len(terminals),
    "payment_id":     terminalReq.GetPayment().ID,
    "merchant_id":    terminalReq.GetMerchantID(),
}
logger.Info(ctx, "ROUTER_RESPONSE_SUCCESS", responseLog)
```

### Detection Strategy
1. Find `ExecuteFetchTerminalsRequest` calls
2. Check for logger calls before and after execution
3. Verify PII fields (card IIN, name) are masked via ToTrace()

### Flag Conditions
- **MEDIUM**: No request logging → Hard to debug router issues
- **MEDIUM**: No response logging → Can't verify terminal selection
- **HIGH**: Logging raw request with PII (card IIN, name) → Security risk

### Severity
📋 **MEDIUM** - Missing logs make production debugging difficult; PII leaks are HIGH severity

---

## Summary

| Check | Severity | Description |
|-------|----------|-------------|
| 1. Client Initialization | 🚨 CRITICAL | Mode validation (test/live) from config |
| 2. Request Validation | 🚨 CRITICAL | Nil checks before router call |
| 3. Request Factory Method | ⚠️ HIGH | Use correct method for payment type |
| 4. Mode Consistency | ⚠️ HIGH | Client mode matches request mode |
| 5. Response Validation | 📋 MEDIUM | Check terminals not empty before access |
| 6. Error Handling | ⚠️ HIGH | Map status codes to app errors |
| 7. Basic Auth Config | ⚠️ HIGH | Username/password from config |
| 8. Context Propagation | ⚠️ HIGH | Trace headers in request context |
| 9. Timeout Configuration | ⚠️ HIGH | Explicit timeout (30s recommended) |
| 10. Request Logging | 📋 MEDIUM | Log with PII masking |

## File Triggers

Load this reference when PR modifies files matching:
- `**/router*.go` - Router client usage
- `**/*payment*initiate*.go` - Payment initiation with routing
- `**/*intent*create*.go` - Intent flows with routing
- `import.*goutils/router` - Router SDK imports
- Files calling `GetTerminals`, `NewFetchTerminalsRequest`

## Common Repositories

- **payments-upi**: `internal/pkg/clients/router/router.go` (client wrapper)
- **pg-router**: The routing service itself (not checked by this reference)
- **goutils/router**: SDK library (not checked, used as reference)

## References

Based on real patterns from:
- `razorpay/payments-upi` - Router client implementation
- `razorpay/goutils/router` - Router SDK
- `razorpay/pg-router` - Smart routing service (HLD)
