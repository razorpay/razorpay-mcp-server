# Query & Index Analysis Checks (Static)

## Overview

Validates that database queries have appropriate indexes by analyzing code patterns and migration files. This is **static analysis** during PR review - no live database connection required.

**Load when:** PR modifies database queries, adds WHERE/JOIN clauses, or changes migration files

**Total Checks:** 6

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 3
- 📋 Medium: 1

---

## Scope & Limitations

### ✅ What This Check DOES (Static Analysis)
- Scans code for query patterns (WHERE, JOIN, ORDER BY)
- Checks migration files for index definitions
- Validates foreign key columns have indexes
- Detects N+1 query patterns
- Checks composite indexes match query patterns

### ❌ What This Check CANNOT DO (Requires Live DB)
- Execute EXPLAIN/EXPLAIN ANALYZE
- Measure actual query performance
- Check index usage statistics
- Benchmark query plans

**For dynamic analysis,** use separate `postgres-query-analyzer` or `mysql-query-analyzer` skill (similar to trino-analyzer).

---

## Check 1: Foreign Key Columns Have Indexes 🚨 CRITICAL

### What to Check

All foreign key columns must have indexes to prevent table scans on JOIN operations.

### Razorpay Context

Foreign keys without indexes cause severe performance degradation on:
- JOINs
- DELETE CASCADE operations
- Foreign key constraint checks

### Bad Pattern ❌

```go
// Migration: Add foreign key
func AddGatewayCredentials(tx *gorm.DB) error {
    type GatewayCredential struct {
        ID         string `gorm:"primaryKey"`
        GatewayID  string `gorm:"column:gateway_id"`  // ❌ Foreign key, no index!
        MerchantID string `gorm:"column:merchant_id"` // ❌ Foreign key, no index!
        Acquirer   string `gorm:"column:acquirer"`
    }

    tx.AutoMigrate(&GatewayCredential{})

    // ❌ Foreign keys created but NO indexes!
    tx.Exec(`ALTER TABLE gateway_credentials
             ADD CONSTRAINT fk_gateway
             FOREIGN KEY (gateway_id) REFERENCES gateways(id)`)

    return nil
}

// Query: JOIN on unindexed foreign key
SELECT gc.*, g.name
FROM gateway_credentials gc
JOIN gateways g ON gc.gateway_id = g.id  -- ❌ Table scan on gateway_credentials!
WHERE gc.merchant_id = 'merch_123'       -- ❌ Another table scan!
```

**Problem:**
- Full table scan on every JOIN
- DELETE CASCADE scans entire table
- Constraint checks are slow
- Performance degrades as table grows

### Good Pattern ✅

```go
// CORRECT: Foreign keys with indexes (PostgreSQL)
func AddGatewayCredentials(tx *gorm.DB) error {
    type GatewayCredential struct {
        ID         string `gorm:"primaryKey"`
        GatewayID  string `gorm:"column:gateway_id;index"`  // ✅ Index added
        MerchantID string `gorm:"column:merchant_id;index"` // ✅ Index added
        Acquirer   string `gorm:"column:acquirer"`
    }

    tx.AutoMigrate(&GatewayCredential{})

    // ✅ Explicit index creation
    tx.Exec(`CREATE INDEX IF NOT EXISTS idx_gateway_credentials_gateway_id
             ON gateway_credentials(gateway_id)`)

    tx.Exec(`CREATE INDEX IF NOT EXISTS idx_gateway_credentials_merchant_id
             ON gateway_credentials(merchant_id)`)

    // Add foreign key (index already exists)
    tx.Exec(`ALTER TABLE gateway_credentials
             ADD CONSTRAINT fk_gateway
             FOREIGN KEY (gateway_id) REFERENCES gateways(id)`)

    return nil
}

// Query now uses indexes
SELECT gc.*, g.name
FROM gateway_credentials gc
JOIN gateways g ON gc.gateway_id = g.id  -- ✅ Index scan
WHERE gc.merchant_id = 'merch_123'       -- ✅ Index scan
```

### Detection Strategy

```bash
# 1. Find foreign key definitions in migrations
grep -rn "FOREIGN KEY\|REFERENCES\|foreignKey" <migration_files>

# 2. Extract foreign key columns
# Example: FOREIGN KEY (gateway_id) REFERENCES gateways(id)
#          → Column: gateway_id

# 3. Check if index exists on that column
grep -rn "CREATE INDEX.*gateway_id\|index.*gateway_id" <migration_files>

# 4. Flag if foreign key exists but no index
```

### Flag Conditions

Flag if:
- `FOREIGN KEY (column_name)` exists in migration
- No `CREATE INDEX` on `column_name`
- GORM struct has foreign key tag but no `index` tag
- Column name ends with `_id` and used in JOIN but no index

### Severity

🚨 **Critical** - Performance catastrophe:
- Table scans on every JOIN
- DELETE CASCADE locks entire table
- Query time grows O(n) instead of O(log n)
- Production incidents under load

---

## Check 2: WHERE Clause Columns Have Indexes ⚠️ HIGH

### What to Check

Columns frequently used in WHERE clauses must have indexes.

### Bad Pattern ❌

```go
// Migration: No index on status
type Payment struct {
    ID        string `gorm:"primaryKey"`
    Status    string `gorm:"column:status"`     // ❌ No index!
    CreatedAt time.Time
}

// Query: Filter by unindexed column
db.Where("status = ?", "pending").Find(&payments)
// ❌ Full table scan! Scans millions of rows
```

### Good Pattern ✅

```go
// CORRECT: Index on frequently filtered column
type Payment struct {
    ID        string `gorm:"primaryKey"`
    Status    string `gorm:"column:status;index"`  // ✅ Indexed
    CreatedAt time.Time `gorm:"index"`             // ✅ Indexed
}

// Migration
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);

// Query uses index
db.Where("status = ?", "pending").Find(&payments)  // ✅ Index scan
```

### Detection Strategy

```bash
# 1. Find WHERE clauses in code
grep -rn "\.Where\|WHERE" <pr_files> --include="*.go"

# 2. Extract column names
# .Where("status = ?", ...) → Column: status
# WHERE merchant_id = ? → Column: merchant_id

# 3. Check if column has index in migrations
grep "CREATE INDEX.*status\|index.*status" <migration_files>

# 4. Flag if frequently queried but no index
```

### Severity

⚠️ **High** - Performance degradation:
- Full table scans
- Slow API responses
- High database CPU
- Poor user experience

---

## Check 3: Composite Indexes Match Query Patterns ⚠️ HIGH

### What to Check

Multi-column WHERE clauses need composite indexes with correct column order.

### Bad Pattern ❌

```go
// Migration: Separate indexes
CREATE INDEX idx_merchant_id ON transactions(merchant_id);
CREATE INDEX idx_status ON transactions(status);

// Query: Two columns in WHERE
SELECT * FROM transactions
WHERE merchant_id = 'merch_123'  -- Uses idx_merchant_id
  AND status = 'pending'         -- ❌ Can't use idx_status (already filtered)
-- Result: Index scan on merchant_id + filter on status (not optimal)
```

### Good Pattern ✅

```go
// CORRECT: Composite index matching query pattern
CREATE INDEX idx_transactions_merchant_status
ON transactions(merchant_id, status);

// Query benefits from composite index
SELECT * FROM transactions
WHERE merchant_id = 'merch_123'
  AND status = 'pending'
-- ✅ Composite index scan (optimal)

// Order matters!
// ✅ Works: WHERE merchant_id = ? AND status = ?
// ✅ Works: WHERE merchant_id = ?
// ❌ Doesn't work well: WHERE status = ? (can't use index efficiently)

// CORRECT: Multiple composite indexes for different query patterns
-- Pattern 1: Filter by merchant + status
CREATE INDEX idx_merchant_status ON transactions(merchant_id, status);

-- Pattern 2: Filter by status + created_at (reports)
CREATE INDEX idx_status_created ON transactions(status, created_at);

-- Pattern 3: Order by created_at within merchant
CREATE INDEX idx_merchant_created ON transactions(merchant_id, created_at);
```

### Column Order Rules

**PostgreSQL/MySQL B-tree indexes:**
1. Most selective column first (highest cardinality)
2. Equality filters before range filters
3. Columns in ORDER BY should match index order

```sql
-- ✅ GOOD: Equality first, range second
CREATE INDEX idx_merchant_date ON orders(merchant_id, created_at);
WHERE merchant_id = ? AND created_at > ?  -- Uses index efficiently

-- ❌ BAD: Range first, equality second
CREATE INDEX idx_date_merchant ON orders(created_at, merchant_id);
WHERE merchant_id = ? AND created_at > ?  -- Index less effective
```

### Detection Strategy

```bash
# 1. Find multi-column WHERE clauses
grep -A 5 "WHERE" <pr_files> | grep "AND\|OR"

# 2. Extract column combinations
# WHERE merchant_id = ? AND status = ? → Columns: merchant_id, status

# 3. Check for composite index in migrations
grep "CREATE INDEX.*merchant_id.*status" <migration_files>

# 4. Flag if no composite index for common query pattern
```

### Severity

⚠️ **High** - Suboptimal performance:
- Partial index usage
- Higher I/O than necessary
- Slower queries under load

---

## Check 4: N+1 Query Detection 🚨 CRITICAL

### What to Check

Queries inside loops cause N+1 query problem (1 query + N queries in loop).

### Razorpay Context

Common in:
- Fetching related entities (merchants, terminals, gateways)
- API response building
- Report generation

### Bad Pattern ❌

```go
// ANTI-PATTERN: N+1 queries
func GetMerchantsWithGateways(merchantIDs []string) ([]*MerchantInfo, error) {
    var results []*MerchantInfo

    // 1 query: Fetch merchants
    var merchants []Merchant
    db.Where("id IN ?", merchantIDs).Find(&merchants)

    for _, merchant := range merchants {
        // ❌ N queries: 1 query per merchant!
        var gateways []Gateway
        db.Where("merchant_id = ?", merchant.ID).Find(&gateways)

        results = append(results, &MerchantInfo{
            Merchant: merchant,
            Gateways: gateways,
        })
    }
    // Total: 1 + N queries (if N=100, that's 101 queries!)

    return results, nil
}
```

**Problem:**
- 100 merchants = 101 database queries
- High latency (each query ~10ms → 1000ms total)
- Database connection pool exhaustion
- Doesn't scale

### Good Pattern ✅

```go
// CORRECT: Batch fetch (2 queries total)
func GetMerchantsWithGateways(merchantIDs []string) ([]*MerchantInfo, error) {
    // Query 1: Fetch all merchants
    var merchants []Merchant
    db.Where("id IN ?", merchantIDs).Find(&merchants)

    // Query 2: Fetch all gateways for these merchants (batch)
    var gateways []Gateway
    db.Where("merchant_id IN ?", merchantIDs).Find(&gateways)

    // Group gateways by merchant_id in memory
    gatewayMap := make(map[string][]Gateway)
    for _, gw := range gateways {
        gatewayMap[gw.MerchantID] = append(gatewayMap[gw.MerchantID], gw)
    }

    // Build results
    var results []*MerchantInfo
    for _, merchant := range merchants {
        results = append(results, &MerchantInfo{
            Merchant: merchant,
            Gateways: gatewayMap[merchant.ID],
        })
    }

    return results, nil
}

// CORRECT: GORM Preload (2 queries)
func GetMerchantsWithGateways(merchantIDs []string) ([]Merchant, error) {
    var merchants []Merchant
    // ✅ GORM automatically batches the gateway query
    db.Preload("Gateways").Where("id IN ?", merchantIDs).Find(&merchants)
    return merchants, nil
}

// CORRECT: JOIN (1 query)
func GetMerchantsWithGateways(merchantIDs []string) ([]*MerchantInfo, error) {
    var results []struct {
        Merchant
        Gateway
    }

    db.Table("merchants m").
        Select("m.*, g.*").
        Joins("LEFT JOIN gateways g ON g.merchant_id = m.id").
        Where("m.id IN ?", merchantIDs).
        Scan(&results)

    // Group by merchant...
}
```

### Detection Strategy

```bash
# 1. Find loops over database results
grep -B 5 -A 10 "for.*range" <pr_files> | grep -E "\.Where|\.Find|\.First"

# 2. Check if query uses loop variable
# for _, merchant := range merchants {
#     db.Where("merchant_id = ?", merchant.ID)  // ❌ N+1!
# }

# 3. Flag if query inside loop references loop variable
```

### Flag Conditions

Flag if:
- Database query inside `for ... range` loop
- Query references loop variable (e.g., `merchant.ID`)
- No `Preload()` or `IN (?)` batch query
- Comment doesn't explain batching strategy

### Severity

🚨 **Critical** - Performance catastrophe:
- Linear growth with data size (O(n))
- API timeouts
- Database connection exhaustion
- Production incidents

---

## Check 5: ORDER BY Columns Have Indexes ⚠️ HIGH

### What to Check

Columns used in ORDER BY must have indexes to avoid sorting large datasets.

### Bad Pattern ❌

```go
// No index on created_at
SELECT * FROM payments
WHERE status = 'completed'
ORDER BY created_at DESC  -- ❌ Filesort! Sorts in memory
LIMIT 100;

-- EXPLAIN shows: Using filesort (BAD)
```

### Good Pattern ✅

```go
// Composite index covers WHERE + ORDER BY
CREATE INDEX idx_payments_status_created
ON payments(status, created_at DESC);

SELECT * FROM payments
WHERE status = 'completed'
ORDER BY created_at DESC  -- ✅ Uses index, no sort needed
LIMIT 100;

-- EXPLAIN shows: Using index (GOOD)
```

### Detection Strategy

```bash
# Find ORDER BY clauses
grep -rn "ORDER BY\|\.Order(" <pr_files>

# Extract columns
# ORDER BY created_at DESC → Column: created_at

# Check for index
grep "CREATE INDEX.*created_at" <migration_files>
```

### Severity

⚠️ **High** - Performance issues:
- In-memory sorts on large tables
- High memory usage
- Slow pagination queries

---

## Check 6: Unique Constraints as Indexes 📋 MEDIUM

### What to Check

Unique constraints can serve dual purpose as indexes for queries.

### Good Pattern ✅

```go
// Unique constraint on email (also serves as index)
CREATE UNIQUE INDEX idx_users_email ON users(email);

// Query benefits from unique index
SELECT * FROM users WHERE email = 'user@example.com';  -- ✅ Unique index scan
```

### Severity

📋 **Medium** - Optimization opportunity

---

## Summary Table

| Check # | Pattern | Severity | Impact |
|---------|---------|----------|--------|
| 1 | Foreign key columns indexed | 🚨 Critical | JOIN performance |
| 2 | WHERE clause columns indexed | ⚠️ High | Query performance |
| 3 | Composite indexes match queries | ⚠️ High | Multi-column queries |
| 4 | N+1 query detection | 🚨 Critical | Batch vs loop queries |
| 5 | ORDER BY columns indexed | ⚠️ High | Sorting performance |
| 6 | Unique constraints as indexes | 📋 Medium | Optimization |

---

## Detection Workflow

### Step 1: Scan Code for Query Patterns

```bash
# Find all database queries
grep -rn "\.Where\|\.Find\|\.First\|SELECT" <pr_files> --include="*.go" --include="*.sql"

# Extract patterns:
# - WHERE columns
# - JOIN columns
# - ORDER BY columns
# - Queries in loops (N+1)
```

### Step 2: Scan Migration Files for Indexes

```bash
# Find all index definitions
grep -rn "CREATE INDEX\|index:" <migration_files>

# Build index map:
# {
#   "payments.merchant_id": ["idx_payments_merchant"],
#   "payments.status": ["idx_payments_status"],
#   "payments.merchant_id,status": ["idx_payments_merchant_status"]
# }
```

### Step 3: Match Queries to Indexes

```python
# Pseudo-code
for query_pattern in queries:
    columns = extract_where_columns(query_pattern)

    # Check single column indexes
    for col in columns:
        if col not in index_map:
            flag_issue(f"Column {col} queried but no index exists")

    # Check composite indexes for multi-column queries
    if len(columns) > 1:
        composite_key = ",".join(columns)
        if composite_key not in index_map:
            flag_issue(f"Multi-column query but no composite index: {composite_key}")
```

### Step 4: Generate Report

```
📁 File: internal/services/payment_service.go

🚨 Check #1 Failed: Foreign key without index (Line 12)
   Migration: migrations/add_gateway_credentials.go
   Column: gateway_id (foreign key to gateways)
   Fix: CREATE INDEX idx_gateway_credentials_gateway_id ON gateway_credentials(gateway_id)

⚠️  Check #2 Failed: WHERE column not indexed (Line 45)
   Code: db.Where("status = ?", status).Find(&payments)
   Column: status
   Fix: Add index in migration: CREATE INDEX idx_payments_status ON payments(status)

🚨 Check #4 Failed: N+1 query detected (Line 67)
   Code: for _, merchant := range merchants {
             db.Where("merchant_id = ?", merchant.ID).Find(&gateways)
         }
   Issue: 1 + N queries (N = number of merchants)
   Fix: Use batch query:
         var gateways []Gateway
         db.Where("merchant_id IN ?", merchantIDs).Find(&gateways)

⚠️  Check #3 Failed: No composite index for query pattern (Line 89)
   Code: WHERE merchant_id = ? AND status = ?
   Columns: merchant_id, status
   Fix: CREATE INDEX idx_payments_merchant_status ON payments(merchant_id, status)

✅ Check #5 Passed: ORDER BY columns indexed
✅ Check #6 Passed: Unique constraints properly used
```

---

## Limitations & Dynamic Analysis

### Static Analysis Cannot Detect

1. **Actual index usage** - Requires EXPLAIN analysis
2. **Query plan costs** - Needs live database statistics
3. **Index selectivity** - Depends on data distribution
4. **Unused indexes** - Needs production stats
5. **Performance benchmarks** - Requires test data

### For Dynamic Analysis, Use Separate Skill

**Postgres Query Plan Analyzer** (similar to trino-analyzer):
- Connects to dev/stage database
- Runs EXPLAIN ANALYZE
- Estimates query costs
- Suggests optimizations
- Benchmarks query plans

```bash
# Example usage (future skill)
claude postgres-query-analyzer \
  --query "SELECT * FROM payments WHERE status = 'pending'" \
  --database stage \
  --suggest-indexes
```

---

## Integration with Pre-Mortem

### When to Load

Load `infrastructure-query-index-analysis.md` when PR modifies:
- Database queries (`.Where`, `.Find`, `SELECT`)
- Migration files (`CREATE TABLE`, `CREATE INDEX`)
- Repository files (`repo.go`, `*_repository.go`)

### Priority

Run after basic database checks but before domain checks:
1. Database transaction checks
2. **Query & index analysis** ← This
3. Domain constraints validation

---

## Best Practices

1. **Index foreign keys ALWAYS**
2. **Analyze query patterns before creating indexes**
3. **Use composite indexes for multi-column WHERE clauses**
4. **Avoid N+1 queries - use batch fetching**
5. **Test migrations on realistic data volumes**
6. **Monitor index usage in production**

---

## Example: Complete Index Strategy

```sql
-- Table: gateway_credentials
CREATE TABLE gateway_credentials (
    id VARCHAR(14) PRIMARY KEY,
    gateway_id VARCHAR(14) NOT NULL,
    merchant_id VARCHAR(14) NOT NULL,
    acquirer VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Foreign keys + indexes
CREATE INDEX idx_gc_gateway_id ON gateway_credentials(gateway_id);
CREATE INDEX idx_gc_merchant_id ON gateway_credentials(merchant_id);

ALTER TABLE gateway_credentials
ADD CONSTRAINT fk_gateway FOREIGN KEY (gateway_id) REFERENCES gateways(id);

-- Composite indexes for common query patterns
-- Query: WHERE merchant_id = ? AND status = ?
CREATE INDEX idx_gc_merchant_status ON gateway_credentials(merchant_id, status);

-- Query: WHERE gateway_id = ? AND acquirer = ?
CREATE INDEX idx_gc_gateway_acquirer ON gateway_credentials(gateway_id, acquirer);

-- Query: WHERE status = ? ORDER BY created_at DESC
CREATE INDEX idx_gc_status_created ON gateway_credentials(status, created_at DESC);

-- Unique constraint (also serves as index)
CREATE UNIQUE INDEX idx_gc_unique ON gateway_credentials(gateway_id, merchant_id, acquirer);
```

---

## References

**Razorpay patterns:**
- Foreign key indexes in payment tables
- Composite indexes for merchant + status queries
- N+1 prevention in API responses
- Order by created_at in pagination
