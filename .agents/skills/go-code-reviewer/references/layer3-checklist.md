# Layer 3: Quality Intent Checklist

This checklist provides detailed guidance for reviewing Go code quality. Use this as a reference when performing Layer 3 reviews.

## Preservation Principles

### Existing Behavior
- [ ] Changes preserve existing behavior unless explicitly intended to modify it
- [ ] No breaking changes to public APIs without justification
- [ ] Backward compatibility maintained where expected

### Code Style Consistency
- [ ] Matches existing code style in the same package/module
- [ ] Follows established naming conventions in the codebase
- [ ] Uses same error handling patterns as surrounding code
- [ ] Respects existing abstraction boundaries

### Team Conventions
- [ ] Aligns with team's documented coding standards
- [ ] Uses established utility functions instead of reinventing
- [ ] Follows project-specific patterns (logging, config, etc.)

## Go Idioms & Best Practices

### Naming (following Go conventions)
- [ ] Package names: lowercase, single word, no underscores
- [ ] Exported names: MixedCaps (not snake_case)
- [ ] Unexported names: mixedCaps
- [ ] Interface names: often single method + "er" suffix (Reader, Writer)
- [ ] Variable names: short in small scopes, descriptive in larger scopes
- [ ] No unnecessary abbreviations that hurt readability

### Error Handling
- [ ] Errors are checked and handled appropriately
- [ ] Error messages include context (what operation failed, why)
- [ ] Errors are wrapped with fmt.Errorf("...:%w", err) for context
- [ ] Custom error types used when callers need to inspect errors
- [ ] Panic only used for unrecoverable programming errors

### Concurrency
- [ ] Goroutines used only when concurrency is actually needed
- [ ] All goroutines have clear termination conditions
- [ ] Channels used appropriately (buffered vs unbuffered)
- [ ] Mutexes used correctly (no missing unlocks, defer unlock after Lock)
- [ ] Context passed as first parameter for cancellation/timeout
- [ ] No goroutine leaks (all spawned goroutines can terminate)

### Resource Management
- [ ] Resources (files, connections, etc.) properly closed
- [ ] defer used appropriately for cleanup
- [ ] Context properly propagated and respected
- [ ] No resource leaks in error paths

## Design Principles

### Simplicity
- [ ] Code is as simple as possible (but no simpler)
- [ ] Clever code avoided in favor of clear code
- [ ] No premature optimization
- [ ] Complex logic has explanatory comments

### Abstraction
- [ ] No premature abstraction (follows "Rule of Three")
- [ ] Abstractions have clear, single responsibility
- [ ] Interfaces are small and focused
- [ ] No unnecessary indirection

### Structure
- [ ] Functions are focused and do one thing well
- [ ] Function length is reasonable (<50 lines ideal, <100 acceptable)
- [ ] Cyclomatic complexity is reasonable
- [ ] Code is organized logically within files and packages

## Documentation

### Code Comments
- [ ] Complex logic has explanatory comments (why, not what)
- [ ] Non-obvious edge cases are documented
- [ ] Tradeoff decisions are explained
- [ ] TODOs include context and owner if present

### Godoc Comments
- [ ] All exported functions have godoc comments
- [ ] Comments start with the name of the element (e.g., "NewClient creates...")
- [ ] Package has package-level documentation
- [ ] Examples provided for non-trivial public APIs (if helpful)

### Inline Documentation
- [ ] Magic numbers replaced with named constants
- [ ] Complex regex/algorithms have explanatory comments
- [ ] Workarounds for known issues are documented

## Testing Quality

### Test Coverage
- [ ] New/modified code has tests
- [ ] Happy path tested
- [ ] Error paths tested
- [ ] Edge cases covered

### Test Design
- [ ] Test names follow convention: TestFunctionName_Scenario_ExpectedBehavior
- [ ] Tests focus on behavior, not implementation
- [ ] Table-driven tests used for multiple similar cases
- [ ] Test data is clear and self-documenting
- [ ] No brittle tests (over-mocking, tight coupling to implementation)

### Test Maintainability
- [ ] Tests are readable and easy to understand
- [ ] Test setup/teardown is clean
- [ ] No test pollution (tests don't affect each other)
- [ ] Mocks/stubs are minimal and necessary

## Performance Considerations

### Appropriate Choices
- [ ] Data structures appropriate for use case
- [ ] Algorithms have reasonable complexity
- [ ] No obvious performance bugs (N+1 queries, etc.)
- [ ] Allocations minimized in hot paths (if relevant)

### Premature Optimization
- [ ] No premature optimization without profiling data
- [ ] Readability not sacrificed for micro-optimizations
- [ ] Performance tradeoffs documented when made

## Security Considerations

### Input Validation
- [ ] User input validated at boundaries
- [ ] SQL injection prevented (use parameterized queries)
- [ ] XSS prevented (proper escaping)
- [ ] Path traversal prevented

### Sensitive Data
- [ ] No hardcoded credentials
- [ ] Secrets not logged
- [ ] Sensitive data properly encrypted/protected
- [ ] Authentication/authorization implemented correctly

### Common Vulnerabilities
- [ ] No obvious security issues (OWASP Top 10)
- [ ] Dependencies don't have known vulnerabilities
- [ ] Proper use of crypto libraries (no custom crypto)

## Code Health Assessment

### Overall Impression
- [ ] Code is readable and understandable
- [ ] Would you be comfortable modifying this code?
- [ ] Does this improve the codebase's overall health?
- [ ] Is this code maintainable long-term?

### Google's Standard
> "Favor approving a CL once it is in a state where it definitely improves the overall code health of the system being worked on, even if the CL isn't perfect."

Ask yourself: **Does this change make the codebase better than it was before?**

## Common Anti-Patterns to Flag

### Poor Error Handling
- Ignoring errors (using `_` without justification)
- Generic error messages without context
- Returning errors without wrapping/adding context

### Concurrency Issues
- Goroutine leaks
- Race conditions
- Deadlocks from incorrect mutex usage
- Not respecting context cancellation

### Over-Engineering
- Unnecessary abstractions
- Interfaces with single implementation (without clear reason)
- Complex design for simple requirements

### Under-Engineering
- Copy-pasted code that should be abstracted
- Missing error handling
- No tests for complex logic

## Reference Links

**Uber Go Style Guide:** Key patterns and conventions
**Google Go Style Guide:** Readability and clarity guidelines
**Effective Go:** Official Go best practices - https://go.dev/doc/effective_go
