# Log Volume Optimizer Examples

Complete examples of using the log-volume-optimizer skill.

## Example 1: Basic Analysis

### User Request
```
Analyze log volume for pg-router at /Users/parth.anand/razorpay/pg-router
```

### Expected Workflow

1. **Scan Repository**
   - Parse all `.go` files
   - Extract log statements with context
   - Identify loops, error handlers, functions

2. **Estimate Volume**
   - Use default RPS (500) if not specified
   - Calculate daily units per log

3. **Generate Report**

### Expected Output
```markdown
## Log Volume Analysis: pg-router

### Repository Scan
- Files scanned: 45
- Log statements found: 156
- Language: Go

### Volume Summary
| Metric | Value |
|--------|-------|
| Total log statements | 156 |
| Estimated daily units | 2,400 |
| Assumed RPS | 500 |

### By Level
| Level | Count | Units/Day | % |
|-------|-------|-----------|---|
| ERROR | 45 | 39 | 1.6% |
| WARN | 12 | 104 | 4.3% |
| INFO | 89 | 2,200 | 91.7% |
| DEBUG | 10 | 0 | 0% |

### By Category
| Category | Count | Units/Day |
|----------|-------|-----------|
| CRITICAL | 5 | 1,200 |
| HIGH | 20 | 800 |
| MEDIUM | 50 | 350 |
| LOW | 81 | 50 |

### Top 10 High-Impact Logs
| # | File:Line | Level | Function | Units/Day | Issue |
|---|-----------|-------|----------|-----------|-------|
| 1 | handler/payment.go:45 | INFO | HandlePayment | 432 | Hot path |
| 2 | router/dispatch.go:123 | INFO | Dispatch | 389 | Entry log |
| 3 | service/batch.go:78 | INFO | ProcessBatch | 345 | In loop |
| 4 | handler/refund.go:56 | INFO | HandleRefund | 162 | Hot path |
| 5 | gateway/client.go:89 | INFO | CallGateway | 156 | Every request |
```

---

## Example 2: With Traffic Parameters

### User Request
```
Analyze pg-router with:
- Average RPS: 500
- Peak RPS: 2000
- Daily merchants: 50,000
- Assigned quota: 480 units
```

### Expected Output
```markdown
## Log Volume Analysis: pg-router

### Configuration
| Parameter | Value |
|-----------|-------|
| Average RPS | 500 |
| Peak RPS | 2,000 |
| Daily Merchants | 50,000 |
| Assigned Quota | 480 units |

### Volume Summary
| Metric | Value |
|--------|-------|
| Estimated daily units | 2,400 |
| Assigned quota | 480 |
| **Utilization** | **500%** ⚠️ |
| Over quota by | 1,920 units |

### Status: 🔴 OVER QUOTA

The estimated log volume significantly exceeds the assigned quota.
Immediate optimization required.

### Recommended Actions
1. Address CRITICAL category logs (saves ~1,200 units)
2. Address HIGH category logs (saves ~800 units)
3. Target: Reduce to <400 units (83% reduction needed)
```

---

## Example 3: Full Optimization Flow

### User Request
```
1. Scan pg-router for log statements
2. Get actual consumption from Coralogix
3. Compare estimated vs actual
4. Suggest top 5 optimizations
5. Show the code changes needed
```

### Expected Output

#### Phase 1: Repository Scan
```markdown
### Scan Results
- Repository: /Users/parth.anand/razorpay/pg-router
- Files scanned: 45 Go files
- Log statements: 156 found
- In loops: 8
- In error handlers: 45
```

#### Phase 2: Coralogix Data
```markdown
### Actual Consumption (Coralogix)
- Application: pg-router
- Period: Last 7 days
- Daily average: 420 units
- Peak day: 480 units
- Source: coralogix_mcp
```

#### Phase 3: Comparison
```markdown
### Estimated vs Actual
| Metric | Estimated | Actual | Variance |
|--------|-----------|--------|----------|
| Daily Units | 2,400 | 420 | +471% |

### Variance Analysis
The large positive variance suggests:
1. Many estimated logs are DEBUG (disabled in prod)
2. Some code paths rarely execute
3. Error handlers triggered infrequently (~1%)

### Adjusted Estimate
Applying trigger probabilities:
- ERROR logs: 1% trigger rate
- DEBUG logs: 0% (disabled)
- Adjusted estimate: 450 units
- Adjusted variance: +7% ✓
```

#### Phase 4: Recommendations
```markdown
### Top 5 Optimization Recommendations

#### 1. Consolidate batch processing logs
**File:** `service/batch.go:78-85`
**Issue:** 3 logs inside loop, 100 iterations average
**Savings:** 345 units/day (14% of total)

**Current:**
```go
for _, item := range items {
    logger.Log(ctx).Info("processing item", "id", item.ID)
    if err := process(item); err != nil {
        logger.Log(ctx).Error("item failed", "id", item.ID, "error", err)
    }
    logger.Log(ctx).Info("item complete", "id", item.ID)
}
```

**Suggested:**
```go
var processed, failed int
var errors []string
for _, item := range items {
    if err := process(item); err != nil {
        failed++
        if len(errors) < 5 {
            errors = append(errors, fmt.Sprintf("%s: %v", item.ID, err))
        }
    } else {
        processed++
    }
}
logger.Log(ctx).Info("batch complete",
    "total", len(items),
    "processed", processed,
    "failed", failed)
if failed > 0 {
    logger.Log(ctx).Error("batch had failures",
        "failed", failed,
        "sample_errors", errors)
}
```

---

#### 2. Remove entry log in payment handler
**File:** `handler/payment.go:45`
**Issue:** Entry log at INFO in hot path (200 RPS)
**Savings:** 172 units/day (7% of total)

**Current:**
```go
func HandlePayment(ctx context.Context, req *PaymentRequest) (*Response, error) {
    logger.Log(ctx).Info("entering HandlePayment",
        "merchant_id", req.MerchantID,
        "amount", req.Amount)
    // ...
}
```

**Suggested:**
```go
func HandlePayment(ctx context.Context, req *PaymentRequest) (*Response, error) {
    // Entry log removed - using metrics for request tracking
    // If debugging needed, uncomment:
    // logger.Log(ctx).Debug("entering HandlePayment", ...)
    
    // ...
}
```

---

#### 3. Change gateway call log to DEBUG
**File:** `gateway/client.go:89`
**Issue:** Verbose INFO log on every gateway call
**Savings:** 156 units/day (6.5% of total)

**Current:**
```go
logger.Log(ctx).Info("calling gateway",
    "gateway", gateway,
    "timeout", timeout,
    "retry_count", retryCount)
```

**Suggested:**
```go
logger.Log(ctx).Debug("calling gateway",
    "gateway", gateway,
    "timeout", timeout,
    "retry_count", retryCount)
```

---

#### 4. Add sampling to success logs
**File:** `handler/payment.go:120`
**Issue:** Success log on every payment (200 RPS)
**Savings:** 170 units/day (7% of total)

**Current:**
```go
logger.Log(ctx).Info("payment successful",
    "payment_id", paymentID,
    "duration_ms", duration.Milliseconds())
```

**Suggested:**
```go
// Use metric for counting
paymentSuccessTotal.Inc()
paymentDuration.Observe(duration.Seconds())

// Sample 1% for debugging
if shouldSample(0.01) {
    logger.Log(ctx).Info("payment successful (sampled)",
        "payment_id", paymentID,
        "duration_ms", duration.Milliseconds(),
        "sample_rate", "1%")
}
```

---

#### 5. Reduce verbosity in router dispatch
**File:** `router/dispatch.go:123`
**Issue:** Full request object logged
**Savings:** 130 units/day (5.4% of total)

**Current:**
```go
logger.Log(ctx).Info("dispatching request",
    "request", req,  // Full object!
    "route", route)
```

**Suggested:**
```go
logger.Log(ctx).Info("dispatching request",
    "route", route,
    "method", req.Method,
    "content_length", req.ContentLength)
```
```

#### Phase 5: Summary
```markdown
### Optimization Summary

| Recommendation | Savings | % of Total |
|----------------|---------|------------|
| Consolidate batch logs | 345 | 14% |
| Remove entry log | 172 | 7% |
| Change to DEBUG | 156 | 6.5% |
| Add sampling | 170 | 7% |
| Reduce verbosity | 130 | 5.4% |
| **Total Savings** | **973** | **40%** |

### Projected Impact
| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Daily Units | 420 | 252 | -40% |
| Quota Utilization | 87.5% | 52.5% | -35% |

### Next Steps
1. Review suggested changes above
2. Say "Apply these changes" to generate code modifications
3. Say "Create PR" to raise a pull request with changes
```

---

## Example 4: Generate and Raise PR

### User Request
```
Apply the optimizations and create a PR for pg-router
```

### Expected Output
```markdown
### Changes Applied

Modified 5 files:
- `service/batch.go` - Consolidated loop logs
- `handler/payment.go` - Removed entry log, added sampling
- `gateway/client.go` - Changed to DEBUG level
- `router/dispatch.go` - Reduced verbosity

### PR Created

**Branch:** `log-optimization/pg-router-20240115`
**PR Title:** `[pg-router] Reduce log volume by 40%`
**PR URL:** https://github.com/razorpay/pg-router/pull/4125

### PR Description
```markdown
## Log Volume Optimization

### Summary
Reduces log volume by ~40% based on analysis by log-volume-optimizer skill.

### Changes
| File | Change | Savings |
|------|--------|---------|
| service/batch.go | Consolidated 3 loop logs into summary | 345 units/day |
| handler/payment.go | Removed entry log, added 1% sampling | 342 units/day |
| gateway/client.go | Changed INFO to DEBUG | 156 units/day |
| router/dispatch.go | Reduced log verbosity | 130 units/day |

### Impact
| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Estimated Daily Units | 420 | 252 | -40% |
| Quota Utilization | 87.5% | 52.5% | -35% |

### Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Verified error logging still works
- [ ] Confirmed DEBUG disabled in prod config

### Rollback
Revert this PR if issues arise. Log changes don't affect business logic.

---
*Generated by log-volume-optimizer skill*
```

### Awaiting Review
- AI PR Reviewer will analyze the changes
- Human review required from platform-team
```

---

## Example 5: Multi-Service Analysis

### User Request
```
Compare log volume across pg-router, api, and dashboard services
```

### Expected Output
```markdown
## Multi-Service Log Volume Comparison

### Summary
| Service | Quota | Actual | Utilization | Status |
|---------|-------|--------|-------------|--------|
| pg-router | 480 | 420 | 87.5% | ⚠️ Warning |
| api | 800 | 650 | 81.3% | ⚠️ Warning |
| dashboard | 200 | 120 | 60% | ✅ Healthy |

### Optimization Priority
1. **pg-router** - Highest utilization, optimize first
2. **api** - Large absolute volume, significant savings potential
3. **dashboard** - Healthy, lower priority

### Combined Savings Potential
| Service | Current | After Optimization | Savings |
|---------|---------|-------------------|---------|
| pg-router | 420 | 252 | 168 units |
| api | 650 | 455 | 195 units |
| dashboard | 120 | 96 | 24 units |
| **Total** | **1,190** | **803** | **387 units** |

### Next Steps
Say "Optimize pg-router" to start with the highest priority service.
```
