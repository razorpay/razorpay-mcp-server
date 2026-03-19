# Pre-Mortem Skill

Automated pre-mortem analysis for GitHub pull requests to catch reliability, security, quality, and observability issues before they reach production.

## What It Does

This skill performs 199 automated checks across 6 categories:

### 1. Infrastructure (93 checks)
- **Database** (6): Transaction rollback, replica lag, timeouts, migrations, indexes, N+1 queries
- **PostgreSQL** (6): Row-level locking (SELECT FOR UPDATE), deadlock handling, JSONB indexes, SSL mode, prepared statements, connection pool config
- **Aurora** (6): Reader/writer endpoints, RDS Proxy for Lambda, failover retry, replication lag (CDC), Secrets Manager/Kubestash, serverless auto-scaling
- **DynamoDB** (8): Partition key design, conditional writes, batch operations, TTL configuration, GSI projection, TransactWriteItems, ConsistentRead, capacity mode
- **Kafka** (8): Panic recovery, idempotency, DLQ, topic naming, consumer config
- **Kafka Config** (10): Bootstrap servers from config, SSL/SASL security, producer acks, compression (Snappy), retries/timeout, max in-flight requests, auto offset reset, session timeout vs heartbeat, idempotence, connection pooling
- **Query & Index Analysis** (6): Foreign key indexes, WHERE clause indexes, composite indexes, N+1 query detection, ORDER BY indexes, unique constraints
- **Redis** (8): TTL on keys, lock management, error handling, stampede protection
- **SQS** (6): DLQ setup, visibility timeout, message deletion, polling backoff
- **Eventing** (6): Schema validation, type assertions, versioning, context propagation
- **Resilience** (8): Circuit breakers, retry logic, timeouts, error classification
- **Error Handling** (9): Panic recovery (HTTP handlers + goroutines), nil checks, ignored errors, type assertions, error wrapping, defer cleanup
- **Config** (6): Missing prod keys, dangerous defaults, hardcoded secrets

### 2. Services (70 checks)
- **API Contracts** (4): Request/response schemas, error handling, timeout/retry
- **Event Contracts** (4): Event schemas, publisher/consumer compatibility
- **Splitz Integration** (10): Experiment existence, environment activation, variant config, default variant safety, RequestData schema, error handling, identifier consistency, bulk evaluation, environment config, test coverage (6 checks require Splitz MCP)
- **Stork Integration** (12): Service registration (new integrations), environment config, template existence, WhatsApp opt-in, attachment presigned URLs, error handling & retry, rate limiting, DLT registration (India SMS), service name convention, context keys whitelisting, delivery callbacks, mock mode for tests
- **Passport Integration** (10): JWKS host configuration, JWT token extraction, validation & error handling, context storage & retrieval, resource owner extraction, test environment handling, mode & domain validation, role-based access control, initialization timeout, logging & observability
- **ASV Integration** (10): Client configuration & initialization, Paths field validation, error handling & nil checks, write authorization, document handling (FileStoreId prefix), website fields validation, concurrent operations safety, cache configuration, singleton client pattern, gRPC connection tuning
- **Router SDK Integration** (10): Client initialization with mode validation, request validation (nil checks), appropriate request factory method (standard/recurring/intent), mode consistency, terminal response validation, error handling with status code mapping, Basic Auth configuration, context propagation for tracing, timeout configuration, request logging with PII masking
- **PG-Router Service** (10): Mutex lock acquisition with defer unlock, payment status transition validation, service registry validation before calls, gateway field validation (optimizer bypass prevention), callback hash verification (HMAC), API monolith integration error handling, ledger dual-write with Kafka events, error class and identifier code usage (PGPR######), callback idempotency checks, timeout configuration per payment method

### 3. Domain Logic (10 checks)
- **Constraints** (4): Unique constraints, required fields, allowed values, formats
- **Flows** (4): Critical steps, step ordering, cache invalidation, error paths
- **Rules** (2): Validation rules, business logic bypass

### 4. Quality (8 checks)
- **Unit Tests** (4): Test files exist, coverage threshold, new functions covered
- **Integration Tests** (2): SLITs for endpoints, error case coverage
- **Feature Flags** (1): Major features behind Splitz
- **CI Integration** (1): Auto-invoke pr-autopilot on failures

### 5. Performance (18 checks) ⚡ Delegated to db-network-optimizer
- **Duplicate Detection** (8): Duplicate DB queries, duplicate service calls, N+1 patterns in loops, missing param passing, missing request cache, missing GORM preload, repeated config fetches, related entities sequential
- **Query Optimization** (10): Missing indexes (Level 1), suboptimal composite indexes (Level 2), redundant/unused indexes (Level 3), query rewrites
- **How it works:** When performance-relevant files are detected, pre-mortem invokes `/db-network-optimizer` and integrates findings into the comprehensive report

### 6. Observability (5 checks)
- **Monitoring** (2): Error metrics on failures, success metrics on happy path
- **Logging** (3): Trace codes for errors, contextual fields, avoid info log spam

## Related Skills

### Performance Checks: Delegation to db-network-optimizer

pre-mortem **automatically invokes** `/db-network-optimizer` when performance-relevant files are detected in PR. This provides:
- ✅ Single source of truth for performance checks
- ✅ Integrated findings in comprehensive pre-mortem report
- ✅ No duplicate logic (easier maintenance)

**Use `/db-network-optimizer` directly when you need:**
- **Standalone performance analysis** (without other pre-mortem checks)
- **Two-section detailed report** with comprehensive fix strategies
- **Multiple optimization strategies** per issue with before/after code examples
- **Full-repo baseline** with `--full` flag for Reliability Week audits

## Quick Start

```bash
# Review current branch's PR
"Review this PR"

# Review specific PR
"Check PR #456 for issues"

# Quick check
"Is this safe to merge?"

# With CI failures
"Fix CI failures and run pre-mortem"
```

## How It Works

### 1. Repo-Level Context First

**CRITICAL:** Before running any checks, the skill loads repo-level context:

1. **Asks for repo path** if not already in the repository
2. **Loads patterns from**:
   - `.claude/skills/*-skill/` - Repo skill with domain knowledge
   - `CLAUDE.md` - Repo documentation with patterns
   - Code examples - Existing metrics, logging, test patterns

3. **Uses repo patterns** for all subsequent checks:
   - Metrics library syntax
   - Trace code constants
   - Logger format
   - Database patterns
   - Event schemas

**Without repo context, checks use generic patterns that may not match your codebase.**

### 2. Progressive Loading

**Only loads relevant checks based on changed files:**

| Changed Files | References Loaded | Pages |
|---------------|-------------------|-------|
| `configs/*.toml` | config.md | ~15 |
| `internal/*/repo.go` (PostgreSQL/GORM) | postgres.md, error-handling.md | ~30 |
| Aurora config, CDC pipeline | aurora.md, eventing.md | ~30 |
| `pkg/dynamodb/*` | dynamodb.md | ~25 |
| `worker/kafka/*` (handlers) | kafka.md, eventing.md | ~35 |
| Kafka client init, config | kafka-config.md | ~25 |
| `splitz.GetVariant`, experiment IDs | services-splitz.md | ~30 |
| `stork.NewClient`, SMS/Email/WhatsApp | services-stork.md | ~35 |
| `passport.InitHandler`, auth middleware | services-passport.md | ~30 |
| `accountService.NewClient`, ASV operations | services-asv.md | ~30 |
| `router.NewClient`, `GetTerminals`, payment routing | services-router.md | ~25 |
| PG-Router: `internal/payments/`, `callback*.go`, mutex usage | pgrouter-service.md | ~30 |
| Domain entities | constraints.md, flows.md + repo skill | ~40 |

**Total context: 25-80 pages per PR** (repo context + dynamically loaded checks)

### 3. Smart Domain Validation

Auto-discovers domain constraints from repo skill:

```
PR modifies: internal/gateway_credentials/service.go
↓
Loads: .claude/skills/terminals-skill/modules/domain/gateway-credentials/
  - constraints.md (business rules)
  - flows.md (step-by-step processes)
↓
Validates:
  ✓ Unique constraints enforced
  ✓ Required fields validated
  ✓ Critical flow steps present
```

### 4. Actionable Output

```
🔍 PR Pre-Mortem Results for PR #123

📊 Summary:
  🚨 Critical: 2    ⚠️  High: 5
  📋 Medium: 8     ℹ️  Low: 3
  ✅ Passed: 64/87

🚨 CRITICAL ISSUES:

1. [Database] Missing Transaction Rollback
   File: internal/terminals/repo.go:422
   Fix: Add "defer tx.Rollback()" after db.Begin()

2. [Config] Missing prod-live.toml Keys
   File: configs/default.toml:45
   Risk: Production will use dev defaults!
   Fix: Add to prod-live.toml, prod-test.toml

What would you like me to do?
1. 🔧 Fix issues automatically
2. 💬 Add review comments to PR
3. 🔍 Investigate CI failures
```

## File Structure

```
pr-pre-mortem/
├── SKILL.md                                # Main orchestrator
├── README.md                               # This file
└── references/
    ├── infrastructure-database.md          # 6 DB checks
    ├── infrastructure-kafka.md             # 8 Kafka checks
    ├── infrastructure-redis.md             # 8 Redis checks
    ├── infrastructure-sqs.md               # 6 SQS checks
    ├── infrastructure-eventing.md          # 6 event checks
    ├── infrastructure-resilience.md        # 8 resilience checks
    ├── infrastructure-error-handling.md    # 9 error checks
    ├── infrastructure-config.md            # 6 config checks
    ├── services-api-contracts.md           # 4 API checks
    ├── services-event-contracts.md         # 4 event checks
    ├── services-splitz.md                  # 10 Splitz checks (6 require MCP)
    ├── services-stork.md                   # 12 Stork checks
    ├── services-passport.md                # 10 Passport checks
    ├── services-asv.md                     # 10 ASV checks
    ├── services-router.md                  # 10 Router SDK checks
    ├── pgrouter-service.md                 # 10 PG-Router service checks
    ├── domain-constraints.md               # 4 constraint checks
    ├── domain-flows.md                     # 4 flow checks
    ├── domain-rules.md                     # 2 rule checks
    ├── quality-unit-tests.md               # 4 test checks
    ├── quality-integration-tests.md        # 2 SLIT checks
    ├── quality-feature-flags.md            # 1 Splitz check
    └── quality-ci-integration.md           # 1 CI check
```

## Severity Levels

| Level | Icon | When | Action |
|-------|------|------|--------|
| Critical | 🚨 | Incident risk, data corruption, security | Block merge |
| High | ⚠️ | Data loss, performance issue, contract break | Require fix |
| Medium | 📋 | Tech debt, maintainability | Suggest fix |
| Low | ℹ️ | Best practice, minor optimization | Optional |

## Integration with Other Skills

### pr-ci-fixer

```
Pre-mortem detects CI failure
  ↓
Offers to investigate
  ↓
Invokes pr-ci-fixer
  ↓
Fixes failing tests or retriggers flaky ones
  ↓
Re-runs pre-mortem to validate
```

### pr-creator

```
Pre-mortem finds issues
  ↓
User approves auto-fix
  ↓
Applies fixes
  ↓
Uses pr-creator to commit and push
```

### Repo Skills

```
Auto-loads domain docs:
  .claude/skills/*-skill/modules/domain/{entity}/
    - constraints.md
    - flows.md
    - decisions.md
    - integration.md
```

## Supported Repositories

Works best with:
- **terminals** - Full repo skill with domain docs
- **payments-upi** - Complete code review patterns
- **pg-router** - Domain documentation
- **Other Go services** - Generic infrastructure checks

## Configuration Patterns Detected

Automatically handles:
- **Pattern 1**: `default.toml` → `stage-live.toml`, `prod-live.toml`
- **Pattern 2**: `default-live.toml`, `default-test.toml` → env configs
- **Pattern 3**: `default.toml` → `stage.toml`, `prod.toml`

## Examples

### Example 1: New Feature Without Tests

```
Input: PR adds internal/services/payment.go (250 lines)

Output:
🚨 Critical: New code without tests
  File: internal/services/payment.go
  Issue: No corresponding payment_test.go file
  Missing: Unit tests with 80%+ coverage
```

### Example 2: Config Change

```
Input: PR adds key to configs/default.toml

Output:
🚨 Critical: Missing keys in production configs
  Added: newFeature.enabled = true
  Missing from:
    - configs/prod-live.toml
    - configs/prod-test.toml
  Risk: Production uses dev default (enabled=true)!
```

### Example 3: Database Changes

```
Input: PR modifies internal/terminals/repo.go

Output:
🚨 Critical: Missing transaction rollback
  Line 422: if err != nil { return err }
  Issue: No rollback on error path
  Fix: Add "defer tx.Rollback()" after db.Begin()
```

### Example 4: Domain Constraint Violation

```
Input: PR modifies internal/gateway_credentials/service.go

Output:
🚨 Critical: Unique constraint not validated
  Constraint: UNIQUE(gateway, org_id, acquirer)
  From skill: constraints.md:19
  Missing: Duplicate check before Save()
```

## Limitations

**Currently checks:**
- Go services (terminals, payments-upi, pg-router)
- TOML configs
- Kafka/SQS messaging
- PostgreSQL/MySQL databases

**Not yet supported:**
- Python/Node.js services (coming soon)
- gRPC proto validation
- Kubernetes manifests
- Docker configs

## Development

### Adding New Checks

1. Create reference file: `references/{category}-{topic}.md`
2. Follow template:
   ```markdown
   # {Topic} Checks

   ## Check 1: {Name} {Severity}
   ### What to Check
   ### Bad Pattern ❌
   ### Good Pattern ✅
   ### Detection Strategy
   ### Flag Conditions
   ### Severity
   ```
3. Update SKILL.md file mapping
4. Test on sample PRs

### Testing

```bash
# Test on sample PR
cd /path/to/terminals
gh pr view 123
# Run skill manually to verify checks
```

## Credits

Built from analysis of:
- **terminals** service (56+ infrastructure patterns)
- **payments-upi** code review skill (12 quality checks)
- **pg-router** domain documentation (10+ business rules)

## License

Internal Razorpay tool. See company guidelines for usage.

## Support

For issues or enhancements:
1. Check existing code-review skills in `development/skills/`
2. Reference payments-upi-code-review skill patterns
3. Consult #claude-skills Slack channel
