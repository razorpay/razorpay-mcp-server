---
name: go-code-review
description: Performs comprehensive Go code reviews focusing on Uber Go style principles, idiomatic patterns, critical bugs (especially database transaction context issues), error handling, concurrency safety, and performance. Use when reviewing Go pull requests, checking for race conditions, goroutine leaks, or validating transaction context correctness.
license: Part of Agent Skills repository
metadata:
  version: "1.0.0"
  category: "Code Quality & Review"
  language: "Go"
  author: "razorpay"
---

# Go Code Review Skill

## What This Skill Does

Reviews Go (Golang) pull requests against best practices with focus on:

1. **Critical Bug Detection** (HIGHEST PRIORITY)
   - Database transaction context issues (`ctx` vs `tctx`)
   - Race conditions and goroutine leaks
   - Resource leaks (unclosed files, connections)

2. **Uber Go Style Guide Compliance**
   - Error handling (wrapping, sentinel errors)
   - Pointer vs value receivers
   - Interface design and usage

3. **Concurrency Safety**
   - Goroutine lifecycle management
   - Channel usage patterns
   - Synchronization primitives

4. **Performance & Quality**
   - Unnecessary allocations
   - Test coverage and patterns

## When to Use This Skill

Use this skill when:
- Reviewing Go PRs with database operations
- Checking concurrent code (goroutines, channels)
- Validating error handling patterns
- Ensuring Uber Go style compliance
- Analyzing performance-critical Go code

## Usage Examples

### Basic Review
```
Review this PR @Branch (Diff with Main Branch) for Go best practices 
and Uber Go style compliance.
```

### Focused on Critical Issues
```
Review @Branch focusing on:
1. Database transaction context correctness
2. Race conditions and goroutine leaks
3. Error handling patterns
```

### Comprehensive Review
```
Review this PR @Branch based on Uber Go style principles, idiomatic Go code,
performance bottlenecks, error handling, and concurrency safety. Check that all
database operations inside transactions use the correct context (tctx not ctx).
```

## Review Process

1. **Fetch Changes**: Get branch diff and identify `.go` files
2. **Critical Checks**: Scan for transaction context issues, race conditions
3. **Style Analysis**: Verify Uber Go style guide compliance
4. **Performance Review**: Check for allocations, inefficiencies
5. **Test Analysis**: Review coverage and patterns
6. **Generate Report**: Categorize by severity with fix suggestions

## Output Structure

Reports are organized by severity:

- **🚨 Critical** (Must Fix): Transaction bugs, race conditions, data corruption risks
- **⚠️ Important** (Should Fix): Error handling, context propagation
- **💡 Minor** (Nice to Have): Style consistency, naming
- **⚡ Performance** (Optimize): Allocations, string ops

Each issue includes:
- File path and line numbers
- Problem description with code example
- Suggested fix with corrected code
- Impact explanation
- Reference to detailed documentation

## Critical: Transaction Context Pattern

The #1 critical bug this skill detects:

```go
// ❌ CRITICAL BUG
func (r *Repo) Update(ctx context.Context, id string) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        // Using 'ctx' instead of 'tctx' - WRONG!
        data, err := r.Get(ctx, id)
        return r.Save(ctx, data)
    })
}

// ✅ CORRECT
func (r *Repo) Update(ctx context.Context, id string) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        // Using 'tctx' - CORRECT!
        data, err := r.Get(tctx, id)
        return r.Save(tctx, data)
    })
}
```

**Why Critical**: Wrong context causes transaction isolation violations, timeout issues, and potential data corruption.

## Reference Documentation

The skill uses these comprehensive guides (loaded on-demand):

1. **[Transaction Context](references/transaction-context.md)** - Critical DB patterns (386 lines)
2. **[Uber Go Style](references/uber-go-style.md)** - Complete style guide (742 lines)
3. **[Error Handling](references/error-handling.md)** - Patterns and anti-patterns (654 lines)
4. **[Concurrency](references/concurrency-patterns.md)** - Goroutines, channels (825 lines)
5. **[Idiomatic Go](references/idiomatic-go.md)** - Go conventions (860 lines)
6. **[Performance](references/performance-optimization.md)** - Optimization (824 lines)
7. **[Testing](references/testing-best-practices.md)** - Test patterns (721 lines)

See [INDEX.md](INDEX.md) for complete navigation guide.

## Configuration Options

### Strictness Levels

- `relaxed`: Critical issues only
- `standard`: Critical + important (default)
- `strict`: All issues including style

### Focus Areas

Specify in prompt:
```
Review @Branch focusing on: transaction_context, concurrency, performance
```

Available focus areas:
- `transaction_context` - DB transaction correctness
- `error_handling` - Error patterns
- `concurrency` - Race conditions, goroutine safety
- `performance` - Bottlenecks
- `testing` - Test quality
- `security` - Security vulnerabilities

## Quick Reference

For fast reviews, see [QUICK_REFERENCE.md](QUICK_REFERENCE.md) which provides:
- Critical issues checklist
- 60-second scan guide
- Common patterns to flag
- Anti-patterns reference

## Integration Examples

### GitHub Actions
```yaml
- name: Go Code Review
  run: |
    claude-review --skill go-code-review \
      --branch ${{ github.head_ref }} \
      --base main
```

### Pre-commit Hook
```bash
#!/bin/bash
claude-review --skill go-code-review \
  --branch $(git branch --show-current) \
  --base main
```

## Supported Libraries

Transaction context detection works with:
- GORM
- sqlx
- database/sql
- Custom transaction wrappers matching `func(tctx context.Context) error`

## Limitations

- Does not execute code or run tests
- Static analysis only (no runtime behavior)
- Repository-specific rules need custom configuration
- May miss complex control flow patterns

## Troubleshooting

**Skill not detecting transaction issues?**
- Verify function signature: `func(tctx context.Context) error`
- Check if using non-standard transaction pattern

**Too many minor suggestions?**
- Use `strictness: relaxed` in prompt
- Focus on specific areas only

**Need more detail?**
- Reference specific guide in `references/` directory
- See [USAGE_EXAMPLES.md](USAGE_EXAMPLES.md) for detailed scenarios

## Additional Resources

- [README.md](README.md) - Complete user guide
- [USAGE_EXAMPLES.md](USAGE_EXAMPLES.md) - Concrete usage scenarios
- [STRUCTURE.md](STRUCTURE.md) - Directory organization
- [INDEX.md](INDEX.md) - Documentation navigation

## Version

**Current**: 1.0.0 (2026-01-18)

**Changelog**:
- 1.0.0: Initial release with Uber Go style, transaction context detection, concurrency checks, performance analysis
