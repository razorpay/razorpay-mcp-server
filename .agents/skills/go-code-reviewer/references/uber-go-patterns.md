# Uber Go Style Guide - Key Patterns

Quick reference for common patterns from the Uber Go Style Guide.

## Error Handling

### Verify Error Performance
When calling a function multiple times, verify errors are handled consistently:

```go
// Bad
if err := doSomething(); err != nil {
    // handle
}
if err := doSomethingElse(); err != nil {
    // different handling
}

// Good - consistent error handling
for _, item := range items {
    if err := process(item); err != nil {
        return fmt.Errorf("process %v: %w", item, err)
    }
}
```

## Containers

### Specify Container Capacity
Specify capacity for maps and slices when size is known:

```go
// Bad
m := make(map[string]int)

// Good
m := make(map[string]int, len(items))

// Bad
s := make([]int, 0)

// Good
s := make([]int, 0, len(items))
```

## Control Flow

### Reduce Nesting
Use early returns and continue to reduce nesting:

```go
// Bad
func process(item Item) error {
    if item.IsValid() {
        if item.IsReady() {
            // do work
            return nil
        } else {
            return errors.New("not ready")
        }
    } else {
        return errors.New("invalid")
    }
}

// Good
func process(item Item) error {
    if !item.IsValid() {
        return errors.New("invalid")
    }
    if !item.IsReady() {
        return errors.New("not ready")
    }
    // do work
    return nil
}
```

### Unnecessary Else
Omit else when if block returns:

```go
// Bad
if condition {
    return x
} else {
    return y
}

// Good
if condition {
    return x
}
return y
```

## Variables

### Local Variable Declarations
Use `:=` for local variables:

```go
// Bad
var s = "foo"

// Good
s := "foo"
```

### Reduce Scope of Variables
Declare variables in smallest scope:

```go
// Bad
err := doSomething()
if err != nil {
    return err
}

// Good
if err := doSomething(); err != nil {
    return err
}
```

### Avoid Naked Parameters
Use comments or constants for boolean/numeric parameters:

```go
// Bad
client.Connect(true, 30)

// Good
client.Connect(
    true,  // autoReconnect
    30,    // timeoutSeconds
)

// Better
const (
    autoReconnect = true
    timeout = 30 * time.Second
)
client.Connect(autoReconnect, timeout)
```

## Strings

### Use Raw String Literals
Use backticks for strings with escapes:

```go
// Bad
path := "C:\\Program Files\\Application"

// Good
path := `C:\Program Files\Application`
```

## Source: https://github.com/uber-go/guide
