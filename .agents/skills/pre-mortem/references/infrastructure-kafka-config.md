# Kafka Connection & Configuration Checks

## Overview

Validates Kafka connection and configuration patterns in Razorpay services to prevent connection issues, security vulnerabilities, performance problems, and message delivery failures. Based on patterns from goutils, upidp, financial-data-service, metro, and event-driven architecture rules.

**Load when:** PR modifies Kafka client initialization, config files with Kafka settings, or bootstrap configuration

**Total Checks:** 10

**Severity Distribution:**
- 🚨 Critical: 3
- ⚠️ High: 4
- 📋 Medium: 3

---

## Razorpay Kafka Context

### Services Using Kafka

**Found in production code:**
- `goutils/kafka` - Standard Kafka client library (Sarama wrapper)
- `upidp` - UPI payment processing with Kafka events
- `financial-data-service` - Financial data streaming
- `metro` - Message broker abstraction
- `wallet` - Event-driven architecture
- `router`, `ledger`, `gc-order-management-service` - Event patterns

### Common Patterns Found

**Razorpay Standards:**
- **Library**: Shopify Sarama (Go Kafka client)
- **Compression**: Snappy (default across services)
- **Acks**: `sarama.WaitForAll` (all replicas must acknowledge)
- **Kafka Version**: 2.3.0+ (configurable via `DefaultKafkaVersion`)
- **Security**: SSL protocol with keystore authentication
- **Brokers**: From config/env, never hardcoded

**Connection Pattern** (goutils):
```go
config.Producer.RequiredAcks = sarama.WaitForAll
config.Producer.Compression = sarama.CompressionSnappy
config.Producer.Retry.Max = 3
```

---

## Check 1: Bootstrap Servers from Config (Not Hardcoded) 🚨 CRITICAL

### What to Check

Kafka broker addresses must be loaded from configuration or environment variables, never hardcoded in code.

### Razorpay Pattern (metro, wallet, goutils)

**Standard pattern across services:**
```go
"bootstrap.servers": strings.Join(config.Brokers, ",")
// OR
"bootstrap.servers": os.Getenv("KAFKA_BROKERS")
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: Hardcoded broker addresses
producer, err := kafka.NewProducer(&kafka.ConfigMap{
    "bootstrap.servers": "kafka1.prod.razorpay.com:9092,kafka2.prod.razorpay.com:9092",
    // ❌ Hardcoded! Can't switch environments or brokers
})

// ANTI-PATTERN: Hardcoded in Sarama config
config := sarama.NewConfig()
brokers := []string{
    "10.0.1.100:9092",
    "10.0.1.101:9092",
    "10.0.1.102:9092",
}
client, err := sarama.NewClient(brokers, config)
// ❌ IPs hardcoded! Can't change without code deploy
```

**Problem:**
- Can't switch between dev/stage/prod environments
- Broker migrations require code changes
- Kubernetes service discovery doesn't work
- No flexibility for disaster recovery

### Good Pattern ✅

```go
// CORRECT: Bootstrap servers from config (metro pattern)
type BrokerConfig struct {
    Brokers []string `mapstructure:"brokers"`  // ✅ From config
}

func NewKafkaProducer(bConfig BrokerConfig) (*kafka.Producer, error) {
    configMap := &kafka.ConfigMap{
        "bootstrap.servers": strings.Join(bConfig.Brokers, ","),  // ✅ From config
    }

    return kafka.NewProducer(configMap)
}

// TOML config - prod-live.toml
[kafka]
brokers = [
    "kafka-broker-1.prod.svc.cluster.local:9092",
    "kafka-broker-2.prod.svc.cluster.local:9092",
    "kafka-broker-3.prod.svc.cluster.local:9092",
]

// CORRECT: From environment variable (Razorpay pattern)
func NewProducerFromEnv() (*kafka.Producer, error) {
    brokers := os.Getenv("KAFKA_BROKERS")  // ✅ From env
    if brokers == "" {
        return nil, errors.New("KAFKA_BROKERS environment variable not set")
    }

    configMap := &kafka.ConfigMap{
        "bootstrap.servers": brokers,
    }

    return kafka.NewProducer(configMap)
}

// Kubernetes deployment
env:
- name: KAFKA_BROKERS
  value: "kafka-broker-1:9092,kafka-broker-2:9092,kafka-broker-3:9092"

// CORRECT: Sarama with config (goutils pattern)
type KafkaConfig struct {
    Brokers []string `json:"brokers"`
}

func NewSaramaProducer(cfg KafkaConfig) (sarama.SyncProducer, error) {
    config := sarama.NewConfig()
    // ... other config

    return sarama.NewSyncProducer(cfg.Brokers, config)  // ✅ From config
}
```

### Detection Strategy

```bash
# Find Kafka producer/consumer initialization
grep -n "NewProducer\|NewConsumer\|NewClient\|NewConsumerGroup" <pr_files>

# Check for hardcoded addresses (IP or env-specific hostnames)
grep -nE 'bootstrap\.servers.*[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' <pr_files>
grep -nE 'brokers.*kafka[0-9]+\.(prod|stage)' <pr_files>

# Flag files where bootstrap.servers is set without config/env source
# Note: grep -L does not work on piped input — use per-file loop
for file in <pr_files>; do
    if grep -qE '"bootstrap\.servers"' "$file" && \
       ! grep -qE 'os\.Getenv|config\.|cfg\.|\.Brokers|mapstructure|toml' "$file"; then
        echo "⚠️  $file: bootstrap.servers may not be loaded from config or env"
    fi
done
```

### Flag Conditions

Flag if:
- `bootstrap.servers` contains IP addresses or hostnames directly
- Broker list is an array literal with hostnames
- No `config.Brokers` or `os.Getenv("KAFKA_BROKERS")` usage
- Environment detection (dev/stage/prod) but broker URLs hardcoded

### Severity

🚨 **Critical** - Operational flexibility:
- Environment-specific deploys impossible
- Broker migration requires code change
- No disaster recovery flexibility
- Kubernetes service discovery broken

### References

**Razorpay production code:**
- `metro:pkg/messagebroker/kafka.go` - `strings.Join(bConfig.Brokers, ",")`
- `wallet:.cursor/rule-go-event-driven-architecture.mdc` - Standard pattern
- `vendor-experience:.agents/skills/cell-readiness` - `os.Getenv("KAFKA_BROKERS")`

---

## Check 2: Security Protocol (SSL/SASL) in Production 🚨 CRITICAL

### What to Check

Production Kafka connections must use SSL encryption and SASL authentication, not plain text.

### Razorpay Pattern (cc-address-service)

**SSL with keystore authentication:**
```scala
properties.setProperty("security.protocol", "SSL")
properties.setProperty("ssl.keystore.location", "/home/flink/certs/kafka.client.keystore.jks")
properties.setProperty("ssl.keystore.password", sys.env("KAFKA_KEYSTORE_PASSWORD"))
properties.setProperty("sasl.mechanism", "PLAIN")
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: No security protocol (plain text)
configMap := &kafka.ConfigMap{
    "bootstrap.servers": kafkaBrokers,
    // ❌ No security.protocol! Unencrypted traffic
}

// ANTI-PATTERN: Hardcoded credentials
configMap := &kafka.ConfigMap{
    "bootstrap.servers":   kafkaBrokers,
    "security.protocol":   "SASL_SSL",
    "sasl.mechanism":      "SCRAM-SHA-512",
    "sasl.username":       "kafka-user",      // ❌ Hardcoded!
    "sasl.password":       "SuperSecret123",  // ❌ Hardcoded!
}
```

**Problem:**
- Unencrypted Kafka traffic (MITM attacks)
- Credentials in code (Git history exposure)
- No certificate validation
- Compliance violations (PCI-DSS, SOC 2)

### Good Pattern ✅

```go
// CORRECT: SSL with environment-based credentials (Razorpay standard)
func NewSecureProducer(brokers string) (*kafka.Producer, error) {
    configMap := &kafka.ConfigMap{
        "bootstrap.servers": brokers,

        // ✅ SSL encryption
        "security.protocol": "SASL_SSL",

        // ✅ SASL authentication
        "sasl.mechanism":    "SCRAM-SHA-512",  // Or PLAIN
        "sasl.username":     os.Getenv("KAFKA_USERNAME"),
        "sasl.password":     os.Getenv("KAFKA_PASSWORD"),

        // ✅ SSL certificate validation
        "ssl.ca.location":   "/etc/ssl/certs/ca-certificates.crt",
    }

    return kafka.NewProducer(configMap)
}

// CORRECT: Sarama with TLS config
config := sarama.NewConfig()

// ✅ TLS enabled
tlsConfig := &tls.Config{
    InsecureSkipVerify: false,  // ✅ Verify certificates
}
config.Net.TLS.Enable = true
config.Net.TLS.Config = tlsConfig

// ✅ SASL authentication
config.Net.SASL.Enable = true
config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
config.Net.SASL.User = os.Getenv("KAFKA_USERNAME")
config.Net.SASL.Password = os.Getenv("KAFKA_PASSWORD")

// CORRECT: Environment-specific security
// dev-live.toml
[kafka]
security_protocol = "PLAINTEXT"  # ✅ OK for local dev

// prod-live.toml
[kafka]
security_protocol = "SASL_SSL"
sasl_mechanism = "SCRAM-SHA-512"
sasl_username = "${KAFKA_USERNAME}"  # ✅ From env/secrets
sasl_password = "${KAFKA_PASSWORD}"
ssl_ca_location = "/etc/ssl/certs/ca-bundle.crt"

// Kubernetes secret
apiVersion: v1
kind: Secret
metadata:
  name: kafka-credentials
type: Opaque
data:
  KAFKA_USERNAME: <base64>
  KAFKA_PASSWORD: <base64>
```

### Detection Strategy

```bash
# Find Kafka config
grep -n "ConfigMap\|sarama.NewConfig" <pr_files>

# Check for security protocol
grep -A 10 "ConfigMap\|NewConfig" <pr_files> | grep "security.protocol\|Net.TLS.Enable"

# Flag if production config without security
grep -l "prod\|live" <config_files> | xargs grep "kafka" | grep -v "SASL\|SSL\|TLS"
```

### Flag Conditions

Flag if:
- Production config without `security.protocol = SASL_SSL` or `TLS.Enable = true`
- Hardcoded SASL username/password
- `InsecureSkipVerify: true` in production
- No SSL/TLS configuration for production brokers

### Severity

🚨 **Critical** - Security vulnerability:
- Unencrypted Kafka traffic (eavesdropping)
- Credentials exposed in logs/code
- Man-in-the-middle attacks
- Compliance violations

### References

**Razorpay production code:**
- `cc-address-service:src/main/scala/cc/address/completeness/utils/KafkaConfig.scala` - SSL config

---

## Check 3: Producer Acks Configuration ⚠️ HIGH

### What to Check

Kafka producers must have appropriate `acks` setting based on durability requirements.

### Razorpay Standard (goutils)

**Pattern:** `sarama.WaitForAll` (all replicas must acknowledge)

```go
config.Producer.RequiredAcks = sarama.WaitForAll  // ✅ Razorpay standard
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: No acks config (defaults to leader-only)
configMap := &kafka.ConfigMap{
    "bootstrap.servers": brokers,
    // ❌ No 'acks' config! Defaults to acks=1 (leader only)
}
// Risk: Message loss if leader fails before replication

// ANTI-PATTERN: acks=0 for critical data
configMap := &kafka.ConfigMap{
    "bootstrap.servers": brokers,
    "acks":              "0",  // ❌ Fire-and-forget for payment events!
}
// ❌ No acknowledgment, potential message loss
```

**Problem:**
- `acks=0`: Fire-and-forget, no confirmation (message loss)
- `acks=1` (default): Leader only (message loss if leader crashes)
- Critical data (payments, transactions) needs `acks=all`

### Good Pattern ✅

```go
// CORRECT: acks=all for critical data (Razorpay standard)
configMap := &kafka.ConfigMap{
    "bootstrap.servers": brokers,
    "acks":              "all",  // ✅ All in-sync replicas must acknowledge
    "enable.idempotence": "true",  // ✅ Prevents duplicates
}

// CORRECT: Sarama config (goutils pattern)
config := sarama.NewConfig()
config.Producer.RequiredAcks = sarama.WaitForAll  // ✅ Razorpay standard
config.Producer.Idempotent = true                 // ✅ Exactly-once semantics

// CORRECT: Environment-specific acks
// High-throughput, non-critical (logs, analytics)
[kafka.producer.logs]
acks = "1"  # ✅ Leader-only for performance

// Critical data (payments, transactions)
[kafka.producer.payments]
acks = "all"  # ✅ All replicas for durability
enable_idempotence = true

// Decision matrix
// acks=0:   High throughput, OK with loss (metrics, logs)
// acks=1:   Balanced (analytics, non-critical events)
// acks=all: Critical data (payments, transactions, financial events) - Razorpay standard
```

### Detection Strategy

```bash
# Find producer initialization
grep -n "NewProducer\|SyncProducer\|sarama.NewConfig" <pr_files>

# Check for acks configuration near producer setup
grep -A 15 "NewProducer\|sarama.NewConfig" <pr_files> | grep -E "acks|RequiredAcks|WaitForAll"

# Flag files with financial topic producers but no acks setting
# Note: grep -L does not work on piped input — use per-file loop
for file in <pr_files>; do
    if grep -qE "NewProducer|NewSyncProducer|sarama.NewConfig" "$file" && \
       grep -qiE "payment|transaction|refund|financial" "$file" && \
       ! grep -qE '"acks"|RequiredAcks|WaitForAll' "$file"; then
        echo "⚠️  $file: Kafka producer for financial topic without acks configuration"
    fi
done
```

### Flag Conditions

Flag if:
- Producer for financial/critical topics without `acks` config
- `acks=0` or `acks=1` for payment/transaction events
- No `RequiredAcks` in Sarama config
- Comment mentions "critical" but acks not set to "all"

### Severity

⚠️ **High** - Data durability:
- Potential message loss during failures
- Financial data integrity risk
- Audit trail gaps
- Recovery challenges

### References

**Razorpay production code:**
- `goutils:tracing/examples/kafka/utils/kafka.go` - `RequiredAcks = sarama.WaitForAll`
- Standard across all Razorpay services for critical data

---

## Check 4: Compression Type Configuration ⚠️ HIGH

### What to Check

Kafka producers should use compression (Snappy or LZ4) to reduce bandwidth and storage costs.

### Razorpay Standard (upidp, financial-data-service)

**Pattern:** Snappy compression (default)

```go
// upidp, financial-data-service comments:
// "Default kafka compression to handle data transfer costs and better storage
//  on kafka side. We are enabling with SnappyCompression."
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: No compression config (uncompressed)
configMap := &kafka.ConfigMap{
    "bootstrap.servers": brokers,
    // ❌ No compression! Wastes bandwidth and storage
}

// 1MB message → 1MB transferred, 1MB stored
// With 1M messages/day → 1TB/day
```

### Good Pattern ✅

```go
// CORRECT: Snappy compression (Razorpay standard)
configMap := &kafka.ConfigMap{
    "bootstrap.servers":  brokers,
    "compression.type":   "snappy",  // ✅ Razorpay standard (fast + good ratio)
}

// CORRECT: Sarama config (goutils pattern)
config := sarama.NewConfig()
config.Producer.Compression = sarama.CompressionSnappy  // ✅ Razorpay standard

// Compression options:
// - snappy: Fast, good compression (Razorpay default)
// - lz4:    Very fast, decent compression
// - gzip:   Slower, better compression (for cold data)
// - zstd:   Best compression, moderate CPU (Kafka 2.1+)

// CORRECT: Config-based compression
[kafka.producer]
compression_type = "snappy"  # ✅ Razorpay standard

// Performance benefit:
// 1MB message → ~300KB compressed (3x reduction)
// 1M messages/day → 300GB/day (vs 1TB uncompressed)
```

### Severity

⚠️ **High** - Cost and performance:
- Higher bandwidth costs (3-5x)
- More Kafka storage needed
- Slower cross-datacenter replication
- Network congestion under load

### References

**Razorpay production code:**
- `upidp:pkg/publisher/broker/kafka/kafka.go` - Snappy compression default
- `financial-data-service:pkg/queue/kafka/kafkaproducer.go` - Snappy standard

---

## Check 5: Producer Retries and Timeout ⚠️ HIGH

### What to Check

Producers must have retry configuration with appropriate timeouts for transient failures.

### Razorpay Pattern (goutils)

```go
config.Producer.Retry.Max = 3
// "How long to wait for the cluster to settle between retries"
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: No retry config (defaults to 0)
configMap := &kafka.ConfigMap{
    "bootstrap.servers": brokers,
    // ❌ No retries! Fails immediately on transient errors
}

// ANTI-PATTERN: Infinite retries
configMap := &kafka.ConfigMap{
    "bootstrap.servers": brokers,
    "retries":           "999999",  // ❌ Blocks forever!
}
```

### Good Pattern ✅

```go
// CORRECT: Retry with timeout (Razorpay pattern)
configMap := &kafka.ConfigMap{
    "bootstrap.servers":       brokers,
    "retries":                 "3",       // ✅ Retry up to 3 times
    "retry.backoff.ms":        "100",     // ✅ 100ms between retries
    "request.timeout.ms":      "30000",   // ✅ 30s timeout
    "delivery.timeout.ms":     "120000",  // ✅ 2min total delivery timeout
}

// CORRECT: Sarama config (goutils pattern)
config := sarama.NewConfig()
config.Producer.Retry.Max = 3                              // ✅ Max retries
config.Producer.Timeout = 10 * time.Second                 // ✅ Request timeout
config.Producer.Return.Errors = true                       // ✅ Report errors
config.Producer.Return.Successes = true                    // ✅ Report successes
config.Metadata.Retry.Backoff = 100 * time.Millisecond    // ✅ Backoff
```

### Severity

⚠️ **High** - Reliability:
- Fail on transient network issues
- Higher error rates
- Poor user experience

---

## Check 6: Max In-Flight Requests (Ordering) ⚠️ HIGH

### What to Check

Producers requiring message ordering must set `max.in.flight.requests.per.connection = 1`.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No max in-flight config (defaults to 5)
configMap := &kafka.ConfigMap{
    "bootstrap.servers": brokers,
    "acks":              "all",
    // ❌ No max.in.flight.requests! Messages can arrive out-of-order on retry
}

// Scenario:
// Send: msg1, msg2, msg3
// msg2 fails, retried
// Kafka receives: msg1, msg3, msg2 ❌ Out of order!
```

### Good Pattern ✅

```go
// CORRECT: Strict ordering guarantee
configMap := &kafka.ConfigMap{
    "bootstrap.servers": brokers,
    "acks":              "all",
    "max.in.flight.requests.per.connection": "1",  // ✅ Only 1 request at a time
    "enable.idempotence": "true",                  // ✅ Idempotence ensures ordering
}

// CORRECT: Idempotent producer (Kafka 0.11+)
// enable.idempotence = true automatically sets:
// - max.in.flight.requests = 5 (but maintains order via sequence numbers)
// - acks = all
// - retries = MAX_INT

configMap := &kafka.ConfigMap{
    "bootstrap.servers":   brokers,
    "enable.idempotence": "true",  // ✅ Maintains order + prevents duplicates
}

// Decision:
// Strict ordering needed? → max.in.flight = 1
// Order + performance?   → enable.idempotence = true
// No order requirement?  → max.in.flight = 5 (default, better throughput)
```

### Severity

⚠️ **High** - Data consistency:
- Out-of-order message processing
- State corruption
- Incorrect business logic execution

---

## Check 7: Consumer Auto Offset Reset ⚠️ HIGH

### What to Check

Kafka consumers must explicitly set `auto.offset.reset` behavior.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No auto.offset.reset config (defaults to "latest")
config := sarama.NewConfig()
// ❌ Defaults to latest! New consumer skips existing messages
```

**Problem:**
- `latest` (default): Consumer skips all existing messages (data loss for new consumers)
- `earliest`: Reprocesses all messages from beginning (duplicate processing risk)
- No explicit choice = production surprise

### Good Pattern ✅

```go
// CORRECT: Explicit auto.offset.reset
configMap := &kafka.ConfigMap{
    "bootstrap.servers":   brokers,
    "group.id":            "payment-processor",
    "auto.offset.reset":   "earliest",  // ✅ Explicit: start from beginning
}

// CORRECT: Sarama config
config := sarama.NewConfig()
config.Consumer.Offsets.Initial = sarama.OffsetOldest  // ✅ Explicit: earliest
// OR
config.Consumer.Offsets.Initial = sarama.OffsetNewest  // ✅ Explicit: latest

// Decision matrix:
// earliest: New consumer must process all historical data (default for critical)
// latest:   New consumer only needs new messages (metrics, logs)
```

### Severity

⚠️ **High** - Data processing:
- Skipped messages for new consumers
- Duplicate processing
- Inconsistent behavior across environments

---

## Check 8: Session Timeout vs Heartbeat Interval 📋 MEDIUM

### What to Check

Consumer `session.timeout.ms` should be > 3x `heartbeat.interval.ms` to prevent false rebalances.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Session timeout too close to heartbeat
configMap := &kafka.ConfigMap{
    "session.timeout.ms":  "6000",   // 6s
    "heartbeat.interval.ms": "5000", // 5s  ❌ Too close! < 3x
}
// Risk: Network delay → missed heartbeat → rebalance
```

### Good Pattern ✅

```go
// CORRECT: 3x rule (Kafka recommendation)
configMap := &kafka.ConfigMap{
    "session.timeout.ms":    "30000",  // 30s
    "heartbeat.interval.ms": "3000",   // 3s  ✅ 10x safety margin
}

// Razorpay typical values (based on goutils comments):
// - session.timeout: 30s (time before rebalance triggered)
// - heartbeat.interval: 3s (how often to send heartbeats)
// - max.poll.interval: 5min (max time between poll() calls)
```

### Severity

📋 **Medium** - Stability:
- Frequent consumer rebalances
- Processing delays
- Duplicate message processing

---

## Check 9: Enable Idempotence for Retry-Enabled Producers 📋 MEDIUM

### What to Check

Producers that configure `retries > 0` without idempotence can produce duplicate messages on network ack loss. This is distinct from Check 3 (which focuses on acks for durability): idempotence specifically prevents the **duplicate-on-retry** race condition.

> **Scope:** Only flag when the file both enables retries AND targets financial/critical topics.
> Check 3 covers acks durability. Check 9 covers duplicate prevention.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Retries enabled without idempotence
configMap := &kafka.ConfigMap{
    "bootstrap.servers": brokers,
    "acks":              "all",   // ✅ Durability covered (Check 3)
    "retries":           "3",     // ✅ Retries configured (Check 5)
    // ❌ Missing enable.idempotence — a retry after lost ack creates a duplicate
}
```

### Good Pattern ✅

```go
// CORRECT: Retries + idempotence = exactly-once delivery
configMap := &kafka.ConfigMap{
    "bootstrap.servers":  brokers,
    "acks":               "all",
    "retries":            "3",
    "enable.idempotence": "true",  // ✅ Deduplicates retried messages
}

// CORRECT: Sarama
config := sarama.NewConfig()
config.Producer.Retry.Max = 3
config.Producer.Idempotent = true  // ✅ Pairs with retries
config.Producer.RequiredAcks = sarama.WaitForAll
```

### Detection Strategy

```bash
# Flag producers with retries but no idempotence in critical-topic files
for file in <pr_files>; do
    if grep -qE '"retries"|Retry\.Max' "$file" && \
       grep -qiE "payment|transaction|refund|financial" "$file" && \
       ! grep -qE '"enable\.idempotence"|Idempotent\s*=' "$file"; then
        echo "📋 $file: Producer has retries but no idempotence — duplicates possible on ack loss"
    fi
done
```

### Severity

📋 **Medium** - Data integrity (downgraded from High to avoid overlap with Check 3):
- Duplicate message processing on ack loss + retry
- Incorrect aggregations
- Financial reconciliation requires idempotent consumers as additional guard

---

## Check 10: Connection Pool / Producer Reuse 📋 MEDIUM

### What to Check

Kafka producers should be reused (singleton), not created per message.

### Bad Pattern ❌

```go
// ANTI-PATTERN: New producer per message
func PublishEvent(topic string, message []byte) error {
    // ❌ New producer every time!
    producer, _ := kafka.NewProducer(&kafka.ConfigMap{
        "bootstrap.servers": brokers,
    })
    defer producer.Close()

    producer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{Topic: &topic},
        Value:          message,
    }, nil)
}
// Problem: Connection overhead, resource exhaustion
```

### Good Pattern ✅

```go
// CORRECT: Singleton producer (reuse)
var (
    producer *kafka.Producer
    once     sync.Once
)

func GetProducer() (*kafka.Producer, error) {
    var err error
    once.Do(func() {
        producer, err = kafka.NewProducer(&kafka.ConfigMap{
            "bootstrap.servers": os.Getenv("KAFKA_BROKERS"),
        })
    })
    return producer, err
}

func PublishEvent(topic string, message []byte) error {
    p, err := GetProducer()  // ✅ Reuses connection
    if err != nil {
        return err
    }

    return p.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{Topic: &topic},
        Value:          message,
    }, nil)
}

// CORRECT: Graceful shutdown
func Shutdown() {
    if producer != nil {
        producer.Flush(5000)  // Wait up to 5s for pending messages
        producer.Close()
    }
}
```

### Severity

📋 **Medium** - Performance:
- Connection overhead
- Resource exhaustion
- Slower message delivery

---

## Summary Table

| Check # | Pattern | Severity | Razorpay Pattern |
|---------|---------|----------|------------------|
| 1 | Bootstrap Servers from Config | 🚨 Critical | `strings.Join(config.Brokers, ",")` |
| 2 | Security Protocol (SSL/SASL) | 🚨 Critical | `security.protocol = SSL` |
| 3 | Producer Acks Configuration | ⚠️ High | `RequiredAcks = WaitForAll` |
| 4 | Compression Type | ⚠️ High | `CompressionSnappy` (standard) |
| 5 | Producer Retries & Timeout | ⚠️ High | `Retry.Max = 3` |
| 6 | Max In-Flight Requests | ⚠️ High | `max.in.flight = 1` for ordering |
| 7 | Consumer Auto Offset Reset | ⚠️ High | Explicit `earliest` or `latest` |
| 8 | Session Timeout vs Heartbeat | 📋 Medium | 30s session, 3s heartbeat |
| 9 | Enable Idempotence (when retries configured) | 📋 Medium | `Idempotent = true` + retries |
| 10 | Connection Pool / Reuse | 📋 Medium | Singleton producer pattern |

---

## How to Apply

**For each file matching** Kafka client initialization:

1. **Bootstrap Servers**: Check for config/env usage, not hardcoded
2. **Security**: Verify SSL/SASL in production configs
3. **Producer Acks**: Ensure `acks=all` for critical topics
4. **Compression**: Validate Snappy compression enabled
5. **Retries**: Check retry config with timeout
6. **Ordering**: Verify `max.in.flight` or idempotence for ordering needs
7. **Consumer Reset**: Explicit `auto.offset.reset`
8. **Timeouts**: Session timeout > 3x heartbeat
9. **Idempotence**: Enabled for exactly-once semantics
10. **Connection Reuse**: Singleton producer pattern

**Example output:**

```
📁 File: pkg/events/kafka_producer.go

🚨 Check #1 Failed: Hardcoded bootstrap servers (Line 23)
   Code: "bootstrap.servers": "kafka1.prod.com:9092"
   Fix: Load from config: strings.Join(config.Brokers, ",")

🚨 Check #2 Failed: No security protocol (Line 24)
   Fix: Add "security.protocol": "SASL_SSL" for production

⚠️  Check #3 Failed: No acks configuration (Line 25)
   Fix: Add "acks": "all" for durability

⚠️  Check #4 Failed: No compression (Line 26)
   Fix: Add "compression.type": "snappy" (Razorpay standard)

✅ Check #5 Passed: Retry config present
✅ Check #9 Passed: Idempotence enabled
```

---

## Integration with Razorpay Systems

### Goutils Library
- Standard Kafka wrapper used across services
- Enforces: `WaitForAll` acks, Snappy compression, retries
- Kafka version: 2.3.0+

### Event-Driven Architecture
- Cursor rules enforce producer/consumer patterns
- Standard across wallet, router, ledger, gc-order-management

### Security
- SSL/SASL required for production
- Keystore authentication pattern (cc-address-service)
- Credentials from environment/Kubestash

### Data Platform
- Kafka as CDC pipeline (Aurora → Kafka → Analytics)
- Compression critical for cross-DC replication
- High durability requirements (acks=all)
