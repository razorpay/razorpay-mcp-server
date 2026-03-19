# Validation Subskill

Configuration validation and best practices compliance for helmfile deployments.

## Purpose

Validates helm chart configurations against best practices before deployment to prevent common issues.

## When to Use

- Before deploying a new service
- After making configuration changes
- When troubleshooting deployment issues
- For configuration audits

## Validation Checklist

### values.yaml Validation

**Required Fields**:

```yaml
# ✅ Must Have
namespace: <service-namespace>              # Kubernetes namespace
devstack_label: <label>                     # Environment identifier
ttl: <1h|8h|forever>                       # Time-to-live for cleanup
image: <commit-hash>                       # Container image tag
secret_name: <secret-name>                 # Secret reference

# Node placement
node_selector:
  environment: devstack
base_node_selector:
  environment: base

# Web container resources (REQUIRED)
web_requests_cpu: 50m
web_requests_memory: 50Mi
web_limits_memory: 100Mi
# NOTE: CPU limits intentionally NOT set to prevent throttling

# Worker resources (if workers exist)
worker_requests_cpu: 60m
worker_requests_memory: 150Mi
worker_limits_memory: 256Mi
# NOTE: CPU limits intentionally NOT set to prevent throttling
```

**Validation Actions**:

1. **Check all required fields exist**
   ```bash
   grep -E "namespace|devstack_label|ttl|image|secret_name|web_requests_cpu|web_requests_memory|web_limits_memory" charts/<service>/values.yaml
   ```

2. **Validate resource format**
   - CPU: Must end with 'm' (millicores) or be integer
   - Memory: Must end with 'Mi', 'Gi', etc.
   - Examples: `50m`, `100Mi`, `1Gi`

3. **Check TTL values**
   - Valid: `1h`, `8h`, `forever`
   - Invalid: `2h`, `24h`, `1d` (use 8h or forever)

4. **Verify image tag**
   - Should be commit hash or semantic version
   - Should NOT be empty
   - Should NOT be `latest` (not recommended)

**Common Issues**:

❌ **Missing Memory Limit**:
```yaml
# Invalid - no memory limit
web_requests_cpu: 50m
web_requests_memory: 50Mi
```

✅ **Fixed**:
```yaml
web_requests_cpu: 50m
web_requests_memory: 50Mi
web_limits_memory: 100Mi    # Added
# CPU limits intentionally NOT set to prevent throttling
```

❌ **Invalid TTL**:
```yaml
ttl: 2h  # Invalid value
```

✅ **Fixed**:
```yaml
ttl: 8h  # Valid value
```

### deployment.yaml Validation

**Required Configurations**:

```yaml
# ✅ Metadata annotations
metadata:
  annotations:
    janitor/ttl: "{{ .Values.ttl }}"      # REQUIRED for cleanup

# ✅ Metadata labels
  labels:
    devstack_label: {{ .Values.devstack_label }}  # REQUIRED for filtering
    app: {{ .Chart.Name }}

# ✅ Pod spec
spec:
  # DNS configuration
  dnsPolicy: ClusterFirst                  # REQUIRED
  dnsConfig:
    options:
    - name: ndots
      value: "1"

  # Node placement
  nodeSelector:
    {{ toYaml .Values.node_selector | indent 4 }}

  # Containers
  containers:
  - name: web
    image: "c.rzp.io/razorpay/{{ .Chart.Name }}:{{ .Values.image }}"

    # Resource limits (REQUIRED)
    resources:
      requests:
        cpu: {{ .Values.web_requests_cpu }}
        memory: {{ .Values.web_requests_memory }}
      limits:
        memory: {{ .Values.web_limits_memory }}
        # CPU limits intentionally omitted to prevent throttling

    # Liveness probe (REQUIRED)
    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 10
      timeoutSeconds: 2
      failureThreshold: 3

    # Readiness probe (REQUIRED)
    readinessProbe:
      httpGet:
        path: /ready
        port: 8080
      initialDelaySeconds: 10
      periodSeconds: 10
      timeoutSeconds: 2
      failureThreshold: 3
```

**Validation Actions**:

1. **Check for required annotations**
   ```bash
   grep -A 2 "annotations:" charts/<service>/templates/deployment.yaml | grep "janitor/ttl"
   ```

2. **Verify labels**
   ```bash
   grep "devstack_label" charts/<service>/templates/deployment.yaml
   ```

3. **Check DNS policy**
   ```bash
   grep "dnsPolicy" charts/<service>/templates/deployment.yaml
   ```

4. **Verify resource blocks**
   ```bash
   grep -A 5 "resources:" charts/<service>/templates/deployment.yaml
   ```

5. **Check for probes**
   ```bash
   grep -E "livenessProbe|readinessProbe" charts/<service>/templates/deployment.yaml
   ```

**Common Issues**:

❌ **Missing TTL Annotation**:
```yaml
metadata:
  annotations:
    app: payment-service
  # Missing janitor/ttl!
```

✅ **Fixed**:
```yaml
metadata:
  annotations:
    app: payment-service
    janitor/ttl: "{{ .Values.ttl }}"  # Added
```

❌ **Missing DNS Policy**:
```yaml
spec:
  # No dnsPolicy!
  containers:
  - name: web
```

✅ **Fixed**:
```yaml
spec:
  dnsPolicy: ClusterFirst      # Added
  dnsConfig:                   # Added
    options:
    - name: ndots
      value: "1"
  containers:
  - name: web
```

❌ **No Probes**:
```yaml
containers:
- name: web
  image: ...
  # No probes!
```

✅ **Fixed**:
```yaml
containers:
- name: web
  image: ...
  livenessProbe:              # Added
    httpGet:
      path: /health
      port: 8080
    initialDelaySeconds: 30
    periodSeconds: 10
  readinessProbe:             # Added
    httpGet:
      path: /ready
      port: 8080
    initialDelaySeconds: 10
    periodSeconds: 10
```

### helmfile.yaml Validation

**Required Structure**:

```yaml
releases:
  - name: {{ .Chart.Name }}-{{ .Values.devstack_label }}
    namespace: {{ .Values.namespace }}
    chart: ./charts/<service-name>
    values:
      - image: <commit-hash>
      - devstack_label: {{ .Values.devstack_label }}
      - ttl: {{ .Values.ttl }}
      - namespace: {{ .Values.namespace }}
```

**Validation Actions**:

1. **Find service entry**
   ```bash
   grep -n "name: <service>-{{ .Values.devstack_label }}" helmfile.yaml
   ```

2. **Check if commented**
   - Lines starting with `#` are commented
   - Auto-uncomment if found

3. **Verify required values**
   - image
   - devstack_label
   - ttl
   - namespace

4. **Validate chart path**
   - Must point to `./charts/<service-name>`
   - Directory must exist

**Common Issues**:

❌ **Service Commented Out**:
```yaml
# - name: payment-service-{{ .Values.devstack_label }}
#   namespace: payment-service
#   chart: ./charts/payment-service
```

✅ **Auto-Fixed**:
```yaml
- name: payment-service-{{ .Values.devstack_label }}
  namespace: payment-service
  chart: ./charts/payment-service
```

❌ **Missing Values**:
```yaml
- name: api-gateway-{{ .Values.devstack_label }}
  chart: ./charts/api-gateway
  # Missing image and other values!
```

✅ **Fixed**:
```yaml
- name: api-gateway-{{ .Values.devstack_label }}
  namespace: api-gateway
  chart: ./charts/api-gateway
  values:
    - image: abc123def
    - devstack_label: {{ .Values.devstack_label }}
    - ttl: {{ .Values.ttl }}
```

## Template Validation

Run helmfile template to catch syntax errors:

```bash
helmfile -f helmfile.yaml -l name=<service>-<devstack-label> template
```

**What This Checks**:
- YAML syntax errors
- Missing template variables
- Invalid Go template syntax
- Reference to non-existent values

**Example Errors**:

❌ **Missing Value**:
```
Error: template: deployment.yaml:45:20: executing "deployment.yaml"
at <.Values.missing_value>: map has no entry for key "missing_value"
```

**Fix**: Add the missing value to values.yaml

❌ **YAML Syntax Error**:
```
Error: yaml: line 32: mapping values are not allowed in this context
```

**Fix**: Check indentation and YAML formatting at line 32

## Best Practices Validation

### Resource Sizing

**CPU Guidelines**:
- Minimal service: 50m - 100m
- Standard service: 100m - 200m
- CPU-intensive: 200m - 500m

**Memory Guidelines**:
- Minimal service: 50Mi - 100Mi
- Standard service: 100Mi - 512Mi
- Memory-intensive: 512Mi - 2Gi

**Validate Reasonable Limits**:
```
⚠️ Warning: Very high resource requests

Current Configuration:
  web_requests_cpu: 2000m      # 2 full CPU cores
  web_requests_memory: 4Gi     # 4GB memory

Recommendation:
  Consider if this is necessary for devstack environment.
  Typical devstack services use 50m-200m CPU and 100Mi-512Mi memory.
```

### Probe Configuration

**Liveness Probe Best Practices**:
- Use HTTP GET for web services
- Set appropriate `initialDelaySeconds` (30s typical)
- Don't make it too sensitive (`failureThreshold: 3`)
- Use a simple endpoint that checks basic health

**Readiness Probe Best Practices**:
- Shorter `initialDelaySeconds` than liveness (10s typical)
- Can be more strict than liveness
- Should check actual readiness to serve traffic
- May check dependencies (DB, cache, etc.)

**Common Mistakes**:

❌ **Liveness Probe Too Strict**:
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5       # Too short!
  failureThreshold: 1          # Too strict!
```

✅ **Better Configuration**:
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30      # Allow time to start
  periodSeconds: 10
  failureThreshold: 3          # More forgiving
```

### Security Best Practices

**Service Account**:
```yaml
serviceAccountName: <service-name>  # Use dedicated SA
# Not: serviceAccountName: default
```

**Security Context**:
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  capabilities:
    drop:
      - ALL
```

**Resource Quotas**:
- Always set memory limits (prevent OOM affecting node)
- Do NOT set CPU limits (prevents throttling, allows bursting to available CPU)
- Use requests to guarantee minimum resources

## Auto-Fix Strategy

### What Gets Auto-Fixed

The following issues are automatically corrected:

1. **Missing resource limits** → Add defaults
2. **Missing TTL annotation** → Add `janitor/ttl`
3. **Missing devstack_label** → Add to labels
4. **Missing DNS policy** → Add `ClusterFirst`
5. **Commented service** → Uncomment
6. **Missing basic probes** → Add HTTP probes (if port is standard)

### What Requires Manual Review

The following require manual intervention:

1. **Custom probe endpoints** → Need service-specific paths
2. **Non-standard ports** → Need actual port numbers
3. **Environment variables** → Service-specific config
4. **Secrets** → Cannot auto-create (security)
5. **Volume mounts** → Service-specific needs
6. **Init containers** → Complex startup logic

## Validation Report Format

```
## 🔍 Configuration Validation Report

### values.yaml ✅
✅ All required fields present
✅ Resource limits properly configured
✅ TTL value valid (8h)
⚠️ Image tag empty - needs to be provided

### deployment.yaml ⚠️
✅ TTL annotation present
✅ devstack_label in labels
⚠️ Missing liveness probe - AUTO-FIXED
⚠️ Missing DNS policy - AUTO-FIXED
✅ Resource blocks properly configured

### helmfile.yaml ✅
✅ Service entry found (line 892)
⚠️ Service was commented - AUTO-UNCOMMENTED
✅ Chart path valid
✅ Required values present

### Template Rendering ✅
✅ No syntax errors
✅ All variables resolved
✅ Valid Kubernetes YAML generated

### Overall Status: READY TO DEPLOY ✅
- 3 auto-fixes applied
- 1 manual input needed (image tag)
```

## Validation Commands

```bash
# Validate values.yaml exists
test -f charts/<service>/values.yaml && echo "✅ values.yaml found" || echo "❌ values.yaml missing"

# Check for required fields
grep -q "web_limits_memory" charts/<service>/values.yaml && echo "✅ Memory limits set" || echo "❌ Missing memory limits"

# Validate template syntax
helmfile -f helmfile.yaml -l name=<service>-<label> template > /dev/null && echo "✅ Template valid" || echo "❌ Template errors"

# Check for janitor/ttl annotation
grep -q "janitor/ttl" charts/<service>/templates/deployment.yaml && echo "✅ TTL annotation present" || echo "❌ Missing TTL"
```

## Related Subskills

- [Deployment](deployment.md) - Uses validation before deploying
- [Debugging](debugging.md) - Validation helps prevent issues
- [Config Checklist](../references/config-checklist.md) - Complete validation checklist
