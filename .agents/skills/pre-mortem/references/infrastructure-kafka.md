# Kafka Infrastructure Checks

## Overview

Validates Kafka consumer and producer patterns to prevent message loss, duplicate processing, service crashes, and cascading failures.

**Load when:** PR modifies `internal/kafka/*`, `worker/kafka/*`, or Kafka-related code

**Total Checks:** 8

**Severity Distribution:**
- 🚨 Critical: 4
- ⚠️ High: 3
- 📋 Medium: 1

---

## Check 1: Panic Recovery in Consumer Handlers 🚨 CRITICAL

### What to Check

Kafka message handlers and goroutines must have panic recovery to prevent consumer crashes.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Goroutine without panic recovery
go (func() {
    if err1 := k.HandleMessage(ctx, msg); err1 != nil {
        msg.Nack()
    } else {
        msg.Ack()
    }
})()  // ❌ Panic in HandleMessage crashes worker!

// ANTI-PATTERN: Handler without recovery
func (h *Handler) Handle(ctx context.Context, payload []byte) error {
    body := make(map[string]interface{})
    json.Unmarshal(payload, &body)

    merchantId := body["merchant_id"].(string)  // ❌ Panics if missing!
    // ... process
}
```

**Problem:**
- Panic crashes entire consumer/worker
- All topics stop processing
- Manual restart required

### Good Pattern ✅

```go
// CORRECT: Goroutine with panic recovery
go (func() {
    defer func() {
        if r := recover(); r != nil {
            logger.Error(ctx, "panic in kafka handler", "error", r)
            msg.Nack()  // Requeue message
        }
    }()

    if err1 := k.HandleMessage(ctx, msg); err1 != nil {
        msg.Nack()
    } else {
        msg.Ack()
    }
})()

// CORRECT: Handler with panic recovery
func (h *Handler) Handle(ctx context.Context, payload []byte) error {
    defer func() {
        if r := recover(); r != nil {
            logger.Error(ctx, "panic in handle", "error", r)
        }
    }()

    body := make(map[string]interface{})
    if err := json.Unmarshal(payload, &body); err != nil {
        return err
    }

    merchantId, ok := body["merchant_id"].(string)
    if !ok {
        return errors.New("merchant_id missing or invalid")
    }
    // ... process
}
```

### Severity

🚨 **Critical** - Worker crashes, all message processing stops

---

## Check 2: Dead Letter Queue (DLQ) for Failed Messages 🚨 CRITICAL

### What to Check

Failed messages must be sent to DLQ instead of being lost or retried indefinitely.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Failed messages just logged
case failedMsg := <-k.baseConsumer.Failed:
    logger.Error(ctx, "message failed",
        "topic", failedMsg.Topic,
        "offset", failedMsg.Offset)
    // ❌ Message is LOST!
```

### Good Pattern ✅

```go
// CORRECT: Send failed messages to DLQ
func (h *Handler) Handle(ctx context.Context, payload []byte) error {
    terminalId, err := processMessage(payload)
    if err != nil {
        // Push to DLQ for later investigation
        PushToDLQ(ctx, payload, err)
        logger.Error(ctx, "message_failed_pushed_to_dlq", "error", err)
        return nil  // ACK message (it's in DLQ now)
    }
    return nil
}

func PushToDLQ(ctx *gin.Context, payload []byte, originalError error) {
    dlqTopic := fmt.Sprintf("dlq_%s_%s", config.Mode, originalTopic)
    err := producer.DefaultProducer.Produce(ctx, dlqTopic, string(payload))
    if err != nil {
        logger.Error(ctx, "dlq_push_failed", "error", err)
        // Store locally as backup
        saveToLocalDLQ(payload, originalError)
    }
}
```

### Severity

🚨 **Critical** - Data loss, failed messages unrecoverable

---

## Check 3: Idempotency via Distributed Locks or Message ID 🚨 CRITICAL

### What to Check

Kafka consumers must prevent duplicate processing of the same message.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No idempotency check
func (h *Handler) Handle(ctx context.Context, payload []byte) error {
    request := parseRequest(payload)

    // ❌ Processes same message multiple times if delivered twice!
    createTerminal(request.TerminalId)
    return nil
}
```

### Good Pattern ✅

```go
// PATTERN 1: Distributed lock with timestamp check
func (h *Handler) Handle(ctx context.Context, payload []byte) error {
    request := parseRequest(payload)

    // Check if already processed (timestamp-based)
    if isAlreadyProcessed(request.MessageId) {
        logger.Info(ctx, "message_already_processed", "id", request.MessageId)
        return nil
    }

    // Acquire lock to prevent concurrent processing
    mutex := NewMutex(ctx, request.MessageId, 60*time.Second)
    defer mutex.Unlock()

    if err := mutex.Lock(); err != nil {
        logger.Info(ctx, "already_processing", "id", request.MessageId)
        return nil  // Another instance processing
    }

    createTerminal(request.TerminalId)
    markAsProcessed(request.MessageId)
    return nil
}

// PATTERN 2: Database unique constraint
func createTerminal(terminalId string) error {
    // Database constraint prevents duplicates
    // UNIQUE(terminal_id)
    err := db.Create(&Terminal{ID: terminalId})
    if errors.Is(err, gorm.ErrDuplicatedKey) {
        return nil  // Already exists, idempotent
    }
    return err
}
```

### Severity

🚨 **Critical** - Duplicate processing, data corruption

---

## Check 4: Producer Panic Cascade Prevention 🚨 CRITICAL

### What to Check

Kafka producers must not crash the entire service after max panic attempts.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Re-panic after max attempts
const maxPanics = 5

func handleProducerStatus(producer *kafka.KafkaProducerQueue, doneCh chan struct{}) {
    for i := 1; i <= maxPanics; i++ {
        handleProducerStatusHelper(producer, doneCh)
    }

    // ❌ Crashes entire service!
    panic("handleProducerStatus reached max panics")
}
```

### Good Pattern ✅

```go
// CORRECT: Circuit breaker instead of panic
func handleProducerStatus(producer *kafka.KafkaProducerQueue, doneCh chan struct{}) {
    failureCount := 0
    const maxFailures = 5

    for {
        select {
        case err := <-producer.Producer.Errors():
            failureCount++
            logger.Error(ctx, "producer_error", "error", err, "count", failureCount)

            if failureCount >= maxFailures {
                // Open circuit, stop producing
                logger.Error(ctx, "producer_circuit_open", "failures", failureCount)
                circuitOpen = true

                // Wait before retry
                time.Sleep(30 * time.Second)
                failureCount = 0  // Reset
            }
        case <-doneCh:
            return
        }
    }
}
```

### Severity

🚨 **Critical** - Service crashes, complete outage

---

## Check 5: Context Propagation in Handlers ⚠️ HIGH

### What to Check

Kafka handlers should preserve context from consumer, not create new context.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Creates new context, loses trace ID
func (h *Handler) Handle(_ context.Context, payload []byte) error {
    ctx := &gin.Context{}  // ❌ New context - no trace ID!
    ctx.Set(constants.TaskID, utils.NewUUIDAsString())  // ❌ Regenerated!

    processMessage(ctx, payload)
}
```

### Good Pattern ✅

```go
// CORRECT: Use provided context
func (h *Handler) Handle(ctx context.Context, payload []byte) error {
    // Use provided context with trace ID
    processMessage(ctx, payload)
}

// If gin.Context needed, convert properly
func (h *Handler) Handle(ctx context.Context, payload []byte) error {
    ginCtx := convertToGinContext(ctx)  // Preserve trace ID
    processMessage(ginCtx, payload)
}
```

### Severity

⚠️ **High** - Lost tracing, debugging difficult

---

## Check 6: Topic Naming from Config (Not Hardcoded) ⚠️ HIGH

### What to Check

Topic names should come from configuration, not hardcoded strings.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Hardcoded topic names
func publishEvent(ctx *gin.Context, event string) {
    producer.Produce(ctx, "prod-merchant-payments-enabled", event)  // ❌ Hardcoded!
}
```

### Good Pattern ✅

```go
// CORRECT: Topic from config
func publishEvent(ctx *gin.Context, event string) {
    topic := bootstrap.Config.Topics.PaymentEnabled
    producer.Produce(ctx, topic, event)
}

// Config structure
type Topics struct {
    PaymentEnabled    string `toml:"paymentEnabled"`
    TerminalSync      string `toml:"terminalSync"`
}
```

### Severity

⚠️ **High** - Hard to maintain, environment issues

---

## Check 7: Consumer Group Naming Convention ⚠️ HIGH

### What to Check

Consumer groups should follow naming pattern and be configured (not hardcoded).

### Bad Pattern ❌

```go
// ANTI-PATTERN: Unclear consumer group
consumerGroup = "group1"  // ❌ What is this?
```

### Good Pattern ✅

```go
// CORRECT: Descriptive naming pattern
// Pattern: {service-name}-{function}-group
consumerGroup = "terminals-onboarding-group"
consumerGroup = "pg-router-payment-notification-group"

// From config
consumerGroup = config.Kafka.ConsumerGroup
```

### Severity

⚠️ **High** - Message delivery issues, debugging hard

---

## Check 8: Error Handling in Message Unmarshaling 📋 MEDIUM

### What to Check

JSON unmarshaling errors should be handled, not ignored.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Unmarshal error ignored
body := make(map[string]interface{})
json.Unmarshal(payload, &body)  // ❌ Error ignored!
merchantId := body["merchant_id"].(string)  // Panics if unmarshal failed
```

### Good Pattern ✅

```go
// CORRECT: Check unmarshal error
body := make(map[string]interface{})
if err := json.Unmarshal(payload, &body); err != nil {
    logger.Error(ctx, "unmarshal_failed", "error", err)
    PushToDLQ(ctx, payload, err)
    return nil  // ACK message (it's in DLQ)
}

merchantId, ok := body["merchant_id"].(string)
if !ok {
    return errors.New("merchant_id missing")
}
```

### Severity

📋 **Medium** - Can cause panics, but usually caught by panic recovery

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | Panic recovery | 🚨 Critical | Worker crashes |
| 2 | DLQ for failures | 🚨 Critical | Data loss |
| 3 | Idempotency | 🚨 Critical | Duplicate processing |
| 4 | Producer panic | 🚨 Critical | Service crashes |
| 5 | Context propagation | ⚠️ High | Lost tracing |
| 6 | Topic from config | ⚠️ High | Hard to maintain |
| 7 | Consumer group naming | ⚠️ High | Delivery issues |
| 8 | Unmarshal errors | 📋 Medium | Potential panics |
