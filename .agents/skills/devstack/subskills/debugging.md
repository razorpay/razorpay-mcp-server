# Debugging Subskill

Autonomous troubleshooting and root cause analysis for failing Kubernetes pods.

## Quick Reference

**For common issues, check these resources first:**
- 📚 [FAQ](../references/faq.md) - Frequently asked questions (TTL, external access, NAT IPs, etc.)
- 🔍 [Error Patterns](../references/error-patterns.md) - Common errors and solutions
- 🔧 [Auto-Fix Strategies](../references/auto-fix-strategies.md) - What gets auto-fixed

## Purpose

Provides intelligent debugging for pod failures with automatic log analysis, event inspection, and root cause identification.

## Prerequisites

- Helmfile directory configured (see [../SKILL.md#configuration](../SKILL.md#configuration))
- kubectl access to devstack cluster
- Path detection will run automatically (see [../references/path-detection.md](../references/path-detection.md))

**Note**: When debugging requires access to chart files, the skill will use the helmfile directory path from `config.json`. All path references like `<HELMFILE_DIR>` below refer to the configured path.

## When to Use

- Pods in CrashLoopBackOff state
- ImagePullBackOff errors
- OOMKilled pods
- Pending pods that won't schedule
- Running pods that aren't ready
- Any pod errors or failures

## Pod Status Analysis

### ImagePullBackOff

**Root Cause**: Cannot pull container image from registry

**Automatic Debug Actions**:
1. Check image tag in helmfile.yaml
   ```bash
   grep -A 5 "name: <service>-{{ .Values.devstack_label }}" helmfile.yaml | grep image
   ```

2. Verify image format
   - Expected: `c.rzp.io/razorpay/<service>:<commit-hash>`
   - Check if commit hash is valid

3. Get pod description for detailed error
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   ```

**Common Issues**:
- Empty image tag → Request valid image from user
- Typo in image name → Correct and redeploy
- Non-existent commit hash → Verify commit exists
- Registry authentication → Check registry access (rare)

**Report Format**:
```
❌ Issue: ImagePullBackOff

Root Cause: Image not found in registry
Evidence:
  Failed to pull image "c.rzp.io/razorpay/payment-service:"
  rpc error: code = Unknown desc = Error response: manifest unknown

Analysis:
  - Image tag is empty in helmfile.yaml:892
  - Service deployment requires valid commit hash

Fix Needed:
  1. Get valid commit hash from CI/CD or git
  2. Update helmfile.yaml:
     image: <valid-commit-hash>
  3. Redeploy: helmfile -f helmfile.yaml -l name=<service>-<label> sync
```

### CrashLoopBackOff

**Root Cause**: Container keeps crashing after starting

**Automatic Debug Actions**:

1. Get pod name
   ```bash
   kubectl get pods -n <namespace> -l devstack_label=<label> -o jsonpath='{.items[0].metadata.name}'
   ```

2. Check current logs
   ```bash
   kubectl logs <pod-name> -n <namespace> --tail=50
   ```

3. Check previous container logs (critical for crash analysis)
   ```bash
   kubectl logs <pod-name> -n <namespace> --previous --tail=50
   ```

4. Check pod events
   ```bash
   kubectl describe pod <pod-name> -n <namespace> | grep -A 20 Events
   ```

5. Analyze logs for error patterns (see [../references/error-patterns.md](../references/error-patterns.md)):
   - **Database errors**: `connection refused`, `authentication failed`
   - **Config errors**: `environment variable not set`, `missing config`
   - **Port conflicts**: `bind: address already in use`
   - **Application crashes**: `panic:`, `fatal error:`, `segmentation violation`

**Error Pattern Examples**:

#### Database Connection Error
```
Root Cause: Cannot connect to database

Evidence from logs:
  Error: dial tcp 172.20.1.50:3306: connect: connection refused
  Failed to connect to MySQL at mysql-base:3306

Analysis:
  - Application trying to connect to mysql-base service
  - Connection refused indicates service not running or not accessible

Fix Needed:
  1. Check if MySQL service exists:
     kubectl get svc mysql-base -n database

  2. If missing, deploy MySQL first:
     helmfile -f helmfile.yaml -l name=mysql-base sync

  3. If exists, check MySQL pods:
     kubectl get pods -n database -l app=mysql

  4. After MySQL is running, restart this service:
     kubectl delete pods -n <namespace> -l devstack_label=<label>
```

#### Missing Environment Variable
```
Root Cause: Required environment variable not set

Evidence from logs:
  panic: REDIS_HOST environment variable is not set

Analysis:
  - Application requires REDIS_HOST environment variable
  - Variable not defined in values.yaml or secret

Fix Needed:
  1. Add to values.yaml:
     env:
       - name: REDIS_HOST
         value: "redis-base.redis.svc.cluster.local"

  2. Or reference from secret:
     envFrom:
       - secretRef:
           name: {{ .Values.secret_name }}

  3. Redeploy after adding configuration
```

#### Port Already in Use
```
Root Cause: Port conflict

Evidence from logs:
  Error: listen tcp :8080: bind: address already in use

Analysis:
  - Application trying to bind to port 8080
  - Port likely already used by another container/process
  - May indicate multiple containers in pod with same port

Fix Needed:
  Check deployment.yaml for:
  1. Multiple containers using same port
  2. Incorrect containerPort configuration
  3. Liveness probe on wrong port
```

#### Application Panic/Crash
```
Root Cause: Application bug causing crash

Evidence from logs:
  panic: runtime error: invalid memory address or nil pointer dereference
  goroutine 1 [running]:
  main.processPayment(...)
      /app/payment/processor.go:145
  main.main()
      /app/main.go:42 +0x123

Analysis:
  - Application code has nil pointer dereference
  - Crash at payment/processor.go:145
  - This is an application-level bug

Fix Needed:
  This requires code fix:
  1. Review payment/processor.go:145
  2. Add nil checks before dereferencing
  3. Build new image with fix
  4. Deploy with updated image tag

  Cannot be auto-fixed (application bug)
```

### Pending

**Root Cause**: Pod cannot be scheduled to any node

**Automatic Debug Actions**:

1. Describe pod to see scheduling errors
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   ```

2. Check events for specific failure reason
   ```bash
   kubectl get events -n <namespace> --field-selector involvedObject.name=<pod-name> --sort-by='.lastTimestamp'
   ```

**Common Causes**:

#### Insufficient Resources
```
Event: FailedScheduling
Message: 0/5 nodes are available: 5 Insufficient memory

Root Cause: Not enough memory on any node

Fix Needed:
  Option 1: Reduce memory request in values.yaml
    web_requests_memory: 50Mi  # Reduce from current value

  Option 2: Increase node capacity (requires infra team)
```

#### Node Selector Mismatch
```
Event: FailedScheduling
Message: 0/5 nodes are available: 5 node(s) didn't match Pod's node affinity/selector

Root Cause: No nodes match the nodeSelector

Fix Needed:
  Check nodeSelector in values.yaml:
    node_selector:
      environment: devstack  # Verify this label exists on nodes

  Verify nodes have correct labels:
    kubectl get nodes --show-labels | grep environment=devstack
```

#### PVC Pending
```
Event: FailedMount
Message: PersistentVolumeClaim "data-volume" not found

Root Cause: Referenced PVC doesn't exist

Fix Needed:
  1. Check if PVC is defined in templates
  2. Create PVC if missing
  3. Or remove volume mount if not needed
```

### OOMKilled

**Root Cause**: Container exceeded memory limit and was killed

**Automatic Debug Actions**:

1. Check current memory limits in values.yaml
   ```bash
   grep -E "web_limits_memory|worker_limits_memory" charts/<service>/values.yaml
   ```

2. Check previous logs for memory usage patterns
   ```bash
   kubectl logs <pod-name> -n <namespace> --previous --tail=100
   ```

3. **AUTOMATICALLY increase memory limits** based on current value:
   - Current < 100Mi → Increase to 200Mi
   - Current < 200Mi → Increase to 512Mi
   - Current < 512Mi → Increase to 1Gi

4. Apply fix to values.yaml

5. Redeploy with new limits

6. Monitor for 2 minutes

**Auto-Fix Example**:
```
⚠️ OOMKilled Detected - Auto-Fixing

Current Configuration:
  web_limits_memory: 100Mi

Auto-Fix Applied:
  web_limits_memory: 200Mi (increased from 100Mi)

Actions Taken:
  1. ✅ Updated values.yaml:23
  2. ✅ Redeploying with new limits
  3. ⏳ Waiting for pods to restart...
  4. ✅ Pods running with increased memory

Result: Issue resolved
```

**If Still OOMKilled After 3 Increases**:
```
❌ Persistent OOM Issue

Attempted Fixes:
  - Increased to 200Mi → Still OOMKilled
  - Increased to 512Mi → Still OOMKilled
  - Increased to 1Gi → Still OOMKilled

Analysis:
  Application has memory leak or requires > 1Gi

Recommended Actions:
  1. Profile application memory usage
  2. Check for memory leaks in code
  3. Review large object allocations
  4. Consider if application is appropriate for this environment

  Further increases may not be sustainable.
```

### Error

**Root Cause**: Pod terminated with error status

**Automatic Debug Actions**:

1. Get full pod details
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   ```

2. Check all container logs
   ```bash
   kubectl logs <pod-name> -n <namespace> --all-containers=true
   ```

3. Check events for error details

4. Analyze exit code:
   - Exit code 1: General application error
   - Exit code 137: Killed (OOMKilled or manual)
   - Exit code 139: Segmentation fault
   - Exit code 143: Graceful termination (SIGTERM)

### CreateContainerConfigError

**Root Cause**: Invalid container configuration

**Automatic Debug Actions**:

1. Check pod events
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   ```

**Common Causes**:

#### Missing Secret
```
Error: couldn't find key MYSQL_PASSWORD in Secret <namespace>/<secret-name>

Root Cause: Secret doesn't exist or missing key

Fix Needed:
  1. Check if secret exists:
     kubectl get secret <secret-name> -n <namespace>

  2. If missing, create secret:
     kubectl create secret generic <secret-name> \
       --from-literal=MYSQL_PASSWORD=<password> \
       -n <namespace>

  3. If exists, verify it has required keys:
     kubectl get secret <secret-name> -n <namespace> -o yaml
```

#### Invalid ConfigMap
```
Error: ConfigMap "<config-name>" not found

Root Cause: Referenced ConfigMap doesn't exist

Fix Needed:
  1. Create ConfigMap with required data
  2. Or remove ConfigMap reference if not needed
```

### RBAC Permission Denied

**Root Cause**: User lacks permissions to access cluster resources

**Error Pattern**:
```
Error from server (Forbidden): secrets is forbidden: User "dev@razorpay.com" cannot list resource "secrets" in API group "" in the namespace "app"
```

**Automatic Debug Actions**:

1. Verify current cluster context
   ```bash
   kubectl config current-context
   ```

2. Check if context is `dev-serve` or similar

3. **Guide user to provision access**:
   - Run cluster access pipeline at https://deploy.razorpay.com/#/applications/devserve-infra/executions
   - Pipeline Name: "Cluster access"
   - OR: Ask team lead to complete onboarding flow

4. Wait for pipeline completion (2-5 minutes)

5. Verify permissions after provisioning
   ```bash
   kubectl auth can-i list secrets -n app
   kubectl auth can-i get pods -n <namespace>
   ```

**Report Format**:
```
❌ Issue: RBAC Permission Denied

Root Cause: User lacks RBAC permissions for dev-serve cluster
Evidence:
  forbidden: User "dev@razorpay.com" cannot list resource "secrets"
  in API group "" in the namespace "app"

Analysis:
  - Current cluster: dev-serve
  - User permissions not provisioned
  - RBAC access required for cluster operations

Solution Required:
  1. Run cluster access provisioning pipeline:
     Pipeline: Cluster access
     Link: https://deploy.razorpay.com/#/applications/devserve-infra/executions

  2. Click "Start Manual Execution" and run the pipeline

  3. Wait for completion (~2-5 minutes)

  4. Verify access:
     kubectl get secrets -n app
     kubectl get pods -n <namespace>

Alternative:
  - If pipeline access unavailable, request onboarding from team lead
  - Contact DevOps for manual RBAC provisioning

Next Steps After Access Granted:
  - Retry the original command
  - Continue with deployment/debugging workflow
```

**Validation**:
```bash
# Verify cluster context
kubectl config current-context
# Expected: dev-serve

# Test specific permissions
kubectl auth can-i list secrets -n app
kubectl auth can-i get pods --all-namespaces
kubectl auth can-i create deployments -n <namespace>

# If all return "yes", permissions are correctly provisioned
```

### Running but Not Ready (0/1)

**Root Cause**: Readiness probe failing

**Automatic Debug Actions**:

1. Check logs for application startup
   ```bash
   kubectl logs <pod-name> -n <namespace> --tail=100
   ```

2. Check readiness probe configuration in deployment.yaml

3. Verify application is listening on correct port

4. Check if health endpoint is responding

**Common Issues**:

#### Wrong Port
```
Readiness probe failed: dial tcp 10.244.1.5:8080: connect: connection refused

Root Cause: App not listening on probe port

Check:
  1. Application logs for actual port:
     Starting server on port 3000

  2. Update readiness probe port to match:
     readinessProbe:
       httpGet:
         port: 3000  # Change from 8080
```

#### Slow Startup
```
Readiness probe failed: timeout

Root Cause: App takes longer to start than probe allows

Fix:
  Increase initialDelaySeconds:
    readinessProbe:
      httpGet:
        path: /ready
        port: 8080
      initialDelaySeconds: 30  # Increase from 10
      timeoutSeconds: 5
```

## Debugging Workflow

### Step-by-Step Debug Process

1. **Get Pod Status**
   ```bash
   kubectl get pods -n <namespace> -l devstack_label=<label>
   ```

2. **Identify Pod State**
   - Running → Check readiness
   - Pending → Check scheduling
   - CrashLoopBackOff → Check logs
   - ImagePullBackOff → Check image
   - Error → Check exit code
   - OOMKilled → Check memory

3. **Gather Evidence**
   - Pod description (events)
   - Current logs
   - Previous logs (if crashed)
   - Resource usage
   - **Helm chart configuration** (see Helm Chart Validation below)

4. **Analyze Root Cause**
   - Match error patterns
   - Identify specific failure point
   - Determine if auto-fixable

5. **Apply Fix or Report**
   - Auto-fix if possible
   - Provide detailed steps if manual fix needed
   - Include all relevant commands

## Helm Chart Validation During Debugging

When debugging pod failures, always validate the helm chart configuration to ensure deployment specs match expectations.

### Why Check Helm Charts

Many pod failures stem from:
- Misconfigured templates
- Incorrect values in values.yaml
- Template rendering issues
- Mismatch between chart and deployed resources

### Helm Chart Debug Actions

#### 1. Verify Chart Files Exist

```bash
# Check chart structure
ls -la /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile/charts/<service>/

# Expected structure:
# - Chart.yaml
# - values.yaml
# - templates/
#   - deployment.yaml
#   - service.yaml
#   - configmap.yaml (if any)
```

#### 2. Validate Current values.yaml Configuration

```bash
# Read current values
cat /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile/charts/<service>/values.yaml
```

**Check for**:
- Resource limits match pod requirements
- Image tag is correct
- Environment variables are set
- Secret references are valid
- Node selectors are appropriate

#### 3. Render and Validate Templates

```bash
# Render templates to see what will be deployed
cd /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile
helmfile -f helmfile.yaml -l name=<service>-<label> template
```

**This reveals**:
- Template rendering errors
- Missing variable substitutions
- Invalid YAML syntax
- Incorrect resource specifications

**Common Template Issues**:

```
❌ Template Error: Missing Variable
Error: template: deployment.yaml:23:15: executing "deployment.yaml"
at <.Values.db_host>: map has no entry for key "db_host"

Root Cause: values.yaml missing required variable

Fix: Add to values.yaml:
  db_host: mysql-base.database.svc.cluster.local
```

#### 4. Compare Deployed Resources with Chart

```bash
# Get currently deployed spec
kubectl get deployment <deployment-name> -n <namespace> -o yaml > /tmp/deployed.yaml

# Compare with rendered template
helmfile -f helmfile.yaml -l name=<service>-<label> template | grep -A 50 "kind: Deployment" > /tmp/chart-template.yaml

# Look for differences
diff /tmp/deployed.yaml /tmp/chart-template.yaml
```

**This identifies**:
- Configuration drift
- Manual changes not in chart
- Values not being applied correctly

#### 5. Validate Helm Release Status

```bash
# Check helm release
helm list -n <namespace> | grep <service>-<label>

# Get release details
helm get values <service>-<label> -n <namespace>

# Check revision history
helm history <service>-<label> -n <namespace>
```

**Common Issues**:

```
Status: FAILED
Last Deployment: failed
Revision: 3

Actions:
1. Check failed revision details:
   helm get manifest <service>-<label> -n <namespace> --revision 3

2. Review failure reason:
   helm status <service>-<label> -n <namespace>

3. Consider rollback if previous version worked:
   helmfile -f helmfile.yaml -l name=<service>-<label> sync
```

#### 6. Check Chart Template Syntax

```bash
# Validate chart templates locally
helm lint /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile/charts/<service>/
```

**This catches**:
- YAML syntax errors
- Invalid Kubernetes resource specs
- Missing required fields
- Template logic errors

### Helm Chart Debug Checklist

When debugging, systematically check:

- [ ] Chart files exist and are readable
- [ ] values.yaml has all required fields
- [ ] Template rendering succeeds without errors
- [ ] Rendered templates produce valid Kubernetes resources
- [ ] Deployed resources match chart templates
- [ ] Helm release is in DEPLOYED status (not FAILED)
- [ ] Resource limits/requests in chart match pod requirements
- [ ] Image tag in values.yaml matches expected version
- [ ] Environment variables in chart match application needs
- [ ] Probes configuration is appropriate for application
- [ ] Service and endpoint configurations are correct

### Example: Debugging with Helm Chart Validation

```
User: Pods are failing for payment-service with label john

Debug Process:

1. ✅ Check pod status
   kubectl get pods -n payment-service -l devstack_label=john
   Status: CrashLoopBackOff

2. ✅ Get logs
   kubectl logs payment-service-john-xyz -n payment-service
   Error: Database connection refused

3. ✅ Validate Helm Chart Configuration
   cd /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile

   # Check values.yaml
   cat charts/payment-service/values.yaml | grep -i db
   Result: db_host not set!

4. ✅ Render template to confirm
   helmfile -f helmfile.yaml -l name=payment-service-john template | grep -A 5 "DB_HOST"
   Result: DB_HOST environment variable is empty

5. ❌ Root Cause Identified
   values.yaml missing db_host configuration

   Fix Applied:
   - Add db_host: mysql-base.database.svc.cluster.local to values.yaml
   - Redeploy with helmfile sync

6. ✅ Verify fix
   kubectl logs payment-service-john-abc -n payment-service
   Result: Application started successfully
```

### Helm Chart Error Patterns

#### Template Rendering Failures

**Pattern**: `Error: template: <file>:<line>: executing`

**Debug**:
```bash
# Identify problematic template
helmfile -f helmfile.yaml -l name=<service>-<label> template 2>&1 | grep "Error:"

# Check the specific template file
cat charts/<service>/templates/<template-file>.yaml | sed -n '<line>p'

# Verify value exists
grep "<missing-key>" charts/<service>/values.yaml
```

#### Values Not Applied

**Pattern**: Deployed resources don't reflect values.yaml changes

**Debug**:
```bash
# Check if helmfile.yaml passes values correctly
grep -A 10 "name: <service>-{{ .Values.devstack_label }}" helmfile.yaml

# Verify values are being used
helm get values <service>-<label> -n <namespace>

# Compare with values.yaml
diff <(helm get values <service>-<label> -n <namespace>) charts/<service>/values.yaml
```

#### Release Failed

**Pattern**: Helm release status is FAILED

**Debug**:
```bash
# Get failure details
helm status <service>-<label> -n <namespace>

# Check what was attempted
helm get manifest <service>-<label> -n <namespace>

# Review all revisions
helm history <service>-<label> -n <namespace>

# If needed, force redeploy
helmfile -f helmfile.yaml -l name=<service>-<label> sync --force
```

## Multi-Pod Debugging

When multiple pods are failing:

1. **Check all pod statuses**
   ```bash
   kubectl get pods -n <namespace> -l devstack_label=<label>
   ```

2. **Identify patterns**
   - All pods same error → Common issue (config, image)
   - Different errors → Check each individually

3. **Prioritize debugging**
   - Fix common issues first
   - Address unique issues individually

4. **Report all findings**
   ```
   Multiple Pod Failures Detected:

   Common Issue (3/3 pods):
     Status: CrashLoopBackOff
     Root Cause: Database connection failure
     Fix: Deploy MySQL service

   Individual Issues:
     - worker-pod-1: OOMKilled (auto-fixed, redeploying)
   ```

## Commands Reference

### Essential Debug Commands

```bash
# Get pod status
kubectl get pods -n <namespace> -l devstack_label=<label>

# Detailed pod info
kubectl describe pod <pod-name> -n <namespace>

# Current logs
kubectl logs <pod-name> -n <namespace> --tail=100

# Previous container logs
kubectl logs <pod-name> -n <namespace> --previous --tail=100

# Real-time logs
kubectl logs <pod-name> -n <namespace> -f

# Pod events only
kubectl get events -n <namespace> --field-selector involvedObject.name=<pod-name>

# Resource usage
kubectl top pod <pod-name> -n <namespace>

# Execute command in pod (if running)
kubectl exec -it <pod-name> -n <namespace> -- /bin/sh

# Check RBAC permissions
kubectl auth can-i list secrets -n <namespace>
kubectl auth can-i get pods -n <namespace>
kubectl auth can-i create deployments -n <namespace>

# Verify cluster context
kubectl config current-context
kubectl config get-contexts
```

### Helm Chart Debug Commands

```bash
# List helm releases
helm list -n <namespace>

# Check release status
helm status <service>-<label> -n <namespace>

# Get release values
helm get values <service>-<label> -n <namespace>

# Get deployed manifest
helm get manifest <service>-<label> -n <namespace>

# Check release history
helm history <service>-<label> -n <namespace>

# Render helmfile templates
cd /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile
helmfile -f helmfile.yaml -l name=<service>-<label> template

# Validate chart syntax
helm lint /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile/charts/<service>/

# Check values.yaml
cat /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile/charts/<service>/values.yaml

# Check deployment template
cat /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile/charts/<service>/templates/deployment.yaml

# Compare deployed vs chart
diff <(kubectl get deployment <deployment> -n <namespace> -o yaml) \
     <(helmfile -f helmfile.yaml -l name=<service>-<label> template | grep -A 100 "kind: Deployment")
```

## Related Subskills

- [Deployment](deployment.md) - Initial deployment and validation
- [Monitoring](monitoring.md) - Ongoing health checks
- [Validation](validation.md) - Configuration validation

## Additional Resources

- [Error Patterns Reference](../references/error-patterns.md) - Comprehensive error catalog
- [Recovery Workflows](../references/recovery-workflows.md) - Step-by-step recovery procedures
