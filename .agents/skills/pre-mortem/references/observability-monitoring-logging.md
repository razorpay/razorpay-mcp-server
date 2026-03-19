# Observability: Monitoring & Logging

## Overview

Validates that critical code paths have proper monitoring (metrics) and logging (trace codes) for production observability.

**Load when:** All PRs (final validation step after other checks)

**Total Checks:** 5

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 2
- 📋 Medium: 1

---

## Check 1: Error Metrics on Critical Failures 🚨 CRITICAL

### What to Check

Critical error paths must emit metrics for monitoring/alerting.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Error logged but no metric
func ProcessPayment(ctx context.Context, req PaymentRequest) error {
    err := gateway.Charge(req.Amount)
    if err != nil {
        logger.Error(ctx, "payment_failed", "error", err)  // ❌ Only logged
        return err
    }
}
```

**Problem:**
- No metric emitted → Can't alert on spike in failures
- No visibility into failure rate trends
- Manual log searching to detect issues

### Good Pattern ✅

```go
// CORRECT: Error logged + metric emitted
func ProcessPayment(ctx context.Context, req PaymentRequest) error {
    err := gateway.Charge(req.Amount)
    if err != nil {
        logger.Error(ctx, "payment_failed", "error", err, "gateway", req.Gateway)

        // ✅ Emit metric for monitoring
        metrics.Count(ctx, "payment.failed", 1, map[string]string{
            "gateway": req.Gateway,
            "error_type": getErrorType(err),
        })

        return err
    }

    // ✅ Also emit success metric
    metrics.Count(ctx, "payment.success", 1, map[string]string{
        "gateway": req.Gateway,
    })

    return nil
}
```

### Detection Strategy

**Ask user for repo context first:**
```
To suggest monitoring patterns, I need to understand your metrics library.
Can you share:
1. Path to a file that emits metrics (e.g., service.go, handler.go)
2. Or your CLAUDE.md monitoring section
```

**Then analyze:**
```bash
# Find error returns without metrics
grep -n "logger.Error\|logger.Errorw" <changed_file> | while read line; do
    line_num=$(echo "$line" | cut -d: -f1)
    # Check if a metrics call (with library prefix) exists nearby (±8 lines)
    context=$(sed -n "$((line_num-2)),$((line_num+8))p" <changed_file>)
    if ! echo "$context" | grep -qE "metrics\.[A-Za-z]|prometheus\.[A-Za-z]|statsd\.[A-Za-z]|\.Inc\(\)|\.Add\(|\.Observe\("; then
        FLAG: "Error logged without metric at line $line_num"
    fi
done
```

> **Note:** The check looks for a metrics library call (e.g., `metrics.Count`, `prometheus.Inc`) rather than bare identifiers like `Count` or `Increment` which appear in non-metric code. If the project uses a different metrics prefix, adapt accordingly.

### Flag Conditions

Flag if:
- `logger.Error()` in critical path (payment, onboarding, refund)
- No metric emitted within 5 lines
- Error returned to user/external system

### Severity

🚨 **Critical** - No alerting possible, silent production issues

---

## Check 2: Trace Codes for Error Logs 🚨 CRITICAL

### What to Check

Error logs must use standardized trace codes (not free-form strings).

### Bad Pattern ❌

```go
// ANTI-PATTERN: Free-form log message
logger.Error(ctx, "payment processing failed", "error", err)  // ❌ No trace code
logger.Error(ctx, "Failed to create merchant", "error", err)  // ❌ No trace code
```

**Problem:**
- Hard to grep/search for specific errors
- Can't track error frequency
- No standardization across team
- Log analysis tools can't categorize

### Good Pattern ✅

```go
// CORRECT: Use trace code constants
logger.Error(ctx, TraceCode.PAYMENT_PROCESSING_FAILED, "error", err, "payment_id", paymentID)
logger.Error(ctx, TraceCode.MERCHANT_CREATE_FAILED, "error", err, "merchant_id", merchantID)

// Trace codes defined in constants file
const (
    PAYMENT_PROCESSING_FAILED = "PAYMENT_PROCESSING_FAILED"
    MERCHANT_CREATE_FAILED = "MERCHANT_CREATE_FAILED"
    GATEWAY_TIMEOUT = "GATEWAY_TIMEOUT"
)
```

### Trace Code Naming Convention

**Pattern:** `<ENTITY>_<ACTION>_<STATUS>`

**Examples:**
- ✅ `PAYMENT_CREATE_FAILED`
- ✅ `TERMINAL_SYNC_TIMEOUT`
- ✅ `KAFKA_MESSAGE_PARSE_ERROR`
- ❌ `ERROR_OCCURRED` (too generic)
- ❌ `failed` (lowercase, no context)

### Detection Strategy

**Ask user for repo context:**
```
To validate trace codes, I need to see your logging patterns.
Can you share:
1. Path to TraceCode constants file (e.g., internal/tracecode/codes.go)
2. Or a file with existing logger.Error() calls
```

**Then analyze:**
```bash
# Find error logs without trace code constants
grep -n "logger\.Error\|logger\.Errorw" <changed_file> | while read line; do
    # Check if first argument is a constant (all caps with underscores)
    if ! echo "$line" | grep -q 'logger\.Error[w]*([^,]*, [A-Z_][A-Z_]*'; then
        FLAG: "Error log without trace code: $line"
    fi
done
```

### Flag Conditions

Flag if:
- `logger.Error()` with string literal instead of constant
- No trace code defined in constants file
- Trace code doesn't follow naming convention

### Severity

🚨 **Critical** - Log searchability broken, incident response slow

---

## Check 3: Success Metrics on Happy Path ⚠️ HIGH

### What to Check

Happy paths must emit success metrics (not just errors).

### Bad Pattern ❌

```go
// ANTI-PATTERN: Only error metrics
func CreateMerchant(ctx context.Context, req MerchantRequest) error {
    merchant, err := db.Create(req)
    if err != nil {
        metrics.Count(ctx, "merchant.create.failed", 1)  // ✅ Error metric
        return err
    }

    return nil  // ❌ No success metric!
}
```

**Problem:**
- Can't calculate success rate (success / total)
- Can't detect if feature usage is dropping
- Only see errors, not baseline traffic

### Good Pattern ✅

```go
// CORRECT: Metrics on both paths
func CreateMerchant(ctx context.Context, req MerchantRequest) error {
    merchant, err := db.Create(req)
    if err != nil {
        metrics.Count(ctx, "merchant.create.failed", 1, tags)
        return err
    }

    // ✅ Emit success metric
    metrics.Count(ctx, "merchant.create.success", 1, tags)
    return nil
}
```

### Detection Strategy

```bash
# Find functions with error metrics but no success metrics
grep -A 20 "^func.*error" <changed_file> | while read block; do
    if echo "$block" | grep -q "metrics.*failed\|metrics.*error"; then
        if ! echo "$block" | grep -q "metrics.*success\|metrics.*completed"; then
            FLAG: "Function has error metric but no success metric"
        fi
    fi
done
```

### Flag Conditions

Flag if:
- Function emits error metric
- No success metric in same function
- Function is critical path (payment, auth, onboarding)

### Severity

⚠️ **High** - Can't measure success rate, incomplete observability

---

## Check 4: Contextual Fields in Logs ⚠️ HIGH

### What to Check

Logs must include relevant context for debugging.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Generic log without context
logger.Error(ctx, TraceCode.PAYMENT_FAILED, "error", err)  // ❌ No payment_id, gateway, etc.
```

**Problem:**
- Can't correlate logs across services
- Can't filter by merchant/gateway
- Hard to debug specific failures

### Good Pattern ✅

```go
// CORRECT: Rich context
logger.Error(ctx, TraceCode.PAYMENT_FAILED,
    "error", err,
    "payment_id", payment.ID,           // ✅ Entity ID
    "merchant_id", payment.MerchantID,  // ✅ Parent entity
    "gateway", payment.Gateway,         // ✅ Integration point
    "amount", payment.Amount,           // ✅ Business context
    "error_code", err.Code,             // ✅ Error details
)
```

### Required Context Fields

**For error logs, include:**
- Entity ID (payment_id, merchant_id, terminal_id)
- Parent entity (merchant_id if logging payment)
- Integration point (gateway, bank, network)
- Error details (error_code, error_message)
- Business context (amount, currency, status)

### Detection Strategy

```bash
# Check if error logs include sufficient context fields
# Capture the full multi-line logger.Error call (up to closing paren)
grep -n "logger\.Error" <changed_file> | while read match; do
    line_num=$(echo "$match" | cut -d: -f1)
    # Collect 8 lines starting at the logger call (covers multi-line calls)
    call_block=$(sed -n "${line_num},$((line_num+8))p" <changed_file>)
    # Flag if no contextual key is present beyond "error"/"err".
    # Extract all quoted lowercase string keys from the call block, then check
    # whether any key other than "error"/"err" exists. This is open-ended —
    # any domain-specific field name (terminal_id, gateway, amount, source, etc.)
    # satisfies the check without maintaining a closed enumeration.
    extra_keys=$(echo "$call_block" | grep -oE '"[a-z][a-z0-9_]*"' | grep -vE '^"err(or)?"$')
    if [ -z "$extra_keys" ]; then
        FLAG: "Error log at line $line_num may be missing contextual fields beyond the error value"
    fi
done
```

> **Note:** Comma-counting on a single grep line is unreliable — logger calls often span multiple lines. The check looks for named context fields beyond `"error"` itself; `"*_id"`, `"gateway"`, `"amount"` etc. all satisfy it. If the log passes a variadic `fields...` slice, the check won't fire — flag only when a string literal key is absent.

### Flag Conditions

Flag if:
- Error log has only `"error"` / `"err"` as its sole named key (no domain context at all)
- No entity ID, integration point, or business field present in the call block

Do NOT flag if:
- Any non-error named key is present (even one field like `"terminal_id"` or `"gateway"` is sufficient)
- Logger call uses variadic `fields...` — content cannot be statically checked

### Severity

⚠️ **High** - Hard to debug, incident resolution slow

---

## Check 5: Avoid Info/Debug Logs in Critical Path 📋 MEDIUM

### What to Check

Critical paths should minimize info logs (prefer error/warn).

### Bad Pattern ❌

```go
// ANTI-PATTERN: Excessive info logging
func ProcessPayment(ctx context.Context, req PaymentRequest) error {
    logger.Info(ctx, "processing_payment_started", "payment_id", req.ID)  // ❌ Noisy

    logger.Info(ctx, "validating_payment", "payment_id", req.ID)  // ❌ Noisy
    if err := validate(req); err != nil {
        return err
    }

    logger.Info(ctx, "calling_gateway", "gateway", req.Gateway)  // ❌ Noisy
    resp, err := gateway.Charge(req)
    if err != nil {
        logger.Error(ctx, TraceCode.GATEWAY_CHARGE_FAILED, "error", err)
        return err
    }

    logger.Info(ctx, "payment_processed", "payment_id", req.ID)  // ❌ Noisy
    return nil
}
```

**Problem:**
- Logs polluted with info messages
- High log volume → increased costs
- Hard to find actual errors
- Performance impact (I/O)

### Good Pattern ✅

```go
// CORRECT: Only error logs + success metric
func ProcessPayment(ctx context.Context, req PaymentRequest) error {
    if err := validate(req); err != nil {
        logger.Error(ctx, TraceCode.PAYMENT_VALIDATION_FAILED, "error", err, "payment_id", req.ID)
        return err
    }

    resp, err := gateway.Charge(req)
    if err != nil {
        logger.Error(ctx, TraceCode.GATEWAY_CHARGE_FAILED,
            "error", err,
            "payment_id", req.ID,
            "gateway", req.Gateway)
        metrics.Count(ctx, "payment.failed", 1)
        return err
    }

    // ✅ Only metric for success (no info log)
    metrics.Count(ctx, "payment.success", 1)
    return nil
}
```

### When Info Logs Are OK

**Acceptable info logs:**
- Startup/shutdown events
- Configuration loading
- Background job start/completion
- Kafka message processing start (for tracing)

**NOT for:**
- Every function call
- Validation steps
- Business logic flow

### Detection Strategy

```bash
# Detect info log density per function using brace-depth tracking.
# Name-suffix matching (e.g., *Payment) misses functions like RouteToGateway.
# ^}$ as a boundary breaks on closures and nested structs.
# Instead: track open/close brace depth from every func declaration.
awk '
  /^func / {
      # Start tracking a new function: reset state, mark as not yet opened
      # Extract just the function name for readable output
      match($0, /func ([A-Za-z0-9_]+)/, arr)
      depth = 0; info_count = 0; fn_name = arr[1]; body_started = 0
  }
  fn_name != "" {
      # Count braces on this line to track nesting depth
      n = split($0, chars, "")
      for (i = 1; i <= n; i++) {
          if (chars[i] == "{") { depth++; body_started = 1 }
          if (chars[i] == "}") depth--
      }
  }
  /logger\.Info/ && fn_name != "" { info_count++ }
  # Only close the function when we have seen the opening brace (body_started)
  # and depth returns to 0. Guards against firing on the func declaration line
  # itself when the opening brace is on a separate line.
  body_started && depth == 0 && fn_name != "" {
      if (info_count > 3)
          printf "⚠️  %s has %d info logs — consider reducing in hot path\n", fn_name, info_count
      fn_name = ""; info_count = 0; body_started = 0
  }
' <changed_file>
```

> **Note:** This checks **all** functions, not just those with critical-path names. A function named `RouteToGateway` or `ProcessOrder` is just as critical as one named `ProcessPayment`. The 3-log threshold is per-function — a file with many single-log functions is fine.

### Flag Conditions

Flag if:
- More than 3 info logs **within a single critical-path function** (payment, auth, checkout, refund, transfer)
- Info logs inside a loop or per-item processing block
- Info log without business value (e.g., logging intermediate variables)

### Severity

📋 **Medium** - Log pollution, performance impact

---

## Repo Context Questions

**If no CLAUDE.md or monitoring patterns found, ask:**

```
To provide accurate monitoring/logging recommendations, I need your repo context:

1. **Metrics Library:**
   - How do you emit metrics? (e.g., metrics.Count(), prometheus.Inc())
   - Example file path with metrics?

2. **Trace Codes:**
   - Where are trace codes defined? (e.g., internal/tracecode/codes.go)
   - Example trace code constant?

3. **Logging Library:**
   - Logger syntax? (e.g., logger.Error(), log.Errorw())
   - Example error log line?

Without this context, I'll use generic patterns that may not match your codebase.
```

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | Error metrics on failures | 🚨 Critical | No alerting |
| 2 | Trace codes for errors | 🚨 Critical | Hard to search logs |
| 3 | Success metrics | ⚠️ High | Can't measure success rate |
| 4 | Contextual log fields | ⚠️ High | Hard to debug |
| 5 | Avoid info log spam | 📋 Medium | Log pollution |

---

## Example Output

```
📊 Observability Check

File: internal/services/payment_service.go

⚠️ Check #1 Failed: Missing error metric
Location: Line 145

Issue:
  logger.Error(ctx, "payment_failed", "error", err)
  // ❌ No metric emitted

Recommendation:
  metrics.Count(ctx, "payment.failed", 1, map[string]string{
      "gateway": payment.Gateway,
      "error_type": getErrorType(err),
  })

🚨 Check #2 Failed: No trace code
Location: Line 145

Issue:
  logger.Error(ctx, "payment failed", "error", err)  // ❌ String literal

Fix:
  logger.Error(ctx, TraceCode.PAYMENT_PROCESSING_FAILED, "error", err)

⚠️ Check #3 Failed: No success metric
Location: Line 156

Issue:
  Function has error metric but no success metric

Fix:
  metrics.Count(ctx, "payment.success", 1)
```
