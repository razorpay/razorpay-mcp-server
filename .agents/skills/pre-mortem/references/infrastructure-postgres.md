# PostgreSQL Infrastructure Checks

## Overview

Validates PostgreSQL usage patterns in Razorpay services to prevent data corruption, race conditions, deadlocks, and performance issues. Based on patterns from x-payroll services, wallet, terminals, and other Go services using GORM.

**Load when:** PR modifies code using PostgreSQL/GORM

**Total Checks:** 6

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 3
- 📋 Medium: 1

---

## Check 1: Row-Level Locking (SELECT FOR UPDATE) 🚨 CRITICAL

### What to Check

Critical read-modify-write operations must use row-level locking to prevent race conditions.

### Razorpay Services Using This

**Found in production code:**
- `wallet:internal/hold/repo.go` - FindAndLockHoldsByParams
- `payments-card-present:internal/service/payment.go` - Select for update payment
- `mco:internal/onboarding/repo/onboarding.go` - GetByMerchantIDAndProductForUpdate
- `loyalty-booking-engine:internal/booking/repo/repository.go` - SELECT FOR UPDATE
- `rewards-catalogue:internal/brand/repo/repo.go` - Lock row to prevent concurrent modifications

### Bad Pattern ❌

```go
// ANTI-PATTERN: Race condition in balance update
// Found in: Multiple payment services (anti-pattern to avoid)
func DeductBalance(ctx context.Context, merchantID string, amount int) error {
    tx := db.Begin()
    defer tx.Rollback()

    var merchant Merchant
    // ❌ No lock! Another transaction can read same balance
    tx.Where("merchant_id = ?", merchantID).First(&merchant)

    if merchant.Balance < amount {
        return errors.New("insufficient balance")
    }

    merchant.Balance -= amount
    tx.Save(&merchant)

    return tx.Commit().Error
}

// Race scenario:
// T1: Reads balance=1000, checks OK, deducts 600 → balance=400
// T2: Reads balance=1000 (same!), checks OK, deducts 600 → balance=400
// ❌ Should be -200, lost $600!
```

### Good Pattern ✅

```go
// CORRECT: Row-level lock prevents race condition
// Pattern from wallet, payments-card-present services
func DeductBalance(ctx context.Context, merchantID string, amount int) error {
    tx := db.Begin()
    defer tx.Rollback()

    var merchant Merchant
    // ✅ SELECT FOR UPDATE locks the row
    err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("merchant_id = ?", merchantID).
        First(&merchant).Error

    if err != nil {
        return err
    }

    // Row is locked - other transactions wait here
    if merchant.Balance < amount {
        return errors.New("insufficient balance")
    }

    merchant.Balance -= amount
    tx.Save(&merchant)

    return tx.Commit().Error
    // Lock released after commit
}

// Alternative: Raw SQL (loyalty-booking-engine pattern)
func LockBooking(ctx context.Context, bookingID string) (*Booking, error) {
    var booking Booking
    query := `SELECT * FROM bookings WHERE booking_id = $1 FOR UPDATE`
    err := db.Raw(query, bookingID).Scan(&booking).Error
    return &booking, err
}

// ✅ FOR SHARE for non-modifying reads (multiple readers, block writers)
func VerifyBalanceConsistency(ctx context.Context, merchantID string) error {
    tx := db.Begin()
    defer tx.Rollback()

    var merchant Merchant
    // ✅ FOR SHARE: multiple readers OK, writers blocked
    tx.Clauses(clause.Locking{Strength: "SHARE"}).
        Where("merchant_id = ?", merchantID).
        First(&merchant)

    // Read related data with same snapshot...

    return tx.Commit().Error
}
```

### Detection Strategy

```bash
# Find transactions with balance/amount updates
grep -n "\.Save\|\.Update" <pr_files> | grep -i "balance\|amount\|quantity"

# Check if SELECT FOR UPDATE used
grep -B 10 "\.Save\|\.Update" <pr_files> | grep -i "for update\|Locking"

# Flag if critical update without lock
```

### Flag Conditions

Flag if:
- Transaction modifies `balance`, `amount`, `quantity`, `stock` fields
- No `Clauses(clause.Locking{...})` or `FOR UPDATE` before modification
- Comment mentions "// select for update" but no actual lock
- Financial operations (payment, refund, transfer) without lock

### Severity

🚨 **Critical** - Race conditions in financial operations can cause:
- Double spending
- Incorrect balances
- Lost revenue
- Data corruption

### References

**Razorpay production code:**
- `wallet:internal/hold/repo.go:FindAndLockHoldsByParams()`
- `mco:internal/onboarding/repo/onboarding.go:GetByMerchantIDAndProductForUpdate()`
- `rewards-catalogue:internal/brand/repo/repo.go` - Explicit locking comment

---

## Check 2: Deadlock Prevention & Handling ⚠️ HIGH

### What to Check

Transactions must prevent deadlocks through consistent lock ordering and handle deadlock errors with retries.

### Razorpay Patterns Found

**Governor service** (razorpay/governor):
- Metric: `db_deadlock_total` - Tracks deadlocks across services
- Check: `strings.Contains(err.Error(), "Deadlock")`
- Action: Increment counter for monitoring

**Digital billing service** (razorpay/digital-billing-service):
- Function: `checkAndSleepOnDeadlockError()`
- Pattern: Detect deadlock, log warning, retry with backoff

**Parity service** (razorpay/parity):
- Function: `isDeadlock()` - Checks for MySQL error 1213
- Handles both PostgreSQL and MySQL deadlock errors

**Wallet service** (razorpay/wallet):
- Comment: "Debit source - always do pool txn first (debit/credit) to avoid deadlocks"
- Pattern: Consistent ordering (pool first, then container)

### Bad Pattern ❌

```go
// ANTI-PATTERN: Inconsistent lock ordering → deadlock
// Transaction 1:
tx1.Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("id = ?", merchantA).First(&m1)  // Locks A
tx1.Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("id = ?", merchantB).First(&m2)  // Waits for B

// Transaction 2 (concurrent):
tx2.Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("id = ?", merchantB).First(&m2)  // Locks B
tx2.Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("id = ?", merchantA).First(&m1)  // Waits for A
// ❌ Deadlock! T1 has A waiting for B, T2 has B waiting for A

// ANTI-PATTERN: No deadlock retry
err := TransferFunds(fromID, toID, amount)
if err != nil {
    return err  // ❌ Fails immediately on deadlock
}
```

### Good Pattern ✅

```go
// CORRECT: Consistent lock ordering (wallet pattern)
// Pattern from razorpay/wallet: "always do pool txn first"
func TransferFunds(fromID, toID string, amount int) error {
    // ✅ Always lock in alphabetical order
    firstID, secondID := fromID, toID
    if fromID > toID {
        firstID, secondID = toID, fromID
    }

    tx := db.Begin()
    defer tx.Rollback()

    var first, second Account
    tx.Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("id = ?", firstID).First(&first)
    tx.Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("id = ?", secondID).First(&second)

    // Now identify which is from/to
    from, to := &first, &second
    if firstID != fromID {
        from, to = &second, &first
    }

    from.Balance -= amount
    to.Balance += amount

    tx.Save(from)
    tx.Save(to)

    return tx.Commit().Error
}

// CORRECT: Deadlock detection and retry (digital-billing-service pattern)
func TransferWithRetry(ctx context.Context, fromID, toID string, amount int) error {
    maxRetries := 3

    for i := 0; i < maxRetries; i++ {
        err := TransferFunds(fromID, toID, amount)

        if err == nil {
            return nil
        }

        // ✅ Check for deadlock (Razorpay pattern)
        if isDeadlock(err) {
            logger.Warn(ctx, "deadlock_retry",
                "attempt", i+1,
                "from", fromID,
                "to", toID)

            // Exponential backoff with jitter
            backoff := time.Duration(rand.Intn(100)) * time.Millisecond
            time.Sleep(backoff)
            continue
        }

        // Non-deadlock error - fail fast
        return err
    }

    return errors.New("max retries exceeded due to deadlocks")
}

// CORRECT: Deadlock detection (parity + governor pattern)
func isDeadlock(err error) bool {
    if err == nil {
        return false
    }

    errStr := strings.ToLower(err.Error())

    // PostgreSQL deadlock
    if strings.Contains(errStr, "deadlock detected") {
        return true
    }

    // MySQL deadlock (Error 1213)
    if strings.Contains(errStr, "error 1213") ||
       strings.Contains(errStr, "deadlock") {
        return true
    }

    return false
}

// CORRECT: Emit metrics (governor pattern)
func handleDeadlock(tableName string) {
    metric.DBDeadlockTotal.WithLabelValues(tableName).Inc()
}
```

### Detection Strategy

```bash
# Find multi-lock transactions
grep -A 20 "\.Begin()" <pr_files> | grep "Locking.*UPDATE"

# Check for consistent ordering
# Flag if multiple locks without alphabetical/ID ordering

# Check for deadlock handling
grep -n "isDeadlock\|deadlock\|Deadlock" <pr_files>
```

### Flag Conditions

Flag if:
- Multiple `FOR UPDATE` locks in same transaction
- No visible lock ordering (not alphabetical/ID sorted)
- Financial transfer without lock ordering
- No deadlock retry logic for critical operations
- Comment mentions "// avoid deadlocks" but no ordering visible

### Severity

⚠️ **High** - Deadlocks cause:
- Transaction failures
- User experience degradation
- Retry storms
- Service instability

### References

**Razorpay production code:**
- `governor:app/metric/metric.go` - `db_deadlock_total` metric
- `digital-billing-service:internal/migration_job/repo.go:checkAndSleepOnDeadlockError()`
- `parity:backend/internal/store/store.go:isDeadlock()`
- `wallet:instruments/container/recharge/process.go` - Lock ordering comment

---

## Check 3: JSONB Index Usage 📋 MEDIUM

### What to Check

JSONB columns must have appropriate indexes (GIN) for query performance.

### Razorpay Services Using JSONB

**Heavy JSONB usage found in:**
- `mozart:app/models/audit_log.go` - RawRequest, RawResponse, RequestBody, ResponseBody
- `terminals:internal/tidb/types.go` - Product, Card, Emi fields (migration to postgres)
- `splitz:internal/experiment/rule.go` - Experiment rules
- `upi-switch:internal/delegate/service/core/model/delegate.go` - SecondaryPayer, PrimaryPayer
- `business-verification-service:internal/bvs/model/validation.go` - Enrichments, Metadata, Rules
- `offers-engine:internal/marketplace/repo/reward/sql/query_builder.go` - Uses jsonb_agg

### Bad Pattern ❌

```sql
-- ANTI-PATTERN: No index on JSONB column
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    raw_request JSONB,
    raw_response JSONB
);

-- Query is slow (full table scan)
SELECT * FROM audit_logs WHERE raw_request->>'merchant_id' = 'merch_123';
-- Scans all rows!
```

```go
// ANTI-PATTERN: No index for JSONB queries
// Pattern from terminals (before migration optimization)
type Terminal struct {
    ID       string         `gorm:"column:id"`
    Product  postgres.Jsonb `gorm:"type:jsonb;column:product"`  // ❌ No index!
}

// Query without index
db.Where("product->>'name' = ?", "card").Find(&terminals)
// Slow!
```

### Good Pattern ✅

```sql
-- CORRECT: GIN index on entire JSONB column
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    raw_request JSONB,
    raw_response JSONB
);

-- ✅ Index entire JSONB column (general queries)
CREATE INDEX idx_audit_raw_request_gin ON audit_logs USING gin(raw_request);

-- ✅ Index specific JSON path (faster for specific queries)
CREATE INDEX idx_audit_merchant_id ON audit_logs ((raw_request->>'merchant_id'));

-- ✅ GIN index for containment queries
CREATE INDEX idx_audit_request_contains ON audit_logs USING gin(raw_request jsonb_path_ops);

-- Fast queries
SELECT * FROM audit_logs WHERE raw_request->>'merchant_id' = 'merch_123';
SELECT * FROM audit_logs WHERE raw_request @> '{"status": "success"}';
```

```go
// CORRECT: JSONB with proper indexing
type AuditLog struct {
    ID           uint           `gorm:"primaryKey"`
    RawRequest   postgres.Jsonb `gorm:"type:jsonb;column:raw_request"`
    RawResponse  postgres.Jsonb `gorm:"type:jsonb;column:raw_response"`
}

// Migration with GIN index
func AddAuditIndexes(tx *gorm.DB) error {
    // ✅ GIN index on JSONB column
    tx.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_raw_request_gin
              ON audit_logs USING gin(raw_request)`)

    // ✅ Index specific path for common query
    tx.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_merchant_id
              ON audit_logs ((raw_request->>'merchant_id'))`)

    return nil
}

// Query by JSON field (uses index)
func GetAuditsByMerchant(merchantID string) ([]AuditLog, error) {
    var logs []AuditLog

    // ✅ Use JSONB operator (uses GIN index)
    db.Where("raw_request->>'merchant_id' = ?", merchantID).Find(&logs)

    return logs, nil
}

// Containment query (offers-engine pattern)
func GetByFeatures(features []string) ([]Terminal, error) {
    var terminals []Terminal

    // ✅ @> containment operator (uses GIN index with jsonb_path_ops)
    featureJSON := fmt.Sprintf(`{"features": %v}`, features)
    db.Where("product @> ?", featureJSON).Find(&terminals)

    return terminals, nil
}
```

### Detection Strategy

```bash
# Find JSONB column definitions
grep -n "postgres.Jsonb\|type:jsonb" <pr_files>

# Find JSONB queries
grep -n "->>\|@>\|jsonb_agg" <pr_files>

# Check migrations for GIN indexes
grep -n "CREATE INDEX.*gin\|USING gin" <migration_files>
```

### Flag Conditions

Flag if:
- New JSONB column without GIN index in migration
- Query uses `->>`/`@>` operators on JSONB field
- No index on JSONB field being queried
- High-frequency queries (in loops, API handlers) without index

### Severity

📋 **Medium** - Performance degradation as data grows
- Slow queries on large tables
- Full table scans
- High CPU usage

### References

**Razorpay production code:**
- `mozart:app/models/audit_log.go` - Multiple JSONB fields
- `terminals:internal/tidb/types.go` - Comment: "Have to convert below fields to jsonb post migration to postgres"
- `offers-engine:internal/marketplace/repo/reward/sql/query_builder.go` - `jsonb_agg` usage

---

## Check 4: SSL Mode in Production 🚨 CRITICAL

### What to Check

Production database connections must use SSL encryption (`sslmode=require` or `sslmode=verify-full`).

### Razorpay Standard Pattern

**Connection DSN format found across services:**
- `terminals:pkg/db/db.go`: `PostgresConnectionDSNFormat = "host=%s dbname=%s sslmode=%s user=%s password=%s"`
- `x-payroll-compute:pkg/storage/database_config.go`: `PostgresDNS = "host=%s port=%d dbname=%s sslmode=%s user=%s password=%s"`
- `upidp:reporting/pkg/storage/sql/sql.go`: `SslMode: "require"`
- `rproxy:pkg/store/sql/postgres.go`: `PostgresConnectionDSNFormat` includes sslmode

**Dev vs Production:**
- Dev/Test: `sslmode=disable` (local, end-to-end-tests)
- Production: `sslmode=require` (upidp, rproxy, edge-cp, router)

### Bad Pattern ❌

```go
// ANTI-PATTERN: Hardcoded sslmode=disable in production
// Pattern found in dev/test code (should NOT be in production)
dsn := "host=prod-db.razorpay.com dbname=api sslmode=disable user=app password=secret"
db, _ := gorm.Open(postgres.Open(dsn))
// ❌ No encryption! MITM attacks possible

// ANTI-PATTERN: No sslmode specified (defaults to prefer, not enforce)
dsn := "host=prod-db.razorpay.com dbname=api user=app password=secret"
db, _ := gorm.Open(postgres.Open(dsn))
// ❌ May fallback to unencrypted!
```

### Good Pattern ✅

```go
// CORRECT: SSL required in production (Razorpay standard)
// Pattern from terminals, x-payroll-compute, rproxy

// Config structure
type DatabaseConfig struct {
    Host     string `toml:"host"`
    Port     int    `toml:"port"`
    DBName   string `toml:"dbname"`
    SSLMode  string `toml:"sslmode"`  // ✅ Configurable
    User     string `toml:"user"`
    Password string `toml:"password"`
}

// Connection with SSL (terminals pattern)
const PostgresConnectionDSNFormat = "host=%s dbname=%s sslmode=%s user=%s password=%s"

func ConnectDB(cfg DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf(PostgresConnectionDSNFormat,
        cfg.Host,
        cfg.DBName,
        cfg.SSLMode,  // ✅ "require" or "verify-full" in production
        cfg.User,
        cfg.Password,
    )

    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// TOML config - prod-live.toml
[database]
host = "prod-postgres.rds.amazonaws.com"
port = 5432
dbname = "api"
sslmode = "require"  # ✅ Enforced SSL
user = "api_user"
password = "${DB_PASSWORD}"  # From secrets

// TOML config - default.toml (dev)
[database]
host = "localhost"
port = 5432
dbname = "api_dev"
sslmode = "disable"  # ✅ OK for local dev
user = "dev_user"
password = "dev_pass"

// CORRECT: SSL with certificate verification (upidp pattern)
type StorageConfig struct {
    Host     string
    Port     int
    Database string
    Username string
    Password string
    SslMode  string  // ✅ "verify-full" for strongest security
    SslCert  string  // Path to client cert
    SslKey   string  // Path to client key
    SslRootCert string  // Path to CA cert
}

dsn := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=%s user=%s password=%s sslrootcert=%s",
    cfg.Host, cfg.Port, cfg.Database, cfg.SslMode, cfg.Username, cfg.Password, cfg.SslRootCert)
```

### Detection Strategy

```bash
# Find database connection code
grep -n "gorm.Open\|sql.Open" <pr_files>

# Check SSL mode
grep -B 5 -A 5 "sslmode" <pr_files>

# Flag if production config
grep -l "prod\|live" <config_files> | xargs grep "sslmode=disable"
```

### Flag Conditions

Flag if:
- Production config (`prod-live.toml`, `prod.toml`) has `sslmode=disable`
- Connection string without `sslmode` parameter
- Hardcoded DSN with `sslmode=disable` in non-test code
- Environment: production, sslmode not `require` or `verify-full`

### Severity

🚨 **Critical** - Security vulnerability:
- Unencrypted database traffic
- MITM attacks possible
- Credentials exposed on network
- Compliance violations (PCI-DSS, SOC 2)

### References

**Razorpay production code:**
- `terminals:pkg/db/db.go:PostgresConnectionDSNFormat`
- `x-payroll-compute:pkg/storage/database_config.go:PostgresDNS`
- `upidp:reporting/pkg/storage/sql/sql.go` - SslMode: "require"
- `rproxy, edge-cp, router` - All use configurable sslmode

---

## Check 5: Prepared Statements Enabled ⚠️ HIGH

### What to Check

GORM prepared statement mode must be enabled for query plan caching and performance.

### Razorpay Context

While not explicitly found in search results (rate limited), this is a standard GORM best practice that applies to all services using GORM (terminals, x-payroll, upi-switch, etc.).

### Bad Pattern ❌

```go
// ANTI-PATTERN: Prepared statements disabled (default in older GORM)
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
    // ❌ PrepareStmt not set (defaults to false)
})

// Each query parses SQL again
for _, id := range merchantIDs {
    var merchant Merchant
    db.Where("merchant_id = ?", id).First(&merchant)
    // ❌ SQL parsed 100 times for 100 merchants!
}
```

### Good Pattern ✅

```go
// CORRECT: Enable prepared statement caching
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
    PrepareStmt: true,  // ✅ Cache prepared statements

    // Optional: Additional optimization settings
    SkipDefaultTransaction: true,  // Don't wrap single operations in tx
})

// Prepared statement cached and reused
for _, id := range merchantIDs {
    var merchant Merchant
    db.Where("merchant_id = ?", id).First(&merchant)
    // ✅ SQL parsed once, plan cached, reused 100 times
}

// CORRECT: Manual prepared statements for critical queries
func GetMerchantBatch(ctx context.Context, ids []string) ([]Merchant, error) {
    stmt := db.Session(&gorm.Session{PrepareStmt: true}).
        Where("merchant_id = ?", "")  // Template

    var merchants []Merchant
    for _, id := range ids {
        var merchant Merchant
        stmt.Where("merchant_id = ?", id).First(&merchant)
        merchants = append(merchants, merchant)
    }

    return merchants, nil
}
```

### Detection Strategy

```bash
# Find GORM initialization
grep -n "gorm.Open" <pr_files>

# Check for PrepareStmt config
grep -A 5 "gorm.Config" <pr_files> | grep "PrepareStmt"
```

### Flag Conditions

Flag if:
- `gorm.Open()` without `PrepareStmt: true` in config
- Service handles high query volume (API, worker)
- Repeated queries in loops without prepared statements

### Severity

⚠️ **High** - Performance degradation:
- Repeated SQL parsing overhead
- Higher CPU usage
- Slower query execution
- Database load increase

---

## Check 6: Connection Pool Configuration ⚠️ HIGH

### What to Check

Database connection pools must be configured with appropriate limits to prevent connection exhaustion.

### Razorpay Context

From observability skills, connection pool issues are common:
- Canary sentinel checks for: `"too many connections"`, `"pool exhausted"`
- Service opex reports: "connection pool exhaustion", "check pool metrics"

### Bad Pattern ❌

```go
// ANTI-PATTERN: No pool configuration (uses defaults)
db, _ := gorm.Open(postgres.Open(dsn))

// Default settings (often inadequate for production):
// MaxOpenConns: unlimited (can exhaust DB)
// MaxIdleConns: 2 (too low, creates connection churn)
// ConnMaxLifetime: 0 (connections never recycled)

// ❌ Under high load:
// - Opens unlimited connections → DB maxes out
// - Only 2 idle connections → constant connect/disconnect
// - Stale connections never closed
```

### Good Pattern ✅

```go
// CORRECT: Production-ready connection pool
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    PrepareStmt: true,
})

sqlDB, err := db.DB()

// ✅ Configure connection pool
sqlDB.SetMaxOpenConns(25)                  // Limit total connections
sqlDB.SetMaxIdleConns(10)                  // Keep warm connections
sqlDB.SetConnMaxLifetime(5 * time.Minute)  // Recycle connections
sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Close idle connections

// Rule of thumb:
// MaxOpenConns = (number of pods * connections per pod)
// If 10 pods, each needs ~2-3 connections = 25-30 total
// MaxIdleConns = MaxOpenConns / 2 or 3

// CORRECT: Environment-specific configuration
type PoolConfig struct {
    MaxOpenConns    int           `toml:"max_open_conns"`
    MaxIdleConns    int           `toml:"max_idle_conns"`
    ConnMaxLifetime time.Duration `toml:"conn_max_lifetime"`
}

// prod-live.toml
[database.pool]
max_open_conns = 25
max_idle_conns = 10
conn_max_lifetime = "5m"

// stage-live.toml
[database.pool]
max_open_conns = 10
max_idle_conns = 5
conn_max_lifetime = "5m"

// CORRECT: With health check
func InitDB(cfg DatabaseConfig) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn))
    if err != nil {
        return nil, err
    }

    sqlDB, _ := db.DB()

    // Configure pool
    sqlDB.SetMaxOpenConns(cfg.Pool.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.Pool.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(cfg.Pool.ConnMaxLifetime)

    // ✅ Verify connectivity
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := sqlDB.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return db, nil
}
```

### Detection Strategy

```bash
# Find DB initialization
grep -n "gorm.Open\|sql.Open" <pr_files>

# Check for pool configuration
grep -A 10 "gorm.Open" <pr_files> | grep -E "SetMaxOpenConns|SetMaxIdleConns|SetConnMaxLifetime"
```

### Flag Conditions

Flag if:
- `gorm.Open()` or `sql.Open()` without subsequent pool configuration
- No `SetMaxOpenConns()` call (unlimited connections)
- No `SetMaxIdleConns()` call (defaults to 2)
- High-traffic service (API, worker) without pool limits

### Severity

⚠️ **High** - Operational issues:
- Connection exhaustion (common in Razorpay logs)
- Database maxing out connections
- Connection churn (constant connect/disconnect)
- Poor performance under load

### References

**Razorpay observability:**
- `canary-sentinel:SKILL.md` - Checks for "too many connections", "pool exhausted"
- `service-opex:SAMPLE_REPORT.md` - "connection pool exhaustion", "check pool metrics"

---

## Summary Table

| Check # | Pattern | Severity | Razorpay Service Examples |
|---------|---------|----------|---------------------------|
| 1 | Row-Level Locking (FOR UPDATE) | 🚨 Critical | wallet, payments-card-present, mco, loyalty-booking |
| 2 | Deadlock Prevention & Handling | ⚠️ High | governor (metrics), digital-billing (retry), wallet (ordering) |
| 3 | JSONB Index Usage | 📋 Medium | mozart, terminals, splitz, business-verification |
| 4 | SSL Mode in Production | 🚨 Critical | terminals, x-payroll-compute, upidp, rproxy |
| 5 | Prepared Statements | ⚠️ High | All GORM services |
| 6 | Connection Pool Config | ⚠️ High | All services (canary monitors this) |

---

## How to Apply

**For each file matching** `*.go` with PostgreSQL/GORM usage:

1. **Row Locking**: Check financial operations for `Clauses(clause.Locking{...})`
2. **Deadlocks**: Verify lock ordering, check for retry logic
3. **JSONB**: Ensure GIN indexes on JSONB columns being queried
4. **SSL Mode**: Verify production configs use `sslmode=require` or `verify-full`
5. **Prepared Statements**: Check `gorm.Config{PrepareStmt: true}`
6. **Connection Pool**: Verify `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`

**Example output:**

```
📁 File: internal/wallet/repo.go

✅ Check #1 Passed: SELECT FOR UPDATE used (Line 45)
   Code: tx.Clauses(clause.Locking{Strength: "UPDATE"})

✅ Check #2 Passed: Lock ordering implemented (Line 52)
   Code: IDs sorted alphabetically before locking

⚠️  Check #3 Failed: No GIN index on JSONB column (Line 12)
   Code: Metadata postgres.Jsonb `gorm:"type:jsonb"`
   Fix: Add migration with CREATE INDEX USING gin(metadata)

🚨 Check #4 Failed: Production config uses sslmode=disable (configs/prod-live.toml:8)
   Fix: Change to sslmode=require

✅ Check #5 Passed: PrepareStmt enabled (pkg/db/db.go:23)

⚠️  Check #6 Failed: No connection pool configuration (pkg/db/db.go:25)
   Fix: Add SetMaxOpenConns(25), SetMaxIdleConns(10)
```

---

## Integration with Razorpay Monitoring

### Metrics to Emit

Based on governor service pattern:

```go
// Deadlock metric (governor pattern)
metric.DBDeadlockTotal.WithLabelValues(tableName).Inc()

// Connection pool metrics (recommend adding)
metric.DBConnectionPoolActive.Set(float64(stats.OpenConnections))
metric.DBConnectionPoolIdle.Set(float64(stats.Idle))
metric.DBConnectionPoolWait.Add(float64(stats.WaitCount))
```

### Canary Checks

Existing canary-sentinel checks:
- `"too many connections"` - Triggers on pool exhaustion
- `"deadlock"` - Triggers on deadlock errors
- `"connection refused"` - Triggers on DB unavailability

### Service Opex Integration

When investigating P99 latency spikes, check:
- Connection pool metrics
- Slow query logs
- Deadlock frequency
- JSONB query performance
