# Feature Flag Quality Checks

## Overview

Validates that major features are gated behind feature flags (Splitz) for safe rollout and easy rollback.

**Load when:** PR adds significant new features or changes critical business logic

**Total Checks:** 1

**Severity Distribution:**
- ⚠️ High: 1

---

## Check 1: Major Features Behind Splitz ⚠️ HIGH

### What to Check

Significant new features must be wrapped in Splitz (feature flags) to enable:
- Gradual rollout
- A/B testing
- Quick rollback on issues

### What Qualifies as "Major Feature"

**Requires Splitz:**
- ✅ New payment flows
- ✅ New API endpoints (especially write operations)
- ✅ Algorithm changes (pricing, routing, matching)
- ✅ Third-party integrations
- ✅ Database schema changes with data migration
- ✅ Changes to critical paths (checkout, payment processing)
- ✅ New authentication/authorization logic
- ✅ Experimental features

**Doesn't require Splitz:**
- ❌ Bug fixes
- ❌ Refactoring without behavior change
- ❌ Configuration changes
- ❌ Logging/monitoring improvements
- ❌ Internal tools/admin features

### Bad Pattern ❌

```go
// ANTI-PATTERN: Major feature without Splitz

// PR adds new payment routing algorithm
func RoutePayment(payment *Payment) (*Gateway, error) {
    // ❌ New routing logic enabled for ALL merchants immediately!
    // ❌ No way to rollback without code deploy

    // NEW ALGORITHM (risky!)
    if payment.Amount > 100000 {
        return findCheapestGateway(payment)  // New logic
    }

    return findFastestGateway(payment)
}
```

**Problem:**
- Enabled for all users immediately
- Issues affect entire user base
- Rollback requires code deploy (slow)
- Can't A/B test effectiveness

### Good Pattern ✅

```go
// CORRECT: Major feature behind Splitz

func RoutePayment(ctx *gin.Context, payment *Payment) (*Gateway, error) {
    // ✅ Feature flag for new routing algorithm
    useNewRouting := splitz.IsEnabled(ctx,
        "enable_new_payment_routing",
        payment.MerchantID)

    if useNewRouting {
        logger.Info(ctx, "using_new_routing_algorithm", "merchant_id", payment.MerchantID)

        // New algorithm (gradual rollout)
        if payment.Amount > 100000 {
            return findCheapestGateway(payment)
        }
    } else {
        // Old algorithm (fallback)
        logger.Info(ctx, "using_old_routing_algorithm")
    }

    return findFastestGateway(payment)
}
```

**Benefits:**
- ✅ Enable for 1% of merchants first
- ✅ Monitor metrics (success rate, latency)
- ✅ Gradually increase to 10%, 50%, 100%
- ✅ Instant rollback (disable in Splitz dashboard)
- ✅ A/B test new vs old algorithm

### Example Splitz Configuration

```go
// config/splitz.go

var FeatureFlags = map[string]SplitzConfig{
    "enable_new_payment_routing": {
        Description: "New payment routing algorithm (cost-optimized)",
        DefaultValue: false,  // ✅ Disabled by default
        Rollout: RolloutConfig{
            // ✅ Gradual rollout plan
            Phase1: {Percentage: 1},   // Enable for 1%
            Phase2: {Percentage: 10},  // Increase to 10% after 2 days
            Phase3: {Percentage: 50},  // Increase to 50% after 1 week
            Phase4: {Percentage: 100}, // Full rollout after 2 weeks
        },
        Metrics: []string{
            "payment_success_rate",
            "payment_latency_p99",
            "gateway_cost_per_transaction",
        },
    },

    "enable_auto_refund": {
        Description: "Automatic refund for failed payments",
        DefaultValue: false,
        RequiresApproval: true,  // ✅ Requires manual approval
    },
}
```

### Detection Strategy

```bash
# Signal 1: New API endpoints without Splitz in the PR
# Line count is a poor signal — a 200-line refactor needs no flag; a 10-line new endpoint does.
# Use diff with file context (-u) and track the current file from '--- a/' headers,
# so each endpoint is checked against the file it actually lives in.
git diff main..HEAD -U0 | awk '
    /^--- a\// { current_file = substr($0, 6) }
    /^\+.*router\.(POST|PUT|DELETE|PATCH)\(/ { print current_file ":" $0 }
' | grep -v '^---' | while IFS=: read src_file line; do
    endpoint=$(echo "$line" | grep -oE '"[^"]+"' | head -1)
    if [ -n "$src_file" ] && ! git diff main..HEAD -- "$src_file" | \
           grep -qE "splitz\.|IsEnabled|FeatureFlag|SplitzEnabled"; then
        echo "⚠️  New endpoint $endpoint in $src_file — no Splitz feature flag found"
    fi
done

# Signal 2: New exported functions with control flow in critical-path files without Splitz
# Filter by file path (reliable). Narrow to functions with actual branching logic —
# skip constructors, String/Validate/Error helpers that add no rollout-worthy behavior.
git diff main..HEAD --name-only | \
    grep -iE 'service|handler|routing|pricing|checkout' | \
    while read file; do
        # Find new exported functions added in this file
        new_funcs=$(git diff main..HEAD -- "$file" | grep -E '^\+func [A-Z]' | grep -v '^---' | \
            grep -vE 'func (New[A-Z]|String|Validate|Error|MarshalJSON|UnmarshalJSON)\(')
        if [ -n "$new_funcs" ]; then
            # Only flag if the new function body contains branching logic (if/switch/for)
            has_branches=$(git diff main..HEAD -- "$file" | grep -E '^\+\s+(if |switch |for )' | grep -v '^---')
            if [ -n "$has_branches" ] && ! git diff main..HEAD -- "$file" | \
                   grep -qE "splitz\.|IsEnabled|FeatureFlag|SplitzEnabled"; then
                echo "⚠️  New function(s) with control flow in $file — verify if Splitz is needed:"
                echo "$new_funcs" | grep -oE 'func [A-Za-z]+'
            fi
        fi
    done

# Signal 3: New business-logic branches in core files — check per-file for Splitz
# Narrow to branches that implement routing/selection decisions, not nil/zero guards.
# Exclude common validation patterns (== nil, == 0, == "", err != nil) which appear
# in almost every handler and are not rollout-worthy.
git diff main..HEAD --name-only -- internal/services/ internal/routing/ internal/pricing/ | \
    while read file; do
        new_branches=$(git diff main..HEAD -- "$file" | \
            grep -E '^\+\s+if ' | grep -v '^---' | \
            grep -vE '(== nil|!= nil|== 0|!= 0|== ""|err != nil|len\()' | \
            grep -E '(Gateway|Route|Method|Instrument|Provider|Channel|Processor)')
        if [ -n "$new_branches" ]; then
            if ! git diff main..HEAD -- "$file" | grep -qE "splitz\.|IsEnabled|FeatureFlag"; then
                echo "⚠️  $file: New routing/selection logic without Splitz — consider feature flag:"
                echo "$new_branches" | head -3
            fi
        fi
    done
```

> **Rationale:** The previous `> 50 lines changed` threshold generates false positives for refactors, comment changes, and test additions. These signals instead look for structural additions (new endpoints, new exported functions, new branch conditions) that are more reliable indicators of new rollout-worthy behavior.

### Flag Conditions

Flag if:
- New API endpoint (POST/PUT/DELETE/PATCH) added without a Splitz check in its handler
- New exported function in `services/`, `routing/`, or `pricing/` packages touching critical logic
- New conditional branch in routing/pricing/checkout logic without a corresponding feature flag
- Third-party integration added without a kill-switch
- No `splitz.IsEnabled` / `FeatureFlag` anywhere in the changed critical-path files

Do NOT flag if:
- Change is a refactor with no new branches or endpoints (pure restructuring)
- Change is test-only, logging, or monitoring
- Change is a bug fix to existing guarded behavior

### Severity

⚠️ **High** - Risky deployment, no gradual rollout

### Output Format

```
⚠️  Major Feature Without Splitz

File: internal/services/payment_router.go
Changes: 127 lines modified

New functionality detected:
  - New payment routing algorithm (line 45-89)
  - Changes critical path: RoutePayment()

Recommendation:
  1. Wrap in Splitz feature flag:
     useNewRouting := splitz.IsEnabled(ctx, "enable_new_routing", merchantId)

  2. Add Splitz config in config/splitz.go:
     "enable_new_routing": {
         Description: "New cost-optimized routing",
         DefaultValue: false,
         Rollout: GradualRollout{...},
     }

  3. Rollout plan:
     - Phase 1: Enable for 1% (day 1)
     - Phase 2: 10% (day 3)
     - Phase 3: 50% (day 7)
     - Phase 4: 100% (day 14)

  4. Monitor metrics:
     - payment_success_rate
     - payment_latency_p99
     - gateway_cost_per_transaction
```

---

## When to Skip Splitz

You can skip Splitz if:
1. **Bug fix** - Fixing broken functionality
2. **Urgent hotfix** - Critical production issue
3. **Internal tool** - Not customer-facing
4. **Configuration change** - No code logic change
5. **Backward compatible** - Guaranteed safe change

In these cases, add comment explaining why Splitz is not needed:

```go
// No Splitz needed: Bug fix for existing functionality
func FixCalculation() {
    // ... fix
}
```

---

## Splitz Best Practices

1. **Default to disabled**
   ```go
   DefaultValue: false  // ✅ Safe default
   ```

2. **Gradual rollout**
   ```go
   // ✅ Start small
   Phase1: 1%  → Phase2: 10% → Phase3: 50% → Phase4: 100%
   ```

3. **Monitor metrics**
   ```go
   Metrics: ["success_rate", "latency", "error_rate"]
   ```

4. **Cleanup old flags**
   ```go
   // ⚠️ TODO: Remove after 100% rollout (2025-02-01)
   useNewFeature := splitz.IsEnabled(...)
   ```

5. **Document rollout plan**
   ```go
   // Rollout Plan:
   // - 2025-01-15: Enable for 1%
   // - 2025-01-17: Increase to 10%
   // - 2025-01-22: Increase to 50%
   // - 2025-01-29: Full rollout (100%)
   ```

---

## Summary

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | Major features behind Splitz | ⚠️ High | Risky deployment, no rollback |

---

## How to Apply

**For each PR:**

1. Detect significant code changes (50+ lines in critical paths)
2. Check for new API endpoints
3. Verify Splitz is used for major features
4. Suggest Splitz config if missing

**Example output:**

```
📁 File: internal/handlers/payment_handler.go

⚠️  Check #1 Failed: Major feature without Splitz

New endpoint: POST /payments/auto-refund (Line 45)
Changes: 89 lines

Recommendation: Add Splitz flag
  splitz.IsEnabled(ctx, "enable_auto_refund", merchantId)

✅ File: internal/services/logger.go
    Changes: 67 lines (logging improvement)
    No Splitz needed (not customer-facing)
```
