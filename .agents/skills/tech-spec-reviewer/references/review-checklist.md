# Tech Spec Review Checklist

This document provides a comprehensive checklist for reviewing technical specifications across different dimensions.

## 1. Alternative Approaches & Trade-offs

### Questions to Ask
- Are multiple solution approaches documented with clear trade-off analysis?
- Are the pros and cons of each approach specific and measurable?
- Is the recommended approach clearly justified with reasoning?
- Are there obvious alternative solutions that weren't considered?
- Are architectural alternatives explored (microservices vs monolith, sync vs async, etc.)?
- Are vendor/technology alternatives evaluated?

### Red Flags
- Only one approach mentioned without considering alternatives
- Generic pros/cons like "easier" or "faster" without specifics
- No clear decision criteria for choosing the final approach
- Missing comparison on key dimensions (cost, complexity, maintainability)

### Good Practice
```
Approach A: New Service (Recommended)
Pros:
- Clear ownership (Team X)
- Supports 10K TPS (measured in load tests)
- Can scale independently
Cons:
- 3 weeks additional development time
- $500/month infrastructure cost
- Requires new monitoring setup

Approach B: Extend Existing Service
Pros:
- 1 week development time
- No additional infra cost
Cons:
- Shared database becomes bottleneck at 5K TPS
- Tight coupling with legacy code
- Cross-team dependency for deployments
```

## 2. Edge Cases & Failure Scenarios

### How to Explain Edge Cases to Developers

When identifying edge cases, **provide detailed scenarios** so developers can understand the failure path and make informed decisions. Use this format:

**Template for Edge Case Explanation:**
```
Edge Case: [Name of the edge case]
Scenario: [Step-by-step description of what happens]
Example: [Concrete example with data/state]
Impact: [What goes wrong - user experience, data corruption, etc.]
Recommendation: [How to handle it - code snippet, design change, validation]
```

**Example - Concurrent Reserve for Last Item:**
```
Edge Case: Race condition when multiple users reserve last available item
Scenario:
  1. Inventory has 1 item available (item_id: 123, state: CREATED)
  2. User A requests reserve at t=0ms
  3. User B requests reserve at t=5ms (before User A commits)
  4. Both queries: SELECT * FROM items WHERE state='CREATED' LIMIT 1
  5. Both see item_id: 123
  6. Both try to UPDATE item SET state='RESERVED' WHERE id=123

Impact: Without row locking, both succeed → overselling (2 reservations, 1 item)

Recommendation:
  Use row-level locking with SKIP LOCKED:
  ```sql
  SELECT * FROM items
  WHERE state='CREATED' AND reward_variant_id=?
  LIMIT 1
  FOR UPDATE SKIP LOCKED
  ```
  This ensures User A locks item_id: 123, User B gets 0 items (no overselling)
```

**Example - Idempotency Without Proper Key:**
```
Edge Case: Same correlation_id used for different variants
Scenario:
  1. Client calls reserve(correlation_id='order_123', variant='GIFT_CARD_A')
  2. Server creates transaction txn_1, reserves item
  3. Client retries with reserve(correlation_id='order_123', variant='GIFT_CARD_B')
     (maybe buggy client or malicious request)
  4. If idempotency only checks correlation_id, returns txn_1 (wrong variant!)

Impact: User gets wrong reward variant, potential security issue

Recommendation:
  Idempotency key should be composite: (correlation_id + reward_variant_id + user_id)
  ```go
  idempotencyKey := fmt.Sprintf("%s:%s:%s", correlationID, variantID, userID)
  if existingTxn := repo.FindByIdempotencyKey(idempotencyKey); existingTxn != nil {
    return existingTxn
  }
  ```
```

### Critical Edge Cases to Check

#### Data-Related
- Empty input / null values
- Duplicate requests (idempotency)
- Concurrent updates to same resource
- Data size limits (pagination, max payload)
- Invalid or malformed data
- Character encoding issues (unicode, special characters)
- Time zone and date boundary cases
- Numeric overflow/underflow

#### State & Timing
- Requests arriving out of order
- Partial failures (some operations succeed, others fail)
- Retry scenarios and duplicate processing
- State transitions that shouldn't occur
- Race conditions between concurrent operations
- Timeout scenarios at each integration point

#### System Failures
- Downstream service unavailable
- Upstream service timeout
- Database connection failures
- Network partitions
- Partial system degradation
- Cache invalidation failures
- Message queue backlogs

#### Business Logic
- Boundary values (0, negative, maximum limits)
- Invalid state transitions
- Insufficient balance/quota/permissions
- Expired tokens/sessions
- Missing required dependencies
- Conflicting operations

### Migration-Specific Edge Cases
- Data in old system while migration in progress
- Rollback during partial migration
- Dual-write inconsistencies
- Schema version mismatches
- Data loss during migration
- Orphaned records after migration

### Questions to Ask
- What happens when each external dependency fails?
- How does the system handle concurrent requests to the same resource?
- What if the operation is retried? Is it idempotent?
- Are there time-based edge cases (month boundaries, leap years, DST)?
- What are the data volume limits? What happens when exceeded?

## 3. Optimizations

### How to Explain Optimizations to Developers

When recommending optimizations, **provide before/after scenarios** with measurable impact so developers can prioritize and make informed decisions. Use this format:

**Template for Optimization Explanation:**
```
Optimization: [Name of the optimization]
Current Approach: [What the spec proposes]
Problem: [Why it's suboptimal - performance, cost, complexity]
Proposed Approach: [Alternative approach]
Impact: [Measurable improvement - latency, cost, throughput]
Trade-offs: [What do we give up? Complexity, development time, etc.]
Recommendation: [When to apply - always, only if X, not worth it]
```

**Example - Database Index Optimization:**
```
Optimization: Composite index for item allocation query
Current Approach:
  Query: SELECT * FROM items WHERE reward_variant_id=? AND state='CREATED'
         AND is_active=true AND expires_at > NOW() LIMIT 10 FOR UPDATE
  Indexes: Single index on reward_variant_id

Problem:
  - Database does index scan on reward_variant_id, then filters state/is_active/expires_at (slow)
  - For variant with 100K items (80K CLAIMED, 20K CREATED), scans 100K rows to find 20K CREATED
  - Query time: ~200ms at 100K items, grows to ~2s at 1M items

Proposed Approach:
  Add composite index:
  CREATE INDEX idx_items_allocation ON items(reward_variant_id, state, is_active, expires_at)
  WHERE deleted_at IS NULL;

Impact:
  - Query time: ~5ms at 100K items, ~15ms at 1M items (40x improvement)
  - Supports 200 TPS vs 5 TPS previously

Trade-offs:
  - Additional 100MB disk space per 1M items
  - Slightly slower writes (index maintenance) - negligible for read-heavy workload

Recommendation: MUST implement - critical for performance at scale
```

**Example - Caching Optimization:**
```
Optimization: Cache reward variant metadata
Current Approach:
  Every claim/reserve fetches reward_variant from DB to check is_active
  SELECT is_active, expires_at FROM reward_variants WHERE id=?

Problem:
  - 1000 reserve TPS = 1000 DB queries/sec just for metadata
  - Adds 10-20ms latency per request
  - Metadata changes rarely (updated few times per day)

Proposed Approach:
  Cache reward_variant metadata in Redis with 5-minute TTL
  Cache key: "variant:{variant_id}"
  Cache miss: Read from DB, populate cache

Impact:
  - Latency: Reduce p95 from 150ms to 130ms (20ms improvement)
  - DB load: Reduce by 95% (1000 QPS → 50 QPS for cache misses)
  - Cost: ~$20/month for Redis vs ~$100/month for larger DB instance

Trade-offs:
  - Eventual consistency: Changes take up to 5 min to propagate
  - Additional dependency (Redis) - need monitoring, failover
  - Cache invalidation complexity if variant updated

Recommendation:
  - Implement if TPS > 100 (measurable benefit)
  - Skip if TPS < 100 (premature optimization, adds complexity)
  - For KTLO product with low traffic, SKIP
```

**Example - N+1 Query Optimization:**
```
Optimization: Eager load transaction line items
Current Approach:
  1. Query: Fetch transaction: SELECT * FROM transactions WHERE id=?
  2. For each transaction, query line items: SELECT * FROM line_items WHERE txn_id=?
  3. For claim with 10 items: 1 + 10 = 11 queries

Problem:
  - N+1 query pattern (1 query for txn + N queries for items)
  - For batch claim of 50 items: 51 database round trips
  - Latency: 51 × 5ms = 255ms just for queries

Proposed Approach:
  Use JOIN or batch query:
  ```sql
  SELECT t.*, li.* FROM transactions t
  LEFT JOIN line_items li ON t.id = li.txn_id
  WHERE t.id = ?
  ```
  Or use ORM preload: `db.Preload("LineItems").Find(&transaction)`

Impact:
  - Queries: 51 → 1 (98% reduction)
  - Latency: 255ms → 5ms for queries (50x improvement)
  - Total claim latency: 300ms → 50ms

Trade-offs:
  - Slightly more complex query
  - May fetch more data than needed (all fields)

Recommendation: MUST implement - significant performance improvement with minimal cost
```

**Example - Cost Optimization:**
```
Optimization: Data retention policy for old transactions
Current Approach:
  Keep all transactions forever (soft delete only)

Problem:
  - Database grows indefinitely: 10M txn/year × 5 years = 50M rows
  - Storage cost: 50M rows × 2KB = 100GB = $50/month (and growing)
  - Query performance degrades over time (even with indexes)

Proposed Approach:
  Archive transactions older than 2 years to cold storage (S3/Glacier)
  - Keep last 2 years in DB: 20M rows = 40GB = $20/month
  - Archive to S3: 60M rows = 120GB = $3/month (1/15th cost)
  - Provide separate API for archived data retrieval (slower, rarely used)

Impact:
  - Cost: $50/month → $23/month (54% reduction)
  - At 5 years: $250/month → $35/month (86% reduction)
  - Query performance: Improved due to smaller table

Trade-offs:
  - Archived data retrieval slower (acceptable if rare)
  - Need archival job and separate retrieval logic
  - Implementation effort: ~1 week

Recommendation:
  - Implement if transaction volume > 5M/year
  - Skip for low-volume systems (not worth complexity)
```

### Performance Optimizations
- **Caching**: What can be cached? Cache invalidation strategy?
- **Batch Processing**: Can operations be batched?
- **Async Processing**: Can operations be made asynchronous?
- **Database**: Proper indexing? N+1 query issues? Connection pooling?
- **API Calls**: Can multiple calls be combined? GraphQL instead of REST?
- **Lazy Loading**: Can data loading be deferred?

### Scalability Optimizations
- **Horizontal Scaling**: Can the service scale out?
- **Load Balancing**: Is load evenly distributed?
- **Sharding**: Is data partitioned effectively?
- **Rate Limiting**: Prevent resource exhaustion?
- **Backpressure**: Handle downstream slowness?

### Cost Optimizations
- **Resource Utilization**: Right-sized instances/containers?
- **Storage**: Data retention policies? Archival strategy?
- **Network**: Minimize cross-region data transfer?
- **Compute**: Spot instances? Auto-scaling policies?
- **Caching**: Reduce expensive operations?

### Code/Architecture Optimizations
- **Reduce Complexity**: Can the design be simplified?
- **Reuse Existing**: Leverage existing services/libraries?
- **Decouple**: Reduce tight coupling between components?
- **Standardize**: Use common patterns/frameworks?

### Questions to Ask
- What are the most expensive operations? Can they be optimized?
- Are there obvious bottlenecks in the design?
- Can we leverage existing infrastructure/services?
- Is the solution over-engineered for the problem?

## 4. Testing Strategy

### Test Coverage Requirements

#### Unit Tests
- Core business logic functions
- Edge cases and boundary conditions
- Error handling paths
- Input validation

#### Integration Tests
- API contract tests
- Database integration
- External service integration (with mocks)
- Message queue integration

#### End-to-End Tests
- Critical user journeys
- Multi-service workflows
- Error recovery flows

#### Performance Tests
- Load testing (expected TPS)
- Stress testing (beyond capacity)
- Soak testing (sustained load)
- Spike testing (sudden traffic increase)

#### Other Testing
- Security testing (penetration, vulnerability scanning)
- Chaos engineering (failure injection)
- A/B testing strategy
- Backward compatibility testing
- Migration testing (dual-write verification)

### Questions to Ask
- Are test coverage targets specified? (e.g., 80% code coverage)
- Are critical paths explicitly listed for testing?
- Is there a load testing plan with specific TPS targets?
- How will migration be tested without affecting production?
- Are rollback scenarios tested?
- Is there a UAT (User Acceptance Testing) plan?
- Are there automated regression tests?

### Red Flags
- "To be added" for testing sections
- No specific test scenarios listed
- No performance/load testing plan
- No rollback testing mentioned
- Missing integration testing strategy

## 5. Observability & Metrics

### Required Metrics

#### Health Metrics
- Service availability/uptime (%)
- Request rate (requests/second)
- Error rate (errors/total requests)
- Latency (p50, p90, p95, p99)

#### Business Metrics
- Transaction success rate
- Transaction volume
- Revenue impact (if applicable)
- User-facing errors

#### System Metrics
- CPU/Memory utilization
- Database connection pool usage
- Queue depth/lag
- Cache hit ratio
- External API call latency

#### Custom Metrics
- Domain-specific KPIs
- Feature usage metrics
- Data quality metrics

### Logging Requirements
- Structured logging format (JSON)
- Correlation IDs for request tracing
- Log levels (DEBUG, INFO, WARN, ERROR)
- PII/sensitive data masking
- Log retention policy

### Alerting Strategy
- Critical alerts (paging required)
- Warning alerts (investigate during business hours)
- SLO-based alerts
- Alert fatigue prevention
- Escalation policy

### Dashboards
- Service health dashboard
- Business metrics dashboard
- Error tracking dashboard
- Real-time monitoring dashboard

### Questions to Ask
- What metrics will be tracked? Are they specific and measurable?
- What are the alert thresholds? When does someone get paged?
- Is there distributed tracing for debugging?
- How will we detect issues before users report them?
- Are there runbooks for common issues?
- Is there a monitoring strategy during migration?

### Red Flags
- "To be added" for monitoring section
- No specific metrics listed
- No alerting strategy
- No dashboards planned
- No correlation IDs for distributed tracing

## 6. Non-Functional Requirements (NFRs)

### Performance
- **TPS (Transactions Per Second)**: Specific target (e.g., "Handle 10,000 TPS")
- **Latency**: p50, p95, p99 targets (e.g., "p95 < 200ms")
- **Throughput**: Data processing rate
- **Concurrent Users**: How many simultaneous users?

### Scalability
- **Horizontal Scaling**: Can add more instances?
- **Vertical Scaling**: Can increase instance size?
- **Auto-scaling**: Triggers and policies
- **Growth Projection**: Handle 5x traffic in 12 months?

### Availability
- **Uptime SLA**: Specific target (e.g., "99.9% uptime")
- **Downtime Budget**: Planned maintenance windows
- **Multi-region**: Active-active or active-passive?
- **Disaster Recovery**: RPO (Recovery Point Objective) and RTO (Recovery Time Objective)

### Reliability
- **Data Durability**: No data loss guarantees
- **Consistency**: Strong vs eventual consistency
- **Idempotency**: Retry-safe operations
- **Fault Tolerance**: Handle component failures

### Security
- **Authentication**: How are users/services authenticated?
- **Authorization**: Role-based access control (RBAC)?
- **Encryption**: Data at rest and in transit
- **Secrets Management**: How are credentials stored?
- **Audit Logging**: Who did what, when?
- **Data Privacy**: PII handling, GDPR compliance
- **Input Validation**: Prevent injection attacks
- **Rate Limiting**: Prevent abuse

### Compliance
- **Regulatory Requirements**: PCI-DSS, SOC 2, GDPR, etc.
- **Data Residency**: Where data must be stored
- **Audit Trail**: Compliance reporting requirements
- **Data Retention**: How long to keep data

### Maintainability
- **Code Quality**: Linting, code review standards
- **Documentation**: API docs, architecture docs, runbooks
- **Deployment**: CI/CD pipeline, deployment frequency
- **Backward Compatibility**: API versioning strategy

### Cost
- **Infrastructure Cost**: Monthly/annual estimates
- **Operational Cost**: Support, monitoring, maintenance
- **Cost per Transaction**: Unit economics
- **Budget Constraints**: Maximum acceptable cost

### Questions to Ask
- Are NFRs specific and measurable? (avoid "high availability" - specify "99.95%")
- Are SLAs defined for critical operations?
- Is there a security review checklist?
- Are compliance requirements identified?
- Is there a cost estimate with breakdown?
- Are there capacity planning projections?

### Red Flags
- NFR sections marked "NA" without justification
- Generic terms without specific numbers (e.g., "highly available")
- No security considerations
- No performance targets (TPS, latency)
- No availability/uptime targets
- Missing cost estimates

## 7. Data Consistency & Correctness

### Consistency Guarantees
- **ACID Properties**: For database transactions
- **Eventual Consistency**: For distributed systems
- **Idempotency**: Same request produces same result
- **Duplicate Prevention**: Unique constraints, deduplication
- **Referential Integrity**: Foreign key constraints, data relationships

### Data Migration
- **Backfill Strategy**: How existing data will be migrated
- **Dual-Write Verification**: Compare old vs new system
- **Data Validation**: Checksums, row counts, data quality checks
- **Zero Data Loss**: Guarantee no data is lost during migration
- **Rollback Safety**: Can revert without data loss

### Concurrency Control
- **Optimistic Locking**: Version numbers, timestamps
- **Pessimistic Locking**: Row/table locks
- **Distributed Locks**: For cross-service coordination
- **Race Condition Handling**: Concurrent updates to same resource

### Questions to Ask
- How is data consistency guaranteed?
- What happens if dual-write fails partially?
- Are there database transactions wrapping critical operations?
- How are race conditions prevented?
- Is the migration strategy zero-downtime?
- How do we verify data integrity after migration?

## 8. Error Handling & Resilience

### Error Handling Strategy
- **Graceful Degradation**: System remains partially functional
- **Circuit Breakers**: Prevent cascading failures
- **Retry Logic**: Exponential backoff with jitter
- **Timeout Configuration**: At each integration point
- **Fallback Mechanisms**: Cached data, default values
- **Dead Letter Queues**: For failed message processing

### Error Communication
- **User-Facing Errors**: Clear, actionable messages
- **Error Codes**: Standardized error taxonomy
- **Error Logging**: Sufficient context for debugging
- **Error Monitoring**: Track error rates and types

### Recovery Mechanisms
- **Automatic Recovery**: Self-healing systems
- **Manual Recovery**: Runbooks for common failures
- **Data Recovery**: Backups, point-in-time recovery
- **State Recovery**: Resume from failure point

### Questions to Ask
- How are errors surfaced to users vs logged internally?
- Are there retry mechanisms? With what strategy?
- What's the timeout for each external call?
- How does the system handle partial failures?
- Is there a circuit breaker pattern implemented?
- Are errors monitored and alerted on?

## 9. Rollout & Rollback Plan

### Deployment Strategy
- **Canary Deployment**: Gradual rollout to subset of users
- **Blue-Green Deployment**: Switch between environments
- **Feature Flags**: Enable/disable features without deployment
- **Ramp Plan**: 1% → 10% → 50% → 100%

### Rollback Requirements
- **Rollback Trigger**: When to rollback (error rate > X%, latency > Yms)
- **Rollback Procedure**: Specific steps to revert
- **Rollback Time**: Target time to complete rollback (e.g., < 15 minutes)
- **Data Rollback**: How to handle data written in new format
- **Zero-Downtime Rollback**: Can rollback without service disruption

### Migration Specifics
- **Downtime Window**: If required, how long?
- **Communication Plan**: Notify affected users/teams
- **Verification Steps**: How to confirm migration success
- **Merchant/User Migration**: One-by-one or batch?

### Questions to Ask
- Is there a specific rollout plan with percentages and timelines?
- What are the rollback triggers and procedures?
- Can we rollback without data loss?
- Is backward compatibility maintained during rollout?
- Are there automated health checks during rollout?
- How long does rollback take?

### Red Flags
- No rollback plan mentioned
- "Big bang" deployment without phased rollout
- No rollback triggers defined
- No backward compatibility strategy
- Downtime not quantified or justified

## 10. Dependencies & Integration

### Dependency Analysis
- **Upstream Dependencies**: Services this depends on
- **Downstream Dependencies**: Services that depend on this
- **Third-Party Services**: External APIs, SaaS
- **Shared Resources**: Databases, caches, queues

### Integration Contracts
- **API Contracts**: Request/response formats
- **SLAs**: Expected availability and performance of dependencies
- **Rate Limits**: Throttling on external services
- **Authentication**: How to authenticate with dependencies

### Failure Handling
- **Dependency Failure**: What happens when dependency is down?
- **Timeout Strategy**: Timeout for each dependency call
- **Retry Strategy**: When and how to retry failed calls
- **Fallback**: Alternative when dependency unavailable

### Questions to Ask
- Are all dependencies clearly identified with owners?
- What are the SLAs for each dependency?
- How does the system behave when dependencies fail?
- Are there circular dependencies?
- Are rate limits of dependencies considered?
- Is there a dependency health check?
