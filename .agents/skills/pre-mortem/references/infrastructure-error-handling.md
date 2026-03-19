# Error Handling Best Practices Checks

## Overview

Validates error handling patterns to prevent panics, nil pointer dereferences, ignored errors, and poor error propagation.

**Load when:** PR modifies error handling, adds new error paths, or changes critical functions

**Total Checks:** 9

**Severity Distribution:**
- 🚨 Critical: 4
- ⚠️ High: 3
- 📋 Medium: 2

---

## Check 1: Panic Recovery in HTTP Handlers 🚨 CRITICAL

### What to Check

HTTP handlers must have panic recovery middleware to prevent service crashes.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No panic recovery
func HandleTerminalCreate(c *gin.Context) {
    var req TerminalRequest
    c.BindJSON(&req)

    // ❌ Panic here crashes entire service!
    terminal := processTerminal(req)  // Could panic on nil

    c.JSON(200, terminal)
}

func processTerminal(req TerminalRequest) *Terminal {
    // ❌ Panics if merchant not found
    merchant := merchants[req.MerchantID]  // nil map panic
    return &Terminal{
        MerchantName: merchant.Name,  // nil pointer panic
    }
}
```

**Problem:**
- Single panic crashes entire service
- All requests fail
- Manual restart required

### Good Pattern ✅

```go
// CORRECT: Panic recovery middleware
func RecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                // ✅ Recover from panic
                logger.Error(c, "panic_recovered",
                    "error", r,
                    "stack", string(debug.Stack()))

                c.JSON(500, gin.H{
                    "error": "internal server error",
                })
            }
        }()

        c.Next()
    }
}

// Apply middleware
func main() {
    router := gin.New()
    router.Use(RecoveryMiddleware())  // ✅ Catch all panics
    router.POST("/terminals", HandleTerminalCreate)
}

// PATTERN 2: Handler-level recovery
func HandleTerminalCreate(c *gin.Context) {
    defer func() {
        if r := recover(); r != nil {
            logger.Error(c, "handler_panic", "error", r)
            c.JSON(500, gin.H{"error": "internal error"})
        }
    }()

    var req TerminalRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    terminal, err := processTerminal(req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, terminal)
}
```

### Detection Strategy

```bash
# Find HTTP handler functions
grep -n "func Handle.*gin.Context" internal/handlers/*.go

# For each handler, check:
# 1. defer recover() exists in handler or middleware
# 2. Middleware registered in router setup
```

### Flag Conditions

Flag if:
- HTTP handler without `defer recover()`
- No recovery middleware in router
- Critical path (payments, auth) without recovery

### Severity

🚨 **Critical** - Service crashes, complete outage

### Reference

Based on terminals handlers

---

## Check 2: Nil Pointer Checks Before Dereference 🚨 CRITICAL

### What to Check

Pointers must be checked for nil before dereferencing.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No nil check
func GetTerminalMerchant(terminal *Terminal) string {
    // ❌ Panics if terminal is nil!
    return terminal.MerchantID
}

func ProcessPayment(payment *Payment) {
    // ❌ Panics if nested field is nil
    amount := payment.Details.Amount  // payment.Details could be nil
}

// ANTI-PATTERN: Partial nil check
func GetMerchantName(merchantID string) string {
    merchant, _ := repo.FindMerchant(merchantID)  // Can return nil

    // ❌ No nil check before access
    return merchant.Name  // Panic if merchant is nil!
}
```

**Problem:**
- Nil pointer panics crash service
- Unpredictable behavior
- Poor error messages

### Good Pattern ✅

```go
// CORRECT: Nil checks before access
func GetTerminalMerchant(terminal *Terminal) (string, error) {
    // ✅ Check nil pointer
    if terminal == nil {
        return "", errors.New("terminal is nil")
    }

    return terminal.MerchantID, nil
}

func ProcessPayment(payment *Payment) error {
    // ✅ Check all levels
    if payment == nil {
        return errors.New("payment is nil")
    }

    if payment.Details == nil {
        return errors.New("payment details missing")
    }

    amount := payment.Details.Amount
    // Process amount...
    return nil
}

// CORRECT: Safe navigation
func GetMerchantName(merchantID string) (string, error) {
    merchant, err := repo.FindMerchant(merchantID)
    if err != nil {
        return "", fmt.Errorf("merchant not found: %w", err)
    }

    // ✅ Nil check after retrieval
    if merchant == nil {
        return "", errors.New("merchant is nil")
    }

    return merchant.Name, nil
}

// PATTERN 2: Safe accessor methods
func (t *Terminal) GetMerchantID() string {
    if t == nil {
        return ""  // ✅ Safe default for nil receiver
    }
    return t.MerchantID
}
```

### Severity

🚨 **Critical** - Service crashes, nil pointer panics

---

## Check 3: Error Returns Not Ignored 🚨 CRITICAL

### What to Check

Function errors must be checked, not ignored with `_`.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Ignored errors
func CreateTerminal(req TerminalRequest) {
    terminal := &Terminal{
        ID:         utils.NewID(),
        MerchantID: req.MerchantID,
    }

    repo.Save(terminal)  // ❌ Error ignored - database failure silent!

    cache.Set(terminal.ID, terminal)  // ❌ Cache failure ignored

    eventBus.Publish("terminal.created", terminal)  // ❌ Event failure ignored
}

// ANTI-PATTERN: Error assigned but not checked
func UpdateTerminal(terminal *Terminal) {
    err := repo.Update(terminal)
    // ❌ err assigned but never checked
    logger.Info("terminal_updated")
}
```

**Problem:**
- Database failures go unnoticed
- Partial state persisted
- Data inconsistency

### Good Pattern ✅

```go
// CORRECT: All errors checked
func CreateTerminal(req TerminalRequest) error {
    terminal := &Terminal{
        ID:         utils.NewID(),
        MerchantID: req.MerchantID,
    }

    // ✅ Check database error
    if err := repo.Save(terminal); err != nil {
        logger.Error(ctx, "terminal_save_failed", "error", err)
        return fmt.Errorf("failed to save terminal: %w", err)
    }

    // ✅ Cache failure logged but not returned (non-critical)
    if err := cache.Set(terminal.ID, terminal); err != nil {
        logger.Warn(ctx, "cache_set_failed", "error", err)
        // Continue - cache is non-critical
    }

    // ✅ Event publishing error logged
    if err := eventBus.Publish("terminal.created", terminal); err != nil {
        logger.Error(ctx, "event_publish_failed", "error", err)
        // Don't fail request, but log for investigation
    }

    return nil
}

// PATTERN 2: Must pattern for critical operations
func MustCreateIndex(db *gorm.DB, indexName string) {
    if err := db.Exec(fmt.Sprintf("CREATE INDEX %s...", indexName)).Error; err != nil {
        // ✅ Panic on startup errors is acceptable
        panic(fmt.Sprintf("failed to create index %s: %v", indexName, err))
    }
}
```

### Detection Strategy

```bash
# Find function calls with ignored errors
# Look for:
# 1. Function calls without assignment: repo.Save(...)
# 2. Blank identifier: _, err := ...  (but err never used)
# 3. err assigned but not checked in function body
```

### Severity

🚨 **Critical** - Silent failures, data loss

---

## Check 4: Type Assertions with Safety Check 🚨 CRITICAL

### What to Check

Type assertions must use the `, ok` pattern to avoid panics.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Unsafe type assertion
func ProcessEvent(event interface{}) {
    // ❌ Panics if event is not TerminalEvent
    terminalEvent := event.(TerminalEvent)

    processTerminal(terminalEvent.TerminalID)
}

// ANTI-PATTERN: Map value type assertion
func GetMetadata(data map[string]interface{}) {
    // ❌ Panics if merchant_id is not string
    merchantID := data["merchant_id"].(string)

    // ❌ Panics if amount is not float64
    amount := data["amount"].(float64)
}
```

**Problem:**
- Type assertion panic crashes service
- Unexpected data types cause failures

### Good Pattern ✅

```go
// CORRECT: Safe type assertion
func ProcessEvent(event interface{}) error {
    // ✅ Safe type assertion with ok check
    terminalEvent, ok := event.(TerminalEvent)
    if !ok {
        logger.Error(ctx, "unexpected_event_type",
            "expected", "TerminalEvent",
            "got", fmt.Sprintf("%T", event))
        return fmt.Errorf("expected TerminalEvent, got %T", event)
    }

    return processTerminal(terminalEvent.TerminalID)
}

// CORRECT: Safe map value access
func GetMetadata(data map[string]interface{}) error {
    // ✅ Check key exists and type is correct
    merchantIDRaw, exists := data["merchant_id"]
    if !exists {
        return errors.New("merchant_id missing")
    }

    merchantID, ok := merchantIDRaw.(string)
    if !ok {
        return fmt.Errorf("merchant_id must be string, got %T", merchantIDRaw)
    }

    // ✅ Safe numeric conversion with validation
    amountRaw, exists := data["amount"]
    if !exists {
        return errors.New("amount missing")
    }

    var amount float64
    switch v := amountRaw.(type) {
    case float64:
        amount = v
    case int:
        amount = float64(v)
    case string:
        parsed, err := strconv.ParseFloat(v, 64)
        if err != nil {
            return fmt.Errorf("invalid amount: %w", err)
        }
        amount = parsed
    default:
        return fmt.Errorf("amount must be number, got %T", v)
    }

    return processPayment(merchantID, amount)
}
```

### Severity

🚨 **Critical** - Type assertion panics

---

## Check 5: Error Wrapping for Context ⚠️ HIGH

### What to Check

Errors should be wrapped with context using `fmt.Errorf` with `%w`.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No context in error
func CreateTerminal(req TerminalRequest) error {
    if err := validateRequest(req); err != nil {
        return err  // ❌ Lost context of where error happened
    }

    if err := repo.Save(terminal); err != nil {
        return err  // ❌ Can't tell if error from validation or save
    }

    return nil
}

// ANTI-PATTERN: String concatenation loses error chain
func ProcessPayment(payment *Payment) error {
    if err := gateway.Charge(payment); err != nil {
        // ❌ Loses original error, can't use errors.Is()
        return errors.New("payment failed: " + err.Error())
    }
    return nil
}
```

**Problem:**
- Lost error context
- Can't use `errors.Is()` or `errors.As()`
- Difficult debugging

### Good Pattern ✅

```go
// CORRECT: Wrap errors with context
func CreateTerminal(req TerminalRequest) error {
    if err := validateRequest(req); err != nil {
        // ✅ Wrap with context
        return fmt.Errorf("validation failed: %w", err)
    }

    terminal := buildTerminal(req)

    if err := repo.Save(terminal); err != nil {
        // ✅ Add context about what operation failed
        return fmt.Errorf("failed to save terminal %s: %w", terminal.ID, err)
    }

    return nil
}

// CORRECT: Preserve error chain
func ProcessPayment(payment *Payment) error {
    if err := gateway.Charge(payment); err != nil {
        // ✅ Wrapping preserves error chain
        return fmt.Errorf("gateway charge failed for %s: %w", payment.ID, err)
    }
    return nil
}

// Can now use errors.Is()
if err := ProcessPayment(payment); err != nil {
    if errors.Is(err, ErrInsufficientFunds) {
        // Handle specific error
    }
}
```

### Severity

⚠️ **High** - Poor error context, difficult debugging

---

## Check 6: Custom Error Types for Domain Errors ⚠️ HIGH

### What to Check

Domain-specific errors should use custom types for better handling.

### Bad Pattern ❌

```go
// ANTI-PATTERN: String errors
func FindTerminal(id string) (*Terminal, error) {
    terminal, err := repo.Get(id)
    if err != nil {
        // ❌ String comparison needed to detect not found
        return nil, errors.New("terminal not found")
    }
    return terminal, nil
}

// Caller has to parse error string
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        // ❌ Fragile string matching
    }
}
```

### Good Pattern ✅

```go
// CORRECT: Custom error types
var (
    ErrTerminalNotFound     = errors.New("terminal not found")
    ErrInvalidTerminalID    = errors.New("invalid terminal ID")
    ErrTerminalAlreadyExists = errors.New("terminal already exists")
    ErrInsufficientFunds    = errors.New("insufficient funds")
)

// Or structured error type
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}

// Usage
func FindTerminal(id string) (*Terminal, error) {
    if id == "" {
        return nil, ErrInvalidTerminalID
    }

    terminal, err := repo.Get(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // ✅ Return domain error
            return nil, ErrTerminalNotFound
        }
        return nil, fmt.Errorf("database error: %w", err)
    }

    return terminal, nil
}

// Caller can use errors.Is()
terminal, err := FindTerminal(id)
if err != nil {
    if errors.Is(err, ErrTerminalNotFound) {
        // ✅ Type-safe error handling
        return c.JSON(404, gin.H{"error": "terminal not found"})
    }
    return c.JSON(500, gin.H{"error": "internal error"})
}
```

### Severity

⚠️ **High** - Fragile error handling, poor API contracts

---

## Check 7: Logging Errors Before Returning 📋 MEDIUM

### What to Check

Errors should be logged at the point they occur with context.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No logging
func ProcessPayment(payment *Payment) error {
    if err := gateway.Charge(payment); err != nil {
        return err  // ❌ Error returned but not logged
    }
    return nil
}

// Caller also doesn't log
func HandlePayment(c *gin.Context) {
    if err := ProcessPayment(payment); err != nil {
        c.JSON(500, gin.H{"error": "payment failed"})  // ❌ No log with context
    }
}
```

**Problem:**
- No visibility into errors
- Can't debug production issues
- Lost request context

### Good Pattern ✅

```go
// CORRECT: Log errors with context
func ProcessPayment(ctx *gin.Context, payment *Payment) error {
    logger.Info(ctx, "processing_payment",
        "payment_id", payment.ID,
        "amount", payment.Amount)

    if err := gateway.Charge(payment); err != nil {
        // ✅ Log error with full context
        logger.Error(ctx, "gateway_charge_failed",
            "payment_id", payment.ID,
            "amount", payment.Amount,
            "gateway", payment.Gateway,
            "error", err)

        return fmt.Errorf("gateway charge failed: %w", err)
    }

    logger.Info(ctx, "payment_processed",
        "payment_id", payment.ID)

    return nil
}

// Handler logs at its level too
func HandlePayment(c *gin.Context) {
    if err := ProcessPayment(c, payment); err != nil {
        logger.Error(c, "payment_handler_failed",
            "request_id", c.GetString("request_id"),
            "error", err)

        c.JSON(500, gin.H{"error": "payment failed"})
    }
}
```

### Severity

📋 **Medium** - Poor observability, difficult debugging

---

## Check 8: Defer for Cleanup Resources 📋 MEDIUM

### What to Check

Resource cleanup (file close, connection close) must use `defer`.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No defer for cleanup
func ProcessFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }

    data, err := readData(file)
    if err != nil {
        // ❌ File not closed on error!
        return err
    }

    file.Close()  // ❌ Doesn't run if readData panics
    return processData(data)
}

// ANTI-PATTERN: Manual transaction rollback
func UpdateTerminal(terminal *Terminal) error {
    tx := db.Begin()

    if err := tx.Save(terminal).Error; err != nil {
        tx.Rollback()  // ❌ Must remember to rollback
        return err
    }

    if err := updateCache(terminal); err != nil {
        // ❌ Forgot to rollback here!
        return err
    }

    return tx.Commit().Error
}
```

**Problem:**
- Resource leaks (file descriptors, connections)
- Inconsistent cleanup
- Easy to forget cleanup on all paths

### Good Pattern ✅

```go
// CORRECT: Use defer for cleanup
func ProcessFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()  // ✅ Always closes, even on panic

    data, err := readData(file)
    if err != nil {
        return err  // ✅ file.Close() still runs
    }

    return processData(data)  // ✅ file.Close() still runs
}

// CORRECT: Defer rollback for transactions
func UpdateTerminal(terminal *Terminal) error {
    tx := db.Begin()
    defer tx.Rollback()  // ✅ Always rolls back unless Commit succeeds

    if err := tx.Save(terminal).Error; err != nil {
        return err  // ✅ Rollback happens automatically
    }

    if err := updateCache(terminal); err != nil {
        return err  // ✅ Rollback happens automatically
    }

    return tx.Commit().Error  // ✅ Commit overrides rollback
}

// CORRECT: Defer with error check
func ProcessFileWithDeferCheck(filename string) (err error) {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }

    defer func() {
        // ✅ Check close error too
        if closeErr := file.Close(); closeErr != nil && err == nil {
            err = fmt.Errorf("failed to close file: %w", closeErr)
        }
    }()

    return readData(file)
}
```

### Severity

📋 **Medium** - Resource leaks, transaction issues

---

## Check 9: Merge/Patch Function — Comment-Implementation Contract ⚠️ HIGH

### What to Check

Functions that claim to "merge", "partial update", or "preserve existing values for unset fields" must implement that semantics for **all** fields they claim to preserve — not silently only for pointer fields.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Comment overpromises what the function actually preserves

// mergeMerchantBankData merges existing database data with new request data.
// New values take precedence; unset optional fields preserve existing values. ← claims ALL optional fields
func mergeMerchantBankData(existing, newData *MerchantBank) *MerchantBank {
    merged := *existing

    merged.Netbanking = newData.Netbanking  // ❌ ALWAYS overwrites — even 0 (= "disabled")
    merged.Upi        = newData.Upi         // ❌ ALWAYS overwrites — caller can't signal "not provided"
    // ... 40+ more value-type fields, all unconditionally overwritten ...

    if newData.Card != nil {                // ✅ ONLY Card actually gets "preserve if unset" logic
        merged.Card = newData.Card          //    because Card is *int32 — nil = "not provided"
    }
    return &merged
}
```

**Problem:**
- Comment says "unset optional fields preserve existing values" — reader assumes *all* omitted fields are safe
- Only pointer fields (`*int32`, `*string`, etc.) can structurally express "not provided" vs "set to zero"
- Value-type fields (`int32`, `bool`) cannot distinguish "caller sent 0/false" from "caller omitted this field"
- Any value-type field omitted by the caller will silently overwrite the existing DB value with its zero value
- The bug the PR was trying to fix for `Card` still exists for every other non-pointer field

### Good Pattern ✅

```go
// CORRECT: Comment precisely describes which fields get preservation and why

// mergeMerchantBankData merges existing database data with new request data.
// Pointer fields (e.g. Card *int32): if nil in newData, existing value is preserved —
//   nil unambiguously means "caller did not provide this field".
// Value fields (e.g. Netbanking int32): newData ALWAYS wins — zero is a valid value ("disabled"),
//   indistinguishable from "not provided". To add selective preservation, use pointer types.
func mergeMerchantBankData(existing, newData *MerchantBank) *MerchantBank {
    merged := *existing

    merged.Netbanking = newData.Netbanking  // value type: always overwrites (documented above)
    // ...

    if newData.Card != nil {                // pointer type: nil = not provided, preserve existing
        merged.Card = newData.Card
    }
    return &merged
}
```

### Detection Strategy

The trigger must be **code structure first, not function name or comment**.
Function names and comments can be wrong (that is exactly the bug this check catches).
The code's own assignment pattern is the reliable signal.

**Signal 1 — Same-type merge signature (structural trigger):**
```bash
# Both parameters are the same type and the return is that type — classic merge shape
# Pattern: func <name>(<varA> *<Type>, <varB> *<Type>) *<Type>
grep -n "func \w\+(\w\+ \*\(\w\+\), \w\+ \*\1) \*\1" <changed_file>
```
This distinguishes a merge (same type in, same type out) from a builder/formatter
(different input and output types), which would be a false positive.

**Signal 2 — Mixed guard pattern inside the function body:**
```bash
# Unconditional bulk assignments from the second parameter
UNCONDITIONAL=$(grep -c "\bmerged\.\w\+ = \w\+\.\w\+" <function_body>)

# Nil-guarded assignments from the second parameter
GUARDED=$(grep -c "if \w\+\.\w\+ != nil" <function_body>)

# Both > 0 means: the author knew some fields need guarding.
# Every unconditional value-type assignment is now suspect.
if [ "$UNCONDITIONAL" -gt 0 ] && [ "$GUARDED" -gt 0 ]; then
    FLAG: "Mixed guard pattern: $GUARDED fields nil-guarded, $UNCONDITIONAL always overwritten"
fi
```

Why this works: the presence of even **one** nil-guard is proof the author intended selective
preservation. Any unconditional assignment of a value-type field in the same function is
potentially a field where "zero value = disabled" is indistinguishable from "not provided."

**Signal 3 — Comment check (secondary, boosts severity if triggered):**
```bash
# Only used to escalate, not to trigger
grep -n "preserve\|optional\|unset\|merge\|partial" <function_comment>
```
If Signal 1 + 2 fire AND the comment claims "preserve unset fields" without qualification,
raise severity. If only 1 + 2 fire with no comment, still flag but as Medium (needs human review).

**Limitation — cannot resolve without proto/domain context:**
Whether `Netbanking = 0` means "caller disabled it" or "caller omitted it" is only knowable from
the proto definition (`optional int32` → pointer → can be nil vs `int32` → zero is a real value).
Static analysis cannot determine this. Always raise the finding anyway with a specific question
the PR author must answer before merge — do not silently skip it.

### Flag Conditions

Flag when **Signal 1 AND Signal 2** both hold:

1. **Function signature**: two parameters of the same struct type, returns the same struct type
2. **Mixed guard pattern**: at least one nil-guarded field assignment AND at least one
   unconditional value-type field assignment from the second parameter in the same function body

Escalate to High (from Medium) if additionally:
- Function comment says "preserve unset/optional fields" without specifying which field types benefit
- OR the struct has pointer fields (`*int32`, `*string`) mixed with value-type fields on
  the same entity — signalling the codebase uses pointers deliberately for optionality

Do **not** flag:
- Builder functions where input and output types differ (different-type signature)
- Functions where ALL fields are guarded (correct selective merge)
- Functions where NO fields are guarded (intentional full overwrite — no mixed intent)

### Required Author Verification

Always surface this finding with the following explicit question, regardless of whether
a comment mismatch was detected. The author must answer it before merge:

> **For each unconditional value-type field assignment in this function (`merged.X = newData.X`
> where X is `int32`, `bool`, `string`, etc.): can the caller express "I did not provide this field"
> differently from "I set this field to zero/false/empty"?**
>
> - If **no** (proto field is non-optional, zero IS a real value the caller can send): the unconditional
>   assignment is correct. Add a comment explaining this so future reviewers don't re-raise it.
> - If **yes** (proto field is `optional`, or the call path always sets every field): the unconditional
>   assignment is a latent bug identical to the one this PR may have just fixed for pointer fields.
>   Either make the field a pointer type, or add a nil/sentinel guard.

This question cannot be answered by static analysis alone — it requires checking the proto
definition and the call sites. The PR author is the right person to verify it.

### Severity

⚠️ **High** (mixed guard pattern + comment overpromise)
📋 **Medium** (mixed guard pattern only, no comment claim — but author verification still required)

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | Panic recovery | 🚨 Critical | Service crashes |
| 2 | Nil pointer checks | 🚨 Critical | Panics |
| 3 | Error returns checked | 🚨 Critical | Silent failures |
| 4 | Safe type assertions | 🚨 Critical | Type panics |
| 5 | Error wrapping | ⚠️ High | Poor debugging |
| 6 | Custom error types | ⚠️ High | Fragile handling |
| 9 | Merge/patch comment contract | ⚠️ High | Silent data corruption |
| 7 | Error logging | 📋 Medium | Poor visibility |
| 8 | Defer cleanup | 📋 Medium | Resource leaks |

---

## How to Apply

**For each file:**

1. Check HTTP handlers have panic recovery
2. Verify nil checks before pointer dereference
3. Check all function errors are handled
4. Verify type assertions use `, ok` pattern
5. Check errors wrapped with `%w`
6. Look for custom error types
7. Verify errors logged with context
8. Check defer used for cleanup
9. For merge/patch functions: verify comment matches which fields are actually guarded

**Example output:**

```
📁 File: internal/services/terminal_service.go

🚨 Check #1 Failed: No panic recovery (Line 23)
   Code: func HandleCreate(c *gin.Context) {...}
   Fix: Add defer recover() or use recovery middleware

🚨 Check #3 Failed: Error ignored (Line 67)
   Code: repo.Save(terminal)
   Fix: if err := repo.Save(terminal); err != nil { return err }

⚠️  Check #5 Failed: Error not wrapped (Line 89)
   Code: return err
   Fix: return fmt.Errorf("save failed: %w", err)

✅ Check #2 Passed: Nil checks present
✅ Check #4 Passed: Type assertions safe
✅ Check #8 Passed: Defer used for cleanup
```
