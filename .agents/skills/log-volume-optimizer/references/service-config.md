# Service Configuration Guide

Configure services for log volume analysis and optimization.

## Configuration File Structure

Create a YAML configuration file for each service:

```yaml
# Required: Service identification
service:
  name: service-name
  description: Brief description of the service
  repository: razorpay/service-name
  language: go  # go, python, java

# Required: Coralogix settings
coralogix:
  application_name: service-name
  subsystem_name: service-name  # Optional, defaults to application_name

# Required: Quota settings
quota:
  assigned_units: 500    # Daily quota in Coralogix units
  tier: M                # H (high), M (medium), L (low)
  warning_threshold: 80  # Warn at this utilization %
  critical_threshold: 95 # Critical at this utilization %

# Required: Traffic parameters
traffic:
  avg_rps: 100           # Average requests per second
  peak_rps: 500          # Peak requests per second
  merchants_daily: 10000 # Daily active merchants

# Optional: Route-level traffic distribution
routes:
  /api/v1/payments: 40   # % of total traffic
  /api/v1/refunds: 20
  /api/v1/orders: 25
  /health: 10
  /other: 5

# Optional: Known high-traffic handlers
hot_paths:
  - function: HandlePayment
    route: /api/v1/payments
    rps: 40  # Specific RPS for this route

# Optional: Metrics configuration
metrics:
  rps_metric: service_http_requests_total
  rps_query: sum(rate(service_http_requests_total[5m])) by (route)
  labels:
    service: service-name
    env: production

# Optional: Log patterns to detect
logging:
  patterns:
    - logger.Log(ctx).Info
    - logger.Log(ctx).Error
  include:
    - "**/*.go"
  exclude:
    - "**/*_test.go"
    - "**/vendor/**"

# Optional: Optimization settings
optimization:
  target_reduction: 30  # Target % reduction
```

## Example Configurations

### pg-router

```yaml
service:
  name: pg-router
  description: Payment Gateway Router - Routes payment requests
  repository: razorpay/pg-router
  language: go

coralogix:
  application_name: pg-router

quota:
  assigned_units: 480
  tier: M
  warning_threshold: 80
  critical_threshold: 95

traffic:
  avg_rps: 500
  peak_rps: 2000
  merchants_daily: 50000

routes:
  /v1/payments: 40
  /v1/refunds: 15
  /v1/orders: 25
  /v1/settlements: 10
  /health: 5
  /other: 5

hot_paths:
  - function: HandlePayment
    route: /v1/payments
    rps: 200
  - function: HandleRefund
    route: /v1/refunds
    rps: 75

metrics:
  rps_metric: pg_router_http_requests_total
  rps_query: sum(rate(pg_router_http_requests_total[5m])) by (route)

logging:
  patterns:
    - logger.Log(ctx).Info
    - logger.Log(ctx).Error
    - logger.Log(ctx).Debug
    - lgr.Logger(ctx).Info
    - lgr.Logger(ctx).Fatal
  include:
    - "**/*.go"
  exclude:
    - "**/*_test.go"
    - "**/testdata/**"
    - "**/vendor/**"
    - "**/mock/**"

optimization:
  target_reduction: 30
```

### api (Monolith)

```yaml
service:
  name: api
  description: Razorpay API Monolith
  repository: razorpay/api
  language: php  # Or mixed

coralogix:
  application_name: api

quota:
  assigned_units: 800
  tier: H
  warning_threshold: 75
  critical_threshold: 90

traffic:
  avg_rps: 2000
  peak_rps: 10000
  merchants_daily: 200000

optimization:
  target_reduction: 25
```

### dashboard

```yaml
service:
  name: dashboard
  description: Merchant Dashboard Backend
  repository: razorpay/dashboard
  language: go

coralogix:
  application_name: dashboard

quota:
  assigned_units: 200
  tier: L
  warning_threshold: 80
  critical_threshold: 95

traffic:
  avg_rps: 100
  peak_rps: 500
  merchants_daily: 30000

optimization:
  target_reduction: 40
```

## Getting Traffic Parameters

### From Grafana

Query average RPS:
```promql
avg(rate(http_requests_total{service="SERVICE"}[1h]))
```

Query peak RPS:
```promql
max_over_time(rate(http_requests_total{service="SERVICE"}[5m])[24h:])
```

Query route distribution:
```promql
sum(rate(http_requests_total{service="SERVICE"}[1h])) by (route)
```

### From Coralogix

Query assigned quota:
```
Check service_units.csv or contact Platform team
```

Query actual consumption:
```
sum(cx_data_usage_units{application_name="SERVICE"})
```

## Quota Tiers

| Tier | Description | Typical Quota |
|------|-------------|---------------|
| H | High priority, critical path | 500-1000 units |
| M | Medium priority, standard | 200-500 units |
| L | Low priority, internal | 50-200 units |

## Validation

Before using a configuration:

1. **Verify service exists** in Coralogix with `application_name`
2. **Check quota** matches assigned units from Platform team
3. **Validate RPS** against Grafana dashboards
4. **Test patterns** match actual log statements in code

## Adding New Service

1. Copy the template above
2. Fill in required fields (service, coralogix, quota, traffic)
3. Add optional sections as needed
4. Save to `references/services/{service-name}.yaml`
5. Test with: `Analyze log volume for {service-name}`
