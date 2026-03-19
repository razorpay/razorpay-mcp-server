# PG-Router Service Development Checks

Pre-mortem checks for **pg-router service** code development (not router SDK). Based on real patterns from pg-router repository (.agents/skills/repo-skill). These checks apply when developing **inside pg-router**, not when consuming router SDK.

---

## Check #1: Mutex Lock Acquisition with Defer Unlock 🚨 CRITICAL

### What to Check
Verify distributed mutex is acquired on order_id/payment_id before processing, and unlock is in defer block.

### Bad Pattern ❌
```go
// Missing mutex lock - race condition risk
func ProcessCallback(ctx context.Context, orderId string) error {
    order, err := GetOrder(ctx, orderId)
    if err != nil {
        return err
    }

    // ❌ No mutex - concurrent create + callback can conflict
    return UpdateOrderStatus(ctx, order, "authorized")
}

// Mutex unlock not deferred
callbackMutex := appContext.Context().MutexClient().New(ctx, key, value, ttl)
errMutexLock := callbackMutex.Lock(attempts, retryInterval)
if errMutexLock != nil {
    return err
}
callbackMutex.Unlock() // ❌ Not deferred - won't run if panic or early return
```

### Good Pattern ✅
```go
func ProcessCallback(ctx context.Context, orderId string) error {
    // Acquire mutex with TTL auto-expire
    key := fmt.Sprintf("callback_%s", orderId)
    callbackMutex := appContext.Context().MutexClient().New(
        ctx, key, orderId,
        commonConstants.OrderIDMutexTTL,
    )

    errMutexLock := callbackMutex.Lock(
        commonConstants.OrderIDMutexAttempts,
        commonConstants.OrderIDMutexRetryInterval,
    )

    if errMutexLock != nil {
        logger.Logger(ctx).Errorw(trace.OrderMutexLockFailed,
            map[string]interface{}{"error": errMutexLock.Error()})

        return class.ErrBadRequestError.New(
            codes.BadRequestMutexResourceAlreadyAcquired,
        )
    }

    // Defer unlock immediately after lock success
    defer func() {
        errMutexUnlock := callbackMutex.Unlock()
        if errMutexUnlock != nil {
            logger.Logger(ctx).Errorw(trace.MutexUnlockFailed,
                map[string]interface{}{"error": errMutexUnlock.Error()})
        }
    }()

    // Process callback with mutex held
    return UpdateOrderStatus(ctx, orderId, "authorized")
}
```

### Detection Strategy
1. Find functions processing payments/callbacks/orders
2. Check for `MutexClient().New()` calls with order_id/payment_id
3. Verify `defer` block with `Unlock()` immediately after `Lock()`
4. Check mutex key uses proper format: `callback_{order_id}`

### Flag Conditions
- **CRITICAL**: No mutex on concurrent operations → Race condition, duplicate processing
- **CRITICAL**: Unlock not in defer → Mutex leaked on panic/early return
- **HIGH**: Missing mutex error handling → Silent failure, no retry
- **MEDIUM**: Hardcoded mutex TTL instead of config constant

### Severity
🚨 **CRITICAL** - Race between payment create and callback causes duplicate payments or lost status updates

---

## Check #2: Payment Status Transition Validation 🚨 CRITICAL

### What to Check
Validate payment status transitions follow state machine rules (CREATED → PENDING → AUTHORIZED → CAPTURED → REFUNDED).

### Bad Pattern ❌
```go
// Direct status update without validation
func ProcessCallback(ctx context.Context, payment *Payment, callbackStatus string) error {
    // ❌ No validation - allows invalid transitions
    payment.Status = callbackStatus
    return payment.Save(ctx)
}

// Allowing status downgrade
if callbackStatus == "failed" {
    // ❌ Overwrites CAPTURED status with FAILED
    payment.Status = "failed"
}
```

### Good Pattern ✅
```go
func ProcessCallback(ctx context.Context, payment *Payment, callbackStatus string) error {
    currentStatus := payment.Status

    // Validate status transition is allowed
    if !IsValidStatusTransition(currentStatus, callbackStatus) {
        logger.Logger(ctx).Errorw(trace.InvalidStatusTransition,
            map[string]interface{}{
                "current_status": currentStatus,
                "new_status":     callbackStatus,
                "payment_id":     payment.ID,
            })

        return class.ErrValidationFailure.New(
            codes.BadRequestInvalidStatusTransition,
        ).WithPublicDescriptionParams(
            fmt.Sprintf("cannot transition from %s to %s",
                currentStatus, callbackStatus),
        )
    }

    // Only update if transition is valid
    payment.Status = callbackStatus
    return payment.Save(ctx)
}

func IsValidStatusTransition(from, to string) bool {
    validTransitions := map[string][]string{
        "created":    {"pending", "failed"},
        "pending":    {"authorized", "failed"},
        "authorized": {"captured", "failed"},
        "captured":   {"refunded"}, // Terminal state
        "failed":     {}, // Terminal state
        "refunded":   {}, // Terminal state
    }

    allowedNext, exists := validTransitions[from]
    if !exists {
        return false
    }

    for _, allowed := range allowedNext {
        if allowed == to {
            return true
        }
    }

    return false
}
```

### Detection Strategy
1. Find payment/order status update code
2. Check for status transition validation before Save()
3. Verify terminal states (captured, refunded, failed) cannot be overwritten
4. Look for callback processing that directly assigns status

### Flag Conditions
- **CRITICAL**: Status downgrade allowed (CAPTURED → FAILED) → Lost payment data
- **CRITICAL**: No transition validation → Invalid state machine
- **HIGH**: Terminal state overwrite allowed → Data corruption
- **MEDIUM**: Missing logging on invalid transition

### Severity
🚨 **CRITICAL** - Invalid transitions corrupt payment state, cause settlement mismatches

---

## Check #3: Service Registry Validation Before Call 🚨 CRITICAL

### What to Check
Verify service clients are initialized and registered before use (CPS, Merchant, Optimizer).

### Bad Pattern ❌
```go
// No nil check - panic risk
func ProcessCallback(ctx context.Context, payment *Payment) error {
    // ❌ CpsServiceClient may be nil
    resp, err := c.CpsServiceClient.CallbackPaymentCPS(ctx, payment)
    return err
}
```

### Good Pattern ✅
```go
func ProcessCallback(ctx context.Context, payment *Payment) error {
    // Validate service is registered
    if c.CpsServiceClient == nil {
        logger.Logger(ctx).Errorw(trace.ServiceNotRegistered,
            map[string]interface{}{"service": "CPS"})

        return class.ErrInternalServerError.New(
            codes.ServerErrorCPSServiceNotInRegistry,
        )
    }

    resp, err := c.CpsServiceClient.CallbackPaymentCPS(ctx, payment)
    if err != nil {
        logger.Logger(ctx).Errorw(trace.CPSCallbackFailed,
            map[string]interface{}{
                "error":      err.Error(),
                "payment_id": payment.ID,
            })
        return err
    }

    return nil
}
```

### Detection Strategy
1. Find service client calls (CpsServiceClient, OptimizerClient, etc.)
2. Check for nil validation before method invocation
3. Verify error code matches pattern: `Server Error{Service}NotInRegistry`
4. Look for initialization in main.go or boot package

### Flag Conditions
- **CRITICAL**: No nil check on service client → Panic
- **HIGH**: Service not registered in init → Runtime failure
- **MEDIUM**: Missing error logging when service unavailable

### Severity
🚨 **CRITICAL** - Nil service client causes panic, crashes entire service

---

## Check #4: Gateway Field Validation (Optimizer Bypass Prevention) ⚠️ HIGH

### What to Check
Prevent merchants from manually specifying gateway unless explicitly allowed (enterprise merchants).

### Bad Pattern ❌
```go
// Accepting gateway from merchant request without validation
func CreatePayment(ctx context.Context, req *PaymentRequest) error {
    payment := &Payment{
        Gateway: req.Gateway, // ❌ Allows optimizer bypass
        Amount:  req.Amount,
    }

    return payment.Save(ctx)
}
```

### Good Pattern ✅
```go
func CreatePayment(ctx context.Context, req *PaymentRequest, merchant *Merchant) error {
    // Validate gateway parameter usage
    if req.Gateway != "" && !merchant.HasFeature("allow_gateway_selection") {
        logger.Logger(ctx).Errorw(trace.GatewayFieldNotAllowed,
            map[string]interface{}{
                "merchant_id": merchant.ID,
                "gateway":     req.Gateway,
            })

        return class.ErrValidationFailure.New(
            codes.BadRequestExtraFieldsProvided,
        ).WithPublicDescriptionParams(
            "gateway field should not be sent",
        )
    }

    var selectedGateway string
    if req.Gateway != "" {
        // Enterprise merchant - use specified gateway
        selectedGateway = req.Gateway
    } else {
        // Use Optimizer for gateway selection
        gateway, err := c.OptimizerClient.SelectProvider(ctx, req)
        if err != nil {
            return err
        }
        selectedGateway = gateway
    }

    payment := &Payment{
        Gateway: selectedGateway,
        Amount:  req.Amount,
    }

    return payment.Save(ctx)
}
```

### Detection Strategy
1. Find payment create code accepting gateway parameter
2. Check for merchant feature flag validation
3. Verify optimizer is called when gateway not specified
4. Look for `BadRequestExtraFieldsProvided` error code

### Flag Conditions
- **HIGH**: No validation on gateway field → Optimizer bypass
- **HIGH**: Missing feature flag check → Unauthorized routing control
- **MEDIUM**: No logging when gateway field rejected

### Severity
⚠️ **HIGH** - Merchants bypassing optimizer → unpredictable routing, higher failure rates

---

## Check #5: Callback Hash Verification Before Processing 🚨 CRITICAL

### What to Check
Verify callback authenticity via HMAC hash before processing payment status updates.

### Bad Pattern ❌
```go
// No hash verification - security risk
func ProcessCallback(ctx context.Context, paymentId, hash string, params map[string]string) error {
    payment, _ := GetPayment(ctx, paymentId)

    // ❌ No hash verification - anyone can forge callbacks
    return UpdatePaymentStatus(ctx, payment, params["status"])
}
```

### Good Pattern ✅
```go
func ProcessCallback(ctx context.Context, paymentId, hash string, params map[string]string) error {
    payment, err := GetPayment(ctx, paymentId)
    if err != nil {
        return err
    }

    // Verify callback hash using gateway secret
    expectedHash := GenerateCallbackHash(payment.Gateway, paymentId, params)
    if hash != expectedHash {
        logger.Logger(ctx).Errorw(trace.InvalidCallbackHash,
            map[string]interface{}{
                "payment_id":    paymentId,
                "gateway":       payment.Gateway,
                "expected_hash": expectedHash,
                "received_hash": hash,
            })

        return class.ErrBadRequestError.New(
            codes.BadRequestInvalidHash,
        ).WithPublicDescriptionParams("invalid callback hash")
    }

    // Hash verified - safe to process
    return UpdatePaymentStatus(ctx, payment, params["status"])
}

func GenerateCallbackHash(gateway, paymentId string, params map[string]string) string {
    gatewaySecret := config.GetGatewaySecret(gateway)
    data := fmt.Sprintf("%s|%s|%s", paymentId, params["status"], params["amount"])

    h := hmac.New(sha256.New, []byte(gatewaySecret))
    h.Write([]byte(data))

    return hex.EncodeToString(h.Sum(nil))
}
```

### Detection Strategy
1. Find callback handler functions (PaymentCallbackController)
2. Check for hash parameter extraction from URL
3. Verify hash comparison before payment update
4. Look for `BadRequestInvalidHash` error code

### Flag Conditions
- **CRITICAL**: No hash verification → Anyone can forge callbacks
- **CRITICAL**: Hash checked after payment update → Race condition
- **HIGH**: Using weak hash algorithm (MD5) instead of HMAC-SHA256
- **MEDIUM**: Gateway secret hardcoded instead of from config

### Severity
🚨 **CRITICAL** - Missing hash verification allows fraudulent payment confirmations

---

## Check #6: API Monolith Integration Error Handling ⚠️ HIGH

### What to Check
Handle API monolith service failures gracefully with fallback or clear error codes.

### Bad Pattern ❌
```go
// Synchronous API call without timeout
func CreatePayment(ctx context.Context, req *PaymentRequest) error {
    // ❌ No timeout - can block forever
    order, err := apiClient.GetOrder(ctx, req.OrderId)
    if err != nil {
        // ❌ Generic error - loses context
        return err
    }

    return ProcessPayment(ctx, order)
}
```

### Good Pattern ✅
```go
func CreatePayment(ctx context.Context, req *PaymentRequest) error {
    // Add timeout to API call
    apiCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    order, err := apiClient.GetOrder(apiCtx, req.OrderId)
    if err != nil {
        logger.Logger(ctx).Errorw(trace.APIServiceErrorResponse,
            map[string]interface{}{
                "error":    err.Error(),
                "order_id": req.OrderId,
            })

        // Map to specific error code
        if errors.Is(err, context.DeadlineExceeded) {
            return class.ErrInternalServerError.New(
                codes.ServerErrorAPIServiceTimeout,
            )
        }

        return class.ErrInternalServerError.New(
            codes.ServerErrorAPIServiceIntegrationFailed,
        ).Wrap(err)
    }

    return ProcessPayment(ctx, order)
}
```

### Detection Strategy
1. Find API service calls (apiClient.ExecuteRequest)
2. Check for context timeout wrapping
3. Verify error logging with trace codes
4. Look for specific error codes (ServerErrorAPIServiceIntegrationFailed)

### Flag Conditions
- **HIGH**: No timeout on API calls → Indefinite blocking
- **HIGH**: Generic error wrapping → Lost debugging context
- **MEDIUM**: No circuit breaker on API client → Cascading failures
- **MEDIUM**: Missing error metrics emission

### Severity
⚠️ **HIGH** - API monolith timeouts block payment creation, no fallback

---

## Check #7: Ledger Dual-Write with Kafka Event ⚠️ HIGH

### What to Check
Ensure payment state changes publish Kafka events for ledger journal creation (dual-write consistency).

### Bad Pattern ❌
```go
// Payment updated without ledger event
func CapturePayment(ctx context.Context, payment *Payment) error {
    payment.Status = "captured"

    // ❌ No ledger event - journal entry not created
    return payment.Save(ctx)
}
```

### Good Pattern ✅
```go
func CapturePayment(ctx context.Context, payment *Payment) error {
    payment.Status = "captured"

    // Save payment first
    err := payment.Save(ctx)
    if err != nil {
        return err
    }

    // Publish Kafka event for ledger worker
    event := &kafka_models.PaymentCapturedEvent{
        PaymentId: payment.ID,
        Amount:    payment.Amount,
        Timestamp: time.Now().Unix(),
    }

    err = c.EventProducer.PublishPaymentEvent(ctx, event)
    if err != nil {
        logger.Logger(ctx).Errorw(trace.KafkaEventPublishFailed,
            map[string]interface{}{
                "error":      err.Error(),
                "payment_id": payment.ID,
                "event_type": "payment_captured",
            })

        // Don't fail payment - ledger parity check will reconcile
        // But emit metric for monitoring
        metrics.IncrementLedgerEventFailure(ctx, "payment_captured")
    }

    return nil
}
```

### Detection Strategy
1. Find payment status update code (Save, Update)
2. Check for Kafka event publication after save
3. Verify event includes transactor_id and transactor_event
4. Look for parity check reconciliation mentions

### Flag Conditions
- **HIGH**: No Kafka event after payment save → Ledger entry missing
- **HIGH**: Event publish failure blocks payment → Unnecessary coupling
- **MEDIUM**: Missing parity check metrics → Silent ledger drift
- **MEDIUM**: Hardcoded event topic instead of config

### Severity
⚠️ **HIGH** - Missing ledger events cause accounting mismatches, hard to reconcile

---

## Check #8: Error Class and Identifier Code Usage 📋 MEDIUM

### What to Check
Use correct error class (ErrValidationFailure, ErrInternalServerError) with unique identifier code (PGPR######).

### Bad Pattern ❌
```go
// Generic error without error class
func ValidatePayment(ctx context.Context, payment *Payment) error {
    if payment.Amount <= 0 {
        // ❌ Generic error - no HTTP status mapping
        return fmt.Errorf("invalid amount")
    }

    return nil
}

// Wrong error class for validation
if payment.Token == "" && payment.IsRecurring {
    // ❌ Should use ErrValidationFailure, not ErrInternalServerError
    return class.ErrInternalServerError.New(
        codes.BadRequestCreateValidationFailure,
    )
}
```

### Good Pattern ✅
```go
func ValidatePayment(ctx context.Context, payment *Payment) errors.IError {
    if payment.Amount <= 0 {
        return class.ErrValidationFailure.New(
            codes.BadRequestInvalidAmount,
        ).WithPublicDescriptionParams("amount must be greater than 0")
    }

    if payment.Token == "" && payment.IsRecurring {
        return class.ErrValidationFailure.New(
            codes.BadRequestCreateValidationFailure,
        ).WithPublicDescriptionParams("token is required for recurring payment")
    }

    return nil
}

// Error class mapping
// ErrValidationFailure → 400 Bad Request (client error)
// ErrInternalServerError → 500 Internal Server Error (server error)
// ErrBadRequestError → 400 Bad Request (malformed request)
// ErrUnauthenticated → 401 Unauthorized (auth failure)
```

### Detection Strategy
1. Find error return statements
2. Check for `class.Err*` usage instead of `fmt.Errorf`
3. Verify error code starts with `PGPR` and has 6 digits
4. Match error class to HTTP semantics (validation → 400, server → 500)

### Flag Conditions
- **MEDIUM**: Using `fmt.Errorf` instead of error class → No HTTP mapping
- **MEDIUM**: Wrong error class for error type → Incorrect HTTP status
- **LOW**: Missing public description → Poor client error messages
- **LOW**: Error code doesn't follow PGPR###### format

### Severity
📋 **MEDIUM** - Wrong error codes confuse API consumers, break error monitoring

---

## Check #9: Callback Idempotency Check 📋 MEDIUM

### What to Check
Detect and handle duplicate callbacks from gateway (same status, same timestamp).

### Bad Pattern ❌
```go
// No deduplication - processes every callback
func ProcessCallback(ctx context.Context, payment *Payment, status string) error {
    // ❌ No duplicate check - same callback processed multiple times
    payment.Status = status
    return payment.Save(ctx)
}
```

### Good Pattern ✅
```go
func ProcessCallback(ctx context.Context, payment *Payment, callbackData map[string]interface{}) error {
    // Check if this is a duplicate callback
    isDuplicate, err := c.CpsServiceClient.CheckDuplicateCallback(ctx, callbackData)
    if err != nil {
        logger.Logger(ctx).Errorw(trace.DuplicateCheckFailed,
            map[string]interface{}{"error": err.Error()})
        // Continue processing - better to risk duplicate than miss callback
    }

    if isDuplicate {
        logger.Logger(ctx).Infow(trace.DuplicateCallbackDetected,
            map[string]interface{}{
                "payment_id": payment.ID,
                "status":     callbackData["status"],
            })

        // Return success - already processed
        return nil
    }

    // Process callback
    newStatus := callbackData["status"].(string)

    // Additional check: same status as current
    if payment.Status == newStatus {
        logger.Logger(ctx).Infow(trace.CallbackStatusUnchanged,
            map[string]interface{}{
                "payment_id": payment.ID,
                "status":     newStatus,
            })
        return nil
    }

    payment.Status = newStatus
    return payment.Save(ctx)
}
```

### Detection Strategy
1. Find callback processing code
2. Check for duplicate detection logic (CPS or independent)
3. Verify status comparison before update
4. Look for idempotency logging

### Flag Conditions
- **MEDIUM**: No duplicate callback detection → Multiple status updates
- **MEDIUM**: Status updated even when unchanged → Unnecessary DB writes
- **LOW**: Missing logging on duplicate detection
- **LOW**: Duplicate check failure blocks callback → Too strict

### Severity
📋 **MEDIUM** - Duplicate callbacks waste resources, can trigger duplicate notifications

---

## Check #10: Timeout Configuration Per Payment Method ⚠️ HIGH

### What to Check
Use method-specific timeouts (bank transfer 31 days, card 10 minutes, UPI 5 minutes).

### Bad Pattern ❌
```go
// Hardcoded timeout for all methods
func CreatePayment(ctx context.Context, req *PaymentRequest) error {
    payment := &Payment{
        Method:    req.Method,
        ExpiresAt: time.Now().Add(10 * time.Minute), // ❌ Fixed 10min
    }

    return payment.Save(ctx)
}
```

### Good Pattern ✅
```go
func CreatePayment(ctx context.Context, req *PaymentRequest) error {
    // Get method-specific timeout from constants
    timeout := getMethodTimeout(req.Method)

    payment := &Payment{
        Method:    req.Method,
        ExpiresAt: time.Now().Add(timeout),
    }

    return payment.Save(ctx)
}

func getMethodTimeout(method string) time.Duration {
    methodTimeouts := map[string]time.Duration{
        "card":          10 * time.Minute,
        "upi":           5 * time.Minute,
        "netbanking":    15 * time.Minute,
        "bank_transfer": 31 * 24 * time.Hour, // 31 days
        "emandate":      2 * time.Hour,
    }

    timeout, exists := methodTimeouts[method]
    if !exists {
        logger.Warnf("Unknown payment method: %s, using default timeout", method)
        return 10 * time.Minute // Default
    }

    return timeout
}
```

### Detection Strategy
1. Find payment creation code setting ExpiresAt field
2. Check if timeout varies by payment method
3. Verify timeout values match method expectations
4. Look for timeout constants in `constants/methods.go`

### Flag Conditions
- **HIGH**: Hardcoded timeout for all methods → Bank transfers expire too early
- **MEDIUM**: Missing timeout for new payment method → Uses wrong default
- **MEDIUM**: Timeout not configurable → Can't adjust per environment
- **LOW**: No logging when default timeout used

### Severity
⚠️ **HIGH** - Wrong timeout causes valid payments to expire prematurely

---

## Summary

| Check | Severity | Description |
|-------|----------|-------------|
| 1. Mutex Lock & Unlock | 🚨 CRITICAL | Defer unlock after mutex acquisition |
| 2. Status Transition Validation | 🚨 CRITICAL | Validate payment state machine rules |
| 3. Service Registry Validation | 🚨 CRITICAL | Check service clients initialized before use |
| 4. Gateway Field Validation | ⚠️ HIGH | Prevent optimizer bypass via gateway param |
| 5. Callback Hash Verification | 🚨 CRITICAL | HMAC verify before processing callbacks |
| 6. API Monolith Integration | ⚠️ HIGH | Timeout and error handling on API calls |
| 7. Ledger Dual-Write Events | ⚠️ HIGH | Kafka events for journal entry creation |
| 8. Error Class Usage | 📋 MEDIUM | Use proper error classes and PGPR codes |
| 9. Callback Idempotency | 📋 MEDIUM | Detect and skip duplicate callbacks |
| 10. Method-Specific Timeouts | ⚠️ HIGH | Use correct timeout per payment method |

## File Triggers

Load this reference when PR modifies files in **pg-router repository**:
- `internal/payments/*` - Payment processing core
- `internal/orders/*` - Order management
- `**/callback*.go` - Callback handlers
- `internal/optimizer/*` - Gateway selection
- `internal/ledger/*` - Ledger integration
- `internal/api/*` - API monolith integration
- `internal/middleware/*` - Request middleware

## Common Repositories

This reference is ONLY for:
- **pg-router** - The payment routing service itself

**NOT** for:
- **goutils/router** - Router SDK (use services-router.md instead)
- **payments-upi** - UPI service (uses router SDK)
- **terminals** - Terminal management (uses router SDK)

## References

Based on real patterns from:
- `razorpay/pg-router/.agents/skills/repo-skill/` - PG-Router repo skill
- `internal/payments/core/create.go` - Payment creation patterns
- `internal/payments/core/callback.go` - Callback processing patterns
- `modules/domain/payment/` - Payment constraints and flows
- `modules/foundation/` - Error handling, testing, logging patterns
