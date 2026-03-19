# Go Testing Best Practices

## Test File Organization

### File Naming

```go
// ✅ Good - _test.go suffix
payment.go       // Implementation
payment_test.go  // Tests
```

### Package Naming

```go
// ✅ Good - same package for unit tests
package payment

func TestProcessPayment(t *testing.T) {
    // Can test unexported functions
}

// ✅ Good - _test package for integration tests
package payment_test

import (
    "testing"
    "github.com/company/app/payment"
)

func TestPaymentService(t *testing.T) {
    // Only tests exported API
}
```

## Table-Driven Tests

### Basic Pattern

```go
// ✅ Good - table-driven test
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
                Status: StatusPending,
            },
            wantErr: false,
        },
        {
            name: "missing ID",
            payment: Payment{
                Amount: 1000,
                Status: StatusPending,
            },
            wantErr: true,
        },
        {
            name: "negative amount",
            payment: Payment{
                ID:     "pay_123",
                Amount: -100,
                Status: StatusPending,
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

### Advanced Pattern

```go
// ✅ Good - with setup and verification
func TestProcessPayment(t *testing.T) {
    tests := []struct {
        name       string
        payment    Payment
        setupMock  func(*MockRepo)
        want       *Payment
        wantErr    bool
        checkError func(*testing.T, error)
    }{
        {
            name: "successful processing",
            payment: Payment{
                ID:     "pay_123",
                Amount: 1000,
            },
            setupMock: func(m *MockRepo) {
                m.EXPECT().Save(gomock.Any()).Return(nil)
            },
            want: &Payment{
                ID:     "pay_123",
                Amount: 1000,
                Status: StatusCompleted,
            },
            wantErr: false,
        },
        {
            name: "database error",
            payment: Payment{
                ID:     "pay_123",
                Amount: 1000,
            },
            setupMock: func(m *MockRepo) {
                m.EXPECT().Save(gomock.Any()).Return(errors.New("db error"))
            },
            wantErr: true,
            checkError: func(t *testing.T, err error) {
                if !strings.Contains(err.Error(), "db error") {
                    t.Errorf("expected db error, got %v", err)
                }
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            repo := NewMockRepo(ctrl)
            if tt.setupMock != nil {
                tt.setupMock(repo)
            }
            
            service := NewPaymentService(repo)
            got, err := service.Process(context.Background(), tt.payment)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.checkError != nil && err != nil {
                tt.checkError(t, err)
            }
            
            if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Process() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Test Helpers

### Helper Functions

```go
// ✅ Good - mark as helper
func assertNoError(t *testing.T, err error) {
    t.Helper() // Error points to caller, not this line
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func assertEqual(t *testing.T, got, want interface{}) {
    t.Helper()
    if !reflect.DeepEqual(got, want) {
        t.Errorf("got %v, want %v", got, want)
    }
}

// Usage
func TestSomething(t *testing.T) {
    result, err := DoSomething()
    assertNoError(t, err)
    assertEqual(t, result, expected)
}
```

### Setup and Teardown

```go
// ✅ Good - TestMain for global setup
func TestMain(m *testing.M) {
    // Setup
    db = setupTestDatabase()
    cache = setupTestCache()
    
    // Run tests
    code := m.Run()
    
    // Teardown
    db.Close()
    cache.Close()
    
    os.Exit(code)
}

// ✅ Good - per-test setup
func setupTest(t *testing.T) (*Database, func()) {
    t.Helper()
    
    db := createTestDB(t)
    
    cleanup := func() {
        db.Close()
    }
    
    return db, cleanup
}

func TestQuery(t *testing.T) {
    db, cleanup := setupTest(t)
    defer cleanup()
    
    // Test with db
}
```

## Mocking

### Interface-Based Mocking

```go
// ✅ Good - define interface
type PaymentRepository interface {
    Save(ctx context.Context, payment *Payment) error
    Get(ctx context.Context, id string) (*Payment, error)
}

// ✅ Good - simple mock
type MockPaymentRepo struct {
    SaveFunc func(ctx context.Context, payment *Payment) error
    GetFunc  func(ctx context.Context, id string) (*Payment, error)
}

func (m *MockPaymentRepo) Save(ctx context.Context, payment *Payment) error {
    if m.SaveFunc != nil {
        return m.SaveFunc(ctx, payment)
    }
    return nil
}

func (m *MockPaymentRepo) Get(ctx context.Context, id string) (*Payment, error) {
    if m.GetFunc != nil {
        return m.GetFunc(ctx, id)
    }
    return nil, nil
}

// Usage in tests
func TestProcessPayment(t *testing.T) {
    mock := &MockPaymentRepo{
        SaveFunc: func(ctx context.Context, payment *Payment) error {
            if payment.Amount <= 0 {
                return errors.New("invalid amount")
            }
            return nil
        },
    }
    
    service := NewPaymentService(mock)
    err := service.Process(context.Background(), &Payment{Amount: 100})
    assertNoError(t, err)
}
```

### Using testify/mock

```go
import (
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/assert"
)

// ✅ Good - testify mock
type MockPaymentRepo struct {
    mock.Mock
}

func (m *MockPaymentRepo) Save(ctx context.Context, payment *Payment) error {
    args := m.Called(ctx, payment)
    return args.Error(0)
}

func (m *MockPaymentRepo) Get(ctx context.Context, id string) (*Payment, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Payment), args.Error(1)
}

// Usage
func TestProcessPayment(t *testing.T) {
    mock := new(MockPaymentRepo)
    mock.On("Save", mock.Anything, mock.Anything).Return(nil)
    
    service := NewPaymentService(mock)
    err := service.Process(context.Background(), &Payment{Amount: 100})
    
    assert.NoError(t, err)
    mock.AssertExpectations(t)
}
```

## Test Coverage

### Measuring Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out

# View in terminal
go tool cover -func=coverage.out

# View in browser
go tool cover -html=coverage.out

# Coverage for specific package
go test -coverprofile=coverage.out ./internal/payment

# Coverage for all packages
go test -coverprofile=coverage.out ./...
```

### Coverage Requirements

```go
// ✅ Good - test error paths
func TestValidatePayment(t *testing.T) {
    tests := []struct {
        name    string
        payment Payment
        wantErr bool
    }{
        {name: "valid", payment: validPayment(), wantErr: false},
        {name: "nil", payment: Payment{}, wantErr: true},
        {name: "negative amount", payment: Payment{Amount: -100}, wantErr: true},
        {name: "invalid status", payment: Payment{Status: "invalid"}, wantErr: true},
        // Cover all error conditions
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePayment(tt.payment)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Integration Tests

### Database Tests

```go
// ✅ Good - integration test with real database
func TestPaymentRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewPaymentRepository(db)
    
    // Test Create
    payment := &Payment{
        ID:     "pay_test_123",
        Amount: 1000,
    }
    
    err := repo.Save(context.Background(), payment)
    if err != nil {
        t.Fatalf("failed to save: %v", err)
    }
    
    // Test Read
    retrieved, err := repo.Get(context.Background(), "pay_test_123")
    if err != nil {
        t.Fatalf("failed to get: %v", err)
    }
    
    if retrieved.Amount != payment.Amount {
        t.Errorf("amount = %d, want %d", retrieved.Amount, payment.Amount)
    }
}

// Run only unit tests:
// go test -short

// Run all tests including integration:
// go test
```

### HTTP Tests

```go
// ✅ Good - HTTP handler test
func TestPaymentHandler(t *testing.T) {
    handler := NewPaymentHandler(mockService)
    
    req := httptest.NewRequest("POST", "/payments", strings.NewReader(`{
        "amount": 1000,
        "currency": "USD"
    }`))
    req.Header.Set("Content-Type", "application/json")
    
    rec := httptest.NewRecorder()
    
    handler.ServeHTTP(rec, req)
    
    if rec.Code != http.StatusOK {
        t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
    }
    
    var response PaymentResponse
    if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }
    
    if response.Amount != 1000 {
        t.Errorf("amount = %d, want 1000", response.Amount)
    }
}
```

## Benchmarks

### Basic Benchmark

```go
// ✅ Good - benchmark function
func BenchmarkProcessPayment(b *testing.B) {
    payment := &Payment{
        ID:     "pay_123",
        Amount: 1000,
    }
    
    b.ResetTimer() // Don't count setup time
    for i := 0; i < b.N; i++ {
        ProcessPayment(payment)
    }
}

// Run: go test -bench=.
// Run with memory stats: go test -bench=. -benchmem
```

### Table Benchmarks

```go
// ✅ Good - benchmark different scenarios
func BenchmarkSum(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
            data := make([]int, size)
            for i := range data {
                data[i] = i
            }
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                Sum(data)
            }
        })
    }
}
```

### Parallel Benchmarks

```go
// ✅ Good - benchmark concurrent access
func BenchmarkCacheGetParallel(b *testing.B) {
    cache := NewCache()
    cache.Set("key", "value")
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            cache.Get("key")
        }
    })
}
```

## Testing Patterns

### Golden Files

```go
// ✅ Good - golden file testing
func TestRender(t *testing.T) {
    data := Data{Name: "Test", Value: 123}
    
    got := Render(data)
    
    goldenFile := "testdata/render_golden.txt"
    
    if *update {
        os.WriteFile(goldenFile, []byte(got), 0644)
    }
    
    want, err := os.ReadFile(goldenFile)
    if err != nil {
        t.Fatalf("failed to read golden file: %v", err)
    }
    
    if got != string(want) {
        t.Errorf("Render() = %v, want %v", got, string(want))
    }
}

// Update golden files:
// go test -update
```

### Fuzz Testing

```go
// ✅ Good - fuzz test (Go 1.18+)
func FuzzValidateEmail(f *testing.F) {
    // Seed corpus
    f.Add("test@example.com")
    f.Add("invalid")
    f.Add("")
    
    f.Fuzz(func(t *testing.T, email string) {
        // Should not panic
        ValidateEmail(email)
    })
}

// Run: go test -fuzz=FuzzValidateEmail
```

## Testing Anti-Patterns

### Don't Test Implementation

```go
// ❌ Bad - testing implementation details
func TestProcessPayment_CallsValidate(t *testing.T) {
    mock := &MockValidator{}
    service := NewPaymentService(mock)
    
    service.Process(payment)
    
    if !mock.ValidateCalled {
        t.Error("expected Validate to be called")
    }
}

// ✅ Good - test behavior
func TestProcessPayment_RejectsInvalidPayment(t *testing.T) {
    service := NewPaymentService()
    
    err := service.Process(invalidPayment)
    
    if !errors.Is(err, ErrInvalidPayment) {
        t.Errorf("expected ErrInvalidPayment, got %v", err)
    }
}
```

### Don't Use time.Sleep

```go
// ❌ Bad - using sleep
func TestAsyncOperation(t *testing.T) {
    go doAsync()
    time.Sleep(100 * time.Millisecond) // Flaky!
    checkResult()
}

// ✅ Good - use synchronization
func TestAsyncOperation(t *testing.T) {
    done := make(chan bool)
    go func() {
        doAsync()
        done <- true
    }()
    
    select {
    case <-done:
        checkResult()
    case <-time.After(time.Second):
        t.Fatal("timeout")
    }
}
```

### Don't Ignore Errors in Tests

```go
// ❌ Bad - ignoring setup errors
func TestSomething(t *testing.T) {
    db, _ := setupDB() // What if this fails?
    // Test continues...
}

// ✅ Good - fail fast on setup errors
func TestSomething(t *testing.T) {
    db, err := setupDB()
    if err != nil {
        t.Fatalf("setup failed: %v", err)
    }
    // Test continues...
}
```

## Test Organization

### Group Related Tests

```go
// ✅ Good - subtests for grouping
func TestPaymentService(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        t.Run("ValidPayment", func(t *testing.T) {
            // ...
        })
        
        t.Run("InvalidAmount", func(t *testing.T) {
            // ...
        })
    })
    
    t.Run("Get", func(t *testing.T) {
        t.Run("ExistingPayment", func(t *testing.T) {
            // ...
        })
        
        t.Run("NotFound", func(t *testing.T) {
            // ...
        })
    })
}

// Run specific subtest:
// go test -run TestPaymentService/Create/ValidPayment
```

## Testing Documentation

### Document Test Intent

```go
// ✅ Good - clear test documentation
func TestProcessPayment_RetriesOnTransientError(t *testing.T) {
    // Given: a service that returns a transient error twice
    callCount := 0
    mock := &MockService{
        ProcessFunc: func() error {
            callCount++
            if callCount <= 2 {
                return ErrTransient
            }
            return nil
        },
    }
    
    // When: processing a payment
    err := ProcessPayment(mock)
    
    // Then: should succeed after retries
    if err != nil {
        t.Errorf("expected success after retries, got %v", err)
    }
    
    if callCount != 3 {
        t.Errorf("expected 3 attempts, got %d", callCount)
    }
}
```

## Review Checklist

- [ ] Test files named with `_test.go` suffix
- [ ] Table-driven tests for multiple scenarios
- [ ] Helper functions marked with `t.Helper()`
- [ ] All error paths tested
- [ ] Edge cases covered (nil, empty, boundary values)
- [ ] Mocks/stubs used for external dependencies
- [ ] Integration tests tagged with `testing.Short()`
- [ ] Benchmarks for performance-critical code
- [ ] Tests don't depend on execution order
- [ ] Tests clean up resources (defer cleanup)
- [ ] No `time.Sleep()` in tests
- [ ] Setup errors cause test failure
- [ ] Test names clearly describe what's tested
- [ ] Tests are deterministic (no flakiness)
- [ ] Coverage adequate (>80% for critical paths)

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [testify](https://github.com/stretchr/testify)
- [gomock](https://github.com/golang/mock)
- [Go Fuzzing](https://go.dev/security/fuzz/)

