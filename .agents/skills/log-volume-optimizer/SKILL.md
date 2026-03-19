---
name: log-volume-optimizer
description: Analyzes and optimizes log volume in Go services by scanning repositories for log statements, estimating daily volume, and generating optimization recommendations. Supports batch processing across multiple repositories with automated PR creation. Use when reducing Coralogix units consumption, optimizing logging costs across services, or auditing log hygiene for Go repositories.
---

# Log Volume Optimizer

Analyzes and optimizes log volume in Razorpay services by combining:
1. **Coralogix MCP** - Real production log frequency data
2. **Static Code Analysis** - Multi-language log detection (Go/PHP/TypeScript/Python)
3. **Smart Decision Engine** - Team-approved optimization rules
4. **Automated PR Creation** - GitHub API-based changes

## Activation

This skill activates when the user asks to:
- Analyze log volume for a service
- Optimize logging costs
- Reduce Coralogix units consumption
- Scan a repository for log statements
- Audit log hygiene for a repository

---

## WORKFLOW (Follow This Exactly)

### Step 1: Get Real Log Frequencies from Coralogix

**REQUIRED: Always start by querying Coralogix MCP for actual production data.**

```
Tool: mcp_coralogix-server_get_logs

Query (DataPrime):
source logs
| filter $l.applicationname == '{APPLICATION_NAME}'
| countby $d.message
| sortby _count desc
| limit 500

Time range: last 15 minutes (start_date, end_date in ISO 8601)
```

**After getting results, save them for the scripts:**

Save the MCP response to a file:
```
reports/{APPLICATION_NAME}-coralogix-frequencies.json
```

Format:
```json
{
  "application": "{APPLICATION_NAME}",
  "query_minutes": 15,
  "frequencies": [
    {"message": "actual log message from coralogix", "count": 12345},
    ...
  ]
}
```

This tells you which log messages have the highest frequency in production.

### Step 2: Scan Repository for Log Statements

Run the multi-language scanner:

```bash
cd ~/.claude/skills/log-volume-optimizer/scripts
python scan_logs.py --path /path/to/repo --language auto --output logs.json
```

This detects:
- All log statements with file, line, function, level, message
- Context flags: `in_loop`, `in_error_handler`
- Message templates for matching

### Step 3: Match Code Logs to Coralogix Data

Use the coralogix_integration module to cross-reference:

```bash
python coralogix_integration.py match \
  --coralogix reports/{APP}-coralogix-frequencies.json \
  --scan-results logs.json \
  --output matched.json
```

Or programmatically:
```python
from coralogix_integration import CoralogixData
cx = CoralogixData()
cx.load_from_json("reports/{APP}-coralogix-frequencies.json")
results, summary = cx.match_to_code_logs(scanned_logs)
```

The matching uses:
1. Exact template match (normalized message → same template)
2. Substring match (code template is prefix/subset of production message)
3. Fuzzy match (>75% similarity via SequenceMatcher)

### Step 4: Apply Decision Rules

For EACH log statement, apply these rules IN ORDER (first match wins):

| Priority | Condition | Action | Reason |
|----------|-----------|--------|--------|
| 1 | Level is ERROR or FATAL | **KEEP** | Always preserve error logs |
| 2 | Level is WARN | **KEEP** | Warnings indicate issues |
| 3 | In error handler (`if err != nil`) | **KEEP** | Error context is critical |
| 4 | Message contains: failed, error, timeout, panic | **KEEP** | Error semantics |
| 5 | Business-critical: payment, transaction, settlement, refund | **KEEP** | Business events |
| 6 | In loop AND frequency > 100K/day | **AGGREGATE** | Batch to summary |
| 7 | In loop AND frequency > 10K/day | **SAMPLE** | Add 1% sampling |
| 8 | Entry/exit pattern: entering, exiting, starting, completed, begin, end | **DOWNGRADE** | Not production-relevant |
| 9 | Success pattern: successfully, done processing | **DOWNGRADE** | Absence indicates failure |
| 10 | INFO level AND frequency > 1M/day | **DOWNGRADE** | Too verbose |
| 11 | Metrics pattern: count, total, processed, duration, latency | **USE_METRIC** | Should be Prometheus |
| 12 | Contains secrets: password, token, key, secret, credential | **FLAG_REVIEW** | Security concern |
| 13 | DEBUG level | **KEEP** | Already lowest level |
| 14 | Default | **KEEP** | Conservative approach |

### Step 5: Generate Changes and Self-Validate

For logs requiring action:
1. Apply transformations (DOWNGRADE changes `.Info(` to `.Debug(`)
2. **VALIDATE each change before committing:**
   - Verify log level was actually changed (Info → Debug)
   - Ensure no random sampling wrappers were added
   - Check brackets are balanced (syntax valid)
   - Reject files with validation errors

### Step 6: Create PR (After Validation Passes)

1. Create branch via GitHub API
2. Commit only validated changes
3. Create PR with impact summary
4. Include validation status in PR description

---

## Decision Actions Reference

| Action | What It Does | When to Use |
|--------|--------------|-------------|
| **KEEP** | No change | ERROR/WARN, error handlers, business events |
| **DOWNGRADE** | INFO → DEBUG | Entry/exit, success confirmations, verbose tracing |
| **REMOVE** | Comment out the log | Duplicates, per-item logs in loops |
| **AGGREGATE** | Move out of loop, log summary | Loop logs with >100K/day |
| **SAMPLE** | INFO → DEBUG (loop logs) | Loop logs that need to be reduced |
| **USE_METRIC** | Replace with Prometheus counter/histogram | Counting, timing, rate patterns |
| **FLAG_REVIEW** | Mark for human review | PII/secrets, large payloads |

---

## MCP Integration

### Coralogix MCP (Required)

**Tool:** `mcp_coralogix-server_get_logs`

**Simple Frequency Query:**
```dataprime
source logs
| filter $l.applicationname == '{APP_NAME}'
| countby $d.message
| sortby _count desc
| limit 500
```

**Full Cost Analysis Query:**
```dataprime
source logs
| filter $l.applicationname == '{APP_NAME}'
| create size_bytes from $d:string.length()+$l:string.length()+$m:string.length()
| create msg_string from substr(trim($d.message), 0, 500)
| groupby msg_string, $m.priorityclass, $m.timestamp/5m as per_5_min
    sum(size_bytes)/(1024*1024*1024) as bytes_gb_per_5_min
| groupby msg_string, priorityclass
    agg avg(bytes_gb_per_5_min) as avg_size
| create estimated_cost_month from case {
    priorityclass == 'low' -> round(avg_size * 12 * 12 * 30 * 0.59 * 0.12),
    priorityclass == 'medium' -> round(avg_size * 12 * 12 * 30 * 0.59 * 0.32),
    priorityclass == 'high' -> round(avg_size * 12 * 12 * 30 * 0.59 * 0.75)
}
| sortby estimated_cost_month desc
```

**Get Units Consumption:**
```
Tool: mcp_coralogix-server_metrics__range_query
Query: sum(cx_data_usage_units{application_name="{APP_NAME}"}) by (application_name)
```

---

## Language Support

| Language | Extensions | Detected Patterns |
|----------|------------|-------------------|
| Go | .go | `logger.Info()`, `log.Error()`, `lgr.Debugw()`, `zap.S().Info()` |
| PHP | .php | `Log::info()`, `$trace->info()`, `$this->trace->debug()` |
| TypeScript | .ts, .tsx, .js | `console.log()`, `logger.info()`, `Logger.debug()` |
| Python | .py | `logging.info()`, `logger.debug()`, `log.error()` |

Auto-detection uses `--language auto` flag.

---

## Directory Exclusions

Always skip these directories (per Rajeev's approach):
- `vendor/`
- `build/`
- `config/`
- `bin/`
- `scripts/`
- `templates/`
- `node_modules/`
- `.git/`

---

## Script Usage

### Single Repository

```bash
cd ~/.claude/skills/log-volume-optimizer/scripts

# Scan
python scan_logs.py --path /path/to/repo --language auto --output analysis.json

# Estimate volume (with RPS)
python estimate_volume.py --input analysis.json --rps 500 --output report.json
```

### Batch Processing

```bash
# Analyze multiple repos
python batch_analyze.py \
    --org razorpay \
    --repos "pg-router,payments-upi,ledger" \
    --output-dir ./reports

# Create PRs via GitHub API (WITHOUT Coralogix - uses estimates only)
python github_api_prs.py \
    --report ./reports/target-apps-analysis.json \
    --org razorpay \
    --branch log-volume-optimizer/optimize-logs

# Create PRs WITH real Coralogix data (RECOMMENDED)
python github_api_prs.py \
    --report ./reports/target-apps-analysis.json \
    --org razorpay \
    --branch log-volume-optimizer/optimize-logs \
    --coralogix-dir ./reports

# Create PRs with single Coralogix data file
python github_api_prs.py \
    --report ./reports/target-apps-analysis.json \
    --org razorpay \
    --coralogix-data reports/pg-router-coralogix-frequencies.json

# Dry-run to preview decisions with Coralogix data
python github_api_prs.py \
    --report ./reports/target-apps-analysis.json \
    --repos "pg-router" \
    --coralogix-dir ./reports \
    --dry-run
```

### Coralogix Data Preparation

```bash
# Generate the query to run via MCP
python coralogix_integration.py query --app pg-router

# After saving MCP results, analyze them
python coralogix_integration.py analyze \
    --input reports/pg-router-coralogix-frequencies.json \
    --top 20

# Match Coralogix data to code scan
python coralogix_integration.py match \
    --coralogix reports/pg-router-coralogix-frequencies.json \
    --scan-results logs.json \
    --output matched.json
```

---

## Target Applications

Priority services (sorted by Coralogix consumption):

```
api, terminals, payments-card, pg-router, payments-upi,
checkout-affordability-api, dashboard, ledger, router,
banking-accounts, subscriptions, optimizer-core, bin-service,
vault, 1cc-rto-prediction-service, settlements, payment-links,
tokens, invoices, checkout-service, shield, payouts,
capital-collections, stork, offers-engine, ui-config-service,
reminders, fts, payment-methods, scrooge
```

---

## Example Session

**User:** "Optimize logs for pg-router"

**Claude Code:**
```
1. Query Coralogix for pg-router log frequencies
   [Call mcp_coralogix-server_get_logs with DataPrime query]
   [Save results to reports/pg-router-coralogix-frequencies.json]

2. Found 1,247 unique log messages. Top 10 by frequency:
   - "request received" - 2.3M/day (Coralogix actual)
   - "processing payment" - 1.8M/day (Coralogix actual)
   - "payment successful" - 1.5M/day (Coralogix actual)
   ...

3. Scan repository code
   [Run scan_logs.py on pg-router]
   Found 156 log statements

4. Match code logs to Coralogix data:
   [Run coralogix_integration.py match]
   Matched 89/156 code logs (57% match rate)
   - exact: 42, substring: 31, fuzzy: 16, unmatched: 67

5. Apply decision rules (with real Coralogix frequencies):
   - Line 45: logger.Info("request received") - [Coralogix: 2.3M/day] → DOWNGRADE (Rule 10)
   - Line 67: logger.Error("payment failed") - [Coralogix: 1.2K/day] → KEEP (Rule 1)
   - Line 89: logger.Info("item processed") in loop - [Coralogix: 500K/day] → AGGREGATE (Rule 6)

6. Create PR with changes (via github_api_prs.py --coralogix-dir reports/):
   - 12 logs downgraded to DEBUG
   - 3 loops aggregated
   - Estimated savings: 4.2M units/day
   - Frequency data source: 89 from Coralogix, 67 from estimates
```

---

## Files in This Skill

| File | Purpose |
|------|---------|
| `SKILL.md` | This file - workflow and rules |
| `scripts/scan_logs.py` | Multi-language log scanner |
| `scripts/estimate_volume.py` | Volume estimation |
| `scripts/decision_engine.py` | Smart decision rules engine |
| `scripts/coralogix_integration.py` | **Coralogix MCP data loading, matching, and frequency lookup** |
| `scripts/github_api_prs.py` | GitHub API PR creation (uses Coralogix data) |
| `scripts/batch_analyze.py` | Multi-repo batch processing |
| `references/team-approaches.md` | Team approaches (Monark, Ankur, Rajeev) |
| `references/logging-standards.md` | Razorpay logging standards |
| `references/optimization-strategies.md` | Optimization strategies |
| `references/mcp-integration.md` | MCP integration guide |

---

## PR Guidelines

1. **Never auto-merge** - Human reviews all changes
2. **Safe operations only** - DOWNGRADE is safest, REMOVE requires confidence
3. **Include impact summary** - Estimated savings in PR description
4. **One PR per repo** - Consolidate all changes
5. **Branch naming** - `log-volume-optimizer/optimize-logs`

---

## Dependencies

1. **Coralogix MCP** - Configured in `~/.claude/mcp_settings.json`
2. **GitHub CLI (`gh`)** - Authenticated with org access
3. **Python 3.9+** - For analysis scripts

---

## Related Documentation

- `references/team-approaches.md` - Full prompts from Monark, Ankur, Rajeev
- `references/logging-standards.md` - Razorpay logging best practices
- `references/mcp-integration.md` - MCP configuration and usage
