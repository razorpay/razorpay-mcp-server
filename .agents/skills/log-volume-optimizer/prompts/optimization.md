# Log Optimization Recommendations Prompt

## Context
Based on the log volume analysis, generate specific, actionable recommendations to reduce logging costs while maintaining observability.

## Optimization Strategies

### Strategy 1: Remove Unnecessary Logs
**Target**: Logs that provide no operational value
**Savings**: 100% of that log's volume

Identify and recommend removal of:
- Entry/exit logs at INFO level (should be DEBUG or removed)
- Success logs in hot paths (use metrics instead)
- Duplicate information already captured elsewhere
- Debug statements left in production code

**Example Change**:
```go
// REMOVE - Entry log in hot path (500 RPS = 43M logs/day)
- logger.Log(ctx).Info("entering HandlePayment")

// KEEP - Only log on errors
+ // Removed verbose entry log, using metrics for request counting
```

### Strategy 2: Change Log Levels
**Target**: INFO logs that should be DEBUG
**Savings**: ~95% (DEBUG typically disabled in production)

Identify logs that should be DEBUG:
- Flow tracing logs
- Variable value dumps
- Internal state logging
- Development-time debugging

**Example Change**:
```go
// BEFORE - INFO level, always logged
- logger.Log(ctx).Info("parsed request body", "size", len(body))

// AFTER - DEBUG level, disabled in production
+ logger.Log(ctx).Debug("parsed request body", "size", len(body))
```

### Strategy 3: Add Sampling
**Target**: High-frequency logs that are still needed occasionally
**Savings**: 90-99% depending on sample rate

Apply sampling to:
- Health check logs
- Metrics/stats logs
- High-RPS success paths
- Periodic status updates

**Example Change**:
```go
// BEFORE - Every request logged
- logger.Log(ctx).Info("request processed", "duration", duration)

// AFTER - 1% sampling
+ if rand.Float64() < 0.01 {
+     logger.Log(ctx).Info("request processed (sampled)", "duration", duration)
+ }
```

### Strategy 4: Consolidate Loop Logs
**Target**: Logs inside loops
**Savings**: (N-1)/N where N = loop iterations

Replace per-iteration logs with summaries:
- Count successes/failures
- Collect important data points
- Log summary after loop

**Example Change**:
```go
// BEFORE - Log per item (100 items = 100 logs)
- for _, item := range items {
-     logger.Log(ctx).Info("processing item", "id", item.ID)
-     process(item)
- }

// AFTER - Single summary log
+ var processed, failed int
+ for _, item := range items {
+     if err := process(item); err != nil {
+         failed++
+     } else {
+         processed++
+     }
+ }
+ logger.Log(ctx).Info("batch complete", "processed", processed, "failed", failed)
```

### Strategy 5: Use Metrics Instead
**Target**: Counting/measuring logs
**Savings**: 100% of log volume

Replace logs with Prometheus metrics:
- Request counts
- Duration histograms
- Error rates
- Success rates

**Example Change**:
```go
// BEFORE - Log every success
- logger.Log(ctx).Info("payment successful", "amount", amount)

// AFTER - Increment metric (use ONLY low-cardinality labels)
+ paymentSuccessCounter.WithLabelValues(paymentMode, gateway).Inc()
+ paymentAmountHistogram.WithLabelValues(paymentMode).Observe(float64(amount))
```

> **WARNING**: Never use high-cardinality fields as Prometheus metric labels.
> Forbidden labels: `merchantID`, `paymentID`, `userID`, `orderID`, `transactionID`
> Allowed labels: `payment_mode`, `gateway`, `method`, `status_code`, `error_type`

### Strategy 6: Reduce Log Verbosity
**Target**: Overly detailed logs
**Savings**: Proportional to data reduction

Reduce logged data:
- Remove redundant fields
- Truncate large payloads
- Use IDs instead of full objects
- Remove stack traces for non-errors

**Example Change**:
```go
// BEFORE - Full object logged
- logger.Log(ctx).Info("user data", "user", user)

// AFTER - Only relevant fields
+ logger.Log(ctx).Info("user action", "userID", user.ID, "action", action)
```

## Recommendation Format

For each recommendation, provide:

```markdown
### Recommendation #X: [Title]

**File**: `path/to/file.go`
**Line**: 123
**Current Log**:
```go
logger.Log(ctx).Info("message", "field", value)
```

**Issue**: [Why this is a problem]
**Strategy**: [Which strategy applies]
**Estimated Savings**: X units/day (Y% of this log)

**Suggested Change**:
```go
// New code here
```

**Risk Assessment**: LOW/MEDIUM/HIGH
**Justification**: [Why this change is safe]
```

## Prioritization

Rank recommendations by:
1. **Impact**: Higher savings first
2. **Risk**: Lower risk first
3. **Effort**: Easier changes first

Calculate priority score:
```
priority = (daily_units_saved × 10) - (risk_factor × 100) - (effort_hours × 5)
```

## PR Description Template

```markdown
## Log Volume Optimization

### Summary
This PR reduces log volume for {{service_name}} based on analysis of current logging patterns.

### Changes
- Removed X unnecessary log statements
- Changed Y logs from INFO to DEBUG
- Added sampling to Z high-frequency logs
- Consolidated N loop logs into summaries

### Impact
| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Estimated Daily Units | X | Y | -Z% |
| Log Statements | X | Y | -Z |
| Quota Utilization | X% | Y% | -Z% |

### Testing
- [ ] Verified critical error paths still log appropriately
- [ ] Confirmed DEBUG logs are disabled in production config
- [ ] Validated metrics are recording correctly (if applicable)

### Rollback
If issues arise, revert this PR. Log changes are low-risk as they don't affect business logic.

---
*Generated by Log Volume Optimizer Skill*
```

## Safety Guidelines

### Never Remove
- ERROR logs for actual errors
- Security-related audit logs
- Compliance-required logs
- Transaction completion logs

### Never Use as Prometheus Metric Labels (High Cardinality)
- `merchantID`, `merchant_id`
- `paymentID`, `payment_id`, `orderID`, `order_id`
- `userID`, `user_id`, `customerID`
- `transactionID`, `request_id`, `trace_id`
- Any field with unbounded unique values

### Always Verify
- Error paths still have adequate logging
- Debug information available when needed
- No PII exposed in remaining logs
- Log levels match environment config

### Risk Indicators
- HIGH RISK: Removing error logs, changing error handling
- MEDIUM RISK: Consolidating logs, changing levels
- LOW RISK: Adding sampling, reducing verbosity
