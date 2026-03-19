# Razorpay Logging Standards

Comprehensive guidelines for logging in Razorpay services.

## Log Levels

### ERROR
**Use for:** Actual errors that require attention
- Unhandled exceptions
- Failed operations that impact users
- Integration failures
- Data corruption or loss

**Trigger rate:** ~1% of requests

```go
// Good - Actual error
logger.Log(ctx).Error("payment failed", 
    "payment_id", paymentID,
    "error", err.Error(),
    "gateway", gateway)

// Bad - Not an error
logger.Log(ctx).Error("starting payment processing") // Use INFO or DEBUG
```

### WARN
**Use for:** Recoverable issues, deprecation warnings
- Retried operations
- Degraded service (using fallback)
- Deprecated API usage
- Rate limiting applied

**Trigger rate:** ~10% of requests

```go
// Good - Recoverable issue
logger.Log(ctx).Warn("gateway timeout, using fallback",
    "gateway", primary,
    "fallback", secondary)

// Bad - Normal operation
logger.Log(ctx).Warn("processing payment") // Use INFO
```

### INFO
**Use for:** Significant business events
- Transaction completions
- User actions
- State changes
- Audit-worthy events

**Trigger rate:** ~100% (for the code path)

```go
// Good - Business event
logger.Log(ctx).Info("payment completed",
    "payment_id", paymentID,
    "amount", amount,
    "merchant_id", merchantID)

// Bad - Entry/exit tracing
logger.Log(ctx).Info("entering HandlePayment") // Use DEBUG
```

### DEBUG
**Use for:** Development and troubleshooting
- Entry/exit points
- Variable values
- Flow tracing
- Intermediate states

**Trigger rate:** 0% in production (typically disabled)

```go
// Good - Development tracing
logger.Log(ctx).Debug("parsed request",
    "body_size", len(body),
    "headers_count", len(headers))
```

## Anti-Patterns

### 1. Logs in Loops

**Problem:** Generates N logs for N iterations
**Impact:** 10x-1000x volume multiplier

```go
// BAD - 1000 logs per request
for _, item := range items {
    logger.Log(ctx).Info("processing item", "id", item.ID)
    process(item)
}

// GOOD - 1 log per request
var processed, failed int
for _, item := range items {
    if err := process(item); err != nil {
        failed++
    } else {
        processed++
    }
}
logger.Log(ctx).Info("batch complete", 
    "processed", processed, 
    "failed", failed,
    "total", len(items))
```

### 2. Entry/Exit Logs at INFO Level

**Problem:** Always triggers, high volume
**Impact:** Doubles log volume

```go
// BAD - Generates 2 logs per request at INFO
logger.Log(ctx).Info("entering HandlePayment")
// ... processing ...
logger.Log(ctx).Info("exiting HandlePayment")

// GOOD - Use DEBUG or remove entirely
logger.Log(ctx).Debug("entering HandlePayment")
// Or use metrics for counting
requestCounter.Inc()
```

### 3. Success Logs in Hot Paths

**Problem:** High RPS = high log volume
**Impact:** 86,400 logs per RPS per day

```go
// BAD - 43M logs/day at 500 RPS
func HandlePayment(ctx context.Context, req *Request) {
    // process...
    logger.Log(ctx).Info("payment successful")
}

// GOOD - Use metrics
paymentSuccessCounter.Inc()
// Or sample
if shouldSample(0.01) { // 1% sampling
    logger.Log(ctx).Info("payment successful (sampled)")
}
```

### 4. Sensitive Data in Logs

**Problem:** Security risk, compliance violation
**Impact:** PII exposure, regulatory issues

```go
// BAD - Exposes sensitive data
logger.Log(ctx).Info("user authenticated",
    "password", req.Password,        // NEVER
    "card_number", card.Number,      // NEVER
    "cvv", card.CVV,                 // NEVER
    "api_key", config.APIKey)        // NEVER

// GOOD - Log identifiers only
logger.Log(ctx).Info("user authenticated",
    "user_id", user.ID,
    "card_last_four", card.LastFour)
```

### 5. String Interpolation

**Problem:** Allocations in hot path, harder to query
**Impact:** Performance overhead

```go
// BAD - String allocation
logger.Log(ctx).Info(fmt.Sprintf("processing payment %s for merchant %s", 
    paymentID, merchantID))

// GOOD - Structured fields
logger.Log(ctx).Info("processing payment",
    "payment_id", paymentID,
    "merchant_id", merchantID)
```

## Context Fields

Always include relevant context identifiers:

```go
logger.Log(ctx).Info("payment processed",
    // Request context
    "request_id", requestID,
    "trace_id", traceID,
    
    // Business context
    "payment_id", paymentID,
    "merchant_id", merchantID,
    "transaction_id", txnID,
    
    // Operation context
    "gateway", gateway,
    "amount", amount,
    "currency", currency)
```

## Cost Guidelines

### Units Calculation
```
1 Coralogix Unit = 1 MB of log data
Daily logs = RPS × 86,400 seconds × trigger_probability
Daily bytes = Daily logs × avg_log_size (typically 200-500 bytes)
Daily units = Daily bytes / 1,000,000
```

### Example Calculation
```
Route RPS: 500
Log size: 300 bytes
Trigger probability: 1.0 (always)

Daily logs: 500 × 86,400 × 1.0 = 43,200,000
Daily bytes: 43,200,000 × 300 = 12,960,000,000 (12.96 GB)
Daily units: 12,960 units
```

### Cost Reduction Strategies

| Strategy | Typical Savings |
|----------|-----------------|
| Remove entry/exit logs | 95% (disabled DEBUG) |
| Consolidate loop logs | 90% (1 instead of N) |
| Add 1% sampling | 99% |
| Use metrics instead | 100% |
| Reduce log verbosity | 50% |

## Quick Reference

### Do's
- ✅ Use structured logging with fields
- ✅ Include request_id, transaction_id context
- ✅ Use appropriate log levels
- ✅ Log errors with full context
- ✅ Use metrics for counting
- ✅ Consolidate loop logs

### Don'ts
- ❌ Log in loops
- ❌ Log sensitive data (PII, credentials)
- ❌ Use INFO for entry/exit
- ❌ Log success for every request
- ❌ Use string interpolation
- ❌ Create excessive fields
