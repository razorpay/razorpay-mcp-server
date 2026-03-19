# TOML Configuration Checks

## Overview

Validates TOML configuration changes to prevent production incidents from misconfigured environments, missing keys, or dangerous default values.

**Load when:** PR modifies `configs/*.toml` files

**Total Checks:** 6

**Severity Distribution:**
- 🚨 Critical: 3
- ⚠️ High: 2
- 📋 Medium: 1

---

## Check 1: Missing Keys in Production Configs 🚨 CRITICAL

### What to Check

New keys added to `default.toml` must exist in all critical production configs. Otherwise, production will silently use default values, which are often meant for development.

### Configuration Patterns

**Pattern 1: Single default.toml**
```
configs/
├── default.toml          ← Base config
├── stage-live.toml      ← Stage (live mode)
├── stage-test.toml      ← Stage (test mode)
├── prod-live.toml       ← Production (live mode) ⚠️ CRITICAL
├── prod-test.toml       ← Production (test mode) ⚠️ CRITICAL
```

**Pattern 2: Split default configs**
```
configs/
├── default-live.toml    ← Base for live mode
├── default-test.toml    ← Base for test mode
├── stage-live.toml
├── stage-test.toml
├── prod-live.toml       ⚠️ CRITICAL
├── prod-test.toml       ⚠️ CRITICAL
```

**Pattern 3: Simple (some repos)**
```
configs/
├── default.toml
├── stage.toml           ← Stage
├── prod.toml            ← Production ⚠️ CRITICAL
```

### Detection Strategy

```bash
# Step 1: Detect config pattern
if [ -f "configs/prod.toml" ] && [ -f "configs/stage.toml" ]; then
    PATTERN="simple"
    CRITICAL_CONFIGS=("stage.toml" "prod.toml")
elif [ -f "configs/default-live.toml" ]; then
    PATTERN="split"
    BASE_CONFIGS=("default-live.toml" "default-test.toml")
    CRITICAL_CONFIGS=("stage-live.toml" "stage-test.toml" "prod-live.toml" "prod-test.toml")
else
    PATTERN="standard"
    BASE_CONFIGS=("default.toml")
    CRITICAL_CONFIGS=("stage-live.toml" "stage-test.toml" "prod-live.toml" "prod-test.toml")
fi

# Step 2: Extract new/modified keys from base config
git diff main..HEAD -- configs/default.toml | grep "^+" | grep -v "^+++" | grep "=" > /tmp/new_keys.txt

# Step 3: For each new key, verify it exists in ALL critical configs
for key in $(cat /tmp/new_keys.txt); do
    key_name=$(echo $key | cut -d'=' -f1 | tr -d ' +')

    for config in "${CRITICAL_CONFIGS[@]}"; do
        if ! grep -q "^${key_name}\s*=" "configs/$config"; then
            FLAG: "Key '${key_name}' in default.toml but missing in $config"
        fi
    done
done
```

### Bad Example ❌

**PR adds to configs/default.toml:**
```toml
[newFeature]
enabled = true      # ❌ Testing locally!
debugMode = true    # ❌ Dev setting!
timeout = 5         # ❌ Short timeout for dev!
```

**configs/prod-live.toml NOT updated:**
```toml
# [newFeature] section missing entirely!
```

**Result in production:**
- ✅ `prod-live.toml` loads successfully (no error)
- ❌ **Production uses `enabled = true`, `debugMode = true`, `timeout = 5` from default.toml!**
- 💥 **Untested feature goes live with dev settings!**

### Good Example ✅

**PR updates ALL critical configs:**

```toml
# configs/default.toml
[newFeature]
enabled = true
debugMode = true
timeout = 5

# configs/stage-live.toml
[newFeature]
enabled = false     # ✅ Disabled in stage initially
debugMode = false
timeout = 30

# configs/prod-live.toml
[newFeature]
enabled = false     # ✅ Disabled in prod initially
debugMode = false
timeout = 60

# configs/prod-test.toml
[newFeature]
enabled = false     # ✅ Disabled in prod-test
debugMode = false
timeout = 60
```

### Flag Conditions

Flag if:
- New section `[...]` added to default.toml but missing in any critical config
- New key `key = value` added but missing in critical configs
- Key exists in base but value type different in prod (string vs int)

### Severity

🚨 **Critical** - Production uses untested default values, potential incident

### Output Format

```
🚨 Missing Keys in Production Config

File: configs/default.toml
New keys added:
  - newFeature.enabled = true
  - newFeature.debugMode = true
  - newFeature.timeout = 5

Missing from critical configs:
  ❌ configs/prod-live.toml (all 3 keys missing)
  ❌ configs/prod-test.toml (all 3 keys missing)
  ⚠️  configs/stage-live.toml (debugMode missing)

Risk: Production will use dev default values!

Fix: Add keys to all critical configs with appropriate values
```

---

## Check 2: Dangerous Default Values 🚨 CRITICAL

### What to Check

Default config should not have values that are dangerous for production. Development-oriented defaults can cause incidents if used in production.

### Dangerous Patterns

**Mock Services:**
```toml
# ❌ DANGEROUS DEFAULT
[mozart]
mock = true  # Service returns fake data!
```

**Debug/Verbose Modes:**
```toml
# ❌ DANGEROUS DEFAULT
[application]
debug = true          # Verbose logging, security risk
logLevel = "DEBUG"    # Excessive logs
```

**Feature Flags Enabled:**
```toml
# ❌ DANGEROUS DEFAULT
[features]
autoRefundEnabled = true  # Untested feature!
```

**Short Timeouts:**
```toml
# ❌ DANGEROUS DEFAULT
[database]
timeout = 5  # Too short for production queries
connectionLifetime = 10  # Connections recycled too quickly
```

**Local URLs:**
```toml
# ❌ DANGEROUS DEFAULT
[mozart]
baseUrl = "http://localhost:8080"  # Won't work in prod!
```

### Detection Strategy

```bash
# Parse default.toml for dangerous patterns
grep -E "(mock\s*=\s*true|debug\s*=\s*true|enabled\s*=\s*true|localhost)" configs/default.toml

# Check against dangerous keywords
DANGEROUS_PATTERNS=(
    "mock = true"
    "debug = true"
    "debugMode = true"
    "enabled = true"  # For feature flags
    "localhost"
    "127.0.0.1"
    "timeout = [0-5]$"  # Very short timeouts
    "logLevel = \"DEBUG\""
)

for pattern in "${DANGEROUS_PATTERNS[@]}"; do
    if grep -q "$pattern" configs/default.toml; then
        # Check if overridden in prod configs
        if ! grep -q "$pattern_override" configs/prod-live.toml; then
            FLAG: "Dangerous default not overridden: $pattern"
        fi
    fi
done
```

### Bad Example ❌

```toml
# configs/default.toml
[mozart]
mock = true          # ❌ Returns fake gateway responses
baseUrl = "http://localhost:8080"  # ❌ Local URL

[features]
autoRefundEnabled = true  # ❌ Experimental feature

[database]
timeout = 3          # ❌ Too short for production

# configs/prod-live.toml (MISSING OVERRIDES)
[mozart]
baseUrl = "https://mozart.razorpay.com"  # Fixed URL
# ❌ But mock still true from default!
```

### Good Example ✅

```toml
# configs/default.toml
[mozart]
mock = true          # OK for local dev
baseUrl = "http://localhost:8080"

# configs/prod-live.toml (ALL DANGEROUS VALUES OVERRIDDEN)
[mozart]
mock = false         # ✅ Disabled
baseUrl = "https://mozart.razorpay.com"  # ✅ Prod URL
timeout = 30         # ✅ Appropriate timeout
```

### Flag Conditions

Flag if new keys in default.toml have:
- `mock = true` for any service
- `debug = true` or `logLevel = "DEBUG"`
- Feature flag `enabled = true`
- `timeout < 10` for any service
- `baseUrl` with `localhost` or `127.0.0.1`

And these are NOT overridden in `prod-live.toml` / `prod-test.toml`

### Severity

🚨 **Critical** - Mock data in production, security leaks, feature incidents

### Output Format

```
🚨 Dangerous Default Values

configs/default.toml:
  Line 45: mozart.mock = true
  Line 48: autoRefund.enabled = true
  Line 52: database.timeout = 3

Danger: These dev-friendly defaults could be used in production!

Verification:
  ❌ prod-live.toml does not override mozart.mock
  ✅ prod-live.toml overrides autoRefund.enabled = false
  ❌ prod-live.toml does not override database.timeout

Fix: Add overrides to prod configs:
  [mozart]
  mock = false

  [database]
  timeout = 30
```

---

## Check 3: Sensitive Data in TOML 🚨 CRITICAL

### What to Check

Secrets, passwords, tokens, and API keys must NOT be hardcoded in TOML files. Use environment variables or Vault references.

### Detection Strategy

```bash
# Search for sensitive patterns
grep -iE "(password|secret|token|api_key|private_key|credentials)" configs/*.toml

# Check if values are hardcoded vs placeholders
# Good: password = "${DB_PASSWORD}" or password = ""
# Bad: password = "mySecretPassword123"
```

### Bad Example ❌

```toml
# configs/prod-live.toml
[database]
username = "admin"
password = "SuperSecret123!"  # ❌ Hardcoded password!

[mozart]
apiKey = "sk_live_abc123xyz"  # ❌ Hardcoded API key!

[aws]
accessKeyId = "AKIAIOSFODNN7EXAMPLE"  # ❌ AWS credentials!
secretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
```

### Good Example ✅

```toml
# Method 1: Environment variables
[database]
username = "${DB_USERNAME}"
password = "${DB_PASSWORD}"

# Method 2: Empty (loaded from Vault at runtime)
[mozart]
apiKey = ""  # Loaded from Vault

# Method 3: Explicit env var syntax (pg-router style)
[aws]
accessKeyId = "env|AWS_ACCESS_KEY_ID|default_value|slice"
```

### Flag Conditions

Flag if:
- Line contains `password`, `secret`, `token`, `apiKey`, `privateKey`
- AND value is NOT:
  - Empty string: `password = ""`
  - Environment variable: `password = "${VAR}"`
  - Vault placeholder: `password = "vault://secret/path"`
  - Env var syntax: `password = "env|VAR|default|type"`

### Severity

🚨 **Critical** - Security vulnerability, credential leak

---

## Check 4: Section Structure Mismatch ⚠️ HIGH

### What to Check

All environment configs should have consistent section structure. Missing sections can cause config parsing errors or unexpected defaults.

### Bad Example ❌

```toml
# configs/default.toml
[application]
appName = "terminals"

[database]
host = "localhost"

[database.master]
host = "localhost"
port = 5432

[database.replica]
host = "localhost"
port = 5433

# configs/prod-live.toml
[application]
appName = "terminals"

[database]
host = "prod-db.razorpay.vpc"
# ❌ Missing [database.master] and [database.replica] sections!
```

**Problem:** Code expects `database.master` config, gets nothing or nil

### Detection Strategy

```bash
# Extract all section headers from default.toml
grep "^\[" configs/default.toml > /tmp/default_sections.txt

# For each critical config, verify all sections exist
for config in "${CRITICAL_CONFIGS[@]}"; do
    while read section; do
        if ! grep -q "$section" "configs/$config"; then
            FLAG: "Section $section missing in $config"
        fi
    done < /tmp/default_sections.txt
done
```

### Severity

⚠️ **High** - Config parsing errors, nil pointer dereference

---

## Check 5: Value Type Consistency ⚠️ HIGH

### What to Check

Same key should have same type across all configs (string vs int vs bool).

### Bad Example ❌

```toml
# configs/default.toml
[database]
port = 5432  # Integer

maxConnections = "25"  # String

# configs/prod-live.toml
[database]
port = "5432"  # ❌ String instead of int!

maxConnections = 100  # ❌ Int instead of string!
```

**Problem:** Type assertion failures at runtime

### Detection Strategy

```bash
# Parse TOML and compare types
# For each key in default.toml:
#   - Extract value and determine type (int, string, bool)
#   - Verify same type in all configs
```

### Severity

⚠️ **High** - Runtime errors, type assertion panics

---

## Check 6: Environment-Specific Values 📋 MEDIUM

### What to Check

Ensure environment-specific values are actually different (not copy-pasted).

### Bad Example ❌

```toml
# configs/stage-live.toml
[database]
host = "prod-db.razorpay.vpc"  # ❌ Copy-pasted prod value!

# configs/prod-live.toml
[database]
host = "prod-db.razorpay.vpc"

# Both point to prod database!
```

### Detection Strategy

```bash
# Compare critical values across stage/prod
# Flag if identical when they should differ:
# - database host
# - kafka brokers
# - redis host
# - external service URLs
```

### Severity

📋 **Medium** - Staging uses production resources, dangerous

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | Missing prod keys | 🚨 Critical | Prod uses dev defaults |
| 2 | Dangerous defaults | 🚨 Critical | Mock/debug in prod |
| 3 | Hardcoded secrets | 🚨 Critical | Security breach |
| 4 | Section mismatch | ⚠️ High | Config parse errors |
| 5 | Type mismatch | ⚠️ High | Runtime panics |
| 6 | Env value duplication | 📋 Medium | Wrong resources used |

---

## Critical Configs Priority

**Always check these (in order):**
1. `prod-live.toml` - Production live mode (most critical)
2. `prod-test.toml` - Production test mode
3. `stage-live.toml` - Pre-production validation
4. `stage-test.toml` - Staging tests

**Can skip:**
- `dev.toml`, `automation-*.toml`, `bvt-*.toml`, `func-*.toml`, `perf-*.toml`
- These are test environments, less critical

---

## Example Output

```
📁 File: configs/default.toml

🚨 Check #1 Failed: Missing keys in production configs
   New section: [newFeature]
   Keys: enabled, debugMode, timeout
   Missing from:
     - configs/prod-live.toml (3 keys)
     - configs/prod-test.toml (3 keys)

🚨 Check #2 Failed: Dangerous default value
   Line 45: newFeature.enabled = true
   Risk: Untested feature could activate in production
   Fix: Override to false in prod configs

⚠️  Check #4 Failed: Section structure mismatch
   Section [database.replica] in default.toml
   Missing from: configs/stage-live.toml

✅ Check #3 Passed: No hardcoded secrets
✅ Check #5 Passed: Types consistent
✅ Check #6 Passed: Environment values differ
```
