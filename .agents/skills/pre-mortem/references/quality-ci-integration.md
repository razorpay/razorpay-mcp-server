# CI Integration Quality Check

## Overview

Validates CI check status and **automatically invokes pr-autopilot skill** to fix CI failures, handle flaky tests, or address AI review comments.

**Load when:** Every PR (as final validation step)

**Total Checks:** 1

**Severity Distribution:**
- 🚨 Critical: 1

**Integration:** Uses **pr-autopilot** skill to autonomously fix CI failures

---

## Check 1: CI Checks Passing 🚨 CRITICAL

### What to Check

All required CI checks must pass before merge:
- ✅ Build successful
- ✅ Unit tests passing
- ✅ Integration tests passing
- ✅ Coverage threshold met
- ✅ Linting passing
- ✅ AI review comments resolved (rCoRe check)

### Detection Strategy

```bash
# Check CI status using GitHub CLI
gh pr checks $PR_NUMBER --json name,status,conclusion

# Example output:
# {
#   "name": "build",
#   "status": "completed",
#   "conclusion": "success"
# },
# {
#   "name": "unit-tests",
#   "status": "completed",
#   "conclusion": "failure"  ← FAILING
# },
# {
#   "name": "quality-gate-utExpected",
#   "status": "pending",
#   "conclusion": null  ← PENDING (coverage check incomplete)
# },
# {
#   "name": "rCoRe / comment_resolution_validator",
#   "status": "completed",
#   "conclusion": "failure"  ← AI comments not resolved
# }
```

### Automated Fix with pr-autopilot

When CI checks fail, **automatically offer to invoke pr-autopilot**:

```
🚨 CI Checks Failing

PR #456: Add payment routing feature

Failed checks:
  ❌ unit-tests (3 failures)
  ❌ rCoRe / comment_resolution_validator (2 AI comments)
  ⏸️  quality-gate-utExpected (waiting for coverage)

Would you like me to investigate and fix these issues?
1. ✅ Yes, auto-fix CI failures (uses pr-autopilot)
2. Show me the failing tests
3. Skip for now
```

**If user confirms:**

```bash
# Invoke pr-autopilot skill
invoke_skill("pr-autopilot", {
    "pr_number": 456,
    "action": "fix_ci_failures"
})
```

**pr-autopilot will:**
1. **Check CI status** (gh pr checks)
2. **Identify failed workflows** (gh run list --status failure)
3. **Download CI logs** (gh run view $RUN_ID --log)
4. **Parse failed tests**
5. **Run tests locally** to detect flaky vs genuine failures
6. **Handle based on failure type:**
   - **Flaky tests**: Retrigger CI (gh run rerun)
   - **Genuine failures**: Fix code and push
   - **Coverage failures**: Generate tests
   - **AI comments**: Address comments in-thread
7. **Verify CI passes** after fixes

### Failure Type Handling

#### Type 1: Flaky Tests 🔄

**Detection:**
- Test fails in CI
- Test passes locally
- Test has timing dependencies

**pr-autopilot action:**
```bash
# Retrigger failed jobs
gh run rerun $RUN_ID --failed

# Monitor
gh run watch $RUN_ID
```

**Output:**
```
🔄 Flaky Tests Detected

Flaky tests (pass locally, fail in CI):
  - TestPaymentProcessing_Concurrent (race condition)
  - TestGatewayTimeout_Retry (timing issue)

Actions taken:
  ✓ Retriggered failed CI jobs
  ✓ Tests passed on retry ✅
  ✓ CI now passing
```

#### Type 2: Genuine Test Failures 🔧

**Detection:**
- Test fails both in CI and locally
- Consistent failure

**pr-autopilot action:**
1. Analyze error from CI logs
2. Run test locally to reproduce
3. Identify root cause
4. Fix the code
5. Run tests to verify
6. Commit and push fix

**Output:**
```
🔧 Test Failures Fixed

Failed tests analyzed:
  ❌ TestCreateTerminal_Validation
     Error: Expected validation error, got nil
     Root cause: Missing merchant_id validation

Fixed:
  ✓ Added validation check in terminal_service.go:45
  ✓ Test now passes ✅
  ✓ All tests passing locally
  ✓ Committed: "Fix terminal validation logic"
  ✓ Pushed to PR branch
  ✓ CI retriggered
```

#### Type 3: Coverage Failures 📊

**Detection:**
- `quality-gate-utExpected` check failing or pending
- Coverage below threshold

**pr-autopilot action:**
```bash
# Generate tests to improve coverage
invoke_skill("pr-autopilot", {
    "action": "improve_coverage",
    "target": 80
})
```

**Output:**
```
📊 Coverage Improved

Before: 67.3%
Target: 80%

Generated tests:
  ✓ payment_service_test.go (12 new test cases)
  ✓ terminal_validator_test.go (8 new test cases)

After: 85.2% ✅

Actions:
  ✓ Tests committed and pushed
  ✓ Coverage check now passing
```

#### Type 4: AI Review Comments (rCoRe) 💬

**Detection:**
- `rCoRe / comment_resolution_validator` failing
- Unresolved AI PR review comments

**pr-autopilot action:**
1. Fetch AI comments (gh pr view --json comments)
2. Analyze each comment
3. Determine action (implement, respond, or reject)
4. Address comments in-thread (using in_reply_to)
5. Mark as resolved if implemented

**Output:**
```
💬 AI Review Comments Addressed

Unresolved comments: 2

Comment 1 (Line 45):
  AI: "Add null check for payment.Details"
  Action: Implemented ✅
  Code: Added if payment.Details == nil check
  Reply: "Added null check as suggested"

Comment 2 (Line 89):
  AI: "Consider caching this query"
  Action: Implemented ✅
  Code: Added Redis cache with 15min TTL
  Reply: "Implemented caching with 15min TTL"

Result:
  ✓ All comments addressed
  ✓ rCoRe check now passing
```

### Workflow Integration

```
1. pr-pre-mortem runs all 82 checks
   ↓
2. Detects CI failures at end
   ↓
3. Offers to invoke pr-autopilot
   ↓
4. User confirms
   ↓
5. pr-autopilot:
   - Downloads CI logs
   - Identifies failure type
   - Fixes issues automatically
   - Pushes fixes
   - Retrigs CI
   ↓
6. pr-pre-mortem re-runs to verify all checks pass
   ↓
7. Reports final status to user
```

### Auto-invocation Scenarios

Pre-mortem **automatically offers** pr-autopilot when:
- ✅ Any CI check failing
- ✅ `quality-gate-utExpected` pending (coverage issue)
- ✅ `rCoRe` failing (AI comments unresolved)
- ✅ Flaky tests detected (pass locally, fail in CI)

### Flag Conditions

Flag if:
- Any required CI check status != "success"
- Coverage check pending or failing
- AI comment resolution failing
- Build failing
- Tests failing

### Severity

🚨 **Critical** - PR cannot merge with failing CI

### Output Format

```
🚨 CI Status Check

PR #456: Add payment routing feature
Branch: feature/payment-routing

CI Checks (4/7 passing):
  ✅ build
  ✅ lint
  ✅ integration-tests
  ✅ security-scan

  ❌ unit-tests (3 failures)
     - TestPaymentRouting_InvalidGateway
     - TestPaymentRouting_Timeout
     - TestFeeCalculation_EdgeCase

  ⏸️  quality-gate-utExpected (coverage: 67.3%, target: 80%)

  ❌ rCoRe / comment_resolution_validator (2 unresolved comments)
     - Line 45: Add null check
     - Line 89: Consider caching

🤖 Auto-Fix Available

I can use pr-autopilot to:
  ✓ Run failing tests locally
  ✓ Identify flaky vs genuine failures
  ✓ Fix code issues
  ✓ Generate tests for coverage
  ✓ Address AI review comments
  ✓ Retrigger CI

This will take ~5-10 minutes.

Would you like me to proceed? (yes/no)
```

**After user confirms:**

```
✅ Invoking pr-autopilot...

Step 1: Analyzing CI failures...
  ✓ Downloaded CI logs
  ✓ Identified 3 test failures

Step 2: Running tests locally...
  ✓ TestPaymentRouting_InvalidGateway: PASS (flaky) 🔄
  ✓ TestPaymentRouting_Timeout: PASS (flaky) 🔄
  ✓ TestFeeCalculation_EdgeCase: FAIL (genuine) 🔧

Step 3: Fixing genuine failures...
  ✓ Root cause: Missing edge case for 0 amount
  ✓ Fixed fee calculation logic
  ✓ Test now passing

Step 4: Improving coverage...
  ✓ Generated 8 new test cases
  ✓ Coverage: 67.3% → 86.1%

Step 5: Addressing AI comments...
  ✓ Comment 1: Implemented null check
  ✓ Comment 2: Implemented caching
  ✓ Replied in-thread

Step 6: Pushing fixes...
  ✓ Committed: "Fix test failures and improve coverage"
  ✓ Pushed to feature/payment-routing

Step 7: Retriggering CI...
  ✓ Retriggered failed jobs (flaky tests)
  ✓ New CI run started

Waiting for CI...
  ⏳ Running... (2 min)

✅ All CI checks passing!

Summary:
  - Fixed 1 genuine test failure
  - Retriggered 2 flaky tests (passed)
  - Improved coverage 67.3% → 86.1%
  - Addressed 2 AI review comments
  - All checks green ✅

PR is ready to merge!
```

---

## Manual Investigation Option

If user chooses "Show me the failing tests":

```
📋 Failed Tests Details

1. TestPaymentRouting_InvalidGateway
   File: internal/services/payment_router_test.go:45
   Error: assertion failed: expected error, got nil
   Log: https://github.com/.../runs/12345

2. TestPaymentRouting_Timeout
   File: internal/services/payment_router_test.go:89
   Error: context deadline exceeded
   Log: https://github.com/.../runs/12345

3. TestFeeCalculation_EdgeCase
   File: internal/services/fee_calculator_test.go:123
   Error: panic: division by zero
   Stack trace: ...

Commands to debug locally:
  go test -v ./internal/services -run TestPaymentRouting_InvalidGateway
  go test -race -v ./internal/services -run TestPaymentRouting_Timeout
  go test -v ./internal/services -run TestFeeCalculation_EdgeCase

Would you like me to:
1. Fix these automatically (pr-autopilot)
2. Run them locally for you
3. Skip
```

---

## Summary

| Check # | Pattern | Severity | Auto-fix |
|---------|---------|----------|----------|
| 1 | CI checks passing | 🚨 Critical | pr-autopilot |

---

## How to Apply

**For every PR (final step):**

1. Check CI status (gh pr checks)
2. If any failures, offer pr-autopilot
3. Handle based on failure type:
   - Flaky → retrigger
   - Genuine → fix and push
   - Coverage → generate tests
   - AI comments → address in-thread
4. Verify all checks pass
5. Report status

**Example output:**

```
🎯 Final Validation: CI Status

✅ All CI checks passing
✅ Coverage: 87.5% (target: 80%)
✅ AI comments: All resolved
✅ Build: Successful
✅ Tests: 142/142 passing

PR #456 is ready to merge! 🚀
```

---

## CI Check Priority

**Critical (Block Merge):**
- 🚨 build
- 🚨 unit-tests
- 🚨 quality-gate-utExpected (coverage)
- 🚨 rCoRe (AI comment resolution)

**Important (Should Pass):**
- ⚠️ integration-tests (SLITs)
- ⚠️ lint
- ⚠️ security-scan

**Optional (Nice to Have):**
- ℹ️ performance-tests
- ℹ️ e2e-tests
