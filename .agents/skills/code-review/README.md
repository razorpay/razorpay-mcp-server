# Go Code Review - Claude Skill

A comprehensive code review skill for Go (Golang) projects with deep focus on Uber Go style guide, idiomatic patterns, and critical bug detection (especially database transaction context issues).

## Quick Start

### Basic Usage

```
Review this PR @Branch (Diff with Main Branch) based on Uber Go style principles and best practices.
```

### Comprehensive Review

```
Review this PR @Branch (Diff with Main Branch) focusing on:
1. Uber Go style guide compliance
2. Database transaction context correctness
3. Error handling patterns
4. Concurrency issues
5. Performance bottlenecks
```

### Transaction Context Review

```
Check this PR @Branch for database transaction context issues - verify all DB operations inside transactions use tctx, not ctx.
```

## What This Skill Does

This skill performs expert-level Go code reviews with emphasis on:

### 1. Critical Bug Detection

- **Database Transaction Context Issues** (CRITICAL) - Detects incorrect context usage in DB transactions
- Race conditions in concurrent code
- Goroutine leaks
- Resource leaks (unclosed files, connections)
- nil pointer dereferences

### 2. Uber Go Style Guide Compliance

- Error handling patterns (wrapping, sentinel errors)
- Pointer vs value receivers
- Interface design
- Proper use of goroutines and channels
- Context propagation

### 3. Go Best Practices

- Idiomatic Go patterns
- Effective concurrency patterns
- Proper defer usage
- String building efficiency
- Slice and map pre-allocation

### 4. Performance Analysis

- Unnecessary allocations
- String concatenation inefficiencies
- Goroutine pool usage
- Database query optimization
- Lock contention issues

### 5. Testing Quality

- Table-driven test patterns
- Test coverage adequacy
- Mock usage appropriateness
- Benchmark presence for critical paths

## The Critical Transaction Context Issue

One of the **most important** checks this skill performs is detecting incorrect context usage in database transactions:

### The Problem

```go
// ❌ CRITICAL BUG
func (r *Repo) UpdatePayment(ctx context.Context, id string) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        // BUG: Using 'ctx' instead of 'tctx'
        payment, err := r.GetPayment(ctx, id)  // Wrong context!
        if err != nil {
            return err
        }
        return r.UpdateStatus(ctx, payment.ID, "completed")  // Wrong context!
    })
}
```

### The Fix

```go
// ✅ CORRECT
func (r *Repo) UpdatePayment(ctx context.Context, id string) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        // Correctly using 'tctx'
        payment, err := r.GetPayment(tctx, id)
        if err != nil {
            return err
        }
        return r.UpdateStatus(tctx, payment.ID, "completed")
    })
}
```

This bug can cause:
- Transaction isolation violations
- Data corruption
- Incorrect timeout handling
- Difficult-to-debug race conditions

## Reference Documents

The skill uses these comprehensive references:

| Document | Purpose |
|----------|---------|
| [uber-go-style.md](./references/uber-go-style.md) | Complete Uber Go style guide |
| [transaction-context.md](./references/transaction-context.md) | Critical DB transaction patterns |
| [error-handling.md](./references/error-handling.md) | Error patterns and anti-patterns |
| [concurrency-patterns.md](./references/concurrency-patterns.md) | Goroutines, channels, sync primitives |
| [idiomatic-go.md](./references/idiomatic-go.md) | Go idioms and conventions |
| [performance-optimization.md](./references/performance-optimization.md) | Performance best practices |
| [testing-best-practices.md](./references/testing-best-practices.md) | Testing patterns and coverage |

## Output Format

### Review Summary

The skill provides a structured review with:

```markdown
## Go Code Review Summary

**Branch**: feature/payment-gateway
**Files Changed**: 8 Go files
**Overall Assessment**: ⚠️ NEEDS CHANGES

### Key Metrics
- Critical Issues: 2
- Important Issues: 5  
- Minor Suggestions: 8
- Performance Concerns: 3
```

### Critical Issues

Issues that **must** be fixed before merging:

```markdown
## 🚨 Critical Issues

### 1. Incorrect Transaction Context Usage
**File**: `internal/repo/payment_repo.go`
**Lines**: 145-158
**Severity**: CRITICAL

[Detailed explanation with before/after code examples]
```

### Important Issues

Should be fixed:

```markdown
## ⚠️ Important Issues

### 1. Missing Error Wrapping
**File**: `internal/service/payment.go`
**Line**: 67

[Code example with suggestion]
```

### Performance Concerns

```markdown
## ⚡ Performance Concerns

### 1. Unnecessary Goroutine Creation
**File**: `internal/handler/webhook.go`
**Lines**: 89-95

[Analysis with optimization suggestion]
```

### Minor Suggestions

Nice-to-have improvements:

```markdown
## 💡 Minor Suggestions

- Extract magic number to constant (line 42)
- Use early return to reduce nesting (line 67)
```

## Configuration

### Review Strictness

- **relaxed**: Critical issues only
- **standard**: Critical + important issues (default)
- **strict**: All issues including minor style

### Focus Areas

```yaml
focus_areas:
  - transaction_context  # DB transaction context correctness
  - error_handling       # Error patterns
  - concurrency          # Race conditions, goroutine leaks
  - performance          # Performance bottlenecks
  - testing              # Test coverage and quality
  - security             # Security vulnerabilities
```

## Examples

### Example 1: Transaction Context Review

**Input**: PR with database operations

**Output**:
- Identifies all transaction blocks
- Checks context usage in each operation
- Reports violations with line numbers
- Provides corrected code

### Example 2: Concurrency Review

**Input**: PR introducing goroutines

**Output**:
- Identifies potential race conditions
- Checks for goroutine leaks
- Verifies proper synchronization
- Suggests worker pool patterns if needed

### Example 3: Performance Review

**Input**: PR with hot path changes

**Output**:
- Identifies allocation hotspots
- Suggests pre-allocation for slices/maps
- Recommends `strings.Builder` for concatenation
- Checks for lock contention

## Integration

### CI/CD Integration

```yaml
# .github/workflows/code-review.yml
name: Go Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run Go Code Review
        run: |
          claude-review \
            --skill go-code-review \
            --branch ${{ github.head_ref }} \
            --base main
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-push

echo "Running Go code review..."
claude-review --skill go-code-review --branch $(git branch --show-current) --base main
```

## Common Review Scenarios

### Scenario 1: Payment Processing PR

**Common Issues Found**:
- Transaction context incorrect ✓
- Error not wrapped ✓
- Missing validation ✓
- No test for error case ✓

### Scenario 2: Concurrent Worker PR

**Common Issues Found**:
- Loop variable captured incorrectly ✓
- Goroutine leak on timeout ✓
- No WaitGroup for coordination ✓
- Race condition in map access ✓

### Scenario 3: API Handler PR

**Common Issues Found**:
- Context not propagated ✓
- Panic not recovered ✓
- Error logged and returned (duplicate) ✓
- No input validation ✓

## Troubleshooting

### Skill not detecting transaction issues?

1. Ensure transaction function matches pattern: `func(tctx context.Context)`
2. Check if using different transaction library
3. Add pattern to reference docs

### Too many minor suggestions?

1. Adjust strictness to "standard" or "relaxed"
2. Configure focus areas to skip minor checks
3. Use `--focus transaction_context,concurrency` flag

### Missing performance issues?

1. Enable performance analysis explicitly
2. Provide benchmark baseline if available
3. Check if performance concerns are real bottlenecks

## Extending the Skill

### Adding New Patterns

To detect new anti-patterns:

1. Add pattern to appropriate reference doc
2. Include bad/good examples
3. Document severity (critical/important/minor)
4. Add test case if possible

### Repository-Specific Rules

Create `.go-review.yaml` in repo:

```yaml
rules:
  transaction_context:
    enabled: true
    severity: critical
  custom_patterns:
    - pattern: "log.Printf.*password"
      message: "Don't log passwords"
      severity: critical
```

## Best Practices

1. **Run Early**: Review code before it reaches PR stage
2. **Focus on Critical**: Address critical issues first
3. **Learn Patterns**: Use reference docs to understand why
4. **Iterate**: Re-run review after fixes
5. **Automate**: Integrate into CI/CD pipeline

## Metrics

The skill tracks:

- Total files reviewed
- Issues found by severity
- Issue categories (transaction, concurrency, etc.)
- Review time
- Test coverage percentage

## Version History

- **1.0.0** (2026-01-18): Initial release
  - Uber Go style guide integration
  - Transaction context detection
  - Error handling analysis
  - Concurrency safety checks
  - Performance optimization suggestions
  - 7 comprehensive reference documents

## Support

For issues or questions:

1. Check the reference documents in `./references/`
2. Review example scenarios above
3. Consult Uber Go Style Guide
4. File an issue with code example

## License

Part of the Agent Skills repository.

---

**Remember**: The goal is to catch critical bugs (especially transaction context issues) and ensure idiomatic, maintainable Go code. Focus on what matters most for code quality and correctness.

