# Idiomatic Go Code Patterns

## Package Organization

### Package Naming

```go
// ✅ Good - short, clear, singular
package payment
package user
package http

// ❌ Bad - plural, verbose
package payments
package userservice
package httputils
```

### Package Structure

```
✅ Good structure:
myapp/
  cmd/
    myapp/
      main.go
  internal/
    payment/
      payment.go
      payment_test.go
      repository.go
      service.go
    user/
      user.go
      user_test.go
  pkg/
    client/
      client.go
```

### Import Organization

```go
// ✅ Good - grouped by category
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // Third-party
    "github.com/pkg/errors"
    "go.uber.org/zap"
    
    // Internal
    "github.com/company/app/internal/payment"
)

// ❌ Bad - mixed, not grouped
import (
    "github.com/pkg/errors"
    "fmt"
    "github.com/company/app/internal/payment"
    "time"
)
```

## Variable Declaration

### Use Short Declarations

```go
// ✅ Good - short and clear
name := "John"
count := 10

// ❌ Bad - unnecessary verbosity
var name string = "John"
var count int = 10
```

### Use var for Zero Values

```go
// ✅ Good - implicit zero values
var (
    count   int
    message string
    ready   bool
)

// ❌ Bad - explicit zero values
count := 0
message := ""
ready := false
```

### Group Related Declarations

```go
// ✅ Good - grouped constants
const (
    StatusPending  = "pending"
    StatusApproved = "approved"
    StatusRejected = "rejected"
)

// ✅ Good - grouped variables
var (
    ErrNotFound = errors.New("not found")
    ErrInvalid  = errors.New("invalid")
)

// ❌ Bad - separate declarations
const StatusPending = "pending"
const StatusApproved = "approved"
const StatusRejected = "rejected"
```

## Functions

### Function Naming

```go
// ✅ Good - clear verb-based names
func GetUser(id string) (*User, error)
func CreatePayment(req PaymentRequest) (*Payment, error)
func ValidateInput(input string) bool
func IsValid() bool
func HasPermission() bool

// ❌ Bad - unclear or redundant
func UserGet(id string) (*User, error)
func MakePayment(req PaymentRequest) (*Payment, error)
func InputValidator(input string) bool
func CheckIfValid() bool
```

### Function Parameters

```go
// ✅ Good - context first, options last
func ProcessPayment(ctx context.Context, id string, opts ...Option) error

// ✅ Good - named result for documentation
func Divide(a, b int) (result int, err error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// ❌ Bad - too many parameters
func CreateUser(name, email, phone, address, city, country, zip string) error

// ✅ Good - use struct for many parameters
type UserRequest struct {
    Name    string
    Email   string
    Phone   string
    Address Address
}

func CreateUser(req UserRequest) error
```

### Early Returns

```go
// ✅ Good - early returns reduce nesting
func ProcessPayment(p *Payment) error {
    if p == nil {
        return ErrInvalidPayment
    }
    
    if p.Amount <= 0 {
        return ErrInvalidAmount
    }
    
    if p.Status != StatusPending {
        return ErrInvalidStatus
    }
    
    // Main logic here
    return process(p)
}

// ❌ Bad - deeply nested
func ProcessPayment(p *Payment) error {
    if p != nil {
        if p.Amount > 0 {
            if p.Status == StatusPending {
                // Main logic deeply nested
                return process(p)
            } else {
                return ErrInvalidStatus
            }
        } else {
            return ErrInvalidAmount
        }
    } else {
        return ErrInvalidPayment
    }
}
```

## Structs

### Struct Definition

```go
// ✅ Good - clear field types and tags
type Payment struct {
    ID        string    `json:"id" db:"id"`
    Amount    int64     `json:"amount" db:"amount"`
    Currency  string    `json:"currency" db:"currency"`
    Status    string    `json:"status" db:"status"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ✅ Good - embedding for composition
type Service struct {
    BaseService
    repo   Repository
    logger *zap.Logger
}

// ❌ Bad - unclear naming
type P struct {
    I string
    A int64
    C string
}
```

### Struct Initialization

```go
// ✅ Good - named fields
payment := &Payment{
    ID:       "pay_123",
    Amount:   1000,
    Currency: "USD",
    Status:   StatusPending,
}

// ❌ Bad - positional (fragile)
payment := &Payment{"pay_123", 1000, "USD", "pending", time.Now()}
```

### Constructor Functions

```go
// ✅ Good - New* constructor
func NewPaymentService(repo Repository, logger *zap.Logger) *PaymentService {
    return &PaymentService{
        repo:   repo,
        logger: logger,
    }
}

// ✅ Good - zero value is useful
type Config struct {
    Timeout time.Duration
    Retries int
}

func (c *Config) withDefaults() *Config {
    if c.Timeout == 0 {
        c.Timeout = 30 * time.Second
    }
    if c.Retries == 0 {
        c.Retries = 3
    }
    return c
}
```

## Interfaces

### Small Interfaces

```go
// ✅ Good - focused, single responsibility
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// Compose when needed
type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}

// ❌ Bad - too many methods
type DataService interface {
    Create(...) error
    Read(...) error
    Update(...) error
    Delete(...) error
    List(...) error
    Search(...) error
    Validate(...) error
}
```

### Accept Interfaces, Return Structs

```go
// ✅ Good
type PaymentProcessor interface {
    Process(ctx context.Context, p *Payment) error
}

func NewPaymentService(processor PaymentProcessor) *PaymentService {
    return &PaymentService{processor: processor}
}

// ❌ Bad - returning interface
func NewPaymentService() PaymentService {
    return &paymentService{}
}
```

## Methods

### Receiver Naming

```go
// ✅ Good - short, consistent (1-2 letters)
type Payment struct {
    ID     string
    Amount int64
}

func (p *Payment) SetStatus(status string) {
    p.Status = status
}

func (p *Payment) Validate() error {
    if p.Amount <= 0 {
        return ErrInvalidAmount
    }
    return nil
}

// ❌ Bad - inconsistent or verbose
func (payment *Payment) SetStatus(status string) {
    payment.Status = status
}

func (p *Payment) Validate() error {
    // Inconsistent with above
}
```

### Pointer vs Value Receivers

```go
// ✅ Good - pointer for mutation
func (p *Payment) SetAmount(amount int64) {
    p.Amount = amount
}

// ✅ Good - value for small immutable types
type Status string

func (s Status) IsValid() bool {
    return s == StatusPending || s == StatusApproved
}

// ❌ Bad - value receiver for mutation (doesn't work)
func (p Payment) SetAmount(amount int64) {
    p.Amount = amount // Only modifies copy!
}
```

## Error Handling

### Return Errors, Don't Panic

```go
// ✅ Good - return error
func GetPayment(id string) (*Payment, error) {
    payment, err := db.Find(id)
    if err != nil {
        return nil, fmt.Errorf("get payment: %w", err)
    }
    return payment, nil
}

// ❌ Bad - panic for regular errors
func GetPayment(id string) *Payment {
    payment, err := db.Find(id)
    if err != nil {
        panic(err)
    }
    return payment
}
```

### Handle Errors Immediately

```go
// ✅ Good
if err := doSomething(); err != nil {
    return fmt.Errorf("failed: %w", err)
}

// ❌ Bad - deferred error check
err := doSomething()
// ... many lines later ...
if err != nil {
    return err
}
```

## Nil Handling

### Nil Slice vs Empty Slice

```go
// ✅ Good - prefer nil for empty
func GetUsers() []User {
    // ... fetch from DB
    if len(users) == 0 {
        return nil // Not return []User{}
    }
    return users
}

// ✅ Good - nil-safe operations
var users []User // nil slice
users = append(users, User{}) // Works fine
length := len(users)          // Works fine (0)
```

### Nil Interface

```go
// ⚠️ Be careful - typed nil is not nil!
var user *User = nil
var any interface{} = user

if any == nil {
    // This is FALSE! any contains typed nil
}

// ✅ Good - check before assigning to interface
func GetUser() interface{} {
    var user *User
    // ... fetch user ...
    if user == nil {
        return nil // Return untyped nil
    }
    return user
}
```

## Strings

### Use strings.Builder for Concatenation

```go
// ✅ Good - efficient
var b strings.Builder
for _, word := range words {
    b.WriteString(word)
    b.WriteString(" ")
}
result := b.String()

// ❌ Bad - inefficient
result := ""
for _, word := range words {
    result += word + " "
}
```

### String Formatting

```go
// ✅ Good - use fmt.Sprintf
message := fmt.Sprintf("User %s has %d credits", user.Name, user.Credits)

// ❌ Bad - manual concatenation
message := "User " + user.Name + " has " + strconv.Itoa(user.Credits) + " credits"
```

## Slices and Maps

### Make with Capacity

```go
// ✅ Good - specify capacity when known
users := make([]User, 0, len(ids))
for _, id := range ids {
    users = append(users, fetchUser(id))
}

cache := make(map[string]Value, expectedSize)

// ❌ Bad - grows multiple times
users := []User{}
for _, id := range ids {
    users = append(users, fetchUser(id))
}
```

### Copy Slices

```go
// ✅ Good - proper copy
original := []int{1, 2, 3}
copied := make([]int, len(original))
copy(copied, original)

// ❌ Bad - both reference same array
original := []int{1, 2, 3}
copied := original
```

### Check Map Existence

```go
// ✅ Good - check existence
value, ok := myMap[key]
if !ok {
    // Key doesn't exist
}

// ❌ Bad - can't distinguish zero value from missing
value := myMap[key]
if value == 0 {
    // Is it zero or missing?
}
```

## Constants and Enums

### Use iota for Enums

```go
// ✅ Good - iota for auto-incrementing
type Status int

const (
    StatusPending Status = iota
    StatusApproved
    StatusRejected
)

// ✅ Good - with String() method
func (s Status) String() string {
    switch s {
    case StatusPending:
        return "pending"
    case StatusApproved:
        return "approved"
    case StatusRejected:
        return "rejected"
    default:
        return "unknown"
    }
}
```

### String Enums

```go
// ✅ Good - type safety with strings
type PaymentMethod string

const (
    PaymentMethodCard   PaymentMethod = "card"
    PaymentMethodBank   PaymentMethod = "bank"
    PaymentMethodWallet PaymentMethod = "wallet"
)

func (pm PaymentMethod) IsValid() bool {
    switch pm {
    case PaymentMethodCard, PaymentMethodBank, PaymentMethodWallet:
        return true
    }
    return false
}
```

## Defer

### Defer for Cleanup

```go
// ✅ Good - defer cleanup
func ReadFile(filename string) ([]byte, error) {
    f, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    
    return ioutil.ReadAll(f)
}

// ✅ Good - defer unlock
func (c *Cache) Update(key string, value Value) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.data[key] = value
}
```

### Defer Gotchas

```go
// ⚠️ Careful - defer evaluates arguments immediately
func example() {
    startTime := time.Now()
    defer log.Printf("took %v", time.Since(startTime)) // Calculates at defer time!
    
    // Do work...
}

// ✅ Good - use closure
func example() {
    startTime := time.Now()
    defer func() {
        log.Printf("took %v", time.Since(startTime)) // Calculates at defer execution
    }()
    
    // Do work...
}
```

## Comments

### Package Comments

```go
// ✅ Good - package doc comment
// Package payment provides functionality for processing payments.
// It supports multiple payment methods including cards, bank transfers,
// and digital wallets.
package payment
```

### Exported Names

```go
// ✅ Good - document exported functions
// ProcessPayment processes a payment request and returns the payment ID.
// It validates the request, creates a payment record, and initiates
// the payment with the provider.
//
// Returns ErrInvalidRequest if the request is invalid.
// Returns ErrInsufficientFunds if the account has insufficient balance.
func ProcessPayment(ctx context.Context, req PaymentRequest) (string, error) {
    // ...
}

// ❌ Bad - no documentation
func ProcessPayment(ctx context.Context, req PaymentRequest) (string, error) {
    // ...
}
```

### Inline Comments

```go
// ✅ Good - explain why, not what
// Use buffered channel to prevent goroutine leak if context is cancelled
ch := make(chan Result, 1)

// ❌ Bad - obvious comment
// Create a channel
ch := make(chan Result)
```

## Testing

### Table-Driven Tests

```go
// ✅ Good - table-driven
func TestValidatePayment(t *testing.T) {
    tests := []struct {
        name    string
        payment Payment
        wantErr bool
    }{
        {
            name:    "valid payment",
            payment: Payment{ID: "1", Amount: 100},
            wantErr: false,
        },
        {
            name:    "invalid amount",
            payment: Payment{ID: "1", Amount: -100},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePayment(tt.payment)
            if (err != nil) != tt.wantErr {
                t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Helpers

```go
// ✅ Good - helper with t.Helper()
func assertNoError(t *testing.T, err error) {
    t.Helper() // Marks this as helper for better error messages
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

// Usage
func TestSomething(t *testing.T) {
    result, err := DoSomething()
    assertNoError(t, err) // Error points to this line, not inside helper
}
```

## Code Organization

### Organize by Feature, Not Type

```
✅ Good - feature-based:
internal/
  payment/
    payment.go
    repository.go
    service.go
    handler.go
  user/
    user.go
    repository.go
    service.go
    handler.go

❌ Bad - type-based:
internal/
  models/
    payment.go
    user.go
  repositories/
    payment_repository.go
    user_repository.go
  services/
    payment_service.go
    user_service.go
```

### Group Related Code

```go
// ✅ Good - related constants together
const (
    // HTTP status codes
    StatusOK         = 200
    StatusBadRequest = 400
    StatusNotFound   = 404
    
    // Timeouts
    DefaultTimeout = 30 * time.Second
    MaxTimeout     = 60 * time.Second
)

// ✅ Good - blank line between groups
type PaymentService struct {
    repo   Repository
    logger *zap.Logger
    
    config Config
    cache  Cache
}
```

## Miscellaneous

### Use crypto/rand for Randomness

```go
// ✅ Good - cryptographically secure
import "crypto/rand"

func generateID() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}

// ❌ Bad - math/rand for security-sensitive
import "math/rand"

func generateToken() string {
    return fmt.Sprintf("%d", rand.Int63())
}
```

### Prefer strconv over fmt

```go
// ✅ Good - more efficient
s := strconv.Itoa(42)
i, err := strconv.Atoi("42")

// ❌ Bad - slower
s := fmt.Sprint(42)
fmt.Sscanf("42", "%d", &i)
```

## Review Checklist

- [ ] Package names are short, clear, singular
- [ ] Imports are grouped (stdlib, external, internal)
- [ ] Variables use short declaration (`:=`) when appropriate
- [ ] Zero values use `var` without explicit initialization
- [ ] Functions have clear, verb-based names
- [ ] Context is first parameter in functions
- [ ] Early returns reduce nesting
- [ ] Struct fields are exported with appropriate tags
- [ ] Interfaces are small (1-3 methods)
- [ ] Functions accept interfaces, return concrete types
- [ ] Method receivers are short and consistent
- [ ] Pointer receivers used for mutation or large structs
- [ ] Errors returned, not panicked (except programmer errors)
- [ ] `strings.Builder` used for string concatenation
- [ ] Slice/map capacity specified when size known
- [ ] Map existence checked with comma-ok idiom
- [ ] `defer` used for cleanup operations
- [ ] Exported functions have documentation comments
- [ ] Tests are table-driven where appropriate
- [ ] Code organized by feature, not by type

## References

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)

