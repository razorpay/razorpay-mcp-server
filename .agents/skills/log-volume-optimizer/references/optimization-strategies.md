# Log Optimization Strategies

Detailed techniques for reducing log volume while maintaining observability.

## Strategy Overview

| Strategy | Target | Typical Savings | Risk |
|----------|--------|-----------------|------|
| Remove Entry/Exit | INFO entry/exit logs | 95% | Low |
| Consolidate Loops | Logs inside loops | 90% | Low |
| Add Sampling | High-frequency logs | 99% | Medium |
| Use Metrics | Counting/measuring | 100% | Low |
| Change Level | Verbose INFO logs | 95% | Low |
| Reduce Verbosity | Large log payloads | 50% | Low |

---

## 1. Remove Entry/Exit Logs

**Target:** Entry and exit logs at INFO or higher level

**Problem:** These logs trigger on every request, providing minimal debugging value while consuming significant volume.

**Identification:**
```go
// Look for these patterns
logger.Log(ctx).Info("entering HandlePayment")
logger.Log(ctx).Info("exiting HandlePayment")
logger.Log(ctx).Info("starting process")
logger.Log(ctx).Info("finished processing")
```

**Solution:**
```go
// Option 1: Remove entirely (recommended for hot paths)
// Removed: Entry/exit logs not needed, using metrics for request tracking

// Option 2: Change to DEBUG (if needed for development)
logger.Log(ctx).Debug("entering HandlePayment")

// Option 3: Use metrics
requestsTotal.WithLabelValues("HandlePayment").Inc()
```

**Savings Calculation:**
```
RPS: 500
Log size: 250 bytes
Daily logs: 500 × 86,400 = 43,200,000
Daily bytes: 43,200,000 × 250 = 10.8 GB
Daily units saved: 10,800
```

---

## 2. Consolidate Loop Logs

**Target:** Log statements inside loops

**Problem:** Generates N logs for N iterations, multiplying volume by loop size.

**Identification:**
```go
// Logs inside for/range loops
for _, item := range items {
    logger.Log(ctx).Info("processing item", "id", item.ID)  // N logs!
    if err := process(item); err != nil {
        logger.Log(ctx).Error("item failed", "id", item.ID)  // Error logs OK
    }
}
```

**Solution:**
```go
// Collect stats during loop, log summary after
var processed, failed int
var failedIDs []string

for _, item := range items {
    if err := process(item); err != nil {
        failed++
        if len(failedIDs) < 10 {  // Cap sample size
            failedIDs = append(failedIDs, item.ID)
        }
    } else {
        processed++
    }
}

// Single summary log
logger.Log(ctx).Info("batch processing complete",
    "total", len(items),
    "processed", processed,
    "failed", failed,
    "sample_failed_ids", failedIDs)

// Log error if any failures (still important for alerting)
if failed > 0 {
    logger.Log(ctx).Error("batch had failures",
        "failed_count", failed,
        "sample_ids", failedIDs)
}
```

**Savings Calculation:**
```
Loop iterations: 100
RPS: 50
Original logs: 100 × 50 × 86,400 = 432,000,000
Consolidated: 1 × 50 × 86,400 = 4,320,000
Reduction: 99%
```

---

## 3. Add Sampling

**Target:** High-frequency logs that are occasionally useful

**Problem:** Every request logs create massive volume for diminishing returns.

**Identification:**
- Logs in handlers with RPS > 100
- Success/completion logs
- Timing/performance logs

**Solution:**

```go
// Simple random sampling
import "math/rand"

func shouldSample(rate float64) bool {
    return rand.Float64() < rate
}

// 1% sampling for high-frequency logs
if shouldSample(0.01) {
    logger.Log(ctx).Info("request processed (sampled)",
        "duration_ms", duration.Milliseconds(),
        "sample_rate", "1%")
}

// Deterministic sampling (same request always sampled or not)
func deterministicSample(requestID string, rate float64) bool {
    hash := fnv.New32a()
    hash.Write([]byte(requestID))
    return float64(hash.Sum32()%10000)/10000 < rate
}
```

**Advanced: Adaptive Sampling:**
```go
// Sample more errors, less successes
func adaptiveSample(err error) float64 {
    if err != nil {
        return 1.0  // Always log errors
    }
    return 0.01    // 1% for successes
}
```

**Savings Calculation:**
```
Original: 500 RPS × 86,400 = 43,200,000 logs/day
1% sampling: 432,000 logs/day
Reduction: 99%
```

---

## 4. Use Metrics Instead

**Target:** Logs used for counting or measuring

**Problem:** Logging every event for aggregation is expensive; metrics are designed for this.

**Identification:**
```go
// Counting patterns
logger.Log(ctx).Info("payment successful")
logger.Log(ctx).Info("request processed", "duration", duration)
logger.Log(ctx).Info("cache hit")
```

**Solution:**
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    paymentTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "payments_total",
            Help: "Total payments processed",
        },
        []string{"status", "gateway"},
    )
    
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "request_duration_seconds",
            Help:    "Request duration histogram",
            Buckets: prometheus.DefBuckets,
        },
        []string{"handler", "status"},
    )
)

// Replace logs with metrics
paymentTotal.WithLabelValues("success", gateway).Inc()
requestDuration.WithLabelValues("HandlePayment", "200").Observe(duration.Seconds())
```

**Savings:** 100% (logs completely eliminated)

**When to Keep Logs:**
- Unique identifiers needed (payment_id, merchant_id)
- Complex context for debugging
- Audit requirements

---

## 5. Change Log Level

**Target:** Verbose logs at INFO that should be DEBUG

**Problem:** INFO logs are always enabled in production; DEBUG is typically disabled.

**Identification:**
```go
// Verbose INFO logs that aren't business events
logger.Log(ctx).Info("parsed request body", "size", len(body))
logger.Log(ctx).Info("calling gateway", "gateway", gateway)
logger.Log(ctx).Info("database query executed", "rows", count)
```

**Solution:**
```go
// Change to DEBUG - disabled in production
logger.Log(ctx).Debug("parsed request body", "size", len(body))
logger.Log(ctx).Debug("calling gateway", "gateway", gateway)
logger.Log(ctx).Debug("database query executed", "rows", count)
```

**Savings:** ~95% (DEBUG typically disabled in prod)

**Keep as INFO:**
- Business events (payment completed, order created)
- Significant state changes
- Audit-required events

---

## 6. Reduce Verbosity

**Target:** Logs with excessive or large payloads

**Problem:** Large log sizes directly increase unit consumption.

**Identification:**
```go
// Full object dumps
logger.Log(ctx).Info("request received", "body", req)
logger.Log(ctx).Info("response", "data", response)

// Excessive fields
logger.Log(ctx).Info("payment",
    "payment", payment,           // Full object
    "merchant", merchant,         // Full object
    "customer", customer,         // Full object
    "request_headers", headers)   // All headers
```

**Solution:**
```go
// Log only IDs and relevant fields
logger.Log(ctx).Info("payment processed",
    "payment_id", payment.ID,
    "merchant_id", merchant.ID,
    "amount", payment.Amount,
    "status", payment.Status)

// Truncate large payloads
func truncate(s string, max int) string {
    if len(s) <= max {
        return s
    }
    return s[:max] + "...(truncated)"
}

logger.Log(ctx).Debug("request body", 
    "body", truncate(string(body), 500))
```

**Savings:** 50-80% depending on original verbosity

---

## Priority Order

When optimizing, address issues in this order:

1. **CONSOLIDATE_LOOP** - Highest multiplier impact
2. **USE_METRICS** - Complete elimination of log
3. **CHANGE_TO_DEBUG** - 95% reduction, low risk
4. **ADD_SAMPLING** - 99% reduction, some visibility loss
5. **REMOVE_ENTRY_EXIT** - Clean removal, low risk
6. **REDUCE_VERBOSITY** - Smaller impact, easy wins

## Safety Guidelines

### Always Keep
- ERROR logs for actual errors
- Security audit logs
- Compliance-required logs
- Transaction completion logs

### Always Verify
- Error paths still have logging
- Debug info available when needed
- No PII in remaining logs
- Metrics are working correctly

### Testing Changes
1. Deploy to staging first
2. Verify alerting still works
3. Check dashboards populate
4. Confirm debug capability in dev mode
