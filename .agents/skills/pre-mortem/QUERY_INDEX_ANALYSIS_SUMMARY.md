# Query & Index Analysis Integration - Complete ✅

## Summary

Successfully added comprehensive **static query and index analysis** to pre-mortem skill. This validates that database queries have appropriate indexes by analyzing code patterns and migration files - **no live database connection required**.

---

## Two-Part Approach

### ✅ Part 1: STATIC Analysis (Pre-Mortem - DONE)

**What it does:**
- Scans code for query patterns (WHERE, JOIN, ORDER BY)
- Checks migration files for index definitions
- Validates foreign keys have indexes
- Detects N+1 query patterns
- Checks composite indexes match query patterns

**When:** During PR review (no DB needed)
**File:** `references/infrastructure-query-index-analysis.md`
**Checks:** 6 (2 Critical, 3 High, 1 Medium)

### ⚠️ Part 2: DYNAMIC Analysis (Separate Skill - RECOMMENDED)

**What it needs:**
- Live database connection (dev/stage)
- Execute EXPLAIN/EXPLAIN ANALYZE
- Query plan cost estimation
- Index usage statistics
- Performance benchmarking

**When:** Post-merge or on-demand
**Similar to:** `trino-analyzer` skill (already exists for Trino)
**Recommendation:** Create `postgres-query-analyzer` and `mysql-query-analyzer` skills

---

## What Was Added (Part 1)

### New Check File: Query & Index Analysis (6 Checks)

**File:** `infrastructure-query-index-analysis.md`

| Check # | Pattern | Severity | What It Detects |
|---------|---------|----------|-----------------|
| 1 | Foreign Key Indexes | 🚨 Critical | Foreign key columns without indexes → table scans on JOIN |
| 2 | WHERE Clause Indexes | ⚠️ High | Frequently queried columns without indexes |
| 3 | Composite Indexes | ⚠️ High | Multi-column WHERE clauses need composite indexes |
| 4 | N+1 Query Detection | 🚨 Critical | Queries inside loops (1 + N queries) |
| 5 | ORDER BY Indexes | ⚠️ High | Sort columns without indexes → filesort |
| 6 | Unique Constraints | 📋 Medium | Unique constraints serve as indexes |

---

## Check 1: Foreign Key Indexes 🚨 CRITICAL

### Problem

Foreign keys without indexes cause severe performance issues:
- Full table scans on every JOIN
- DELETE CASCADE locks entire table
- Constraint checks are slow

### Detection

```bash
# 1. Find foreign key definitions
grep "FOREIGN KEY\|foreignKey" migrations/*.sql

# 2. Extract FK columns
# FOREIGN KEY (gateway_id) REFERENCES gateways(id) → gateway_id

# 3. Check if index exists
grep "CREATE INDEX.*gateway_id" migrations/*.sql

# 4. Flag if no index
```

### Example

```go
// ❌ BAD: Foreign key without index
ALTER TABLE gateway_credentials
ADD CONSTRAINT fk_gateway
FOREIGN KEY (gateway_id) REFERENCES gateways(id);

SELECT * FROM gateway_credentials gc
JOIN gateways g ON gc.gateway_id = g.id;  -- ❌ Table scan!

// ✅ GOOD: Index created first
CREATE INDEX idx_gateway_credentials_gateway_id
ON gateway_credentials(gateway_id);

ALTER TABLE gateway_credentials
ADD CONSTRAINT fk_gateway
FOREIGN KEY (gateway_id) REFERENCES gateways(id);

-- Query now uses index
```

---

## Check 2: WHERE Clause Indexes ⚠️ HIGH

### Problem

Columns in WHERE clauses without indexes cause full table scans.

### Detection

```bash
# 1. Find WHERE clauses in code
grep "\.Where\|WHERE" *.go

# 2. Extract columns
# .Where("status = ?", ...) → status
# WHERE merchant_id = ? → merchant_id

# 3. Check if indexed
grep "CREATE INDEX.*status" migrations/*.sql
```

### Example

```go
// ❌ BAD: No index on status
db.Where("status = ?", "pending").Find(&payments)
// Scans entire payments table!

// ✅ GOOD: Index exists
CREATE INDEX idx_payments_status ON payments(status);

db.Where("status = ?", "pending").Find(&payments)
// Index scan - fast!
```

---

## Check 3: Composite Indexes ⚠️ HIGH

### Problem

Multi-column WHERE clauses need composite indexes with correct column order.

### Detection

```bash
# Find multi-column WHERE
grep "WHERE.*AND" *.go

# Extract column pairs
# WHERE merchant_id = ? AND status = ? → merchant_id, status

# Check for composite index
grep "CREATE INDEX.*merchant_id.*status" migrations/*.sql
```

### Example

```go
// ❌ BAD: Separate indexes
CREATE INDEX idx_merchant ON payments(merchant_id);
CREATE INDEX idx_status ON payments(status);

SELECT * FROM payments
WHERE merchant_id = ? AND status = ?;
-- Can only use one index efficiently!

// ✅ GOOD: Composite index
CREATE INDEX idx_payments_merchant_status
ON payments(merchant_id, status);

SELECT * FROM payments
WHERE merchant_id = ? AND status = ?;
-- Uses composite index - optimal!
```

**Column Order Matters:**
```sql
-- ✅ Works: WHERE merchant_id = ? AND status = ?
-- ✅ Works: WHERE merchant_id = ?
-- ❌ Doesn't work well: WHERE status = ?
```

---

## Check 4: N+1 Query Detection 🚨 CRITICAL

### Problem

Queries inside loops cause catastrophic performance (1 query + N queries in loop).

### Detection

```bash
# Find loops with queries
grep -B 3 -A 5 "for.*range" *.go | grep "\.Where\|\.Find"

# Flag if query uses loop variable
```

### Example

```go
// ❌ BAD: N+1 queries
var merchants []Merchant
db.Find(&merchants)  // 1 query

for _, merchant := range merchants {
    var gateways []Gateway
    db.Where("merchant_id = ?", merchant.ID).Find(&gateways)  // N queries!
}
// Total: 1 + N queries (100 merchants = 101 queries!)

// ✅ GOOD: Batch fetch (2 queries)
var merchants []Merchant
db.Find(&merchants)  // 1 query

var merchantIDs []string
for _, m := range merchants {
    merchantIDs = append(merchantIDs, m.ID)
}

var gateways []Gateway
db.Where("merchant_id IN ?", merchantIDs).Find(&gateways)  // 1 query!

// Group in memory
gatewayMap := make(map[string][]Gateway)
for _, gw := range gateways {
    gatewayMap[gw.MerchantID] = append(gatewayMap[gw.MerchantID], gw)
}

// ✅ BETTER: GORM Preload (2 queries)
var merchants []Merchant
db.Preload("Gateways").Find(&merchants)  // GORM auto-batches!
```

**Impact:**
- 100 merchants = 101 queries (N+1)
- vs 2 queries (batch)
- **50x performance improvement**

---

## Check 5: ORDER BY Indexes ⚠️ HIGH

### Problem

ORDER BY without index causes filesort (sorting large datasets in memory).

### Example

```sql
-- ❌ BAD: No index on created_at
SELECT * FROM payments
WHERE status = 'completed'
ORDER BY created_at DESC
LIMIT 100;
-- EXPLAIN: Using filesort (BAD)

-- ✅ GOOD: Composite index
CREATE INDEX idx_payments_status_created
ON payments(status, created_at DESC);

SELECT * FROM payments
WHERE status = 'completed'
ORDER BY created_at DESC
LIMIT 100;
-- EXPLAIN: Using index (GOOD)
```

---

## Check 6: Unique Constraints 📋 MEDIUM

### Tip

Unique constraints automatically create indexes - use them for queries.

```sql
CREATE UNIQUE INDEX idx_users_email ON users(email);

SELECT * FROM users WHERE email = ?;
-- Uses unique index automatically!
```

---

## Detection Workflow

### Step 1: Scan Code

```bash
# Find all database queries
grep -rn "\.Where\|\.Find\|SELECT" *.go

# Extract query patterns:
# - WHERE columns
# - JOIN columns
# - ORDER BY columns
# - Queries in loops
```

### Step 2: Scan Migrations

```bash
# Find all indexes
grep -rn "CREATE INDEX\|index:" migrations/

# Build index map:
# {
#   "payments.merchant_id": ["idx_payments_merchant"],
#   "payments.status": ["idx_payments_status"]
# }
```

### Step 3: Match & Flag

```python
for query in queries:
    columns = extract_where_columns(query)

    for col in columns:
        if col not in index_map:
            flag(f"Column {col} queried but no index")
```

---

## Example Output

```
📁 File: internal/services/payment_service.go

🚨 Check #1 Failed: Foreign key without index (Line 12)
   Migration: migrations/add_gateway_credentials.go
   Column: gateway_id (foreign key to gateways)
   Issue: JOIN on unindexed foreign key → table scan
   Fix: CREATE INDEX idx_gateway_credentials_gateway_id
        ON gateway_credentials(gateway_id)

⚠️  Check #2 Failed: WHERE column not indexed (Line 45)
   Code: db.Where("status = ?", status).Find(&payments)
   Column: status
   Issue: Full table scan on WHERE clause
   Fix: CREATE INDEX idx_payments_status ON payments(status)

🚨 Check #4 Failed: N+1 query detected (Line 67-72)
   Code: for _, merchant := range merchants {
             db.Where("merchant_id = ?", merchant.ID).Find(&gateways)
         }
   Issue: 1 + N queries (N = number of merchants)
   Impact: 100 merchants = 101 queries!
   Fix: Batch query:
        merchantIDs := extractIDs(merchants)
        db.Where("merchant_id IN ?", merchantIDs).Find(&gateways)
        // Or use: db.Preload("Gateways").Find(&merchants)

⚠️  Check #3 Failed: No composite index (Line 89)
   Code: WHERE merchant_id = ? AND status = ?
   Columns: merchant_id, status
   Fix: CREATE INDEX idx_payments_merchant_status
        ON payments(merchant_id, status)

⚠️  Check #5 Failed: ORDER BY without index (Line 102)
   Code: ORDER BY created_at DESC
   Fix: CREATE INDEX idx_payments_status_created
        ON payments(status, created_at DESC)

✅ Check #6 Passed: Unique constraints properly used
```

---

## Updated Stats

### Before
- **Total Checks:** 113
- **Infrastructure Checks:** 87

### After
- **Total Checks:** 119 (+6)
- **Infrastructure Checks:** 93 (+6)
- **Query & Index Checks:** 6 (NEW)

---

## Files Updated

1. ✅ `references/infrastructure-query-index-analysis.md` - New file (6 checks)
2. ✅ `SKILL.md` - Updated file mapping and totals
3. ✅ `README.md` - Updated check counts
4. ✅ `QUERY_INDEX_ANALYSIS_SUMMARY.md` - This summary

---

## Limitations (Static Analysis)

### ✅ CAN Detect
- Foreign keys without indexes
- WHERE clauses without indexes
- N+1 query patterns
- Missing composite indexes
- ORDER BY without indexes

### ❌ CANNOT Detect (Needs Live DB)
- Actual index usage (requires EXPLAIN)
- Query plan costs
- Index selectivity
- Unused indexes
- Performance benchmarks

---

## Part 2: Dynamic Analysis (Recommended Future Work)

### Postgres Query Plan Analyzer (Separate Skill)

Similar to existing `trino-analyzer` skill, create:

**Skills to build:**
1. `postgres-query-analyzer`
2. `mysql-query-analyzer`

**Features:**
- Connect to dev/stage database
- Execute EXPLAIN ANALYZE
- Parse query plans
- Estimate costs
- Suggest optimizations
- Benchmark queries

**Example usage:**
```bash
# Analyze query performance (future)
claude postgres-query-analyzer \
  --query "SELECT * FROM payments WHERE status = 'pending'" \
  --database stage \
  --explain-analyze \
  --suggest-indexes

# Output:
# Query Plan:
# -> Index Scan on payments_status (cost=0.42..8.44 rows=100)
#    Index Cond: (status = 'pending')
#
# Estimated cost: 8.44
# Execution time: 0.123ms
#
# ✅ Query uses index efficiently
# 💡 Suggestion: Add composite index for common ORDER BY:
#    CREATE INDEX idx_payments_status_created
#    ON payments(status, created_at DESC);
```

**Implementation approach:**
```python
# Similar to trino-analyzer/scripts/analyze_query_plan.py

import psycopg2
import sqlparse

def analyze_query(query, db_config):
    conn = psycopg2.connect(**db_config)
    cursor = conn.cursor()

    # Execute EXPLAIN ANALYZE
    cursor.execute(f"EXPLAIN ANALYZE {query}")
    plan = cursor.fetchall()

    # Parse plan
    cost = extract_cost(plan)
    indexes_used = extract_indexes(plan)

    # Suggest optimizations
    suggestions = analyze_plan(plan)

    return {
        'cost': cost,
        'indexes_used': indexes_used,
        'suggestions': suggestions
    }
```

**When to use:**
- After PR is merged (dev database available)
- Performance investigation
- Pre-production optimization
- Load testing preparation

---

## Best Practices

### Static Checks (Pre-Mortem)
1. **Always index foreign keys**
2. **Analyze query patterns before creating indexes**
3. **Use composite indexes for multi-column WHERE**
4. **Avoid N+1 queries - batch fetch**

### Dynamic Checks (Separate Skill)
5. **Run EXPLAIN on new queries**
6. **Test with realistic data volumes**
7. **Monitor index usage in production**
8. **Remove unused indexes**

---

## Integration Workflow

### During PR Review (Pre-Mortem)
```
1. Developer creates PR with new queries
2. Pre-mortem runs static analysis
3. Flags missing indexes, N+1 patterns
4. Developer adds indexes to migration
5. PR approved and merged
```

### After Merge (Query Analyzer)
```
1. PR merged to dev branch
2. Run postgres-query-analyzer on dev database
3. Execute EXPLAIN ANALYZE
4. Validate index usage
5. Fine-tune indexes if needed
6. Promote to production
```

---

## Success Criteria Met

1. ✅ Foreign key index validation
2. ✅ WHERE clause index checking
3. ✅ Composite index pattern matching
4. ✅ N+1 query detection
5. ✅ ORDER BY index validation
6. ✅ Static analysis (no DB needed)
7. ✅ Clear separation: static vs dynamic
8. ✅ Example outputs provided
9. ✅ Documentation complete

---

## Next Steps

### Immediate
1. ✅ Integration complete
2. 📝 Test on real PRs with database queries
3. 📝 Validate detection accuracy

### Future (Recommended)
1. 📝 Create `postgres-query-analyzer` skill (like trino-analyzer)
2. 📝 Create `mysql-query-analyzer` skill
3. 📝 Integrate with CI/CD for automated EXPLAIN
4. 📝 Build index recommendation engine

---

## References

**Existing patterns:**
- `trino-analyzer` - Query plan analysis for Trino
- `infrastructure-database.md` - Check #5 (basic index check)
- `infrastructure-postgres.md` - Check #3 (JSONB indexes)

**New patterns:**
- Foreign key indexing across all Razorpay services
- N+1 query prevention in API responses
- Composite index strategies for common queries

---

**Pre-mortem now has comprehensive static query & index analysis!** 🚀

For **dynamic query plan analysis**, create separate skill with database connectivity.
