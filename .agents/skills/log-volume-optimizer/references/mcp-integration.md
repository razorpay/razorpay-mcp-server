# MCP Integration Guide

How to integrate with Coralogix and Grafana MCP servers for log volume analysis.

## Overview

The log-volume-optimizer skill uses two MCP servers:
1. **Coralogix MCP** - Fetch actual log consumption data
2. **Grafana MCP** - Fetch RPS data from Prometheus

## Coralogix MCP

### Configuration

**Endpoint:** `https://api.coralogix.in/mgmt/api/v1/mcp`

**Required Environment Variable:**
```bash
export CORALOGIX_API_KEY="your-coralogix-api-key"
```

### Available Tools

#### `mcp_coralogix-server_metrics__range_query`
Execute a PromQL range query against Coralogix metrics.

**Parameters:**
- `query`: PromQL expression
- `start`: Start time (RFC3339)
- `end`: End time (RFC3339)
- `step`: Query resolution (e.g., "1h")
- `limit`: Maximum series to return

**Example Usage:**
```python
# Fetch units consumption for pg-router over 7 days
result = mcp_coralogix-server_metrics__range_query(
    query='sum(cx_data_usage_units{application_name="pg-router"}) by (application_name)',
    start="2024-01-01T00:00:00Z",
    end="2024-01-08T00:00:00Z",
    step="1h",
    limit=200
)
```

#### `mcp_coralogix-server_get_logs`
Query logs using Dataprime syntax.

**Example:**
```python
# Get recent error logs
result = mcp_coralogix-server_get_logs(
    query='source logs | filter $m.severity == ERROR | filter $l.applicationname == "pg-router"',
    start_date="2024-01-07T00:00:00Z",
    end_date="2024-01-08T00:00:00Z",
    limit=100
)
```

### Units Consumption Query

The primary query for fetching units consumption:

```promql
sum(cx_data_usage_units{application_name="SERVICE_NAME"}) by (application_name)
```

**Understanding the Data:**
- `cx_data_usage_units` is a cumulative daily counter
- Resets at midnight UTC
- Query with hourly resolution to detect daily peaks
- Average the peaks to get daily consumption

**Peak Detection Algorithm:**
```python
# Detect daily resets (counter drops by >70%)
daily_peaks = []
for i in range(len(data_points) - 1):
    current_val = data_points[i]
    next_val = data_points[i + 1]
    
    # Reset detected: significant drop
    if current_val > 100 and next_val < current_val * 0.3:
        daily_peaks.append(current_val)

daily_avg = sum(daily_peaks) / len(daily_peaks)
```

### Breakdown Queries

**By Severity:**
```promql
sum(cx_data_usage_units{application_name="SERVICE"}) by (severity)
```

**By Subsystem:**
```promql
sum(cx_data_usage_units{application_name="SERVICE"}) by (subsystem_name)
```

**By Tier:**
```promql
sum(cx_data_usage_units{application_name="SERVICE"}) by (pillar)
```

---

## Grafana MCP

### Configuration

**Available via Cursor MCP integration**

Uses the standard Grafana MCP tools.

### Available Tools

#### `mcp_grafana-mcp-server_query_prometheus`
Execute PromQL queries against Prometheus datasources.

**Parameters:**
- `datasourceUid`: UID of the Prometheus datasource
- `expr`: PromQL expression
- `startTime`: Start time (RFC3339 or relative like "now-1h")
- `endTime`: End time (optional)
- `stepSeconds`: Query resolution in seconds
- `queryType`: "range" or "instant"

**Example Usage:**
```python
# Fetch RPS by route for pg-router
result = mcp_grafana-mcp-server_query_prometheus(
    datasourceUid="prometheus-prod",
    expr='sum(rate(pg_router_http_requests_total[5m])) by (route)',
    startTime="now-1h",
    queryType="instant"
)
```

#### `mcp_grafana-mcp-server_list_prometheus_metric_names`
List available metrics matching a pattern.

```python
# Find available metrics for pg-router
result = mcp_grafana-mcp-server_list_prometheus_metric_names(
    datasourceUid="prometheus-prod",
    regex="pg_router.*"
)
```

### Common RPS Queries

**Total RPS:**
```promql
sum(rate(http_requests_total{service="SERVICE"}[5m]))
```

**RPS by Route:**
```promql
sum(rate(http_requests_total{service="SERVICE"}[5m])) by (route)
```

**RPS by Handler:**
```promql
sum(rate(http_requests_total{service="SERVICE"}[5m])) by (handler)
```

**Peak RPS (last 24h):**
```promql
max_over_time(sum(rate(http_requests_total{service="SERVICE"}[5m]))[24h:])
```

---

## Integration Workflow

### Step 1: Get Traffic Data from Grafana

```python
# Fetch current RPS by route
rps_data = mcp_grafana-mcp-server_query_prometheus(
    datasourceUid="prometheus-prod",
    expr='sum(rate(pg_router_http_requests_total[5m])) by (route)',
    startTime="now-1h",
    queryType="instant"
)

# Parse into route -> RPS map
route_rps = {}
for series in rps_data:
    route = series['metric']['route']
    rps = float(series['value'][1])
    route_rps[route] = rps
```

### Step 2: Get Consumption from Coralogix

```python
# Fetch daily consumption over 7 days
consumption = mcp_coralogix-server_metrics__range_query(
    query='sum(cx_data_usage_units{application_name="pg-router"})',
    start=seven_days_ago,
    end=now,
    step="1h",
    limit=200
)

# Calculate daily average using peak detection
daily_avg = calculate_daily_peaks(consumption)
```

### Step 3: Compare and Analyze

```python
# Calculate estimated volume from scanned logs
estimated_units = sum(
    log.daily_units for log in scanned_logs
)

# Compare with actual
variance = (actual_units - estimated_units) / actual_units * 100

# Generate report
report = {
    "estimated": estimated_units,
    "actual": daily_avg,
    "variance_percent": variance,
    "quota_utilization": (daily_avg / assigned_quota) * 100
}
```

---

## Error Handling

### Coralogix MCP Errors

```python
# Handle missing API key
if not os.getenv("CORALOGIX_API_KEY"):
    return {
        "error": "CORALOGIX_API_KEY not set",
        "source": "error",
        "daily_avg_units": 0
    }

# Handle query failures
try:
    result = await mcp_call("range_query", ...)
except Exception as e:
    return {
        "error": f"MCP query failed: {e}",
        "source": "error"
    }

# Handle empty data
if not result or len(result) == 0:
    return {
        "error": "No data found for application",
        "source": "error"
    }
```

### Grafana MCP Errors

```python
# Handle datasource not found
try:
    result = mcp_grafana_query(...)
except Exception as e:
    if "datasource not found" in str(e):
        # Fall back to default RPS
        return {"rps": default_rps, "source": "fallback"}
    raise
```

---

## Testing MCP Integration

### Test Coralogix Connection

```
Use Coralogix MCP to query units for pg-router over the last 7 days
```

Expected output:
```
Application: pg-router
Daily Average: 420 units
Max Daily: 480 units
Min Daily: 380 units
Data Points: 7 days
Source: coralogix_mcp
```

### Test Grafana Connection

```
Use Grafana MCP to get RPS by route for pg-router
```

Expected output:
```
Route: /v1/payments, RPS: 200
Route: /v1/refunds, RPS: 75
Route: /v1/orders, RPS: 125
Total RPS: 500
Source: grafana_mcp
```

---

## Troubleshooting

### "CORALOGIX_API_KEY not set"
```bash
# Set the environment variable
export CORALOGIX_API_KEY="your-key-here"

# Or in .env file
echo "CORALOGIX_API_KEY=your-key" >> ~/.env
```

### "No data found for application"
- Verify `application_name` matches exactly in Coralogix
- Check if service is sending logs to Coralogix
- Verify time range has data

### "Datasource not found"
- Check datasource UID is correct
- Verify access to Grafana instance
- List available datasources first

### Timeout Errors
- Reduce query time range
- Increase step size
- Add limits to query
