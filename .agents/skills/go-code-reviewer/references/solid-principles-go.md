# SOLID Principles for Go Code Review

## Overview

SOLID principles adapted for Go's interface-centric design and composition-based architecture. Use this checklist during Layer 3 quality assessment.

---

## Single Responsibility Principle (SRP)

**Definition**: Each function, struct, or package should have one reason to change.

### Checklist
- [ ] Functions have a single, focused purpose
- [ ] Structs encapsulate one cohesive set of data/behavior
- [ ] Packages have clear, focused responsibilities

### Examples

**❌ Violation**:
```go
// ProcessAndSaveOrder does too much
func ProcessAndSaveOrder(order Order) error {
    // Validates order
    if err := validateOrder(order); err != nil {
        return err
    }

    // Calculates pricing
    total := calculateTotal(order.Items)

    // Saves to database
    if err := db.Save(order); err != nil {
        return err
    }

    // Sends email
    return emailService.SendConfirmation(order)
}
```

**✅ Correct**:
```go
// Each function has single responsibility
func ValidateOrder(order Order) error { ... }
func CalculateTotal(items []Item) float64 { ... }
func SaveOrder(order Order) error { ... }
func SendOrderConfirmation(order Order) error { ... }

// Orchestrator delegates to focused functions
func ProcessOrder(order Order) error {
    if err := ValidateOrder(order); err != nil {
        return err
    }
    order.Total = CalculateTotal(order.Items)
    if err := SaveOrder(order); err != nil {
        return err
    }
    return SendOrderConfirmation(order)
}
```

---

## Open/Closed Principle (OCP)

**Definition**: Software entities should be open for extension but closed for modification.

### Checklist
- [ ] Interfaces allow behavior extension without modifying existing code
- [ ] New functionality added via new implementations, not code changes
- [ ] Composition used over modification

### Examples

**❌ Violation**:
```go
func ProcessPayment(payment Payment, method string) error {
    switch method {
    case "credit_card":
        return processCreditCard(payment)
    case "paypal":
        return processPayPal(payment)
    // Adding UPI requires modifying this function
    default:
        return errors.New("unknown payment method")
    }
}
```

**✅ Correct**:
```go
// Interface allows extension without modification
type PaymentProcessor interface {
    Process(payment Payment) error
}

// New payment methods = new implementations
type CreditCardProcessor struct{}
func (c *CreditCardProcessor) Process(payment Payment) error { ... }

type PayPalProcessor struct{}
func (p *PayPalProcessor) Process(payment Payment) error { ... }

// Adding UPI = new struct implementing interface (no modification)
type UPIProcessor struct{}
func (u *UPIProcessor) Process(payment Payment) error { ... }
```

---

## Liskov Substitution Principle (LSP)

**Definition**: Interface implementations must maintain behavioral contracts.

### Checklist
- [ ] All implementations of an interface behave consistently
- [ ] No implementation throws unexpected errors
- [ ] Implementations maintain semantic compatibility

### Examples

**❌ Violation**:
```go
type Repository interface {
    Get(ctx context.Context, id string) (*Entity, error)
}

// Violation: Returns nil without error (breaks contract)
type CacheRepo struct{}
func (c *CacheRepo) Get(ctx context.Context, id string) (*Entity, error) {
    entity := cache.Lookup(id)
    // Returns nil, nil instead of error for missing entity
    return entity, nil
}
```

**✅ Correct**:
```go
type Repository interface {
    Get(ctx context.Context, id string) (*Entity, error)
}

type CacheRepo struct{}
func (c *CacheRepo) Get(ctx context.Context, id string) (*Entity, error) {
    entity := cache.Lookup(id)
    if entity == nil {
        // Returns proper error like other implementations
        return nil, ErrNotFound
    }
    return entity, nil
}
```

---

## Interface Segregation Principle (ISP)

**Definition**: Clients should not depend on interfaces they don't use. Keep interfaces small and focused.

### Checklist
- [ ] Interfaces have minimal methods (ideally 1-3)
- [ ] No "fat interfaces" forcing unnecessary implementations
- [ ] Clients depend only on methods they need

### Examples

**❌ Violation**:
```go
// Fat interface - clients forced to implement unused methods
type Repository interface {
    Get(id string) (*Entity, error)
    List() ([]*Entity, error)
    Create(e *Entity) error
    Update(e *Entity) error
    Delete(id string) error
    Validate(e *Entity) error
    Transform(e *Entity) *DTO
}
```

**✅ Correct**:
```go
// Segregated interfaces - clients implement only what they need
type Reader interface {
    Get(id string) (*Entity, error)
    List() ([]*Entity, error)
}

type Writer interface {
    Create(e *Entity) error
    Update(e *Entity) error
    Delete(id string) error
}

// Compose when both needed
type ReadWriter interface {
    Reader
    Writer
}
```

---

## Dependency Inversion Principle (DIP)

**Definition**: Depend on abstractions (interfaces), not concretions (structs).

### Checklist
- [ ] High-level modules depend on interfaces
- [ ] Concrete implementations injected at runtime
- [ ] No direct instantiation of dependencies

### Examples

**❌ Violation**:
```go
// Directly depends on concrete PostgresRepo
type UserService struct {
    repo *PostgresRepo  // Tightly coupled
}

func NewUserService() *UserService {
    return &UserService{
        repo: &PostgresRepo{},  // Cannot swap implementation
    }
}
```

**✅ Correct**:
```go
// Depends on interface abstraction
type UserService struct {
    repo Repository  // Interface dependency
}

// Implementation injected via constructor
func NewUserService(repo Repository) *UserService {
    return &UserService{
        repo: repo,  // Can inject any Repository implementation
    }
}

// Usage
postgresRepo := &PostgresRepo{}
service := NewUserService(postgresRepo)

// Easy to swap for testing or different implementation
mockRepo := &MockRepo{}
testService := NewUserService(mockRepo)
```

---

## Review Template

Use this template during Layer 3 assessment:

```markdown
### SOLID Principles Assessment

**Single Responsibility**:
- ✅ Functions focused (e.g., `ValidateOrder` only validates)
- ⚠️ `ProcessBatch()` handles validation + processing (consider splitting)

**Open/Closed**:
- ✅ Interface-based design allows extension
- ✅ New entities added via proto annotations (no code modification)

**Liskov Substitution**:
- ✅ All Repository implementations maintain contracts
- ✅ No unexpected nil returns or panics

**Interface Segregation**:
- ✅ Interfaces minimal (e.g., `IRepo` has focused methods)
- ✅ No fat interfaces forcing unused method implementations

**Dependency Inversion**:
- ✅ Services depend on interface abstractions
- ✅ Concrete implementations injected via constructors
- ✅ Easy to mock for testing
```

---

## Common Go-Specific Patterns

### Accept Interfaces, Return Structs
```go
// ✅ Correct: Accept interface, return concrete type
func ProcessOrder(repo Repository) (*Order, error) {
    // Implementation
}
```

### Small Interfaces
```go
// ✅ Go idiom: Interfaces of 1-2 methods
type Stringer interface {
    String() string
}

type Reader interface {
    Read(p []byte) (n int, err error)
}
```

### Composition Over Inheritance
```go
// ✅ Compose behaviors using embedding
type LoggedRepo struct {
    Repository  // Embed interface
    logger Logger
}
```

---

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Proverbs](https://go-proverbs.github.io/)
- [SOLID Principles in Go](https://dave.cheney.net/2016/08/20/solid-go-design)
