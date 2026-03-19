# Auto-Fix Strategies

Guide to automatic fixes applied by the helmfile-debug skill.

## Auto-Fix Philosophy

**Auto-fix when**:
- Fix is safe and predictable
- Fix follows established best practices
- No risk of breaking existing functionality
- User would likely apply the same fix

**Manual fix when**:
- Fix requires business logic decisions
- Multiple valid approaches exist
- Potential security implications
- Service-specific knowledge required

## Auto-Fixable Issues

### 1. Missing Resource Limits

**Detection**: values.yaml missing `web_limits_memory`

**Auto-Fix Applied**:
```yaml
# Added to values.yaml
web_limits_memory: 100Mi
```

**Rationale**: Memory limits prevent one pod from consuming all node memory and causing OOM on the node

**Default Values**:
- Memory limit: 2x memory request or 100Mi (whichever higher)

**Important Note**: CPU limits are intentionally NOT set. CPU limits cause throttling even when CPU is available on the node, leading to unnecessary performance degradation. Without CPU limits, applications can burst to use available CPU resources without artificial restrictions.

### 2. OOMKilled Memory Increase

**Detection**: Pod terminated with `OOMKilled` status

**Auto-Fix Applied**:
Progressive increase based on current limit:
```yaml
# If current < 100Mi
web_limits_memory: 200Mi

# If current < 200Mi
web_limits_memory: 512Mi

# If current < 512Mi
web_limits_memory: 1Gi
```

**Rationale**: Container needs more memory to run successfully

**Limits**: Maximum 3 automatic increases, then report for manual intervention

**Actions**:
1. Update values.yaml
2. Redeploy service
3. Monitor for 2 minutes
4. If still OOMKilled, repeat (up to 3 times)

### 3. Missing DNS Policy

**Detection**: deployment.yaml missing `dnsPolicy`

**Auto-Fix Applied**:
```yaml
# Added to deployment.yaml spec
dnsPolicy: ClusterFirst
dnsConfig:
  options:
  - name: ndots
    value: "1"
```

**Rationale**:
- `ClusterFirst` enables proper Kubernetes DNS resolution
- `ndots: 1` reduces unnecessary DNS lookups

**Benefits**:
- Services can resolve by short names
- Faster DNS resolution
- Fewer DNS queries

### 4. Missing TTL Annotation

**Detection**: deployment.yaml missing `janitor/ttl` annotation

**Auto-Fix Applied**:
```yaml
# Added to metadata.annotations
metadata:
  annotations:
    janitor/ttl: "{{ .Values.ttl }}"
```

**Rationale**: Enables automatic cleanup of devstack resources

**Importance**: Prevents resource accumulation in cluster

### 5. Missing devstack_label in Labels

**Detection**: deployment.yaml missing `devstack_label` in labels

**Auto-Fix Applied**:
```yaml
# Added to metadata.labels
metadata:
  labels:
    devstack_label: {{ .Values.devstack_label }}
    app: {{ .Chart.Name }}
```

**Rationale**: Required for filtering and identifying pods

**Use Cases**:
- kubectl filtering by label
- Service selectors
- Monitoring queries

### 6. Commented Service in helmfile.yaml

**Detection**: Service release block has lines starting with `#`

**Auto-Fix Applied**:
```yaml
# Before
# - name: payment-service-{{ .Values.devstack_label }}
#   namespace: payment-service
#   chart: ./charts/payment-service

# After (uncommented)
- name: payment-service-{{ .Values.devstack_label }}
  namespace: payment-service
  chart: ./charts/payment-service
```

**Rationale**: Service needs to be active to deploy

**Caution**: Verifies entire block before uncommenting

### 7. Basic HTTP Health Probes

**Detection**: deployment.yaml missing `livenessProbe` or `readinessProbe`

**Auto-Fix Applied** (only if port 8080 detected or standard):
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 2
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 2
  failureThreshold: 3
```

**Rationale**: Health probes are critical for Kubernetes to manage pods

**Limitations**: Only auto-fixes for standard HTTP services on port 8080

**Manual Required**: Custom ports, TCP probes, exec probes

## Conditional Auto-Fixes

### Worker Resource Limits (if workers exist)

**Detection**: deployment.yaml has worker container but values.yaml missing worker limits

**Auto-Fix Applied**:
```yaml
worker_requests_cpu: 60m
worker_requests_memory: 150Mi
worker_limits_memory: 256Mi
```

**Condition**: Only applied if worker container detected in templates

**Note**: CPU limits are intentionally NOT set to prevent throttling.

### Missing Namespace in Values

**Detection**: values.yaml missing `namespace` but namespace in helmfile.yaml

**Auto-Fix Applied**:
```yaml
# Extract from helmfile.yaml and add to values.yaml
namespace: <extracted-namespace>
```

**Rationale**: Consistency between files

## Manual Intervention Required

These issues CANNOT be auto-fixed:

### 1. Environment Variables

**Reason**: Service-specific configuration

**Example**:
```yaml
env:
  - name: REDIS_HOST
    value: ?  # Unknown - needs user input
```

**Recommendation Provided**:
```
Add to values.yaml or secret:
  env:
    - name: REDIS_HOST
      value: "redis-base.redis.svc.cluster.local"
```

### 2. Database Connection Strings

**Reason**: Environment-specific, security-sensitive

**Recommendation Provided**:
```
Update secret or values with correct database URL
```

### 3. Secrets

**Reason**: Security - cannot auto-create or modify secrets

**Recommendation Provided**:
```bash
kubectl create secret generic <name> \
  --from-literal=KEY=value \
  -n <namespace>
```

### 4. Network Policies

**Reason**: Security-sensitive, requires network architecture knowledge

**Recommendation Provided**:
```
Consult with platform team for network policy requirements
```

### 5. Custom Probe Configurations

**Reason**: Application-specific health endpoints

**Recommendation Provided**:
```yaml
# Customize for your application
livenessProbe:
  httpGet:
    path: /your-health-endpoint
    port: <your-port>
```

### 6. Ingress Configurations

**Reason**: Domain and routing specific to environment

**Recommendation Provided**:
```yaml
# Add ingress configuration based on your domain
```

### 7. Service Mesh Settings

**Reason**: Complex configuration requiring architectural knowledge

**Recommendation Provided**:
```
Consult service mesh documentation for your use case
```

### 8. Application Bugs

**Reason**: Code-level fixes required

**Recommendation Provided**:
```
Fix identified code issue:
  File: payment/processor.go:145
  Issue: Nil pointer dereference
  Action: Add nil check before access
```

## Auto-Fix Workflow

### Step 1: Detection
```
Scan configuration files for issues
├── values.yaml validation
├── deployment.yaml validation
└── helmfile.yaml validation
```

### Step 2: Classification
```
For each issue:
├── Is it auto-fixable? → Apply fix
├── Requires manual input? → Provide recommendation
└── Complex decision? → Provide options
```

### Step 3: Application
```
Apply fixes:
├── Update files
├── Validate syntax
└── Record changes
```

### Step 4: Reporting
```
Report to user:
├── List of auto-fixes applied
├── Files modified with line numbers
└── Actions that need manual intervention
```

## Auto-Fix Safety

### Pre-Fix Validation
- Verify fix won't break existing config
- Check for dependencies
- Validate syntax

### Post-Fix Validation
- Run template rendering
- Verify YAML validity
- Check for conflicts

### Rollback
If auto-fix causes template errors:
- Revert changes
- Report issue
- Provide manual fix recommendation

## Auto-Fix Limits

### Retry Limits
- OOMKilled: Maximum 3 automatic increases
- Deployment failures: Maximum 2 retry attempts
- Configuration fixes: No limit (safe operations)

### Resource Limits
- Memory: Don't auto-increase beyond 2Gi

**Rationale**: Prevent resource waste, may indicate architectural issue

**Note**: CPU limits are not set as they cause unnecessary throttling

## Reporting Format

### Auto-Fix Success Report
```
⚠️ Auto-Fixes Applied

Files Modified:
1. charts/<service>/values.yaml:23
   + web_limits_memory: 200Mi

2. charts/<service>/templates/deployment.yaml:15
   + janitor/ttl: "{{ .Values.ttl }}"

3. helmfile.yaml:892-897
   Uncommented service block

Validation: ✅ All changes validated
Template Rendering: ✅ Success
Ready to Deploy: ✅ Yes
```

### Mixed Auto-Fix and Manual Report
```
Configuration Issues Found

Auto-Fixed (3):
✅ Added web_limits_memory: 200Mi
✅ Added janitor/ttl annotation
✅ Uncommented service in helmfile.yaml

Manual Action Required (2):
❌ Empty image tag
   Fix: Update helmfile.yaml with valid commit hash

❌ Missing REDIS_HOST environment variable
   Fix: Add to values.yaml:
   env:
     - name: REDIS_HOST
       value: "redis-base.redis.svc.cluster.local"
```

## Configuration

### Enable/Disable Auto-Fix

By default, all safe auto-fixes are enabled. To disable:

```
Deploy with manual fixes only
```

### Selective Auto-Fix

```
Auto-fix resource limits only, skip other fixes
```

## Best Practices

1. **Review Auto-Fixes**: Always check what was auto-fixed
2. **Understand Changes**: Know why each fix was applied
3. **Validate Results**: Verify deployment works as expected
4. **Update Source**: Commit auto-fixes to git
5. **Monitor Impact**: Watch for any unexpected behavior

## Future Enhancements

Potential future auto-fixes:
- Service mesh integration
- Horizontal Pod Autoscaler configuration
- PodDisruptionBudget setup
- Network policies (basic patterns)
- Resource quota checks
