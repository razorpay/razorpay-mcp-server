# Services: ASV Integration (10 Checks)

Validates Account Service (ASV) integration patterns. ASV is Razorpay's account/merchant entity service that decouples merchant data from API monolith. Provides gRPC API for fetching/updating accounts, stakeholders, contacts, documents.

**Total Checks**: 10 (5 Critical, 4 High, 1 Medium)

---

## Check #1: Client Configuration & Initialization (Critical)

**Problem**: Missing or incorrect ASV configuration causes all API calls to fail.

**Detection Strategy**:
```bash
# Find ASV client initialization
grep -E "accountService\.NewClient|asv\.NewClient|NewAsvService" <pr_files>

# Check config files for ASV settings
grep -E "\[.*accountService\]|\[.*asv\]" configs/*.toml

# Verify credentials configuration
grep -E "baseUrl|BaseURL|ServerURL" configs/*.toml | grep -i "asv\|account"
```

**What to Flag**:
```go
// ❌ BAD: Hardcoded server URL and credentials
config := asvConfig.DefaultConfig().
    WithServerURL("asv.grpc.int.dev.razorpay.in:443").  // <-- Hardcoded!
    WithCredentials(&asvConfig.UserCredentials{
        ClientID: "terminals",      // <-- Hardcoded!
        Password: "password",       // <-- Hardcoded!
    })

// ❌ BAD: No validation of config values
baseURL := os.Getenv("ASV_BASE_URL")  // <-- Could be empty!
client, _ := accountService.NewClient(ctx,
    accountService.WithConfig(config))

// ❌ BAD: Wrong server URL for environment
// In production deployment:
config := asvConfig.DefaultConfig().
    WithServerURL("asv.grpc.int.dev.razorpay.in:443")  // <-- Dev URL in prod!

// ❌ BAD: Missing ASV config in TOML
// configs/prod.toml - no [service.accountService] section!
```

**How to Fix**:
```go
// ✅ GOOD: Load from config with validation
func initASVClient(ctx context.Context, cfg *config.Config) (*accountService.Client, error) {
    if cfg.Service.AccountService.BaseURL == "" {
        return nil, errors.New("ASV base URL not configured")
    }
    if cfg.Service.AccountService.Auth.Key == "" || cfg.Service.AccountService.Auth.Secret == "" {
        return nil, errors.New("ASV credentials not configured")
    }

    asvConfig := asvConfig.DefaultConfig().
        WithServerURL(cfg.Service.AccountService.BaseURL).
        WithCredentials(&asvConfig.UserCredentials{
            ClientID: cfg.Service.AccountService.Auth.Key,
            Password: cfg.Service.AccountService.Auth.Secret,
        })

    client, err := accountService.NewClient(ctx,
        accountService.WithConfig(asvConfig),
        accountService.WithClientId("terminals"))
    if err != nil {
        return nil, fmt.Errorf("failed to init ASV client: %w", err)
    }

    return client, nil
}

// ✅ GOOD: Config in TOML with environment-specific values
// configs/default.toml (dev):
[service.accountService]
    baseUrl = "asv.grpc.int.dev.razorpay.in:443"
    [service.accountService.auth]
        key = "terminals"
        secret = "password"

// configs/stage.toml:
[service.accountService]
    baseUrl = "asv.grpc.int.stage.razorpay.in:443"
    [service.accountService.auth]
        key = "terminals"
        secret = "{{ vault_secret }}"

// configs/prod-live.toml:
[service.accountService]
    baseUrl = "asv.grpc.razorpay.com:443"
    [service.accountService.auth]
        key = "terminals"
        secret = "{{ vault_secret }}"
```

**Razorpay Standard - ASV Hosts**:
- **Dev**: `asv.grpc.int.dev.razorpay.in:443`
- **Stage**: `asv.grpc.int.stage.razorpay.in:443`
- **Production**: `asv.grpc.razorpay.com:443`

**Client Registration**:
- Service must be registered in ASV server's `config/default.toml`
- Contact: #platform_account_service on Slack
- Tag: @acct-svc-devs

**Rationale**: Wrong host or missing credentials cause all ASV calls to fail with auth errors.

---

## Check #2: Paths Field Validation (Critical)

**Problem**: Missing `Paths` field fetches entire object (performance issue), invalid paths cause errors.

**Detection Strategy**:
```bash
# Find GetByID calls
grep -E "GetByID|GetAccountByIDRequest" <pr_files>

# Check if Paths field is specified
grep -A 10 "GetAccountByIDRequest" <pr_files> | grep "Paths:"

# Find Save calls
grep -E "SaveRequest|Write\(\)\.Save" <pr_files>

# Check if Paths are specified for Save
grep -A 10 "SaveRequest" <pr_files> | grep "Paths:"
```

**What to Flag**:
```go
// ❌ BAD: No Paths specified (fetches everything!)
req := &dto.GetAccountByIDRequest{
    Id: merchantID,
    // Missing: Paths field!
}
response, _ := asvClient.Account().GetByID(ctx, req)
// This fetches ALL account fields - huge performance hit!

// ❌ BAD: Invalid path syntax
req := &dto.GetAccountByIDRequest{
    Id: merchantID,
    Paths: []string{
        "account/name",  // <-- Wrong! Should use dot notation
        "accountName",   // <-- Wrong! Missing entity prefix
    },
}

// ❌ BAD: Requesting non-existent fields
req := &dto.GetAccountByIDRequest{
    Id: merchantID,
    Paths: []string{
        "account.merchant_name",  // <-- Field doesn't exist!
        "account.details.phone",  // <-- Wrong path structure!
    },
}

// ❌ BAD: Save without Paths (updates nothing!)
saveReq := &dto.SaveRequest{
    Account: &dto.Account{
        Id:   merchantID,
        Name: ConvertToPointer("New Name"),
    },
    // Missing: Paths field!
}
```

**How to Fix**:
```go
// ✅ GOOD: Specify only needed fields with correct paths
req := &dto.GetAccountByIDRequest{
    Id: merchantID,
    Paths: []string{
        "account.name",
        "account.email",
        "account.account_detail.contact_name",
        "account.account_detail.contact_mobile",
        "account.business_detail.website_details",
        "documents.id",
        "documents.file_store_id",
        "documents.document_type",
    },
}

response, err := asvClient.Account().GetByID(ctx, req)

// ✅ GOOD: Save with explicit Paths
saveReq := &dto.SaveRequest{
    Account: &dto.Account{
        Id:   merchantID,
        Name: ConvertToPointer("ABC Technology"),
        Email: ConvertToPointer("abc@gmail.com"),
        AccountDetail: &dto.AccountDetail{
            ContactName: ConvertToPointer("John Doe"),
        },
    },
    Paths: []string{
        "account.name",
        "account.email",
        "account.account_detail.contact_name",
    },
}

_, err := asvClient.Write().Save(ctx, saveReq)
```

**Valid Path Patterns**:
- **Account**: `account.{field}` (e.g., `account.name`, `account.email`)
- **Account Detail**: `account.account_detail.{field}`
- **Business Detail**: `account.business_detail.{field}`
- **Website Detail**: `account.business_detail.website_details`
- **Stakeholders**: `stakeholders.{field}`
- **Stakeholder Address**: `stakeholders.residential_address.{field}`
- **Contacts**: `contacts.{field}`
- **Documents**: `documents.{field}`

**Common Fields**:
```go
// Account
"account.id", "account.name", "account.email", "account.org_id"

// Account Detail
"account.account_detail.contact_name", "account.account_detail.contact_mobile",
"account.account_detail.transaction_volume", "account.account_detail.business_description"

// Business Detail
"account.business_detail.app_urls", "account.business_detail.website_details"

// Documents
"documents.id", "documents.account_id", "documents.file_store_id",
"documents.document_type", "documents.source", "documents.entity_type"

// Stakeholders
"stakeholders.id", "stakeholders.name", "stakeholders.email"

// Contacts
"contacts.id", "contacts.type", "contacts.email", "contacts.verified"
```

**Rationale**: Missing Paths fetches all fields (slow + high memory), invalid paths cause API errors.

---

## Check #3: Error Handling & Nil Checks (Critical)

**Problem**: Ignored errors or missing nil checks cause panics.

**Detection Strategy**:
```bash
# Find ASV API calls
grep -E "GetByID|Save|Delete" <pr_files> | grep -v "test"

# Check if errors are handled
grep -A 3 "asvClient\." <pr_files> | grep "if err"

# Check for nil response checks
grep -A 5 "GetByID" <pr_files> | grep "!= nil\|== nil"
```

**What to Flag**:
```go
// ❌ BAD: Ignoring errors
response, _ := asvClient.Account().GetByID(ctx, req)  // <-- Ignores error!
name := response.Name  // <-- Panic if response is nil!

// ❌ BAD: No nil check on response
response, err := asvClient.Account().GetByID(ctx, req)
if err != nil {
    logger.Error(ctx, "ASV_ERROR", "error", err)
    // Continues without return!
}
name := *response.Name  // <-- Panic if response/Name is nil!

// ❌ BAD: Test mode check in wrong place
response, err := asvClient.Account().GetByID(ctx, req)
if err != nil {
    if gin.Mode() != gin.TestMode {  // <-- Wrong! Should check before call
        return err
    }
}

// ❌ BAD: Accessing nested fields without nil checks
docs := response.Documents
fileId := *docs[0].FileStoreId  // <-- Multiple panic risks!
```

**How to Fix**:
```go
// ✅ GOOD: Proper error handling with nil checks
response, err := asvClient.Account().GetByID(ctx, req)
if err != nil {
    logger.Error(ctx, "ASV_GET_ACCOUNT_FAILED", "error", err, "merchant_id", merchantID)
    return nil, fmt.Errorf("failed to get account: %w", err)
}

if response == nil {
    logger.Error(ctx, "ASV_RESPONSE_NIL", "merchant_id", merchantID)
    return nil, errors.New("ASV returned nil response")
}

// Safe field access
var name string
if response.Name != nil {
    name = *response.Name
}

// ✅ GOOD: Test mode check at service level
func (s *asvService) GetMerchantDocuments(ctx *gin.Context, merchantId string) ([]*dto.Document, error) {
    if asvClient == nil {
        logger.Info(ctx, "ASV_CLIENT_NIL", "merchant_id", merchantId)
        if gin.Mode() != gin.TestMode {
            return nil, errors.New("ASV client is not initialized")
        }
        return []*dto.Document{}, nil  // <-- Safe return for tests
    }

    response, err := asvClient.Account().GetByID(ctx, req)
    if err != nil {
        if gin.Mode() != gin.TestMode {
            return nil, err
        }
        return []*dto.Document{}, nil  // <-- Safe return for tests
    }

    return response.Documents, nil
}

// ✅ GOOD: Safe nested field access
var documents []*dto.Document
if response != nil && response.Documents != nil {
    for _, doc := range response.Documents {
        if doc != nil && doc.FileStoreId != nil {
            documents = append(documents, doc)
        }
    }
}
```

**Razorpay Standard**:
- **Always check errors** from all ASV API calls
- **Always check nil response** before field access
- **Use pointer helpers** for optional fields
- **Test mode gates** should be at service initialization or early in function
- **Log errors** with context (merchant_id, operation)

**Rationale**: Unchecked errors and nil values cause runtime panics.

---

## Check #4: Write Authorization (Critical)

**Problem**: Calling Write APIs without being authorized in ASV server config causes permission errors.

**Detection Strategy**:
```bash
# Find Write API calls
grep -E "Write\(\)\.Save|Write\(\)\.Delete" <pr_files>

# Check if service is likely authorized (look for PR link or comment)
grep -E "authorized|ASV.*PR|account-service.*PR" <pr_description>
```

**What to Flag**:
```go
// ❌ BAD: Using Write API without authorization check
saveReq := &dto.SaveRequest{
    Account: &dto.Account{
        Id:   merchantID,
        Name: ConvertToPointer("New Name"),
    },
    Paths: []string{"account.name"},
}

// No evidence of authorization in ASV server!
_, err := asvClient.Write().Save(ctx, saveReq)
```

**How to Fix**:
```
✅ GOOD: Authorization workflow before using Write APIs

Before using Write().Save() or Write().Delete():

1. Check if service is already authorized:
   - Look at ASV server config: https://github.com/razorpay/account-service
   - File: config/default.toml
   - Section: [AuthorizedClients]
   - Check if your service is listed

2. If NOT authorized, raise PR to ASV server:
   - Repo: https://github.com/razorpay/account-service
   - Add your service to [AuthorizedClients]
   - Example PR: https://github.com/razorpay/account-service/pull/584
   - Tag: @acct-svc-devs for review

3. After PR merged and deployed, use Write APIs:
   ```go
   // Now authorized to write
   saveReq := &dto.SaveRequest{
       Account: &dto.Account{
           Id:   merchantID,
           Name: ConvertToPointer("ABC Technology"),
       },
       Paths: []string{"account.name"},
   }

   _, err := asvClient.Write().Save(ctx, saveReq)
   if err != nil {
       return fmt.Errorf("ASV save failed: %w", err)
   }
   ```

4. Document authorization in PR description:
   "ASV Write access granted via https://github.com/razorpay/account-service/pull/XXX"
```

**Razorpay Standard**:
- **Read APIs** (GetByID): No authorization needed
- **Write APIs** (Save, Delete): Requires authorization in ASV server
- **Process**: Raise PR to account-service repo before using Write APIs
- **Contact**: #platform_account_service, @acct-svc-devs

**Rationale**: Unauthorized write attempts fail with permission errors.

---

## Check #5: Document Handling (Critical)

**Problem**: Incorrect FileStoreId handling or duplicate documents.

**Detection Strategy**:
```bash
# Find document operations
grep -E "FileStoreId|file_store_id|Documents" <pr_files>

# Check for prefix handling
grep -E "file_.*Replace|strings\.Replace.*file_" <pr_files>

# Check for document merging
grep -E "mergeExisting|existingDoc|GetMerchantDocuments" <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Not stripping "file_" prefix before saving
document := &dto.Document{
    AccountId:    merchantID,
    FileStoreId:  ConvertToPointer("file_ABC123"),  // <-- Prefix not stripped!
    DocumentType: "sebi_registration_certificate",
}

saveReq := &dto.SaveRequest{
    Documents: []*dto.Document{document},
    Paths:     []string{"documents.file_store_id", "documents.document_type"},
}

// ❌ BAD: Creating duplicate documents (not checking existing)
newDoc := &dto.Document{
    AccountId:    merchantID,
    DocumentType: "ffmc_license",  // <-- May already exist!
    FileStoreId:  ConvertToPointer("XYZ789"),
}
// No check if document_type already exists, creates duplicate!

// ❌ BAD: Invalid document type
document := &dto.Document{
    DocumentType: "merchant_license",  // <-- Invalid type!
}
```

**How to Fix**:
```go
// ✅ GOOD: Strip "file_" prefix before saving
const fileStoreIdPrefix = "file_"

for _, document := range documents {
    if document.FileStoreId != nil {
        *document.FileStoreId = strings.Replace(*document.FileStoreId,
            fileStoreIdPrefix, "", 1)
    }
}

saveReq := &dto.SaveRequest{
    Documents: documents,
    Paths:     []string{"documents.file_store_id", "documents.document_type"},
}

// ✅ GOOD: Merge with existing documents to avoid duplicates
func (s *asvService) PatchMerchantDocuments(ctx *gin.Context, documents []*dto.Document, merchantID string) error {
    // 1. Fetch existing documents
    existingDocs, err := s.GetMerchantDocuments(ctx, merchantID)
    if err != nil {
        return err
    }

    // 2. Create map by document_type
    existingDocsByType := make(map[string]*dto.Document)
    for _, doc := range existingDocs {
        if doc.DocumentType != "" {
            existingDocsByType[doc.DocumentType] = doc
        }
    }

    // 3. Merge: Update existing docs instead of creating duplicates
    for _, document := range documents {
        document.AccountId = merchantID

        // If document_type exists, reuse its ID (update instead of create)
        if existingDoc, exists := existingDocsByType[document.DocumentType]; exists {
            document.Id = existingDoc.Id
            logger.Info(ctx, "ASV_UPDATING_EXISTING_DOCUMENT",
                "document_type", document.DocumentType,
                "document_id", existingDoc.Id)
        }

        // Strip prefix
        if document.FileStoreId != nil {
            *document.FileStoreId = strings.Replace(*document.FileStoreId,
                fileStoreIdPrefix, "", 1)
        }
    }

    // 4. Save (creates new or updates existing)
    saveReq := &dto.SaveRequest{
        Documents: documents,
        Paths:     []string{"documents.document_type", "documents.file_store_id", "documents.source"},
    }

    _, err = asvClient.Write().Save(ctx, saveReq)
    return err
}

// ✅ GOOD: Valid document types
validDocTypes := []string{
    "sebi_registration_certificate",
    "irdai_registration_certificate",
    "ffmc_license",
    "nbfc_registration_certificate",
    "amfi_certificate",
    "shop_establishment",
    "gst_certificate",
    "business_pan",
    "cancelled_cheque",
    "form_60",
}
```

**Razorpay Standard - Document Types**:
- `sebi_registration_certificate`
- `irdai_registration_certificate`
- `ffmc_license`
- `nbfc_registration_certificate`
- `amfi_certificate`
- `shop_establishment`
- `gst_certificate`
- `business_pan`
- `cancelled_cheque`
- `form_60`

**FileStoreId Handling**:
- **Incoming**: May have `file_` prefix
- **Before Save**: Strip prefix using `strings.Replace(*id, "file_", "", 1)`
- **Storage**: ASV stores without prefix

**Rationale**: Wrong FileStoreId format or duplicate documents cause save failures.

---

## Check #6: Website Fields Validation (High)

**Problem**: Invalid URLs in website fields cause validation errors.

**Detection Strategy**:
```bash
# Find website fields usage
grep -E "WebsiteFields|website_details|WebsiteDetails" <pr_files>

# Check for URL validation
grep -E "Validate\(\)|validateURL" <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Not validating website URLs
websiteFields := WebsiteFields{
    Terms:   "not-a-url",  // <-- Invalid!
    Privacy: "http://",    // <-- Incomplete!
    Refund:  "example.com", // <-- Missing protocol!
}

// No validation before saving!
s.PatchMerchantWebsiteDetails(ctx, websiteFields, merchantID)

// ❌ BAD: Skipping validation in non-test mode
if err := websiteFields.Validate(); err != nil {
    if gin.Mode() != gin.TestMode {
        // Logs error but continues!
        logger.Error(ctx, "VALIDATION_ERROR", "error", err)
    }
}
// Saves invalid data!
```

**How to Fix**:
```go
// ✅ GOOD: Define validation rules
const urlValidationPattern = `^(https?:\/\/)[-a-zA-Z0-9@:%._\+~#?&//=]{2,256}$`
var urlRegex = regexp.MustCompile(urlValidationPattern)

type WebsiteFields struct {
    Pricing      string `json:"pricing"`
    Terms        string `json:"terms"`
    Privacy      string `json:"privacy"`
    Contact      string `json:"contact"`
    About        string `json:"about"`
    Refund       string `json:"refund"`
    Cancellation string `json:"cancellation"`
}

// Validate validates all website fields
func (w WebsiteFields) Validate() error {
    return validation.ValidateStruct(&w,
        validation.Field(&w.Pricing, validation.By(validateURL)),
        validation.Field(&w.Terms, validation.By(validateURL)),
        validation.Field(&w.Privacy, validation.By(validateURL)),
        validation.Field(&w.Contact, validation.By(validateURL)),
        validation.Field(&w.About, validation.By(validateURL)),
        validation.Field(&w.Refund, validation.By(validateURL)),
        validation.Field(&w.Cancellation, validation.By(validateURL)),
    )
}

func validateURL(value interface{}) error {
    urlString, ok := value.(string)
    if !ok {
        return validation.NewError("validation_url_invalid", "must be a valid URL string")
    }
    if urlString == "" {
        return nil // Allow empty strings
    }
    if !urlRegex.MatchString(urlString) {
        return validation.NewError("validation_url_format",
            "must be a valid URL with http:// or https://")
    }
    return nil
}

// ✅ GOOD: Validate before saving
func (s *asvService) PatchMerchantWebsiteDetails(ctx *gin.Context, websiteFields WebsiteFields, merchantID string) error {
    // Validate website fields
    if err := websiteFields.Validate(); err != nil {
        if gin.Mode() != gin.TestMode {
            return errors.New("Website field validation failed: " + err.Error())
        }
    }

    // Build website details map
    details := make(map[string]interface{})
    if websiteFields.Terms != "" {
        details["terms"] = websiteFields.Terms
    }
    if websiteFields.Privacy != "" {
        details["privacy"] = websiteFields.Privacy
    }
    // ... other fields

    req := &dto.SaveRequest{
        Account: &dto.Account{
            Id: merchantID,
            BusinessDetail: &dto.AccountBusinessDetail{
                WebsiteDetails: details,
            },
        },
        Paths: []string{"account.business_detail.website_details"},
    }

    _, err := asvClient.Write().Save(ctx, req)
    return err
}
```

**Razorpay Standard - Website Fields**:
- **Required protocol**: `http://` or `https://`
- **Valid format**: Standard URL format
- **Empty allowed**: Optional fields can be empty
- **Validation library**: `github.com/go-ozzo/ozzo-validation/v4`

**Rationale**: Invalid URLs cause ASV save failures and compliance issues.

---

## Check #7: Concurrent Operations Safety (High)

**Problem**: Goroutines without panic recovery or unsafe error handling.

**Detection Strategy**:
```bash
# Find goroutines in ASV operations
grep -B 5 -A 15 "go func" <pr_files> | grep -E "asv|GetMerchantDocuments|GetByID"

# Check for panic recovery
grep -A 10 "go func" <pr_files> | grep "defer.*recover"

# Check for WaitGroup
grep -E "sync\.WaitGroup|wg\.Add|wg\.Done" <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Goroutine without panic recovery
func GetMerchantData(ctx *gin.Context, merchantID string) (*MerchantData, error) {
    var asvDocs []*dto.Document
    var apiInfo *MerchantInfo
    var wg sync.WaitGroup

    wg.Add(2)

    go func() {
        defer wg.Done()
        // No panic recovery!
        asvDocs, _ = asvService.GetMerchantDocuments(ctx, merchantID)
    }()

    go func() {
        defer wg.Done()
        // No panic recovery!
        apiInfo, _ = apiService.GetMerchantInfo(ctx, merchantID)
    }()

    wg.Wait()
    return &MerchantData{Docs: asvDocs, Info: apiInfo}, nil
}

// ❌ BAD: No error aggregation
var apiErr, asvErr error
go func() {
    apiInfo, apiErr = apiService.GetMerchantInfo(ctx, merchantID)
}()
go func() {
    asvDocs, asvErr = asvService.GetMerchantDocuments(ctx, merchantID)
}()
wg.Wait()
// Errors silently ignored!

// ❌ BAD: Race condition on shared variable
var result *MerchantData
go func() {
    result = fetchFromAPI()  // <-- Race!
}()
go func() {
    result = fetchFromASV()  // <-- Race!
}()
```

**How to Fix**:
```go
// ✅ GOOD: Goroutines with panic recovery and error handling
func GetMerchantInfoWithDocuments(ctx *gin.Context, merchantID string) (*MerchantData, error) {
    var (
        merchantInfo *MerchantInfo
        asvDocuments []*dto.Document
        apiErr       error
        asvErr       error
        wg           sync.WaitGroup
    )

    wg.Add(2)

    // Fetch merchant info from API
    go func() {
        defer wg.Done()
        defer func() {
            if r := recover(); r != nil {
                logger.Error(ctx, "PANIC_API_FETCH", "panic", r, "merchant_id", merchantID)
                apiErr = fmt.Errorf("API fetch panicked: %v", r)
            }
        }()

        merchantInfo, apiErr = apiService.GetMerchantInfo(ctx, merchantID)
        if apiErr != nil {
            logger.Error(ctx, "API_FETCH_ERROR", "error", apiErr, "merchant_id", merchantID)
        }
    }()

    // Fetch documents from ASV
    go func() {
        defer wg.Done()
        defer func() {
            if r := recover(); r != nil {
                logger.Error(ctx, "PANIC_ASV_FETCH", "panic", r, "merchant_id", merchantID)
                asvErr = fmt.Errorf("ASV fetch panicked: %v", r)
            }
        }()

        asvDocuments, asvErr = asvService.GetMerchantDocuments(ctx, merchantID)
        if asvErr != nil {
            logger.Error(ctx, "ASV_FETCH_ERROR", "error", asvErr, "merchant_id", merchantID)
        }
    }()

    wg.Wait()

    // Handle errors - fail if both failed, succeed if at least one succeeded
    if apiErr != nil && asvErr != nil {
        return nil, fmt.Errorf("both services failed - API: %w, ASV: %v", apiErr, asvErr)
    }
    if apiErr != nil {
        return nil, fmt.Errorf("API fetch failed: %w", apiErr)
    }
    if asvErr != nil {
        // ASV failed but API succeeded - return partial data with error
        return nil, fmt.Errorf("ASV fetch failed: %w", asvErr)
    }

    // Map ASV documents to response
    merchantInfo.DocumentDetails = MapDocuments(asvDocuments)

    return &MerchantData{
        Info:      merchantInfo,
        Documents: asvDocuments,
    }, nil
}

// ✅ GOOD: Using mutex for shared state (if needed)
var (
    mu     sync.Mutex
    result *MerchantData
    err    error
)

go func() {
    data, fetchErr := fetchFromAPI()
    mu.Lock()
    if fetchErr == nil && result == nil {
        result = data
    }
    mu.Unlock()
}()
```

**Razorpay Standard**:
- **Always use panic recovery** in goroutines (`defer recover()`)
- **Aggregate errors** from parallel operations
- **Use mutex** for shared variable access
- **Log panics** with context for debugging
- **Graceful degradation**: Continue if one service fails

**Rationale**: Panics in goroutines crash the process, race conditions cause data corruption.

---

## Check #8: Cache Configuration (High)

**Problem**: No cache configured leads to performance issues.

**Detection Strategy**:
```bash
# Find ASV client initialization
grep -E "NewClient|accountService\.NewClient" <pr_files>

# Check if cache is configured
grep -E "WithCache|InMemoryCache|cache\.New" <pr_files>

# Check cache TTL configuration
grep -E "WithEviction|TTL|CacheConfig" <pr_files>
```

**What to Flag**:
```go
// ⚠️ SUBOPTIMAL: No cache configured (performance hit)
asvClient, err := accountService.NewClient(ctx,
    accountService.WithConfig(config),
    accountService.WithClientId("terminals"))
// Every call hits ASV server!

// ❌ BAD: Cache with very short TTL (defeats purpose)
cacheConfig := cache.NewInMemoryCacheConfig().
    WithEviction(1 * time.Second).  // <-- Too short!
    WithMaxCacheSize(1)              // <-- Too small!

// ❌ BAD: Cache with very long TTL (stale data risk)
cacheConfig := cache.NewInMemoryCacheConfig().
    WithEviction(1 * time.Hour)  // <-- Too long for merchant data!
```

**How to Fix**:
```go
// ✅ GOOD: Configure cache with optimal settings
func initASVClient(ctx context.Context, config *Config) (*accountService.Client, error) {
    // 1. Create cache with recommended settings
    cacheConfig := cache.NewInMemoryCacheConfig().
        WithEviction(30 * time.Second).  // 30s TTL (default: 15s)
        WithMaxCacheSize(1024).          // 1 GB max size (default: 8 MB)
        WithCleanWindow(1 * time.Second) // Cleanup interval

    inMemoryCache, err := cache.NewInMemoryCache(ctx, cacheConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create cache: %w", err)
    }

    // 2. Create ASV client with cache
    asvConfig := asvConfig.DefaultConfig().
        WithServerURL(config.ASV.BaseURL).
        WithCredentials(&asvConfig.UserCredentials{
            ClientID: config.ASV.Auth.Key,
            Password: config.ASV.Auth.Secret,
        })

    asvClient, err := accountService.NewClient(ctx,
        accountService.WithCache(inMemoryCache),  // <-- Enable cache
        accountService.WithConfig(asvConfig),
        accountService.WithClientId("terminals"))

    if err != nil {
        return nil, fmt.Errorf("failed to create ASV client: %w", err)
    }

    logger.Info(ctx, "ASV_CLIENT_INITIALIZED_WITH_CACHE",
        "cache_ttl_seconds", 30,
        "cache_max_size_mb", 1024)

    return asvClient, nil
}

// ✅ GOOD: Environment-specific cache settings
var cacheTTL time.Duration
var cacheSize int

switch config.Env {
case "production":
    cacheTTL = 30 * time.Second
    cacheSize = 2048  // 2 GB for prod
case "staging":
    cacheTTL = 15 * time.Second
    cacheSize = 1024  // 1 GB for stage
default:
    cacheTTL = 10 * time.Second
    cacheSize = 512   // 512 MB for dev
}

cacheConfig := cache.NewInMemoryCacheConfig().
    WithEviction(cacheTTL).
    WithMaxCacheSize(cacheSize)
```

**Razorpay Standard - Cache Settings**:
- **TTL**: 30 seconds (balance between freshness and performance)
- **Max Size**: 1-2 GB (based on expected merchant data volume)
- **Clean Window**: 1 second (default, don't change)
- **Cache Library**: `github.com/razorpay/goutils/account-service/v2/cache`

**When NOT to use cache**:
- Write operations (always fresh)
- Real-time critical data (KYC status changes)
- Very low traffic services (cache overhead > benefit)

**Rationale**: No cache means every GetByID call hits ASV server (high latency, load).

---

## Check #9: Singleton Client Pattern (High)

**Problem**: Creating new ASV client per request wastes resources.

**Detection Strategy**:
```bash
# Find client initialization in handlers/services
grep -E "NewClient|accountService\.NewClient" <pr_files>

# Check if using sync.Once
grep -E "sync\.Once|once\.Do" <pr_files>

# Look for client as function-local variable
grep -B 5 "NewClient" <pr_files> | grep "func.*Handler\|func.*Service"
```

**What to Flag**:
```go
// ❌ BAD: Creating new client per request
func GetMerchantDocuments(ctx *gin.Context, merchantID string) ([]*dto.Document, error) {
    // Creates new client on every call!
    config := asvConfig.DefaultConfig().WithServerURL(baseURL)
    asvClient, _ := accountService.NewClient(ctx, accountService.WithConfig(config))

    req := &dto.GetAccountByIDRequest{Id: merchantID, Paths: []string{"documents.id"}}
    response, _ := asvClient.Account().GetByID(ctx, req)
    return response.Documents, nil
}

// ❌ BAD: Multiple client instances
var asvClient1 *accountService.Client
var asvClient2 *accountService.Client

func init() {
    asvClient1, _ = accountService.NewClient(ctx, ...)  // Instance 1
    asvClient2, _ = accountService.NewClient(ctx, ...)  // Instance 2!
}
```

**How to Fix**:
```go
// ✅ GOOD: Singleton pattern with sync.Once
var (
    once      sync.Once
    asvClient *accountService.Client
)

func getAsvClient(ctx context.Context) *accountService.Client {
    once.Do(func() {
        var err error

        cacheConfig := cache.NewInMemoryCacheConfig().
            WithEviction(30 * time.Second).
            WithMaxCacheSize(1024)

        inMemoryCache, err := cache.NewInMemoryCache(ctx, cacheConfig)
        if err != nil {
            logger.Error(ctx, "ASV_CACHE_INIT_ERROR", "error", err)
            return
        }

        config := asvConfig.DefaultConfig().
            WithServerURL(bootstrap.Config.Service.AccountService.BaseURL).
            WithCredentials(&asvConfig.UserCredentials{
                ClientID: bootstrap.Config.Service.AccountService.Auth.Key,
                Password: bootstrap.Config.Service.AccountService.Auth.Secret,
            })

        asvClient, err = accountService.NewClient(ctx,
            accountService.WithCache(inMemoryCache),
            accountService.WithConfig(config),
            accountService.WithClientId("terminals"))

        if err != nil {
            logger.Error(ctx, "ASV_CLIENT_INIT_ERROR", "error", err)
            return
        }

        logger.Info(ctx, "ASV_CLIENT_INITIALIZED")
    })

    return asvClient
}

// ✅ GOOD: Service wrapper with singleton client
type AsvService interface {
    GetMerchantDocuments(ctx *gin.Context, merchantID string) ([]*dto.Document, error)
}

type asvService struct {
    client *accountService.Client
}

func NewAsvService(ctx *gin.Context) (AsvService, error) {
    client := getAsvClient(ctx)  // <-- Reuses singleton
    if client == nil {
        if gin.Mode() != gin.TestMode {
            return nil, errors.New("ASV client initialization failed")
        }
    }
    return &asvService{client: client}, nil
}

func (s *asvService) GetMerchantDocuments(ctx *gin.Context, merchantID string) ([]*dto.Document, error) {
    req := &dto.GetAccountByIDRequest{
        Id:    merchantID,
        Paths: []string{"documents.id", "documents.file_store_id"},
    }

    response, err := s.client.Account().GetByID(ctx, req)
    if err != nil {
        return nil, err
    }

    return response.Documents, nil
}
```

**Razorpay Standard**:
- **One client per service**: Use singleton pattern
- **Initialization**: `sync.Once` for thread-safe lazy init
- **Reuse**: Service wrapper gets client from singleton getter
- **Testing**: Allow nil client in test mode

**Benefits**:
- Avoids repeated client initialization overhead
- Shares single gRPC connection pool
- Reuses cache across requests
- Thread-safe initialization

**Rationale**: Creating new clients per request wastes memory and connections.

---

## Check #10: gRPC Connection Tuning (Medium)

**Problem**: Default timeouts cause slow connections and poor recovery.

**Detection Strategy**:
```bash
# Find client initialization
grep -E "NewClient|accountService\.NewClient" <pr_files>

# Check for gRPC config
grep -E "GrpcDialOptions|ConnectParams|KeepaliveClientParameters" <pr_files>
```

**What to Flag**:
```go
// ⚠️ SUBOPTIMAL: Using default gRPC settings
asvClient, _ := accountService.NewClient(ctx,
    accountService.WithConfig(config))
// Uses defaults: 120s MaxDelay, no keepalive, etc.

// ❌ BAD: Too aggressive timeouts (may fail unnecessarily)
grpcConfig := config.GrpcDialOptions{
    ConnectParams: &grpc.ConnectParams{
        Backoff: backoff.Config{
            MaxDelay: 1 * time.Second,  // <-- Too low!
        },
        MinConnectTimeout: 1 * time.Second,  // <-- Too low!
    },
}
```

**How to Fix**:
```go
// ✅ GOOD: Optimized gRPC connection config
grpcClientConfig := config.GrpcDialOptions{
    ConnectParams: &grpc.ConnectParams{
        Backoff: backoff.Config{
            BaseDelay:  1 * time.Second,  // Initial delay
            Multiplier: 1.6,               // Exponential multiplier
            Jitter:     0.2,               // Random jitter
            MaxDelay:   15 * time.Second,  // Max backoff (default: 120s)
        },
        MinConnectTimeout: 30 * time.Second,  // Min time for connection
    },
    KeepaliveClientParameters: &keepalive.ClientParameters{
        Time:                20 * time.Second,  // Ping interval
        Timeout:             10 * time.Second,  // Ping timeout
        PermitWithoutStream: false,             // Don't ping without active RPCs
    },
}

asvConfig := asvConfig.DefaultConfig().
    WithServerURL(baseURL).
    WithGrpcDialOptions(grpcClientConfig).  // <-- Use optimized config
    WithCredentials(credentials)

asvClient, err := accountService.NewClient(ctx,
    accountService.WithConfig(asvConfig),
    accountService.WithClientId("terminals"))

// ✅ GOOD: Environment-specific tuning
var maxDelay time.Duration
var keepaliveTime time.Duration

if config.Env == "production" {
    maxDelay = 15 * time.Second       // Production: faster recovery
    keepaliveTime = 20 * time.Second  // More frequent pings
} else {
    maxDelay = 30 * time.Second       // Dev: allow slower recovery
    keepaliveTime = 60 * time.Second  // Less frequent pings
}

grpcConfig := config.GrpcDialOptions{
    ConnectParams: &grpc.ConnectParams{
        Backoff: backoff.Config{
            BaseDelay:  1 * time.Second,
            Multiplier: 1.6,
            Jitter:     0.2,
            MaxDelay:   maxDelay,
        },
        MinConnectTimeout: 30 * time.Second,
    },
    KeepaliveClientParameters: &keepalive.ClientParameters{
        Time:    keepaliveTime,
        Timeout: 10 * time.Second,
        PermitWithoutStream: false,
    },
}
```

**Razorpay Standard - Recommended Settings**:
- **BaseDelay**: 1s (initial backoff)
- **Multiplier**: 1.6 (exponential growth)
- **Jitter**: 0.2 (20% randomness)
- **MaxDelay**: 15s (prod), 30s (dev) - default is 120s!
- **MinConnectTimeout**: 30s
- **Keepalive Time**: 20s (prod), 60s (dev)
- **Keepalive Timeout**: 10s
- **PermitWithoutStream**: false

**Rationale**: Default 120s MaxDelay is too slow for production recovery needs.

---

## Integration Instructions

### When to Load This File
Load when PR contains any of:
- `import.*account-service` or `github.com/razorpay/goutils/account-service`
- `accountService.NewClient` or `asv.NewClient`
- `GetByID`, `Write().Save`, `Write().Delete`
- `dto.Account`, `dto.Document`, `dto.Stakeholder`

### Progressive Loading
Only load if ASV-related code changes detected. Defer until actual ASV usage confirmed.

### File Pattern Detection

```bash
# Detect ASV usage
grep -r "account-service" go.mod
grep -r "accountService\.NewClient\|asv\.NewClient" --include="*.go"
grep -r "dto\.GetAccountByIDRequest\|dto\.SaveRequest" --include="*.go"

# Check config
grep -r "\[.*accountService\]|\[.*asv\]" --include="*.toml"
```

### Common Issue Patterns

| Issue | Checks | Severity |
|-------|--------|----------|
| Missing config | #1 | Critical |
| No Paths field | #2 | Critical |
| Ignored errors | #3 | Critical |
| Unauthorized writes | #4 | Critical |
| Wrong FileStoreId | #5 | Critical |
| Invalid URLs | #6 | High |
| No panic recovery | #7 | High |
| No cache | #8 | High |
| Per-request client | #9 | High |
| Default timeouts | #10 | Medium |

### Environment-Specific Hosts

| Environment | ASV gRPC Host |
|-------------|---------------|
| Dev | `asv.grpc.int.dev.razorpay.in:443` |
| Stage | `asv.grpc.int.stage.razorpay.in:443` |
| Production | `asv.grpc.razorpay.com:443` |

### Reference Links
- SDK: `github.com/razorpay/goutils/account-service/v2`
- Documentation: https://idocs.razorpay.com/platform/identity-account-safety/asv
- API Docs: https://idocs.razorpay.com/openapi/account-service/master
- Support: #platform_account_service on Slack, @acct-svc-devs
