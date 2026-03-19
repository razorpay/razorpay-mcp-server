# Goroutine Panic Recovery Integration - Complete ✅

## Summary

Successfully integrated comprehensive goroutine panic recovery checks into the pre-mortem skill, based on patterns from the `go-panic-recovery` skill (PR #310).

---

## What Was Added

### New Check: Goroutine Panic Recovery (Check #9) 🚨 CRITICAL

**File:** `references/infrastructure-error-handling.md`

**Comprehensive coverage beyond existing HTTP handler panic checks:**

#### Before (Existing Coverage)
- ✅ Check #1: Panic recovery in HTTP handlers (middleware-based)
- ✅ Kafka consumer panic recovery (infrastructure-kafka.md Check #1)
- ✅ Type assertion safety (infrastructure-error-handling.md Check #4)

#### After (New Coverage)
- ✨ Check #9: **All spawned goroutines** (fire-and-forget, parallel, workers)

---

## Key Insights Encoded

### Critical Difference
**HTTP middleware CANNOT protect spawned goroutines!**

```go
// ✅ Protected by middleware
func Handler(c *gin.Context) {
    // Panic here caught by RecoveryMiddleware()
}

// ❌ NOT protected by middleware
go func() {
    // Panic here CRASHES entire service!
}()
```

### Goroutine Isolation
Each goroutine must protect itself - parent context cannot help:

```go
// WRONG - parent recover doesn't help child
defer func() { recover() }()
go func() { panic("oops") }()  // NOT RECOVERED ❌

// CORRECT - recover in same goroutine
go func() {
    defer func() { recover() }()
    panic("oops")  // Recovered ✅
}()
```

---

## Patterns Covered

### 1. Fire-and-Forget Goroutines
```go
// Cache invalidation, background notifications, async metrics
go func() {
    defer func() {
        if r := recover(); r != nil {
            logger.Error(ctx, "PANIC_CACHE_INVALIDATION", "panic", r)
        }
    }()
    deleteCache(ctx)
}()
```

### 2. Parallel Processing (WaitGroup)
```go
// Fetching from multiple services, batch processing
var wg sync.WaitGroup
var mu sync.Mutex
var err error

wg.Add(1)
go func() {
    defer wg.Done()
    defer func() {
        if r := recover(); r != nil {
            logger.Error(ctx, "PANIC_FETCH_DATA", "panic", r)
            mu.Lock()
            err = fmt.Errorf("panic: %v", r)
            mu.Unlock()
        }
    }()
    // Work...
}()
```

### 3. Long-Running Workers
```go
// Server lifecycle, event listeners, queue consumers
go func() {
    defer func() {
        if r := recover(); r != nil {
            logger.Error(ctx, "PANIC_METRICS_SERVER", "panic", r, "stack", string(debug.Stack()))
        }
    }()
    metricsServer.Run(ctx)
}()
```

### 4. Error Channel Pattern
```go
errChan := make(chan error, numWorkers)

go func(id int) {
    defer func() {
        if r := recover(); r != nil {
            logger.Error(ctx, "PANIC_WORKER", "panic", r, "worker_id", id)
            errChan <- fmt.Errorf("worker %d panicked: %v", id, r)
        }
    }()
    // Work...
    errChan <- nil
}(i)
```

### 5. Reusable Wrapper
```go
func SafeGo(ctx context.Context, name string, fn func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                logger.Error(ctx, name, "panic", r, "stack", string(debug.Stack()))
            }
        }()
        fn()
    }()
}

// Usage
SafeGo(ctx, "PANIC_SEND_NOTIFICATION", func() {
    sendNotification(userID, message)
})
```

---

## Runtime Panic Scenarios Protected Against

1. **Nil pointer dereference** - `c.Cache.Get("key")` when Cache is nil
2. **Nil map assignment** - `m["key"] = 42` on nil map
3. **Out of bounds access** - `slice[5]` on length-3 slice
4. **Type assertion without check** - `value := x.(string)` when x is not string
5. **Send on closed channel** - `ch <- 42` after `close(ch)`
6. **Integer divide by zero** - `x / 0` for integers
7. **Third-party library panics** - Unexpected data formats, API violations

---

## Detection Strategy

```bash
# Find all goroutine launches
grep -n "go func\|go [a-zA-Z]" <pr_files> --include="*.go" --exclude="*_test.go"

# Check for defer recover() pattern
grep -A 10 "go func" <pr_files> | grep "defer.*recover"

# Flag if missing
```

### Flag Conditions

- `go func()` without `defer func() { recover() }` in first 5 lines
- `go someFunction()` where function lacks recovery
- Fire-and-forget goroutine (no WaitGroup, no error channel)
- Parallel processing with WaitGroup but no panic recovery
- Server lifecycle goroutine without recovery

---

## Common Pitfalls Documented

### Pitfall 1: recover() in Wrong Goroutine
```go
// WRONG
defer func() { recover() }()
go func() { panic("oops") }()  // NOT RECOVERED
```

### Pitfall 2: recover() Outside defer
```go
// WRONG - must be in defer
func handlePanic() { recover() }
defer handlePanic()  // Won't work
```

### Pitfall 3: Silently Swallowing Panics
```go
// BAD - no logging
defer func() { recover() }()

// GOOD - log with context
defer func() {
    if r := recover(); r != nil {
        logger.Error(ctx, "PANIC_NAME", "panic", r, "stack", string(debug.Stack()))
    }
}()
```

---

## Best Practices

1. **Every spawned goroutine needs recovery**
   - Fire-and-forget: Log panic
   - WaitGroup: Propagate error via mutex/channel
   - Long-running: Log panic, don't crash main service

2. **Use descriptive panic log names**
   - ✅ `PANIC_CACHE_INVALIDATION`
   - ✅ `PANIC_FETCH_MERCHANT_INFO`
   - ❌ `panic_occurred`

3. **Include context in panic logs**
   - merchant_id, request_id, gateway, etc.
   - Stack trace: `string(debug.Stack())`

4. **Test panic recovery**
   ```go
   func TestPanicRecovery(t *testing.T) {
       done := make(chan bool)
       go func() {
           defer func() {
               if r := recover(); r != nil {
                   done <- true
               }
           }()
           panic("test panic")
       }()
       <-done  // Should complete, not crash
   }
   ```

---

## Integration with go-panic-recovery Skill

### Complementary Purposes

**go-panic-recovery skill** (observability category):
- **Purpose:** Find and fix ALL unprotected goroutines in entire codebase
- **When:** User requests "add panic recovery", "scan for unsafe goroutines"
- **Scope:** Entire codebase analysis
- **Output:** Fixes applied to all goroutines + PR created
- **Use case:** One-time codebase hardening

**pre-mortem Check #9** (development category):
- **Purpose:** Validate PRs don't introduce NEW unprotected goroutines
- **When:** PR is created/updated
- **Scope:** Only changed files in PR
- **Output:** Flag violations, block merge
- **Use case:** Continuous validation in PR workflow

### Workflow Together

1. **Initial cleanup** → Use `go-panic-recovery` skill to fix entire codebase
2. **Ongoing validation** → Pre-mortem Check #9 prevents regression in new PRs

---

## Updated Stats

### Before
- **Total Checks:** 112
- **Infrastructure Checks:** 86
- **Error Handling Checks:** 8

### After
- **Total Checks:** 113 (+1)
- **Infrastructure Checks:** 87 (+1)
- **Error Handling Checks:** 9 (+1)
- **Critical Checks:** +1 (goroutine panic recovery)

---

## Files Modified

1. ✅ `references/infrastructure-error-handling.md` - Added Check #9
2. ✅ `SKILL.md` - Updated total checks (112 → 113)
3. ✅ `README.md` - Updated infrastructure and error handling counts
4. ✅ `PANIC_RECOVERY_INTEGRATION.md` - This summary

---

## Example Output

```
📁 File: internal/services/merchant_service.go

🚨 Check #9 Failed: Goroutine without panic recovery (Line 45)
   Code: go fetchMerchantData(merchantID)
   Issue: Spawned goroutine lacks defer recover()
   Risk: Type assertion panic crashes entire service!

   Fix: Wrap with panic recovery:
   go func() {
       defer func() {
           if r := recover(); r != nil {
               logger.Error(ctx, "PANIC_FETCH_MERCHANT",
                   "panic", r,
                   "merchant_id", merchantID)
           }
       }()
       fetchMerchantData(merchantID)
   }()

   Reference: infrastructure-error-handling.md #9

🚨 Check #9 Failed: WaitGroup goroutine without panic recovery (Line 67)
   Code: go func() { defer wg.Done(); processItem(item) }()
   Issue: Panic blocks WaitGroup, causes deadlock
   Risk: All parallel processing stops!

   Fix: Add panic recovery before wg.Done()

✅ Check #1 Passed: HTTP handlers have recovery middleware
✅ Check #2 Passed: Nil checks present
```

---

## Severity Justification

🚨 **Critical** - Goroutine panics:
- Crash entire process (not just request)
- All services unavailable
- HTTP middleware cannot help
- Manual restart required
- Production outages

**More severe than HTTP handler panics** because:
- HTTP middleware catches handler panics → 500 error, service continues
- Goroutine panics → entire process dies, all requests fail

---

## References

**Source patterns:**
- `go-panic-recovery` skill (PR #310, merged)
- `go-panic-recovery/references/panic-patterns.md` - Comprehensive Go panic scenarios
- Razorpay production patterns:
  - Cache invalidation goroutines
  - Server lifecycle (HTTP, gRPC, metrics)
  - Parallel merchant data fetching
  - Background notification workers

---

## Testing Recommendations

1. **Unit tests** - Verify panic recovery works
2. **Integration tests** - Test goroutine panic scenarios
3. **Load tests** - Ensure no performance impact
4. **Manual verification** - Check production goroutines have recovery

---

## Next Steps

1. ✅ Integration complete
2. ✅ Documentation updated
3. 📝 Test on real PRs with goroutines
4. 📝 Validate detection accuracy
5. 📝 Gather feedback from engineers

---

## Success Criteria Met

1. ✅ Comprehensive goroutine panic coverage added
2. ✅ All goroutine patterns documented (fire-and-forget, parallel, workers)
3. ✅ Runtime panic scenarios listed
4. ✅ Detection strategy defined
5. ✅ Best practices included
6. ✅ Integration with go-panic-recovery skill explained
7. ✅ Updated all documentation (SKILL.md, README.md)
8. ✅ Example outputs provided

---

**Pre-mortem skill now provides complete panic protection for Go services!** 🎉
