# Resilience Patterns Checks

## Overview

Validates circuit breakers, retry logic, timeout handling, and graceful degradation patterns to prevent cascading failures and improve system resilience.

**Load when:** PR modifies HTTP clients, external service calls, or resilience configurations

**Total Checks:** 8

**Severity Distribution:**
- 🚨 Critical: 3
- ⚠️ High: 3
- 📋 Medium: 2

---

## Check 1: Circuit Breaker for External Services 🚨 CRITICAL

### What to Check

Calls to external services must use circuit breaker to prevent cascading failures.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No circuit breaker
func CallGatewayAPI(terminalID string) (*Response, error) {
    // ❌ Gateway down = all requests hang/fail
    // ❌ No protection from cascading failure
    resp, err := http.Post(gatewayURL, "application/json", payload)
    if err != nil {
        return nil, err
    }
    return parseResponse(resp), nil
}

// In production:
// - Gateway has outage
// - All 1000 requests timeout waiting
// - Connection pool exhausted
// - Service crashes
```

**Problem:**
- External service outage crashes your service
- No fail-fast mechanism
- Thread/connection pool exhaustion
- Cascading failure

### Good Pattern ✅

```go
// CORRECT: Circuit breaker with hystrix or gobreaker
import "github.com/sony/gobreaker"

var gatewayCircuitBreaker *gobreaker.CircuitBreaker

func init() {
    settings := gobreaker.Settings{
        Name:        "GatewayAPI",
        MaxRequests: 3,                    // ✅ Allow 3 requests in half-open state
        Interval:    time.Duration(0),     // ✅ Reset counts on state change
        Timeout:     60 * time.Second,     // ✅ Half-open after 60s
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            // ✅ Open circuit after 5 consecutive failures
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 10 && failureRatio >= 0.6
        },
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            logger.Info("circuit_breaker_state_change",
                "name", name,
                "from", from.String(),
                "to", to.String())
        },
    }

    gatewayCircuitBreaker = gobreaker.NewCircuitBreaker(settings)
}

func CallGatewayAPI(terminalID string) (*Response, error) {
    // ✅ Execute with circuit breaker
    result, err := gatewayCircuitBreaker.Execute(func() (interface{}, error) {
        return callGatewayAPIInternal(terminalID)
    })

    if err != nil {
        if err == gobreaker.ErrOpenState {
            logger.Warn(ctx, "circuit_breaker_open", "service", "gateway")
            // ✅ Fail fast when circuit is open
            return nil, errors.New("gateway service unavailable")
        }
        return nil, err
    }

    return result.(*Response), nil
}

func callGatewayAPIInternal(terminalID string) (*Response, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := httpClient.Post(ctx, gatewayURL, payload)
    if err != nil {
        return nil, err
    }

    return parseResponse(resp), nil
}
```

**Circuit Breaker States:**
- **Closed**: Normal operation, requests flow through
- **Open**: Too many failures, fail fast without calling service
- **Half-Open**: After timeout, allow limited requests to test recovery

### Detection Strategy

```bash
# Find external service calls — stdlib HTTP and gRPC are reliable signals.
# Avoid generic [A-Za-z]+Client.* patterns — they match DB, cache, and Redis clients
# which have their own resilience patterns and don't need circuit breakers here.

# Stdlib HTTP (strong signal — always external)
grep -n "http\.Post\|http\.Get\|http\.Do\|http\.NewRequest" <pr_files>

# gRPC client calls (strong signal — always external)
grep -nE "grpc\.Dial|grpc\.DialContext" <pr_files>

# Named external service clients (Razorpay-specific patterns, not DB/cache)
# Look for client types from known external packages: mozart, asv, passport, stork, gateway
grep -nE "(mozart|asv|passport|stork|gateway|razorpay)\w*[Cc]lient\." <pr_files>

# Custom HTTP client wrappers (e.g., httpClient.Post, client.Do with http.Request)
grep -nE "httpClient\.(Post|Get|Put|Delete|Do)|\.Do\(req\)" <pr_files>

# For each call site found, check for a circuit breaker wrapper:
grep -n "gobreaker\|hystrix\|CircuitBreaker\|circuitBreaker\|\.Execute(" <pr_files>

# Flag files with strong external call signals but no circuit breaker
for file in <pr_files>; do
    has_external=$(grep -cE "http\.(Post|Get|Do|NewRequest)|grpc\.Dial|httpClient\.(Post|Get|Do)|\.Do\(req\)" "$file" 2>/dev/null)
    has_breaker=$(grep -cE "gobreaker|hystrix|CircuitBreaker|\.Execute\(" "$file" 2>/dev/null)
    if [ "${has_external:-0}" -gt 0 ] && [ "${has_breaker:-0}" -eq 0 ]; then
        echo "⚠️  $file: External HTTP/gRPC calls without circuit breaker"
    fi
done
```

> **Note:** Generic `[A-Za-z]+Client.*` patterns were dropped — they match `dbClient`, `cacheClient`, `redisClient` etc. which legitimately need no circuit breaker. The check focuses on stdlib HTTP, gRPC dials, and named external-service client patterns instead.

### Flag Conditions

Flag if:
- External API call without circuit breaker
- No timeout on HTTP request
- No failure threshold configuration
- Critical service call (payment gateway, auth, etc.)

### Severity

🚨 **Critical** - Cascading failures, service outages

### Reference

Pattern seen in terminals Mozart integration

---

## Check 2: Retry Logic with Exponential Backoff 🚨 CRITICAL

### What to Check

Transient failures must use exponential backoff retry, not infinite retry or no retry.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No retry
func FetchMerchantConfig(merchantID string) (*Config, error) {
    resp, err := http.Get(configURL)
    if err != nil {
        return nil, err  // ❌ Single transient error fails request
    }
    return parseConfig(resp), nil
}

// ANTI-PATTERN: Constant retry interval
func FetchWithBadRetry(merchantID string) (*Config, error) {
    for i := 0; i < 10; i++ {
        resp, err := http.Get(configURL)
        if err == nil {
            return parseConfig(resp), nil
        }
        time.Sleep(1 * time.Second)  // ❌ Same delay every time - hammers failing service
    }
    return nil, errors.New("max retries exceeded")
}

// ANTI-PATTERN: Infinite retry
func FetchWithInfiniteRetry(merchantID string) (*Config, error) {
    for {
        resp, err := http.Get(configURL)
        if err == nil {
            return parseConfig(resp), nil
        }
        time.Sleep(1 * time.Second)  // ❌ Retries forever, goroutine leak
    }
}
```

**Problem:**
- No retry: Fails on transient errors
- Constant retry: Hammers failing service
- Infinite retry: Goroutine leaks, connection pool exhaustion

### Good Pattern ✅

```go
// CORRECT: Exponential backoff with max retries
import "github.com/cenkalti/backoff/v4"

func FetchMerchantConfig(merchantID string) (*Config, error) {
    var config *Config

    operation := func() error {
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()

        resp, err := httpClient.Get(ctx, configURL)
        if err != nil {
            // Check if error is retryable
            if isRetryable(err) {
                logger.Warn(ctx, "fetch_config_failed_retrying", "error", err)
                return err  // ✅ Retry
            }
            // Non-retryable error (4xx client error)
            logger.Error(ctx, "fetch_config_failed_permanent", "error", err)
            return backoff.Permanent(err)  // ✅ Don't retry
        }

        config = parseConfig(resp)
        return nil
    }

    // ✅ Exponential backoff: 100ms, 200ms, 400ms, 800ms...
    exponentialBackoff := backoff.NewExponentialBackOff()
    exponentialBackoff.InitialInterval = 100 * time.Millisecond
    exponentialBackoff.MaxInterval = 10 * time.Second
    exponentialBackoff.MaxElapsedTime = 30 * time.Second  // ✅ Total max time

    err := backoff.Retry(operation, exponentialBackoff)
    if err != nil {
        return nil, fmt.Errorf("failed after retries: %w", err)
    }

    return config, nil
}

func isRetryable(err error) bool {
    // ✅ Retry on network errors, timeouts, 5xx
    if errors.Is(err, context.DeadlineExceeded) {
        return true
    }
    if errors.Is(err, syscall.ECONNREFUSED) {
        return true
    }
    // Check HTTP status code
    if statusCode >= 500 && statusCode < 600 {
        return true
    }
    return false
}
```

**Retry Strategy:**
- Initial: 100ms
- After 1st retry: 200ms
- After 2nd retry: 400ms
- After 3rd retry: 800ms
- Max interval: 10s
- Max total time: 30s
- Stop on permanent errors (4xx)

### Severity

🚨 **Critical** - Poor resilience, hammering services, goroutine leaks

---

## Check 3: Timeout on All External Calls 🚨 CRITICAL

### What to Check

Every external HTTP call, database query, and RPC must have timeout.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No timeout
func CallExternalService() (*Response, error) {
    // ❌ Can hang indefinitely if service is slow/stuck
    resp, err := http.Get("https://external-api.com/data")
    return parseResponse(resp), err
}

// ANTI-PATTERN: Context without deadline
func CallWithContext(ctx context.Context) error {
    // ❌ ctx has no timeout, can still hang
    resp, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    return err
}
```

**Problem:**
- Requests hang forever
- Connection pool exhaustion
- Goroutine leaks
- No fail-fast

### Good Pattern ✅

```go
// CORRECT: Timeout on every external call
func CallExternalService() (*Response, error) {
    // ✅ Context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := httpClient.Do(req)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            logger.Warn(ctx, "external_call_timeout", "url", url)
            return nil, ErrTimeout
        }
        return nil, err
    }

    return parseResponse(resp), nil
}

// CORRECT: HTTP client with default timeout
var httpClient = &http.Client{
    Timeout: 10 * time.Second,  // ✅ Default timeout for all requests
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   2 * time.Second,  // ✅ Connection timeout
            KeepAlive: 30 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   3 * time.Second,  // ✅ TLS timeout
        ResponseHeaderTimeout: 5 * time.Second,  // ✅ Header timeout
        ExpectContinueTimeout: 1 * time.Second,
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   10,
        IdleConnTimeout:       90 * time.Second,
    },
}
```

### Severity

🚨 **Critical** - Hanging requests, connection exhaustion

---

## Check 4: Fallback for Non-Critical Services ⚠️ HIGH

### What to Check

Non-critical service failures should have fallback values/behavior.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No fallback for non-critical data
func GetMerchantLogo(merchantID string) (*Logo, error) {
    logo, err := externalCDN.FetchLogo(merchantID)
    if err != nil {
        // ❌ Logo fetch failure fails entire request
        return nil, err
    }
    return logo, nil
}
```

**Problem:**
- Non-critical failure breaks user experience
- No graceful degradation

### Good Pattern ✅

```go
// CORRECT: Fallback for non-critical data
func GetMerchantLogo(merchantID string) *Logo {
    logo, err := externalCDN.FetchLogo(merchantID)
    if err != nil {
        logger.Warn(ctx, "cdn_fetch_failed_using_default", "error", err)

        // ✅ Return default logo instead of failing
        return &Logo{
            URL: "/static/default-merchant-logo.png",
            Size: "medium",
        }
    }
    return logo
}

// PATTERN 2: Cached fallback
func GetMerchantConfig(merchantID string) (*Config, error) {
    // Try live config
    config, err := configService.Fetch(merchantID)
    if err != nil {
        logger.Warn(ctx, "config_fetch_failed_trying_cache", "error", err)

        // ✅ Fallback to cached config
        cachedConfig, cacheErr := cache.Get(fmt.Sprintf("config:%s", merchantID))
        if cacheErr == nil {
            logger.Info(ctx, "using_cached_config")
            return cachedConfig, nil
        }

        // ✅ Ultimate fallback: default config
        logger.Warn(ctx, "using_default_config")
        return getDefaultConfig(), nil
    }

    // Cache successful fetch
    cache.Set(fmt.Sprintf("config:%s", merchantID), config, 15*time.Minute)
    return config, nil
}
```

### Severity

⚠️ **High** - Poor user experience, unnecessary failures

---

## Check 5: Bulkhead Pattern for Resource Isolation ⚠️ HIGH

### What to Check

Different services should use separate connection pools to prevent one failure from affecting all.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Shared HTTP client for all services
var sharedClient = &http.Client{
    Timeout: 10 * time.Second,
}

func CallGatewayAPI() {
    sharedClient.Get(gatewayURL)  // ❌ Shares pool with all other services
}

func CallPaymentAPI() {
    sharedClient.Get(paymentURL)  // ❌ If gateway is slow, payment calls blocked
}

func CallNotificationAPI() {
    sharedClient.Get(notificationURL)  // ❌ All services compete for same pool
}
```

**Problem:**
- One slow service blocks all others
- Connection pool exhaustion affects unrelated calls
- No resource isolation

### Good Pattern ✅

```go
// CORRECT: Separate clients for critical vs non-critical services
var (
    // ✅ Dedicated client for critical payment services
    criticalServicesClient = &http.Client{
        Timeout: 5 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 20,  // Higher for critical
        },
    }

    // ✅ Separate client for non-critical services
    nonCriticalServicesClient = &http.Client{
        Timeout: 3 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        50,
            MaxIdleConnsPerHost: 5,  // Lower for non-critical
        },
    }
)

func CallGatewayAPI() {
    // ✅ Uses dedicated critical pool
    criticalServicesClient.Get(gatewayURL)
}

func CallNotificationAPI() {
    // ✅ Uses separate non-critical pool
    // Slow notifications don't affect payments
    nonCriticalServicesClient.Get(notificationURL)
}
```

### Severity

⚠️ **High** - Lack of isolation, cascading failures

---

## Check 6: Rate Limiting for Outbound Calls ⚠️ HIGH

### What to Check

Outbound API calls should be rate-limited to prevent overwhelming external services.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No rate limiting
func ProcessBatch(terminals []Terminal) {
    for _, terminal := range terminals {
        // ❌ 10,000 terminals = 10,000 simultaneous API calls!
        go func(t Terminal) {
            CallGatewayAPI(t.ID)
        }(terminal)
    }
}
```

**Problem:**
- Overwhelms external service
- Gets rate-limited/banned
- Wastes resources

### Good Pattern ✅

```go
// CORRECT: Rate limit with token bucket
import "golang.org/x/time/rate"

var gatewayRateLimiter = rate.NewLimiter(rate.Limit(100), 100)  // ✅ 100 req/s, burst 100

func CallGatewayAPI(terminalID string) error {
    // ✅ Wait for rate limit token
    if err := gatewayRateLimiter.Wait(ctx); err != nil {
        return fmt.Errorf("rate limiter error: %w", err)
    }

    return callGatewayAPIInternal(terminalID)
}

// PATTERN 2: Worker pool for concurrency control
func ProcessBatch(terminals []Terminal) {
    const maxWorkers = 10  // ✅ Limit concurrent requests

    semaphore := make(chan struct{}, maxWorkers)
    var wg sync.WaitGroup

    for _, terminal := range terminals {
        wg.Add(1)
        semaphore <- struct{}{}  // ✅ Acquire slot

        go func(t Terminal) {
            defer wg.Done()
            defer func() { <-semaphore }()  // ✅ Release slot

            CallGatewayAPI(t.ID)
        }(terminal)
    }

    wg.Wait()
}
```

### Severity

⚠️ **High** - Service bans, resource waste

---

## Check 7: Health Check Endpoints 📋 MEDIUM

### What to Check

Service must expose health check endpoint for dependency monitoring.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No health checks
func main() {
    http.HandleFunc("/api/terminals", handleTerminals)
    // ❌ No /health or /ready endpoint
    http.ListenAndServe(":8080", nil)
}
```

**Problem:**
- Can't monitor service health
- Load balancer can't detect unhealthy instances
- No dependency health visibility

### Good Pattern ✅

```go
// CORRECT: Health and readiness checks
func main() {
    // ✅ Liveness: Is service running?
    http.HandleFunc("/health", healthHandler)

    // ✅ Readiness: Is service ready to serve traffic?
    http.HandleFunc("/ready", readinessHandler)

    http.HandleFunc("/api/terminals", handleTerminals)
    http.ListenAndServe(":8080", nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    // ✅ Basic health check
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
    })
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
    // ✅ Check dependencies
    checks := map[string]bool{
        "database": checkDatabaseConnection(),
        "redis":    checkRedisConnection(),
        "kafka":    checkKafkaConnection(),
    }

    allHealthy := true
    for _, healthy := range checks {
        if !healthy {
            allHealthy = false
            break
        }
    }

    status := http.StatusOK
    if !allHealthy {
        status = http.StatusServiceUnavailable
    }

    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": map[string]bool{
            "ready": allHealthy,
        },
        "dependencies": checks,
    })
}
```

### Severity

📋 **Medium** - Poor observability, manual detection of issues

---

## Check 8: Graceful Shutdown 📋 MEDIUM

### What to Check

Service must handle SIGTERM gracefully, finishing in-flight requests.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Abrupt shutdown
func main() {
    http.HandleFunc("/api/terminals", handleTerminals)
    http.ListenAndServe(":8080", nil)  // ❌ Kills mid-request on SIGTERM
}
```

**Problem:**
- In-flight requests fail
- Partial database transactions
- Poor user experience during deploys

### Good Pattern ✅

```go
// CORRECT: Graceful shutdown
func main() {
    srv := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }

    // Start server in goroutine
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("server_start_failed", "error", err)
        }
    }()

    // ✅ Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("shutting_down_server")

    // ✅ Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Error("server_forced_shutdown", "error", err)
    }

    logger.Info("server_exited")
}
```

### Severity

📋 **Medium** - Failed requests during deploys

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | Circuit breaker | 🚨 Critical | Cascading failures |
| 2 | Exponential backoff | 🚨 Critical | Service hammering |
| 3 | Timeouts | 🚨 Critical | Hanging requests |
| 4 | Fallback values | ⚠️ High | Poor UX |
| 5 | Bulkhead isolation | ⚠️ High | Resource contention |
| 6 | Rate limiting | ⚠️ High | Service bans |
| 7 | Health checks | 📋 Medium | Poor monitoring |
| 8 | Graceful shutdown | 📋 Medium | Failed requests |

---

## How to Apply

**For each file with external calls:**

1. Check circuit breaker wraps critical services
2. Verify retry logic uses exponential backoff
3. Check all external calls have timeout
4. Look for fallback on non-critical failures
5. Verify separate clients for different service criticality
6. Check rate limiting on high-volume calls
7. Verify health check endpoints exist
8. Check graceful shutdown on SIGTERM

**Example output:**

```
📁 File: internal/services/gateway_service.go

🚨 Check #1 Failed: No circuit breaker (Line 45)
   Code: http.Post(gatewayURL, ...)
   Fix: Wrap with gobreaker circuit breaker

🚨 Check #2 Failed: No retry logic (Line 67)
   Code: Single API call without retry
   Fix: Add exponential backoff retry

⚠️  Check #4 Failed: No fallback (Line 89)
   Code: Logo fetch failure fails request
   Fix: Return default logo on error

✅ Check #3 Passed: Timeout set (5s)
✅ Check #6 Passed: Rate limiter configured
```
