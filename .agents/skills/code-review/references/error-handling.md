# Go Error Handling Patterns

## Error Handling Philosophy

Go's error handling is explicit and straightforward. Errors are values that should be handled at the appropriate level with proper context.

## Basic Error Handling

### Check Errors Immediately

```go
// ✅ Good
result, err := DoSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
// Use result

// ❌ Bad - delayed check
result, err := DoSomething()
// ... other code ...
if err != nil {
    return err
}
```

### Don't Ignore Errors

```go
// ❌ Bad - ignoring error
result, _ := DoSomething()

// ✅ Good - handle or log
result, err := DoSomething()
if err != nil {
    log.Printf("failed to do something: %v", err)
    // Decide what to do
}

// ✅ Good - if truly doesn't matter (rare)
result, _ := DoSomething() // error safely ignored because X
```

## Error Wrapping

### Use `%w` for Error Wrapping

```go
// ✅ Good - preserves error chain
if err := processPayment(payment); err != nil {
    return fmt.Errorf("failed to process payment %s: %w", payment.ID, err)
}

// ❌ Bad - breaks error chain
if err := processPayment(payment); err != nil {
    return fmt.Errorf("failed to process payment %s: %v", payment.ID, err)
}
```

### Error Unwrapping

```go
// Check wrapped errors
if err := doSomething(); err != nil {
    if errors.Is(err, ErrNotFound) {
        // Handle not found
    }
}

// Extract specific error type
var validationErr *ValidationError
if errors.As(err, &validationErr) {
    // Handle validation error
    fmt.Printf("Field %s failed: %s\n", validationErr.Field, validationErr.Reason)
}
```

## Sentinel Errors

### Define Package-Level Errors

```go
// ✅ Good - exported sentinel errors
var (
    ErrNotFound          = errors.New("resource not found")
    ErrAlreadyExists     = errors.New("resource already exists")
    ErrInvalidInput      = errors.New("invalid input")
    ErrUnauthorized      = errors.New("unauthorized access")
    ErrInsufficientFunds = errors.New("insufficient funds")
)

// Usage
func GetPayment(id string) (*Payment, error) {
    payment, err := db.Find(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get payment: %w", err)
    }
    return payment, nil
}

// Caller
payment, err := GetPayment("pay_123")
if errors.Is(err, ErrNotFound) {
    // Handle not found specifically
}
```

### Naming Convention

```go
// ✅ Good - starts with Err
var (
    ErrInvalidRequest = errors.New("invalid request")
    ErrTimeout        = errors.New("operation timed out")
)

// ❌ Bad - doesn't follow convention
var (
    InvalidRequest = errors.New("invalid request")
    TimedOut       = errors.New("operation timed out")
)
```

## Custom Error Types

### Structured Errors

```go
// ✅ Good - rich error information
type ValidationError struct {
    Field   string
    Value   interface{}
    Reason  string
    Details map[string]string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Reason)
}

// Usage
func ValidatePayment(p *Payment) error {
    if p.Amount <= 0 {
        return &ValidationError{
            Field:  "amount",
            Value:  p.Amount,
            Reason: "amount must be positive",
        }
    }
    return nil
}

// Caller
if err := ValidatePayment(payment); err != nil {
    var validationErr *ValidationError
    if errors.As(err, &validationErr) {
        // Access structured information
        log.Printf("Field: %s, Reason: %s", validationErr.Field, validationErr.Reason)
    }
}
```

### HTTP Error Types

```go
// ✅ Good - HTTP-aware errors
type HTTPError struct {
    StatusCode int
    Message    string
    Err        error
}

func (e *HTTPError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("HTTP %d: %s: %v", e.StatusCode, e.Message, e.Err)
    }
    return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

func (e *HTTPError) Unwrap() error {
    return e.Err
}

// Helper constructors
func NewBadRequest(message string) *HTTPError {
    return &HTTPError{StatusCode: 400, Message: message}
}

func NewNotFound(resource string) *HTTPError {
    return &HTTPError{
        StatusCode: 404,
        Message:    fmt.Sprintf("%s not found", resource),
    }
}

func NewInternalError(err error) *HTTPError {
    return &HTTPError{
        StatusCode: 500,
        Message:    "internal server error",
        Err:        err,
    }
}

// Usage in handler
func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    if id == "" {
        writeError(w, NewBadRequest("payment ID is required"))
        return
    }
    
    payment, err := h.service.GetPayment(r.Context(), id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            writeError(w, NewNotFound("payment"))
            return
        }
        writeError(w, NewInternalError(err))
        return
    }
    
    writeJSON(w, payment)
}
```

## Error Context

### Add Meaningful Context

```go
// ❌ Bad - no context
if err != nil {
    return err
}

// ❌ Bad - vague context
if err != nil {
    return fmt.Errorf("error: %w", err)
}

// ✅ Good - specific context
if err != nil {
    return fmt.Errorf("failed to create payment for user %s with amount %d: %w", 
        userID, amount, err)
}
```

### Context at Each Layer

```go
// Repository layer
func (r *PaymentRepo) Create(ctx context.Context, payment *Payment) error {
    if err := r.db.Create(payment).Error; err != nil {
        return fmt.Errorf("db create payment %s: %w", payment.ID, err)
    }
    return nil
}

// Service layer
func (s *PaymentService) ProcessPayment(ctx context.Context, req PaymentRequest) error {
    payment := &Payment{
        ID:     generateID(),
        Amount: req.Amount,
    }
    
    if err := s.repo.Create(ctx, payment); err != nil {
        return fmt.Errorf("process payment for user %s: %w", req.UserID, err)
    }
    return nil
}

// Handler layer
func (h *PaymentHandler) HandlePayment(w http.ResponseWriter, r *http.Request) {
    var req PaymentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    
    if err := h.service.ProcessPayment(r.Context(), req); err != nil {
        log.Printf("handle payment request: %v", err)
        http.Error(w, "failed to process payment", http.StatusInternalServerError)
        return
    }
}
```

## Error Handling Patterns

### Multiple Return Values

```go
// ✅ Good - error as last return value
func GetPayment(id string) (*Payment, error) {
    // ...
}

func ValidateAndSave(payment *Payment) error {
    // ...
}

// ❌ Bad - error not last
func GetPayment(id string) (error, *Payment) {
    // ...
}
```

### Defer for Cleanup

```go
// ✅ Good - cleanup with defer
func ProcessFile(filename string) error {
    f, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("open file %s: %w", filename, err)
    }
    defer f.Close()
    
    // Process file
    // Even if processing fails, file will be closed
    
    return nil
}

// ✅ Good - multiple defers
func ProcessPayment(ctx context.Context, payment *Payment) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback() // Rollback does nothing if already committed
    
    span := trace.StartSpan(ctx, "process_payment")
    defer span.End()
    
    // Process payment
    
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    
    return nil
}
```

### Panic and Recover

```go
// ❌ Bad - using panic for normal errors
func GetPayment(id string) *Payment {
    payment, err := db.Find(id)
    if err != nil {
        panic(err) // Don't do this!
    }
    return payment
}

// ✅ Good - panic only for programmer errors
func MustCompileRegex(pattern string) *regexp.Regexp {
    re, err := regexp.Compile(pattern)
    if err != nil {
        // Pattern is hardcoded, this is a programmer error
        panic(fmt.Sprintf("invalid regex pattern %s: %v", pattern, err))
    }
    return re
}

// Usage in package initialization
var emailRegex = MustCompileRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ✅ Good - recover from panics in goroutines
func (w *Worker) Run() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("worker panic: %v\n%s", r, debug.Stack())
            // Potentially restart worker
        }
    }()
    
    // Worker code
}
```

## Error Aggregation

### Collecting Multiple Errors

```go
// ✅ Good - using multierr package
import "go.uber.org/multierr"

func ValidatePayment(p *Payment) error {
    var err error
    
    if p.ID == "" {
        err = multierr.Append(err, errors.New("ID is required"))
    }
    
    if p.Amount <= 0 {
        err = multierr.Append(err, errors.New("amount must be positive"))
    }
    
    if p.Currency == "" {
        err = multierr.Append(err, errors.New("currency is required"))
    }
    
    return err
}

// ✅ Good - custom error list
type ErrorList []error

func (e ErrorList) Error() string {
    if len(e) == 0 {
        return ""
    }
    if len(e) == 1 {
        return e[0].Error()
    }
    
    var b strings.Builder
    b.WriteString("multiple errors:\n")
    for i, err := range e {
        fmt.Fprintf(&b, "  %d. %v\n", i+1, err)
    }
    return b.String()
}

func (e ErrorList) Errors() []error {
    return []error(e)
}
```

## Testing Error Handling

### Test Error Cases

```go
func TestGetPayment_NotFound(t *testing.T) {
    repo := NewMockRepo()
    repo.SetError(sql.ErrNoRows)
    
    service := NewPaymentService(repo)
    
    _, err := service.GetPayment(context.Background(), "pay_123")
    if !errors.Is(err, ErrNotFound) {
        t.Errorf("expected ErrNotFound, got %v", err)
    }
}

func TestValidatePayment_InvalidAmount(t *testing.T) {
    payment := &Payment{
        ID:     "pay_123",
        Amount: -100, // Invalid
    }
    
    err := ValidatePayment(payment)
    if err == nil {
        t.Error("expected validation error for negative amount")
    }
    
    var validationErr *ValidationError
    if !errors.As(err, &validationErr) {
        t.Error("expected ValidationError type")
    }
    
    if validationErr.Field != "amount" {
        t.Errorf("expected field 'amount', got '%s'", validationErr.Field)
    }
}
```

### Table-Driven Error Tests

```go
func TestProcessPayment(t *testing.T) {
    tests := []struct {
        name       string
        payment    *Payment
        setupMock  func(*MockRepo)
        wantErr    bool
        checkError func(t *testing.T, err error)
    }{
        {
            name: "success",
            payment: &Payment{
                ID:     "pay_123",
                Amount: 1000,
            },
            setupMock: func(m *MockRepo) {
                m.CreateSuccess()
            },
            wantErr: false,
        },
        {
            name: "validation error",
            payment: &Payment{
                ID:     "pay_123",
                Amount: -100,
            },
            wantErr: true,
            checkError: func(t *testing.T, err error) {
                var validationErr *ValidationError
                if !errors.As(err, &validationErr) {
                    t.Error("expected ValidationError")
                }
            },
        },
        {
            name: "database error",
            payment: &Payment{
                ID:     "pay_123",
                Amount: 1000,
            },
            setupMock: func(m *MockRepo) {
                m.SetError(errors.New("db error"))
            },
            wantErr: true,
            checkError: func(t *testing.T, err error) {
                if !strings.Contains(err.Error(), "db") {
                    t.Error("expected database-related error")
                }
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := NewMockRepo()
            if tt.setupMock != nil {
                tt.setupMock(repo)
            }
            
            service := NewPaymentService(repo)
            err := service.ProcessPayment(context.Background(), tt.payment)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessPayment() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if err != nil && tt.checkError != nil {
                tt.checkError(t, err)
            }
        })
    }
}
```

## Anti-Patterns

### Don't Panic in Libraries

```go
// ❌ Bad - library panicking
func GetPayment(id string) *Payment {
    payment, err := db.Find(id)
    if err != nil {
        panic(err)
    }
    return payment
}

// ✅ Good - return error
func GetPayment(id string) (*Payment, error) {
    payment, err := db.Find(id)
    if err != nil {
        return nil, fmt.Errorf("get payment: %w", err)
    }
    return payment, nil
}
```

### Don't Use `_` for Errors

```go
// ❌ Bad - ignoring errors
payment, _ := GetPayment("pay_123")

// ✅ Good - handle errors
payment, err := GetPayment("pay_123")
if err != nil {
    log.Printf("failed to get payment: %v", err)
    return
}
```

### Don't Log and Return

```go
// ❌ Bad - logging and returning (causes duplicate logs)
func ProcessPayment(payment *Payment) error {
    if err := validate(payment); err != nil {
        log.Printf("validation failed: %v", err)
        return err // Caller will also log this
    }
    return nil
}

// ✅ Good - either log OR return
func ProcessPayment(payment *Payment) error {
    if err := validate(payment); err != nil {
        return fmt.Errorf("validate payment: %w", err)
    }
    return nil
}

// Log at the top level
func main() {
    if err := ProcessPayment(payment); err != nil {
        log.Printf("failed to process payment: %v", err)
    }
}
```

### Don't Use Generic Error Messages

```go
// ❌ Bad
return errors.New("error")
return errors.New("failed")
return fmt.Errorf("something went wrong")

// ✅ Good
return errors.New("payment ID is required")
return fmt.Errorf("invalid payment amount: %d", amount)
return fmt.Errorf("failed to create payment for user %s: %w", userID, err)
```

## Review Checklist

- [ ] All errors are checked immediately after function calls
- [ ] Errors are not ignored (no `_` for errors without justification)
- [ ] Error wrapping uses `%w` to preserve error chain
- [ ] Sentinel errors are defined at package level with `Err` prefix
- [ ] Custom error types implement `Error()` method
- [ ] Custom error types implement `Unwrap()` if wrapping another error
- [ ] Error messages provide meaningful context
- [ ] Each layer adds appropriate context when wrapping errors
- [ ] Panic is only used for programmer errors, not runtime errors
- [ ] Defer is used appropriately for cleanup
- [ ] Errors are tested with both `errors.Is()` and `errors.As()`
- [ ] No logging-and-returning pattern (choose one)
- [ ] Error messages are specific and actionable

## References

- [Go Blog: Error Handling](https://go.dev/blog/error-handling-and-go)
- [Go Blog: Working with Errors](https://go.dev/blog/go1.13-errors)
- [errors package](https://pkg.go.dev/errors)
- [fmt.Errorf documentation](https://pkg.go.dev/fmt#Errorf)

