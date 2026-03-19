# Go Code Review - Quick Reference Guide

This is a quick reference for the most common issues to check during Go code reviews.

## ⚠️ Critical Issues Checklist

These issues **MUST** be caught and fixed:

### 1. Transaction Context Usage

```go
// ❌ CRITICAL BUG
func (r *Repo) Update(ctx context.Context) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        r.Get(ctx, id)  // Using ctx instead of tctx!
    })
}

// ✅ CORRECT
func (r *Repo) Update(ctx context.Context) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        r.Get(tctx, id)  // Using tctx
    })
}
```

**Check**: Every DB operation inside a transaction uses `tctx`, not `ctx`

### 2. Goroutine Leaks

```go
// ❌ Goroutine leak
ch := make(chan int)
ch <- 1  // Blocks forever

// ✅ Correct
ch := make(chan int, 1)  // Buffered
ch <- 1
```

**Check**: Unbuffered channels with potential blocking

### 3. Race Conditions

```go
// ❌ Race condition
var counter int
go func() { counter++ }()
go func() { counter++ }()

// ✅ Correct
var counter atomic.Int64
go func() { counter.Add(1) }()
go func() { counter.Add(1) }()
```

**Check**: Shared variables accessed by multiple goroutines without protection

### 4. Resource Leaks

```go
// ❌ File not closed
f, err := os.Open(name)
// ... forget to close

// ✅ Correct
f, err := os.Open(name)
if err != nil {
    return err
}
defer f.Close()
```

**Check**: Files, connections, locks are properly closed/released

## 🔍 Important Issues Checklist

### Error Handling

```go
// ❌ Not wrapped
return err

// ✅ Wrapped
return fmt.Errorf("failed to process payment %s: %w", id, err)
```

**Check**: Errors wrapped with context using `%w`

### Context Propagation

```go
// ❌ Context not passed
func Process(ctx context.Context) {
    doWork()  // Missing ctx
}

// ✅ Correct
func Process(ctx context.Context) {
    doWork(ctx)
}
```

**Check**: Context passed to all downstream calls

### Mutex Protection

```go
// ❌ Not all accesses protected
func (c *Cache) Get() {
    return c.data[key]  // No lock!
}

// ✅ All accesses protected
func (c *Cache) Get() {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}
```

**Check**: All accesses to shared data are protected

## 💡 Performance Checklist

### Allocations

```go
// ❌ Multiple allocations
var result []int
for _, item := range items {
    result = append(result, process(item))
}

// ✅ Pre-allocated
result := make([]int, 0, len(items))
for _, item := range items {
    result = append(result, process(item))
}
```

**Check**: Slices/maps pre-allocated when size is known

### String Building

```go
// ❌ Inefficient
s := ""
for _, part := range parts {
    s += part
}

// ✅ Efficient
var b strings.Builder
for _, part := range parts {
    b.WriteString(part)
}
s := b.String()
```

**Check**: Use `strings.Builder` for concatenation in loops

### Unnecessary Goroutines

```go
// ❌ Overkill
go func() {
    result := simpleCalc()
    ch <- result
}()

// ✅ Direct call
result := simpleCalc()
```

**Check**: Goroutines only when needed for concurrency

## 🧪 Testing Checklist

### Table-Driven Tests

```go
// ✅ Good pattern
tests := []struct {
    name    string
    input   Input
    want    Output
    wantErr bool
}{
    {name: "valid", input: valid, want: output, wantErr: false},
    {name: "invalid", input: invalid, wantErr: true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := Function(tt.input)
        if (err != nil) != tt.wantErr {
            t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
        }
    })
}
```

**Check**: Multiple test cases use table-driven pattern

### Error Cases Tested

```go
tests := []struct{...}{
    {name: "success", ...},
    {name: "not found error", ...},
    {name: "validation error", ...},
    {name: "db error", ...},
}
```

**Check**: All error paths have test cases

## 🎯 Quick Scan Pattern

When reviewing a Go PR, scan in this order:

1. **Search for `Transaction(`** → Check all uses of context inside
2. **Search for `go func(`** → Check for leaks, race conditions
3. **Search for `make(chan`** → Check buffer size, closing
4. **Search for `defer`** → Check resource cleanup
5. **Search for `return err`** → Check error wrapping
6. **Search for `sync.Mutex`** → Check all accesses protected
7. **Search for `append(`** → Check pre-allocation
8. **Check test files** → Coverage of new code

## 📋 Common Patterns to Flag

### Pattern 1: DB Transaction

```regex
\.Transaction\(ctx,\s*func\((\w+)\s+context\.Context\)
```

Then search inside that block for uses of `ctx` (should be using captured context name instead)

### Pattern 2: Loop Variable Capture

```go
for _, item := range items {
    go func() {
        process(item)  // ❌ Captures loop variable
    }()
}
```

### Pattern 3: Unbuffered Channel Timeout

```go
ch := make(chan T)  // Unbuffered
go func() {
    ch <- value  // May leak if timeout
}()

select {
case v := <-ch:
case <-time.After(...):  // ⚠️ Goroutine still blocked
}
```

## 🚫 Anti-Patterns

| Anti-Pattern | Why Bad | Better |
|--------------|---------|--------|
| `panic()` in libraries | Caller can't recover | Return error |
| `_` for errors | Silent failures | Check or justify |
| `time.Sleep()` in tests | Flaky tests | Use channels/sync |
| Magic numbers | Unclear meaning | Named constants |
| Log and return error | Duplicate logs | Choose one |
| Deep nesting | Hard to read | Early returns |

## 📚 Reference Quick Links

- **Transaction Context**: [Full guide](./references/transaction-context.md)
- **Error Handling**: [Full guide](./references/error-handling.md)
- **Concurrency**: [Full guide](./references/concurrency-patterns.md)
- **Performance**: [Full guide](./references/performance-optimization.md)
- **Idiomatic Go**: [Full guide](./references/idiomatic-go.md)
- **Testing**: [Full guide](./references/testing-best-practices.md)
- **Uber Go Style**: [Full guide](./references/uber-go-style.md)

## ⏱️ 60-Second Review

Can't do a full review? Check these in 60 seconds:

1. ✓ Any `Transaction(` blocks → verify context usage (30s)
2. ✓ Any `go func(` → check for leaks (15s)
3. ✓ Any `defer` → verify cleanup happens (10s)
4. ✓ Any new tests → spot check coverage (5s)

## 🎓 Learning Path

1. Start with **Critical Issues** - Learn transaction context pattern
2. Master **Error Handling** - Understand wrapping and sentinel errors
3. Study **Concurrency** - Goroutines, channels, sync primitives
4. Optimize **Performance** - When and how to optimize
5. Perfect **Testing** - Table-driven tests and coverage

---

**Pro Tip**: Focus on critical issues first. A bug in transaction context can corrupt data. Missing a constant name suggestion won't.

