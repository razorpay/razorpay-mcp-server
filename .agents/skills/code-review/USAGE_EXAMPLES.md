# Go Code Review - Usage Examples

This document provides concrete examples of how to use the Go Code Review skill.

## Basic Usage

### Example 1: Simple Review Request

**Prompt**:
```
Review this PR @Branch (Diff with Main Branch) based on Uber Go style principles and Go best practices.
```

**What Happens**:
1. Skill reads the branch diff
2. Analyzes all `.go` files
3. Checks against Uber Go style guide
4. Reports issues by severity
5. Provides fix suggestions

---

## Focused Reviews

### Example 2: Transaction Context Focus

**Prompt**:
```
Review this PR @Branch for database transaction context issues. 
Check if all database operations inside transactions use the correct context (tctx instead of ctx).
```

**What Skill Checks**:
```go
// Searches for patterns like:
r.db.Transaction(ctx, func(tctx context.Context) error {
    // Verifies ALL calls inside use 'tctx' not 'ctx'
    r.GetPayment(tctx, id)     // ✅ Correct
    r.UpdateStatus(tctx, id)   // ✅ Correct
})
```

**Sample Output**:
```markdown
## 🚨 Critical Issue Found

### Incorrect Transaction Context Usage
**File**: `internal/repo/payment_repo.go`
**Lines**: 145-158

**Problem**: Database operations using outer context 'ctx' instead of transaction context 'tctx'

[Code example with fix]
```

---

### Example 3: Concurrency Review

**Prompt**:
```
Review @Branch focusing on:
- Goroutine leaks
- Race conditions
- Channel usage
- Proper synchronization
```

**What Skill Checks**:
- Goroutines with no exit condition
- Shared variables without mutex
- Unbuffered channels that might block
- Missing WaitGroup coordination
- Loop variable captures in closures

**Sample Finding**:
```markdown
## 🚨 Critical: Goroutine Leak Detected

**File**: `internal/worker/processor.go`
**Line**: 89

```go
// ❌ PROBLEM
ch := make(chan Result)
go func() {
    ch <- expensiveOperation()  // Blocks if timeout!
}()

select {
case result := <-ch:
    return result
case <-time.After(time.Second):
    return Result{}  // Goroutine still running!
}
```

**Fix**: Use buffered channel or context cancellation
```

---

### Example 4: Error Handling Review

**Prompt**:
```
Review this PR focusing on error handling:
- Are errors wrapped properly?
- Sentinel errors used correctly?
- Error context provided?
```

**Sample Finding**:
```markdown
## ⚠️ Important: Error Not Wrapped

**File**: `internal/service/payment.go`
**Line**: 67

```go
// ❌ Current
if err := r.ProcessPayment(payment); err != nil {
    return err  // Lost context!
}

// ✅ Suggested
if err := r.ProcessPayment(payment); err != nil {
    return fmt.Errorf("failed to process payment %s: %w", payment.ID, err)
}
```
```

---

### Example 5: Performance Review

**Prompt**:
```
Review @Branch for performance issues:
- Unnecessary allocations
- String concatenation in loops
- Missing pre-allocation
- Inefficient data structures
```

**Sample Finding**:
```markdown
## ⚡ Performance: Unnecessary Allocations

**File**: `internal/handler/api.go`
**Lines**: 45-52

```go
// ❌ Current - grows multiple times
var results []Result
for _, item := range items {
    results = append(results, process(item))
}

// ✅ Suggested - single allocation
results := make([]Result, 0, len(items))
for _, item := range items {
    results = append(results, process(item))
}
```

**Impact**: Reduces allocations from O(log n) to O(1)
```

---

## Complete Review Examples

### Example 6: Full Payment Service PR Review

**Prompt**:
```
Review this PR @feature/payment-refund (Diff with Main) comprehensively:
- Uber Go style compliance
- Database transaction correctness
- Error handling
- Concurrency safety
- Performance
- Test coverage
```

**Expected Output Structure**:

```markdown
# Go Code Review: feature/payment-refund

## Summary
- **Files Changed**: 6 Go files, 2 test files
- **Lines Added**: +347, -128
- **Overall Assessment**: ⚠️ Needs Changes

### Quick Stats
- Critical Issues: 1
- Important Issues: 3
- Performance Concerns: 2
- Minor Suggestions: 5
- Test Coverage: Adequate ✓

---

## 🚨 Critical Issues (Must Fix)

### 1. Transaction Context Violation
**File**: `internal/repo/refund_repo.go:78-95`
**Severity**: CRITICAL

Inside `CreateRefund` transaction, database operations use wrong context.

[Detailed code example with fix]

**Impact**: Could cause transaction isolation issues and data inconsistency.

---

## ⚠️ Important Issues (Should Fix)

### 1. Error Not Wrapped
**File**: `internal/service/refund_service.go:45`
[Code example]

### 2. Missing Context Propagation
**File**: `internal/service/refund_service.go:67`
[Code example]

### 3. No Validation for Edge Case
**File**: `internal/handler/refund_handler.go:123`
[Code example]

---

## ⚡ Performance Concerns

### 1. String Concatenation in Loop
**File**: `internal/handler/refund_handler.go:89-95`
[Code example with strings.Builder suggestion]

### 2. Unnecessary Goroutine Creation
**File**: `internal/worker/refund_processor.go:56`
[Code example with worker pool suggestion]

---

## 💡 Minor Suggestions

1. Extract magic number 30 to constant (refund_service.go:45)
2. Use early return to reduce nesting (refund_handler.go:78)
3. Rename variable `r` to `refund` for clarity (refund_repo.go:123)
4. Add package documentation comment (refund.go:1)
5. Consistent receiver name (refund_service.go:34,45 uses different names)

---

## ✅ Positive Feedback

- Good test coverage for happy path
- Proper use of defer for cleanup
- Clear error messages
- Consistent naming conventions

---

## 📋 Test Coverage Analysis

### Covered
- ✓ Successful refund creation
- ✓ Invalid amount validation
- ✓ Payment not found error

### Missing
- ❌ Concurrent refund attempts on same payment
- ❌ Database transaction rollback scenario
- ❌ Network timeout during external API call

---

## Recommendations

1. **Immediate**: Fix transaction context issue (critical)
2. **Before Merge**: Address error wrapping and context propagation
3. **Consider**: Add missing test cases for edge scenarios
4. **Optional**: Apply minor style suggestions for consistency
```

---

## Specialized Reviews

### Example 7: Test Coverage Review

**Prompt**:
```
Review test coverage for @Branch:
- Are all error paths tested?
- Table-driven tests used?
- Edge cases covered?
- Mocks appropriate?
```

**Sample Output**:
```markdown
## 🧪 Test Coverage Analysis

### File: payment_service_test.go

#### ✅ Well-Tested
- Happy path with valid payment
- Invalid amount validation
- Payment not found error

#### ❌ Missing Coverage
1. **Concurrent payment processing** - No test for race conditions
2. **Database transaction rollback** - Only success path tested
3. **Context cancellation** - No test for timeout behavior
4. **Boundary values** - No test for amount = 0 or max int

#### 💡 Suggestions
- Use table-driven tests for validation scenarios
- Add benchmark for ProcessPayment (performance critical)
- Mock external dependencies (currently using real HTTP)
```

---

### Example 8: Security Review

**Prompt**:
```
Review @Branch for security concerns:
- SQL injection risks
- Input validation
- Secret handling
- Authentication/authorization
```

**Sample Finding**:
```markdown
## 🔒 Security Concerns

### 1. Missing Input Validation
**File**: `internal/handler/user_handler.go:45`

```go
// ❌ Unsafe - no validation
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    user, _ := h.service.GetUser(r.Context(), userID)  // No validation!
    json.NewEncoder(w).Encode(user)
}

// ✅ Safe
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    if userID == "" {
        http.Error(w, "user ID required", http.StatusBadRequest)
        return
    }
    
    if !isValidUserID(userID) {
        http.Error(w, "invalid user ID format", http.StatusBadRequest)
        return
    }
    
    user, err := h.service.GetUser(r.Context(), userID)
    if err != nil {
        http.Error(w, "user not found", http.StatusNotFound)
        return
    }
    
    json.NewEncoder(w).Encode(user)
}
```
```

---

## Command Line Usage (if available)

```bash
# Basic review
claude-review --skill go-code-review --branch feature/payment --base main

# Focus on specific areas
claude-review --skill go-code-review \
  --branch feature/payment \
  --focus transaction_context,concurrency

# Strict mode (all issues)
claude-review --skill go-code-review \
  --branch feature/payment \
  --strictness strict

# Generate report file
claude-review --skill go-code-review \
  --branch feature/payment \
  --output review-report.md
```

---

## Integration Examples

### GitHub Actions

```yaml
name: Go Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
        with:
          fetch-depth: 0
      
      - name: Run Go Code Review
        run: |
          claude-review \
            --skill go-code-review \
            --branch ${{ github.head_ref }} \
            --base ${{ github.base_ref }} \
            --output review.md
      
      - name: Comment on PR
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const review = fs.readFileSync('review.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: review
            });
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-push

echo "🔍 Running Go code review..."

CURRENT_BRANCH=$(git branch --show-current)

if [ "$CURRENT_BRANCH" != "main" ]; then
    claude-review \
        --skill go-code-review \
        --branch "$CURRENT_BRANCH" \
        --base main \
        --strictness standard
    
    if [ $? -ne 0 ]; then
        echo "❌ Code review found issues. Please fix before pushing."
        exit 1
    fi
fi

echo "✅ Code review passed!"
```

---

## Tips for Effective Reviews

### 1. Start with Critical Issues
Always address critical issues first:
- Transaction context bugs
- Race conditions
- Resource leaks

### 2. Use Focus Mode for Large PRs
For PRs with 1000+ lines:
```
Review @Branch focusing only on:
1. Database transaction context
2. Concurrency safety
3. Critical error paths
```

### 3. Iterate on Fixes
After fixing issues:
```
Re-review @Branch - verify fixes for transaction context issues reported earlier
```

### 4. Learn from Patterns
Use the skill as a learning tool:
```
Review this code and explain the Uber Go style principles being violated
```

---

## Common Patterns to Request

### Pattern 1: New Service Review
```
Review new payment service in @Branch:
- Architecture and design
- Error handling
- Testing coverage
- Performance considerations
```

### Pattern 2: Refactoring Review
```
Review refactoring in @Branch:
- Backward compatibility
- No new bugs introduced
- Performance impact
- Test coverage maintained
```

### Pattern 3: Bug Fix Review
```
Review bug fix in @Branch:
- Root cause addressed
- No new issues introduced
- Test case added
- Similar issues elsewhere?
```

---

**Remember**: The skill is most effective when you:
1. Be specific about what to check
2. Mention if it's a critical path (payment, auth, etc.)
3. Ask for explanations, not just findings
4. Iterate on fixes and re-review

