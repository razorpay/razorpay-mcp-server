# Configuration Checklist

Complete validation checklist for helmfile deployments.

## values.yaml Checklist

### Core Configuration

- [ ] `namespace` - Kubernetes namespace where service will be deployed
- [ ] `devstack_label` - Unique identifier for this deployment instance
- [ ] `ttl` - Time-to-live for automatic cleanup (1h/8h/forever)
- [ ] `image` - Container image tag (commit hash or version)
- [ ] `secret_name` - Name of Kubernetes secret for environment variables

### Node Placement

- [ ] `node_selector` - Node selector for pod placement
  ```yaml
  node_selector:
    environment: devstack
  ```

- [ ] `base_node_selector` - Base environment selector
  ```yaml
  base_node_selector:
    environment: base
  ```

### Web Container Resources (Required)

- [ ] `web_requests_cpu` - Minimum CPU (e.g., `50m`)
- [ ] `web_requests_memory` - Minimum memory (e.g., `50Mi`)
- [ ] `web_limits_memory` - Maximum memory (e.g., `100Mi`)

**Note**: CPU limits are intentionally NOT set to prevent throttling. Applications can use available CPU on the node without artificial restrictions.

### Worker Container Resources (If workers exist)

- [ ] `worker_requests_cpu` - Minimum CPU (e.g., `60m`)
- [ ] `worker_requests_memory` - Minimum memory (e.g., `150Mi`)
- [ ] `worker_limits_memory` - Maximum memory (e.g., `256Mi`)

**Note**: CPU limits are intentionally NOT set to prevent throttling.

## deployment.yaml Checklist

### Metadata

- [ ] `janitor/ttl` annotation in `metadata.annotations`
  ```yaml
  annotations:
    janitor/ttl: "{{ .Values.ttl }}"
  ```

- [ ] `devstack_label` in `metadata.labels`
  ```yaml
  labels:
    devstack_label: {{ .Values.devstack_label }}
    app: {{ .Chart.Name }}
  ```

### Pod Specification

- [ ] `dnsPolicy: ClusterFirst`
- [ ] `dnsConfig` with ndots option
  ```yaml
  dnsConfig:
    options:
    - name: ndots
      value: "1"
  ```

- [ ] `nodeSelector` referencing values
  ```yaml
  nodeSelector:
    {{ toYaml .Values.node_selector | indent 4 }}
  ```

### Container Specification

- [ ] Correct image reference
  ```yaml
  image: "c.rzp.io/razorpay/{{ .Chart.Name }}:{{ .Values.image }}"
  ```

- [ ] Resource requests defined
  ```yaml
  resources:
    requests:
      cpu: {{ .Values.web_requests_cpu }}
      memory: {{ .Values.web_requests_memory }}
  ```

- [ ] Resource limits defined (memory only, no CPU limits)
  ```yaml
  resources:
    limits:
      memory: {{ .Values.web_limits_memory }}
      # CPU limits intentionally omitted to prevent throttling
  ```

### Health Checks

- [ ] Liveness probe configured
  ```yaml
  livenessProbe:
    httpGet:
      path: /health
      port: 8080
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 2
    failureThreshold: 3
  ```

- [ ] Readiness probe configured
  ```yaml
  readinessProbe:
    httpGet:
      path: /ready
      port: 8080
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 2
    failureThreshold: 3
  ```

## helmfile.yaml Checklist

### Release Configuration

- [ ] Release name includes devstack_label
  ```yaml
  - name: {{ .Chart.Name }}-{{ .Values.devstack_label }}
  ```

- [ ] Namespace specified
  ```yaml
    namespace: {{ .Values.namespace }}
  ```

- [ ] Chart path correct
  ```yaml
    chart: ./charts/<service-name>
  ```

- [ ] Required values provided
  ```yaml
    values:
      - image: <commit-hash>
      - devstack_label: {{ .Values.devstack_label }}
      - ttl: {{ .Values.ttl }}
      - namespace: {{ .Values.namespace }}
  ```

- [ ] Service is not commented out

## Security Checklist

### Best Practices

- [ ] Service account specified (not `default`)
- [ ] Security context defined
  ```yaml
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    capabilities:
      drop:
        - ALL
  ```

- [ ] Secrets referenced from Kubernetes secrets (not hardcoded)
- [ ] Sensitive data not in values.yaml

## Common Issues Checklist

### Resource Configuration

- [ ] Memory limits set (prevents node OOM)
- [ ] CPU limits NOT set (prevents throttling - this is intentional)
- [ ] Requests lower than or equal to limits
- [ ] Reasonable values (not 10Gi for simple service)

### Probe Configuration

- [ ] Liveness probe path exists in application
- [ ] Readiness probe path exists in application
- [ ] Ports match actual application ports
- [ ] `initialDelaySeconds` adequate for startup time
- [ ] `failureThreshold` not too strict (3 is good)

### Labels and Selectors

- [ ] Labels match selectors
- [ ] devstack_label in all resources
- [ ] Consistent naming across files

### DNS and Networking

- [ ] dnsPolicy set to ClusterFirst
- [ ] ndots set to 1 (reduces DNS lookups)
- [ ] Service ports match container ports
- [ ] No port conflicts

## Validation Commands

```bash
# Check values.yaml exists
test -f charts/<service>/values.yaml

# Validate required fields
grep -E "namespace|devstack_label|ttl|image|web_requests_cpu|web_requests_memory|web_limits_memory" charts/<service>/values.yaml

# Check deployment.yaml
grep "janitor/ttl" charts/<service>/templates/deployment.yaml
grep "devstack_label" charts/<service>/templates/deployment.yaml
grep "dnsPolicy" charts/<service>/templates/deployment.yaml

# Template validation
helmfile -f helmfile.yaml -l name=<service>-<label> template

# Find service in helmfile
grep -n "name: <service>-{{ .Values.devstack_label }}" helmfile.yaml
```
