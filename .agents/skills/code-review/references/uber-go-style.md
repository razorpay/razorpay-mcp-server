# Uber Go Style Guide - Key Principles

This document summarizes the most important principles from the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) for code review purposes.

## Error Handling

### Error Wrapping

Always wrap errors with context:

```go
// ❌ Bad
if err != nil {
    return err
}

// ✅ Good
if err != nil {
    return fmt.Errorf("failed to process payment %s: %w", paymentID, err)
}
```

### Error Types

Define custom error types for complex errors:

```go
// ✅ Good
type ValidationError struct {
    Field string
    Value interface{}
    Reason string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Reason)
}
```

### Sentinel Errors

Use sentinel errors for expected error conditions:

```go
// ✅ Good
var (
    ErrNotFound = errors.New("payment not found")
    ErrInvalidStatus = errors.New("invalid payment status")
    ErrInsufficientFunds = errors.New("insufficient funds")
)

// Usage
if err := validatePayment(p); errors.Is(err, ErrInsufficientFunds) {
    // Handle insufficient funds
}
```

### Error Checking

Check errors immediately:

```go
// ❌ Bad
result, err := DoSomething()
// ... many lines later ...
if err != nil {
    return err
}

// ✅ Good
result, err := DoSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

## Pointer Receivers vs Value Receivers

### Use Pointer Receivers When:

1. Method modifies the receiver
2. Receiver is a large struct
3. Consistency (if one method uses pointer, all should)

```go
// ✅ Good - modifies receiver
func (p *Payment) SetStatus(status string) {
    p.Status = status
    p.UpdatedAt = time.Now()
}

// ✅ Good - large struct
type LargeConfig struct {
    // ... many fields
}

func (c *LargeConfig) Validate() error {
    // Read-only but uses pointer for efficiency
}
```

### Use Value Receivers When:

1. Method doesn't modify receiver
2. Receiver is a small struct or primitive type
3. Receiver is a map, function, or channel

```go
// ✅ Good - small immutable type
type Status string

func (s Status) IsValid() bool {
    return s == "active" || s == "inactive"
}
```

## Interfaces

### Accept Interfaces, Return Structs

```go
// ❌ Bad - returning interface
func NewPaymentService() PaymentService {
    return &paymentService{}
}

// ✅ Good - returning struct
func NewPaymentService() *PaymentService {
    return &PaymentService{}
}

// ❌ Bad - accepting struct
func ProcessPayment(service *PaymentService, payment Payment) error {
    // ...
}

// ✅ Good - accepting interface
func ProcessPayment(service PaymentProcessor, payment Payment) error {
    // ...
}
```

### Interface Size

Keep interfaces small (ideally 1-3 methods):

```go
// ✅ Good - focused interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// ❌ Bad - too many responsibilities
type DataManager interface {
    Read() ([]byte, error)
    Write([]byte) error
    Validate() error
    Transform() error
    Encrypt() error
    Decrypt() error
}
```

### Zero-Value Interfaces

```go
// ✅ Good - valid zero value
type Config struct {
    Logger Logger // nil logger is valid
}

func (c *Config) log(msg string) {
    if c.Logger != nil {
        c.Logger.Log(msg)
    }
}
```

## Initialization

### Use `var` for Zero Values

```go
// ✅ Good
var (
    count int
    name  string
    ready bool
)

// ❌ Bad - explicit zero values
count := 0
name := ""
ready := false
```

### Struct Initialization

```go
// ✅ Good - field names for clarity
payment := Payment{
    ID:        id,
    Amount:    amount,
    Currency:  "USD",
    Status:    "pending",
    CreatedAt: time.Now(),
}

// ❌ Bad - positional initialization
payment := Payment{id, amount, "USD", "pending", time.Now()}
```

## Reduce Nesting

### Early Returns

```go
// ❌ Bad - deeply nested
func ProcessPayment(p Payment) error {
    if p.ID != "" {
        if p.Amount > 0 {
            if p.Status == "pending" {
                // do work
                return nil
            } else {
                return ErrInvalidStatus
            }
        } else {
            return ErrInvalidAmount
        }
    } else {
        return ErrInvalidID
    }
}

// ✅ Good - early returns
func ProcessPayment(p Payment) error {
    if p.ID == "" {
        return ErrInvalidID
    }
    if p.Amount <= 0 {
        return ErrInvalidAmount
    }
    if p.Status != "pending" {
        return ErrInvalidStatus
    }
    
    // do work
    return nil
}
```

## Unnecessary Else

```go
// ❌ Bad
func Status(err error) string {
    if err != nil {
        return "error"
    } else {
        return "success"
    }
}

// ✅ Good
func Status(err error) string {
    if err != nil {
        return "error"
    }
    return "success"
}
```

## Goroutines

### Always Provide Exit Condition

```go
// ❌ Bad - no way to stop
go func() {
    for {
        work()
    }
}()

// ✅ Good - can be stopped
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            work()
        }
    }
}()
```

### Use WaitGroups for Coordination

```go
// ✅ Good
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        process(i)
    }(item)
}
wg.Wait()
```

### Don't Start Goroutines in Libraries

```go
// ❌ Bad - library controls goroutine
func (s *Service) Watch() {
    go func() {
        for {
            s.poll()
            time.Sleep(time.Second)
        }
    }()
}

// ✅ Good - caller controls goroutine
func (s *Service) Poll() {
    s.poll()
}

// Caller decides to use goroutine
go func() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for range ticker.C {
        service.Poll()
    }
}()
```

## Channel Size

### Zero or One

```go
// ✅ Good - unbuffered (synchronous)
ch := make(chan int)

// ✅ Good - buffer of 1 (one async send)
ch := make(chan int, 1)

// ❌ Bad - magic number
ch := make(chan int, 64)

// ✅ Good if needed - named constant with justification
const (
    // workerPoolSize is the number of concurrent workers
    // based on profiling results showing optimal throughput
    workerPoolSize = 10
)
ch := make(chan int, workerPoolSize)
```

## Enums

### Use Type and Const

```go
// ✅ Good
type Status string

const (
    StatusPending   Status = "pending"
    StatusApproved  Status = "approved"
    StatusRejected  Status = "rejected"
)

// Add validation
func (s Status) IsValid() bool {
    switch s {
    case StatusPending, StatusApproved, StatusRejected:
        return true
    }
    return false
}
```

### Use iota for Integer Enums

```go
// ✅ Good
type Priority int

const (
    PriorityLow Priority = iota + 1
    PriorityMedium
    PriorityHigh
    PriorityCritical
)
```

## Time

### Use `time.Time` for Instants

```go
// ❌ Bad
type Config struct {
    CreatedAt int64 // Unix timestamp
}

// ✅ Good
type Config struct {
    CreatedAt time.Time
}
```

### Use `time.Duration` for Periods

```go
// ❌ Bad
func Wait(seconds int) {
    time.Sleep(time.Duration(seconds) * time.Second)
}

// ✅ Good
func Wait(duration time.Duration) {
    time.Sleep(duration)
}

// Usage
Wait(5 * time.Second)
Wait(100 * time.Millisecond)
```

## Naming

### Package Names

- Short, concise, lowercase
- No underscores or mixedCaps
- Not plural

```go
// ✅ Good
package payment
package user
package http

// ❌ Bad
package payments
package user_service
package httpServer
```

### Function Names

```go
// ✅ Good
func ProcessPayment()
func GetUser()
func NewService()

// ❌ Bad
func process_payment() // snake_case
func GETUSER()         // all caps
func get_user()        // snake_case
```

### Variable Names

```go
// ✅ Good - short names for short scopes
for i, item := range items {
    // i and item are clear in context
}

// ✅ Good - descriptive names for package level
var (
    ErrPaymentNotFound = errors.New("payment not found")
    DefaultTimeout     = 30 * time.Second
)

// ❌ Bad - too short for wide scope
var (
    e = errors.New("payment not found")
    t = 30 * time.Second
)
```

### Receiver Names

```go
// ✅ Good - short, consistent
func (p *Payment) SetStatus(status string) {
    p.Status = status
}

func (p *Payment) Validate() error {
    return validatePayment(p)
}

// ❌ Bad - inconsistent
func (p *Payment) SetStatus(status string) {
    p.Status = status
}

func (payment *Payment) Validate() error {
    return validatePayment(payment)
}

// ❌ Bad - using 'this' or 'self'
func (this *Payment) SetStatus(status string) {
    this.Status = status
}
```

## Grouping

### Group Similar Declarations

```go
// ✅ Good
const (
    StatusPending  = "pending"
    StatusApproved = "approved"
    StatusRejected = "rejected"
)

var (
    ErrNotFound = errors.New("not found")
    ErrInvalid  = errors.New("invalid")
)

// ❌ Bad
const StatusPending = "pending"
const StatusApproved = "approved"
const StatusRejected = "rejected"
```

### Organize Imports

```go
// ✅ Good - grouped by standard, external, internal
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // External packages
    "github.com/pkg/errors"
    "go.uber.org/zap"
    
    // Internal packages
    "github.com/company/project/internal/payment"
    "github.com/company/project/internal/user"
)
```

## Testing

### Table-Driven Tests

```go
// ✅ Good
func TestValidatePayment(t *testing.T) {
    tests := []struct {
        name    string
        payment Payment
        wantErr bool
    }{
        {
            name: "valid payment",
            payment: Payment{
                ID:     "pay_123",
                Amount: 1000,
            },
            wantErr: false,
        },
        {
            name: "missing ID",
            payment: Payment{
                Amount: 1000,
            },
            wantErr: true,
        },
        {
            name: "zero amount",
            payment: Payment{
                ID:     "pay_123",
                Amount: 0,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePayment(tt.payment)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidatePayment() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Functional Options

```go
// ✅ Good - flexible configuration
type Option func(*Service)

func WithTimeout(timeout time.Duration) Option {
    return func(s *Service) {
        s.timeout = timeout
    }
}

func WithLogger(logger *zap.Logger) Option {
    return func(s *Service) {
        s.logger = logger
    }
}

func NewService(opts ...Option) *Service {
    s := &Service{
        timeout: DefaultTimeout,
        logger:  zap.NewNop(),
    }
    
    for _, opt := range opts {
        opt(s)
    }
    
    return s
}

// Usage
service := NewService(
    WithTimeout(30 * time.Second),
    WithLogger(logger),
)
```

## Performance

### Prefer strconv over fmt

```go
// ❌ Bad
s := fmt.Sprint(123)

// ✅ Good
s := strconv.Itoa(123)
```

### Avoid String Concatenation

```go
// ❌ Bad - many concatenations
s := "Hello"
s += " "
s += "World"
s += "!"

// ✅ Good - use strings.Builder
var b strings.Builder
b.WriteString("Hello")
b.WriteString(" ")
b.WriteString("World")
b.WriteString("!")
s := b.String()
```

### Specify Map Capacity

```go
// ❌ Bad
m := make(map[string]int)
for _, item := range items {
    m[item.Key] = item.Value
}

// ✅ Good
m := make(map[string]int, len(items))
for _, item := range items {
    m[item.Key] = item.Value
}
```

### Specify Slice Capacity

```go
// ❌ Bad
var results []Result
for _, item := range items {
    results = append(results, process(item))
}

// ✅ Good
results := make([]Result, 0, len(items))
for _, item := range items {
    results = append(results, process(item))
}
```

## Review Checklist

When reviewing code for Uber Go Style compliance:

- [ ] Errors are wrapped with context using `%w`
- [ ] Sentinel errors use `errors.Is()` for comparison
- [ ] Pointer vs value receivers are used appropriately
- [ ] Interfaces are small and focused
- [ ] Functions accept interfaces, return concrete types
- [ ] Early returns are used to reduce nesting
- [ ] Unnecessary `else` blocks are removed
- [ ] Goroutines have clear exit conditions
- [ ] Channel sizes are justified (preferably 0 or 1)
- [ ] `time.Time` and `time.Duration` are used correctly
- [ ] Naming follows Go conventions
- [ ] Imports are properly grouped
- [ ] Tests use table-driven approach
- [ ] Map and slice capacities are specified when size is known
- [ ] String concatenation uses `strings.Builder` for efficiency

## References

- [Uber Go Style Guide (Official)](https://github.com/uber-go/guide/blob/master/style.md)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

