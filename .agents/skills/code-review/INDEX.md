# Go Code Review Skill - Complete Index

Welcome to the Go Code Review Skill! This index helps you navigate all documentation.

## 📚 Documentation Structure

### 🎯 Start Here

1. **[SKILL.md](./SKILL.md)** - Main skill definition
   - What this skill does
   - Usage instructions
   - Review process
   - Output format

2. **[README.md](./README.md)** - User guide
   - Quick start
   - Critical features (transaction context!)
   - Integration examples
   - Troubleshooting

3. **[USAGE_EXAMPLES.md](./USAGE_EXAMPLES.md)** - Concrete examples
   - Basic usage
   - Focused reviews
   - Complete review examples
   - Integration patterns

### ⚡ Quick Access

4. **[QUICK_REFERENCE.md](./QUICK_REFERENCE.md)** - Fast checklist
   - Critical issues checklist
   - Common patterns to flag
   - 60-second review guide
   - Anti-patterns

### 📖 Reference Documentation

#### Core References

5. **[transaction-context.md](./references/transaction-context.md)** ⭐ **MOST CRITICAL**
   - Database transaction context patterns
   - Detection of incorrect context usage
   - Library-specific patterns (GORM, sqlx, database/sql)
   - Real-world examples and fixes
   - **Read this first if you work with databases**

6. **[uber-go-style.md](./references/uber-go-style.md)**
   - Complete Uber Go Style Guide
   - Error handling patterns
   - Pointer vs value receivers
   - Interfaces and naming
   - Testing patterns

7. **[error-handling.md](./references/error-handling.md)**
   - Error wrapping with `%w`
   - Sentinel errors
   - Custom error types
   - Error aggregation
   - Testing error cases

#### Advanced References

8. **[concurrency-patterns.md](./references/concurrency-patterns.md)**
   - Goroutine lifecycle management
   - Channel patterns (fan-out, fan-in, pipeline)
   - Worker pools
   - Synchronization primitives
   - Race condition detection
   - Deadlock prevention

9. **[idiomatic-go.md](./references/idiomatic-go.md)**
   - Package organization
   - Variable declarations
   - Function patterns
   - Struct design
   - Method receivers
   - Code organization

10. **[performance-optimization.md](./references/performance-optimization.md)**
    - Memory allocation optimization
    - String operations
    - Loop optimization
    - Concurrency performance
    - I/O operations
    - Benchmarking and profiling

11. **[testing-best-practices.md](./references/testing-best-practices.md)**
    - Table-driven tests
    - Test helpers
    - Mocking strategies
    - Coverage measurement
    - Integration tests
    - Benchmarks

### 📊 Meta Documentation

12. **[STRUCTURE.md](./STRUCTURE.md)** - Visual directory structure
    - File organization
    - Size statistics
    - Content summary

13. **[IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md)** - Technical overview
    - Deliverables
    - Features
    - Success metrics
    - Future enhancements

## 🎯 Use Cases & Where to Start

### "I need to review a PR quickly"
→ Start with **[QUICK_REFERENCE.md](./QUICK_REFERENCE.md)** (60-second scan)

### "I'm reviewing payment/database code"
→ **CRITICAL**: Read **[transaction-context.md](./references/transaction-context.md)** first!

### "I'm new to Go code reviews"
→ Read in order:
1. [README.md](./README.md) - Overview
2. [uber-go-style.md](./references/uber-go-style.md) - Style guide
3. [USAGE_EXAMPLES.md](./USAGE_EXAMPLES.md) - Concrete examples

### "I'm reviewing concurrent code"
→ Focus on **[concurrency-patterns.md](./references/concurrency-patterns.md)**

### "I need to optimize performance"
→ Check **[performance-optimization.md](./references/performance-optimization.md)**

### "I'm writing tests"
→ Reference **[testing-best-practices.md](./references/testing-best-practices.md)**

### "I want comprehensive coverage"
→ Use **[SKILL.md](./SKILL.md)** for full review capabilities

## 🔍 Find Information By Topic

### Critical Issues

| Topic | Document | Section |
|-------|----------|---------|
| Transaction Context | [transaction-context.md](./references/transaction-context.md) | Complete |
| Race Conditions | [concurrency-patterns.md](./references/concurrency-patterns.md) | Race Conditions |
| Goroutine Leaks | [concurrency-patterns.md](./references/concurrency-patterns.md) | Goroutine Lifecycle |
| Resource Leaks | [uber-go-style.md](./references/uber-go-style.md) | Defer |

### Error Handling

| Topic | Document | Section |
|-------|----------|---------|
| Error Wrapping | [error-handling.md](./references/error-handling.md) | Error Wrapping |
| Sentinel Errors | [error-handling.md](./references/error-handling.md) | Sentinel Errors |
| Custom Error Types | [error-handling.md](./references/error-handling.md) | Custom Error Types |

### Concurrency

| Topic | Document | Section |
|-------|----------|---------|
| Goroutines | [concurrency-patterns.md](./references/concurrency-patterns.md) | Goroutines |
| Channels | [concurrency-patterns.md](./references/concurrency-patterns.md) | Channels |
| Mutexes | [concurrency-patterns.md](./references/concurrency-patterns.md) | Synchronization |
| Worker Pools | [concurrency-patterns.md](./references/concurrency-patterns.md) | Worker Pool Pattern |

### Performance

| Topic | Document | Section |
|-------|----------|---------|
| Allocations | [performance-optimization.md](./references/performance-optimization.md) | Memory Allocation |
| String Building | [performance-optimization.md](./references/performance-optimization.md) | String Operations |
| Benchmarking | [performance-optimization.md](./references/performance-optimization.md) | Benchmarking |
| Profiling | [performance-optimization.md](./references/performance-optimization.md) | Profiling |

### Testing

| Topic | Document | Section |
|-------|----------|---------|
| Table-Driven Tests | [testing-best-practices.md](./references/testing-best-practices.md) | Table-Driven Tests |
| Mocking | [testing-best-practices.md](./references/testing-best-practices.md) | Mocking |
| Coverage | [testing-best-practices.md](./references/testing-best-practices.md) | Test Coverage |

### Style & Idioms

| Topic | Document | Section |
|-------|----------|---------|
| Naming | [idiomatic-go.md](./references/idiomatic-go.md) | Naming |
| Package Organization | [idiomatic-go.md](./references/idiomatic-go.md) | Package Organization |
| Interfaces | [uber-go-style.md](./references/uber-go-style.md) | Interfaces |
| Receivers | [uber-go-style.md](./references/uber-go-style.md) | Pointer vs Value |

## 📏 Document Sizes & Complexity

| Document | Lines | Complexity | Read Time |
|----------|-------|------------|-----------|
| QUICK_REFERENCE.md | 307 | Simple | 5 min |
| transaction-context.md | 386 | Medium | 10 min |
| README.md | 401 | Simple | 8 min |
| SKILL.md | 332 | Simple | 7 min |
| error-handling.md | 654 | Medium | 15 min |
| testing-best-practices.md | 721 | Medium | 18 min |
| uber-go-style.md | 742 | Medium | 20 min |
| performance-optimization.md | 824 | Advanced | 25 min |
| concurrency-patterns.md | 825 | Advanced | 25 min |
| idiomatic-go.md | 860 | Medium | 22 min |

**Total**: 7,092 lines, ~2.5 hours to read everything

## 🚀 Quick Start Paths

### Path 1: Minimum Viable Knowledge (30 minutes)
1. [README.md](./README.md) - 8 min
2. [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - 5 min
3. [transaction-context.md](./references/transaction-context.md) - 10 min
4. [USAGE_EXAMPLES.md](./USAGE_EXAMPLES.md) - 7 min

### Path 2: Comprehensive Understanding (2.5 hours)
Read all documents in order listed above

### Path 3: As-Needed Reference
Keep [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) handy, deep-dive into specific references when needed

## 🎓 Learning Progression

### Beginner (Week 1)
- [x] Read README.md
- [x] Read QUICK_REFERENCE.md
- [x] Read transaction-context.md
- [x] Try basic PR review

### Intermediate (Week 2-3)
- [x] Read uber-go-style.md
- [x] Read error-handling.md
- [x] Read idiomatic-go.md
- [x] Review 5+ PRs with skill

### Advanced (Week 4+)
- [x] Read concurrency-patterns.md
- [x] Read performance-optimization.md
- [x] Read testing-best-practices.md
- [x] Contribute improvements to skill

## 🔗 External References

The skill is based on these authoritative sources:

1. **[Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)** (Official)
2. **[Effective Go](https://go.dev/doc/effective_go)** (Official)
3. **[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)** (Go team)
4. **[Go Concurrency Patterns](https://go.dev/blog/pipelines)** (Go blog)
5. **[Go Performance](https://github.com/golang/go/wiki/Performance)** (Go wiki)

## 📝 Document Status

| Document | Status | Last Updated |
|----------|--------|--------------|
| All files | ✅ Complete | Jan 18, 2026 |

## 💡 Pro Tips

1. **Bookmark QUICK_REFERENCE.md** for fast access during reviews
2. **Always check transaction-context.md** when reviewing DB code
3. **Use USAGE_EXAMPLES.md** to craft better review prompts
4. **Reference specific sections** when explaining issues to developers
5. **Keep learning** - Go best practices evolve

## 🤝 Contributing

To improve this skill:

1. Add new patterns to reference documents
2. Share real-world examples
3. Report false positives/negatives
4. Suggest new checks

## 📞 Support

- **Quick Questions**: Check [QUICK_REFERENCE.md](./QUICK_REFERENCE.md)
- **Usage Help**: See [USAGE_EXAMPLES.md](./USAGE_EXAMPLES.md)
- **Deep Dive**: Read specific reference document
- **Technical Details**: Check [IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md)

---

**Version**: 1.0.0  
**Created**: January 18, 2026  
**Total Documentation**: 7,092 lines across 13 files  
**Status**: ✅ Production Ready

**Start Your Review Journey**: Begin with [README.md](./README.md) →

