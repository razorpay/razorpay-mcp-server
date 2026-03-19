# Edge Cases Catalog by Domain

This document catalogs common edge cases organized by system domain to help reviewers identify missing scenarios in tech specs.

**How to Use This Catalog:**
When reviewing a tech spec, identify the relevant domain(s) and check if the spec addresses these edge cases. For each missing edge case, provide a detailed scenario (see review-checklist.md for example format) explaining:
- The specific sequence of events that triggers the issue
- Concrete example data/state
- User impact or system behavior
- Recommended mitigation

**Example Pattern:**
```
Edge Case from Catalog: "Concurrent reserve for last N items"

In Review, Explain Like This:
---
🔴 CRITICAL: Concurrent Reserve for Last Item Not Addressed
Scenario:
  1. Inventory: 1 item available (item_id: 123, state: CREATED)
  2. User A: reserve(variant_id=X, qty=1) at t=0ms
  3. User B: reserve(variant_id=X, qty=1) at t=5ms
  4. Both: SELECT * FROM items WHERE variant_id=X AND state='CREATED' LIMIT 1
  5. Both see item_id: 123 (no lock yet)
  6. Both: UPDATE items SET state='RESERVED' WHERE id=123
  7. Result: Both succeed → 2 reservations, 1 item (overselling)

Impact: Overselling inventory - deliver voucher to 2 users, only have 1 code
Recommendation:
  ```sql
  SELECT * FROM items WHERE variant_id=? AND state='CREATED'
  LIMIT 1 FOR UPDATE SKIP LOCKED
  ```
  User B gets 0 items (fails fast), User A succeeds
---
```

## Payment Systems

### Transaction Processing
- **Insufficient balance during payment**
  - *Example*: User has ₹100 balance, tries to pay ₹150. Check balance before payment, but balance consumed by concurrent transaction during payment processing.
  - *Mitigation*: Use database transaction with SELECT FOR UPDATE on balance, or optimistic locking with version numbers.

- **Payment timeout during processing**
  - *Example*: Payment gateway call takes 35s (exceeds 30s timeout). Gateway actually processed payment but service marked as failed due to timeout.
  - *Mitigation*: Implement idempotent payment check after timeout - query gateway for payment status before marking failed.

- **Duplicate payment requests (same idempotency key)**
  - *Example*: User clicks "Pay" twice rapidly (double-click). Both requests have same idempotency_key='order_12345'.
  - *Expected*: Return same payment_id for both requests (idempotent).
  - *Edge*: Second request arrives while first is PROCESSING (not committed yet) - return 409 Conflict or wait?

- **Concurrent payment attempts from same user**
  - *Example*: User has 2 browser tabs, attempts payment in both tabs for different orders. Both check balance (₹100), both try to deduct ₹80.
  - *Impact*: User balance goes negative, or second payment fails incorrectly.
  - *Mitigation*: Row-level lock on user balance, or queue payments per user.

- Payment amount of 0 or negative
- Payment exceeding maximum limit
- Partial refund exceeding available amount
- Refund on already refunded transaction
- Payment state changes after user abandonment
- Currency mismatch between payment and account

### Settlement & Reconciliation
- Settlement batch processing failure midway
- Duplicate settlement entries
- Settlement amount mismatch with transactions
- Failed settlement retry logic
- Settlement during month/year boundary
- Missing transactions in settlement batch
- Settlement to closed/invalid bank account

### Wallet Operations
- Transfer when wallet balance is 0
- Concurrent transfers depleting balance
- Transfer to non-existent customer
- Transfer amount exceeding daily/monthly limits
- Wallet balance going negative
- Credit after debit failure (balance inconsistency)
- Wallet locked/frozen during operation

## Distributed Systems

### Network & Communication
- **Request timeout at various stages**
  - *Example*:
    ```
    Client → API Gateway → Service A → Service B → Database
    Timeout at each layer: Client=60s, Gateway=30s, ServiceA=15s, ServiceB=10s

    Issue: ServiceA calls ServiceB with 15s timeout, but ServiceB needs 20s
    Result: ServiceA times out, returns error, BUT ServiceB completes successfully
    Impact: Client sees failure, database has committed transaction → inconsistency
    ```
  - *Mitigation*: Timeout chain should decrease at each layer, idempotent operations, check status after timeout.

- **Connection pool exhaustion**
  - *Example*:
    ```
    Service has 20 DB connections in pool
    Traffic spike: 100 concurrent requests
    First 20 requests: Get DB connections, start processing
    Next 80 requests: Wait for connection (queue)
    If queries are slow (10s each): 80 requests timeout waiting for connection
    ```
  - *Impact*: Service appears down (all requests timeout) even though it's healthy.
  - *Mitigation*: Circuit breaker, connection pool monitoring, alert on pool utilization > 80%.

- Network partition between services
- Service discovery failure
- DNS resolution failure
- SSL certificate expiration
- Message queue backlog/overflow

### Data Consistency
- **Dual-write partial failure (write to system A succeeds, B fails)**
  - *Example*:
    ```
    Migration phase: Write to both old and new inventory systems

    1. Claim inventory in OLD system → Success (item marked CLAIMED)
    2. Claim inventory in NEW system → Failure (network timeout)
    3. User sees error, but item actually claimed in OLD system
    4. Retry: OLD system says "already claimed", NEW system has no record

    Result: Data divergence - item CLAIMED in old, CREATED in new
    ```
  - *Mitigation*:
    - Write to NEW first, then OLD (so rollback leaves consistent state)
    - Log dual-write failures for manual reconciliation
    - Background job to detect and fix divergence

- **Read after write inconsistency**
  - *Example*:
    ```
    User: Update profile (name="Alice") → writes to primary DB
    System: Redirect to profile page → reads from replica DB
    Replication lag: 500ms
    Result: User sees old name ("Bob") for 500ms after update
    ```
  - *Impact*: User confused, thinks update failed, retries.
  - *Mitigation*: Read from primary after write, or use session stickiness, or show "updating..." state.

- **Clock skew between servers**
  - *Example*:
    ```
    Server A (clock: 10:00:00): Creates item with expires_at=10:15:00
    Server B (clock: 10:05:00 due to skew): Checks if item expired
    Server B sees: current_time(10:05:00) < expires_at(10:15:00) → valid
    But actually: Server B's clock is 5min fast, item should be expired
    ```
  - *Mitigation*: Use NTP sync, use monotonic clocks, or store TTL (duration) instead of absolute timestamp.

- Out-of-order message delivery
- Duplicate message consumption
- Lost messages in queue
- Stale cache after database update
- Split brain scenario in distributed system

### State Management
- Distributed transaction rollback
- Orphaned resources after failure
- Inconsistent state across replicas
- Zombie processes after crash
- Lost locks after process crash
- Two-phase commit failure scenarios

## Database Systems

### Concurrency
- **Deadlock between transactions**
  - *Example*:
    ```
    Txn A: Lock item_1, then try to lock item_2
    Txn B: Lock item_2, then try to lock item_1
    Result: Deadlock - both wait forever
    ```
  - *Common scenario*: Two concurrent claim requests allocating items in different order.
  - *Mitigation*: Always acquire locks in consistent order (e.g., ORDER BY id ASC), or use SKIP LOCKED.

- **Lost update problem**
  - *Example*:
    ```
    T=0: User A reads balance = ₹100
    T=1: User B reads balance = ₹100
    T=2: User A deducts ₹50, writes balance = ₹50 (100-50)
    T=3: User B deducts ₹30, writes balance = ₹70 (100-30)
    Result: Balance should be ₹20, but is ₹70 (lost A's update)
    ```
  - *Mitigation*: Use UPDATE SET balance = balance - 50 WHERE id=? (atomic), or optimistic locking.

- **Optimistic locking version mismatch**
  - *Example*:
    ```
    User A: Read item (version=5)
    User B: Read item (version=5)
    User A: Update item SET state='CLAIMED', version=6 WHERE id=? AND version=5 → Success
    User B: Update item SET state='CANCELLED', version=6 WHERE id=? AND version=5 → 0 rows affected (version now 6)
    ```
  - *Expected*: User B's update fails (version mismatch), retry with latest version.
  - *Edge*: Spec doesn't handle version mismatch → silent failure or incorrect error?

- Dirty read scenarios
- Phantom read issues
- Write-write conflict
- Pessimistic lock timeout
- Lock escalation affecting performance

### Data Integrity
- Foreign key constraint violation
- Unique constraint violation on retry
- Null value in NOT NULL column
- Data type overflow (integer, varchar length)
- Character encoding issues (emoji, special chars)
- Timezone inconsistencies
- Floating point precision errors
- JSON/JSONB parsing errors

### Operations
- Database connection pool exhaustion
- Long-running query blocking others
- Index lock during heavy writes
- Replication lag affecting reads
- Disk space exhaustion
- Table/database size limits
- Query timeout
- Transaction isolation level issues

## API & HTTP

### Request Handling
- Missing required headers
- Invalid authentication token
- Expired JWT token
- Malformed JSON payload
- Content-Type mismatch
- Request body size exceeding limit
- Invalid URL parameters
- Special characters in URL
- Request rate limiting triggered

### Response Handling
- 5xx errors from downstream
- Unexpected response format
- Response timeout
- Partial response (connection closed)
- HTTP 429 (Too Many Requests)
- Redirect loops
- Response size exceeding client buffer

### Caching
- Cache stampede (many requests miss cache simultaneously)
- Stale data served from cache
- Cache invalidation race condition
- Cache key collision
- Cache memory exhaustion
- Negative caching of errors

## Async Processing & Queues

### Message Processing
- Message processing failure after dequeue
- Poison message causing repeated failures
- Message ordering violation
- Duplicate message consumption
- Message expiration/TTL
- Queue full/overflow
- Consumer lag exceeding threshold
- Message size exceeding limit

### Job Processing
- Job timeout during execution
- Job retry exhausting max attempts
- Job stuck in processing state
- Concurrent job execution when not intended
- Job dependency failure
- Job scheduler failure
- Cron job overlap (previous run not finished)

### Event Streaming
- Event replay scenarios
- Event out of order
- Event duplication
- Consumer group rebalancing
- Offset commit failure
- Partition unavailability
- Event schema evolution issues

## Data Migration

### During Migration
- Migration script failure midway
- Data validation failure during migration
- Source data changed during migration
- Target schema change during migration
- Network interruption during bulk transfer
- Foreign key constraint violation
- Unique constraint violation
- Data type conversion errors

### Dual Write Period
- Write to old system succeeds, new fails
- Write to new system succeeds, old fails
- Data divergence between systems
- Race condition in dual write
- Duplicate entry detection
- Consistency verification failure

### Cutover & Rollback
- Cutover during active transaction
- Partial rollback leaving inconsistent state
- Rollback when new data format incompatible
- Rollback with data loss
- Cutover trigger failure
- Traffic not fully switched

## Authentication & Authorization

### Authentication
- Token expiration during long operation
- Token refresh failure
- Invalid credentials
- Account locked after failed attempts
- Multi-factor authentication timeout
- Session hijacking
- Concurrent login from different devices
- Password reset token expired/invalid

### Authorization
- Permission changed mid-request
- Role/permission cache stale
- Insufficient permissions
- Resource ownership verification failure
- Cross-tenant data access attempt
- Privilege escalation attempt

## File Operations

### Upload
- File size exceeding limit
- Invalid file type
- Virus/malware in uploaded file
- File name with special characters
- Duplicate file name
- Upload timeout
- Incomplete upload
- Corrupted file

### Processing
- File encoding issues (UTF-8, ASCII)
- Unsupported file format
- File locked by another process
- Insufficient disk space
- File read/write permission denied
- File handle leak
- Binary vs text mode mismatch

## Time & Date

### Time-Based Logic
- Daylight Saving Time transitions
- Leap second handling
- Leap year calculation
- Month boundary (Feb 28/29)
- Year boundary (Dec 31 → Jan 1)
- Timezone conversion errors
- Clock skew between servers
- Time going backward (NTP adjustment)

### Scheduling
- Cron expression edge cases (Feb 30, etc.)
- Schedule overlap
- Missed schedule (system down during trigger)
- Schedule in past
- Infinite loop in retry schedule

## Rate Limiting & Throttling

### Limit Enforcement
- Burst traffic exceeding limit
- Limit reset boundary race condition
- Distributed rate limiting coordination
- Rate limit per user vs per IP vs per API key
- Rate limit bypass via different endpoints
- Rate limit counter overflow

## Batch Processing

### Batch Operations
- Empty batch
- Batch size exceeding memory limit
- Partial batch failure
- Batch timeout
- Duplicate items in batch
- Batch processing order dependency
- Batch retry logic
- Last item in batch handling

## Frontend/Mobile

### Network
- Offline mode
- Intermittent connectivity
- Request timeout on slow network
- Background app suspension mid-request
- App killed during operation

### State
- Stale data in UI
- Optimistic update rollback
- Cache invalidation on app update
- Local storage quota exceeded
- Session expiry while app in background

## Third-Party Integrations

### External Services
- Third-party API down
- Third-party API changed without notice
- Third-party rate limit hit
- API key expired/revoked
- Webhook delivery failure
- Webhook signature verification failure
- OAuth token refresh failure
- Third-party timeout
- Unexpected response format from third-party
- Third-party returning intermittent errors

## Resource Limits

### System Resources
- Memory exhaustion (OOM)
- CPU throttling
- Disk I/O bottleneck
- Network bandwidth saturation
- File descriptor limit reached
- Thread pool exhaustion
- Port exhaustion

### Application Limits
- Maximum connections exceeded
- Maximum concurrent requests
- Buffer overflow
- Stack overflow
- Heap overflow
- Collection size limits (array, map)

## Data Validation

### Input Validation
- SQL injection attempt
- XSS payload
- Command injection
- Path traversal attempt
- Integer overflow in calculation
- Division by zero
- Null/undefined value
- Empty string vs null
- Whitespace-only input
- Very long input string
- Nested object depth limit
- Circular references in JSON

## Compliance & Security

### Data Privacy
- PII in logs
- PII in error messages
- Unencrypted sensitive data in transit
- Unencrypted sensitive data at rest
- Data retention violation
- Right to deletion request
- Data export request
- Cross-border data transfer

### Security
- Brute force attack
- DDoS attack
- Man-in-the-middle attack
- Replay attack
- CSRF attack
- Session fixation
- Insecure deserialization
- Server-side request forgery (SSRF)

## Backward Compatibility

### API Changes
- Field removed that clients depend on
- Field type changed
- Required field added
- Enum value removed
- Error response format changed
- API endpoint deprecated
- Version mismatch between client and server

### Data Format Changes
- Schema evolution (add/remove/rename columns)
- Serialization format change
- Breaking changes in message format
- Protocol version incompatibility
