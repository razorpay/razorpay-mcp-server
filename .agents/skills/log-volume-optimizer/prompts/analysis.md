# Log Volume Analysis Prompt

## Context
You are analyzing log statements in a Razorpay service to estimate their volume impact and identify optimization opportunities.

## Service Information
- **Service Name**: {{service_name}}
- **Repository Path**: {{repo_path}}
- **Language**: {{language}} (Go/PHP/TypeScript/Node.js/Python)
- **Assigned Quota**: {{assigned_units}} units/day
- **Current Tier**: {{tier}}

## Traffic Parameters (Fallback — used only when Coralogix MCP data is unavailable)
> **Note**: With Coralogix MCP integration, real production log frequencies are used.
> These parameters are only needed as a fallback for log messages that don't match any Coralogix entry.

- **Daily Active Merchants**: {{merchants_daily}}
- **Peak RPS**: {{peak_rps}}
- **Average RPS**: {{avg_rps}}
- **Request Distribution**: {{request_distribution}}

## Analysis Instructions

### Step 1: Scan Repository
Scan all source files for log statements. Patterns by language:

**Go:**
```go
logger.Info("message")
logger.Log(ctx).Info("message")
logger.Log(ctx).With("key", value).Info("message")
lgr.Logger(ctx).Fatal("message")
```

**PHP:**
```php
Log::info("message")
$trace->info("message")
$this->trace->debug("message", ["key" => $value])
```

**TypeScript / Node.js:**
```typescript
console.log("message")
logger.info("message")
Logger.debug("message", { key: value })
```

**Python:**
```python
logging.info("message")
logger.debug("message")
log.error("message", extra={"key": value})
```

### Step 2: Extract Metadata
For each log statement, extract:
1. **File path** and **line number**
2. **Function name** containing the log
3. **Log level** (DEBUG, INFO, WARN, ERROR, FATAL)
4. **Log message** and any structured fields
5. **Context flags**:
   - Is it in a loop?
   - Is it in an error handler?
   - Is it in a hot path (high RPS route)?

### Step 3: Estimate Volume
For each log statement, calculate:

```
trigger_probability = 1.0  # Default
if in_error_handler:
    trigger_probability = 0.01  # 1% error rate assumption
if in_conditional:
    trigger_probability = 0.5   # 50% branch probability

daily_invocations = avg_rps × 86400 × trigger_probability
log_size_bytes = len(message) + len(structured_fields) + 200  # overhead
daily_bytes = daily_invocations × log_size_bytes
daily_units = daily_bytes / 1_000_000
```

### Step 4: Categorize by Impact
Group logs into categories:
- **Critical** (>100 units/day): Must optimize
- **High** (10-100 units/day): Should optimize
- **Medium** (1-10 units/day): Consider optimizing
- **Low** (<1 unit/day): Acceptable

### Step 5: Compare with Actual
Use Coralogix MCP to get actual consumption:
```
Query: sum(cx_data_usage_units{application_name="{{service_name}}"}) by (application_name)
Time range: Last 7 days
Method: Daily peak detection
```

Calculate variance:
```
variance = (actual_units - estimated_units) / actual_units × 100
```

If variance > 20%, investigate:
- Missing log sources
- External services logging to same app
- Log sampling already in place
- Different log sizes than estimated

## Output Format

### Summary Table
```markdown
| Metric | Value |
|--------|-------|
| Total Log Statements | X |
| Estimated Daily Units | X |
| Actual Daily Units | X |
| Variance | X% |
| Quota Utilization | X% |
```

### Top 10 High-Impact Logs
```markdown
| Rank | File:Line | Level | Function | Est. Units/Day | Issue |
|------|-----------|-------|----------|----------------|-------|
| 1 | file.go:123 | INFO | HandleRequest | 500 | Hot path |
| 2 | ... | ... | ... | ... | ... |
```

### Category Breakdown
```markdown
| Category | Count | Units/Day | % of Total |
|----------|-------|-----------|------------|
| Critical | X | X | X% |
| High | X | X | X% |
| Medium | X | X | X% |
| Low | X | X | X% |
```

## Questions to Ask User
If information is missing, ask:
1. What is the average RPS for this service?
2. How many merchants use this service daily?
3. Are there any routes with significantly higher traffic?
4. Is log sampling already implemented?
5. Are there external services logging to the same Coralogix application?
