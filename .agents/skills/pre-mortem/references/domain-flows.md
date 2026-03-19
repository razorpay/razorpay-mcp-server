# Domain Business Flow Validation Checks

## Overview

Validates that code changes preserve critical business flow steps, ordering, cache invalidation, and error paths documented in the repo skill.

**Load when:** PR modifies domain entity code (`internal/{domain}/*`)

**Total Checks:** 4

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 2

**How It Works:**
1. Detect which domain is modified (e.g., `gateway_credentials`, `terminal`, `merchant_instrument_request`)
2. Load flow documentation from repo skill: `.claude/skills/*-skill/modules/domain/{domain}/flows.md`
3. Verify PR code preserves critical flow steps
4. Flag violations

---

## Check 1: Critical Flow Steps Present 🚨 CRITICAL

### What to Check

Critical steps documented in flow must be present in code.

### How to Find Flows

**From repo skill:**
```bash
# Check if repo has skill
SKILL_DIR=$(find . -type d -name "*-skill" -path "*/.claude/skills/*" | head -1)

if [ -n "$SKILL_DIR" ]; then
    # Determine domain from modified files
    # E.g., internal/gateway_credentials/* → gateway-credentials
    DOMAIN=$(extract_domain)

    # Load flows
    FLOWS_FILE="$SKILL_DIR/modules/domain/$DOMAIN/flows.md"

    if [ -f "$FLOWS_FILE" ]; then
        # Parse critical flow steps
        grep -A 20 "## Flow:" "$FLOWS_FILE"
    fi
fi
```

### Example from terminals repo skill

**File:** `.claude/skills/terminals-skill/modules/domain/gateway-credentials/flows.md`

```markdown
## Flow: Create Gateway Credential

### Critical Steps (Must be Present)

1. **Validate Request**
   - Check merchant_id exists
   - Validate gateway is supported
   - Check org_id belongs to merchant

2. **Check for Existing Credential**
   - Query by (gateway, org_id, acquirer) where deleted_at IS NULL
   - Return error if active credential exists

3. **Encrypt Secrets**
   - Extract secrets from request
   - Encrypt using KMS
   - Store encrypted version only

4. **Save to Database**
   - Create gateway_credential record
   - Set status = 'active'

5. **Invalidate Cache**
   - Delete cached credentials for org_id
   - Clear gateway config cache

6. **Publish Event**
   - Emit gateway.credential.created event
   - Include encrypted=true flag
```

### Bad Pattern ❌

**PR modifies:** `internal/gateway_credentials/service.go`

```go
// ANTI-PATTERN: Missing critical steps
func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    // ❌ Step 1 missing: No validation!

    // ❌ Step 2 missing: No duplicate check!

    // Step 3: Encrypt (present)
    EncryptSecrets(dao)

    // Step 4: Save (present)
    model := TransformToModel(dao)
    repo.Save(model)

    // ❌ Step 5 missing: No cache invalidation!
    // ❌ Step 6 missing: No event published!

    return nil
}
```

**Problem:**
- Duplicate credentials created (missing step 2)
- Stale cache served (missing step 5)
- Downstream services not notified (missing step 6)

### Good Pattern ✅

```go
// CORRECT: All critical steps present
func CreateGatewayCredential(ctx *gin.Context, dao *daos.GatewayCredential) error {
    // ✅ Step 1: Validate request
    if err := validateGatewayRequest(dao); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    // ✅ Step 2: Check for existing credential
    existing := repo.FindByGatewayOrgAcquirer(
        dao.Gateway,
        dao.OrgId,
        dao.Acquirer,
    )
    if existing != nil {
        return errors.New("active credential already exists")
    }

    // ✅ Step 3: Encrypt secrets
    if err := EncryptSecrets(dao); err != nil {
        return fmt.Errorf("encryption failed: %w", err)
    }

    // ✅ Step 4: Save to database
    model := TransformToModel(dao)
    if err := repo.Save(model); err != nil {
        return fmt.Errorf("save failed: %w", err)
    }

    // ✅ Step 5: Invalidate cache
    cache.Delete(fmt.Sprintf("credentials:org:%s", dao.OrgId))
    cache.Delete(fmt.Sprintf("gateway_config:%s", dao.Gateway))

    // ✅ Step 6: Publish event
    event := buildCredentialCreatedEvent(model)
    if err := eventBus.Publish("gateway.credential.created", event); err != nil {
        logger.Warn(ctx, "event_publish_failed", "error", err)
    }

    return nil
}
```

### Detection Strategy

```bash
# Step 1: Parse critical steps from skill
STEPS=$(grep -A 1 "^[0-9]\\." "$FLOWS_FILE" | grep -v "^--$")

# Example output:
# 1. **Validate Request**
# 2. **Check for Existing Credential**
# 3. **Encrypt Secrets**
# 4. **Save to Database**
# 5. **Invalidate Cache**
# 6. **Publish Event**

# Step 2: For each step, verify implementation exists in PR
for step in $STEPS; do
    case $step in
        "Validate Request")
            # Look for validation function call
            if ! git diff main..HEAD -- internal/${DOMAIN}/* | grep -q "validate"; then
                FLAG: "Step 1 missing: Validate Request"
            fi
            ;;
        "Check for Existing Credential")
            # Look for duplicate check query
            if ! git diff main..HEAD | grep -q "Find.*Existing\|FindBy"; then
                FLAG: "Step 2 missing: Duplicate check"
            fi
            ;;
        "Encrypt Secrets")
            if ! git diff main..HEAD | grep -q "Encrypt"; then
                FLAG: "Step 3 missing: Encryption"
            fi
            ;;
        "Invalidate Cache")
            if ! git diff main..HEAD | grep -q "cache\\.Delete\|cache\\.Invalidate"; then
                FLAG: "Step 5 missing: Cache invalidation"
            fi
            ;;
        "Publish Event")
            if ! git diff main..HEAD | grep -q "Publish\|PublishEvent"; then
                FLAG: "Step 6 missing: Event publishing"
            fi
            ;;
    esac
done
```

### Flag Conditions

Flag if:
- PR adds/modifies domain method
- Critical step from flow is missing in code
- Step order changed without justification

### Severity

🚨 **Critical** - Business logic violated, data corruption, stale cache

### Output Format

```
🚨 Critical Flow Steps Missing

Domain: Gateway Credentials
File: internal/gateway_credentials/service.go:CreateGatewayCredential

Flow (from skill): Create Gateway Credential
  .claude/skills/terminals-skill/modules/domain/gateway-credentials/flows.md:23

Missing steps:
  ❌ Step 2: Check for Existing Credential
      Expected: FindByGatewayOrgAcquirer() call before Save()
      Impact: Duplicate credentials can be created

  ❌ Step 5: Invalidate Cache
      Expected: cache.Delete() for org credentials
      Impact: Stale cache served to API

  ❌ Step 6: Publish Event
      Expected: eventBus.Publish("gateway.credential.created")
      Impact: Downstream services not notified

Present steps:
  ✅ Step 1: Validate Request
  ✅ Step 3: Encrypt Secrets
  ✅ Step 4: Save to Database
```

---

## Check 2: Step Ordering Preserved ⚠️ HIGH

### What to Check

Flow steps must execute in documented order to maintain data consistency.

### Example from skill

```markdown
## Flow: Update Terminal Status

**Step Order (CRITICAL):**

1. **Validate new status is allowed**
   - Check status transition is valid (e.g., active → inactive allowed, suspended → active not allowed without approval)

2. **Begin transaction**

3. **Update terminal record**
   - Set status = new_status
   - Set updated_at = now()

4. **Update dependent records**
   - Disable all active instruments for this terminal
   - Update gateway configuration status

5. **Commit transaction**

6. **Invalidate cache AFTER commit**
   - Delete terminal cache
   - Delete merchant terminal list cache

7. **Publish event**
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: Wrong order
func UpdateTerminalStatus(terminalID, newStatus string) error {
    // ❌ Step 6 before Step 2: Cache invalidated before transaction!
    cache.Delete(fmt.Sprintf("terminal:%s", terminalID))

    // Step 2: Begin transaction
    tx := db.Begin()
    defer tx.Rollback()

    // ❌ Step 3 missing: No status validation!

    // Step 4: Update terminal
    tx.Model(&Terminal{}).Where("id = ?", terminalID).Update("status", newStatus)

    // Step 5: Commit
    tx.Commit()

    // ❌ Step 7 before Step 6: Event published before all cache cleared!
    eventBus.Publish("terminal.status.updated", event)

    return nil
}
```

**Problem:**
- Cache invalidated before transaction commit → race condition
- New status in cache, old status in DB during transaction
- Dependent records not updated

### Good Pattern ✅

```go
// CORRECT: Steps in documented order
func UpdateTerminalStatus(terminalID, newStatus string) error {
    // ✅ Step 1: Validate status transition
    terminal, _ := repo.FindTerminal(terminalID)
    if !isStatusTransitionAllowed(terminal.Status, newStatus) {
        return errors.New("status transition not allowed")
    }

    // ✅ Step 2: Begin transaction
    tx := db.Begin()
    defer tx.Rollback()

    // ✅ Step 3: Update terminal record
    if err := tx.Model(&Terminal{}).
        Where("id = ?", terminalID).
        Updates(map[string]interface{}{
            "status":     newStatus,
            "updated_at": time.Now(),
        }).Error; err != nil {
        return err
    }

    // ✅ Step 4: Update dependent records
    tx.Model(&Instrument{}).
        Where("terminal_id = ? AND status = 'active'", terminalID).
        Update("status", "inactive")

    tx.Model(&GatewayConfig{}).
        Where("terminal_id = ?", terminalID).
        Update("status", newStatus)

    // ✅ Step 5: Commit transaction
    if err := tx.Commit().Error; err != nil {
        return err
    }

    // ✅ Step 6: Invalidate cache AFTER commit
    cache.Delete(fmt.Sprintf("terminal:%s", terminalID))
    cache.Delete(fmt.Sprintf("merchant_terminals:%s", terminal.MerchantID))

    // ✅ Step 7: Publish event
    eventBus.Publish("terminal.status.updated", buildEvent(terminal))

    return nil
}
```

### Severity

⚠️ **High** - Race conditions, data inconsistency

---

## Check 3: Error Paths Handled ⚠️ HIGH

### What to Check

Documented error scenarios must have proper handling.

### Example from skill

```markdown
## Flow: Process Payment

### Error Paths (Must Handle)

**E1: Insufficient Funds**
- Return user-friendly error
- Log for fraud detection
- DO NOT retry automatically

**E2: Gateway Timeout**
- Retry up to 3 times with backoff
- If all retries fail, mark payment as pending
- Create manual review task

**E3: Duplicate Payment**
- Check idempotency key before processing
- Return original payment result if duplicate
- Log duplicate attempt
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: Error paths not handled per flow
func ProcessPayment(payment *Payment) error {
    result, err := gateway.Charge(payment)
    if err != nil {
        // ❌ All errors handled same way (no differentiation)
        return err
    }

    return result
}
```

### Good Pattern ✅

```go
// CORRECT: Error paths per flow documentation
func ProcessPayment(payment *Payment) error {
    // Check idempotency (E3)
    if existing := checkDuplicatePayment(payment.IdempotencyKey); existing != nil {
        logger.Info("duplicate_payment_detected", "payment_id", existing.ID)
        return existing  // ✅ Return original result
    }

    // Process payment
    result, err := gateway.ChargeWithRetry(payment)

    if err != nil {
        // ✅ E1: Insufficient funds
        if errors.Is(err, ErrInsufficientFunds) {
            logger.Info("insufficient_funds",
                "payment_id", payment.ID,
                "merchant_id", payment.MerchantID)
            // DO NOT retry
            return &PaymentResult{
                Status: "failed",
                Error:  "insufficient funds",
            }
        }

        // ✅ E2: Gateway timeout
        if errors.Is(err, context.DeadlineExceeded) {
            logger.Warn("gateway_timeout_all_retries_failed")
            // Mark as pending for manual review
            payment.Status = "pending_review"
            repo.Save(payment)
            createManualReviewTask(payment)
            return &PaymentResult{
                Status: "pending",
                Error:  "gateway timeout",
            }
        }

        // Other errors
        return err
    }

    return result
}
```

### Severity

⚠️ **High** - Incorrect error handling, poor UX

---

## Check 4: Cache Invalidation on Updates 🚨 CRITICAL

### What to Check

Updates must invalidate all related cache keys documented in flow.

### Example from skill

```markdown
## Flow: Update Terminal

### Cache Invalidation (CRITICAL)

When terminal is updated, invalidate:
1. `terminal:{terminal_id}` - Direct terminal cache
2. `merchant_terminals:{merchant_id}` - Merchant's terminal list
3. `org_terminals:{org_id}` - Organization's terminal list
4. `gateway_terminals:{gateway}:{org_id}` - Gateway-specific list
5. `terminal_config:{terminal_id}` - Terminal configuration

**Order:** Invalidate AFTER database commit, BEFORE event publish
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: Incomplete cache invalidation
func UpdateTerminal(terminal *Terminal) error {
    if err := repo.Save(terminal); err != nil {
        return err
    }

    // ❌ Only invalidates direct cache, misses related caches
    cache.Delete(fmt.Sprintf("terminal:%s", terminal.ID))

    // ❌ Missing:
    //   - merchant_terminals cache
    //   - org_terminals cache
    //   - gateway_terminals cache
    //   - terminal_config cache

    return nil
}
```

**Problem:**
- Stale terminal lists served
- Old configuration used
- Inconsistent data across endpoints

### Good Pattern ✅

```go
// CORRECT: Comprehensive cache invalidation per flow
func UpdateTerminal(terminal *Terminal) error {
    if err := repo.Save(terminal); err != nil {
        return err
    }

    // ✅ Invalidate all related caches from flow documentation

    // 1. Direct terminal cache
    cache.Delete(fmt.Sprintf("terminal:%s", terminal.ID))

    // 2. Merchant's terminal list
    cache.Delete(fmt.Sprintf("merchant_terminals:%s", terminal.MerchantID))

    // 3. Organization's terminal list
    cache.Delete(fmt.Sprintf("org_terminals:%s", terminal.OrgID))

    // 4. Gateway-specific list
    cache.Delete(fmt.Sprintf("gateway_terminals:%s:%s", terminal.Gateway, terminal.OrgID))

    // 5. Terminal configuration
    cache.Delete(fmt.Sprintf("terminal_config:%s", terminal.ID))

    logger.Info("cache_invalidated",
        "terminal_id", terminal.ID,
        "keys_invalidated", 5)

    return nil
}

// PATTERN 2: Helper function for cache invalidation
func InvalidateTerminalCaches(terminal *Terminal) {
    cacheKeys := []string{
        fmt.Sprintf("terminal:%s", terminal.ID),
        fmt.Sprintf("merchant_terminals:%s", terminal.MerchantID),
        fmt.Sprintf("org_terminals:%s", terminal.OrgID),
        fmt.Sprintf("gateway_terminals:%s:%s", terminal.Gateway, terminal.OrgID),
        fmt.Sprintf("terminal_config:%s", terminal.ID),
    }

    for _, key := range cacheKeys {
        if err := cache.Delete(key); err != nil {
            logger.Warn("cache_delete_failed", "key", key, "error", err)
        }
    }
}
```

### Severity

🚨 **Critical** - Stale data served, data inconsistency

---

## Summary Table

| Check # | Validates | Severity | From Skill |
|---------|-----------|----------|------------|
| 1 | Critical steps present | 🚨 Critical | Flow step list |
| 2 | Step ordering | ⚠️ High | Documented order |
| 3 | Error path handling | ⚠️ High | Error scenarios |
| 4 | Cache invalidation | 🚨 Critical | Cache keys list |

---

## How to Apply

**For each domain modified:**

1. **Detect domain:**
   ```bash
   DOMAIN=$(basename $(dirname $CHANGED_FILE) | tr '_' '-')
   ```

2. **Load flows from skill:**
   ```bash
   SKILL_DIR=$(find . -name "*-skill" -path "*/.claude/skills/*" | head -1)
   FLOWS="$SKILL_DIR/modules/domain/$DOMAIN/flows.md"
   ```

3. **Parse flow steps:**
   - Critical steps: `grep "^[0-9]\\." $FLOWS`
   - Step order: `grep "Step Order" $FLOWS`
   - Error paths: `grep "Error Paths" $FLOWS`
   - Cache keys: `grep "Cache Invalidation" $FLOWS`

4. **Verify in PR code:**
   - Check all steps present
   - Verify order preserved
   - Check error handling
   - Validate cache invalidation

### Example Output

```
Domain: Gateway Credentials
Modified: internal/gateway_credentials/service.go

🚨 Check #1 Failed: Missing critical steps (Line 48)
   Flow: Create Gateway Credential
   From skill: flows.md:23

   Missing:
     - Step 2: Check for Existing Credential
     - Step 5: Invalidate Cache
     - Step 6: Publish Event

🚨 Check #4 Failed: Incomplete cache invalidation (Line 92)
   Expected cache keys (from skill):
     - credentials:org:{org_id}
     - gateway_config:{gateway}
   Found:
     - credentials:org:{org_id} ✓
   Missing:
     - gateway_config:{gateway} ✗

✅ Check #2 Passed: Step ordering correct
✅ Check #3 Passed: Error paths handled
```

---

## Fallback (No Repo Skill)

If repo doesn't have `.claude/skills/*-skill/modules/domain/`:

```
ℹ️  Domain flow validation skipped
Reason: No repo skill found with flow documentation

Recommendation: Create repo skill with domain flows for automated validation
See: skill-creator for how to document business flows
```
