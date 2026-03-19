# Database Infrastructure Checks

## Overview

Validates database usage patterns to prevent data corruption, race conditions, performance issues, and connection leaks.

**Load when:** PR modifies `internal/*/repo.go`, `pkg/db/*`, `internal/daos/*`, or database-related code

**Total Checks:** 6

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 2
- 📋 Medium: 2

---

## Check 1: Transaction Rollback Patterns 🚨 CRITICAL

### What to Check

Transaction errors must trigger rollback to prevent partial commits and data corruption.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Error doesn't trigger rollback
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()  // Only rolls back on panic!
    }
}()

err := tx.Save(&model).Error
if err != nil {
    return err  // ❌ Returns without rollback - transaction stays open!
}

err = tx.Model(existingModel).Save().Error
if err != nil {
    return err  // ❌ Again, no rollback
}

tx.Commit()
```

**Problem:**
- Rollback only happens on panic, not on business logic errors
- Error paths leak transactions
- Partial data committed if second Save() fails

### Good Pattern ✅

```go
// CORRECT: Defer rollback, let commit override
tx := db.Begin()
defer tx.Rollback()  // ✅ Always rollback unless commit succeeds

err := tx.Save(&model).Error
if err != nil {
    return err  // Rollback happens via defer
}

err = tx.Model(existingModel).Save().Error
if err != nil {
    return err  // Rollback happens via defer
}

return tx.Commit().Error  // Overrides the deferred rollback
```

**Why this works:**
- `defer tx.Rollback()` executes after all returns
- If `Commit()` succeeds, `Rollback()` is a no-op
- If error before `Commit()`, rollback happens automatically

### Detection Strategy

Search PR diff for:
```bash
# Find transaction starts
grep -n "db.Begin()" <pr_files>
grep -n "\.Begin()" <pr_files>

# For each Begin(), verify:
# 1. defer tx.Rollback() exists within 5 lines
# 2. No conditional rollback (only in defer)
# 3. Error paths don't skip rollback
```

### Flag Conditions

Flag if:
- `db.Begin()` or `tx.Begin()` found
- No `defer tx.Rollback()` within next 10 lines
- Rollback inside conditional: `if err != nil { tx.Rollback() }`
- Loop with DB operations not checking errors individually

### Severity

🚨 **Critical** - Can cause data corruption, partial writes, orphaned records

### Reference

Based on terminals analysis: `/internal/terminals/repo.go:422`

---

## Check 2: Replica Lag Implementation 🚨 CRITICAL

### What to Check

Read replica lag must be actually checked, not stubbed out with hardcoded value.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Unimplemented replica lag check
func replicaLag(replica *gorm.DB) int {
    return 0  // ❌ Always returns 0 - no actual check!
}

func (db *Db) ReadOnlyInstance() *gorm.DB {
    for _, replica := range db.replicas {
        if replicaLag(replica) > MaxAllowedReplicaLag {  // Never true!
            continue
        }
        return replica  // Always uses replica, even if lagging
    }
    return db.master
}
```

**Problem:**
- Replicas with 10+ second lag serve stale data
- Read-after-write inconsistency
- Users see outdated information

### Good Pattern ✅

```go
// CORRECT: Actually query replica lag
func replicaLag(replica *gorm.DB) int {
    var lag int
    // PostgreSQL
    err := replica.Raw(`
        SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))::int
    `).Scan(&lag).Error

    if err != nil {
        logger.Warn("Failed to check replica lag", "error", err)
        return MaxAllowedReplicaLag + 1  // Treat as lagging
    }
    return lag
}

// MySQL alternative
func replicaLagMySQL(replica *gorm.DB) int {
    var lag int
    replica.Raw("SHOW SLAVE STATUS").Scan(&result)
    return result.SecondsBehindMaster
}
```

### Detection Strategy

```bash
# Find replica lag functions
grep -A 5 "func.*replicaLag" <pr_files>

# Check for:
# 1. Hardcoded return values (return 0, return lag, etc.)
# 2. No database query (Raw, Exec)
# 3. No actual lag calculation
```

### Flag Conditions

Flag if:
- Function named `replicaLag` or similar
- Function body returns constant (0, hardcoded number)
- No database query to check actual lag
- Comment like "TODO" or "Unimplemented"

### Severity

🚨 **Critical** - Serves stale data, read-after-write inconsistency

### Reference

Based on terminals analysis: `/pkg/db/db.go:257`

---

## Check 3: Query Timeouts ⚠️ HIGH

### What to Check

Long-running database queries must have timeout to prevent connection pool exhaustion.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No timeout on query
query := db.Model(&TerminalModel{}).Where("org_id = ?", orgId).Find(&terminals)
// Query can run indefinitely!

// Also bad: Context without deadline
func FetchTerminals(ctx *gin.Context) {
    // ctx has no timeout
    db.WithContext(ctx).Find(&terminals)
}
```

**Problem:**
- Slow queries block connections
- Connection pool exhaustion
- Cascading failures

### Good Pattern ✅

```go
// CORRECT: Query with timeout
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel()

query := db.WithContext(ctx).Model(&TerminalModel{}).
    Where("org_id = ?", orgId).
    Find(&terminals)

if errors.Is(query.Error, context.DeadlineExceeded) {
    return ErrQueryTimeout
}
```

### Detection Strategy

```bash
# Find database queries
grep -n "\.Find(" <pr_files>
grep -n "\.First(" <pr_files>
grep -n "\.Scan(" <pr_files>

# For each query, check:
# 1. Is .WithContext() used?
# 2. Does context have timeout? (WithTimeout, WithDeadline)
```

### Flag Conditions

Flag if:
- Database query without `.WithContext()`
- Context passed but no timeout visible in function
- Long-running query (joins, aggregations) without timeout

### Severity

⚠️ **High** - Connection pool exhaustion, cascading failures

---

## Check 4: Migration Safety 📋 MEDIUM

### What to Check

Database migrations must be idempotent and handle partial failures.

### Bad Pattern ❌

```sql
-- ANTI-PATTERN: Fails if index exists
CREATE INDEX terminals_merchant_id_idx ON terminals (merchant_id);

-- ANTI-PATTERN: Multiple operations, if first fails, rest skipped
CREATE INDEX idx1 ON table1 (col1);
-- If above fails, below doesn't run
CREATE INDEX idx2 ON table2 (col2);
```

```go
// ANTI-PATTERN: Large INSERT as single statement
_, err = tx.Exec(`INSERT INTO merchant_instrument (instrument, tat) VALUES ` +
    `('pg.cards.domestic.visa', 25), ` +
    `('pg.cards.domestic.mastercard', 25), ` +
    // ... 1000+ lines of values
)
```

**Problem:**
- Re-running migration fails
- Partial index creation
- Large INSERT timeout on production data

### Good Pattern ✅

```sql
-- CORRECT: Idempotent index creation
CREATE INDEX IF NOT EXISTS terminals_merchant_id_idx
ON terminals (merchant_id);

-- CORRECT: Each operation independent
CREATE INDEX IF NOT EXISTS idx1 ON table1 (col1);
CREATE INDEX IF NOT EXISTS idx2 ON table2 (col2);
```

```go
// CORRECT: Batched INSERT
const batchSize = 100
for i := 0; i < len(values); i += batchSize {
    batch := values[i:min(i+batchSize, len(values))]
    _, err := tx.Exec(buildInsertQuery(batch))
    if err != nil {
        return err
    }
}
```

### Detection Strategy

```bash
# Find migration files
find internal/migration -name "*.go" -o -name "*.sql"

# Check for:
grep "CREATE INDEX" <migration_files>  # Without IF NOT EXISTS
grep "INSERT INTO.*VALUES.*,.*,.*," <migration_files>  # Large multi-row INSERT
```

### Flag Conditions

Flag if:
- `CREATE INDEX` without `IF NOT EXISTS`
- `INSERT INTO ... VALUES` with 50+ rows in single statement
- No Down migration (rollback function empty)
- Data migration without batching

### Severity

📋 **Medium** - Failed deploys, manual intervention needed

---

## Check 5: Missing Indexes on New Columns 📋 MEDIUM

### What to Check

New columns used in WHERE/JOIN must have indexes to prevent table scans.

### Bad Pattern ❌

```go
// New query added in PR
func FindByGateway(gateway string) {
    db.Where("gateway = ?", gateway).Find(&terminals)
    // ❌ gateway column not indexed!
}

// New foreign key
type TerminalModel struct {
    OrgId string `gorm:"column:org_id"`  // ❌ No index
}

// Used in query
db.Where("org_id = ?", orgId).Find(&terminals)
```

**Problem:**
- Full table scan on every query
- Slow queries as table grows
- High CPU/IO load

### Good Pattern ✅

```sql
-- Migration adds index with column
ALTER TABLE terminals ADD COLUMN gateway VARCHAR(255);
CREATE INDEX terminals_gateway_idx ON terminals (gateway);

-- Or composite index for common query pattern
CREATE INDEX terminals_gateway_org_idx ON terminals (gateway, org_id);
```

### Detection Strategy

```bash
# Find new WHERE clauses in PR
git diff main..HEAD | grep -A 2 "\.Where("

# For each column in WHERE:
# 1. Check if it's a new column (added in migration)
# 2. Search for CREATE INDEX on that column
# 3. Flag if column exists but no index
```

### Flag Conditions

Flag if:
- New WHERE clause on column
- Column not in any index
- Foreign key column without index
- Frequently queried column (in multiple files) without index

### Severity

📋 **Medium** - Performance degradation as data grows

---

## Check 6: N+1 Query Prevention ℹ️ LOW

### What to Check

Database queries inside loops should be batch queries to avoid N+1 problem.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Query in loop (N+1)
for _, terminal := range terminals {
    submerchants, _ := repo.FetchSubmerchants(terminal.ID)
    // ❌ Executes one query per terminal!
    terminal.Submerchants = submerchants
}
```

**Problem:**
- For 100 terminals = 101 queries (1 for terminals + 100 for submerchants)
- Slow API response
- High database load

### Good Pattern ✅

```go
// CORRECT: Batch query with IN clause
terminalIds := make([]string, len(terminals))
for i, t := range terminals {
    terminalIds[i] = t.ID
}

// Single query for all submerchants
submerchantsByTerminalId := repo.FetchSubmerchantsByTerminalIds(terminalIds)

// Map results
for i, terminal := range terminals {
    terminals[i].Submerchants = submerchantsByTerminalId[terminal.ID]
}
```

### Detection Strategy

```bash
# Find loops with database operations
git diff main..HEAD | grep -B 5 -A 10 "for.*range"

# Look for patterns:
# - Find(), First(), Where() inside for loop
# - Repository method calls in loop
```

### Flag Conditions

Flag if:
- Database query method inside `for` loop
- Not using batch query pattern (WHERE IN)
- Visible in hot path (API handlers, workers)

### Severity

ℹ️ **Low** - Performance issue, but query cache may mitigate

---

## Summary Table

| Check # | Pattern | Severity | Common Location |
|---------|---------|----------|-----------------|
| 1 | Transaction Rollback | 🚨 Critical | `internal/*/repo.go` |
| 2 | Replica Lag | 🚨 Critical | `pkg/db/db.go` |
| 3 | Query Timeouts | ⚠️ High | `internal/*/repo.go` |
| 4 | Migration Safety | 📋 Medium | `internal/migration/*.go` |
| 5 | Missing Indexes | 📋 Medium | Migrations + queries |
| 6 | N+1 Queries | ℹ️ Low | `internal/services/*.go` |

---

## How to Apply

**For each file matching** `internal/*/repo.go`, `pkg/db/*`:

1. Parse PR diff to extract code changes
2. For each check (1-6):
   - Search for the anti-pattern
   - Compare against good/bad examples
   - Flag violations with severity + line number
3. Collect all findings
4. Report to user with file references

**Example output:**

```
📁 File: internal/terminals/repo.go

🚨 Check #1 Failed: Transaction rollback missing (Line 422)
   Code: err := tx.Save(&model).Error; if err != nil { return err }
   Issue: Rollback not triggered on Save() error
   Fix: Add "defer tx.Rollback()" after db.Begin()

⚠️  Check #3 Failed: Query without timeout (Line 156)
   Code: db.Where("org_id = ?", orgId).Find(&terminals)
   Issue: No context timeout on potentially slow query
   Fix: Use db.WithContext(ctx) with 5-second timeout

✅ Check #2 Passed: Replica lag properly implemented
✅ Check #4 Passed: Migrations have IF NOT EXISTS
✅ Check #5 Passed: All WHERE columns indexed
✅ Check #6 Passed: No N+1 queries detected
```
