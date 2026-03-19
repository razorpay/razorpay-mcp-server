# Error Patterns Reference

Comprehensive catalog of common errors, their causes, and solutions.

## Database Errors

### Connection Refused
**Pattern**: `dial tcp <ip>:<port>: connect: connection refused`

**Cause**: Database service not running or not accessible

**Solution**:
1. Check if database service exists
2. Verify database pods are running
3. Check service endpoints
4. Verify network policies allow connection

**Commands**:
```bash
kubectl get svc <db-service> -n <db-namespace>
kubectl get pods -n <db-namespace> -l app=<db-app>
kubectl get endpoints <db-service> -n <db-namespace>
```

### Authentication Failed
**Pattern**: `authentication failed for user`

**Cause**: Wrong database credentials

**Solution**:
1. Verify secret contains correct credentials
2. Check if secret is mounted correctly
3. Verify environment variables reference correct secret keys

**Commands**:
```bash
kubectl get secret <secret-name> -n <namespace> -o yaml
kubectl describe pod <pod-name> -n <namespace> | grep -A 10 Environment
```

### Unknown Database
**Pattern**: `unknown database '<database-name>'`

**Cause**: Database doesn't exist

**Solution**:
1. Create database if missing
2. Verify database name in configuration
3. Check if migrations have run

### Too Many Connections
**Pattern**: `too many connections`

**Cause**: Connection pool exhausted or not closed properly

**Solution**:
1. Review connection pool configuration
2. Check for connection leaks in code
3. Increase max connections (temporary)
4. Fix code to properly close connections

## Application Errors

### Panic
**Pattern**: `panic: <error message>`

**Cause**: Unhandled runtime error in application

**Examples**:
```
panic: runtime error: invalid memory address or nil pointer dereference
panic: runtime error: index out of range
panic: interface conversion: interface {} is nil
```

**Solution**:
- Review stack trace for exact location
- Fix application code
- Add nil checks
- Add bounds checking
- Build new image and redeploy

### Fatal Error
**Pattern**: `fatal error: <error message>`

**Cause**: Critical application error

**Examples**:
```
fatal error: concurrent map writes
fatal error: all goroutines are asleep - deadlock!
```

**Solution**:
- Fix concurrency issues
- Add proper synchronization
- Review goroutine usage

### Segmentation Violation
**Pattern**: `segmentation violation` or `SIGSEGV`

**Cause**: Invalid memory access

**Solution**:
- Check C/Go interop if using CGO
- Review unsafe code usage
- Check for race conditions
- Fix memory access patterns

### Port Already in Use
**Pattern**: `bind: address already in use`

**Cause**: Port conflict

**Solution**:
1. Check if multiple containers using same port
2. Verify port configuration
3. Check for lingering processes
4. Ensure one process per port

**Check**:
```yaml
# deployment.yaml
containers:
- name: web
  ports:
  - containerPort: 8080  # Should be unique
- name: worker
  ports:
  - containerPort: 8081  # Different port
```

## Configuration Errors

### Environment Variable Not Set
**Pattern**: `environment variable <VAR_NAME> not set` or `<VAR_NAME> is required`

**Cause**: Missing required environment variable

**Solution**:
Add to values.yaml or secret:
```yaml
# values.yaml
env:
  - name: REDIS_HOST
    value: "redis-base.redis.svc.cluster.local"

# Or from secret
envFrom:
  - secretRef:
      name: {{ .Values.secret_name }}
```

### File Not Found
**Pattern**: `no such file or directory: <path>`

**Cause**: Missing file or incorrect path

**Solution**:
1. Check if file should be in container image
2. Verify volume mounts
3. Check ConfigMap or Secret mounts
4. Verify working directory

### Permission Denied
**Pattern**: `permission denied`

**Cause**: Insufficient file or resource permissions

**Solution**:
1. Check securityContext runAsUser
2. Verify file permissions in image
3. Check volume mount permissions
4. Review pod security policies

## Resource Errors

### OOMKilled
**Pattern**: Pod status shows `OOMKilled` in events

**Cause**: Container exceeded memory limit

**Solution**:
**AUTO-FIX**: Increase memory limits
```yaml
# Increase limits in values.yaml
web_limits_memory: 200Mi  # Double current value
```

**Commands**:
```bash
# Check current limits
kubectl describe pod <pod-name> -n <namespace> | grep -A 5 Limits

# Check previous logs for memory usage
kubectl logs <pod-name> -n <namespace> --previous --tail=100
```

### Insufficient CPU
**Pattern**: `Insufficient cpu`

**Cause**: Not enough CPU available on nodes

**Solution**:
1. Reduce CPU requests
2. Add more nodes
3. Remove CPU limits (allows bursting)

### Insufficient Memory
**Pattern**: `Insufficient memory`

**Cause**: Not enough memory available on nodes

**Solution**:
1. Reduce memory requests
2. Add more nodes
3. Check if other pods can be evicted

## Network Errors

### Connection Timeout
**Pattern**: `dial tcp <ip>:<port>: i/o timeout`

**Cause**: Network connectivity issue or service too slow

**Solution**:
1. Check if service is responding
2. Verify network policies
3. Check firewall rules
4. Increase timeout if service is slow

### DNS Lookup Failed
**Pattern**: `dial tcp: lookup <service>: no such host`

**Cause**: DNS resolution failure

**Solution**:
1. Verify service name is correct
2. Check if service exists
3. Verify DNS configuration
4. Check dnsPolicy setting

**Fix DNS**:
```yaml
dnsPolicy: ClusterFirst
dnsConfig:
  options:
  - name: ndots
    value: "1"
```

### TLS Handshake Timeout
**Pattern**: `net/http: TLS handshake timeout`

**Cause**: TLS connection issue

**Solution**:
1. Verify certificates
2. Check TLS configuration
3. Verify service supports TLS
4. Increase timeout

## Kubernetes Errors

### RBAC Permission Denied
**Pattern**: `forbidden: User "<user@email.com>" cannot list resource "secrets" in API group "" in the namespace "<namespace>"`

**Cause**: User lacks RBAC permissions for the dev-serve cluster

**Solution**:
**AUTO-ACTION**: Guide user through cluster access provisioning

1. Verify current cluster context:
   ```bash
   kubectl config current-context
   ```
   - Should be: `dev-serve` or similar

2. Run the cluster access pipeline to provision access:
   - **Pipeline Name**: Cluster access
   - **Link**: https://deploy.razorpay.com/#/applications/devserve-infra/executions
   - Click "Start Manual Execution" and run the pipeline

3. Alternative: Complete the onboarding flow:
   - If pipeline access is not available, ask team lead to run onboarding for your user
   - This will provision all necessary RBAC permissions

4. After pipeline completes, verify access:
   ```bash
   kubectl get secrets -n app
   kubectl get pods -n <namespace>
   ```

**Common Scenarios**:
```
Error: forbidden: User "dev@razorpay.com" cannot list resource "secrets"

Root Cause: RBAC permissions not provisioned for dev-serve cluster

Resolution Steps:
1. ✅ Confirmed current cluster: dev-serve
2. ⏳ User needs to run cluster access pipeline
3. 📋 Pipeline: https://deploy.razorpay.com/#/applications/devserve-infra/executions
4. ⏱️ Wait for pipeline to complete (~2-5 minutes)
5. ✅ Verify access with: kubectl get secrets -n app

If pipeline fails or access is still denied:
- Check if your user is in the correct LDAP/AD group
- Contact DevOps team for manual RBAC provisioning
- Verify cluster context is correct
```

**Validation Commands**:
```bash
# Check current context
kubectl config current-context

# List available contexts
kubectl config get-contexts

# Switch to dev-serve if needed
kubectl config use-context dev-serve

# Test permissions after pipeline
kubectl auth can-i list secrets -n app
kubectl auth can-i get pods -n <namespace>
```

### FailedScheduling
**Pattern**: `0/N nodes are available: <reason>`

**Reasons**:
- `Insufficient memory` → Need more memory or nodes
- `Insufficient cpu` → Need more CPU or nodes
- `node(s) didn't match Pod's node affinity` → Check nodeSelector

**Solution**:
```bash
# Check node resources
kubectl describe nodes

# Check node labels
kubectl get nodes --show-labels

# Verify pod nodeSelector
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 5 nodeSelector
```

### ImagePullBackOff
**Pattern**: `Failed to pull image` or `ErrImagePull`

**Cause**: Cannot pull container image

**Solution**:
1. Verify image tag is correct and exists
2. Check registry authentication
3. Verify image name format
4. Check network connectivity to registry

**Check**:
```bash
# Verify image in helmfile
grep -A 3 "name: <service>-{{ .Values.devstack_label }}" helmfile.yaml | grep image

# Check pod events
kubectl describe pod <pod-name> -n <namespace> | grep -A 10 Events
```

### FailedMount
**Pattern**: `Unable to mount volumes` or `PersistentVolumeClaim "<name>" not found`

**Cause**: Volume mount issue

**Solution**:
1. Check if PVC exists
2. Verify volume name matches
3. Check PVC status
4. Verify storage class

### Liveness/Readiness Probe Failed
**Pattern**: `Liveness probe failed` or `Readiness probe failed`

**Cause**: Probe check failing

**Solution**:
1. Check if health endpoint exists
2. Verify endpoint returns 200 OK
3. Check if port is correct
4. Increase initialDelaySeconds if app starts slowly
5. Review application logs

**Debug**:
```bash
# Check probe config
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 10 livenessProbe

# Test endpoint manually
kubectl exec <pod-name> -n <namespace> -- wget -O- http://localhost:8080/health
```

## Helm/Helmfile Errors

### Release Failed
**Pattern**: `Error: release <name> failed`

**Cause**: Helm operation failed

**Solution**:
1. Check Helm output for specific error
2. Verify YAML syntax
3. Check template rendering
4. Review values
5. Validate helm release status

**Debug Commands**:
```bash
# Check release status
helm status <service>-<label> -n <namespace>

# Get release history
helm history <service>-<label> -n <namespace>

# Check deployed manifest
helm get manifest <service>-<label> -n <namespace>

# Review values being used
helm get values <service>-<label> -n <namespace>
```

### Template Rendering Error
**Pattern**: `template: <file>:<line>: executing "<template>" at <field>: map has no entry for key "<key>"`

**Cause**: Missing value in values.yaml

**Solution**:
Add missing value to values.yaml:
```yaml
<key>: <value>
```

**Debug**:
```bash
# Render templates to see exact error
cd /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile
helmfile -f helmfile.yaml -l name=<service>-<label> template

# Check if value exists in values.yaml
grep "<key>" charts/<service>/values.yaml

# Validate chart syntax
helm lint charts/<service>/
```

### YAML Syntax Error
**Pattern**: `yaml: line <N>: mapping values are not allowed in this context`

**Cause**: Invalid YAML syntax

**Solution**:
1. Check indentation at specified line
2. Verify colons and spacing
3. Check for tabs (use spaces)
4. Validate with YAML linter

**Debug**:
```bash
# Check specific template file syntax
cat charts/<service>/templates/<file>.yaml | head -n <N+5> | tail -n 10

# Validate entire chart
helm lint charts/<service>/
```

### Chart Values Not Applied
**Pattern**: Deployed resources don't reflect changes in values.yaml

**Cause**: Values not being passed correctly from helmfile to chart

**Solution**:
1. Verify helmfile.yaml passes values correctly
2. Check helm release is using correct values
3. Ensure helmfile sync was successful

**Debug**:
```bash
# Check what values helm is using
helm get values <service>-<label> -n <namespace>

# Compare with values.yaml
diff <(helm get values <service>-<label> -n <namespace>) \
     charts/<service>/values.yaml

# Check helmfile.yaml configuration
grep -A 10 "name: <service>-{{ .Values.devstack_label }}" helmfile.yaml

# Force re-sync
helmfile -f helmfile.yaml -l name=<service>-<label> sync --force
```

### Helm Release in FAILED State
**Pattern**: `helm list` shows status as FAILED

**Cause**: Previous deployment failed, release stuck in failed state

**Solution**:
1. Review what caused the failure
2. Fix the underlying issue
3. Redeploy to update release

**Debug**:
```bash
# Check failure reason
helm status <service>-<label> -n <namespace>

# See what was attempted in failed revision
helm get manifest <service>-<label> -n <namespace> --revision <N>

# Check all revisions
helm history <service>-<label> -n <namespace>

# After fixing issue, redeploy
helmfile -f helmfile.yaml -l name=<service>-<label> sync
```

### Configuration Drift
**Pattern**: Deployed resources differ from chart templates

**Cause**: Manual kubectl edits or values not syncing

**Solution**:
1. Identify differences
2. Update chart to match desired state
3. Redeploy via helmfile

**Debug**:
```bash
# Compare deployed vs chart template
kubectl get deployment <deployment> -n <namespace> -o yaml > /tmp/deployed.yaml
helmfile -f helmfile.yaml -l name=<service>-<label> template | grep -A 100 "kind: Deployment" > /tmp/chart.yaml
diff /tmp/deployed.yaml /tmp/chart.yaml

# If drift detected, redeploy from chart
helmfile -f helmfile.yaml -l name=<service>-<label> sync
```

## Error Pattern Quick Reference

| Error Message | Likely Cause | Quick Fix |
|--------------|--------------|-----------|
| `connection refused` | Service not running | Start dependency service |
| `authentication failed` | Wrong credentials | Update secret |
| `panic:` | Application bug | Fix code, rebuild |
| `address already in use` | Port conflict | Change port |
| `not set` (env var) | Missing config | Add environment variable |
| `no such file` | Missing file | Add to image or mount |
| `permission denied` | Wrong permissions | Fix securityContext |
| `OOMKilled` | Out of memory | Increase memory limit (auto) |
| `Insufficient cpu/memory` | Not enough resources | Reduce requests or add nodes |
| `lookup ... no such host` | DNS issue | Fix service name or DNS config |
| `TLS handshake timeout` | Certificate issue | Fix certificates |
| `ImagePullBackOff` | Image not found | Fix image tag |
| `FailedMount` | Volume issue | Create PVC |
| `probe failed` | Health check failing | Fix endpoint or config |
| `forbidden: User ... cannot` | RBAC permissions missing | Run cluster access pipeline |
| `release failed` | Helm operation failed | Check helm status and logs |
| `template: ... map has no entry` | Missing value | Add to values.yaml |
| `yaml: line N:` | YAML syntax error | Fix indentation/syntax |
| `values not applied` | Helmfile sync issue | Check helm get values |
| `configuration drift` | Manual changes | Redeploy from chart |

## Debugging Commands by Error Type

```bash
# Database errors
kubectl logs <pod> -n <ns> | grep -i "database\|mysql\|postgres\|connection"

# Application panics
kubectl logs <pod> -n <ns> | grep -i "panic\|fatal"

# Configuration errors
kubectl logs <pod> -n <ns> | grep -i "environment\|config\|missing"

# Network errors
kubectl logs <pod> -n <ns> | grep -i "timeout\|connection\|dial\|lookup"

# RBAC/Permission errors
kubectl auth can-i list secrets -n <ns>
kubectl auth can-i get pods -n <ns>
kubectl config current-context

# Resource errors
kubectl describe pod <pod> -n <ns> | grep -i "OOM\|Insufficient"

# All errors
kubectl logs <pod> -n <ns> | grep -iE "error|exception|fatal|panic|fail"

# Helm/Chart errors
cd /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile
helmfile -f helmfile.yaml -l name=<service>-<label> template 2>&1 | grep -i "error"
helm status <service>-<label> -n <ns>
helm get values <service>-<label> -n <ns>
helm lint charts/<service>/

# Configuration drift detection
diff <(kubectl get deployment <dep> -n <ns> -o yaml) \
     <(helmfile -f helmfile.yaml -l name=<service>-<label> template | grep -A 100 "kind: Deployment")
```
