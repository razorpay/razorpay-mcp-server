# Event Contract Validation Checks

## Overview

Validates event schemas for publisher-consumer compatibility, versioning, and backward compatibility across services.

**Load when:** PR modifies event publishers, consumers, or event schemas

**Total Checks:** 4

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 2

---

## Check 1: Event Schema Documentation 🚨 CRITICAL

### What to Check

Published events must have documented schema that consumers can reference.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Undocumented event schema
func PublishTerminalCreated(terminal *Terminal) {
    // ❌ Event structure not documented anywhere
    event := map[string]interface{}{
        "terminal_id": terminal.ID,
        "merchant_id": terminal.MerchantID,
        "created_at":  terminal.CreatedAt,
        // What other fields exist? Consumers don't know!
    }

    eventBus.Publish("terminal.created", event)
}
```

**Problem:**
- Consumers don't know expected schema
- Breaking changes not detected
- Impossible to validate contracts

### Good Pattern ✅

```go
// CORRECT: Documented event schema

// events/terminal_events.go
// TerminalCreatedEvent is published when a new terminal is created
// Topic: terminals.terminal.created.v1
// Publishers: terminal-service
// Consumers: notification-service, analytics-service
type TerminalCreatedEvent struct {
    // Schema version - always "v1" for this event
    Version string `json:"version" validate:"required,eq=v1"`

    // Terminal unique identifier
    TerminalID string `json:"terminal_id" validate:"required,len=14"`

    // Merchant who owns this terminal
    MerchantID string `json:"merchant_id" validate:"required,len=14"`

    // Organization ID
    OrgID string `json:"org_id" validate:"required,len=14"`

    // Terminal status: active, inactive, suspended
    Status string `json:"status" validate:"required,oneof=active inactive suspended"`

    // Gateway assigned to terminal
    Gateway string `json:"gateway" validate:"required"`

    // When terminal was created
    CreatedAt time.Time `json:"created_at" validate:"required"`

    // Context for tracing
    RequestID string `json:"request_id" validate:"required"`
    TraceID   string `json:"trace_id" validate:"required"`
}

// Publisher
func PublishTerminalCreated(ctx *gin.Context, terminal *Terminal) error {
    event := TerminalCreatedEvent{
        Version:    "v1",
        TerminalID: terminal.ID,
        MerchantID: terminal.MerchantID,
        OrgID:      terminal.OrgID,
        Status:     terminal.Status,
        Gateway:    terminal.Gateway,
        CreatedAt:  terminal.CreatedAt,
        RequestID:  ctx.GetString(constants.RequestID),
        TraceID:    ctx.GetString(constants.TraceID),
    }

    // ✅ Validate against schema
    if err := validator.Validate(event); err != nil {
        logger.Error(ctx, "event_validation_failed", "error", err)
        return fmt.Errorf("invalid event schema: %w", err)
    }

    return eventBus.Publish("terminals.terminal.created.v1", event)
}

// Consumer knows exact schema
func HandleTerminalCreated(ctx context.Context, payload []byte) error {
    var event TerminalCreatedEvent

    if err := json.Unmarshal(payload, &event); err != nil {
        return fmt.Errorf("failed to unmarshal event: %w", err)
    }

    // ✅ Validate consumed event
    if err := validator.Validate(event); err != nil {
        return fmt.Errorf("invalid event: %w", err)
    }

    return processTerminalCreated(event)
}
```

### Detection Strategy

```bash
# Find event publish calls
grep -n "Publish(" internal/events/*.go pkg/eventing/*.go

# For each Publish:
# 1. Check if event has documented struct
# 2. Verify struct has JSON tags
# 3. Check struct has validation tags
# 4. Look for schema documentation comments
```

### Flag Conditions

Flag if:
- Event published with `map[string]interface{}` instead of struct
- Event struct without documentation comments
- No validation tags on event fields
- Missing publishers/consumers documentation

### Severity

🚨 **Critical** - Contract violations, breaking changes undetected

### Reference

Based on terminals event patterns

---

## Check 2: Backward Compatibility on Schema Changes 🚨 CRITICAL

### What to Check

Event schema changes must be backward compatible - can't remove/rename fields.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Breaking change to event schema

// V1 (existing, consumed by multiple services)
type TerminalCreatedEventV1 struct {
    TerminalID string `json:"terminal_id"`
    MerchantID string `json:"merchant_id"`
    Gateway    string `json:"gateway"`
}

// V2 - BREAKING CHANGE
type TerminalCreatedEventV2 struct {
    TerminalID string `json:"tid"`  // ❌ Renamed field!
    MerchantID string `json:"mid"`  // ❌ Renamed field!
    // ❌ gateway field removed!
    GatewayID  string `json:"gateway_id"`  // New field replacing gateway
    OrgID      string `json:"org_id"`      // New optional field
}

// ❌ Old consumers break when they receive V2 events!
```

**Problem:**
- Consumers expecting `terminal_id` get `tid`
- Missing `gateway` field breaks existing logic
- Cascading failures across services

### Good Pattern ✅

```go
// CORRECT: Backward compatible changes

// V1 (existing)
type TerminalCreatedEventV1 struct {
    TerminalID string `json:"terminal_id"`
    MerchantID string `json:"merchant_id"`
    Gateway    string `json:"gateway"`
}

// V2 - BACKWARD COMPATIBLE
type TerminalCreatedEventV2 struct {
    // ✅ Keep all V1 fields with same names
    TerminalID string `json:"terminal_id"`  // Same name
    MerchantID string `json:"merchant_id"`  // Same name
    Gateway    string `json:"gateway"`      // Same name

    // ✅ Add new fields only (optional for consumers)
    OrgID      string `json:"org_id,omitempty"`
    GatewayID  string `json:"gateway_id,omitempty"`  // Additional, not replacing
}

// PATTERN 1: Publish both V1 and V2 during transition
func PublishTerminalCreated(terminal *Terminal) error {
    // Publish V1 for old consumers
    eventV1 := TerminalCreatedEventV1{
        TerminalID: terminal.ID,
        MerchantID: terminal.MerchantID,
        Gateway:    terminal.Gateway,
    }
    eventBus.Publish("terminals.terminal.created.v1", eventV1)

    // Also publish V2 for new consumers
    eventV2 := TerminalCreatedEventV2{
        TerminalID: terminal.ID,
        MerchantID: terminal.MerchantID,
        Gateway:    terminal.Gateway,
        OrgID:      terminal.OrgID,
        GatewayID:  terminal.GatewayID,
    }
    eventBus.Publish("terminals.terminal.created.v2", eventV2)

    return nil
}

// PATTERN 2: Version-aware consumer
func HandleTerminalCreated(ctx context.Context, payload []byte) error {
    // Try V2 first
    var eventV2 TerminalCreatedEventV2
    if err := json.Unmarshal(payload, &eventV2); err == nil {
        if eventV2.OrgID != "" {
            // V2 event
            return processTerminalV2(eventV2)
        }
    }

    // Fallback to V1
    var eventV1 TerminalCreatedEventV1
    if err := json.Unmarshal(payload, &eventV1); err != nil {
        return err
    }

    return processTerminalV1(eventV1)
}
```

**Backward Compatibility Rules:**
1. ✅ **Can add** new optional fields
2. ✅ **Can deprecate** fields (keep but mark deprecated)
3. ❌ **Cannot remove** existing fields
4. ❌ **Cannot rename** fields
5. ❌ **Cannot change** field types

### Severity

🚨 **Critical** - Breaking changes, service failures

---

## Check 3: Consumer Registration Validation ⚠️ HIGH

### What to Check

All event consumers must be documented and registered at startup.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Undocumented consumers

// Somewhere in terminal-service
func PublishTerminalCreated(terminal *Terminal) {
    eventBus.Publish("terminal.created", event)
    // ❌ Who consumes this event? Unknown!
}

// Somewhere in notification-service
func init() {
    // ❌ Consumer not registered, event is published but ignored
}

// Somewhere else in analytics-service
func HandleTerminalCreated(event TerminalCreatedEvent) {
    // Handler exists but NOT REGISTERED
    // ❌ Silent failure - events published but not processed
}
```

**Problem:**
- Events published to void
- Missing business logic
- Silent failures

### Good Pattern ✅

```go
// CORRECT: Document consumers in event schema

// events/terminal_events.go
// TerminalCreatedEvent is published when a new terminal is created
//
// Topic: terminals.terminal.created.v1
//
// Publishers:
//   - terminal-service (internal/services/terminal_service.go:CreateTerminal)
//
// Consumers:
//   - notification-service (sends email to merchant)
//   - analytics-service (tracks terminal creation metrics)
//   - audit-service (logs terminal creation for compliance)
//
// Contract Owner: platform-team
// Last Updated: 2025-01-15
type TerminalCreatedEvent struct {
    // ... schema
}

// notification-service: events/consumers.go
func RegisterEventConsumers(eventBus *EventBus) {
    // ✅ All consumers registered in one place
    eventBus.Subscribe("terminals.terminal.created.v1", handlers.HandleTerminalCreated)
    eventBus.Subscribe("terminals.terminal.updated.v1", handlers.HandleTerminalUpdated)

    logger.Info("registered_event_consumers",
        "service", "notification-service",
        "count", 2)
}

// analytics-service: events/consumers.go
func RegisterEventConsumers(eventBus *EventBus) {
    eventBus.Subscribe("terminals.terminal.created.v1", handlers.TrackTerminalCreation)

    logger.Info("registered_event_consumers",
        "service", "analytics-service",
        "count", 1)
}

// main.go
func main() {
    eventBus := NewEventBus()

    // ✅ Register all consumers at startup
    events.RegisterEventConsumers(eventBus)

    // Start service
}
```

### Detection Strategy

Look for event Publish calls and verify corresponding Subscribe calls exist and are documented.

### Severity

⚠️ **High** - Missing functionality, silent failures

---

## Check 4: Event Field Type Consistency ⚠️ HIGH

### What to Check

Event fields must use consistent types across all events (string IDs, timestamps, etc.).

### Bad Pattern ❌

```go
// ANTI-PATTERN: Inconsistent types

type TerminalCreatedEvent struct {
    TerminalID int    `json:"terminal_id"`  // ❌ int
    CreatedAt  string `json:"created_at"`   // ❌ string timestamp
}

type TerminalUpdatedEvent struct {
    TerminalID string    `json:"terminal_id"`  // ❌ Different type (string)!
    UpdatedAt  time.Time `json:"updated_at"`   // ❌ Different timestamp format!
}

type PaymentCreatedEvent struct {
    TerminalID uint64 `json:"terminal_id"`  // ❌ Yet another type!
    CreatedAt  int64  `json:"created_at"`   // ❌ Unix timestamp!
}
```

**Problem:**
- Consumers must handle multiple types for same field
- Difficult to correlate events
- Parsing errors

### Good Pattern ✅

```go
// CORRECT: Consistent types across all events

// Standard field types (define in shared package)
type EventMetadata struct {
    // ✅ IDs always string (14-char Razorpay IDs)
    RequestID string `json:"request_id"`
    TraceID   string `json:"trace_id"`

    // ✅ Timestamps always time.Time (RFC3339 in JSON)
    Timestamp time.Time `json:"timestamp"`
}

type TerminalCreatedEvent struct {
    EventMetadata

    // ✅ Consistent: string ID
    TerminalID string `json:"terminal_id"`
    MerchantID string `json:"merchant_id"`

    // ✅ Consistent: time.Time
    CreatedAt time.Time `json:"created_at"`
}

type TerminalUpdatedEvent struct {
    EventMetadata

    // ✅ Same type as TerminalCreatedEvent
    TerminalID string `json:"terminal_id"`
    MerchantID string `json:"merchant_id"`

    // ✅ Same timestamp type
    UpdatedAt time.Time `json:"updated_at"`
}

type PaymentCreatedEvent struct {
    EventMetadata

    // ✅ Consistent string ID
    PaymentID  string `json:"payment_id"`
    TerminalID string `json:"terminal_id"`  // Same type for cross-event correlation

    // ✅ Consistent timestamp type
    CreatedAt time.Time `json:"created_at"`

    // ✅ Amounts always int (paise/cents)
    Amount int64 `json:"amount"`
}

// Consumers can rely on consistent types
func HandleEvent(payload []byte) error {
    // ✅ All events have same metadata structure
    var meta EventMetadata
    json.Unmarshal(payload, &meta)

    logger.Info("processing_event",
        "trace_id", meta.TraceID,
        "timestamp", meta.Timestamp)

    // Process...
}
```

**Type Standards:**
- **IDs**: `string` (14-char Razorpay format: `term_abc123xyz`)
- **Timestamps**: `time.Time` (RFC3339 in JSON)
- **Amounts**: `int64` (smallest currency unit - paise/cents)
- **Booleans**: `bool` (not `0/1` or `"true"/"false"`)
- **Enums**: `string` with validation (not magic numbers)

### Severity

⚠️ **High** - Parsing errors, integration complexity

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | Schema documentation | 🚨 Critical | Contract violations |
| 2 | Backward compatibility | 🚨 Critical | Breaking changes |
| 3 | Consumer registration | ⚠️ High | Silent failures |
| 4 | Type consistency | ⚠️ High | Parsing errors |

---

## How to Apply

**For each file matching** `internal/events/*`, `pkg/eventing/*`:

1. Check event structs are documented
2. Verify schema changes are backward compatible
3. Check all consumers are registered
4. Verify consistent field types across events

**Example output:**

```
📁 File: internal/events/terminal_events.go

🚨 Check #1 Failed: Event schema not documented (Line 23)
   Code: type TerminalCreatedEvent struct {...}
   Fix: Add documentation with publishers and consumers

🚨 Check #2 Failed: Breaking schema change (Line 45)
   Code: Renamed field from "terminal_id" to "tid"
   Fix: Keep old field name, add new field separately

⚠️  Check #4 Failed: Inconsistent timestamp type (Line 67)
   Code: CreatedAt string (other events use time.Time)
   Fix: Change to time.Time for consistency

✅ Check #3 Passed: All consumers registered in RegisterEventConsumers()
```
