# Eventing Framework Checks

## Overview

Validates event-driven architecture patterns to prevent schema mismatches, version conflicts, missing handlers, and tracing issues.

**Load when:** PR modifies event publishers, subscribers, or event schemas

**Total Checks:** 6

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 3
- 📋 Medium: 1

---

## Check 1: Event Schema Validation on Publish 🚨 CRITICAL

### What to Check

Events must be validated against schema before publishing to prevent downstream failures.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No schema validation
func PublishTerminalCreatedEvent(terminal *Terminal) error {
    event := map[string]interface{}{
        "terminal_id": terminal.ID,
        "merchant_id": terminal.MerchantID,
        // ❌ Typo: should be "created_at"
        "createdat": terminal.CreatedAt,
        // ❌ Missing required field: "status"
    }

    // ❌ Publishes invalid event - downstream consumers crash!
    return eventBus.Publish("terminal.created", event)
}
```

**Problem:**
- Invalid events published
- Downstream consumers fail parsing
- Cascading failures across services

### Good Pattern ✅

```go
// CORRECT: Define event schema
type TerminalCreatedEvent struct {
    TerminalID  string    `json:"terminal_id" validate:"required,len=14"`
    MerchantID  string    `json:"merchant_id" validate:"required,len=14"`
    Status      string    `json:"status" validate:"required,oneof=active inactive"`
    CreatedAt   time.Time `json:"created_at" validate:"required"`
    Version     string    `json:"version" validate:"required"`
}

// Validate before publishing
func PublishTerminalCreatedEvent(terminal *Terminal) error {
    event := TerminalCreatedEvent{
        TerminalID: terminal.ID,
        MerchantID: terminal.MerchantID,
        Status:     terminal.Status,
        CreatedAt:  terminal.CreatedAt,
        Version:    "v1",
    }

    // ✅ Validate schema
    if err := validator.Validate(event); err != nil {
        logger.Error(ctx, "event_validation_failed", "error", err)
        return fmt.Errorf("invalid event schema: %w", err)
    }

    // ✅ Publish only after validation
    return eventBus.Publish("terminal.created", event)
}
```

### Detection Strategy

```bash
# Find event publish calls
grep -n "Publish(" internal/events/*.go pkg/eventing/*.go

# For each Publish, check:
# 1. Event struct has validation tags
# 2. Validation called before Publish
# 3. Required fields present
```

### Flag Conditions

Flag if:
- `Publish()` with `map[string]interface{}` (untyped)
- Event struct without validation tags
- No validation call before `Publish()`
- Event created inline without struct

### Severity

🚨 **Critical** - Downstream service crashes, cascading failures

### Reference

Based on terminals event publishing patterns

---

## Check 2: Event Type Assertions Safety 🚨 CRITICAL

### What to Check

Event consumers must safely handle type assertions to prevent panics.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Unsafe type assertion
func HandleTerminalEvent(ctx context.Context, event interface{}) error {
    // ❌ Panics if event is wrong type!
    terminalEvent := event.(TerminalCreatedEvent)

    // ❌ Panics if field is nil or wrong type!
    merchantId := terminalEvent.Metadata["merchant_id"].(string)

    processTerminal(merchantId)
    return nil
}
```

**Problem:**
- Panic on unexpected event type
- Consumer crashes
- All events stop processing

### Good Pattern ✅

```go
// CORRECT: Safe type assertion
func HandleTerminalEvent(ctx context.Context, event interface{}) error {
    // ✅ Safe type assertion with ok check
    terminalEvent, ok := event.(TerminalCreatedEvent)
    if !ok {
        logger.Error(ctx, "unexpected_event_type",
            "expected", "TerminalCreatedEvent",
            "got", fmt.Sprintf("%T", event))
        return fmt.Errorf("unexpected event type: %T", event)
    }

    // ✅ Safe field access with nil check
    merchantId, ok := terminalEvent.Metadata["merchant_id"].(string)
    if !ok || merchantId == "" {
        logger.Error(ctx, "invalid_merchant_id")
        return errors.New("merchant_id missing or invalid")
    }

    return processTerminal(merchantId)
}

// PATTERN 2: Unmarshal from bytes
func HandleEventFromBytes(ctx context.Context, payload []byte) error {
    var event TerminalCreatedEvent

    // ✅ Unmarshal with error check
    if err := json.Unmarshal(payload, &event); err != nil {
        logger.Error(ctx, "event_unmarshal_failed", "error", err)
        return fmt.Errorf("failed to unmarshal event: %w", err)
    }

    // ✅ Validate unmarshaled event
    if err := validator.Validate(event); err != nil {
        logger.Error(ctx, "event_validation_failed", "error", err)
        return fmt.Errorf("invalid event: %w", err)
    }

    return processEvent(event)
}
```

### Severity

🚨 **Critical** - Consumer crashes, event processing stops

---

## Check 3: Event Versioning for Schema Evolution ⚠️ HIGH

### What to Check

Events must include version field to support backward-compatible changes.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No version field
type TerminalCreatedEvent struct {
    TerminalID string `json:"terminal_id"`
    MerchantID string `json:"merchant_id"`
    // ❌ No version field!
}

// Later, schema changes...
type TerminalCreatedEvent struct {
    TerminalID string `json:"terminal_id"`
    MerchantID string `json:"merchant_id"`
    GatewayID  string `json:"gateway_id"`  // New field added
    // ❌ Old consumers break!
}
```

**Problem:**
- Can't evolve schema safely
- Old consumers break on new fields
- No way to identify event version

### Good Pattern ✅

```go
// CORRECT: Version-aware events
type TerminalCreatedEventV1 struct {
    Version    string `json:"version"`  // ✅ Always "v1"
    TerminalID string `json:"terminal_id"`
    MerchantID string `json:"merchant_id"`
}

// New version with additional field
type TerminalCreatedEventV2 struct {
    Version    string `json:"version"`  // ✅ Always "v2"
    TerminalID string `json:"terminal_id"`
    MerchantID string `json:"merchant_id"`
    GatewayID  string `json:"gateway_id"`  // New field
}

// Consumer handles multiple versions
func HandleTerminalCreated(ctx context.Context, payload []byte) error {
    // Parse version first
    var versionOnly struct {
        Version string `json:"version"`
    }

    if err := json.Unmarshal(payload, &versionOnly); err != nil {
        return err
    }

    // ✅ Route based on version
    switch versionOnly.Version {
    case "v1":
        var event TerminalCreatedEventV1
        json.Unmarshal(payload, &event)
        return handleV1(event)

    case "v2":
        var event TerminalCreatedEventV2
        json.Unmarshal(payload, &event)
        return handleV2(event)

    default:
        logger.Warn(ctx, "unknown_event_version", "version", versionOnly.Version)
        // ✅ Graceful handling of unknown versions
        return nil
    }
}
```

### Detection Strategy

Look for event struct definitions without `version` or `schema_version` field.

### Severity

⚠️ **High** - Breaking changes, backward compatibility issues

---

## Check 4: Event Topic Naming Convention ⚠️ HIGH

### What to Check

Event topics must follow consistent naming pattern for discoverability.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Inconsistent naming
eventBus.Publish("terminalCreated", event)       // camelCase
eventBus.Publish("terminal_updated", event)      // snake_case
eventBus.Publish("TERMINAL-DELETED", event)      // UPPER-KEBAB
eventBus.Publish("terms", event)                 // Abbreviation
```

**Problem:**
- Hard to find related events
- Subscription errors from typos
- Unclear event ownership

### Good Pattern ✅

```go
// CORRECT: Consistent naming convention
// Pattern: {domain}.{entity}.{action}.{version}

const (
    EventTerminalCreated  = "terminals.terminal.created.v1"
    EventTerminalUpdated  = "terminals.terminal.updated.v1"
    EventTerminalDeleted  = "terminals.terminal.deleted.v1"

    EventMerchantOnboarded = "terminals.merchant.onboarded.v1"
    EventGatewayConfigured = "terminals.gateway.configured.v1"
)

// Usage
eventBus.Publish(EventTerminalCreated, event)
```

**Benefits:**
- Clear domain ownership (terminals)
- Grouped by entity (terminal, merchant)
- Action is verb (created, updated)
- Versioned for evolution

### Severity

⚠️ **High** - Maintainability, subscription errors

---

## Check 5: Context Propagation in Events ⚠️ HIGH

### What to Check

Events must preserve tracing context (request ID, correlation ID) for debugging.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Lost trace ID
func PublishTerminalCreated(terminal *Terminal) error {
    event := TerminalCreatedEvent{
        TerminalID: terminal.ID,
        // ❌ No trace ID or request ID!
    }

    // ❌ Can't trace this event back to original request
    return eventBus.Publish("terminal.created", event)
}
```

**Problem:**
- Can't correlate events with requests
- Debugging cross-service flows impossible
- Lost distributed tracing

### Good Pattern ✅

```go
// CORRECT: Preserve context in event
type BaseEvent struct {
    RequestID     string    `json:"request_id"`
    CorrelationID string    `json:"correlation_id"`
    TraceID       string    `json:"trace_id"`
    Timestamp     time.Time `json:"timestamp"`
}

type TerminalCreatedEvent struct {
    BaseEvent               // ✅ Embedded tracing fields
    TerminalID string `json:"terminal_id"`
    MerchantID string `json:"merchant_id"`
}

// Extract context when publishing
func PublishTerminalCreated(ctx *gin.Context, terminal *Terminal) error {
    event := TerminalCreatedEvent{
        BaseEvent: BaseEvent{
            RequestID:     ctx.GetString(constants.RequestID),      // ✅ From context
            CorrelationID: ctx.GetString(constants.CorrelationID),
            TraceID:       ctx.GetString(constants.TraceID),
            Timestamp:     time.Now(),
        },
        TerminalID: terminal.ID,
        MerchantID: terminal.MerchantID,
    }

    return eventBus.Publish("terminal.created", event)
}

// Consumer extracts and continues trace
func HandleTerminalCreated(ctx context.Context, event TerminalCreatedEvent) error {
    // ✅ Create new context with trace info
    newCtx := context.WithValue(ctx, constants.RequestID, event.RequestID)
    newCtx = context.WithValue(newCtx, constants.TraceID, event.TraceID)

    logger.Info(newCtx, "processing_terminal_created",
        "terminal_id", event.TerminalID,
        "trace_id", event.TraceID)  // ✅ Logged for correlation

    return processTerminal(newCtx, event)
}
```

### Severity

⚠️ **High** - Poor observability, debugging difficult

---

## Check 6: Subscriber Registration Completeness 📋 MEDIUM

### What to Check

All event subscribers must be registered at application startup.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Subscriber not registered
// internal/handlers/terminal_handler.go
func HandleTerminalCreated(event TerminalCreatedEvent) error {
    // Handler exists but never called!
    return processTerminal(event)
}

// main.go
func main() {
    // ❌ Handler exists but not registered
    // Events are published but nothing listens!
}
```

**Problem:**
- Events published but not processed
- Silent failures
- Missing business logic execution

### Good Pattern ✅

```go
// CORRECT: Central subscriber registration

// internal/events/subscribers.go
func RegisterAllSubscribers(eventBus *EventBus) {
    // ✅ All subscribers registered in one place
    eventBus.Subscribe(EventTerminalCreated, handlers.HandleTerminalCreated)
    eventBus.Subscribe(EventTerminalUpdated, handlers.HandleTerminalUpdated)
    eventBus.Subscribe(EventTerminalDeleted, handlers.HandleTerminalDeleted)

    eventBus.Subscribe(EventMerchantOnboarded, handlers.HandleMerchantOnboarded)

    logger.Info("registered_event_subscribers",
        "count", eventBus.SubscriberCount())
}

// main.go
func main() {
    eventBus := NewEventBus()

    // ✅ Register all subscribers at startup
    events.RegisterAllSubscribers(eventBus)

    // Start application
}

// PATTERN 2: Auto-registration with reflection
type EventHandler interface {
    EventType() string
    Handle(ctx context.Context, event interface{}) error
}

func RegisterHandlers(eventBus *EventBus, handlers []EventHandler) {
    for _, handler := range handlers {
        eventBus.Subscribe(handler.EventType(), handler.Handle)
        logger.Info("registered_handler", "event", handler.EventType())
    }
}
```

### Severity

📋 **Medium** - Silent failures, missing functionality

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | Schema validation | 🚨 Critical | Downstream crashes |
| 2 | Type assertion safety | 🚨 Critical | Consumer panics |
| 3 | Event versioning | ⚠️ High | Breaking changes |
| 4 | Topic naming | ⚠️ High | Maintainability issues |
| 5 | Context propagation | ⚠️ High | Lost tracing |
| 6 | Subscriber registration | 📋 Medium | Silent failures |

---

## How to Apply

**For each file matching** `internal/events/*`, `pkg/eventing/*`:

1. Check event publish has schema validation
2. Verify type assertions use `, ok` pattern
3. Check event structs have version field
4. Validate topic naming follows convention
5. Verify events include trace context
6. Check all handlers are registered

**Example output:**

```
📁 File: internal/events/terminal_publisher.go

🚨 Check #1 Failed: No schema validation (Line 45)
   Code: eventBus.Publish(topic, map[string]interface{}{...})
   Fix: Define typed event struct with validation

🚨 Check #2 Failed: Unsafe type assertion (Line 67)
   Code: event.(TerminalEvent)
   Fix: Use event, ok := event.(TerminalEvent)

⚠️  Check #3 Failed: No version field (Line 23)
   Code: type TerminalEvent struct {...}
   Fix: Add Version string `json:"version"`

⚠️  Check #5 Failed: No trace context (Line 45)
   Code: Event missing RequestID, TraceID fields
   Fix: Embed BaseEvent with tracing fields

✅ Check #4 Passed: Consistent topic naming
✅ Check #6 Passed: All handlers registered in RegisterSubscribers()
```
