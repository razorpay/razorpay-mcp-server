---
name: pre-mortem
description: Automated pre-mortem checks for GitHub PRs before merge. Validates infrastructure patterns (PostgreSQL, Aurora, DynamoDB, Kafka config, Kafka consumers/producers, Redis, SQS), query & index patterns (foreign keys, N+1 queries, composite indexes), service integrations (API contracts, events, Splitz experiments, Stork notifications, Passport auth, ASV account service, Router SDK), payment platform services (PG-Router service development), performance optimization (duplicate queries, missing indexes, N+1 patterns - delegated to db-network-optimizer), domain business logic, quality standards (tests, coverage), and observability (monitoring, logging). Includes comprehensive goroutine panic recovery checks. Works for Go/PHP services. 199 automated checks (181 static analysis + 18 performance via delegation, 6 require Splitz MCP) run in parallel across 5 specialized agents (Infrastructure, Service Integration, Domain+Quality, Performance, Observability). Use when PR is created or updated to catch reliability, security, quality, and performance issues before they reach production.
---

# Pre-Mortem Skill

## Purpose

Runs automated checks on GitHub PRs to catch reliability, security, quality, and performance issues before merge. Prevents production incidents by validating:

- **Infrastructure Patterns**: Database transactions (PostgreSQL, Aurora, DynamoDB), Kafka consumers, Redis caching, SQS queues, circuit breakers, error handling, TOML configs
- **Service Contracts**: API request/response schemas, event schemas between services
- **Domain Business Logic**: Constraints validation, flow integrity, business rules (from repo skills)
- **Quality Standards**: Unit test coverage, integration tests (SLITs), feature flags (Splitz)
- **Performance Optimization**: Duplicate database queries, duplicate service calls, N+1 patterns, missing indexes (delegated to db-network-optimizer)

## When to Use

Invoke this skill when:
- User creates or updates a PR
- User asks: "Review this PR", "Check for issues", "Pre-mortem review", "Is this safe to merge?"
- CI passes but you want deeper validation beyond automated tests
- Before approving or merging a PR

**DO NOT** use for:
- Non-PR code reviews (use code-review skill instead)
- Documentation-only changes (unless TOML configs)
- Initial exploratory code discussion

## How It Works

### Step 1: Detect PR Context

```bash
# Get current PR if in a branch
PR_NUMBER=$(gh pr view --json number --jq .number 2>/dev/null)

# Or user provides PR number
PR_NUMBER=<user-provided>

# Get PR details including repo
gh pr view $PR_NUMBER --json number,title,headRefName,files,url,headRepositoryOwner,headRepository
```

If gh CLI not available or not authenticated:
```
⚠️ GitHub CLI required. Please run: gh auth login
```

**Extract repo information:**
```bash
# Get repo name from PR
REPO_NAME=$(gh pr view $PR_NUMBER --json headRepository --jq .headRepository.name)
REPO_OWNER=$(gh pr view $PR_NUMBER --json headRepositoryOwner --jq .headRepositoryOwner.login)
```

### Step 2: Load Repo-Level Context

**CRITICAL:** Always load repo-level context before running any checks.

**Ask user for repo path:**
```
I need the local path to the ${REPO_NAME} repository to access repo-level context.

Please provide the path (e.g., /path/to/repo/terminals):
```

**Once repo path is confirmed, load context in this order:**

1. **Check for repo skill** (highest priority):
   ```bash
   # Check for .claude/skills or .agents/skills
   SKILL_DIR=$(find "${REPO_PATH}" -type d \( -path "*/.claude/skills/*-skill" -o -path "*/.agents/skills/*-skill" \) | head -1)

   if [ -n "$SKILL_DIR" ]; then
       # Load patterns from repo skill
       - modules/infrastructure/* (DB, Kafka, Redis patterns)
       - modules/domain/* (business constraints, flows)
       - modules/monitoring/* (metrics, logging patterns)
       - CLAUDE.md (overview and patterns)
   fi
   ```

2. **Check for CLAUDE.md** (repo documentation):
   ```bash
   if [ -f "${REPO_PATH}/CLAUDE.md" ]; then
       # Extract patterns for:
       - Metrics library and usage
       - Trace code constants location
       - Logger syntax
       - Database patterns
       - Event schemas
       - Testing patterns
   fi
   ```

3. **Check for code examples** (fallback):
   ```bash
   # Find example files for patterns
   - Find existing metrics usage: grep -r "metrics\\.Count\\|metrics\\.Increment" --include="*.go"
   - Find trace codes: find . -name "*tracecode*.go" -o -name "*trace_code*.go"
   - Find logger usage: grep -r "logger\\.Error" --include="*.go" | head -5
   ```

**If no repo-level context found:**
```
⚠️ No repo-level context found (.claude/skills, CLAUDE.md).

I'll use generic patterns, but results may not match your codebase conventions.

Would you like to:
1. Point me to repo documentation
2. Provide example files for patterns
3. Continue with generic checks
```

**Run the extraction script to write `/tmp/premortem/$PR_NUMBER/repo_context.json`:**

```bash
SKILL_SCRIPTS_DIR="$(find ~/.claude/skills ~/.agents/skills -name extract_repo_context.sh 2>/dev/null | head -1 | xargs -r dirname)"
bash "$SKILL_SCRIPTS_DIR/extract_repo_context.sh" "$REPO_PATH" "$PR_NUMBER"
```

The script greps the actual codebase for real examples of metrics calls, logger calls, and tracecode constants — far more reliable than writing JSON manually. Fields it cannot detect are written as `""` and agents fall back to generic patterns for empty fields.

**Do NOT write `repo_context.json` by hand** — JSON escaping of code snippets is error-prone and produces broken files.

### Step 3: Analyze Changed Files

```bash
# Get list of changed files
gh pr diff $PR_NUMBER --name-only > /tmp/pr_files.txt

# Categorize by type for progressive loading
```

**File Pattern → Reference Mapping:**

| Files Changed | Load Reference | Checks |
|---------------|----------------|--------|
| `internal/*/repo.go`, `pkg/db/*` (Generic DB) | `infrastructure-database.md` | 6 DB checks |
| `internal/*/repo.go`, GORM code (PostgreSQL) | `infrastructure-postgres.md` | 6 PostgreSQL checks |
| Database queries, migrations with indexes | `infrastructure-query-index-analysis.md` | 6 Query & index checks |
| Aurora/RDS config, `aurora.*endpoint` | `infrastructure-aurora.md` | 6 Aurora checks |
| `pkg/dynamodb/*`, DynamoDB client code | `infrastructure-dynamodb.md` | 8 DynamoDB checks |
| `internal/kafka/*`, `worker/kafka/*` | `infrastructure-kafka.md` | 8 Kafka consumer/producer checks |
| Kafka client init, config with brokers | `infrastructure-kafka-config.md` | 10 Kafka config checks |
| `internal/cache/*`, `pkg/queue/redis.go` | `infrastructure-redis.md` | 8 Redis checks |
| `pkg/queue/sqs.go`, `worker/dispatcher/*` | `infrastructure-sqs.md` | 6 SQS checks |
| `internal/events/*`, `internal/event_*` | `infrastructure-eventing.md` | 6 event checks |
| `pkg/httpclient/*`, retry/circuit breaker | `infrastructure-resilience.md` | 8 resilience checks |
| Error handling code | `infrastructure-error-handling.md` | 9 error checks |
| `configs/*.toml` | `infrastructure-config.md` | 6 config checks |
| `internal/mozart/*`, service clients | `services-api-contracts.md` | 4 API checks |
| Event publishing/consuming | `services-event-contracts.md` | 4 event checks |
| `splitz.GetVariant`, `splitz.Client`, experiment IDs | `services-splitz.md` | 10 Splitz checks (6 require MCP) |
| `stork.NewClient`, `SendSMS`, `SendEmail`, `SendWhatsApp` | `services-stork.md` | 12 Stork checks |
| `passport.InitHandler`, `FromToken`, `X-Passport-JWT-V1` | `services-passport.md` | 10 Passport checks |
| `accountService.NewClient`, `GetByID`, `Write().Save` | `services-asv.md` | 10 ASV checks |
| `router.NewClient`, `GetTerminals`, `NewFetchTerminalsRequest` | `services-router.md` | 10 Router SDK checks |
| PG-Router: `MutexClient`, callback handlers, payment create | `pgrouter-service.md` | 10 PG-Router service checks |
| Domain entity changes | `domain-constraints.md`, `domain-flows.md` | 10 domain checks |
| `*_test.go` files | `quality-unit-tests.md` | 4 test checks |
| `slit/*` changes | `quality-integration-tests.md` | 2 SLIT checks |
| `internal/*/repo.go`, `**/service.go`, `**/handler.go`, migrations, workers | **Invoke: `/db-network-optimizer`** | 18 performance checks (delegated) |
| **ALL PRs (final step)** | `observability-monitoring-logging.md` | 5 observability checks |

### Step 4: Plan Which References to Load (PLANNING ONLY)

> **⚠️ PLANNING STEP — Do NOT run checks here.** This step produces a plan (which reference files are relevant) that is passed to each parallel agent in Step 5. The agents read and apply the reference files, not you.

**Decide which references to load based on changed files:**

```python
# Pseudo-code — planning only, not execution
changed_files = get_pr_files()
references_to_load = []

for file in changed_files:
    if matches(file, "internal/*/repo.go"):
        # Detect database type
        if uses_gorm(file):
            references_to_load.append("infrastructure-postgres.md")
        elif contains_aurora_endpoint(config):
            references_to_load.append("infrastructure-aurora.md")
        else:
            references_to_load.append("infrastructure-database.md")
    if matches(file, "pkg/dynamodb/*"):
        references_to_load.append("infrastructure-dynamodb.md")
    if matches(file, "configs/*.toml"):
        references_to_load.append("infrastructure-config.md")
    # ... etc

# Remove duplicates
references_to_load = unique(references_to_load)

# → Pass references_to_load to agents in Step 5 via their prompts.
# → Agents use the Read tool to load and apply each reference file.
# → The preprocess_pr.sh manifest (Step 5a) already covers most routing;
#    use this step to supplement with finer-grained file-type detection.
```

**Benefits:**
- Base load: SKILL.md (~5 pages)
- Per file type: 1-3 reference files (~10-15 pages each)
- Domain checks: Load from repo skill if exists (~20 pages)
- **Total context: 15-50 pages per PR** (not all 181 checks at once)

### Step 5: Preprocess PR + Launch Parallel Check Agents

#### 5a: Preprocess (run before spawning agents)

Run the preprocessing script to fetch, filter, and cache the PR diff. This is the key token-saving step — the diff is fetched once, split into category-specific slices, and written to disk. Agents never receive inline diff content.

```bash
SKILL_SCRIPTS_DIR="$(find ~/.claude/skills ~/.agents/skills -name preprocess_pr.sh 2>/dev/null | head -1 | xargs -r dirname)"
CACHE_DIR="/tmp/premortem/$PR_NUMBER"
REFERENCES_DIR="$(dirname "$SKILL_SCRIPTS_DIR")/references"

if [ -z "$SKILL_SCRIPTS_DIR" ]; then
  echo "⚠️  pre-mortem scripts not found. Install the skill first: make install SKILL=pre-mortem"
  echo "   Falling back to inline diff mode (no token optimisation)."
  mkdir -p "$CACHE_DIR"
  gh pr diff "$PR_NUMBER" > "$CACHE_DIR/diff.txt"
  # Set flag so agents know to read diff.txt instead of category slices
  PREPROCESS_OK=false
else
  bash "$SKILL_SCRIPTS_DIR/preprocess_pr.sh" "$PR_NUMBER"
  PREPROCESS_OK=true
fi
# → writes $CACHE_DIR/{manifest.json, *_diff.txt, changed_files.txt}
```

Note: `xargs -r` (GNU) and `xargs` without args (BSD) both skip invocation on empty input. Use `-r` to be safe and note it may not work on macOS — the validation guard `[ -z "$SKILL_SCRIPTS_DIR" ]` is the real safety net.

**If the script fails** (gh auth, network, Python unavailable): fall back to using the full diff inline. Run `gh pr diff $PR_NUMBER > $CACHE_DIR/diff.txt` and point all agents at `diff.txt` instead of the category slices. Token savings are lost but quality is fully preserved.

**Read the manifest:**
```bash
# Parse manifest to decide which agents to spawn
INFRA_HAS_CONTENT=$(python3 -c "import json; d=json.load(open('$CACHE_DIR/manifest.json')); print(str(d['infra']['has_content']).lower())" 2>/dev/null || echo "true")
SERVICES_HAS_CONTENT=$(python3 -c "import json; d=json.load(open('$CACHE_DIR/manifest.json')); print(str(d['services']['has_content']).lower())" 2>/dev/null || echo "true")
# domain_quality and observability agents always spawn
```

**Re-run optimisation:** If SHA is unchanged the script exits immediately (cache hit).

#### 5b: Launch agents in a SINGLE parallel Task tool call

**Do NOT run checks sequentially — launch all applicable agents simultaneously.**

Skip Agent 1 if `INFRA_HAS_CONTENT == "false"` (from manifest parse above).
Skip Agent 2 if `SERVICES_HAS_CONTENT == "false"` (from manifest parse above).
Always spawn Agents 3 and 4.

**CRITICAL — substitute actual values before building each agent prompt:**
Replace every `{CACHE_DIR}`, `{REFERENCES_DIR}` placeholder below with the real paths you set in step 5a before passing the prompt to the Task tool. Do NOT pass a prompt containing literal `{CACHE_DIR}` — the sub-agent cannot resolve shell variables from the parent shell.

**Canonical `check_id` format used by all agents:** `<reference-basename>-<check-number>`
Examples: `infrastructure-database-1`, `infrastructure-kafka-2`, `observability-monitoring-logging-1`, `services-splitz-3`.
This format is deterministic and enables exact deduplication in Step 6.

---

#### Agent 1: Infrastructure Checks
*(skip if manifest shows no infra content)*

Prompt *(substitute {CACHE_DIR} and {REFERENCES_DIR} before launching)*:
```
You are running infrastructure checks for a pre-mortem PR review.
You have all context you need. Do NOT ask the user any questions.
If any context field is empty, use generic Go patterns and continue.

Use the Read tool to read the diff:  {CACHE_DIR}/infra_diff.txt
Use the Read tool for changed files: {CACHE_DIR}/changed_files.txt
Use the Read tool for repo context:  {CACHE_DIR}/repo_context.json
If repo_context.json contains a non-empty "claude_md_path" field, also use the Read tool to read that file — it contains repo-level patterns and conventions that supplement the grep-extracted context.

For each file pattern in the diff, read the matching reference file from {REFERENCES_DIR}/ and apply every check in that file:
- DB/GORM (gorm.DB, .Begin())           → infrastructure-postgres.md (or infrastructure-aurora.md if aurora.*endpoint in diff)
- Generic DB repo patterns              → infrastructure-database.md
- dynamodb. usage                       → infrastructure-dynamodb.md
- kafka. / worker/kafka/ paths          → infrastructure-kafka.md
- Kafka client init, broker config      → infrastructure-kafka-config.md
- redis. / internal/cache/ paths        → infrastructure-redis.md
- sqs. / worker/dispatcher/ paths       → infrastructure-sqs.md
- internal/events/ paths or EventBus    → infrastructure-eventing.md
- http client, circuit breaker          → infrastructure-resilience.md
- error wrapping, sentinel errors       → infrastructure-error-handling.md
- configs/*.toml                        → infrastructure-config.md
- WHERE / JOIN / new columns in queries → infrastructure-query-index-analysis.md

For each check in each reference file:
1. Read the reference file once
2. For every "Flag if" condition, scan the diff
3. Collect violations with exact file path and line number from the diff

Return ONLY a JSON array — no prose, no markdown, no explanation:
[{
  "severity": "Critical|High|Medium|Low",
  "check_id": "<reference-basename>-<check-number>",
  "title": "one-line description",
  "file": "relative/path/to/file.go",
  "line": "42",
  "issue": "what is wrong",
  "fix": "how to fix it",
  "reference": "<reference-basename>.md #<check-number>"
}]

Empty array [] if no violations found.
```

---

#### Agent 2: Service Integration Checks
*(skip if manifest shows no services content)*

Prompt *(substitute {CACHE_DIR} and {REFERENCES_DIR} before launching)*:
```
You are running service integration checks for a pre-mortem PR review.
You have all context you need. Do NOT ask the user any questions.
If any context field is empty, use generic Go patterns and continue.

Use the Read tool to read the diff:  {CACHE_DIR}/services_diff.txt
Use the Read tool for changed files: {CACHE_DIR}/changed_files.txt
Use the Read tool for repo context:  {CACHE_DIR}/repo_context.json
If repo_context.json contains a non-empty "claude_md_path" field, also use the Read tool to read that file — it contains repo-level patterns and conventions that supplement the grep-extracted context.

For each service pattern detected, read the matching reference from {REFERENCES_DIR}/ and apply every check:
- Mozart / HTTP gateway client         → services-api-contracts.md
- Event publish or consume             → services-event-contracts.md
- splitz.GetVariant, experiment IDs    → services-splitz.md
- stork.NewClient, Send{SMS,Email,WA}  → services-stork.md
- passport.InitHandler, FromToken      → services-passport.md
- accountService., GetByID             → services-asv.md
- router.NewClient, GetTerminals       → services-router.md
- MutexClient, PaymentCallback         → pgrouter-service.md

Return ONLY a JSON array — no prose, no markdown, no explanation:
[{
  "severity": "Critical|High|Medium|Low",
  "check_id": "<reference-basename>-<check-number>",
  "title": "one-line description",
  "file": "relative/path/to/file.go",
  "line": "42",
  "issue": "what is wrong",
  "fix": "how to fix it",
  "reference": "<reference-basename>.md #<check-number>"
}]

Empty array [] if no violations found.
```

---

#### Agent 3: Domain + Quality Checks
*(always spawn)*

Prompt *(substitute {CACHE_DIR} and {REFERENCES_DIR} before launching)*:
```
You are running domain and quality checks for a pre-mortem PR review.
You have all context you need. Do NOT ask the user any questions.
If any context field is empty, use generic patterns and continue.

Use the Read tool to read the diff:  {CACHE_DIR}/domain_quality_diff.txt
Use the Read tool for changed files: {CACHE_DIR}/changed_files.txt
Use the Read tool for repo context:  {CACHE_DIR}/repo_context.json
If repo_context.json contains a non-empty "claude_md_path" field, also use the Read tool to read that file — it contains repo-level patterns and conventions that supplement the grep-extracted context.

IMPORTANT: This diff is a superset — it contains ALL non-test Go files from the PR,
not only domain-specific ones. For infrastructure or service files in this diff,
apply domain-level checks (business constraints, flow integrity) only — NOT
low-level infra checks (those are handled by Agents 1 and 2). This prevents
duplicate findings.

**Domain checks** (read {REFERENCES_DIR}/domain-constraints.md and domain-flows.md):
1. If repo_context.repo_skill_dir is non-empty, find domain modules there
2. Extract domain name from changed paths (e.g. internal/gateway_credentials/ → gateway-credentials)
3. Load domain constraints and flows from the repo skill if present
4. Verify: unique constraints enforced, required fields validated, flow steps not skipped

**Quality checks:**
- *_test.go or new .go files without tests → {REFERENCES_DIR}/quality-unit-tests.md
- slit/* changes                           → {REFERENCES_DIR}/quality-integration-tests.md
- Splitz experiment usage                  → {REFERENCES_DIR}/quality-feature-flags.md
- CI config changes                        → {REFERENCES_DIR}/quality-ci-integration.md

Return ONLY a JSON array — no prose, no markdown, no explanation:
[{
  "severity": "Critical|High|Medium|Low",
  "check_id": "<reference-basename>-<check-number>",
  "title": "one-line description",
  "file": "relative/path/to/file.go",
  "line": "42",
  "issue": "what is wrong",
  "fix": "how to fix it",
  "reference": "<reference-basename>.md #<check-number>"
}]

Empty array [] if no violations found.
```

---

#### Agent 4: Observability Checks
*(always spawn — runs for ALL PRs)*

Prompt *(substitute {CACHE_DIR} and {REFERENCES_DIR} before launching)*:
```
You are running observability checks for a pre-mortem PR review.
You have all context you need. Do NOT ask the user any questions.
Use repo_context patterns for actual syntax; fall back to generic Go patterns if fields are empty.

Use the Read tool for the diff:      {CACHE_DIR}/observability_diff.txt
Use the Read tool for repo context:  {CACHE_DIR}/repo_context.json
If repo_context.json contains a non-empty "claude_md_path" field, also use the Read tool to read that file — it contains repo-level patterns and conventions that supplement the grep-extracted context.

Read {REFERENCES_DIR}/observability-monitoring-logging.md fully and apply EVERY check defined in that file.
Use the check number from each "## Check N:" header as the check-number in your check_id.
When the reference file says "Ask user for repo context", skip that step — use the patterns from repo_context.json instead.

Return ONLY a JSON array — no prose, no markdown, no explanation:
[{
  "severity": "Critical|High|Medium|Low",
  "check_id": "<reference-basename>-<check-number>",
  "title": "one-line description",
  "file": "relative/path/to/file.go",
  "line": "42",
  "issue": "what is wrong",
  "fix": "how to fix it",
  "reference": "<reference-basename>.md #<check-number>"
}]

Empty array [] if no violations found.
```

---

#### Agent 5: Performance Checks
*(skip if no performance-relevant files detected)*

**When to Spawn:**
Check if PR modifies performance-relevant files:
```bash
PERF_FILES=$(gh pr diff $PR_NUMBER --name-only | grep -E "(repo\.go|service\.go|handler\.go|migrations/.*\.sql|worker/)")
```

If `$PERF_FILES` is non-empty, spawn this agent.

**Delegation Pattern:**
This agent **delegates to the db-network-optimizer skill** rather than running checks directly.

Prompt:
```
You are running performance checks for a pre-mortem PR review by delegating to the db-network-optimizer skill.

Check if the db-network-optimizer skill is installed:
- Look for ~/.claude/skills/db-network-optimizer/
- Or ~/.agents/skills/db-network-optimizer/

If NOT installed:
Return JSON with a single informational finding:
[{
  "severity": "Low",
  "check_id": "performance-delegation-info",
  "title": "Performance checks available via db-network-optimizer skill",
  "file": "",
  "line": "",
  "issue": "Install db-network-optimizer skill for 18 performance checks (duplicate queries, missing indexes, N+1 patterns)",
  "fix": "Run: npx skills add razorpay/agent-skills --skill db-network-optimizer -y -g",
  "reference": "performance-delegation.md"
}]

If installed:
Invoke the db-network-optimizer skill with: "Analyze PR #$PR_NUMBER for performance issues"

The db-network-optimizer will:
- Detect 8 duplicate patterns (duplicate DB queries, service calls, N+1 loops)
- Check 10 query optimization issues (missing indexes Levels 1-3)
- Return findings in its detailed two-section report

Extract findings from db-network-optimizer's output and convert to pre-mortem JSON format:
[{
  "severity": "Critical|High|Medium|Low",  # map from db-network-optimizer severity
  "check_id": "performance-<pattern-name>",
  "title": "one-line description",
  "file": "relative/path/to/file.go",
  "line": "42",
  "issue": "what is wrong",
  "fix": "how to fix it (from db-network-optimizer recommendation)",
  "reference": "db-network-optimizer"
}]

Empty array [] if no violations found.
```

---

**Wait for all agents to complete before proceeding.**

### Step 6: Aggregate + Deduplicate Results

Agent 3 is a catch-all that receives all Go files, so it may report the same violation as
Agent 1 or 2 (e.g., both flag a missing rollback on the same `repo.go` line). Deduplicate
before building the final report.

```python
all_raw = (
    agent1_infrastructure_findings +
    agent2_service_findings +
    agent3_domain_quality_findings +
    agent4_observability_findings +
    agent5_performance_findings  # from db-network-optimizer delegation
)

# Deduplicate: same (file, line, check_id) = same finding.
# When duplicated, keep the higher severity instance.
# Use .get() with default=3 (Low) to avoid KeyError if an agent returns
# an unrecognised severity string.
severity_order = {"Critical": 0, "High": 1, "Medium": 2, "Low": 3}

def sev(finding):
    raw = finding.get("severity", "Low")
    normalized = raw.strip().capitalize()  # "critical" → "Critical", "HIGH" → "High"
    return severity_order.get(normalized, 3)

seen = {}
for f in all_raw:
    key = (f.get("file", ""), f.get("line", ""), f.get("check_id", ""))
    if key not in seen or sev(f) < sev(seen[key]):
        seen[key] = f

all_findings = sorted(seen.values(), key=sev)

# Count
counts = {
    "Critical": len([f for f in all_findings if f.get("severity","").capitalize() == "Critical"]),
    "High":     len([f for f in all_findings if f.get("severity","").capitalize() == "High"]),
    "Medium":   len([f for f in all_findings if f.get("severity","").capitalize() == "Medium"]),
    "Low":      len([f for f in all_findings if f.get("severity","").capitalize() == "Low"]),
}
total_checks = 199  # total defined checks (181 static + 18 performance via delegation)
passed = total_checks - len(all_findings)
```

### Step 7: Generate Report

**Output Format (Crisp & Concise):**

```
🔍 Pre-Mortem: PR #123 "Add gateway credentials feature"

📊 Summary: 🚨 2 Critical | ⚠️ 5 High | 📋 8 Medium | ✅ 181/199 Passed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚨 CRITICAL (Must Fix):

1. Missing Transaction Rollback
   Location: internal/terminals/repo.go:422
   Issue: Save() error doesn't rollback → data corruption risk

   Fix: Add `defer tx.Rollback()` before operations

   Reference: infrastructure-database.md #1

2. Missing prod-live.toml Keys
   Location: configs/default.toml:45
   Issue: New key 'newFeature.enabled = true' not in prod → uses dev default!

   Fix: Add key to prod-live.toml, prod-test.toml, stage-live.toml, stage-test.toml

   Reference: infrastructure-config.md #1

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️ HIGH (Should Fix):

3. No Panic Recovery in Kafka Consumer
   Location: worker/kafka/handler/new_handler.go:25
   Issue: Goroutine without defer recover() → worker crashes

   Fix: Wrap with `defer func() { if r := recover(); r != nil { ... } }()`

   Reference: infrastructure-kafka.md #1

4. No Error Metric Emitted
   Location: internal/services/payment.go:156
   Issue: logger.Error() but no metrics.Count() → can't alert on failures

   Fix: Add `metrics.Count(ctx, "payment.failed", 1, tags)`

   Reference: observability-monitoring-logging.md #1

5. Missing Trace Code
   Location: internal/services/payment.go:156
   Issue: logger.Error(ctx, "payment failed", ...) → string literal, not constant

   Fix: Use `logger.Error(ctx, TraceCode.PAYMENT_PROCESSING_FAILED, ...)`

   Reference: observability-monitoring-logging.md #2

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 MEDIUM (8 issues) | ℹ️ LOW (3 issues)
  → Ask "show medium issues" for details

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ PASSED: 64 checks (Database, Kafka, Redis, Error Handling, Domain, Quality)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔗 PR Link: https://github.com/razorpay/terminals/pull/123
```

### Step 8: Offer Actions

After generating the report, ask the user what they'd like to do:

```
What would you like me to do?

1. 🔧 Fix issues automatically (where possible)
2. 💬 Add review comments to PR (via gh pr review)
3. 🔍 Investigate CI failures (invoke pr-ci-fixer)
4. 📄 Show detailed report for specific issue
5. ✅ Mark as reviewed (I'll fix manually)
```

**Action Handlers:**

#### Option 1: Auto-fix issues
```bash
# For fixable issues (e.g., defer tx.Rollback, panic recovery)
Apply fixes to code
Run go fmt
Show diff
Ask for confirmation
If yes → Use pr-creator to commit & push
```

#### Option 2: Add PR comments
```bash
# For each critical/high issue
gh pr review $PR_NUMBER --comment --body "
**[Critical] Missing Transaction Rollback**
File: internal/terminals/repo.go:422

Issue: Save() error doesn't trigger rollback
Reference: infrastructure-database.md (Check #1)

Fix: Add \`defer tx.Rollback()\` after \`db.Begin()\`
"
```

#### Option 3: Invoke pr-ci-fixer
```bash
# If CI is failing
Check CI status: gh pr checks $PR_NUMBER
If failed → Invoke pr-ci-fixer skill
After fix → Re-run pre-mortem
```

#### Option 4: Detailed report
```
Show expanded issue with:
- Full code context
- Related constraints from domain skill
- Similar patterns in codebase
- Step-by-step fix instructions
```

### Step 9: CI Integration

**Check CI status and offer to fix:**

```bash
# gh pr checks exits non-zero if any required check fails; use that to detect failure
if ! gh pr checks "$PR_NUMBER" --required 2>/dev/null; then
    CI_FAILING=true
else
    CI_FAILING=false
fi

if [ "$CI_FAILING" = "true" ]; then
    echo "⚠️  CI is failing. Should I investigate?"

    if user_confirms; then
        # Invoke pr-ci-fixer skill
        invoke_skill("pr-ci-fixer", pr_number=$PR_NUMBER)

        # After fix, re-run pre-mortem
        echo "Re-running pre-mortem after CI fix..."
        rerun_checks()
    fi
fi
```

**Integration workflow:**
1. Load repo-level context (Step 2) → Understand codebase patterns
2. Pre-mortem runs → Finds static issues using repo patterns
3. Detects CI failure
4. Asks to investigate
5. Invokes pr-ci-fixer → Fixes failing tests
6. Re-runs pre-mortem → Validates fixes didn't introduce new issues

## Progressive Disclosure Strategy

**Context Management:**

- **Repo Context** (Step 2): Load first — repo skills, CLAUDE.md, code patterns (~10-20 pages)
- **Base**: SKILL.md (5 pages) - always loaded
- **Infrastructure**: 1-3 reference files (~30 pages) - based on changed files
- **Services**: 0-2 reference files (~15 pages) - if service integration modified
- **Domain**: 1-2 domain modules (~20 pages) - from repo skill if exists
- **Quality**: 1-2 reference files (~15 pages) - if tests modified

**Total: 25-60 pages per PR** (repo context + dynamically loaded checks)

**Example loading scenarios:**

| PR Changes | References Loaded | Total Pages |
|------------|-------------------|-------------|
| TOML config only | config.md | ~15 pages |
| PostgreSQL repo changes | postgres.md, error-handling.md | ~30 pages |
| Aurora CDC pipeline | aurora.md, eventing.md | ~30 pages |
| DynamoDB operations | dynamodb.md | ~25 pages |
| New Kafka handler | kafka.md, eventing.md, error-handling.md | ~35 pages |
| Service integration | api-contracts.md, resilience.md | ~25 pages |
| Domain entity | constraints.md, flows.md + repo skill | ~40 pages |
| Full stack feature | 8-10 references + domain | ~60 pages |

## Severity Levels

| Level | Icon | When to Flag | Action Required |
|-------|------|--------------|-----------------|
| Critical | 🚨 | Production incident risk, data corruption, security vulnerability | Block merge until fixed |
| High | ⚠️ | Data loss risk, performance degradation, contract breakage | Strongly recommend fix |
| Medium | 📋 | Maintainability issue, technical debt, best practice violation | Suggest fix |
| Low | ℹ️ | Code style, minor optimization opportunity | Optional improvement |

**Severity Assignment:**

```
Critical:
- Missing transaction rollback
- No panic recovery in critical paths
- Unvalidated unique constraints
- Prod config missing keys (with dangerous defaults)
- Breaking API/event contracts

High:
- No timeout on queries
- Missing idempotency checks
- Weakened validation rules
- Flow steps skipped
- No DLQ for queue

Medium:
- Missing indexes on new columns
- No feature flag for major changes
- SLIT missing for new endpoints
- Coverage drop < 10%

Low:
- Code style issues
- Minor N+1 queries
- Missing comments
```

## Integration with Other Skills

### pr-creator Integration

When fixes are made, use pr-creator workflow:

```
1. Apply fixes to code
2. Run formatter: go fmt ./...
3. Stage files: git add <files>
4. Commit: git commit -m "Fix pre-mortem issues: ..."
5. Push: git push origin $BRANCH
6. Verify: gh pr checks $PR_NUMBER
```

### pr-ci-fixer Integration

When CI fails, delegate to pr-ci-fixer:

```
1. Detect CI failure
2. Ask user: "Should I investigate CI failures?"
3. Invoke pr-ci-fixer:
   - Parse CI logs
   - Run tests locally
   - Determine flaky vs genuine
   - Fix or retrigger
4. After fix: Re-run pre-mortem
```

### Repo Skill Integration

Load domain knowledge from repo skills:

```
Priority order:
1. .claude/skills/*-skill/modules/domain/
2. .agents/skills/*-skill/modules/domain/
3. Fall back to generic checks if no repo skill
```

## Example Usage

**User commands that trigger this skill:**

```
"Review this PR"
"Check PR #456 for issues"
"Run pre-mortem on my changes"
"Is this PR safe to merge?"
"Pre-mortem check"
"Validate PR #123"
```

**Workflow examples:**

```
User: "Review PR #456"