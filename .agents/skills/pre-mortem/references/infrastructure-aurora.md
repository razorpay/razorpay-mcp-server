# Aurora (MySQL-compatible) Infrastructure Checks

## Overview

Validates AWS Aurora MySQL usage patterns in Razorpay services to prevent connection issues, replication lag, failover problems, and security vulnerabilities. Aurora is used extensively in Razorpay's data platform for transactional workloads with CDC to analytics tier.

**Load when:** PR modifies Aurora/RDS database connections or configurations

**Total Checks:** 6

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 3
- 📋 Medium: 1

---

## Razorpay Aurora Context

### Architecture Pattern (from data platform advisor)

```
Application Services
        ↓
    Aurora MySQL (Primary)
        ↓
   Aurora Replicas (Read)
        ↓
    Debezium CDC
        ↓
   Kafka Topics
        ↓
  Analytics Tier (TiDB, Pinot, Iceberg)
```

**CDC Flow:**
- `Aurora → Debezium → Kafka → Spark → Downstream`
- Real-time data replication for analytics
- Replication lag monitoring critical

**Connection String Example:**
- `database.hostname: aurora-payments.rds.amazonaws.com`

**Common Issues (from observability):**
- Connection pool exhaustion
- "too many connections"
- "connection refused" (during failover)
- Replication lag affecting CDC

---

## Check 1: Reader Endpoint for Read Queries 🚨 CRITICAL

### What to Check

SELECT queries must use Aurora reader endpoint to distribute load and optimize costs.

### Razorpay Pattern

From data platform architecture:
- **Writer endpoint**: Application writes (INSERT, UPDATE, DELETE)
- **Reader endpoint**: Analytics queries, reporting, CDC consumers
- **Cost**: Reader capacity is cheaper than writer

### Bad Pattern ❌

```go
// ANTI-PATTERN: All queries hit writer endpoint
type Database struct {
    DB *gorm.DB  // Single connection to writer!
}

func InitDB(writerHost string) *Database {
    dsn := fmt.Sprintf("host=%s user=app dbname=payments sslmode=require", writerHost)
    db, _ := gorm.Open(mysql.Open(dsn))

    return &Database{DB: db}
}

// ❌ Read query hits writer!
func (d *Database) GetMerchants(ids []string) ([]*Merchant, error) {
    var merchants []*Merchant
    // ❌ SELECT on writer - wastes expensive writer capacity
    d.DB.Where("merchant_id IN ?", ids).Find(&merchants)
    return merchants, nil
}

// ❌ Hardcoded endpoints (can't failover)
const (
    WriterHost = "prod-cluster.cluster-abc.us-east-1.rds.amazonaws.com"
    ReaderHost = "prod-cluster.cluster-ro-abc.us-east-1.rds.amazonaws.com"
)
```

### Good Pattern ✅

```go
// CORRECT: Separate writer and reader connections (Razorpay standard)
type Database struct {
    Writer *gorm.DB  // Primary endpoint (writes)
    Reader *gorm.DB  // Reader endpoint (reads)
}

type AuroraConfig struct {
    WriterEndpoint string `toml:"writer_endpoint"`
    ReaderEndpoint string `toml:"reader_endpoint"`
    User           string `toml:"user"`
    Password       string `toml:"password"`
    DBName         string `toml:"dbname"`
}

func InitDB(cfg AuroraConfig) (*Database, error) {
    // ✅ Writer connection
    writerDSN := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true",
        cfg.User, cfg.Password, cfg.WriterEndpoint, cfg.DBName)
    writer, err := gorm.Open(mysql.Open(writerDSN), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("failed to connect to writer: %w", err)
    }

    // ✅ Reader connection
    readerDSN := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true",
        cfg.User, cfg.Password, cfg.ReaderEndpoint, cfg.DBName)
    reader, err := gorm.Open(mysql.Open(readerDSN), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("failed to connect to reader: %w", err)
    }

    return &Database{
        Writer: writer,
        Reader: reader,
    }, nil
}

// ✅ Use reader for SELECT queries
func (d *Database) GetMerchants(ids []string) ([]*Merchant, error) {
    var merchants []*Merchant
    // ✅ Read from reader endpoint
    d.Reader.Where("merchant_id IN ?", ids).Find(&merchants)
    return merchants, nil
}

// ✅ Use writer for writes
func (d *Database) CreateMerchant(merchant *Merchant) error {
    // ✅ Write to writer endpoint
    return d.Writer.Create(merchant).Error
}

// ✅ Use writer for read-modify-write (transaction consistency)
func (d *Database) DeductBalance(merchantID string, amount int) error {
    tx := d.Writer.Begin()  // ✅ Must use writer for transactions
    defer tx.Rollback()

    var merchant Merchant
    // ✅ SELECT FOR UPDATE requires writer
    tx.Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("merchant_id = ?", merchantID).
        First(&merchant)

    merchant.Balance -= amount
    tx.Save(&merchant)

    return tx.Commit().Error
}

// TOML config - prod-live.toml
[database.aurora]
writer_endpoint = "prod-payments.cluster-abc.us-east-1.rds.amazonaws.com"
reader_endpoint = "prod-payments.cluster-ro-abc.us-east-1.rds.amazonaws.com"
user = "api_user"
password = "${DB_PASSWORD}"  # From secrets
dbname = "payments"
```

### Detection Strategy

```bash
# Find database initialization
grep -n "gorm.Open\|sql.Open" <pr_files>

# Check for separate reader/writer
grep -n "Reader\|Writer" <pr_files>

# Find SELECT queries
grep -n "\.Find\|\.First\|\.Where" <pr_files>

# Flag if no reader endpoint config
grep -L "reader" <config_files>
```

### Flag Conditions

Flag if:
- Only one database connection (no reader/writer split)
- Hardcoded endpoints (not from config)
- SELECT queries using writer connection
- Config has `writer_endpoint` but no `reader_endpoint`
- Heavy read workload (analytics, reporting) without reader

### Severity

🚨 **Critical** - Cost and performance impact:
- Expensive writer capacity wasted on reads
- Writer overloaded, affects write performance
- Higher Aurora costs (writer > reader)
- Scalability issues (writer is bottleneck)

### References

**Razorpay architecture:**
- Data platform: Aurora → CDC → Kafka → Analytics
- Pattern: Transactional access via Aurora (hot storage)
- CDC pattern requires reader endpoint for Debezium

---

## Check 2: RDS Proxy for Serverless/Lambda ⚠️ HIGH

### What to Check

Lambda functions and serverless workloads must use RDS Proxy to prevent connection exhaustion.

### Aurora Connection Limits

- **r5.large**: ~90 connections
- **r5.xlarge**: ~180 connections
- **Lambda**: Can spawn 1000s of concurrent executions → exhausts Aurora connections

### Bad Pattern ❌

```go
// ANTI-PATTERN: Lambda creating new connections (exhausts Aurora)
var db *gorm.DB  // Global variable

func Handler(ctx context.Context, event events.APIGatewayProxyRequest) error {
    // ❌ New connection per Lambda invocation!
    dsn := os.Getenv("AURORA_ENDPOINT")
    db, _ := gorm.Open(mysql.Open(dsn))
    defer db.Close()  // ❌ Connection killed after each request

    // Query...
    var merchants []Merchant
    db.Find(&merchants)

    return nil
}

// Problem: 1000 concurrent Lambda invocations = 1000 connections!
// Aurora r5.large max = 90 → ❌ Connection exhaustion
```

### Good Pattern ✅

```go
// CORRECT: Use RDS Proxy for connection pooling
var db *gorm.DB  // Global variable, reused across invocations

func init() {
    // ✅ Connect to RDS Proxy (not direct Aurora endpoint)
    proxyEndpoint := os.Getenv("RDS_PROXY_ENDPOINT")
    // Example: my-proxy.proxy-abc.us-east-1.rds.amazonaws.com

    dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        proxyEndpoint,  // ✅ RDS Proxy endpoint
        os.Getenv("DB_NAME"),
    )

    db, _ = gorm.Open(mysql.Open(dsn), &gorm.Config{})

    // Configure for serverless
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(2)       // ✅ Low for Lambda
    sqlDB.SetMaxIdleConns(1)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)
    sqlDB.SetConnMaxIdleTime(1 * time.Minute)
}

func Handler(ctx context.Context, event events.APIGatewayProxyRequest) error {
    // ✅ Reuse global connection through RDS Proxy
    var merchants []Merchant
    db.WithContext(ctx).Find(&merchants)

    return nil
}

// CORRECT: RDS Proxy with IAM authentication (no passwords!)
func initWithIAM() {
    proxyEndpoint := os.Getenv("RDS_PROXY_ENDPOINT")

    // ✅ Generate IAM auth token
    cfg, _ := config.LoadDefaultConfig(context.Background())
    authToken, _ := auth.BuildAuthToken(
        context.Background(),
        proxyEndpoint+":3306",
        "us-east-1",
        os.Getenv("DB_USER"),
        cfg.Credentials,
    )

    dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?tls=true&allowCleartextPasswords=true",
        os.Getenv("DB_USER"),
        authToken,  // ✅ IAM token (rotates automatically)
        proxyEndpoint,
        os.Getenv("DB_NAME"),
    )

    db, _ = gorm.Open(mysql.Open(dsn))
}

// Terraform/CDK - RDS Proxy configuration
resource "aws_db_proxy" "payments" {
  name                   = "payments-proxy"
  engine_family          = "MYSQL"
  auth {
    auth_scheme = "SECRETS"  # Or IAM
    iam_auth    = "REQUIRED"
    secret_arn  = aws_secretsmanager_secret.db_credentials.arn
  }

  role_arn               = aws_iam_role.proxy.arn
  vpc_subnet_ids         = var.private_subnet_ids

  # ✅ Connection pooling settings
  require_tls            = true
  idle_client_timeout    = 300   # 5 minutes
  max_connections_percent = 50   # Use 50% of Aurora capacity
}
```

### Detection Strategy

```bash
# Find Lambda handlers
grep -n "events.APIGatewayProxyRequest\|lambda.Start" <pr_files>

# Check for RDS Proxy
grep -n "RDS_PROXY\|proxy.*endpoint" <pr_files>

# Flag direct Aurora connections in Lambda
grep -n "rds.amazonaws.com" <lambda_files>
```

### Flag Conditions

Flag if:
- Lambda function connects directly to Aurora (not via proxy)
- Environment variable is `AURORA_ENDPOINT` not `RDS_PROXY_ENDPOINT`
- New connection created per Lambda invocation
- Serverless workload without connection pooling

### Severity

⚠️ **High** - Connection exhaustion:
- Lambda cold starts exhaust Aurora connections
- "too many connections" errors
- Service degradation
- Cascading failures

---

## Check 3: Failover Retry Logic ⚠️ HIGH

### What to Check

Application must handle Aurora automatic failover (<30 seconds) with retry logic.

### Razorpay Context

From canary-sentinel, common errors:
- `"connection refused"` (during failover)
- `"connection timeout"`
- `"broken pipe"`

Aurora failover:
- Primary fails → Replica promoted to primary (~10-30 seconds)
- Application must retry connections

### Bad Pattern ❌

```go
// ANTI-PATTERN: No retry on failover
func QueryMerchants(ctx context.Context) ([]*Merchant, error) {
    var merchants []*Merchant
    err := db.WithContext(ctx).Find(&merchants).Error
    if err != nil {
        return nil, err  // ❌ Fails immediately on failover
    }
    return merchants, nil
}

// During Aurora failover:
// - Primary down
// - App tries to query → connection refused
// - Returns error to user
// - ❌ User sees "Database unavailable"
```

### Good Pattern ✅

```go
// CORRECT: Retry with exponential backoff (Razorpay pattern)
func QueryMerchantsWithRetry(ctx context.Context) ([]*Merchant, error) {
    var merchants []*Merchant

    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        err := db.WithContext(ctx).Find(&merchants).Error

        if err == nil {
            return merchants, nil
        }

        // ✅ Check if retriable error (connection issues during failover)
        if isRetriableDBError(err) {
            backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
            logger.Warn(ctx, "db_failover_retry",
                "attempt", i+1,
                "backoff", backoff,
                "error", err)

            select {
            case <-time.After(backoff):
                continue
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }

        // Non-retriable error - fail fast
        return nil, err
    }

    return nil, errors.New("max retries exceeded")
}

// CORRECT: Detect retriable errors (from canary-sentinel patterns)
func isRetriableDBError(err error) bool {
    if err == nil {
        return false
    }

    errStr := err.Error()

    // Aurora failover errors (from observability patterns)
    retriablePatterns := []string{
        "connection refused",
        "broken pipe",
        "connection reset",
        "connection timeout",
        "no such host",          // DNS propagation during failover
        "server has gone away",  // MySQL error during failover
        "communications link failure",
    }

    for _, pattern := range retriablePatterns {
        if strings.Contains(strings.ToLower(errStr), pattern) {
            return true
        }
    }

    return false
}

// CORRECT: Connection health check with retry
func (d *Database) HealthCheck(ctx context.Context) error {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        sqlDB, _ := d.Writer.DB()
        err := sqlDB.PingContext(ctx)

        if err == nil {
            return nil
        }

        if isRetriableDBError(err) {
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }

        return err
    }

    return errors.New("health check failed after retries")
}
```

### Detection Strategy

```bash
# Find database query operations
grep -n "\.Find\|\.First\|\.Create\|\.Update" <pr_files>

# Check for retry logic
grep -n "retry\|Retry" <pr_files>

# Flag files with critical DB operations but no retry logic
# Note: grep -L does not work on piped input — use per-file loop instead
for file in <pr_files>; do
    if grep -qE "\.Find\(|\.First\(|\.Create\(|\.Update\(" "$file" && \
       grep -qiE "payment|refund|transfer" "$file" && \
       ! grep -qiE "retry|Retry|maxAttempt|backoff" "$file"; then
        echo "⚠️  $file: Critical DB operations without retry logic"
    fi
done
```

### Flag Conditions

Flag if:
- Critical database operations without retry logic
- No connection error handling
- Payment/financial operations fail immediately on error
- Health check without retry

### Severity

⚠️ **High** - Service availability:
- Unnecessary downtime during failover
- Poor user experience
- Failed transactions during brief outages
- SLA impact

---

## Check 4: Replication Lag Monitoring (CDC) ⚠️ HIGH

### What to Check

Aurora → CDC → Kafka pipelines must monitor replication lag to prevent stale data in analytics.

### Razorpay Data Platform Pattern

```
Aurora Primary (Writes)
    ↓
Aurora Replica (Debezium reads binlog)
    ↓
Kafka (CDC events)
    ↓
Analytics Tier (TiDB, Pinot, Iceberg)
```

**Issue:** Replica lag > 1 second → Stale analytics data

### Bad Pattern ❌

```go
// ANTI-PATTERN: No replication lag monitoring
func ProcessCDCEvents(ctx context.Context) {
    // ❌ Reads from replica assuming zero lag
    var merchants []Merchant
    db.Reader.Find(&merchants)

    // Publish to Kafka for CDC
    for _, m := range merchants {
        kafka.Publish("merchants", m)
    }
    // ❌ May publish stale data if replica is lagging!
}
```

### Good Pattern ✅

```go
// CORRECT: Check replica lag before reading
func getReplicaLag(ctx context.Context) (time.Duration, error) {
    var seconds float64

    // ✅ Query Aurora replica lag (MySQL compatible)
    query := `
        SELECT
            TIMESTAMPDIFF(SECOND,
                          CONVERT_TZ(MAX(ts), '+00:00', @@global.time_zone),
                          NOW()) AS replica_lag_seconds
        FROM mysql.rds_heartbeat2
    `

    err := db.Reader.Raw(query).Scan(&seconds).Error
    if err != nil {
        return 0, err
    }

    return time.Duration(seconds) * time.Second, nil
}

// CORRECT: Alert on high lag
func ProcessCDCEventsWithLagCheck(ctx context.Context) error {
    lag, err := getReplicaLag(ctx)
    if err != nil {
        logger.Error(ctx, "failed_to_check_replica_lag", "error", err)
        // Fall back to writer (consistent)
        return readFromWriter(ctx)
    }

    // ✅ Alert if lag > threshold
    if lag > 5*time.Second {
        logger.Warn(ctx, "high_replication_lag",
            "lag_seconds", lag.Seconds(),
            "threshold", 5)

        metric.AuroraReplicaLag.Set(lag.Seconds())

        // Switch to writer for consistent data
        return readFromWriter(ctx)
    }

    // Replica is fresh enough
    return readFromReplica(ctx)
}

// CORRECT: CloudWatch metric alarm
// Terraform/CDK
resource "aws_cloudwatch_metric_alarm" "replica_lag" {
  alarm_name          = "aurora-replica-lag-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "AuroraReplicaLag"
  namespace           = "AWS/RDS"
  period              = "60"
  statistic           = "Average"
  threshold           = "1000"  # 1 second in milliseconds
  alarm_description   = "Alert when Aurora replica lag > 1s"

  dimensions = {
    DBClusterIdentifier = aws_rds_cluster.payments.id
  }
}
```

### Detection Strategy

```bash
# Find CDC/Debezium configuration
grep -n "debezium\|cdc\|binlog" <pr_files>

# Check for lag monitoring
grep -n "replica.*lag\|replication.*lag" <pr_files>

# Flag files that publish to Kafka after reading from .Reader. without lag monitoring.
# CDC config (debezium settings) lives in infra files, not Go service code —
# co-locating the debezium keyword requirement in the same file silences this check.
# Instead, use Kafka publish alongside Reader reads as the CDC proxy signal.
for file in <pr_files>; do
    if grep -qE "\.Reader\." "$file" && \
       grep -qE "kafka\.(Publish|Produce|Send)|\.Produce\(|\.SendMessage\(" "$file" && \
       ! grep -qiE "replica.?lag|replication.?lag|lag.*check|checkLag|rds_heartbeat" "$file"; then
        echo "⚠️  $file: Reader endpoint used alongside Kafka publish — add replication lag check to avoid stale CDC data"
    fi
done
```

### Flag Conditions

Flag if:
- CDC pipeline reads from replica
- No replication lag monitoring
- No alerting on high lag
- Analytics queries without lag consideration

### Severity

⚠️ **High** - Data quality issues:
- Stale analytics data
- Incorrect business decisions
- User-visible inconsistencies
- CDC pipeline delays

---

## Check 5: Connection String from Secrets Manager 🚨 CRITICAL

### What to Check

Aurora credentials must be fetched from AWS Secrets Manager or Kubestash (not hardcoded).

### Razorpay Pattern

From code security skill:
- **Kubestash**: Pulls secrets from DynamoDB `kubestash-*` tables
- Pushes to Kubernetes secrets (runs every 10 minutes)
- Application reads from Kubernetes secrets

### Bad Pattern ❌

```go
// ANTI-PATTERN: Hardcoded credentials
dsn := "user:password123@tcp(aurora-cluster.us-east-1.rds.amazonaws.com:3306)/mydb"
db, _ := gorm.Open(mysql.Open(dsn))
// ❌ Password in code!

// ANTI-PATTERN: Credentials in config file
// prod-live.toml
[database]
host = "aurora-payments.rds.amazonaws.com"
user = "api_user"
password = "SuperSecret123"  # ❌ Committed to Git!

// ANTI-PATTERN: Credentials in environment (plain text)
// Dockerfile
ENV DB_PASSWORD=hardcoded_password  # ❌ Visible in image
```

### Good Pattern ✅

```go
// CORRECT: Fetch from Secrets Manager (Razorpay pattern)
func getDBCredentials(ctx context.Context) (*DBCredentials, error) {
    sess := session.Must(session.NewSession())
    svc := secretsmanager.New(sess)

    // ✅ Fetch from Secrets Manager
    result, err := svc.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
        SecretId: aws.String("prod/aurora/payments/credentials"),
    })
    if err != nil {
        return nil, err
    }

    var creds DBCredentials
    json.Unmarshal([]byte(*result.SecretString), &creds)

    return &creds, nil
}

type DBCredentials struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
    DBName   string `json:"dbname"`
}

func InitDB(ctx context.Context) (*gorm.DB, error) {
    // ✅ Fetch secrets at runtime
    creds, err := getDBCredentials(ctx)
    if err != nil {
        return nil, err
    }

    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        creds.Username,
        creds.Password,
        creds.Host,
        creds.Port,
        creds.DBName,
    )

    return gorm.Open(mysql.Open(dsn))
}

// CORRECT: Kubestash pattern (Razorpay standard)
// 1. Store secret in DynamoDB kubestash-* table
// 2. Kubestash pulls and pushes to K8s secret (every 10 min)
// 3. App reads from K8s secret via env var

// Kubernetes deployment
apiVersion: v1
kind: Secret
metadata:
  name: aurora-credentials
type: Opaque
data:
  DB_HOST: <base64>
  DB_USERNAME: <base64>
  DB_PASSWORD: <base64>  # ✅ Managed by Kubestash

---
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: api
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: aurora-credentials
              key: DB_HOST
        - name: DB_USERNAME
          valueFrom:
            secretKeyRef:
              name: aurora-credentials
              key: DB_USERNAME
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: aurora-credentials
              key: DB_PASSWORD

// Application code
func InitDB() (*gorm.DB, error) {
    // ✅ Read from environment (injected by K8s from secret)
    dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
        os.Getenv("DB_USERNAME"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_NAME"),
    )

    return gorm.Open(mysql.Open(dsn))
}

// CORRECT: TOML with secret reference (not actual secret)
[database.aurora]
writer_endpoint = "prod-payments.cluster-abc.us-east-1.rds.amazonaws.com"
reader_endpoint = "prod-payments.cluster-ro-abc.us-east-1.rds.amazonaws.com"
user = "api_user"
password = "${DB_PASSWORD}"  # ✅ Reference to env var, not actual password
dbname = "payments"
```

### Detection Strategy

```bash
# Find database connections
grep -n "gorm.Open\|sql.Open" <pr_files>

# Check for hardcoded credentials
grep -n "password.*=.*\"" <pr_files> <config_files>

# Flag suspicious patterns
grep -E "(password|passwd|pwd).*[=:].*[\"']" <pr_files>
```

### Flag Conditions

Flag if:
- Hardcoded password in code or config
- Config file contains actual password (not `${ENV_VAR}`)
- Connection string with embedded credentials
- No Secrets Manager or Kubestash usage

### Severity

🚨 **Critical** - Security vulnerability:
- Credentials exposed in Git history
- Easy to leak in logs, errors
- No automatic rotation
- Compliance violations (PCI-DSS, SOC 2)

### References

**Razorpay security:**
- Kubestash pulls from DynamoDB `kubestash-*` tables
- Pushes to K8s secrets every 10 minutes
- Standard pattern across Razorpay services

---

## Check 6: Aurora Serverless Auto-Scaling Config 📋 MEDIUM

### What to Check

Aurora Serverless v2 must have appropriate ACU (Aurora Capacity Units) limits to prevent cost runaway.

### Bad Pattern ❌

```toml
# ANTI-PATTERN: Unbounded auto-scaling
[database.aurora_serverless]
min_capacity = 0.5    # ❌ Too low - cold starts on every spike
max_capacity = 256    # ❌ Unbounded - potential $$$$ bill
timeout_action = "ForceApplyCapacityChange"  # ❌ Can drop connections
auto_pause = true     # ❌ Cold starts in production
```

### Good Pattern ✅

```toml
# CORRECT: Bounded auto-scaling
[database.aurora_serverless]
min_capacity = 2      # ✅ Warm, handles baseline load
max_capacity = 16     # ✅ Capped at reasonable limit (monitor and adjust)
timeout_action = "RollbackCapacityChange"  # ✅ Don't drop connections
auto_pause = false    # ✅ No cold starts in production
```

### Severity

📋 **Medium** - Cost and performance:
- Unbounded scaling → surprise AWS bill
- Auto-pause → cold start latency
- Too low min → constant scaling churn

---

## Summary Table

| Check # | Pattern | Severity | Impact |
|---------|---------|----------|--------|
| 1 | Reader Endpoint for Reads | 🚨 Critical | Cost optimization, writer offload |
| 2 | RDS Proxy for Lambda | ⚠️ High | Connection exhaustion prevention |
| 3 | Failover Retry Logic | ⚠️ High | Service availability during failover |
| 4 | Replication Lag (CDC) | ⚠️ High | Analytics data freshness |
| 5 | Secrets Manager/Kubestash | 🚨 Critical | Security, credential protection |
| 6 | Serverless Auto-Scaling | 📋 Medium | Cost control |

---

## Integration with Razorpay Systems

### Data Platform
- Aurora → Debezium CDC → Kafka → Analytics tier
- Monitor replication lag for CDC accuracy
- Use reader endpoint for Debezium

### Observability
- Canary checks: "connection refused", "too many connections"
- CloudWatch: `AuroraReplicaLag`, `DatabaseConnections`
- Service opex: Connection pool metrics

### Security
- Kubestash manages secrets in DynamoDB
- Rotates credentials automatically
- Injects into K8s secrets
