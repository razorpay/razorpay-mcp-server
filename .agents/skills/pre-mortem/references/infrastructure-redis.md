# Redis/Cache Infrastructure Checks

## Overview

Validates Redis and caching patterns to prevent memory leaks, lock issues, cache stampede, and stale data problems.

**Load when:** PR modifies `internal/cache/*`, `pkg/redis/*`, or caching-related code

**Total Checks:** 8

**Severity Distribution:**
- 🚨 Critical: 3
- ⚠️ High: 3
- 📋 Medium: 2

---

## Check 1: TTL on All Cache Keys 🚨 CRITICAL

### What to Check

Every Redis key must have TTL to prevent unbounded memory growth.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No TTL on cache key
func CacheTerminal(terminalId string, data interface{}) error {
    key := fmt.Sprintf("terminal:%s", terminalId)
    redis.Set(ctx, key, data, 0)  // ❌ TTL = 0 means no expiration!
    return nil
}

// ANTI-PATTERN: Missing TTL on hash
func CacheMultipleFields(merchantId string, fields map[string]interface{}) {
    key := fmt.Sprintf("merchant:%s", merchantId)
    for field, value := range fields {
        redis.HSet(ctx, key, field, value)  // ❌ No EXPIRE set!
    }
}
```

**Problem:**
- Keys accumulate indefinitely
- Redis memory fills up
- Eviction starts affecting hot keys
- Production incident

### Good Pattern ✅

```go
// CORRECT: Always set TTL
const TerminalCacheTTL = 15 * time.Minute

func CacheTerminal(terminalId string, data interface{}) error {
    key := fmt.Sprintf("terminal:%s", terminalId)
    redis.Set(ctx, key, data, TerminalCacheTTL)  // ✅ TTL from config
    return nil
}

// CORRECT: Set EXPIRE after HSET
func CacheMultipleFields(merchantId string, fields map[string]interface{}) {
    key := fmt.Sprintf("merchant:%s", merchantId)

    pipe := redis.Pipeline()
    for field, value := range fields {
        pipe.HSet(ctx, key, field, value)
    }
    pipe.Expire(ctx, key, 1*time.Hour)  // ✅ Set TTL
    pipe.Exec(ctx)
}
```

### Detection Strategy

```bash
# Find Redis Set/HSet operations
grep -n "\.Set(" internal/cache/*.go pkg/redis/*.go
grep -n "\.HSet(" internal/cache/*.go pkg/redis/*.go

# For each Set/HSet, verify:
# 1. Set() has non-zero TTL parameter
# 2. HSet() followed by Expire() within 10 lines
# 3. Or uses SetEX/SetNX with TTL
```

### Flag Conditions

Flag if:
- `redis.Set(ctx, key, value, 0)` - zero TTL
- `redis.HSet()` without `redis.Expire()` call
- `redis.SetNX()` without TTL parameter
- No TTL constant defined

### Severity

🚨 **Critical** - Memory leak, Redis OOM, production incident

### Reference

Based on terminals: `pkg/redis/cache.go:89`

---

## Check 2: Distributed Lock Management 🚨 CRITICAL

### What to Check

Distributed locks must have timeout, proper unlock, and handle edge cases.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Lock without timeout
func ProcessWithLock(resourceId string) error {
    lockKey := fmt.Sprintf("lock:%s", resourceId)

    // Acquire lock
    redis.SetNX(ctx, lockKey, "locked", 0)  // ❌ No TTL - lock never expires!

    // Process
    processResource(resourceId)

    // Unlock
    redis.Del(ctx, lockKey)
    return nil
}

// ANTI-PATTERN: No unlock on error
func ProcessWithLock2(resourceId string) error {
    lockKey := fmt.Sprintf("lock:%s", resourceId)
    redis.SetNX(ctx, lockKey, "locked", 30*time.Second)

    if err := processResource(resourceId); err != nil {
        return err  // ❌ Lock not released on error!
    }

    redis.Del(ctx, lockKey)
    return nil
}

// ANTI-PATTERN: Deleting someone else's lock
func ProcessWithLock3(resourceId string) error {
    lockKey := fmt.Sprintf("lock:%s", resourceId)
    redis.SetNX(ctx, lockKey, "locked", 5*time.Second)

    time.Sleep(10 * time.Second)  // Processing takes longer than lock TTL

    redis.Del(ctx, lockKey)  // ❌ Deletes another process's lock!
    return nil
}
```

**Problem:**
- Locks never released (deadlock)
- Multiple processes access critical section
- Lock stolen by wrong owner

### Good Pattern ✅

```go
// CORRECT: Proper distributed lock
type DistributedLock struct {
    redis     *redis.Client
    key       string
    value     string  // Unique token (UUID)
    ttl       time.Duration
}

func NewLock(resourceId string, ttl time.Duration) *DistributedLock {
    return &DistributedLock{
        redis: redisClient,
        key:   fmt.Sprintf("lock:%s", resourceId),
        value: uuid.NewString(),  // Unique identifier
        ttl:   ttl,
    }
}

func (l *DistributedLock) Acquire(ctx context.Context) (bool, error) {
    acquired, err := l.redis.SetNX(ctx, l.key, l.value, l.ttl).Result()
    return acquired, err
}

func (l *DistributedLock) Release(ctx context.Context) error {
    // Lua script to ensure we only delete our own lock
    script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `
    return l.redis.Eval(ctx, script, []string{l.key}, l.value).Err()
}

// Usage
func ProcessWithLock(resourceId string) error {
    lock := NewLock(resourceId, 30*time.Second)

    acquired, err := lock.Acquire(ctx)
    if err != nil {
        return err
    }
    if !acquired {
        return errors.New("resource already locked")
    }

    defer lock.Release(ctx)  // ✅ Always unlock

    return processResource(resourceId)
}
```

### Severity

🚨 **Critical** - Deadlocks, race conditions, data corruption

---

## Check 3: Cache Stampede Protection 🚨 CRITICAL

### What to Check

Multiple concurrent requests shouldn't all regenerate cache on miss.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Cache stampede
func GetTerminal(terminalId string) (*Terminal, error) {
    cacheKey := fmt.Sprintf("terminal:%s", terminalId)

    // Try cache
    cached, err := redis.Get(ctx, cacheKey).Result()
    if err == redis.Nil {
        // ❌ 1000 concurrent requests all call DB!
        terminal, err := db.FindTerminal(terminalId)
        if err != nil {
            return nil, err
        }

        // All 1000 requests write to cache
        redis.Set(ctx, cacheKey, terminal, 15*time.Minute)
        return terminal, nil
    }

    return unmarshal(cached), nil
}
```

**Problem:**
- Cache expires
- 1000 requests hit simultaneously
- All see cache miss
- All query database (thundering herd)
- Database overload

### Good Pattern ✅

```go
// PATTERN 1: Single-flight (Go-specific)
import "golang.org/x/sync/singleflight"

var group singleflight.Group

func GetTerminal(terminalId string) (*Terminal, error) {
    cacheKey := fmt.Sprintf("terminal:%s", terminalId)

    // Try cache
    cached, err := redis.Get(ctx, cacheKey).Result()
    if err != redis.Nil {
        return unmarshal(cached), nil
    }

    // ✅ Only one request executes, others wait
    result, err, _ := group.Do(terminalId, func() (interface{}, error) {
        terminal, err := db.FindTerminal(terminalId)
        if err != nil {
            return nil, err
        }

        redis.Set(ctx, cacheKey, terminal, 15*time.Minute)
        return terminal, nil
    })

    return result.(*Terminal), err
}

// PATTERN 2: Probabilistic early expiration
func GetTerminalWithEarlyExpire(terminalId string) (*Terminal, error) {
    cacheKey := fmt.Sprintf("terminal:%s", terminalId)

    type CachedTerminal struct {
        Data      *Terminal
        ExpiresAt time.Time
    }

    cached, err := redis.Get(ctx, cacheKey).Result()
    if err == redis.Nil {
        // Cache miss - regenerate
        return regenerateCache(terminalId)
    }

    cachedData := unmarshal(cached)

    // ✅ Regenerate early probabilistically
    delta := cachedData.ExpiresAt.Sub(time.Now())
    if delta < 0 {
        return regenerateCache(terminalId)
    }

    // Calculate probability of early refresh
    // As expiry approaches, probability increases
    probability := 1.0 - (delta.Seconds() / (15 * 60))  // 15 min TTL
    if rand.Float64() < probability * 0.1 {  // Max 10% chance
        go regenerateCache(terminalId)  // Background refresh
    }

    return cachedData.Data, nil
}
```

### Detection Strategy

Look for cache Get → database query pattern without singleflight or locking.

### Severity

🚨 **Critical** - Database overload, cascading failures

---

## Check 4: Redis Connection Pool Configuration ⚠️ HIGH

### What to Check

Redis client must have proper pool size, timeouts, and retry settings.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No pool configuration
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    // ❌ No PoolSize - defaults to 10!
    // ❌ No Timeouts
    // ❌ No MaxRetries
})
```

**Problem:**
- Default pool too small for production
- No timeout → hanging connections
- No retries → transient failures fail requests

### Good Pattern ✅

```go
// CORRECT: Production-ready config
redisClient := redis.NewClient(&redis.Options{
    Addr:         config.Redis.Host,
    Password:     config.Redis.Password,
    DB:           config.Redis.DB,

    // Pool settings
    PoolSize:     100,              // ✅ Adequate for load
    MinIdleConns: 10,               // ✅ Keep warm connections
    MaxConnAge:   5 * time.Minute,  // ✅ Recycle connections
    PoolTimeout:  2 * time.Second,  // ✅ Fail fast if pool exhausted

    // Timeouts
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,

    // Retries
    MaxRetries:      3,
    MinRetryBackoff: 8 * time.Millisecond,
    MaxRetryBackoff: 512 * time.Millisecond,
})
```

### Severity

⚠️ **High** - Connection exhaustion, slow responses

---

## Check 5: Error Handling on Cache Operations ⚠️ HIGH

### What to Check

Cache errors should not fail the request - degrade gracefully.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Cache error fails request
func GetTerminal(terminalId string) (*Terminal, error) {
    cached, err := redis.Get(ctx, cacheKey).Result()
    if err != nil {
        return nil, err  // ❌ Returns error on cache failure!
    }

    return unmarshal(cached), nil
}
```

**Problem:**
- Redis down = all requests fail
- No graceful degradation

### Good Pattern ✅

```go
// CORRECT: Cache errors logged, not returned
func GetTerminal(terminalId string) (*Terminal, error) {
    cached, err := redis.Get(ctx, cacheKey).Result()
    if err != nil && err != redis.Nil {
        // ✅ Log error but continue
        logger.Warn(ctx, "cache_read_failed", "error", err)
    }

    if err == nil {
        return unmarshal(cached), nil  // Cache hit
    }

    // Cache miss or error - fetch from DB
    terminal, err := db.FindTerminal(terminalId)
    if err != nil {
        return nil, err
    }

    // Try to cache (ignore errors)
    if err := redis.Set(ctx, cacheKey, terminal, 15*time.Minute).Err(); err != nil {
        logger.Warn(ctx, "cache_write_failed", "error", err)
    }

    return terminal, nil
}
```

### Severity

⚠️ **High** - Cascading failures, poor resilience

---

## Check 6: Cache Key Naming Convention ⚠️ HIGH

### What to Check

Cache keys should follow consistent naming pattern for debuggability.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Inconsistent, unclear keys
redis.Set(ctx, terminalId, data, ttl)  // ❌ What is this?
redis.Set(ctx, "t:"+terminalId, data, ttl)  // ❌ Unclear prefix
redis.Set(ctx, terminalId+"_cache", data, ttl)  // ❌ Inconsistent
```

### Good Pattern ✅

```go
// CORRECT: Clear, consistent naming
// Pattern: {service}:{entity}:{id}:{field}

const (
    KeyPrefixTerminal         = "terminals:terminal:"
    KeyPrefixMerchantSettings = "terminals:merchant:settings:"
    KeyPrefixGatewayConfig    = "terminals:gateway:config:"
)

func GetTerminalCacheKey(terminalId string) string {
    return fmt.Sprintf("%s%s", KeyPrefixTerminal, terminalId)
    // Result: "terminals:terminal:term_abc123"
}

func GetMerchantSettingsKey(merchantId string, settingName string) string {
    return fmt.Sprintf("%s%s:%s", KeyPrefixMerchantSettings, merchantId, settingName)
    // Result: "terminals:merchant:settings:merch_123:auto_refund"
}
```

### Severity

⚠️ **High** - Debugging difficult, key collisions

---

## Check 7: Cache Invalidation on Updates 📋 MEDIUM

### What to Check

Data updates must invalidate corresponding cache keys.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Update without cache invalidation
func UpdateTerminal(terminal *Terminal) error {
    // Update database
    err := db.Save(terminal).Error
    if err != nil {
        return err
    }

    // ❌ Cache still has old value!
    return nil
}
```

**Problem:**
- Stale data served from cache
- Inconsistency between DB and cache

### Good Pattern ✅

```go
// PATTERN 1: Invalidate on update
func UpdateTerminal(terminal *Terminal) error {
    tx := db.Begin()
    defer tx.Rollback()

    if err := tx.Save(terminal).Error; err != nil {
        return err
    }

    if err := tx.Commit().Error; err != nil {
        return err
    }

    // ✅ Invalidate cache after commit
    cacheKey := GetTerminalCacheKey(terminal.ID)
    redis.Del(ctx, cacheKey)

    return nil
}

// PATTERN 2: Write-through cache
func UpdateTerminal(terminal *Terminal) error {
    // Update DB
    if err := db.Save(terminal).Error; err != nil {
        return err
    }

    // ✅ Update cache immediately
    cacheKey := GetTerminalCacheKey(terminal.ID)
    redis.Set(ctx, cacheKey, terminal, 15*time.Minute)

    return nil
}
```

### Severity

📋 **Medium** - Stale data, but temporary (until TTL)

---

## Check 8: Serialization Errors Handled 📋 MEDIUM

### What to Check

Cache serialization/deserialization errors must be handled.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Ignoring unmarshal errors
func GetTerminal(terminalId string) (*Terminal, error) {
    cached, _ := redis.Get(ctx, cacheKey).Result()

    var terminal Terminal
    json.Unmarshal([]byte(cached), &terminal)  // ❌ Error ignored!

    return &terminal, nil  // Returns zero-value struct on error
}
```

### Good Pattern ✅

```go
// CORRECT: Handle unmarshal errors
func GetTerminal(terminalId string) (*Terminal, error) {
    cached, err := redis.Get(ctx, cacheKey).Result()
    if err == redis.Nil {
        return fetchFromDB(terminalId)
    }
    if err != nil {
        logger.Warn(ctx, "cache_get_failed", "error", err)
        return fetchFromDB(terminalId)
    }

    var terminal Terminal
    if err := json.Unmarshal([]byte(cached), &terminal); err != nil {
        logger.Error(ctx, "cache_unmarshal_failed", "error", err)
        // ✅ Invalidate corrupted cache entry
        redis.Del(ctx, cacheKey)
        return fetchFromDB(terminalId)
    }

    return &terminal, nil
}
```

### Severity

📋 **Medium** - Returns corrupt data, poor error handling

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | TTL on all keys | 🚨 Critical | Memory leak, Redis OOM |
| 2 | Distributed lock | 🚨 Critical | Deadlocks, race conditions |
| 3 | Cache stampede | 🚨 Critical | DB overload |
| 4 | Connection pool | ⚠️ High | Connection exhaustion |
| 5 | Error handling | ⚠️ High | Cascading failures |
| 6 | Key naming | ⚠️ High | Debugging hard |
| 7 | Cache invalidation | 📋 Medium | Stale data |
| 8 | Serialization errors | 📋 Medium | Corrupt data |

---

## How to Apply

**For each file matching** `internal/cache/*`, `pkg/redis/*`:

1. Check all Set/HSet operations have TTL
2. Verify distributed lock patterns
3. Look for cache stampede protection
4. Check Redis client configuration
5. Verify error handling degrades gracefully
6. Validate key naming consistency
7. Check updates invalidate cache
8. Verify unmarshal error handling

**Example output:**

```
📁 File: internal/cache/terminal_cache.go

🚨 Check #1 Failed: No TTL on cache key (Line 45)
   Code: redis.Set(ctx, key, data, 0)
   Fix: Use redis.Set(ctx, key, data, 15*time.Minute)

🚨 Check #2 Failed: Lock without timeout (Line 78)
   Code: redis.SetNX(ctx, lockKey, "locked", 0)
   Fix: Add TTL: redis.SetNX(ctx, lockKey, uuid, 30*time.Second)

⚠️  Check #5 Failed: Cache error fails request (Line 92)
   Code: if err != nil { return nil, err }
   Fix: Log error and fall back to DB

✅ Check #3 Passed: Using singleflight for stampede protection
✅ Check #4 Passed: Pool size configured
✅ Check #6 Passed: Consistent key naming
✅ Check #7 Passed: Cache invalidated on update
✅ Check #8 Passed: Unmarshal errors handled
```
