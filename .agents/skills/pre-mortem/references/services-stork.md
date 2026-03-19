# Services: Stork Integration (12 Checks)

Validates Stork notification service integration patterns. Stork is Razorpay's unified notification platform for SMS, Email, WhatsApp, and Push Notifications. Focuses on correct config setup and validates new integrations are properly registered.

**Total Checks**: 12 (6 Critical, 4 High, 2 Medium)

---

## Check #1: Service Registration (Critical - New Integration Only)

**Problem**: Code uses Stork but service isn't registered, causing all API calls to fail with 403 Forbidden.

**Detection Strategy**:
1. Detect new Stork integration:
   ```bash
   # Check if this is first time using Stork in this repo
   git log --all --oneline -- . | grep -i "stork" | wc -l
   # If 0 results → likely new integration

   # Or check if Stork client initialization is new in PR
   git diff main..HEAD | grep -E "stork\.NewClient|NewStorkClient"
   ```

2. For new integrations, verify registration evidence:
   ```bash
   # Look for PR link or comment mentioning registration
   grep -E "#platform-spine-stork|@spine-stork|stork.*registered" <pr_description>

   # Check if deployment config has credentials
   grep -E "STORK_KEY|STORK_SECRET|STORK_USERNAME|STORK_PASSWORD" kubernetes/*.yaml .env* configs/*.toml
   ```

**What to Flag**:
```go
// ❌ BAD: New Stork integration without registration
// No evidence of:
// - #platform-spine-stork contact
// - Credentials in kubestash
// - Service added to stork/internal/config/config.go

config := stork.Config{
    BaseUrl: os.Getenv("STORK_ENDPOINT"),
    Service: "new-service",  // <-- Not registered!
    Auth: stork.Auth{
        Key:    os.Getenv("STORK_KEY"),    // <-- Won't exist!
        Secret: os.Getenv("STORK_SECRET"),  // <-- Won't exist!
    },
}
```

**How to Fix**:
```
✅ GOOD: Proper registration flow

Before merging this PR:

1. Contact @spine-stork on #platform-spine-stork Slack:
   - Request: "Need to integrate Stork for [service-name]"
   - Provide: Service name, owner, use case

2. Wait for credentials:
   - They will add service to stork/internal/config/config.go
   - Receive username/password for HTTP Basic Auth
   - Credentials added to Stork's kubestash

3. Add credentials to your service's secrets:
   - Kubestash: {service}/stork-key, {service}/stork-secret
   - Or Vault: vault/{service}/stork/credentials

4. Deploy config with environment variables:
   - STORK_ENDPOINT
   - STORK_KEY
   - STORK_SECRET

Reference: https://idocs.razorpay.com/platform/stork/integrate-stork/sms/#create-a-stork-client
```

**Razorpay Standard**:
- Service registration is **mandatory** before Stork can be used
- 80+ services are already registered in `stork/internal/config/config.go`
- Unregistered services will get `403 Forbidden` on all API calls
- Credentials stored in kubestash, never in code

**Rationale**: Unregistered services cannot authenticate with Stork, causing complete integration failure.

---

## Check #2: Environment Configuration (Critical)

**Problem**: Missing or incorrect Stork environment variables cause runtime failures.

**Detection Strategy**:
```bash
# Find Stork client initialization
grep -E "stork\.NewClient|stork\.Config" <pr_files>

# Check if environment variables are validated
grep -A 10 "stork\.Config" <pr_files> | grep -E "Getenv|env\[|config\."

# Verify deployment configs have required variables
grep -E "STORK_ENDPOINT|STORK_BASE_URL" kubernetes/*.yaml .env* configs/*.toml
grep -E "STORK_KEY|STORK_USERNAME" kubernetes/*.yaml .env* configs/*.toml
grep -E "STORK_SECRET|STORK_PASSWORD" kubernetes/*.yaml .env* configs/*.toml
```

**What to Flag**:
```go
// ❌ BAD: No validation of environment variables
config := stork.Config{
    BaseUrl: os.Getenv("STORK_ENDPOINT"),  // <-- Could be empty!
    Service: "api",
    Auth: stork.Auth{
        Key:    os.Getenv("STORK_KEY"),
        Secret: os.Getenv("STORK_SECRET"),
    },
}
client, _ := stork.NewClient(config, nil)  // <-- Will fail if env vars missing

// ❌ BAD: Wrong environment URL
config := stork.Config{
    BaseUrl: "https://stork.dev.razorpay.in",  // <-- Dev URL in prod code!
}

// ❌ BAD: Hardcoded credentials (security issue!)
config := stork.Config{
    Auth: stork.Auth{
        Key:    "hardcoded_key",  // <-- NEVER DO THIS!
        Secret: "hardcoded_secret",
    },
}
```

**How to Fix**:
```go
// ✅ GOOD: Validate environment variables at startup
func NewStorkClient() (*stork.Client, error) {
    endpoint := os.Getenv("STORK_ENDPOINT")
    key := os.Getenv("STORK_KEY")
    secret := os.Getenv("STORK_SECRET")
    service := os.Getenv("STORK_SERVICE_NAME")

    if endpoint == "" || key == "" || secret == "" || service == "" {
        return nil, errors.New("missing required Stork configuration")
    }

    config := stork.Config{
        BaseUrl: endpoint,
        Service: service,
        Auth: stork.Auth{
            Key:    key,
            Secret: secret,
        },
    }

    return stork.NewClient(config, nil)
}

// ✅ GOOD: Deployment config with correct URLs
// kubernetes/deployment.yaml or configs/prod.toml
env:
  - name: STORK_ENDPOINT
    value: "https://stork.razorpay.com"  # Prod
  - name: STORK_KEY
    valueFrom:
      secretKeyRef:
        name: stork-credentials
        key: username
  - name: STORK_SECRET
    valueFrom:
      secretKeyRef:
        name: stork-credentials
        key: password
```

**Razorpay Standard**:
- **Required variables**: `STORK_ENDPOINT`, `STORK_KEY`, `STORK_SECRET`, `STORK_SERVICE_NAME`
- **Environment URLs**:
  - Stage: `https://stork.dev.razorpay.in`
  - Dev-serve: `https://stork-{label}.dev.razorpay.in`
  - Prod: `https://stork.razorpay.com`
- **Credentials**: Always from Vault/kubestash/K8s secrets, never hardcoded
- **Validation**: Check at service startup, fail fast if missing

**Rationale**: Missing or incorrect config causes all Stork calls to fail silently or with auth errors.

---

## Check #3: Template Existence & Naming (Critical)

**Problem**: Code references templates that don't exist or have incorrect names, causing notifications to fail.

**Detection Strategy**:
1. Extract template references from code:
   ```bash
   # Find template_name usage
   grep -E 'TemplateName:\s*"[^"]+"' <pr_files>
   grep -E 'template_name":\s*"[^"]+"' <pr_files>

   # Find template_namespace usage
   grep -E 'TemplateNamespace:\s*"[^"]+"' <pr_files>
   grep -E 'template_namespace":\s*"[^"]+"' <pr_files>
   ```

2. Check if templates are documented or exist in templating service:
   ```bash
   # Look for template registration PR links
   grep -E "raven.*PR|template.*created|DLT.*approved" <pr_description>
   ```

**What to Flag**:
```go
// ❌ BAD: Template name likely doesn't exist (typo or not created)
req := &stork.SMSRequest{
    TemplateName:      "sms.otp.login",  // <-- Check if exists!
    TemplateNamespace: "authentication",
    // ... other fields
}

// ❌ BAD: Namespace mismatch
req := &stork.SMSRequest{
    TemplateName:      "payment_success",  // Template in "payments" namespace
    TemplateNamespace: "transactions",     // <-- Wrong namespace!
}

// ❌ BAD: WhatsApp V2 without template_name (using deprecated `text` field)
req := map[string]interface{}{
    "whatsapp_channels": []map[string]interface{}{
        {
            "destination": "+919876543210",
            "text":        "Hello User",  // <-- NOT SUPPORTED in V2!
        },
    },
}
```

**How to Fix**:
```go
// ✅ GOOD: Template exists and name/namespace match
// Template created via PR: https://github.com/razorpay/raven/pull/335
req := &stork.SMSRequest{
    TemplateName:      "sms.otp.login",  // <-- Exists in templating service
    TemplateNamespace: "authentication",  // <-- Correct namespace
    ContentParams: map[string]string{
        "otp":      "123456",  // <-- Matches template variables
        "validity": "5 minutes",
    },
}

// ✅ GOOD: WhatsApp V2 with template_name
req := map[string]interface{}{
    "whatsapp_channels": []map[string]interface{}{
        {
            "destination":   "+919876543210",
            "template_name": "welcome_message",  // <-- Required for V2
            "params": map[string]interface{}{
                "template_data": map[string]interface{}{
                    "variables": map[string]string{
                        "name": "Merchant Name",
                    },
                },
            },
        },
    },
}
```

**Razorpay Standard**:
- **SMS**: Templates must be registered in Templating Service + DLT platform (India)
- **Email**: Templates in Templating Service (migrate from blade templates)
- **WhatsApp**: Templates approved by Meta via vendor dashboard (Gupshup/Interakt)
- **Push**: No template needed (raw FCM payload)
- **Template naming**: Use dot notation (e.g., `sms.payments.success`, `email.merchant.onboarding`)
- **Context keys**: Must be whitelisted in Stork (raise PR to add new keys)

**Rationale**: Non-existent templates cause notification failures with `not_found` errors.

---

## Check #4: WhatsApp Opt-In Requirement (Critical)

**Problem**: WhatsApp messages sent to non-opted-in users are silently dropped by Meta.

**Detection Strategy**:
```bash
# Find WhatsApp send calls
grep -E "SendWhatsApp|whatsapp_channels" <pr_files>

# Check if opt-in is called first
grep -B 20 "SendWhatsApp" <pr_files> | grep -E "OptInUser|opt.*in"
```

**What to Flag**:
```go
// ❌ BAD: Sending WhatsApp without opt-in check
func NotifyMerchant(ctx context.Context, merchantID, phone string) {
    req := map[string]interface{}{
        "whatsapp_channels": []map[string]interface{}{
            {
                "destination":   phone,  // <-- User may not be opted in!
                "template_name": "payment_success",
            },
        },
    }

    client.SendWhatsApp(ctx, req)  // <-- Message silently dropped!
}

// ❌ BAD: Opt-in called but error ignored
_, err := client.OptInUser(ctx, phone)
// No error check - continues even if opt-in failed!
client.SendWhatsApp(ctx, req)
```

**How to Fix**:
```go
// ✅ GOOD: Opt-in user before sending WhatsApp
func NotifyMerchant(ctx context.Context, merchantID, phone string) error {
    // Step 1: Opt-in user for WhatsApp
    optInReq := map[string]interface{}{
        "owner_id":      merchantID,
        "owner_type":    "merchant",
        "phone_numbers": []string{phone},
        "service":       "api",
    }

    _, err := client.OptInUser(ctx, optInReq)
    if err != nil {
        return fmt.Errorf("whatsapp opt-in failed: %w", err)
    }

    // Step 2: Send WhatsApp message
    req := map[string]interface{}{
        "whatsapp_channels": []map[string]interface{}{
            {
                "destination":   phone,
                "template_name": "payment_success",
            },
        },
    }

    _, err = client.SendWhatsApp(ctx, req)
    return err
}

// ✅ BETTER: Check if user is already opted in (avoid redundant opt-in)
func NotifyMerchant(ctx context.Context, merchantID, phone string) error {
    // Check if user is opted in (cache this in your service)
    if !isUserOptedIn(merchantID, phone) {
        _, err := client.OptInUser(ctx, optInReq)
        if err != nil {
            return fmt.Errorf("whatsapp opt-in failed: %w", err)
        }
        cacheOptInStatus(merchantID, phone, true)
    }

    // Send message
    _, err := client.SendWhatsApp(ctx, req)
    return err
}
```

**Razorpay Standard**:
- **Always opt-in before first WhatsApp message** to a user
- Opt-in is per phone number, not per message
- Cache opt-in status to avoid redundant API calls
- Handle opt-out requests (GDPR compliance)
- Silently dropped messages don't return errors from Stork

**Rationale**: Meta requires explicit user consent for WhatsApp messaging. Non-opted-in users won't receive messages.

---

## Check #5: Attachment Presigned URL Handling (High)

**Problem**: Presigned URLs expire in 15 minutes; delays cause attachment upload failures.

**Detection Strategy**:
```bash
# Find presigned URL usage
grep -E "GetPresignedUrl|GetPresignedURL|presigned_url" <pr_files>

# Check if upload happens immediately after getting URL
grep -A 20 "GetPresignedUrl" <pr_files> | grep -E "http\.Put|PUT|upload"
```

**What to Flag**:
```go
// ❌ BAD: Getting presigned URL but uploading much later
func SendEmailWithAttachment(ctx context.Context, attachmentPath string) {
    // Get presigned URL
    urlResp, _ := client.GetPresignedUrl(ctx, &stork.PresignedUrlRequest{
        Channel:  "email",
        UrlCount: 1,
    })
    presignedUrl := urlResp.AttachmentResponses[0].PresignedUrl
    fileId := urlResp.AttachmentResponses[0].FileId

    // ❌ Do other work (time passes...)
    time.Sleep(20 * time.Minute)  // <-- URL expires after 15 min!

    // Try to upload - WILL FAIL!
    uploadFile(presignedUrl, attachmentPath)

    // Send email with attachment
    client.SendEmail(ctx, emailReq)
}

// ❌ BAD: No error handling on presigned URL request
urlResp, _ := client.GetPresignedUrl(ctx, req)  // <-- Ignores error!
fileId := urlResp.AttachmentResponses[0].FileId  // <-- Panic if urlResp is nil!
```

**How to Fix**:
```go
// ✅ GOOD: Upload immediately after getting presigned URL
func SendEmailWithAttachment(ctx context.Context, attachmentPath string) error {
    // Step 1: Get presigned URL
    urlResp, err := client.GetPresignedUrl(ctx, &stork.PresignedUrlRequest{
        Channel:   "email",
        OwnerId:   merchantID,
        OwnerType: "merchant",
        Service:   "api",
        UrlCount:  1,
    })
    if err != nil {
        return fmt.Errorf("failed to get presigned URL: %w", err)
    }

    presignedUrl := urlResp.AttachmentResponses[0].PresignedUrl
    fileId := urlResp.AttachmentResponses[0].FileId

    // Step 2: Upload IMMEDIATELY (within 15-min window)
    err = uploadFileToS3(ctx, presignedUrl, attachmentPath)
    if err != nil {
        return fmt.Errorf("failed to upload attachment: %w", err)
    }

    // Step 3: Send email with file_id
    emailReq := &stork.EmailRequest{
        // ... other fields
        Attachments: []stork.Attachment{
            {
                FileId:      fileId,  // <-- Reference uploaded file
                DisplayName: "Invoice.pdf",
                Extension:   "pdf",
            },
        },
    }

    _, err = client.SendEmail(ctx, emailReq)
    return err
}

// Helper: Upload file via presigned URL
func uploadFileToS3(ctx context.Context, presignedUrl, filePath string) error {
    fileData, err := os.ReadFile(filePath)
    if err != nil {
        return err
    }

    req, _ := http.NewRequestWithContext(ctx, "PUT", presignedUrl, bytes.NewReader(fileData))
    req.Header.Set("Content-Type", "application/pdf")

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return fmt.Errorf("upload failed with status %d", resp.StatusCode)
    }

    return nil
}
```

**Razorpay Standard**:
- **Upload window**: 15 minutes from presigned URL generation
- **Max attachment size**:
  - Email: 10 MB total
  - WhatsApp Images: 5 MB
  - WhatsApp Documents/Videos: 10 MB
- **Upload method**: HTTP PUT with file contents
- **Error handling**: Check upload response status code

**Rationale**: Expired presigned URLs cause silent upload failures, leading to emails/WhatsApp without attachments.

---

## Check #6: Error Handling & Retry Logic (High)

**Problem**: No retry logic for transient Stork failures causes notification delivery failures.

**Detection Strategy**:
```bash
# Find Stork API calls
grep -E "SendSMS|SendEmail|SendWhatsApp|SendPushNotification" <pr_files>

# Check if errors are handled
grep -A 5 "Send" <pr_files> | grep "if err"

# Check for retry logic
grep -E "retry|Retry|backoff" <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: No error handling
resp, _ := client.SendSMS(ctx, req)  // <-- Ignores errors!
log.Info("SMS sent", "message_id", resp.MessageID)

// ❌ BAD: Error logged but not retried for transient failures
resp, err := client.SendSMS(ctx, req)
if err != nil {
    log.Error("SMS failed", "error", err)  // <-- No retry!
    return err
}

// ❌ BAD: Retrying non-retryable errors (400, 404)
for i := 0; i < 3; i++ {
    resp, err := client.SendSMS(ctx, req)
    if err == nil {
        break
    }
    time.Sleep(time.Second)  // <-- Retries even for bad requests!
}
```

**How to Fix**:
```go
// ✅ GOOD: Proper error handling with retry for transient failures
func SendSMSWithRetry(ctx context.Context, req *stork.SMSRequest) error {
    maxRetries := 3
    backoff := time.Second

    for attempt := 0; attempt < maxRetries; attempt++ {
        resp, err := client.SendSMS(ctx, req)
        if err == nil {
            log.Info("SMS sent", "message_id", resp.MessageID)
            return nil
        }

        // Check if error is retryable
        if !isRetryable(err) {
            return fmt.Errorf("permanent error: %w", err)
        }

        // Retry transient errors with exponential backoff
        if attempt < maxRetries-1 {
            jitter := time.Duration(rand.Intn(500)) * time.Millisecond
            sleepTime := backoff + jitter
            log.Warn("SMS failed, retrying", "attempt", attempt+1, "backoff", sleepTime)
            time.Sleep(sleepTime)
            backoff *= 2  // Exponential backoff
        }
    }

    return fmt.Errorf("SMS failed after %d attempts", maxRetries)
}

// Helper: Determine if error is retryable
func isRetryable(err error) bool {
    // Retry on 429 (rate limit), 500 (internal), 503 (unavailable)
    if strings.Contains(err.Error(), "resource_exhausted") {
        return true  // 429
    }
    if strings.Contains(err.Error(), "internal") {
        return true  // 500
    }
    if strings.Contains(err.Error(), "unavailable") {
        return true  // 503
    }

    // Don't retry 400 (invalid_argument), 403 (permission_denied), 404 (not_found)
    return false
}
```

**Razorpay Standard**:
- **Retry strategy**:
  - Transient errors (429, 500, 503): Max 3 retries
  - Exponential backoff: 1s, 2s, 4s + random jitter (0-500ms)
  - Permanent errors (400, 403, 404): No retry
- **Circuit breaker**: If >50% failures in 60s, open for 30s
- **Timeout**: 30s per API call
- **Error codes**:
  - `invalid_argument` (400): Fix request
  - `not_found` (404): Template doesn't exist
  - `permission_denied` (403): Auth failed
  - `resource_exhausted` (429): Rate limit
  - `internal` (500), `unavailable` (503): Retry

**Rationale**: Transient failures without retry cause unnecessary notification delivery failures.

---

## Check #7: Rate Limiting Awareness (High)

**Problem**: Code doesn't respect rate limits, causing throttling and message failures.

**Detection Strategy**:
```bash
# Find loops or bulk sending
grep -B 5 -A 10 "SendSMS\|SendEmail\|SendWhatsApp" <pr_files> | grep -E "for.*range|for i :=|while"

# Check for rate limiting handling
grep -E "429|resource_exhausted|rate.*limit|throttl" <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Sending SMS in tight loop without rate limiting
for _, merchant := range merchants {
    req := &stork.SMSRequest{
        Destination: merchant.Phone,
        // ... other fields
    }
    client.SendSMS(ctx, req)  // <-- Will hit rate limit!
}

// ❌ BAD: Sending duplicate SMS to same user
for i := 0; i < 10; i++ {
    req := &stork.SMSRequest{
        Destination:  "+919876543210",  // <-- Same user!
        TemplateName: "sms.otp.login",
    }
    client.SendSMS(ctx, req)  // <-- Hits 5 SMS per 30-min limit!
}
```

**How to Fix**:
```go
// ✅ GOOD: Batch sending with rate limiting
func SendBulkSMS(ctx context.Context, merchants []Merchant) error {
    // Rate limit: 5 SMS per 30-min per service+receiver
    // To avoid hitting limit, add delay between sends

    rateLimiter := time.NewTicker(time.Second)  // 1 SMS per second
    defer rateLimiter.Stop()

    for _, merchant := range merchants {
        <-rateLimiter.C  // Wait for rate limiter

        req := &stork.SMSRequest{
            OwnerId:           merchant.ID,
            OwnerType:         "merchant",
            Destination:       merchant.Phone,
            TemplateName:      "sms.payment.reminder",
            TemplateNamespace: "payments",
        }

        err := SendSMSWithRetry(ctx, req)
        if err != nil {
            log.Error("SMS failed", "merchant_id", merchant.ID, "error", err)
            // Continue with other merchants
        }
    }

    return nil
}

// ✅ GOOD: Deduplicate recipients before sending
func SendBulkSMS(ctx context.Context, phones []string) {
    // Deduplicate to avoid rate limit
    uniquePhones := make(map[string]bool)
    for _, phone := range phones {
        uniquePhones[phone] = true
    }

    for phone := range uniquePhones {
        // Send SMS...
    }
}
```

**Razorpay Standard - Rate Limits**:
- **SMS**: 5 messages per 30-min window per service+receiver combination
- **SMS International**: 250 messages per 24h per owner_id+template combination
- **WhatsApp**: Per-merchant limits (enforced by Meta policy)
- **Push Notifications**:
  - General: 1 request/second (CleverTap limit)
  - Targeted: Max 1000 users per request
- **Email**: No explicit limit, but avoid spam patterns

**Handling 429 Errors**:
```go
err := client.SendSMS(ctx, req)
if strings.Contains(err.Error(), "resource_exhausted") {
    // Exponential backoff for rate limit
    time.Sleep(30 * time.Second)
    // Retry...
}
```

**Rationale**: Exceeding rate limits causes message failures and potential service blocking.

---

## Check #8: DLT Registration for India SMS (Critical - SMS Only)

**Problem**: SMS sent without DLT-registered templates/sender IDs are rejected by operators in India.

**Detection Strategy**:
```bash
# Find SMS sending code
grep -E "SendSMS|SMSRequest" <pr_files>

# Check if template/sender ID are documented as DLT-approved
grep -E "DLT.*approved|DLT.*registered|sender.*id" <pr_description> <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Using non-DLT-registered sender ID
req := &stork.SMSRequest{
    Sender:       "MYAPP",  // <-- Not registered in DLT!
    Destination:  "+919876543210",
    TemplateName: "sms.otp.login",
}

// ❌ BAD: Custom template not registered in DLT
req := &stork.SMSRequest{
    Sender:            "RZRPAY",
    TemplateName:      "sms.new.feature",  // <-- Not in DLT!
    TemplateNamespace: "beta-testing",
}
```

**How to Fix**:
```
✅ GOOD: DLT registration workflow

For new SMS templates or sender IDs in India:

1. Register template in Templating Service:
   - Create PR to razorpay/raven repo
   - Add template with Mustache placeholders
   - Example: https://github.com/razorpay/raven/pull/335

2. Contact @spine-stork before merging:
   - Tag @Gorang (Gorang Verma)
   - Request DLT registration for template

3. Wait for DLT approval:
   - Stork team registers with DLT platform
   - Approval takes 2-5 business days
   - DO NOT merge PR until approved

4. For sender IDs:
   - Create PR: https://github.com/razorpay/raven/pull/269
   - Same approval flow as templates
   - Cannot use until DLT-approved

5. Deployment:
   - After DLT approval, merge PR
   - Deploy Raven changes
   - Start using template in your code

✅ Use existing DLT-approved sender IDs:
req := &stork.SMSRequest{
    Sender:            "RZRPAY",  // <-- Already DLT-approved
    TemplateName:      "sms.otp.login",  // <-- Already in system
    TemplateNamespace: "authentication",
}
```

**Razorpay Standard**:
- **TRAI compliance**: Mandatory for all India SMS
- **DLT platform**: Register sender IDs and templates before use
- **Sender ID format**: 6 alphanumeric characters (e.g., `RZRPAY`, `RZRPAYX`)
- **Template ID**: Unique identifier from DLT platform
- **Process owner**: @spine-stork team handles DLT registration
- **Timeline**: 2-5 business days for DLT approval

**International SMS**: DLT not required, but follow local regulations.

**Rationale**: Non-DLT-registered SMS are rejected by operators, causing 100% delivery failure in India.

---

## Check #9: Service Name Convention (High)

**Problem**: Inconsistent or invalid service names break metrics and analytics.

**Detection Strategy**:
```bash
# Find service field in Stork calls
grep -E 'Service:\s*"[^"]+"' <pr_files>
grep -E '"service":\s*"[^"]+"' <pr_files>
```

**What to Flag**:
```go
// ❌ BAD: Invalid service name (not following convention)
req := &stork.SMSRequest{
    Service: "my-new-service",  // <-- Not in standard list!
}

// ❌ BAD: Using prod service name in test/stage code
// In staging environment:
req := &stork.SMSRequest{
    Service: "api-live",  // <-- Should use "api-test" in stage!
}

// ❌ BAD: Empty or missing service name
req := &stork.SMSRequest{
    Service: "",  // <-- Will break metrics!
}
```

**How to Fix**:
```go
// ✅ GOOD: Use environment-specific service names
func getStorkServiceName() string {
    env := os.Getenv("ENVIRONMENT")  // "staging" or "production"

    switch env {
    case "production":
        return "api-live"
    case "staging":
        return "api-test"
    default:
        return "api-test"
    }
}

req := &stork.SMSRequest{
    Service: getStorkServiceName(),  // <-- Environment-aware
    // ... other fields
}

// ✅ GOOD: Use existing standard service names
// Standard names (80+ services registered):
// - api-test, api-live
// - rx-test, rx-live
// - opfin
// - payment-links
// - mandate-hq
// - settlements

req := &stork.SMSRequest{
    Service: "api-live",  // <-- Follow existing convention
}
```

**Razorpay Standard - Service Names**:
- **Use existing names** where applicable
- **Naming convention**: `{product}-{environment}`
  - Examples: `api-live`, `api-test`, `rx-live`, `opfin`
- **Avoid**: Custom names unless necessary
- **New service names**: Discuss with @spine-stork first
- **Metrics impact**: Service name is primary dimension in analytics

**Rationale**: Inconsistent service names break metrics dashboards and make debugging difficult.

---

## Check #10: Context Keys Whitelisting (Medium)

**Problem**: Custom context keys aren't whitelisted in Stork, causing them to be silently dropped.

**Detection Strategy**:
```bash
# Find context field usage
grep -E 'Context:\s*map|context":\s*{' <pr_files>

# Extract context keys
grep -A 10 "Context:" <pr_files> | grep -E '"[^"]+"\s*:'
```

**What to Flag**:
```go
// ❌ BAD: Custom context keys not whitelisted
req := &stork.SMSRequest{
    // ... other fields
    Context: map[string]interface{}{
        "campaign_id":    "camp_123",  // <-- May not be whitelisted!
        "custom_metric":  "value",     // <-- Likely dropped!
        "new_field":      "data",      // <-- Not whitelisted!
    },
}

// ❌ BAD: Sensitive data in context (PII)
req := &stork.SMSRequest{
    Context: map[string]interface{}{
        "email":      "user@example.com",  // <-- PII in analytics!
        "phone":      "+919876543210",      // <-- PII!
        "card_last4": "1234",               // <-- Sensitive!
    },
}
```

**How to Fix**:
```go
// ✅ GOOD: Use whitelisted context keys
req := &stork.SMSRequest{
    // ... other fields
    Context: map[string]interface{}{
        // Common whitelisted keys:
        "batch_id":     "batch_123",
        "org_id":       "org_100000Razorpay",
        "use_case":     "payment_reminder",
        "application":  "merchant_dashboard",
    },
}

// ✅ GOOD: Add new keys via PR if needed
// 1. Create PR to Stork repo
// 2. Example: https://github.com/razorpay/stork/pull/481
// 3. Add key to context whitelist
// 4. Wait for merge and deployment
// 5. Use new key in your code

// ✅ GOOD: Don't include PII in context
req := &stork.SMSRequest{
    Context: map[string]interface{}{
        "merchant_id":  merchantID,  // <-- OK (identifier, not PII)
        "campaign_id":  campaignID,  // <-- OK
        // NO: email, phone, name, card details
    },
}
```

**Razorpay Standard**:
- **Whitelisting required**: Custom context keys must be added to Stork's whitelist
- **Common whitelisted keys**:
  - `batch_id`, `org_id`, `use_case`
  - `application`, `campaign_id`
  - `merchant_id` (identifier, not PII)
- **PR process**: Raise PR to `razorpay/stork` to add new keys
- **No PII**: Never include email, phone, name, card details in context
- **Analytics only**: Context is for metrics/debugging, not business logic

**Rationale**: Non-whitelisted keys are silently dropped, breaking metrics tracking.

---

## Check #11: Delivery Callback Setup (Medium - Optional but Recommended)

**Problem**: No delivery status tracking configured, making debugging failures difficult.

**Detection Strategy**:
```bash
# Check if delivery callback is requested
grep -E "DeliveryCallbackRequested|delivery_callback_requested" <pr_files>

# Check if callback handling is implemented
grep -E "stork.*callback|delivery.*status|SNS.*subscription" <pr_files>
```

**What to Flag**:
```go
// ⚠️ OPTIONAL: Delivery callback not requested (makes debugging harder)
req := &stork.SMSRequest{
    // ... other fields
    // Missing: DeliveryCallbackRequested: true
}

// ❌ BAD: Callback requested but no handler setup
req := &stork.SMSRequest{
    DeliveryCallbackRequested: true,  // <-- Callback enabled
}
// But no SQS queue or SNS subscription configured to receive callbacks!
```

**How to Fix**:
```go
// ✅ GOOD: Request delivery callbacks for critical notifications
req := &stork.SMSRequest{
    OwnerId:                   merchantID,
    OwnerType:                 "merchant",
    Service:                   "api-live",
    Sender:                    "RZRPAY",
    Destination:               phone,
    TemplateName:              "sms.payment.success",
    TemplateNamespace:         "payments",
    DeliveryCallbackRequested: true,  // <-- Enable callbacks
}

// ✅ GOOD: Set up SNS subscription to receive callbacks
// 1. Create PR to razorpay/vishnu repo
// 2. Subscribe SQS queue to SNS topic:
//    - Stage: stage-stork-sms-callbacks
//    - Prod: prod-stork-sms-callbacks
// 3. Add subscription filter policy:
//    {"service": ["api-live"]}  // Only your service's events
// 4. Implement SQS consumer to handle callbacks

// Callback payload structure:
type DeliveryStatusMessage struct {
    Service         string  // "api-live"
    DeliveryChannel string  // "sms" or "email"
    ReferenceID     string  // Message ID from Stork
    Status          string  // "delivered", "failed", "bounced"
    DeliveryTime    int64   // Unix timestamp
    StatusCode      string  // "Delivery", "Bounced", "Reject", etc.
}
```

**Razorpay Standard**:
- **SNS Topics**:
  - Stage: `stage-stork-sms-callbacks`, `stage-stork-email-callbacks`
  - Prod: `prod-stork-sms-callbacks`, `prod-stork-email-callbacks`
- **Subscription**: Create PR to `razorpay/vishnu` to subscribe SQS queue
- **Filter policy**: Filter by `service` field to get only your events
- **Use cases**:
  - Track delivery success rate
  - Alert on delivery failures
  - Debug bounce/rejection reasons
  - Compliance (delivery proof)

**Rationale**: Delivery callbacks enable proactive monitoring and debugging of notification failures.

---

## Check #12: Mock Mode for Testing (High)

**Problem**: Integration tests hit real Stork API, causing test flakiness and SMS/email spam.

**Detection Strategy**:
```bash
# Find Stork client initialization in tests
grep -E "stork\.NewClient" **/*_test.go

# Check if mock mode is used
grep -E "Mock.*true|MockClient" **/*_test.go
```

**What to Flag**:
```go
// ❌ BAD: Test hits real Stork API
// File: notification_test.go
func TestSendSMS(t *testing.T) {
    config := stork.Config{
        BaseUrl: "https://stork.razorpay.com",  // <-- Real API!
        Service: "api-test",
        Auth: stork.Auth{
            Key:    os.Getenv("STORK_KEY"),
            Secret: os.Getenv("STORK_SECRET"),
        },
    }
    client, _ := stork.NewClient(config, nil)

    req := &stork.SMSRequest{
        Destination: "+919876543210",  // <-- Real phone number!
        // ...
    }
    resp, err := client.SendSMS(context.Background(), req)
    // Test will send real SMS!
}
```

**How to Fix**:
```go
// ✅ GOOD: Use mock mode in tests
// File: notification_test.go
func TestSendSMS(t *testing.T) {
    config := stork.Config{
        BaseUrl: "https://stork.dev.razorpay.in",
        Service: "api-test",
        Auth: stork.Auth{
            Key:    "test-key",
            Secret: "test-secret",
        },
        Mock: true,  // <-- Enable mock mode!
    }
    client, err := stork.NewClient(config, nil)
    require.NoError(t, err)

    req := &stork.SMSRequest{
        Destination: "+919876543210",
        TemplateName: "sms.otp.login",
        // ...
    }

    // Mock client returns success without calling API
    resp, err := client.SendSMS(context.Background(), req)
    require.NoError(t, err)
    assert.NotEmpty(t, resp.MessageID)  // Mock returns generated ID
}

// ✅ BETTER: Use interface and mock implementation
type NotificationService interface {
    SendSMS(ctx context.Context, req *stork.SMSRequest) error
}

type StorkNotificationService struct {
    client stork.IStorkClient
}

func (s *StorkNotificationService) SendSMS(ctx context.Context, req *stork.SMSRequest) error {
    _, err := s.client.SendSMS(ctx, req)
    return err
}

// Mock implementation for tests
type MockNotificationService struct {
    SendSMSFunc func(ctx context.Context, req *stork.SMSRequest) error
}

func (m *MockNotificationService) SendSMS(ctx context.Context, req *stork.SMSRequest) error {
    if m.SendSMSFunc != nil {
        return m.SendSMSFunc(ctx, req)
    }
    return nil  // Default: no-op
}

// In tests:
func TestBusinessLogic(t *testing.T) {
    mockNotif := &MockNotificationService{
        SendSMSFunc: func(ctx context.Context, req *stork.SMSRequest) error {
            assert.Equal(t, "+919876543210", req.Destination)
            return nil
        },
    }

    // Test business logic with mock
    err := ProcessPayment(mockNotif, payment)
    require.NoError(t, err)
}
```

**Razorpay Standard**:
- **Always use mock mode** (`Mock: true`) in unit/integration tests
- **Test credentials**: Use `"test-key"` / `"test-secret"` in tests
- **No real API calls** in CI/CD pipelines
- **Interface pattern**: Abstract Stork client behind interface for better testing
- **Stage environment**: Use for manual/E2E testing only (with test phone numbers)

**Rationale**: Real API calls in tests cause flakiness, cost, and spam (real SMS/emails sent).

---

## Integration Instructions

### When to Load This File
Load when PR contains any of:
- `import.*stork` or `github.com/razorpay/goutils/stork`
- `stork.NewClient` or `stork.Config`
- `SendSMS`, `SendEmail`, `SendWhatsApp`, `SendPushNotification`
- New notification integration

### Progressive Loading
Only load if Stork-related code changes detected. Defer until actual Stork usage confirmed.

### Check Priority by Integration State

**New Integration (service never used Stork before)**:
- Run Check #1 FIRST (Service Registration) - CRITICAL
- Then run all other checks

**Existing Integration**:
- Skip Check #1 (already registered)
- Focus on #2-#12

### File Pattern Detection

```bash
# Detect Stork usage
grep -r "stork\\.NewClient" --include="*.go"
grep -r "github.com/razorpay/goutils/stork" --include="*.go"

# Detect new integration (Check #1)
git log --all --oneline | grep -i "stork" | wc -l  # 0 = new integration
```

### Common Issue Patterns

| Issue | Checks | Severity |
|-------|--------|----------|
| New integration not registered | #1 | Critical |
| Missing env vars | #2 | Critical |
| Template doesn't exist | #3 | Critical |
| WhatsApp without opt-in | #4 | Critical |
| SMS without DLT (India) | #8 | Critical |
| No retry logic | #6 | High |
| Rate limit not handled | #7 | High |
| Presigned URL expires | #5 | High |
| Wrong service name | #9 | High |
| Tests hit real API | #12 | High |
| Context keys not whitelisted | #10 | Medium |
| No delivery callbacks | #11 | Medium |
