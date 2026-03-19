# Infrastructure Extensions Brainstorm - Pre-Mortem Skill

Based on analysis of Razorpay codebase patterns discovered through code scanning.

---

## Current State

**Existing checks:**
- Generic Database (6 checks): Transactions, replica lag, timeouts, migrations, indexes, N+1
- Redis (8 checks): TTL, locks, stampede, pool, errors, naming, invalidation
- Kafka (8 checks): Panic recovery, DLQ, idempotency, topics
- Eventing (6 checks): Schema validation, versioning, context propagation

**Total: 28 checks**

---

## Patterns Found in Razorpay Codebase

### Database Patterns
- **PostgreSQL + GORM** (x-payroll services: flexible-benefits, compute, compliance)
- **Aurora** for transactional workloads with CDC to data platform
- Health checks: `db.PingContext(ctx)` standard
- **Common issues identified:**
  - Connection pool exhaustion (frequent logs)
  - N+1 queries (mentioned in multiple skills)
  - Missing indexes (Java/Go code reviews)
  - Deadlocks (canary sentinel checks for this)
  - Slow query logs analysis
  - Replica lag monitoring

### Redis/Cache Patterns
- **Mutex with TTL**: 30s TTL, 3 retries (600ms, 1200ms delays) - BharatQR example
- Health checks: `rdb.Ping(ctx)` standard
- **Common issues:**
  - Pool exhaustion (logs mention "pool exhausted")
  - Redis connection failures
  - Cache stampede (mentioned in Java reviews)

### DynamoDB Usage
- QR data migration to DynamoDB
- DynamoDB Streams for CDC
- Credstash/Kubestash: Secret storage in DynamoDB tables (`kubestash-*`)

### Config Patterns
- Environment variables: `DB_HOST`, `DB_USERNAME`, `DB_PASSWORD`
- Secret references: `<namespace>/<secret-name>/DB_HOST`
- Config from Secrets Manager (Kubestash pulls from DynamoDB)

---

## Extended Coverage Plan

### 1. DynamoDB Checks (8 checks)

#### **A. Data Modeling** (Critical)
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 1 | **Partition Key Design** | 🚨 Critical | Generic best practice |
| 2 | **Sort Key for Range Queries** | 📋 Medium | Generic best practice |

**Patterns to detect:**
```python
# ❌ Low cardinality partition key
partition_key = "status"  # All "active" items in same partition

# ✅ High cardinality
partition_key = "merchant_id"  # Distributed across partitions
```

#### **B. Write Safety** (Critical)
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 3 | **Conditional Writes** | 🚨 Critical | Similar to BharatQR duplicate check |
| 4 | **Transaction Usage** | ⚠️ High | Multi-table updates need atomicity |

**Pattern from codebase:**
- BharatQR uses duplicate detection (`findByProviderReferenceIdAndAmount`)
- DynamoDB equivalent: `ConditionExpression: "attribute_not_exists(payment_id)"`

#### **C. Performance** (High)
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 5 | **Batch Operations** | ⚠️ High | Avoid N+1 (common Razorpay issue) |
| 6 | **GSI/LSI Design** | 📋 Medium | Query optimization |

**Pattern:**
- Code reviews mention N+1 query issues
- DynamoDB: Use BatchGetItem instead of loops

#### **D. Data Lifecycle** (Critical)
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 7 | **TTL Configuration** | 🚨 Critical | Session/temp data expiration |
| 8 | **Read Consistency** | ⚠️ High | Read-after-write scenarios |

**Pattern from codebase:**
- TTL=30s for mutex locks (BharatQR)
- DynamoDB: Enable TTL attribute on tables

---

### 2. Aurora Checks (6 checks)

#### **A. Endpoint Management** (Critical)
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 1 | **Reader Endpoint for Reads** | 🚨 Critical | Data platform refs Aurora |
| 2 | **RDS Proxy for Lambda** | ⚠️ High | Serverless connection pooling |
| 3 | **Failover Retry Logic** | ⚠️ High | Connection refused errors logged |

**Patterns from codebase:**
```yaml
# Data platform references
database.hostname: aurora-payments.rds.amazonaws.com

# Health checks
db.PingContext(ctx)  # Standard pattern
```

**Common errors (from canary-sentinel):**
- `connection refused`
- `connection timeout`
- `too many connections`
- `deadlock`

#### **B. Replication & Performance**
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 4 | **Global DB Replication Lag** | ⚠️ High | Aurora → CDC → Kafka pattern |
| 5 | **Serverless Auto-Scaling** | 📋 Medium | ACU limits config |
| 6 | **Connection String Security** | 🚨 Critical | Kubestash/DynamoDB secrets |

**Patterns from codebase:**
- **CDC Flow**: `Aurora → Debezium → Kafka → TiDB/Iceberg/Pinot`
- **Secrets**: Kubestash pulls from DynamoDB `kubestash-*` tables
- **Config**: Environment vars `DB_HOST`, `DB_USERNAME`, `DB_PASSWORD`

---

### 3. PostgreSQL Checks (6 checks)

#### **A. Concurrency & Locking** (Critical)
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 1 | **Row-Level Locking (FOR UPDATE)** | 🚨 Critical | x-payroll uses GORM + PostgreSQL |
| 2 | **Deadlock Handling** | ⚠️ High | Canary checks for deadlocks |

**Patterns from codebase:**
```go
// x-payroll services use GORM + PostgreSQL
// Table names from TableName() functions
// Transactions critical (claims, assignments, SDI)
```

**Canary monitors for:**
- `deadlock` (literal string in logs)
- Transaction conflicts

#### **B. Advanced Features**
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 3 | **JSONB Index Usage** | 📋 Medium | AI pentester refs JSONB operators |
| 4 | **Partition Table Optimization** | ⚠️ High | Large tables (payments, transactions) |
| 5 | **SSL Mode in Production** | 🚨 Critical | Security requirement |
| 6 | **Prepared Statements** | 📋 Medium | Performance optimization |

**Patterns from security skill:**
- JSONB operators: `->`, `->>`, `@>`, `?|`
- ORM bypass risks: `whereRaw`, `orderByRaw`
- SQL injection via JSON operators

---

### 4. Kafka Config Checks (10 checks)

#### **Connection & Security** (Critical)
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 1 | **Bootstrap Servers from Config** | 🚨 Critical | Never hardcode |
| 2 | **Security Protocol (SASL_SSL)** | 🚨 Critical | Production auth |
| 3 | **Producer Acks Configuration** | 🚨 Critical | `acks=all` for durability |

**Patterns from codebase:**
```python
# Data platform uses Kafka extensively
.format("kafka")
.option("kafka.bootstrap.servers", "kafka:9092")

# CDC Pattern
Aurora → Debezium → Kafka → Spark → Downstream
```

#### **Performance & Reliability** (High)
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 4 | **Compression Type** | ⚠️ High | snappy/lz4 for efficiency |
| 5 | **Retries and Timeout** | ⚠️ High | BharatQR uses 3 retries |
| 6 | **Max In-Flight Requests** | ⚠️ High | Ordering guarantees |
| 7 | **Consumer Auto Offset Reset** | ⚠️ High | earliest/latest explicit |
| 8 | **Session Timeout vs Heartbeat** | 📋 Medium | Rebalance prevention |
| 9 | **Enable Idempotence** | ⚠️ High | Exactly-once semantics |
| 10 | **Connection Pool Settings** | 📋 Medium | Pool exhaustion prevention |

**Pattern from BharatQR:**
- Retry logic: 3 retries with 600ms, 1200ms delays
- Kafka consumer groups for event processing

---

### 5. Redis/Valkey Extensions (Add to existing)

#### **New Checks to Add**
| # | Check | Severity | Found in Code |
|---|-------|----------|---------------|
| 9 | **Mutex Lock Patterns** | 🚨 Critical | BharatQR mutex with TTL=30s |
| 10 | **Connection Pool Exhaustion** | 🚨 Critical | Canary logs "pool exhausted" |

**Patterns from codebase:**
```go
// BharatQR mutex pattern
mutex := NewMutex(ctx, resourceID, 30*time.Second)
retries := 3
delays := []time.Duration{600*time.Millisecond, 1200*time.Millisecond}

// Health check
rdb.Ping(ctx)  // Standard pattern

// Common errors (canary-sentinel)
"redis error"
"pool exhausted" (redis context)
```

**Valkey Note:**
- Redis-compatible
- Add section to `infrastructure-redis.md`
- Mention dual-write pattern for migration

---

## Common Operational Issues (Found in Logs/Monitoring)

### High Priority (from canary-sentinel, service-opex)
1. **Connection pool exhaustion** (DB and Redis)
   - Logs: "too many connections", "pool exhausted"
   - P99 latency spikes correlated with DB_ERROR
   - Investigation: Check pool metrics, slow queries

2. **Deadlocks**
   - Explicit canary check
   - Transaction conflicts

3. **N+1 Queries**
   - Mentioned in Java and general code reviews
   - Performance degradation

4. **Replica Lag**
   - Aurora → CDC flow sensitive to lag
   - Global DB replication lag monitoring

5. **Slow Queries**
   - P99 latency investigation points to DB slowness
   - Recommendation: Check slow query logs

---

## Proposed File Structure

```
pre-mortem/references/
├── infrastructure-database.md              (existing - generic)
├── infrastructure-aurora.md                (NEW - 6 checks)
├── infrastructure-postgres.md              (NEW - 6 checks)
├── infrastructure-dynamodb.md              (NEW - 8 checks)
├── infrastructure-redis.md                 (existing - add Valkey + 2 new checks)
├── infrastructure-kafka.md                 (existing - consumer/producer)
├── infrastructure-kafka-config.md          (NEW - 10 checks)
└── infrastructure-eventing.md              (existing)
```

**Total new checks: 30**
**Updated checks: 2 (Redis)**
**New total: 60 infrastructure checks**

---

## Implementation Priority

### Phase 1: High Impact (Week 1)
1. **PostgreSQL checks** (6) - x-payroll services use this
2. **Kafka config checks** (10) - CDC heavily used

### Phase 2: AWS Specific (Week 2)
3. **Aurora checks** (6) - data platform dependency
4. **DynamoDB checks** (8) - QR migration, secrets

### Phase 3: Enhancement (Week 3)
5. **Redis extensions** (2 new checks)
6. **Integration testing** on real PRs

---

## Validation Strategy

### Test on Real Razorpay Services
1. **x-payroll-flexible-benefits** (PostgreSQL + GORM)
2. **payment-router** (Aurora, Redis health checks)
3. **bharat-qr** (DynamoDB migration, mutex patterns)
4. **Data platform** (Kafka CDC, Aurora → Kafka flow)

### Success Criteria
- Detects connection pool issues (real logs show this)
- Catches N+1 queries (mentioned in reviews)
- Validates mutex TTL patterns (BharatQR example)
- Checks Kafka config (bootstrap servers, acks)
- Validates secrets from config (not hardcoded)

---

## Key Insights from Codebase Scan

### What Works Well at Razorpay
1. ✅ **Health checks standardized**: `db.PingContext()`, `rdb.Ping()`
2. ✅ **Secrets management**: Kubestash + DynamoDB pattern
3. ✅ **CDC architecture**: Well-defined Aurora → Kafka → Analytics
4. ✅ **Monitoring**: Canary-sentinel checks for common errors

### Common Anti-Patterns Found
1. ❌ **Connection pool exhaustion** (logs show this is recurring)
2. ❌ **N+1 queries** (mentioned in multiple review skills)
3. ❌ **Hardcoded values** (concerns in config reviews)
4. ❌ **Missing retry logic** (BharatQR had to add explicit retries)

### Razorpay-Specific Patterns to Encode
1. **Mutex pattern**: 30s TTL, 3 retries with exponential backoff
2. **Health checks**: Must ping DB and cache
3. **Secret references**: `<namespace>/<secret-name>/DB_HOST` format
4. **CDC pattern**: Aurora → Debezium → Kafka → Downstream
5. **Error monitoring**: Canary checks for specific error strings

---

## Next Steps

1. **Get user confirmation** on scope and priority
2. **Start with PostgreSQL** (most x-payroll services use this)
3. **Use BharatQR flows.md** as reference for mutex/retry patterns
4. **Validate against x-payroll investigation.md** for GORM patterns
5. **Test on real PRs** from payment-router, x-payroll services

---

**Questions for User:**
1. Should we prioritize PostgreSQL (most x-payroll services) or Aurora (data platform)?
2. Include Razorpay-specific patterns (mutex TTL=30s, 3 retries) or keep generic?
3. Validate against actual PR from x-payroll or payment-router services?
