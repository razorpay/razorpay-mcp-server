# Services: Splitz Integration (10 Checks)

Validates Splitz A/B experimentation and feature flag integration patterns. Focuses on ensuring experiments are properly configured and activated for target environments.

**Total Checks**: 10 (8 Critical, 2 High)

---

## MCP Setup (Required for Checks #1-5, #8)

**IMPORTANT**: Checks #1, #2, #3, #4, #5, and #8 require the Splitz MCP server to validate experiments against live Splitz configuration.

### Step 1: Check if Splitz MCP is Available

Try to use a Splitz MCP tool to verify it's installed:

```
Attempt to call: mcp__splitz-prod__splitz_guide with doc=["experiment_doc"]
```

**If successful**: MCP is installed ✅ - Proceed with all checks

**If failed (tool not found)**: MCP not installed ❌ - Follow installation steps below

### Step 2: Install Splitz MCP (if not available)

Show the user this message and wait for confirmation:

```
⚠️  Splitz MCP Not Installed

To enable full Splitz validation (Checks #1-5, #8), you need to install the Splitz MCP server.

Installation Steps:
1. Run this command in your terminal:

   claude mcp add --scope user --transport http splitz-prod "https://splitz-concierge.razorpay.com/mcp"

2. Restart Claude Code:
   - Exit current session: Type 'exit' or press Ctrl+D
   - Start new session: Run 'claude' command
   - Resume this conversation

3. Verify installation by asking me to check Splitz MCP again

Without MCP:
- Checks #1, #2, #3, #4, #5, #8 will be SKIPPED
- Checks #6, #7, #9, #10 will still run (code-based validation only)

Would you like to install Splitz MCP now? (Recommended for full validation)
```

Wait for user response before proceeding.

### Step 3: Verify MCP Installation

After user restarts and resumes, verify MCP is working:

```
Test: mcp__splitz-prod__splitz_guide with doc=["experiment_doc"]
```

**Expected Response**: Should return Splitz experiment documentation

**If still failing**:
- Check if user restarted Claude Code
- Verify command was run correctly
- Check `~/.claude.json` has splitz-prod entry
- Suggest user check MCP server URL is accessible

### Step 4: Proceed with Checks

Once MCP is verified:
- Run all 10 checks (full validation) ✅
- Use MCP to fetch experiment details for each referenced experiment
- Validate state, variants, audience rules against code

If MCP not installed and user chooses to skip:
- Run only Checks #6, #7, #9, #10 (code-based validation)
- Add warning in summary: "⚠️ Partial validation only - Install Splitz MCP for full coverage"

---

## MCP-Dependent vs Code-Only Checks

### Require Splitz MCP (6 checks):
1. ✅ Check #1: Experiment Existence - Verifies experiment exists in Splitz
2. ✅ Check #2: Environment Activation - Validates experiment is activated for prod/stage
3. ✅ Check #3: Variant Configuration Match - Compares code variants with Splitz config
4. ✅ Check #4: Default Variant Safety - Validates default matches Splitz control variant
5. ✅ Check #5: RequestData Schema Validation - Checks audience rule compatibility
6. ✅ Check #8: Bulk Evaluation Validation - Verifies all bulk experiments exist

### Work Without MCP (4 checks):
7. ✅ Check #6: Evaluation Error Handling - Code pattern analysis
8. ✅ Check #7: Identifier Consistency - Code pattern analysis
9. ✅ Check #9: Environment Configuration - Check env vars and config
10. ✅ Check #10: Evaluation Test Coverage - Test file analysis

---

## Check #1: Experiment Existence (Critical)

**Problem**: Code references Splitz experiment that doesn't exist or has incorrect ID/name.

**Detection Strategy**:
1. Extract all experiment references from code:
   ```bash
   # Find ExperimentId references
   grep -E 'ExperimentId:\s*"[^"]+"' <pr_files>

   # Find ExperimentName references
   grep -E 'ExperimentName:\s*"[^"]+"' <pr_files>
   ```

2. For each experiment found, use Splitz MCP to verify existence:
   ```
   Use mcp__splitz-prod__get_experiment with the experiment ID or name
   ```

3. Check if experiment was found or returned error

**What to Flag**:
```go
// ❌ BAD: Experiment doesn't exist in Splitz
evalRequest := splitz.EvaluateRequest{
    Id:           merchantID,
    ExperimentId: "NonExistentExp123",  // <-- MCP returns error
}
```

**How to Fix**:
```go
// ✅ GOOD: Experiment exists in Splitz
evalRequest := splitz.EvaluateRequest{
    Id:           merchantID,
    ExperimentId: "RDqg4Gh9Bfxero",  // <-- MCP confirms exists
}
```

**Rationale**: References to non-existent experiments will always return nil/errors at runtime.

---

## Check #2: Environment Activation (Critical)

**Problem**: Experiment exists but not activated for prod or stage environment (the core issue).

**Detection Strategy**:
1. Identify target environment from PR context:
   - Check deployment config files
   - Check if code is in prod/stage deployment path
   - Ask user if unclear

2. For each experiment, use Splitz MCP to get experiment details:
   ```
   Use mcp__splitz-prod__get_experiment to fetch experiment configuration
   ```

3. Check experiment state and environment configuration:
   - State should be "Activated" or "Scheduled" (not "Created" or "Terminated")
   - Project metadata should include target environment
   - If experiment has environment-specific config, verify it matches deployment target

**What to Flag**:
```go
// ❌ BAD: Experiment in "Created" state, not activated
// MCP shows: "state": "Created", "environment": "dev"
evalRequest := splitz.EvaluateRequest{
    Id:           merchantID,
    ExperimentId: "RDqg4Gh9Bfxero",
}

// ❌ BAD: Experiment terminated
// MCP shows: "state": "Terminated"
evalRequest := splitz.EvaluateRequest{
    Id:           merchantID,
    ExperimentId: "OldExperiment123",
}
```

**How to Fix**:
```go
// ✅ GOOD: Experiment in "Activated" state for prod
// MCP shows: "state": "Activated", "environment": "production"
evalRequest := splitz.EvaluateRequest{
    Id:           merchantID,
    ExperimentId: "RDqg4Gh9Bfxero",
}
```

**Razorpay Standard**:
- Experiments must be in "Activated" state before code deployment
- Stage experiments should be activated in staging environment first
- Prod experiments require approval before activation

**Rationale**: Code deployed with non-activated experiments will not work as expected in production.

---

## Check #3: Variant Configuration Match (Critical)

**Problem**: Code references variant names/IDs that don't exist in experiment configuration.

**Detection Strategy**:
1. Extract variant references from code:
   ```bash
   # Find variant name comparisons
   grep -E 'variant\.Name\s*==\s*"[^"]+"' <pr_files>
   grep -E 'variant\.Id\s*==\s*"[^"]+"' <pr_files>

   # Find variant variable key access
   grep -E 'variable\.Key\s*==\s*"[^"]+"' <pr_files>
   ```

2. Use Splitz MCP to get experiment configuration and available variants

3. Verify all referenced variants exist in experiment

**What to Flag**:
```go
// ❌ BAD: Variant name doesn't exist in experiment
// MCP shows variants: ["control", "treatment_v1", "treatment_v2"]
variant, _ := client.GetVariant(ctx, evalRequest)
if variant.Name == "treatment_v3" {  // <-- Doesn't exist!
    // This will never execute
}

// ❌ BAD: Variable key doesn't exist
// MCP shows variables: [{"key": "button_color", ...}]
for _, variable := range variant.Variables {
    if variable.Key == "button_size" {  // <-- Doesn't exist!
        // This will never match
    }
}
```

**How to Fix**:
```go
// ✅ GOOD: Variant names match experiment config
// MCP confirms variants: ["control", "treatment_v1"]
variant, _ := client.GetVariant(ctx, evalRequest)
if variant.Name == "treatment_v1" {  // <-- Exists!
    // Use treatment behavior
}

// ✅ GOOD: Variable keys match experiment config
for _, variable := range variant.Variables {
    if variable.Key == "button_color" {  // <-- Exists!
        color := variable.Value.(string)
    }
}
```

**Rationale**: Mismatched variant names lead to dead code and unexpected fallback behavior.

---

## Check #4: Default Variant Safety (High)

**Problem**: GetVariantOrDefault uses default variant that doesn't match experiment configuration.

**Detection Strategy**:
1. Find GetVariantOrDefault calls:
   ```bash
   grep -A 10 "GetVariantOrDefault" <pr_files>
   ```

2. Extract default variant structure from code

3. Use Splitz MCP to get actual experiment configuration

4. Verify default variant structure matches experiment schema:
   - Same variable keys
   - Compatible value types
   - Represents actual control variant

**What to Flag**:
```go
// ❌ BAD: Default variant doesn't match experiment schema
// MCP shows control variant has: [{"key": "enabled", "value": true}]
defaultVariant := splitz.Variant{
    Id:   "control",
    Name: "control_variant",
    Variables: []splitz.Variable{
        {Key: "feature_enabled", Value: false},  // <-- Wrong key name!
    },
}
variant, _ := client.GetVariantOrDefault(ctx, evalRequest, defaultVariant)

// ❌ BAD: Default uses treatment variant instead of control
defaultVariant := splitz.Variant{
    Id:   "treatment",  // <-- Should use control!
    Name: "treatment_variant",
}
```

**How to Fix**:
```go
// ✅ GOOD: Default matches actual control variant from MCP
// MCP confirms control variant structure
defaultVariant := splitz.Variant{
    Id:   "control",
    Name: "control_variant",
    Variables: []splitz.Variable{
        {Key: "enabled", Value: false},  // <-- Matches MCP schema
    },
}
variant, _ := client.GetVariantOrDefault(ctx, evalRequest, defaultVariant)
```

**Razorpay Standard**:
- Default variant should always represent the "control" (current behavior)
- Default should be safe fallback if experiment fails
- Keep default variant in sync with Splitz configuration

**Rationale**: Mismatched defaults can cause unexpected behavior when Splitz is unavailable.

---

## Check #5: RequestData Schema Validation (High)

**Problem**: RequestData JSON doesn't match audience targeting rules in experiment.

**Detection Strategy**:
1. Extract RequestData from code:
   ```bash
   grep -A 5 "RequestData:" <pr_files>
   ```

2. Use Splitz MCP to get experiment audience rules:
   ```
   Get experiment configuration and check audience targeting criteria
   ```

3. Verify RequestData fields match audience rule requirements:
   - All required fields present
   - Field types compatible
   - Field names match exactly (case-sensitive)

**What to Flag**:
```go
// ❌ BAD: Audience rule expects "merchant_id" but code sends "merchantId"
// MCP shows audience: 'merchant_id == "premium_merchant"'
evalRequest := splitz.EvaluateRequest{
    Id:           userID,
    ExperimentId: "RDqg4Gh9Bfxero",
    RequestData:  `{"merchantId": "12345"}`,  // <-- Wrong field name!
}

// ❌ BAD: Missing required field for audience evaluation
// MCP shows audience: 'country == "IN" AND merchant_tier == "premium"'
evalRequest := splitz.EvaluateRequest{
    Id:           userID,
    ExperimentId: "RDqg4Gh9Bfxero",
    RequestData:  `{"country": "IN"}`,  // <-- Missing merchant_tier!
}
```

**How to Fix**:
```go
// ✅ GOOD: RequestData matches audience rule schema
// MCP confirms audience fields: merchant_id, country
evalRequest := splitz.EvaluateRequest{
    Id:           userID,
    ExperimentId: "RDqg4Gh9Bfxero",
    RequestData:  `{"merchant_id": "12345", "country": "IN"}`,  // <-- Matches!
}

// ✅ GOOD: No RequestData when experiment has no audience rules
// MCP shows no audience targeting
evalRequest := splitz.EvaluateRequest{
    Id:           userID,
    ExperimentId: "SimpleRamping",
    RequestData:  nil,  // <-- Correct for no audience rules
}
```

**Rationale**: Mismatched RequestData causes evaluation to fail or return wrong variant.

---

## Check #6: Evaluation Error Handling (Critical)

**Problem**: Code doesn't handle Splitz evaluation errors, leading to crashes or incorrect behavior.

**Detection Strategy**:
```bash
# Find GetVariant calls without error handling
grep -A 5 "GetVariant" <pr_files> | grep -B 5 -v "if err"

# Find nil variant access without checking
grep -A 10 "GetVariant" <pr_files> | grep "variant\." | grep -v "variant != nil"
```

**What to Flag**:
```go
// ❌ BAD: No error handling
variant, _ := client.GetVariant(ctx, evalRequest)
if variant.Name == "treatment" {  // <-- Panic if variant is nil!
    enableFeature()
}

// ❌ BAD: Error ignored, nil variant accessed
variant, err := client.GetVariant(ctx, evalRequest)
// No check for err or variant != nil
color := variant.Variables[0].Value  // <-- Panic if nil!

// ❌ BAD: Incorrect assumption that GetVariant never returns nil
variant, err := client.GetVariant(ctx, evalRequest)
if err == nil {
    // Assumes variant is non-nil, but it can be nil even without error!
    doSomething(variant.Name)
}
```

**How to Fix**:
```go
// ✅ GOOD: Proper error and nil handling
variant, err := client.GetVariant(ctx, evalRequest)
if err != nil {
    log.Error(ctx, "splitz_eval_error", "error", err)
    // Fallback to default behavior
    return defaultBehavior()
}

if variant != nil && variant.Name == "treatment" {
    enableFeature()
}

// ✅ BETTER: Use GetVariantOrDefault for guaranteed non-nil
variant, err := client.GetVariantOrDefault(ctx, evalRequest, defaultVariant)
if err != nil {
    log.Error(ctx, "splitz_eval_error", "error", err)
    // variant still usable as default
}
// No nil check needed - variant is always non-nil
if variant.Name == "treatment" {
    enableFeature()
}
```

**Razorpay Standard**:
- Always check both error and nil variant for GetVariant
- Prefer GetVariantOrDefault when you have a clear default behavior
- Log errors with proper context for debugging
- Never let Splitz errors crash the application

**Rationale**: Splitz can fail due to network issues, timeouts, or misconfiguration. Code must handle failures gracefully.

---

## Check #7: Identifier Consistency (Critical)

**Problem**: Inconsistent entity ID usage leads to different bucketing for same user.

**Detection Strategy**:
1. Find all Splitz evaluation calls in the same service/handler
2. Check if the same user/entity uses consistent `Id` field:
   ```bash
   grep -E 'Id:\s*[^,]+' <pr_files>
   ```

3. Look for mixing different ID types (userID, merchantID, sessionID)

**What to Flag**:
```go
// ❌ BAD: Inconsistent IDs for same user across evaluations
func checkoutHandler(userID, merchantID string) {
    // First experiment uses userID
    variant1, _ := client.GetVariant(ctx, splitz.EvaluateRequest{
        Id:           userID,
        ExperimentId: "CheckoutRedesign",
    })

    // Second experiment uses merchantID for same user context
    variant2, _ := client.GetVariant(ctx, splitz.EvaluateRequest{
        Id:           merchantID,  // <-- Different ID type!
        ExperimentId: "PaymentFlow",
    })
}

// ❌ BAD: Dynamic ID construction can cause bucketing issues
variant, _ := client.GetVariant(ctx, splitz.EvaluateRequest{
    Id:           fmt.Sprintf("%s-%d", userID, time.Now().Unix()),  // <-- Changes every call!
    ExperimentId: "Feature",
})
```

**How to Fix**:
```go
// ✅ GOOD: Consistent ID type for user-level experiments
func checkoutHandler(userID, merchantID string) {
    // Both experiments use userID consistently
    variant1, _ := client.GetVariant(ctx, splitz.EvaluateRequest{
        Id:           userID,  // <-- Consistent
        ExperimentId: "CheckoutRedesign",
    })

    variant2, _ := client.GetVariant(ctx, splitz.EvaluateRequest{
        Id:           userID,  // <-- Consistent
        ExperimentId: "PaymentFlow",
    })
}

// ✅ GOOD: Use merchantID for merchant-level experiments
variant, _ := client.GetVariant(ctx, splitz.EvaluateRequest{
    Id:           merchantID,  // <-- Correct for merchant experiment
    ExperimentId: "MerchantDashboard",
})

// ✅ GOOD: Static ID for deterministic bucketing
variant, _ := client.GetVariant(ctx, splitz.EvaluateRequest{
    Id:           userID,  // <-- Deterministic
    ExperimentId: "Feature",
})
```

**Razorpay Standard**:
- User-level experiments: Use `userID` or `customerId`
- Merchant-level experiments: Use `merchantID`
- Request-level experiments: Use `requestID` (only for request-scoped features)
- Never use time-based or random IDs

**Rationale**: Inconsistent IDs cause users to see different variants on each request, invalidating experiment results.

---

## Check #8: Bulk Evaluation Validation (Critical)

**Problem**: GetVariantInBulk references experiments that don't exist or uses inconsistent entity IDs.

**Detection Strategy**:
1. Find bulk evaluation calls:
   ```bash
   grep -A 20 "GetVariantInBulk" <pr_files>
   ```

2. Extract all experiment IDs from bulk request

3. Use Splitz MCP to verify all experiments exist and are activated

4. Check if all requests use same entity ID

**What to Flag**:
```go
// ❌ BAD: Bulk request mixes different entity IDs
evaluateRequests := []*splitz.EvaluateRequest{
    {
        Id:           userID,  // <-- User ID
        ExperimentId: "Exp1",
    },
    {
        Id:           merchantID,  // <-- Merchant ID (inconsistent!)
        ExperimentId: "Exp2",
    },
}

// ❌ BAD: One experiment in bulk doesn't exist
// MCP returns error for "NonExistentExp"
evaluateRequests := []*splitz.EvaluateRequest{
    {Id: userID, ExperimentId: "ValidExp"},
    {Id: userID, ExperimentId: "NonExistentExp"},  // <-- Doesn't exist!
}
```

**How to Fix**:
```go
// ✅ GOOD: Consistent entity ID across bulk request
evaluateRequests := []*splitz.EvaluateRequest{
    {
        Id:           userID,  // <-- Consistent
        ExperimentId: "CheckoutRedesign",
    },
    {
        Id:           userID,  // <-- Consistent
        ExperimentId: "PaymentFlow",
    },
}

// ✅ GOOD: All experiments exist and are activated (verified via MCP)
variantsMap, err := client.GetVariantInBulk(ctx, bulkRequest)
if err != nil {
    log.Error(ctx, "bulk_eval_error", "error", err)
}

// Access variants safely
if variant := variantsMap["CheckoutRedesign"]; variant != nil {
    // Use variant
}
```

**Razorpay Standard**:
- Use bulk evaluation for multiple experiments evaluated for same entity
- All experiments in bulk request should target same entity type
- Validate all experiment IDs before deploying code

**Rationale**: Bulk evaluation failures can be silent. Pre-validating prevents runtime issues.

---

## Check #9: Environment Configuration (High)

**Problem**: Missing or incorrect Splitz environment configuration prevents evaluation.

**Detection Strategy**:
1. Check if code initializes Splitz client:
   ```bash
   grep -E 'splitz\.New|splitz\.Config' <pr_files>
   ```

2. Verify environment variable usage:
   ```bash
   grep -E 'SPLITZ_ENDPOINT|SPLITZ_KEY|SPLITZ_SECRET' <pr_files>
   ```

3. Check deployment configs (K8s manifests, env files) for required variables

**What to Flag**:
```go
// ❌ BAD: Hardcoded credentials (security issue)
config := splitz.Config{
    Endpoint: "https://splitz.razorpay.com",
    Key:      "hardcoded_key",  // <-- Security risk!
    Secret:   "hardcoded_secret",
}

// ❌ BAD: Missing environment variables
config := splitz.Config{
    Endpoint: os.Getenv("SPLITZ_ENDPOINT"),  // <-- Could be empty!
    Key:      os.Getenv("SPLITZ_KEY"),
    Secret:   os.Getenv("SPLITZ_SECRET"),
}
// No validation if variables exist

// ❌ BAD: Wrong environment URL
config := splitz.Config{
    Endpoint: "https://splitz-dev.razorpay.com",  // <-- Dev URL in prod code!
}
```

**How to Fix**:
```go
// ✅ GOOD: Use environment variables with validation
func initSplitzClient() (*splitz.Client, error) {
    endpoint := os.Getenv("SPLITZ_ENDPOINT")
    key := os.Getenv("SPLITZ_KEY")
    secret := os.Getenv("SPLITZ_SECRET")

    if endpoint == "" || key == "" || secret == "" {
        return nil, errors.New("missing Splitz configuration")
    }

    config := splitz.Config{
        Endpoint: endpoint,
        Key:      key,
        Secret:   secret,
    }

    return splitz.NewClient(config)
}

// ✅ GOOD: Check deployment config has required variables
// In kubernetes/deployment.yaml or .env file:
// SPLITZ_ENDPOINT=https://splitz.razorpay.com
// SPLITZ_KEY={{ secret "splitz-key" }}
// SPLITZ_SECRET={{ secret "splitz-secret" }}
```

**Razorpay Standard**:
- Always use environment variables for Splitz configuration
- Validate required variables at service startup
- Use correct endpoint for environment:
  - Dev: `https://splitz-dev.razorpay.com`
  - Stage: `https://splitz-stage.razorpay.com`
  - Prod: `https://splitz.razorpay.com`
- Store credentials in Vault/K8s secrets, never in code

**Rationale**: Missing or incorrect configuration causes all Splitz calls to fail.

---

## Check #10: Evaluation Test Coverage (Critical)

**Problem**: Code has Splitz integration but no tests validating evaluation behavior.

**Detection Strategy**:
1. Find Splitz evaluation code in PR
2. Check if corresponding test file exists and has Splitz tests:
   ```bash
   # For file foo.go, check if foo_test.go exists
   # Search for Splitz mocking in tests
   grep -E 'mock.*splitz|splitz.*mock' <test_files>
   grep -E 'GetVariant|GetVariantOrDefault' <test_files>
   ```

3. Verify tests cover both success and error cases

**What to Flag**:
```go
// ❌ BAD: Production code has Splitz but no tests
// File: checkout_handler.go
func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
    variant, _ := splitzClient.GetVariant(ctx, evalRequest)
    if variant.Name == "new_flow" {
        handleNewCheckoutFlow(w, r)
    }
}

// File: checkout_handler_test.go
// No Splitz-related tests found!
```

**How to Fix**:
```go
// ✅ GOOD: Tests cover Splitz evaluation scenarios
// File: checkout_handler_test.go
func TestCheckoutHandler_NewFlow(t *testing.T) {
    mockClient := &MockSplitzClient{
        variant: &splitz.Variant{Name: "new_flow"},
        err:     nil,
    }

    // Test new flow behavior
}

func TestCheckoutHandler_ControlFlow(t *testing.T) {
    mockClient := &MockSplitzClient{
        variant: &splitz.Variant{Name: "control"},
        err:     nil,
    }

    // Test control flow behavior
}

func TestCheckoutHandler_SplitzError(t *testing.T) {
    mockClient := &MockSplitzClient{
        variant: nil,
        err:     errors.New("splitz unavailable"),
    }

    // Test error handling and fallback
}

func TestCheckoutHandler_NilVariant(t *testing.T) {
    mockClient := &MockSplitzClient{
        variant: nil,  // No matching variant
        err:     nil,
    }

    // Test nil variant handling
}
```

**Razorpay Standard**:
- Mock Splitz client in unit tests
- Test all variant scenarios (control, treatment, etc.)
- Test error handling (network failure, timeout)
- Test nil variant handling
- Use table-driven tests for multiple variants

**Rationale**: Untested Splitz integration leads to production failures when experiments activate.

---

## Integration Instructions

### When to Load This File
Load when PR contains any of:
- `import.*splitz` or `github.com/razorpay/goutils/splitz`
- `GetVariant`, `GetVariantOrDefault`, `GetVariantInBulk`
- `ExperimentId`, `ExperimentName`
- `splitz.Client` or `splitz.EvaluateRequest`

### Progressive Loading
Only load this file if Splitz-related code changes detected. Defer loading until actual Splitz usage confirmed to reduce context.

### Pre-Flight Check: Verify Splitz MCP Availability

**BEFORE running Checks #1-5 or #8**, verify MCP is available:

```
Step 1: Attempt MCP call
Try: mcp__splitz-prod__splitz_guide with doc=["experiment_doc"]

Step 2: Handle result
If SUCCESS:
  - Set flag: splitz_mcp_available = true
  - Proceed with all 10 checks

If FAILURE (tool not found):
  - Set flag: splitz_mcp_available = false
  - Show installation instructions (see MCP Setup section above)
  - Wait for user to install and restart
  - If user skips: Run only Checks #6, #7, #9, #10
  - Add warning to final summary
```

### MCP Usage Pattern (When Available)

For checks requiring experiment validation:

**Step 1: Extract Experiment References**
```bash
# Find all experiment IDs
grep -oE 'ExperimentId:\s*"([^"]+)"' <pr_files> | cut -d'"' -f2

# Find all experiment names
grep -oE 'ExperimentName:\s*"([^"]+)"' <pr_files> | cut -d'"' -f2
```

**Step 2: Fetch Experiment Details**
```
For each experiment ID/name:
  Call: mcp__splitz-prod__get_experiment(experiment_key)
  Store: experiment details, variants, state, project info
```

**Step 3: Validate Against Code**
```
Check #1: Verify experiment exists (no error from MCP)
Check #2: Verify experiment.status == "activated" (or "scheduled")
Check #3: Verify variant names in code exist in experiment.variants[]
Check #4: Verify default variant matches control variant from MCP
Check #5: Verify RequestData fields match experiment.audience rules
Check #8: Verify all bulk experiment IDs exist
```

**Example Complete Workflow**:
```
1. Detect code: ExperimentId: "SJYmFIMFxAEqSe"

2. Verify MCP available:
   mcp__splitz-prod__splitz_guide(["experiment_doc"]) → Success ✅

3. Fetch experiment:
   mcp__splitz-prod__get_experiment("SJYmFIMFxAEqSe")

4. Parse response:
   {
     "experiment": {
       "id": "SJYmFIMFxAEqSe",
       "name": "Champion vs Challenger Model",
       "status": "activated",  ← Check #2
       "variants": [
         {"name": "champion"},  ← Check #3
         {"name": "challenger"}
       ]
     }
   }

5. Validate code:
   ✅ Experiment exists (Check #1)
   ✅ Status is "activated" (Check #2)
   ✅ Code references "champion" variant which exists (Check #3)

6. If any check fails, flag with details from MCP response
```

### Graceful Degradation (No MCP)

If MCP not installed and user chooses to skip:

```
Skip: Checks #1, #2, #3, #4, #5, #8 (require live Splitz data)

Run: Checks #6, #7, #9, #10 (code-based only)
  ✅ #6: Error handling patterns
  ✅ #7: Identifier consistency
  ✅ #9: Environment configuration
  ✅ #10: Test coverage

Add to summary:
⚠️  Partial Splitz Validation (6/10 checks skipped)
💡 Install Splitz MCP for full experiment validation
   Run: claude mcp add --scope user --transport http splitz-prod "https://splitz-concierge.razorpay.com/mcp"
   Then restart Claude Code
```

### Logging Recommendations

When flagging issues (MCP available):
```
Example:
❌ Check #2: Environment Activation
   Experiment: SJYmFIMFxAEqSe ("Champion vs Challenger Model")
   Current Status: terminated
   Expected: activated or scheduled

   MCP Details:
   - Status: terminated
   - Auto-terminated: 2026-05-25
   - Project: SmartRouting (payments/router)

   Fix: Re-activate experiment in Splitz before deploying code
   URL: https://splitz.razorpay.com/projects/McZjjMPYuE2cTI/experiments/SJYmFIMFxAEqSe
```

When flagging issues (no MCP):
```
Example:
⚠️  Check #6: Evaluation Error Handling
   File: checkout_handler.go:42
   Issue: GetVariant result not checked for nil before access

   Note: Cannot validate experiment configuration (Splitz MCP not available)
   For full validation, install MCP and verify experiment exists
```
