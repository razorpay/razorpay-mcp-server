# Domain Business Rules Validation Checks

## Overview

Validates that code changes don't violate documented business rules and validation logic defined in the repo skill.

**Load when:** PR modifies domain entity code (`internal/{domain}/*`)

**Total Checks:** 2

**Severity Distribution:**
- 🚨 Critical: 2

**How It Works:**
1. Detect which domain is modified
2. Load business rules from repo skill: `.claude/skills/*-skill/modules/domain/{domain}/rules.md` or `decisions.md`
3. Verify PR code enforces rules
4. Flag violations

---

## Check 1: Validation Rules Enforced 🚨 CRITICAL

### What to Check

Business validation rules documented in repo skill must be enforced in code.

### Example from terminals repo skill

**File:** `.claude/skills/terminals-skill/modules/domain/gateway-credentials/rules.md`

```markdown
## Validation Rules

### Rule 1: Gateway Support Validation

**Rule:** Only supported gateways can be used for credentials

**Supported Gateways:**
- Payment: hitachi, cybersource, upi_airtel, wallet_amazonpay
- Card: axis_migs, billdesk, hdfc
- Netbanking: paytm_netbanking, kotak_netbanking

**Validation:**
```go
if !utils.StringInSlice(gateway, SupportedGateways) {
    return errors.New("gateway not supported")
}
```

### Rule 2: Encryption Mandatory for Secrets

**Rule:** All secret fields must be encrypted before storage

**Secret Fields:**
- api_key, api_secret, private_key, password, token

**Validation:**
- Check if field name contains "secret", "key", "password"
- Encrypt using KMS before Save()
- Never log secret values

### Rule 3: Organization-Gateway Pairing

**Rule:** Org can only have ONE active credential per (gateway, acquirer) pair

**Validation:**
```go
existing := repo.FindByGatewayOrgAcquirer(gateway, orgId, acquirer)
if existing != nil && existing.DeletedAt == nil {
    return errors.New("active credential exists for this gateway-acquirer")
}
```
```

### Bad Pattern ❌

**PR modifies:** `internal/gateway_credentials/service.go`

```go
// ANTI-PATTERN: Rules not enforced
func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    // ❌ Rule 1 violated: No gateway support check
    // Allows any gateway string, even unsupported ones

    // ❌ Rule 2 violated: Secrets stored in plaintext
    model := &GatewayCredential{
        Gateway:    dao.Gateway,
        APIKey:     dao.APIKey,      // ❌ Plaintext!
        APISecret:  dao.APISecret,   // ❌ Plaintext!
    }

    // ❌ Rule 3 violated: No duplicate check
    // Can create multiple active credentials for same gateway-org pair
    repo.Save(model)

    return nil
}
```

**Problem:**
- Unsupported gateways accepted → integration failures
- Secrets leaked in logs/database → security breach
- Duplicate credentials → unpredictable routing

### Good Pattern ✅

```go
// CORRECT: All rules enforced
var SupportedGateways = []string{
    "hitachi", "cybersource", "upi_airtel", "wallet_amazonpay",
    "axis_migs", "billdesk", "hdfc",
    "paytm_netbanking", "kotak_netbanking",
}

func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    // ✅ Rule 1: Gateway support validation
    if !utils.StringInSlice(dao.Gateway, SupportedGateways) {
        logger.Warn(ctx, "unsupported_gateway", "gateway", dao.Gateway)
        return fmt.Errorf("gateway %s not supported", dao.Gateway)
    }

    // ✅ Rule 3: Check for existing credential
    existing := repo.FindByGatewayOrgAcquirer(
        dao.Gateway,
        dao.OrgId,
        dao.Acquirer,
    )
    if existing != nil && existing.DeletedAt == nil {
        return errors.New("active credential exists for this gateway-acquirer")
    }

    // ✅ Rule 2: Encrypt secrets before storage
    if err := encryptSecrets(dao); err != nil {
        logger.Error(ctx, "encryption_failed", "error", err)
        return fmt.Errorf("failed to encrypt secrets: %w", err)
    }

    model := TransformToModel(dao)
    if err := repo.Save(model); err != nil {
        return err
    }

    return nil
}

func encryptSecrets(dao *daos.GatewayCredential) error {
    // ✅ Encrypt all secret fields
    secretFields := []struct {
        value *string
        name  string
    }{
        {&dao.APIKey, "api_key"},
        {&dao.APISecret, "api_secret"},
        {&dao.PrivateKey, "private_key"},
    }

    for _, field := range secretFields {
        if *field.value != "" {
            encrypted, err := kms.Encrypt(*field.value)
            if err != nil {
                return fmt.Errorf("failed to encrypt %s: %w", field.name, err)
            }
            *field.value = encrypted

            // ✅ Never log secrets
            logger.Info(ctx, "field_encrypted", "field", field.name)
        }
    }

    return nil
}
```

### Detection Strategy

```bash
# Step 1: Load rules from skill
SKILL_DIR=$(find . -name "*-skill" -path "*/.claude/skills/*" | head -1)
DOMAIN=$(extract_domain_from_pr_files)
RULES_FILE="$SKILL_DIR/modules/domain/$DOMAIN/rules.md"

# Step 2: Parse rules
grep -A 10 "### Rule" "$RULES_FILE"

# Step 3: For each rule, verify enforcement in PR
# Example for Rule 1: Gateway Support
if grep -q "SupportedGateways" "$RULES_FILE"; then
    if ! git diff main..HEAD | grep -q "StringInSlice.*SupportedGateways"; then
        FLAG: "Rule 1 not enforced: Gateway support validation missing"
    fi
fi

# Example for Rule 2: Encryption
if grep -q "Encryption Mandatory" "$RULES_FILE"; then
    if ! git diff main..HEAD | grep -q "Encrypt\|kms\."; then
        FLAG: "Rule 2 not enforced: Secret encryption missing"
    fi
fi

# Example for Rule 3: Uniqueness
if grep -q "ONE active credential" "$RULES_FILE"; then
    if ! git diff main..HEAD | grep -q "FindBy.*Existing\|FindByGatewayOrgAcquirer"; then
        FLAG: "Rule 3 not enforced: Duplicate check missing"
    fi
fi
```

### Flag Conditions

Flag if:
- Documented rule exists in skill
- PR modifies related domain code
- Rule validation not present in code
- Rule can be bypassed

### Severity

🚨 **Critical** - Business rules violated, data corruption, security issues

### Output Format

```
🚨 Business Rules Not Enforced

Domain: Gateway Credentials
File: internal/gateway_credentials/service.go:CreateGatewayCredential

Rules (from skill): rules.md

❌ Rule 1 Violated: Gateway Support Validation
   From skill: rules.md:12
   Requirement: Only supported gateways allowed
   Missing: StringInSlice(gateway, SupportedGateways) check

❌ Rule 2 Violated: Encryption Mandatory for Secrets
   From skill: rules.md:24
   Requirement: Encrypt api_key, api_secret, private_key before Save()
   Missing: encryptSecrets() call before repo.Save()

❌ Rule 3 Violated: Organization-Gateway Pairing
   From skill: rules.md:38
   Requirement: Check for existing (gateway, org_id, acquirer) before creation
   Missing: FindByGatewayOrgAcquirer() query

Impact:
  - Unsupported gateways will cause integration failures
  - Secrets leaked in database (security breach)
  - Duplicate credentials cause routing issues
```

---

## Check 2: Business Logic Not Bypassed 🚨 CRITICAL

### What to Check

Core business logic documented in decisions/rules must not be bypassed.

### Example from terminals repo skill

**File:** `.claude/skills/terminals-skill/modules/domain/terminal/decisions.md`

```markdown
## Decision: Terminal Activation Requires Gateway Configuration

**Context:** Terminals cannot go active without valid gateway configuration

**Decision:** Terminal status can only be set to 'active' if:
1. Gateway is assigned
2. Gateway credential exists and is active
3. Org has payment instruments configured
4. Merchant is verified

**Implementation:**
```go
func ActivateTerminal(terminalId string) error {
    terminal := repo.FindTerminal(terminalId)

    // MUST check these before activation
    if terminal.Gateway == "" {
        return errors.New("gateway not assigned")
    }

    credential := repo.FindActiveCredential(terminal.Gateway, terminal.OrgId)
    if credential == nil {
        return errors.New("no active gateway credential")
    }

    instruments := repo.FindInstruments(terminal.Id)
    if len(instruments) == 0 {
        return errors.New("no payment instruments configured")
    }

    merchant := repo.FindMerchant(terminal.MerchantId)
    if merchant.VerificationStatus != "verified" {
        return errors.New("merchant not verified")
    }

    // Only then set active
    terminal.Status = "active"
    repo.Save(terminal)
}
```

**Rationale:** Prevents terminals from processing payments without proper setup
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: Bypasses business logic
func UpdateTerminalStatus(terminalId, status string) error {
    // ❌ Allows setting to 'active' without checks
    terminal := repo.FindTerminal(terminalId)
    terminal.Status = status  // ❌ No validation!
    repo.Save(terminal)

    return nil
}

// ANTI-PATTERN: Admin bypass without audit
func ForceActivateTerminal(terminalId string) error {
    // ❌ Bypasses all checks - no audit trail
    terminal := repo.FindTerminal(terminalId)
    terminal.Status = "active"  // Dangerous!
    repo.Save(terminal)

    return nil
}
```

**Problem:**
- Terminals activated without gateway → payment failures
- Unverified merchants process payments → fraud risk
- No audit trail for bypass operations

### Good Pattern ✅

```go
// CORRECT: Enforces business logic
func UpdateTerminalStatus(terminalId, newStatus string) error {
    terminal := repo.FindTerminal(terminalId)

    // ✅ Special validation for 'active' status
    if newStatus == "active" {
        if err := validateTerminalActivation(terminal); err != nil {
            logger.Warn(ctx, "terminal_activation_failed",
                "terminal_id", terminalId,
                "reason", err.Error())
            return fmt.Errorf("cannot activate terminal: %w", err)
        }
    }

    terminal.Status = newStatus
    return repo.Save(terminal)
}

func validateTerminalActivation(terminal *Terminal) error {
    // ✅ Check 1: Gateway assigned
    if terminal.Gateway == "" {
        return errors.New("gateway not assigned")
    }

    // ✅ Check 2: Active gateway credential exists
    credential := repo.FindActiveCredential(terminal.Gateway, terminal.OrgId)
    if credential == nil {
        return errors.New("no active gateway credential")
    }

    // ✅ Check 3: Payment instruments configured
    instruments := repo.FindInstruments(terminal.Id)
    if len(instruments) == 0 {
        return errors.New("no payment instruments configured")
    }

    // ✅ Check 4: Merchant verified
    merchant := repo.FindMerchant(terminal.MerchantId)
    if merchant.VerificationStatus != "verified" {
        return errors.New("merchant not verified")
    }

    return nil
}

// PATTERN: Admin bypass with audit trail
func ForceActivateTerminal(ctx *gin.Context, terminalId, reason string) error {
    adminId := ctx.GetString("admin_user_id")

    // ✅ Require explicit reason
    if reason == "" {
        return errors.New("reason required for force activation")
    }

    terminal := repo.FindTerminal(terminalId)

    // ✅ Log bypass with full context
    logger.Warn(ctx, "terminal_force_activated",
        "terminal_id", terminalId,
        "admin_id", adminId,
        "reason", reason,
        "bypassed_checks", "gateway_validation,merchant_verification")

    // ✅ Create audit record
    audit.LogAdminAction(ctx, audit.Action{
        Type:       "FORCE_TERMINAL_ACTIVATION",
        TerminalID: terminalId,
        AdminID:    adminId,
        Reason:     reason,
        Timestamp:  time.Now(),
    })

    terminal.Status = "active"
    terminal.ActivationOverride = true
    terminal.ActivationOverrideReason = reason

    return repo.Save(terminal)
}
```

### Severity

🚨 **Critical** - Business logic violated, fraud risk, payment failures

---

## Summary Table

| Check # | Validates | Severity | From Skill |
|---------|-----------|----------|------------|
| 1 | Validation rules enforced | 🚨 Critical | rules.md |
| 2 | Business logic not bypassed | 🚨 Critical | decisions.md |

---

## How to Apply

**For each domain modified:**

1. **Load rules from skill:**
   ```bash
   SKILL_DIR=$(find . -name "*-skill" -path "*/.claude/skills/*" | head -1)
   DOMAIN=$(extract_domain)
   RULES="$SKILL_DIR/modules/domain/$DOMAIN/rules.md"
   DECISIONS="$SKILL_DIR/modules/domain/$DOMAIN/decisions.md"
   ```

2. **Parse business rules:**
   - Validation rules: `grep "### Rule" $RULES`
   - Core logic: `grep "## Decision" $DECISIONS`

3. **Verify in PR code:**
   - Check validation rules present
   - Verify no bypass paths added
   - Check audit logging for admin overrides

### Example Output

```
Domain: Terminal
Modified: internal/terminal/service.go

🚨 Check #1 Failed: Validation rule not enforced (Line 45)
   Rule: Gateway Support Validation
   From skill: rules.md:12
   Missing: Gateway whitelist check before Save()

🚨 Check #2 Failed: Business logic bypassed (Line 89)
   Decision: Terminal Activation Requires Gateway Configuration
   From skill: decisions.md:23

   Code allows:
     terminal.Status = newStatus  // No validation!

   Should enforce:
     - Gateway assigned
     - Active gateway credential exists
     - Payment instruments configured
     - Merchant verified

   Impact: Terminals can go active without proper setup
```

---

## Fallback (No Repo Skill)

If repo doesn't have `.claude/skills/*-skill/modules/domain/`:

```
ℹ️  Business rules validation skipped
Reason: No repo skill found with rules documentation

Recommendation: Create repo skill with business rules for automated validation
```
