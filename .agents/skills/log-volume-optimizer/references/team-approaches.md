# Log Optimization Approaches

This file contains the consolidated approaches for log volume optimization.
These approaches are the foundation of the decision engine used by this skill.

---

## Approach 1: Frequency-Based Optimization

### Query
```dataprime
source logs | countby $d.message
```

### Workflow
1. Run query in Coralogix
2. Download results as CSV to repo
3. Run the prompt below

### Prompt
```
You are a logging optimization expert. I want you to:

Step 1: Scan the entire codebase
Parse and understand all source files — not just the current file.
Build a call graph and execution flow for the entire application.
Locate all log statements across services, packages, and modules.

Step 2: Use the attached CSV (log_name, count) to:
Identify high-frequency log statements.
Match log names in CSV to actual code locations.

Step 3: Recommend log improvements:
Remove redundant or noisy logs.
Reduce verbosity: eliminate repeated data, shrink payloads.
Replace logs with metrics/traces if applicable.
Add missing logs for critical errors or transitions.
Downgrade log levels where full verbosity isn't needed.
Avoid logging inside loops or retries — recommend moving out or rate-limiting.
Refactor logs that interpolate strings even when log level is disabled.
Highlight logs with secrets/PII or unnecessary large blobs.

Output:
Table with: log_location, action (remove/optimize/add), reason, suggested_code_change
Summary: how much logging volume (and estimated cost) will reduce
Suggestions for ongoing logging best practices
Please be comprehensive and scan the entire codebase. Don't limit your analysis to current files only.
```

### Final Step
Review prompt results and apply recommended changes

---

## Approach 2: Log Level Correction

### Prompt
```
Read the entire codebase and analyze all logging statements. Identify logs that are redundant, overly verbose, lack context, or are misclassified (e.g., info used instead of debug, error without exception info, etc.). Generate a structured list grouped by file with the following:

Line number and log content
Problem description (e.g., duplicated data, wrong level, leaking sensitive data, etc.)
Suggestion for fix (e.g., downgrade level, merge with previous log, add missing context, or remove)

After creating the list, rewrite the affected logging lines with fixes applied directly in code.

Make sure to:
Reduce log injection points where not required.
Remove logs that duplicate information already available in adjacent lines or function context.
Optimize log levels based on severity (e.g., only keep warn/error for actionable issues).
Flag logs that may expose sensitive information.
Ensure all logs are structured and easy to parse (e.g., use consistent key-value pairs).

Once you're done, summarize the improvement: total logs removed, upgraded, downgraded, and refactored.
```

### Key Points
- Identify misclassified logs (info vs debug, error without exception)
- Remove duplicate information
- Flag sensitive information exposure
- Ensure structured logging (key-value pairs)

---

## Approach 3: Comprehensive Analysis with CSV Matching

### Full Prompt
```
You are a logging-optimization expert. You will analyze a codebase and produce actionable, minimal-change suggestions to reduce log volume.

ENVIRONMENT & INPUTS
- Repo root: ~/workspace/{repo}
- Languages: ["golang", "php", "typescript"]
- CSV provided: "_count,message" (CSV file path: ~/workspace/{repo}/{repo}_trace_codes.csv). Interpret `_count` as the number of log lines emitted in a 15-minute window.

STEP 0 — SCOPE & SAFETY
- Exclude third-party / generated directories (vendor, build, config, bin, scripts, templates) unless explicitly asked.
- Work read-only: produce suggested changes as unified diff patches (git diff format), not by committing to the repo.

STEP 1 — CODEBASE ANALYSIS
- Parse all source files under repo root (respect exclusions). Build a call graph and identify modules that emit logs.
- Detect logging statements including custom wrapper patterns (logger.*, log.*, audit.*, zap.S(), etc.).
- Map each discovered log statement to a canonical `message` (prefer explicit event name; else fingerprint message template).
- For each log: capture file:path:line, log level, message template, interpolated fields, surrounding function/method, and whether inside loops/retries.

STEP 2 — CSV MATCHING & FREQUENCY
- Load CSV and match `message` entries to discovered log locations. If multiple locations map to same message, list all.
- For unmatched CSV entries, attempt fuzzy match to templates and report confidence.

STEP 3 — INFO-LEVEL POLICY (explicit)
- Engineers have used INFO liberally. Wherever appropriate, choose exactly one of: KEEP (leave INFO as-is), DOWNGRADE_TO_DEBUG (change level to DEBUG), or DELETE (remove the log).
- For logs inside hot loops/retries, prefer DELETE or DOWNGRADE_TO_DEBUG plus aggregation; never recommend keeping full per-item INFO logs.

STEP 4 — ACTIONABLE RECOMMENDATIONS (for each high-frequency or noisy log)
- Decide one action per log_location: REMOVE / DOWNGRADE_TO_DEBUG / AGGREGATE / METRIC / KEEP / ADD (if missing critical log).
- For each action produce:
  - log_location (file:path:line)
  - action (one of above)
  - reason (1-line)
  - suggested_code_change (unified diff patch; or exact code snippet replacement if diff not possible)
  - estimated_log_volume_reduction (lines/day or percent) based on CSV counts
  - confidence (high/medium/low) and rationale for that confidence
- If SQS consumers are present, include a recommendation that the queue/consumer run in —quiet mode

OUTPUTS (required)
1. log_optimization_summary.csv CSV: columns = log_location, action, reason, suggested_code_change(path to patch), est_lines_reduced_per_period, confidence.
2. Zip file of all suggested patch diffs at ~/workspace/{repo}/log_patches.zip and a single file ~/workspace/{repo}/log_recommendations.md with executive summary.
3. Summary section: total estimated %log lines reduced, total bytes/day reduced + assumptions used.

BEHAVIOR & FORMAT
- Focus on code-level and collector-level removals, downgrades, aggregation, and metric substitution.
- Do NOT change production configuration or commit files — output only diffs/patches.
- Start by printing a single-line summary of how many files and logging statements you found, then proceed with the deliverables above.

Be exhaustive but practical. Start with the single-line summary, then proceed with the required deliverables.
```

### Key Actions
| Action | Description |
|--------|-------------|
| KEEP | Leave as-is (justified INFO/ERROR/WARN) |
| DOWNGRADE_TO_DEBUG | Change INFO to DEBUG |
| DELETE/REMOVE | Remove the log entirely |
| AGGREGATE | Combine loop logs into summary |
| METRIC | Replace with Prometheus metric |
| ADD | Add missing critical log |

### INFO-Level Policy
- **Never keep full per-item INFO logs in loops/retries**
- Choose exactly ONE of: KEEP, DOWNGRADE_TO_DEBUG, DELETE

### Exclusions
- vendor/
- build/
- config/
- bin/
- scripts/
- templates/

---

## Consolidated Rules for Decision Engine

Based on all approaches, here are the consolidated rules:

### ALWAYS KEEP (Do Not Modify)
1. ERROR and FATAL level logs
2. WARN level logs
3. Logs in error handlers (`if err != nil`)
4. Logs with "failed", "error", "timeout", "panic" in message
5. Critical business events (payment, transaction, settlement, refund)
6. Audit/compliance logs
7. Security-related logs

### DOWNGRADE TO DEBUG
1. Entry/exit logs ("entering", "exiting", "starting", "completed", "begin", "end")
2. Success confirmations ("successfully", "completed successfully", "done")
3. High-frequency INFO logs (>1M/day) that aren't business-critical
4. Verbose tracing logs ("processing", "handling", "calling")
5. Request/response logging at INFO level

### DELETE/REMOVE
1. Duplicate information (same data logged in adjacent lines)
2. Logs inside loops that can't be aggregated
3. Redundant noisy logs
4. Logs that interpolate strings even when level is disabled
5. Per-item logs in batch processing

### AGGREGATE
1. Logs inside for/range loops - move to summary after loop
2. Per-item processing logs - batch into count summary
3. Retry logs - aggregate into final summary

### CONVERT TO METRIC
1. Counting patterns ("processed N requests", "items count")
2. Timing patterns ("took X ms", "duration", "latency")
3. Rate patterns ("N per second", "throughput")
4. Size patterns ("bytes processed", "payload size")

### FLAG FOR REVIEW
1. Logs that may expose secrets/PII (password, token, key, secret, credential)
2. Logs with large blob payloads (>1KB)
3. Unstructured logs (not key-value format)
4. Logs with string interpolation in hot paths

---

## Coralogix Queries

### Simple Frequency Query
```dataprime
source logs | countby $d.message
```

### Per-Application Query
```dataprime
source logs
| filter $l.applicationname == '{APP_NAME}'
| countby $d.message
| sortby _count desc
| limit 500
```

### Full Analysis Query with Cost Estimation
```dataprime
source logs
| filter $l.applicationname == '{APP_NAME}'
| create size_bytes from $d:string.length()+$l:string.length()+$m:string.length()
| create msg_string_raw from if($d.message == null, if($d.msg == null, if($d.log == null, 'null', $d.log), $d.msg), $d.message)
| create msg_string_clean from trim(arrayJoin(arraySplit(msg_string_raw, '\n'), ' '))
| create msg_string from substr(msg_string_clean, 0, 500)
| groupby msg_string, $l.applicationname, $m.priorityclass, $m.timestamp/5m as per_5_minutes
    sum(size_bytes)/(1024*1024*1024) as total_bytes_gb_per_5_min
| groupby msg_string, applicationname, priorityclass
    agg avg(total_bytes_gb_per_5_min) as avg_size_per_hour_in_selected_timeframe
| create estimated_cost_dollars_month from case {
    priorityclass == 'low' -> round(avg_size_per_hour_in_selected_timeframe * 12 * 12 * 30 * 0.59 * 0.12),
    priorityclass == 'medium' -> round(avg_size_per_hour_in_selected_timeframe * 12 * 12 * 30 * 0.59 * 0.32),
    priorityclass == 'high' -> round(avg_size_per_hour_in_selected_timeframe * 12 * 12 * 30 * 0.59 * 0.75)
}
| create estimated_units_month from round(estimated_cost_dollars_month/0.59, 2)
| sortby estimated_cost_dollars_month desc
```

---

## Language Support

| Language | File Extensions | Log Patterns |
|----------|-----------------|--------------|
| Go | .go | `logger.Info()`, `log.Error()`, `lgr.Debugw()`, `zap.S().Info()` |
| PHP | .php | `Log::info()`, `$trace->info()`, `$this->trace->debug()`, `$logger->error()` |
| TypeScript | .ts, .tsx, .js | `console.log()`, `logger.info()`, `Logger.debug()` |
| Python | .py | `logging.info()`, `logger.debug()`, `log.error()` |
