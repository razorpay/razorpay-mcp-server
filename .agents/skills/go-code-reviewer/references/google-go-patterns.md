# Google Go Style Guide - Key Principles

Quick reference for core principles from the Google Go Style Guide.

## Core Philosophy

**Primary Goals:**
1. **Clarity** - Code should be easy to understand
2. **Simplicity** - Prefer simple solutions over complex ones
3. **Consistency** - Follow established patterns

> "Code is written once but read many times" - Optimize for readability.

## Naming

### General Principles
- Short names in small scopes: `i`, `r`, `w`
- Longer names in larger scopes: `userRepository`, `httpClient`
- No stuttering: `user.UserID` → `user.ID`

### Packages
- Lowercase, single word: `http`, `json`, `template`
- No underscores or mixedCaps
- Package name is part of the identifier: `json.Encoder` not `json.JSONEncoder`

### Interfaces
- Single method interfaces: name + "er": `Reader`, `Writer`, `Formatter`
- Focus on behavior: what it does, not what it is

## Error Handling

### Adding Context
```go
// Bad
if err != nil {
    return err
}

// Good
if err != nil {
    return fmt.Errorf("process user %d: %w", userID, err)
}
```

### Error Values
- Errors should be actionable
- Include what failed and why
- Preserve the original error with %w

## Functions

### Keep Functions Short
- Ideal: <50 lines
- Acceptable: <100 lines
- If longer: consider splitting

### Single Responsibility
Each function should do one thing well:

```go
// Bad
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // parse input
    // validate
    // business logic
    // database operations
    // format response
}

// Good - split into focused functions
func handleRequest(w http.ResponseWriter, r *http.Request) {
    input, err := parseInput(r)
    // ...
    if err := validate(input); err != nil {
        // ...
    }
    result, err := processRequest(input)
    // ...
    writeResponse(w, result)
}
```

## Concurrency

### Context Usage
- Context is the first parameter
- Always respect context cancellation
- Propagate context through call chains

```go
// Good
func ProcessItems(ctx context.Context, items []Item) error {
    for _, item := range items {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := processItem(ctx, item); err != nil {
                return err
            }
        }
    }
    return nil
}
```

### Goroutine Lifecycle
- Every goroutine should have a clear termination condition
- Document when/how goroutines terminate
- Avoid goroutine leaks

## Testing

### Test Names
Pattern: `TestFunctionName_StateUnderTest_ExpectedBehavior`

```go
func TestAdd_PositiveNumbers_ReturnsSum(t *testing.T)
func TestDivide_ByZero_ReturnsError(t *testing.T)
```

### Table-Driven Tests
For multiple similar cases:

```go
func TestCalculate(t *testing.T) {
    tests := []struct {
        name    string
        input   int
        want    int
        wantErr bool
    }{
        {"positive", 5, 10, false},
        {"zero", 0, 0, false},
        {"negative", -5, -10, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Calculate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("want error %v, got %v", tt.wantErr, err)
            }
            if got != tt.want {
                t.Errorf("want %v, got %v", tt.want, got)
            }
        })
    }
}
```

## Documentation

### Godoc Comments
- Start with the name of the thing being documented
- Full sentence, ending with period
- Don't say "This function..." or "This method..."

```go
// Bad
// This function creates a new client
func NewClient() *Client

// Good
// NewClient creates and initializes a new Client.
func NewClient() *Client
```

### When to Comment
- **Why**, not **what** (code shows what)
- Non-obvious decisions
- Workarounds
- Complex algorithms

```go
// Bad
// Increment i by 1
i++

// Good
// Skip the first element as it's the header
for i := 1; i < len(rows); i++ {
```

## Common Review Focus Areas

From Google's code review guidelines:

1. **Design** - Is the code well-designed and appropriate for the system?
2. **Functionality** - Does the code behave as intended? Is it good for users?
3. **Complexity** - Could it be simpler? Would another developer understand it?
4. **Tests** - Correct, sensible, useful tests?
5. **Naming** - Clear variable, function, class names?
6. **Comments** - Clear, useful, necessary?
7. **Style** - Follows Go style guide?
8. **Documentation** - Updated if needed?

## The Code Health Standard

> "Favor approving a CL once it is in a state where it definitely improves the overall code health of the system being worked on, even if the CL isn't perfect."

**Key question:** Does this make the codebase better than before?

Not: Is this perfect?

## Source: https://google.github.io/styleguide/go/
