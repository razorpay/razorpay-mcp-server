# Database Transaction Context Handling

## Critical Pattern: Transaction Context Usage

### The Problem

One of the most common and **critical** bugs in Go database code is using the wrong context inside database transaction functions. This can lead to:

- **Transaction isolation violations**
- **Incorrect timeout handling**
- **Context cancellation not propagated**
- **Difficult-to-debug race conditions**
- **Potential data corruption**

### The Pattern to Detect

```go
func (r *Repository) SomeOperation(ctx context.Context, params) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        // CRITICAL: All DB operations inside here MUST use 'tctx', NOT 'ctx'
        
        // ❌ WRONG - Using outer context 'ctx'
        result, err := r.GetData(ctx, id)
        
        // ✅ CORRECT - Using transaction context 'tctx'
        result, err := r.GetData(tctx, id)
        
        return nil
    })
}
```

### Why This Matters

1. **Transaction Isolation**: The transaction context `tctx` is bound to the specific database transaction. Using `ctx` may execute queries outside the transaction boundary.

2. **Timeout Handling**: The transaction may have its own timeout. Using `ctx` ignores this and uses the parent timeout instead.

3. **Cancellation Propagation**: If the transaction is rolled back, operations using `tctx` will be cancelled. Operations using `ctx` will continue executing.

4. **Connection Pooling**: The transaction context ensures all operations use the same database connection. Using `ctx` may use different connections.

## Detection Patterns

### Pattern 1: Direct Method Calls

```go
// ❌ WRONG
func (r *PaymentRepo) UpdatePayment(ctx context.Context, id string) error {
    return r.repo.Transaction(ctx, func(tctx context.Context) error {
        payment, err := r.repo.GetPayment(ctx, id)  // Using ctx!
        if err != nil {
            return err
        }
        
        return r.repo.Update(ctx, payment)  // Using ctx!
    })
}

// ✅ CORRECT
func (r *PaymentRepo) UpdatePayment(ctx context.Context, id string) error {
    return r.repo.Transaction(ctx, func(tctx context.Context) error {
        payment, err := r.repo.GetPayment(tctx, id)  // Using tctx!
        if err != nil {
            return err
        }
        
        return r.repo.Update(tctx, payment)  // Using tctx!
    })
}
```

### Pattern 2: Nested Function Calls

```go
// ❌ WRONG - Passing wrong context to helper
func (r *PaymentRepo) ProcessPayment(ctx context.Context, payment Payment) error {
    return r.repo.Transaction(ctx, func(tctx context.Context) error {
        // Passing ctx instead of tctx to helper!
        if err := r.validateAndSave(ctx, payment); err != nil {
            return err
        }
        
        return r.updateStatus(ctx, payment.ID, "completed")
    })
}

// ✅ CORRECT
func (r *PaymentRepo) ProcessPayment(ctx context.Context, payment Payment) error {
    return r.repo.Transaction(ctx, func(tctx context.Context) error {
        // Correctly passing tctx to helpers
        if err := r.validateAndSave(tctx, payment); err != nil {
            return err
        }
        
        return r.updateStatus(tctx, payment.ID, "completed")
    })
}
```

### Pattern 3: Goroutines Inside Transactions

```go
// ❌ WRONG - Goroutine using outer context
func (r *PaymentRepo) BatchProcess(ctx context.Context, ids []string) error {
    return r.repo.Transaction(ctx, func(tctx context.Context) error {
        for _, id := range ids {
            go func(paymentID string) {
                // Using ctx in goroutine - DANGEROUS!
                r.process(ctx, paymentID)
            }(id)
        }
        return nil
    })
}

// ✅ CORRECT - Don't use goroutines in transactions
// Or if you must, handle carefully with proper synchronization
func (r *PaymentRepo) BatchProcess(ctx context.Context, ids []string) error {
    return r.repo.Transaction(ctx, func(tctx context.Context) error {
        for _, id := range ids {
            // Sequential processing inside transaction
            if err := r.process(tctx, id); err != nil {
                return err
            }
        }
        return nil
    })
}
```

### Pattern 4: Context in Structs

```go
// ❌ WRONG - Storing wrong context
func (r *PaymentRepo) CreatePaymentFlow(ctx context.Context) error {
    return r.repo.Transaction(ctx, func(tctx context.Context) error {
        processor := &PaymentProcessor{
            ctx:  ctx,  // Storing outer context!
            repo: r.repo,
        }
        
        return processor.Process()
    })
}

// ✅ CORRECT
func (r *PaymentRepo) CreatePaymentFlow(ctx context.Context) error {
    return r.repo.Transaction(ctx, func(tctx context.Context) error {
        processor := &PaymentProcessor{
            ctx:  tctx,  // Storing transaction context!
            repo: r.repo,
        }
        
        return processor.Process()
    })
}
```

## Common Transaction Libraries

### GORM

```go
// GORM transaction pattern
func (r *Repo) Update(ctx context.Context) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Use tx, not r.db for all operations inside
        
        // ❌ WRONG
        if err := r.db.Create(&record).Error; err != nil {
            return err
        }
        
        // ✅ CORRECT
        if err := tx.Create(&record).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

### sqlx

```go
// sqlx transaction pattern
func (r *Repo) Update(ctx context.Context) error {
    tx, err := r.db.BeginTxx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Use tx for all operations
    // ❌ WRONG
    _, err = r.db.ExecContext(ctx, query, args...)
    
    // ✅ CORRECT
    _, err = tx.ExecContext(ctx, query, args...)
    
    return tx.Commit()
}
```

### Database/SQL

```go
// database/sql transaction pattern
func (r *Repo) Update(ctx context.Context) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Use tx.ExecContext, tx.QueryContext, etc.
    if _, err := tx.ExecContext(ctx, query1); err != nil {
        return err
    }
    
    if _, err := tx.ExecContext(ctx, query2); err != nil {
        return err
    }
    
    return tx.Commit()
}
```

## Review Checklist

When reviewing code, check for:

- [ ] All transaction functions receive a context parameter (usually named `tctx`)
- [ ] **All** method calls inside the transaction use `tctx`, not `ctx`
- [ ] Helper functions called from transactions receive and use `tctx`
- [ ] No goroutines are launched inside transactions (or if they are, they're handled correctly)
- [ ] Context is not stored in structs that span beyond the transaction
- [ ] All database operations use the transaction handle, not the main DB handle
- [ ] Proper defer rollback is implemented
- [ ] Transaction commits happen only after all operations succeed

## Red Flags to Report

Report these as **CRITICAL** issues:

1. Any use of outer `ctx` inside a transaction function body
2. Passing `ctx` instead of `tctx` to helper functions
3. Database operations using `r.db` instead of `tx` inside transactions
4. Goroutines created inside transactions that use `ctx`
5. Context stored in long-lived structs from transaction scope

## Automated Detection

Look for these patterns:

```go
// Pattern: Transaction function signature
r.Transaction(ctx, func(tctx context.Context) error {
    // Inside this block, search for:
    // 1. Any reference to 'ctx' (should be 'tctx')
    // 2. Method calls with 'ctx' as first parameter
    // 3. Function calls with 'ctx' as parameter
})
```

Regular expression patterns:
```
// Find transaction blocks
r\.Transaction\(ctx,\s*func\(tctx\s+context\.Context\)\s*error\s*{

// Inside those blocks, find ctx usage (should be flagged)
\bctx\b
```

## Impact Assessment

| Issue | Severity | Impact |
|-------|----------|--------|
| Using `ctx` instead of `tctx` in direct DB calls | CRITICAL | Data corruption, transaction isolation violated |
| Passing `ctx` to helper functions | CRITICAL | Partial transaction execution, inconsistent state |
| Goroutine with `ctx` in transaction | CRITICAL | Race conditions, deadlocks |
| Using `r.db` instead of `tx` | CRITICAL | Operations outside transaction |

## Examples from Real Code

### Example 1: Payment Processing

```go
// ❌ WRONG - Real bug found in payment service
func (s *PaymentService) ProcessRefund(ctx context.Context, refundReq RefundRequest) error {
    return s.repo.Transaction(ctx, func(tctx context.Context) error {
        // Bug: Using ctx instead of tctx
        payment, err := s.repo.GetPayment(ctx, refundReq.PaymentID)
        if err != nil {
            return err
        }
        
        // Bug: This creates refund outside transaction!
        refund, err := s.repo.CreateRefund(ctx, refundReq)
        if err != nil {
            return err
        }
        
        // Bug: Status update outside transaction
        return s.repo.UpdatePaymentStatus(ctx, payment.ID, "refunded")
    })
}

// ✅ CORRECT - Fixed version
func (s *PaymentService) ProcessRefund(ctx context.Context, refundReq RefundRequest) error {
    return s.repo.Transaction(ctx, func(tctx context.Context) error {
        // Fixed: Using tctx
        payment, err := s.repo.GetPayment(tctx, refundReq.PaymentID)
        if err != nil {
            return err
        }
        
        // Fixed: Refund created in transaction
        refund, err := s.repo.CreateRefund(tctx, refundReq)
        if err != nil {
            return err
        }
        
        // Fixed: Status update in transaction
        return s.repo.UpdatePaymentStatus(tctx, payment.ID, "refunded")
    })
}
```

### Example 2: Bulk Operations

```go
// ❌ WRONG - Mixing contexts
func (r *OrderRepo) BulkUpdateOrders(ctx context.Context, orderIDs []string, status string) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        for _, orderID := range orderIDs {
            // Bug: First call uses correct context
            order, err := r.GetOrder(tctx, orderID)
            if err != nil {
                return err
            }
            
            // Bug: But update uses wrong context!
            if err := r.UpdateOrderStatus(ctx, order.ID, status); err != nil {
                return err
            }
        }
        return nil
    })
}

// ✅ CORRECT
func (r *OrderRepo) BulkUpdateOrders(ctx context.Context, orderIDs []string, status string) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        for _, orderID := range orderIDs {
            // Consistent use of tctx
            order, err := r.GetOrder(tctx, orderID)
            if err != nil {
                return err
            }
            
            if err := r.UpdateOrderStatus(tctx, order.ID, status); err != nil {
                return err
            }
        }
        return nil
    })
}
```

## Best Practices

1. **Name transaction context distinctly**: Use `tctx`, `txCtx`, or `transactionCtx` to make it obvious
2. **Use linters**: Configure golangci-lint to detect context misuse
3. **Code review focus**: Make this a mandatory check in PR reviews
4. **Add comments**: Document that transaction context must be used
5. **Testing**: Write tests that verify transaction isolation

## Further Reading

- [Go Context Package](https://pkg.go.dev/context)
- [Database Transaction Best Practices](https://go.dev/doc/database/execute-transactions)
- [GORM Transactions](https://gorm.io/docs/transactions.html)

