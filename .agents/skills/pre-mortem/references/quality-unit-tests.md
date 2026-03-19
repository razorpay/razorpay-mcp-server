# Unit Test Quality Checks

## Overview

Validates unit test coverage and quality. **Automatically invokes pr-autopilot skill** to generate missing tests or fix coverage issues.

**Load when:** PR adds/modifies code in `internal/*`, `pkg/*`

**Total Checks:** 4

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 1
- 📋 Medium: 1

**Integration:** Uses **pr-autopilot** skill for auto-generating tests and fixing coverage

---

## Check 1: Test Files Exist for New Code 🚨 CRITICAL

### What to Check

Every new `.go` file with business logic must have corresponding `_test.go` file.

### Bad Pattern ❌

```
PR adds:
✅ internal/services/payment_service.go (250 lines)
❌ No internal/services/payment_service_test.go
```

**Problem:**
- New code untested
- No validation of business logic
- Regression risk

### Good Pattern ✅

```
PR includes:
✅ internal/services/payment_service.go (250 lines)
✅ internal/services/payment_service_test.go (150 lines)
```

### Detection Strategy

```bash
# Find new Go files in PR
NEW_FILES=$(git diff --name-only --diff-filter=A main..HEAD | grep "\.go$" | grep -v "_test\.go$")

# For each new file
for file in $NEW_FILES; do
    # Skip test files
    if [[ "$file" == *"_test.go" ]]; then
        continue
    fi

    # Skip non-logic files (models, constants)
    if [[ "$file" == *"/models/"* ]] || [[ "$file" == *"/daos/"* ]]; then
        continue
    fi

    # Check if test file exists
    test_file="${file%.go}_test.go"
    if [ ! -f "$test_file" ]; then
        FLAG: "Test file missing for $file"
    fi
done
```

### Automated Fix with pr-autopilot

When this check fails, pre-mortem **automatically offers to invoke pr-autopilot**:

```
🚨 Test Files Missing

Files without tests:
  - internal/services/payment_service.go (250 lines)
  - internal/terminal/validator.go (120 lines)

Would you like me to generate unit tests for these files?
1. ✅ Yes, generate tests (uses pr-autopilot)
2. Skip for now
```

**If user confirms:**

```bash
# Invoke pr-autopilot skill
invoke_skill("pr-autopilot", {
    "action": "generate_tests",
    "files": [
        "internal/services/payment_service.go",
        "internal/terminal/validator.go"
    ],
    "coverage_target": 80
})
```

**pr-autopilot will:**
1. Read CLAUDE.md for test patterns
2. Analyze function signatures
3. Generate tests following repo patterns
4. Run tests to verify they pass
5. Commit and push using pr-creator

### Flag Conditions

Flag if:
- New `.go` file added in `internal/*` or `pkg/*`
- File has functions (not just structs/constants)
- No corresponding `_test.go` file
- File not in skip list (models, daos, constants)

### Severity

🚨 **Critical** - New code without tests

---

## Check 2: Test Coverage Threshold Met 🚨 CRITICAL

### What to Check

Code coverage must meet minimum threshold (typically 80%+).

### Bad Pattern ❌

```bash
# Run coverage
go test -cover ./...

# Output:
coverage: 45.2% of statements  # ❌ Below 80% threshold
```

### Good Pattern ✅

```bash
# Coverage meets threshold
go test -cover ./...

# Output:
coverage: 85.7% of statements  # ✅ Above 80%
```

### Detection Strategy

```bash
# Run coverage for changed packages
CHANGED_PACKAGES=$(git diff --name-only main..HEAD | grep "\.go$" | xargs dirname | sort -u)

for pkg in $CHANGED_PACKAGES; do
    # Get coverage for package
    coverage=$(go test -cover ./$pkg 2>&1 | grep "coverage:" | awk '{print $2}' | tr -d '%')

    if (( $(echo "$coverage < 80" | bc -l) )); then
        FLAG: "Coverage below 80% for $pkg (current: $coverage%)"
    fi
done
```

### Automated Fix with pr-autopilot

When coverage is low, **offer to auto-generate tests**:

```
🚨 Coverage Below Threshold

Package: internal/services
Current coverage: 45.2%
Target: 80%

Missing coverage in:
  - ProcessPayment() function (0% covered)
  - ValidateTerminal() function (30% covered)
  - CalculateFee() function (0% covered)

Would you like me to generate tests to improve coverage?
1. ✅ Yes, auto-generate tests (uses pr-autopilot)
2. Show me which lines need tests
3. Skip for now
```

**If user confirms:**

```bash
# Invoke pr-autopilot for coverage improvement
invoke_skill("pr-autopilot", {
    "action": "improve_coverage",
    "package": "internal/services",
    "current_coverage": 45.2,
    "target_coverage": 80,
    "focus_functions": [
        "ProcessPayment",
        "ValidateTerminal",
        "CalculateFee"
    ]
})
```

### Flag Conditions

Flag if:
- Package coverage < 80%
- Critical packages (services, handlers) < 90%
- New functions added without any tests

### Severity

🚨 **Critical** - Inadequate test coverage

---

## Check 3: Test Quality (Not Just Coverage) ⚠️ HIGH

### What to Check

Tests should cover edge cases, error paths, not just happy path.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Only happy path tested
func TestCreateTerminal(t *testing.T) {
    terminal := CreateTerminal(validRequest)

    // ✅ Happy path tested
    assert.NotNil(t, terminal)
    assert.Equal(t, "active", terminal.Status)

    // ❌ Missing tests:
    // - Invalid request
    // - Duplicate terminal
    // - Database error
    // - Missing merchant
}
```

### Good Pattern ✅

```go
// CORRECT: Comprehensive test cases
func TestCreateTerminal(t *testing.T) {
    tests := []struct {
        name    string
        request TerminalRequest
        setup   func()
        want    *Terminal
        wantErr bool
    }{
        {
            name:    "success - valid request",
            request: validRequest,
            setup:   func() { setupMerchant() },
            want:    &Terminal{Status: "active"},
            wantErr: false,
        },
        {
            name:    "error - invalid merchant_id",
            request: TerminalRequest{MerchantID: ""},
            wantErr: true,
        },
        {
            name:    "error - duplicate terminal",
            request: validRequest,
            setup:   func() { createExistingTerminal() },
            wantErr: true,
        },
        {
            name:    "error - database failure",
            request: validRequest,
            setup:   func() { mockDatabaseError() },
            wantErr: true,
        },
        {
            name:    "error - merchant not found",
            request: validRequest,
            setup:   func() { /* no merchant */ },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.setup != nil {
                tt.setup()
            }

            got, err := CreateTerminal(tt.request)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want.Status, got.Status)
            }
        })
    }
}
```

### Detection Strategy

```bash
# Check test patterns
grep -A 20 "func Test" *_test.go | \
    grep -c "wantErr\|error\|Error"  # Count error test cases

# Flag if < 30% of tests check errors
```

### Severity

⚠️ **High** - Tests don't catch bugs

---

## Check 4: Test Files Follow Repo Patterns 📋 MEDIUM

### What to Check

Tests should follow patterns documented in CLAUDE.md or existing tests.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Doesn't follow repo test patterns

// Repo uses testify/assert, this uses raw comparisons
func TestTerminal(t *testing.T) {
    result := GetTerminal("123")
    if result == nil {  // ❌ Not using assert
        t.Error("expected terminal")
    }
}

// Repo uses table-driven tests, this doesn't
func TestValidation1(t *testing.T) { /* ... */ }
func TestValidation2(t *testing.T) { /* ... */ }  // ❌ Repetitive
func TestValidation3(t *testing.T) { /* ... */ }
```

### Good Pattern ✅

```go
// CORRECT: Follows repo patterns from CLAUDE.md

// Pattern 1: Use testify/assert
func TestGetTerminal(t *testing.T) {
    result := GetTerminal("123")
    assert.NotNil(t, result)  // ✅ Follows repo pattern
}

// Pattern 2: Table-driven tests
func TestValidation(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  bool
    }{
        {"valid", "term_123", true},
        {"empty", "", false},
        {"invalid_format", "123", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Validate(tt.input)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Severity

📋 **Medium** - Inconsistent test style

---

## Integration with pr-autopilot

### Workflow

```
1. pr-pre-mortem detects missing/low coverage
   ↓
2. Offers to auto-generate tests
   ↓
3. User confirms
   ↓
4. Invokes pr-autopilot skill
   ↓
5. pr-autopilot:
   - Reads CLAUDE.md for test patterns
   - Analyzes functions needing tests
   - Generates tests following repo patterns
   - Runs tests to verify they pass
   - Commits and pushes via pr-creator
   ↓
6. pr-pre-mortem re-runs to verify coverage improved
```

### Auto-invocation Scenarios

Pre-mortem **automatically offers** pr-autopilot when:
- ✅ New file without `_test.go`
- ✅ Coverage < 80%
- ✅ Coverage dropped from previous commit
- ✅ Critical functions (payments, auth) not tested

---

## Summary Table

| Check # | Pattern | Severity | Auto-fix |
|---------|---------|----------|----------|
| 1 | Test files exist | 🚨 Critical | pr-autopilot |
| 2 | Coverage threshold | 🚨 Critical | pr-autopilot |
| 3 | Test quality | ⚠️ High | Manual |
| 4 | Follow repo patterns | 📋 Medium | pr-autopilot |

---

## How to Apply

**For each PR:**

1. Find new/modified Go files
2. Check test files exist
3. Run coverage analysis
4. Verify test quality
5. **Offer pr-autopilot** if issues found

**Example output:**

```
📁 Package: internal/services

🚨 Check #1 Failed: Test file missing
   File: payment_service.go (250 lines)
   Missing: payment_service_test.go

🚨 Check #2 Failed: Coverage below threshold
   Current: 45.2%
   Target: 80%

   Uncovered functions:
     - ProcessPayment() (0%)
     - ValidateTerminal() (30%)

🤖 Auto-Fix Available

I can use pr-autopilot to:
  ✓ Generate payment_service_test.go
  ✓ Add tests for ProcessPayment()
  ✓ Improve ValidateTerminal() coverage
  ✓ Target 80%+ coverage

This will take ~2-3 minutes.

Would you like me to proceed? (yes/no)
```

**After user confirms:**

```
✅ Invoking pr-autopilot...

Generated:
  ✓ payment_service_test.go (15 test cases)
  ✓ Coverage improved: 45.2% → 87.5%
  ✓ All tests passing ✅
  ✓ Committed and pushed

PR updated. Re-running pre-mortem...
```
