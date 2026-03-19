# Domain Business Logic - Constraints Validation

## Overview

Validates that code changes don't violate documented business rules, unique constraints, required field validations, and allowed value restrictions defined in the repo skill.

**Load when:** PR modifies domain entity code (`internal/{domain}/*`)

**Total Checks:** 4

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 2

**How It Works:**
1. Detect which domain is modified (e.g., `gateway_credentials`, `terminal`, `merchant_instrument_request`)
2. Load constraint documentation from repo skill: `.claude/skills/*-skill/modules/domain/{domain}/constraints.md`
3. Verify PR code enforces constraints
4. Flag violations

---

## Check 1: Unique Constraints Enforced 🚨 CRITICAL

### What to Check

Unique constraints documented in `{domain}/constraints.md` must be validated in code before database insertion to prevent duplicates.

### How to Find Constraints

**From repo skill:**
```bash
# Check if repo has skill
SKILL_DIR=$(find . -type d -name "*-skill" -path "*/.claude/skills/*" | head -1)

if [ -n "$SKILL_DIR" ]; then
    # Determine domain from modified files
    # E.g., internal/gateway_credentials/* → gateway-credentials
    DOMAIN=$(extract_domain)

    # Load constraints
    CONSTRAINTS_FILE="$SKILL_DIR/modules/domain/$DOMAIN/constraints.md"

    if [ -f "$CONSTRAINTS_FILE" ]; then
        # Parse unique constraints
        grep -A 10 "UNIQUE(" "$CONSTRAINTS_FILE"
    fi
fi
```

### Example from terminals repo skill

**File:** `.claude/skills/terminals-skill/modules/domain/gateway-credentials/constraints.md`

```markdown
### Unique Constraint

```sql
UNIQUE(gateway, org_id, acquirer, deleted_at)
```

**Rules**:
- Only ONE active credential per (gateway, org_id, acquirer) combination
- Active = `deleted_at IS NULL`
```

### Bad Pattern ❌

**PR modifies:** `internal/gateway_credentials/service.go`

```go
// ANTI-PATTERN: No duplicate check
func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    model := TransformToModel(dao)
    EncryptSecrets(model)
    repo.Save(model)  // ❌ No uniqueness validation!
    return nil
}
```

**Problem:**
- Violates UNIQUE(gateway, org_id, acquirer) constraint
- Database throws error instead of user-friendly message
- Or worse, if constraint not in DB, creates duplicates!

### Good Pattern ✅

```go
// CORRECT: Validate uniqueness before insert
func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    // Check constraint from skill:
    // UNIQUE(gateway, org_id, acquirer, deleted_at)
    existing := repo.FindByGatewayOrgAcquirer(
        dao.Gateway,
        dao.OrgId,
        dao.Acquirer,
    )

    if existing != nil && !updateFlag {
        return errors.New("Duplicate credential exists")
    }

    model := TransformToModel(dao)
    EncryptSecrets(model)
    repo.Save(model)
    return nil
}
```

### Detection Strategy

```bash
# Step 1: Parse constraints from skill
UNIQUE_CONSTRAINTS=$(grep -E "UNIQUE\(" "$CONSTRAINTS_FILE" | sed 's/UNIQUE(\(.*\))/\1/')

# Example output: "gateway, org_id, acquirer, deleted_at"

# Step 2: Check if PR code validates these columns
for constraint in $UNIQUE_CONSTRAINTS; do
    columns=$(echo $constraint | tr ',' '\n' | tr -d ' ')

    # Search for validation in PR diff
    for col in $columns; do
        if ! git diff main..HEAD -- internal/${DOMAIN}/* | grep -q "Find.*${col}"; then
            FLAG: "Unique constraint on $col not validated"
        fi
    done
done
```

### Flag Conditions

Flag if:
- Skill documents UNIQUE constraint
- PR adds insert/update code for that entity
- No `Find` or `Get` query checking unique columns before insert
- No error handling for duplicate case

### Severity

🚨 **Critical** - Data corruption, duplicate records, constraint violations

### Output Format

```
🚨 Unique Constraint Not Validated

Domain: Gateway Credentials
File: internal/gateway_credentials/service.go:48

Constraint (from skill):
  UNIQUE(gateway, org_id, acquirer, deleted_at)
  "Only ONE active credential per (gateway, org_id, acquirer) combination"

Issue: PR adds Save() without checking for duplicates

Missing validation:
  existing := repo.FindByGatewayOrgAcquirer(gateway, orgId, acquirer)
  if existing != nil { return ErrDuplicate }

Reference:
  .claude/skills/terminals-skill/modules/domain/gateway-credentials/constraints.md:19
```

---

## Check 2: Required Field Validations Present 🚨 CRITICAL

### What to Check

Fields marked NOT NULL or "required" in constraints must be validated before database operations.

### Example from skill

**File:** `.claude/skills/terminals-skill/modules/domain/gateway-credentials/constraints.md`

```markdown
#### Not Null Constraints

| Column | Constraint | Rationale |
|--------|------------|-----------|
| `gateway_credential_id` | NOT NULL, CHAR(14) | Unique identifier required |
| `gateway` | NOT NULL, VARCHAR(255) | Must specify which gateway |
| `org_id` | NOT NULL, CHAR(14) | Must link to organization |
| `identifiers` | NOT NULL, JSONB | Required for gateway communication |
| `secrets` | NOT NULL, JSONB | Required for authentication |
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: No validation of required fields
func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    // ❌ No check if gateway, org_id, secrets are present!
    repo.Save(TransformToModel(dao))
    return nil
}
```

**Problem:**
- Database rejects with generic error
- Poor user experience
- No field-specific error messages

### Good Pattern ✅

```go
// CORRECT: Validate all required fields
func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    // Validate required fields from skill
    if dao.GatewayCredentialId == "" {
        dao.GatewayCredentialId = utils.NewRzpID()  // Auto-generate
    }

    if dao.Gateway == "" {
        return errors.New("gateway is required")
    }

    if dao.OrgId == "" || len(dao.OrgId) != 14 {
        return errors.New("org_id must be 14 characters")
    }

    if dao.Identifiers == nil || len(dao.Identifiers) == 0 {
        return errors.New("identifiers required")
    }

    if dao.Secrets == nil || len(dao.Secrets) == 0 {
        return errors.New("secrets required")
    }

    repo.Save(TransformToModel(dao))
    return nil
}
```

### Detection Strategy

```bash
# Step 1: Parse required fields from skill
REQUIRED_FIELDS=$(grep -A 1 "NOT NULL" "$CONSTRAINTS_FILE" | grep "|" | awk -F'|' '{print $2}' | tr -d ' `')

# Step 2: For each required field, check validation exists
for field in $REQUIRED_FIELDS; do
    # Convert to Go field name (org_id → OrgId)
    go_field=$(echo $field | sed 's/_\([a-z]\)/\U\1/g' | sed 's/^\([a-z]\)/\U\1/')

    # Search for validation in PR
    if ! git diff main..HEAD | grep -qE "(if.*${go_field}.*==|\.${go_field}.*required)"; then
        FLAG: "Required field $field not validated"
    fi
done
```

### Severity

🚨 **Critical** - Database errors, poor UX, incomplete records

---

## Check 3: Allowed Values Restricted ⚠️ HIGH

### What to Check

Fields with restricted allowed values (enums, supported lists) must be validated.

### Example from skill

**File:** `.claude/skills/terminals-skill/modules/domain/gateway-credentials/constraints.md`

```markdown
### gateway

**Supported Values**: 80+ predefined gateway names

**Validation Rules**:
```go
rules := govalidator.MapData{
    "gateway": []string{"required", "gateway"},  // "gateway" is custom validator
}
```

**Examples**:
- Valid: `hitachi`, `upi_airtel`, `cybersource`, `wallet_amazonpay`
- Invalid: `HITACHI` (uppercase), `random_gateway` (not supported)
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: Accepts any gateway string
func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    // ❌ No validation of gateway value
    repo.Save(TransformToModel(dao))
}
```

### Good Pattern ✅

```go
// CORRECT: Validate against allowed list
var SupportedGateways = []string{
    "hitachi", "upi_airtel", "cybersource", "wallet_amazonpay", // ... 80+ gateways
}

func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    if !utils.StringInSlice(dao.Gateway, SupportedGateways) {
        return errors.New("invalid gateway")
    }

    repo.Save(TransformToModel(dao))
}
```

### Detection Strategy

Look for fields with "Supported Values", "Allowed Values", or "Valid:" in constraints, verify validation exists in code.

### Severity

⚠️ **High** - Invalid data, integration failures

---

## Check 4: Field Format Validation ⚠️ HIGH

### What to Check

Fields with format requirements (regex, length, pattern) must be validated.

### Example from skill

**File:** `.claude/skills/terminals-skill/modules/domain/gateway-credentials/constraints.md`

```markdown
### gateway_credential_id

**Format**: 14-character alphanumeric Razorpay ID

**Pattern**: `^[a-zA-Z0-9]{14}$`

**Validation Rules**:
```go
rules := govalidator.MapData{
    "gateway_credential_id": []string{"len:14", "alpha_num"},
}
```
```

### Bad Pattern ❌

```go
// ANTI-PATTERN: No format validation
func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    // ❌ gateway_credential_id could be any string!
    if dao.GatewayCredentialId == "" {
        dao.GatewayCredentialId = "invalid-id!"  // Wrong format
    }
    repo.Save(TransformToModel(dao))
}
```

### Good Pattern ✅

```go
// CORRECT: Validate format
func CreateGatewayCredential(dao *daos.GatewayCredential) error {
    if dao.GatewayCredentialId != "" {
        // Validate format: ^[a-zA-Z0-9]{14}$
        if len(dao.GatewayCredentialId) != 14 {
            return errors.New("gateway_credential_id must be 14 characters")
        }
        if !regexp.MustCompile(`^[a-zA-Z0-9]{14}$`).MatchString(dao.GatewayCredentialId) {
            return errors.New("gateway_credential_id must be alphanumeric")
        }
    } else {
        dao.GatewayCredentialId, _ = utils.NewRzpID()  // Auto-generate valid ID
    }

    repo.Save(TransformToModel(dao))
}
```

### Severity

⚠️ **High** - Data integrity issues, downstream failures

---

## Summary Table

| Check # | Validates | Severity | From Skill |
|---------|-----------|----------|------------|
| 1 | Unique constraints | 🚨 Critical | UNIQUE(...) definitions |
| 2 | Required fields | 🚨 Critical | NOT NULL constraints |
| 3 | Allowed values | ⚠️ High | Supported Values lists |
| 4 | Field formats | ⚠️ High | Pattern/regex definitions |

---

## How to Apply

**For each domain modified:**

1. **Detect domain:**
   ```bash
   # E.g., internal/gateway_credentials/* → gateway-credentials
   DOMAIN=$(basename $(dirname $CHANGED_FILE) | tr '_' '-')
   ```

2. **Load constraints from skill:**
   ```bash
   SKILL_DIR=$(find . -name "*-skill" -path "*/.claude/skills/*" | head -1)
   CONSTRAINTS="$SKILL_DIR/modules/domain/$DOMAIN/constraints.md"
   ```

3. **Parse constraints:**
   - Unique constraints: `grep "UNIQUE(" $CONSTRAINTS`
   - Required fields: `grep "NOT NULL" $CONSTRAINTS`
   - Allowed values: `grep "Supported Values" $CONSTRAINTS`
   - Field formats: `grep "Pattern:" $CONSTRAINTS`

4. **Verify in PR code:**
   - Check if validations exist
   - Flag missing validations
   - Report with skill references

### Example Output

```
Domain: Gateway Credentials
Modified: internal/gateway_credentials/service.go

🚨 Check #1 Failed: Unique constraint not validated (Line 48)
   Constraint: UNIQUE(gateway, org_id, acquirer, deleted_at)
   From skill: constraints.md:19
   Missing: FindByGatewayOrgAcquirer() check

🚨 Check #2 Failed: Required field not validated (Line 52)
   Field: org_id (NOT NULL, CHAR(14))
   From skill: constraints.md:34
   Missing: Length check (must be 14 chars)

⚠️  Check #3 Failed: Allowed values not restricted (Line 58)
   Field: gateway
   From skill: constraints.md:82
   Missing: Validation against supported gateway list

✅ Check #4 Passed: gateway_credential_id format validated
```

---

## Fallback (No Repo Skill)

If repo doesn't have `.claude/skills/*-skill/modules/domain/`:

```
ℹ️  Domain validation skipped
Reason: No repo skill found with domain documentation

Recommendation: Create repo skill with domain constraints for automated validation
See: skill-creator for how to document domain rules
```
