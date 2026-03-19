# DynamoDB Infrastructure Checks

## Overview

Validates AWS DynamoDB usage patterns in Razorpay services to prevent hot partitions, race conditions, data loss, and performance issues. Based on patterns from offers-engine, upi-switch, identity-provider, credstash-v3, and virtual-account services.

**Load when:** PR modifies DynamoDB table definitions or operations

**Total Checks:** 8

**Severity Distribution:**
- 🚨 Critical: 4
- ⚠️ High: 2
- 📋 Medium: 2

---

## Razorpay DynamoDB Context

### Services Using DynamoDB

**Found in production code:**
- `offers-engine` - Offers and rewards catalog (has .claude/skills with patterns)
- `upi-switch` - UPI delegate service
- `identity-provider` - User authentication tokens with TTL
- `credstash-v3` - Secrets storage (Kubestash backend)
- `qr-codes` - QR code data (migration mentioned in jargon-explainer)
- `virtual-account` - VA transaction patterns with batch operations
- `bin-service` - BIN data management
- `cms`, `dcs`, `payments-mandate` - Various use cases

### Common Patterns Found

**Batch operations** (offers-engine skill):
- BatchGetItem: 100 items limit
- BatchWriteItem: 25 items limit
- Retry handling for unprocessed items

**Conditional writes** (identity-provider, upi-switch, qr-codes):
- `ConditionExpression: "attribute_not_exists(PK)"` to prevent duplicates
- Version checking for optimistic locking

**TTL** (identity-provider):
- Session/token expiration
- `UpdateTimeToLiveInput` with `AttributeName: "ttl"`

---

## Check 1: Partition Key Design (High Cardinality) 🚨 CRITICAL

### What to Check

Partition keys must have high cardinality to prevent hot partitions and distribute load evenly.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Low cardinality partition key
type Offer struct {
    Status      string `dynamodbav:"status"`         // PARTITION KEY - BAD!
    OfferID     string `dynamodbav:"offer_id"`       // SORT KEY
    MerchantID  string `dynamodbav:"merchant_id"`
}

// Problem: All "active" offers go to same partition
// - 1 million active offers → 1 partition
// - 1000 inactive offers → different partition
// ❌ Hot partition! Throttling, slow queries

// ANTI-PATTERN: Sequential partition key
type Transaction struct {
    Date          string `dynamodbav:"date"`         // "2024-01-15" - PARTITION KEY
    TransactionID string `dynamodbav:"transaction_id"` // SORT KEY
}

// Problem: All transactions on same day → same partition
// January 15th: 1 million transactions → 1 hot partition
// ❌ Write throttling!

// ANTI-PATTERN: Single partition for all data
type Config struct {
    ConfigType string `dynamodbav:"config_type"` // Always "app_config" - PARTITION KEY
    Key        string `dynamodbav:"key"`         // SORT KEY
}

// Problem: Only 1 partition for entire table
// ❌ No distribution, no scalability
```

### Good Pattern ✅

```go
// CORRECT: High cardinality partition key (Razorpay pattern)
// Pattern from offers-engine, identity-provider, upi-switch

type Offer struct {
    OfferID     string `dynamodbav:"offer_id"`     // PARTITION KEY - Unique per offer ✅
    UpdatedAt   int64  `dynamodbav:"updated_at"`   // SORT KEY
    MerchantID  string `dynamodbav:"merchant_id"`
    Status      string `dynamodbav:"status"`
}

// Distributed: Each offer = unique partition key
// Queries: GetItem by offer_id (single partition)
// ✅ No hot partitions

// CORRECT: Composite key for distributed writes
type Transaction struct {
    MerchantID    string `dynamodbav:"merchant_id"`    // PARTITION KEY ✅
    TransactionID string `dynamodbav:"transaction_id"` // SORT KEY (timestamp+UUID)
    Date          string `dynamodbav:"date"`           // Attribute (for GSI)
}

// Distributed: Transactions spread across merchants
// Query by merchant: Single partition scan
// Query by date: Use GSI
// ✅ Distributed writes

// CORRECT: Sharded design for high-volume writes
type EventLog struct {
    ShardID   string `dynamodbav:"shard_id"`   // PARTITION KEY: "shard_0" to "shard_99"
    Timestamp int64  `dynamodbav:"timestamp"`  // SORT KEY
    EventData string `dynamodbav:"event_data"`
}

// Distributed: 100 shards = 100 partitions
// Write: Hash(timestamp) % 100 → shard_id
// Read: Query all 100 shards in parallel
// ✅ Massively parallel writes

// CORRECT: PK/SK pattern (identity-provider pattern)
type RefreshToken struct {
    PK         string `dynamodbav:"PK"`  // PARTITION KEY: "USER#<user_id>"
    SK         string `dynamodbav:"SK"`  // SORT KEY: "TOKEN#<token_id>"
    TTL        int64  `dynamodbav:"ttl"` // Auto-expire
    RevokedAt  *int64 `dynamodbav:"RevokedAt,omitempty"`
}

// Distributed: Each user = unique partition
// Query: All tokens for user (single partition)
// ✅ High cardinality (millions of users)
```

### Detection Strategy

```bash
# Find DynamoDB table definitions
grep -n "dynamodbav.*PARTITION\|HashKey\|PK" <pr_files>

# Check partition key type
grep -A 5 "PARTITION KEY\|HashKey" <pr_files>

# Flag low cardinality
grep -E "(status|type|category|date).*PARTITION" <pr_files>
```

### Flag Conditions

Flag if:
- Partition key is `status`, `type`, `category`, `date`, `country` (low cardinality)
- Partition key is constant (same value for all items)
- Timestamp-only partition key (sequential writes)
- Comment mentions "hot partition" or "throttling"

### Severity

🚨 **Critical** - Performance and scalability:
- Hot partitions → throttling
- ProvisionedThroughputExceededException
- Slow queries
- Cannot scale horizontally

### References

**Razorpay production code:**
- `offers-engine` - offer_id as partition key
- `identity-provider` - PK/SK pattern with user_id
- `upi-switch` - High cardinality keys

---

## Check 2: Conditional Writes (Prevent Duplicates & Race Conditions) 🚨 CRITICAL

### What to Check

PutItem and UpdateItem must use ConditionExpression to prevent race conditions and duplicates.

### Razorpay Pattern Found

**Consistent pattern across services:**
- `identity-provider`: `ConditionExpression: "attribute_not_exists(PK) and attribute_not_exists(SK)"`
- `goutils/kvstore/dynamo`: `ConditionExpression: "attribute_not_exists(%s)"`
- `qr-codes`: `ConditionExpression: "attribute_not_exists(PK)"`
- `offers-engine`: `ConditionExpression: "attribute_not_exists(contact_id)"`
- `upi-switch`: `ConditionExpression: "attribute_not_exists(#key)"`
- `dcs`: `ConditionExpression: "attribute_not_exists(#U) OR #U < :ts"` (version check)

### Bad Pattern ❌

```go
// ANTI-PATTERN: No duplicate prevention
func CreateOffer(ctx context.Context, offer *Offer) error {
    av, _ := attributevalue.MarshalMap(offer)

    putInput := &dynamodb.PutItemInput{
        TableName: aws.String("Offers"),
        Item:      av,
        // ❌ No ConditionExpression - overwrites existing!
    }

    _, err := client.PutItem(ctx, putInput)
    return err
}

// Problem:
// T1: CreateOffer(offer_123) → succeeds
// T2: CreateOffer(offer_123) → overwrites T1's data!
// ❌ Duplicate processing, data loss

// ANTI-PATTERN: No optimistic locking
func UpdateOfferStatus(ctx context.Context, offerID string, newStatus string) error {
    updateInput := &dynamodb.UpdateItemInput{
        TableName: aws.String("Offers"),
        Key: map[string]types.AttributeValue{
            "offer_id": &types.AttributeValueMemberS{Value: offerID},
        },
        UpdateExpression: aws.String("SET #status = :new_status"),
        // ❌ No version check!
    }

    _, err := client.UpdateItem(ctx, updateInput)
    return err
}

// Race condition:
// T1: Read version=5, prepare update
// T2: Update version=5 → version=6
// T1: Update (overwrites T2's changes!)
// ❌ Lost update!
```

### Good Pattern ✅

```go
// CORRECT: Prevent duplicates (identity-provider pattern)
func CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
    av, _ := attributevalue.MarshalMap(token)

    putInput := &dynamodb.PutItemInput{
        TableName: aws.String("RefreshTokens"),
        Item:      av,
        // ✅ Only create if doesn't exist
        ConditionExpression: aws.String("attribute_not_exists(PK) and attribute_not_exists(SK)"),
    }

    _, err := client.PutItem(ctx, putInput)

    // Handle conditional check failure
    var ccf *types.ConditionalCheckFailedException
    if errors.As(err, &ccf) {
        return fmt.Errorf("token already exists: %w", err)
    }

    return err
}

// CORRECT: Optimistic locking with version (goutils pattern)
type Offer struct {
    OfferID   string `dynamodbav:"offer_id"`
    Status    string `dynamodbav:"status"`
    Version   int    `dynamodbav:"version"` // ✅ Version counter
    UpdatedAt int64  `dynamodbav:"updated_at"`
}

func UpdateOfferStatus(ctx context.Context, offerID string, currentVersion int, newStatus string) error {
    updateInput := &dynamodb.UpdateItemInput{
        TableName: aws.String("Offers"),
        Key: map[string]types.AttributeValue{
            "offer_id": &types.AttributeValueMemberS{Value: offerID},
        },
        UpdateExpression: aws.String("SET #status = :new_status, #version = #version + :incr, #updated = :ts"),
        // ✅ Only update if version matches
        ConditionExpression: aws.String("#version = :expected_version"),
        ExpressionAttributeNames: map[string]string{
            "#status":  "status",
            "#version": "version",
            "#updated": "updated_at",
        },
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":new_status":        &types.AttributeValueMemberS{Value: newStatus},
            ":expected_version":  &types.AttributeValueMemberN{Value: strconv.Itoa(currentVersion)},
            ":incr":             &types.AttributeValueMemberN{Value: "1"},
            ":ts":               &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
        },
    }

    _, err := client.UpdateItem(ctx, updateInput)

    // Handle version mismatch
    var ccf *types.ConditionalCheckFailedException
    if errors.As(err, &ccf) {
        return fmt.Errorf("version conflict - item was modified: %w", err)
    }

    return err
}

// CORRECT: Idempotent create-or-update (dcs pattern)
func CreateOrUpdateConfig(ctx context.Context, key string, value string, timestamp int64) error {
    updateInput := &dynamodb.UpdateItemInput{
        TableName: aws.String("Configs"),
        Key: map[string]types.AttributeValue{
            "config_key": &types.AttributeValueMemberS{Value: key},
        },
        UpdateExpression: aws.String("SET #value = :val, #updated = :ts"),
        // ✅ Only update if doesn't exist OR timestamp is newer
        ConditionExpression: aws.String("attribute_not_exists(#updated) OR #updated < :ts"),
        ExpressionAttributeNames: map[string]string{
            "#value":   "value",
            "#updated": "updated_at",
        },
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":val": &types.AttributeValueMemberS{Value: value},
            ":ts":  &types.AttributeValueMemberN{Value: strconv.FormatInt(timestamp, 10)},
        },
    }

    _, err := client.UpdateItem(ctx, updateInput)

    // Idempotent: If condition fails, old value is newer (OK)
    var ccf *types.ConditionalCheckFailedException
    if errors.As(err, &ccf) {
        logger.Info(ctx, "config_already_newer", "key", key)
        return nil
    }

    return err
}
```

### Detection Strategy

```bash
# Find PutItem/UpdateItem operations
grep -n "PutItem\|UpdateItem" <pr_files>

# Check for ConditionExpression
grep -B 5 -A 10 "PutItem\|UpdateItem" <pr_files> | grep "ConditionExpression"

# Flag if missing
grep -A 10 "PutItem" <pr_files> | grep -L "ConditionExpression"
```

### Flag Conditions

Flag if:
- `PutItem` without `ConditionExpression` (create operations)
- `UpdateItem` modifying critical fields without condition
- Comment mentions "prevent duplicates" but no condition
- Financial operations (offers, payments, credits) without condition

### Severity

🚨 **Critical** - Data integrity:
- Duplicate records
- Lost updates
- Race conditions
- Data corruption

### References

**Razorpay production code:**
- `identity-provider:internal/storage/dynamo/domain.go` - PK/SK existence check
- `goutils:kvstore/dynamo/creator.go` - Generic condition builder
- `qr-codes:.claude/skills/code-review` - Best practices
- `upi-switch:pkg/storage/dynamoaws/creator.go` - Partition key check

---

## Check 3: Batch Operations (Avoid N+1) ⚠️ HIGH

### What to Check

Use BatchGetItem/BatchWriteItem instead of loops with GetItem/PutItem to reduce API calls and improve throughput.

### Razorpay Limits (offers-engine skill)

**BatchGetItem**: 100 items max
**BatchWriteItem**: 25 items max

**virtual-account pattern**: Always handle unprocessed items with retry

### Bad Pattern ❌

```go
// ANTI-PATTERN: N+1 queries (avoid this!)
func GetOffers(ctx context.Context, offerIDs []string) ([]*Offer, error) {
    var offers []*Offer

    for _, id := range offerIDs {  // 100 IDs
        // ❌ 100 separate GetItem calls!
        result, _ := client.GetItem(ctx, &dynamodb.GetItemInput{
            TableName: aws.String("Offers"),
            Key: map[string]types.AttributeValue{
                "offer_id": &types.AttributeValueMemberS{Value: id},
            },
        })

        var offer Offer
        attributevalue.UnmarshalMap(result.Item, &offer)
        offers = append(offers, &offer)
    }

    return offers, nil
}

// Problem: 100 offers = 100 API calls
// Cost: 100 RCUs (vs 12.5 RCUs with batch)
// Latency: ~2000ms (vs ~200ms with batch)
```

### Good Pattern ✅

```go
// CORRECT: Batch read with retry (virtual-account pattern)
func BatchGetOffers(ctx context.Context, offerIDs []string) ([]*Offer, error) {
    var offers []*Offer

    // ✅ Split into batches of 100 (DynamoDB limit)
    for i := 0; i < len(offerIDs); i += 100 {
        end := i + 100
        if end > len(offerIDs) {
            end = len(offerIDs)
        }

        batch := offerIDs[i:end]
        batchOffers, err := batchGetWithRetry(ctx, batch)
        if err != nil {
            return nil, err
        }

        offers = append(offers, batchOffers...)
    }

    return offers, nil
}

func batchGetWithRetry(ctx context.Context, ids []string) ([]*Offer, error) {
    // Build request items
    keys := make([]map[string]types.AttributeValue, len(ids))
    for i, id := range ids {
        keys[i] = map[string]types.AttributeValue{
            "offer_id": &types.AttributeValueMemberS{Value: id},
        }
    }

    batchInput := &dynamodb.BatchGetItemInput{
        RequestItems: map[string]types.KeysAndAttributes{
            "Offers": {Keys: keys},
        },
    }

    var offers []*Offer
    maxRetries := 3

    for attempt := 0; attempt < maxRetries; attempt++ {
        result, err := client.BatchGetItem(ctx, batchInput)
        if err != nil {
            return nil, err
        }

        // ✅ Parse results
        for _, item := range result.Responses["Offers"] {
            var offer Offer
            attributevalue.UnmarshalMap(item, &offer)
            offers = append(offers, &offer)
        }

        // ✅ Handle unprocessed keys (retry)
        if len(result.UnprocessedKeys) == 0 {
            break
        }

        logger.Warn(ctx, "unprocessed_keys_retry",
            "attempt", attempt+1,
            "count", len(result.UnprocessedKeys["Offers"].Keys))

        // Exponential backoff
        time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond)

        // Retry unprocessed keys
        batchInput.RequestItems = result.UnprocessedKeys
    }

    return offers, nil
}

// CORRECT: Batch write (25 items limit)
func BatchCreateOffers(ctx context.Context, offers []*Offer) error {
    // ✅ Split into batches of 25 (DynamoDB limit)
    for i := 0; i < len(offers); i += 25 {
        end := i + 25
        if end > len(offers) {
            end = len(offers)
        }

        batch := offers[i:end]
        if err := batchWriteWithRetry(ctx, batch); err != nil {
            return err
        }
    }

    return nil
}

func batchWriteWithRetry(ctx context.Context, offers []*Offer) error {
    // Build write requests
    writeRequests := make([]types.WriteRequest, len(offers))
    for i, offer := range offers {
        av, _ := attributevalue.MarshalMap(offer)
        writeRequests[i] = types.WriteRequest{
            PutRequest: &types.PutRequest{Item: av},
        }
    }

    batchInput := &dynamodb.BatchWriteItemInput{
        RequestItems: map[string][]types.WriteRequest{
            "Offers": writeRequests,
        },
    }

    maxRetries := 3
    for attempt := 0; attempt < maxRetries; attempt++ {
        result, err := client.BatchWriteItem(ctx, batchInput)
        if err != nil {
            return err
        }

        // ✅ Check for unprocessed items
        if len(result.UnprocessedItems) == 0 {
            return nil
        }

        logger.Warn(ctx, "unprocessed_items_retry",
            "attempt", attempt+1,
            "count", len(result.UnprocessedItems["Offers"]))

        // Exponential backoff
        time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond)

        // Retry unprocessed items
        batchInput.RequestItems = result.UnprocessedItems
    }

    return errors.New("max retries exceeded with unprocessed items")
}
```

### Detection Strategy

```bash
# Find GetItem in loops
grep -B 5 "GetItem" <pr_files> | grep "for.*range"

# Find PutItem in loops
grep -B 5 "PutItem" <pr_files> | grep "for.*range"

# Flag if no batch operations
grep -L "BatchGetItem\|BatchWriteItem" <pr_files_with_loops>
```

### Flag Conditions

Flag if:
- `GetItem` inside `for` loop
- `PutItem` inside `for` loop
- More than 5 items fetched/written without batch operation
- No retry handling for unprocessed items

### Severity

⚠️ **High** - Performance and cost:
- 10-20x higher RCU/WCU consumption
- 10x slower queries
- Higher AWS costs
- Throttling under load

### References

**Razorpay production code:**
- `offers-engine:.claude/skills` - BatchGetItem: 100, BatchWriteItem: 25
- `virtual-account:.agents/skills/va-dynamodb-patterns` - Batch operations with retry
- `upi-switch, payments-mandate` - BatchGetItem/BatchWriteItem usage

---

## Check 4: TTL Configuration for Temporary Data 🚨 CRITICAL

### What to Check

Tables storing temporary data (sessions, tokens, cache) must have TTL attribute enabled for automatic expiration.

### Razorpay Pattern (identity-provider)

**TTL usage found in:**
- `identity-provider` - Refresh tokens with auto-expiration
- Pattern: `UpdateTimeToLiveInput` with `AttributeName: "ttl"`

### Bad Pattern ❌

```go
// ANTI-PATTERN: No TTL - data never expires
type Session struct {
    SessionID string `dynamodbav:"session_id"`
    UserID    string `dynamodbav:"user_id"`
    CreatedAt int64  `dynamodbav:"created_at"`
    // ❌ No TTL attribute! Sessions accumulate forever
}

func CreateSession(ctx context.Context, session *Session) error {
    av, _ := attributevalue.MarshalMap(session)

    _, err := client.PutItem(ctx, &dynamodb.PutItemInput{
        TableName: aws.String("Sessions"),
        Item:      av,
    })

    return err
}

// Problem:
// - Sessions never deleted
// - Table grows unbounded
// - Storage costs increase
// - Query performance degrades
// ❌ Manual cleanup needed!

// ANTI-PATTERN: Wrong TTL format
type Token struct {
    TokenID   string `dynamodbav:"token_id"`
    ExpiresAt string `dynamodbav:"expires_at"` // ❌ String, not Unix timestamp!
}
// DynamoDB TTL requires Unix timestamp (seconds since epoch), not ISO8601 string
```

### Good Pattern ✅

```go
// CORRECT: Proper TTL for sessions (identity-provider pattern)
type RefreshToken struct {
    PK        string  `dynamodbav:"PK"`       // "USER#<user_id>"
    SK        string  `dynamodbav:"SK"`       // "TOKEN#<token_id>"
    UserID    string  `dynamodbav:"user_id"`
    IssuedAt  int64   `dynamodbav:"issued_at"`
    TTL       int64   `dynamodbav:"ttl"`      // ✅ Unix timestamp
    RevokedAt *int64  `dynamodbav:"RevokedAt,omitempty"`
}

func CreateRefreshToken(ctx context.Context, userID string, tokenID string) (*RefreshToken, error) {
    now := time.Now()
    token := &RefreshToken{
        PK:       fmt.Sprintf("USER#%s", userID),
        SK:       fmt.Sprintf("TOKEN#%s", tokenID),
        UserID:   userID,
        IssuedAt: now.Unix(),
        // ✅ Expire in 30 days (Unix timestamp)
        TTL:      now.Add(30 * 24 * time.Hour).Unix(),
    }

    av, _ := attributevalue.MarshalMap(token)

    _, err := client.PutItem(ctx, &dynamodb.PutItemInput{
        TableName: aws.String("RefreshTokens"),
        Item:      av,
        ConditionExpression: aws.String("attribute_not_exists(PK) and attribute_not_exists(SK)"),
    })

    return token, err
}

// CORRECT: Enable TTL on table (one-time setup)
// Migration code (identity-provider pattern)
func EnableTTL(ctx context.Context, tableName string) error {
    ttlInput := &dynamodb.UpdateTimeToLiveInput{
        TableName: aws.String(tableName),
        TimeToLiveSpecification: &types.TimeToLiveSpecification{
            Enabled:       aws.Bool(true),
            AttributeName: aws.String("ttl"),  // ✅ Must match field name
        },
    }

    _, err := client.UpdateTimeToLive(ctx, ttlInput)
    return err
}

// Terraform/CDK - Table with TTL
resource "aws_dynamodb_table" "refresh_tokens" {
  name           = "refresh_tokens"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "PK"
  range_key      = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  # ✅ Enable TTL
  ttl {
    attribute_name = "ttl"
    enabled        = true
  }
}

// CORRECT: Different TTL durations by use case
const (
    SessionTTL       = 24 * time.Hour      // User sessions
    TokenTTL         = 30 * 24 * time.Hour // Refresh tokens
    CacheTTL         = 1 * time.Hour       // Cached data
    TempDataTTL      = 15 * time.Minute    // Temporary processing data
)

func CreateSession(userID string, ttlDuration time.Duration) *Session {
    now := time.Now()
    return &Session{
        SessionID: uuid.NewString(),
        UserID:    userID,
        CreatedAt: now.Unix(),
        TTL:       now.Add(ttlDuration).Unix(),  // ✅ Configurable TTL
    }
}
```

### Detection Strategy

```bash
# Find DynamoDB table definitions
grep -n "type.*struct" <pr_files> | grep -i "session\|token\|cache\|temp"

# Check for TTL field
grep -A 10 "type.*Session\|Token\|Cache" <pr_files> | grep "ttl\|TTL"

# Find table creation
grep -n "CreateTable\|aws_dynamodb_table" <pr_files>

# Check TTL enabled
grep "TimeToLiveSpecification\|ttl.*enabled" <pr_files>
```

### Flag Conditions

Flag if:
- Table stores sessions/tokens/cache without TTL field
- TTL field is string type (not int64 Unix timestamp)
- CreateTable/migration without TTL specification
- No `UpdateTimeToLiveInput` for temporary data tables

### Severity

🚨 **Critical** - Operational issues:
- Unbounded table growth
- Storage cost explosion
- Query performance degradation
- Manual cleanup required
- Potential data retention compliance issues

### References

**Razorpay production code:**
- `identity-provider:cmd/migrations/main.go` - TTL setup with UpdateTimeToLiveInput
- Pattern: `AttributeName: "ttl"`, `Enabled: true`

---

## Check 5: GSI Projection Type Optimization 📋 MEDIUM

### What to Check

Global Secondary Indexes should use appropriate projection type (KEYS_ONLY, INCLUDE, ALL) based on query patterns.

### Razorpay Pattern (identity-provider)

**GSI found in:**
- `identity-provider:cmd/migrations/main.go` - GlobalSecondaryIndexes configuration
- `user-service:test/unit/test.go` - GSI setup

### Bad Pattern ❌

```go
// ANTI-PATTERN: Wasteful ALL projection
// GSI projects ALL attributes but only queries a few fields

// Table definition
type Offer struct {
    OfferID      string `dynamodbav:"offer_id"`     // PARTITION KEY
    MerchantID   string `dynamodbav:"merchant_id"`  // GSI KEY
    Status       string `dynamodbav:"status"`
    Description  string `dynamodbav:"description"`  // 5KB text
    Terms        string `dynamodbav:"terms"`        // 10KB text
    Metadata     string `dynamodbav:"metadata"`     // 20KB JSON
}

// Migration with wasteful GSI
GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
    {
        IndexName: aws.String("MerchantIndex"),
        KeySchema: []types.KeySchemaElement{
            {AttributeName: aws.String("merchant_id"), KeyType: types.KeyTypeHash},
        },
        Projection: &types.Projection{
            ProjectionType: types.ProjectionTypeAll,  // ❌ Projects 35KB+ per item!
        },
    },
}

// Query only needs offer_id and status
func GetOffersByMerchant(merchantID string) ([]string, error) {
    result, _ := client.Query(ctx, &dynamodb.QueryInput{
        TableName: aws.String("Offers"),
        IndexName: aws.String("MerchantIndex"),
        // Only need offer_id, status
        ProjectionExpression: aws.String("offer_id, #status"),
    })
    // ❌ But GSI stored 35KB per item (wasted storage cost!)
}
```

### Good Pattern ✅

```go
// CORRECT: KEYS_ONLY when fetching full item anyway
GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
    {
        IndexName: aws.String("MerchantIndex"),
        KeySchema: []types.KeySchemaElement{
            {AttributeName: aws.String("merchant_id"), KeyType: types.KeyTypeHash},
        },
        Projection: &types.Projection{
            ProjectionType: types.ProjectionTypeKeysOnly,  // ✅ Just keys
        },
    },
}

// Query returns keys, then fetch full items
func GetOffersByMerchant(ctx context.Context, merchantID string) ([]*Offer, error) {
    // Step 1: Query GSI for keys
    queryResult, _ := client.Query(ctx, &dynamodb.QueryInput{
        TableName: aws.String("Offers"),
        IndexName: aws.String("MerchantIndex"),
        KeyConditionExpression: aws.String("merchant_id = :mid"),
    })

    // Step 2: Batch get full items (if needed)
    offerIDs := extractOfferIDs(queryResult.Items)
    return BatchGetOffers(ctx, offerIDs)
}

// CORRECT: INCLUDE projection for specific attributes
GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
    {
        IndexName: aws.String("MerchantStatusIndex"),
        KeySchema: []types.KeySchemaElement{
            {AttributeName: aws.String("merchant_id"), KeyType: types.KeyTypeHash},
            {AttributeName: aws.String("status"), KeyType: types.KeyTypeRange},
        },
        Projection: &types.Projection{
            ProjectionType: types.ProjectionTypeInclude,  // ✅ Include specific fields
            NonKeyAttributes: aws.StringSlice([]string{
                "offer_id",
                "updated_at",
                "expires_at",
            }),
        },
    },
}

// Query gets needed fields directly from GSI
func GetActiveOffers(ctx context.Context, merchantID string) ([]OfferSummary, error) {
    result, _ := client.Query(ctx, &dynamodb.QueryInput{
        TableName: aws.String("Offers"),
        IndexName: aws.String("MerchantStatusIndex"),
        KeyConditionExpression: aws.String("merchant_id = :mid AND #status = :active"),
        // ✅ All fields available in GSI, no additional fetch needed
    })
    // Fast and cost-effective!
}

// Decision matrix
// KEYS_ONLY:    When you'll fetch full item anyway
// INCLUDE:      When you need specific fields (5-10 attributes)
// ALL:          When you need all fields AND won't fetch from table
```

### Severity

📋 **Medium** - Cost and performance:
- Higher storage costs (GSI storage = extra cost)
- Slower writes (more data to replicate to GSI)
- Wasted capacity

---

## Check 6: TransactWriteItems for Multi-Item ACID Operations 🚨 CRITICAL

### What to Check

Operations updating multiple items must use TransactWriteItems for atomicity (all succeed or all fail).

### Bad Pattern ❌

```go
// ANTI-PATTERN: No atomicity - partial failure leaves inconsistent state
func TransferCredits(ctx context.Context, fromUserID, toUserID string, amount int) error {
    // Step 1: Deduct from sender
    _, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String("Users"),
        Key: map[string]types.AttributeValue{
            "user_id": &types.AttributeValueMemberS{Value: fromUserID},
        },
        UpdateExpression: aws.String("SET credits = credits - :amount"),
    })
    if err != nil {
        return err  // ❌ Rolled back
    }

    // Step 2: Add to recipient
    _, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String("Users"),
        Key: map[string]types.AttributeValue{
            "user_id": &types.AttributeValueMemberS{Value: toUserID},
        },
        UpdateExpression: aws.String("SET credits = credits + :amount"),
    })
    if err != nil {
        // ❌ Step 1 succeeded, Step 2 failed!
        // Credits deducted but not added → money lost!
        return err
    }

    return nil
}
```

### Good Pattern ✅

```go
// CORRECT: Atomic transaction (all or nothing)
func TransferCredits(ctx context.Context, fromUserID, toUserID string, amount int) error {
    transactInput := &dynamodb.TransactWriteItemsInput{
        TransactItems: []types.TransactWriteItem{
            {
                // Deduct from sender
                Update: &types.Update{
                    TableName: aws.String("Users"),
                    Key: map[string]types.AttributeValue{
                        "user_id": &types.AttributeValueMemberS{Value: fromUserID},
                    },
                    UpdateExpression: aws.String("SET credits = credits - :amount"),
                    // ✅ Ensure sufficient balance
                    ConditionExpression: aws.String("credits >= :amount"),
                    ExpressionAttributeValues: map[string]types.AttributeValue{
                        ":amount": &types.AttributeValueMemberN{Value: strconv.Itoa(amount)},
                    },
                },
            },
            {
                // Add to recipient
                Update: &types.Update{
                    TableName: aws.String("Users"),
                    Key: map[string]types.AttributeValue{
                        "user_id": &types.AttributeValueMemberS{Value: toUserID},
                    },
                    UpdateExpression: aws.String("SET credits = credits + :amount"),
                    ExpressionAttributeValues: map[string]types.AttributeValue{
                        ":amount": &types.AttributeValueMemberN{Value: strconv.Itoa(amount)},
                    },
                },
            },
        },
    }

    _, err := client.TransactWriteItems(ctx, transactInput)

    // ✅ Either both succeed or both fail (ACID)
    var txCanceled *types.TransactionCanceledException
    if errors.As(err, &txCanceled) {
        return errors.New("transaction failed - insufficient balance or conflict")
    }

    return err
}
```

### Severity

🚨 **Critical** - Data consistency:
- Partial updates → inconsistent state
- Lost money/credits
- Audit trail breaks
- Impossible to recover

---

## Check 7: ConsistentRead for Critical Operations ⚠️ HIGH

### What to Check

Critical read-after-write operations must use `ConsistentRead: true`.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Eventual consistency for critical read
func ProcessPayment(ctx context.Context, paymentID string) error {
    // Write payment
    CreatePayment(ctx, payment)

    // Immediately read (eventual consistency)
    getResult, _ := client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String("Payments"),
        Key: paymentKey(paymentID),
        // ❌ ConsistentRead not set (defaults to false)
    })

    // ❌ May not find payment just created!
    if getResult.Item == nil {
        return errors.New("payment not found")
    }
}
```

### Good Pattern ✅

```go
// CORRECT: Strongly consistent read
func ProcessPayment(ctx context.Context, paymentID string) error {
    CreatePayment(ctx, payment)

    // ✅ Strongly consistent read
    getResult, _ := client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String("Payments"),
        Key: paymentKey(paymentID),
        ConsistentRead: aws.Bool(true),  // ✅ Read-after-write consistency
    })

    // Guaranteed to find payment just created
}

// Use eventual consistency for non-critical reads (cheaper, faster)
func GetOfferForDisplay(ctx context.Context, offerID string) (*Offer, error) {
    result, _ := client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String("Offers"),
        Key: offerKey(offerID),
        // ConsistentRead: false (default) - ✅ OK for display/analytics
    })
}
```

### Severity

⚠️ **High** - Data visibility:
- Read-after-write inconsistency
- Transient failures
- User confusion

---

## Check 8: On-Demand vs Provisioned Capacity Mode 📋 MEDIUM

### What to Check

Tables with unpredictable traffic should use on-demand billing mode.

### Good Pattern ✅

```go
// Terraform/CDK
resource "aws_dynamodb_table" "offers" {
  name         = "offers"
  billing_mode = "PAY_PER_REQUEST"  // ✅ On-demand for variable load

  # OR

  billing_mode   = "PROVISIONED"
  read_capacity  = 100   // Baseline RCU
  write_capacity = 50    // Baseline WCU

  # With auto-scaling
}
```

### Severity

📋 **Medium** - Cost optimization

---

## Summary Table

| Check # | Pattern | Severity | Razorpay Services |
|---------|---------|----------|-------------------|
| 1 | High Cardinality Partition Key | 🚨 Critical | offers-engine, identity-provider, upi-switch |
| 2 | Conditional Writes | 🚨 Critical | identity-provider, goutils, qr-codes, upi-switch |
| 3 | Batch Operations | ⚠️ High | offers-engine, virtual-account, upi-switch |
| 4 | TTL Configuration | 🚨 Critical | identity-provider (tokens) |
| 5 | GSI Projection Type | 📋 Medium | identity-provider, user-service |
| 6 | TransactWriteItems | 🚨 Critical | Financial operations |
| 7 | ConsistentRead | ⚠️ High | Read-after-write scenarios |
| 8 | Capacity Mode | 📋 Medium | Cost optimization |

---

## Integration with Razorpay Systems

### Kubestash/Credstash
- Secrets stored in DynamoDB `kubestash-*` tables
- credstash-v3 service reads/writes secrets
- Conditional writes prevent conflicts

### QR Code Migration
- Mentioned in jargon-explainer: "QR data to DynamoDB"
- Likely uses TTL for temporary QR codes

### Observability
- Monitor `ProvisionedThroughputExceededException`
- CloudWatch: `ConsumedReadCapacityUnits`, `ConsumedWriteCapacityUnits`
- Track unprocessed items in batch operations
