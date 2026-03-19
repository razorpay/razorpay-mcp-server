# Recovery Workflows

Step-by-step procedures for recovering from common failure scenarios.

## OOMKilled Recovery Workflow

### Symptoms
- Pod status shows `OOMKilled` in events
- Container restarts frequently
- Logs show sudden termination

### Recovery Steps

**Step 1: Verify OOMKilled Status**
```bash
kubectl describe pod <pod-name> -n <namespace> | grep -i oom
```

**Step 2: Check Current Memory Limits**
```bash
grep web_limits_memory charts/<service>/values.yaml
```

**Step 3: Auto-Increase Memory (Automatic)**
```
Current limit < 100Mi → Increase to 200Mi
Current limit < 200Mi → Increase to 512Mi
Current limit < 512Mi → Increase to 1Gi
```

**Step 4: Update Configuration**
```yaml
# values.yaml
web_limits_memory: <new-limit>
```

**Step 5: Redeploy**
```bash
helmfile -f helmfile.yaml -l name=<service>-<label> sync
```

**Step 6: Monitor (2 minutes)**
```bash
watch kubectl get pods -n <namespace> -l devstack_label=<label>
```

**Step 7: Verify Stability**
```bash
kubectl top pod -n <namespace> -l devstack_label=<label>
```

### Recovery Outcome

**Success**: Pod running and memory usage stable
**Failure**: Still OOMKilled → Repeat (max 3 times)
**Persistent Failure**: Investigate application memory leak

### Escalation
After 3 increases (reached 1Gi+):
- Profile application memory usage
- Check for memory leaks
- Consider if service appropriate for environment

## CrashLoopBackOff Recovery Workflow

### Symptoms
- Pod status shows `CrashLoopBackOff`
- Container starts then immediately crashes
- Increasing backoff delay between restarts

### Recovery Steps

**Step 1: Identify Error**
```bash
# Get current logs
kubectl logs <pod-name> -n <namespace> --tail=50

# Get previous container logs
kubectl logs <pod-name> -n <namespace> --previous --tail=100
```

**Step 2: Classify Error Type**
```
Database Error → Go to Database Recovery
Configuration Error → Go to Configuration Recovery
Application Error → Go to Application Debug
Dependency Missing → Go to Dependency Recovery
```

**Step 3: Apply Appropriate Sub-Workflow**

### Database Connection Recovery

**Symptoms**: Logs show connection refused or auth failures

**Steps**:
1. Check if database service exists
   ```bash
   kubectl get svc <db-service> -n <db-namespace>
   ```

2. Verify database pods running
   ```bash
   kubectl get pods -n <db-namespace> -l app=<db-app>
   ```

3. If missing, deploy database
   ```bash
   helmfile -f helmfile.yaml -l name=<db-service> sync
   ```

4. Wait for database to be ready (30-60s)

5. Restart application pods
   ```bash
   kubectl delete pods -n <namespace> -l devstack_label=<label>
   ```

6. Monitor restart
   ```bash
   kubectl get pods -n <namespace> -l devstack_label=<label> --watch
   ```

### Configuration Error Recovery

**Symptoms**: Logs show missing env vars or config

**Steps**:
1. Identify missing configuration from logs

2. Add to values.yaml or secret
   ```yaml
   # values.yaml
   env:
     - name: <VAR_NAME>
       value: "<value>"
   ```

3. Redeploy
   ```bash
   helmfile -f helmfile.yaml -l name=<service>-<label> sync
   ```

4. Verify pod starts successfully

### Application Error Recovery

**Symptoms**: Panic, fatal error, or segfault in logs

**Steps**:
1. Capture full stack trace
2. Identify error location
3. Report to development team:
   ```
   Application Bug Found:
   File: <file>:<line>
   Error: <error message>
   Stack trace: <trace>
   ```

4. Cannot auto-recover (requires code fix)

### Recovery Outcome

**Success**: Pod reaches Running status and stays running
**Partial**: Pod starts but encounters new error
**Failure**: Cannot start - needs code fix

## ImagePullBackOff Recovery Workflow

### Symptoms
- Pod status shows `ImagePullBackOff` or `ErrImagePull`
- Cannot pull container image
- Registry errors in events

### Recovery Steps

**Step 1: Identify Image Issue**
```bash
kubectl describe pod <pod-name> -n <namespace> | grep -A 5 "Failed to pull"
```

**Step 2: Check Image Tag**
```bash
grep -A 3 "name: <service>-{{ .Values.devstack_label }}" helmfile.yaml | grep image
```

**Step 3: Verify Issue**
```
Empty image tag → Request from user
Invalid format → Correct format
Non-existent commit → Verify commit exists
Registry access → Check credentials (rare)
```

**Step 4: Fix Image Tag**
```yaml
# helmfile.yaml
values:
  - image: <valid-commit-hash>
```

**Step 5: Redeploy**
```bash
helmfile -f helmfile.yaml -l name=<service>-<label> sync
```

**Step 6: Verify Image Pull**
```bash
kubectl get pods -n <namespace> -l devstack_label=<label>
```

### Recovery Outcome

**Success**: Image pulled, pod starts
**Failure**: Still cannot pull → Check registry access

## Pending Pod Recovery Workflow

### Symptoms
- Pod status stuck in `Pending`
- No container created
- Scheduling failures in events

### Recovery Steps

**Step 1: Identify Scheduling Issue**
```bash
kubectl describe pod <pod-name> -n <namespace> | grep -A 10 Events
```

**Step 2: Classify Issue**
```
Insufficient resources → Resource Recovery
Node selector mismatch → Selector Recovery
PVC pending → Volume Recovery
```

### Insufficient Resources Recovery

**Steps**:
1. Check resource requests
   ```bash
   kubectl describe pod <pod-name> -n <namespace> | grep -A 5 "Requests:"
   ```

2. Option A: Reduce requests
   ```yaml
   # values.yaml
   web_requests_memory: 50Mi  # Reduce
   web_requests_cpu: 50m      # Reduce
   ```

3. Option B: Wait for node capacity

4. Redeploy with adjusted resources

### Node Selector Mismatch Recovery

**Steps**:
1. Check node labels
   ```bash
   kubectl get nodes --show-labels | grep environment
   ```

2. Verify pod nodeSelector
   ```bash
   kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 3 nodeSelector
   ```

3. Fix nodeSelector in values.yaml
   ```yaml
   node_selector:
     environment: devstack  # Must match node labels
   ```

4. Redeploy

### Volume Recovery

**Steps**:
1. Check PVC status
   ```bash
   kubectl get pvc -n <namespace>
   ```

2. If PVC pending, check storage class

3. If PVC missing, create or remove from deployment

### Recovery Outcome

**Success**: Pod scheduled and starts
**Failure**: Cannot schedule - needs infrastructure change

## CreateContainerConfigError Recovery Workflow

### Symptoms
- Pod status shows `CreateContainerConfigError`
- Container cannot be created
- Configuration errors in events

### Recovery Steps

**Step 1: Identify Config Issue**
```bash
kubectl describe pod <pod-name> -n <namespace> | grep -A 10 Events
```

**Step 2: Common Issues**

### Missing Secret Recovery

**Steps**:
1. Verify secret exists
   ```bash
   kubectl get secret <secret-name> -n <namespace>
   ```

2. If missing, create secret
   ```bash
   kubectl create secret generic <secret-name> \
     --from-literal=KEY=value \
     -n <namespace>
   ```

3. If exists, verify required keys
   ```bash
   kubectl get secret <secret-name> -n <namespace> -o yaml
   ```

4. Pod will automatically retry

### Missing ConfigMap Recovery

**Steps**:
1. Check if ConfigMap exists
   ```bash
   kubectl get configmap <configmap-name> -n <namespace>
   ```

2. Create if missing or remove reference

3. Pod will automatically retry

### Recovery Outcome

**Success**: Config error resolved, container created
**Failure**: Config still invalid - review requirements

## Readiness Probe Failure Recovery Workflow

### Symptoms
- Pod status `Running 0/1`
- Readiness probe failing
- Pod not receiving traffic

### Recovery Steps

**Step 1: Check Probe Logs**
```bash
kubectl logs <pod-name> -n <namespace> --tail=100
```

**Step 2: Identify Issue**
```
Wrong port → Port Recovery
Slow startup → Timing Recovery
Endpoint missing → Application Issue
Dependencies not ready → Dependency Recovery
```

### Port Mismatch Recovery

**Steps**:
1. Check application port from logs
   ```
   "Starting server on port 3000"
   ```

2. Update probe port
   ```yaml
   readinessProbe:
     httpGet:
       port: 3000  # Match application port
   ```

3. Redeploy

### Timing Recovery (Slow Startup)

**Steps**:
1. Increase initialDelaySeconds
   ```yaml
   readinessProbe:
     initialDelaySeconds: 30  # Increase from 10
     timeoutSeconds: 5
   ```

2. Redeploy

3. Monitor startup time

### Recovery Outcome

**Success**: Probe passes, pod becomes ready
**Failure**: Probe still failing - check endpoint exists

## Multiple Pod Failure Recovery Workflow

### Symptoms
- Multiple pods failing
- Possibly different errors
- Service degraded or down

### Recovery Steps

**Step 1: Assess Situation**
```bash
kubectl get pods -n <namespace> -l devstack_label=<label>
```

**Step 2: Identify Pattern**
```
All same error → Common Issue Recovery
Different errors → Individual Recovery
```

### Common Issue Recovery

**If all pods have same error**:
1. Fix issue once (configuration, image, etc.)
2. Redeploy all
3. Monitor all pods together

**Priority**: Fix common issue first

### Individual Issue Recovery

**If pods have different errors**:
1. List all unique errors
2. Prioritize:
   - Configuration errors (fix once, affects all)
   - Resource errors (may affect subset)
   - Individual pod issues
3. Fix in priority order

### Recovery Outcome

**Success**: All pods running and ready
**Partial**: Some pods fixed, others need attention
**Failure**: Unable to recover - escalate

## Deployment Rollback Workflow

### When to Rollback
- New deployment causing crashes
- Configuration change broke service
- Image version has critical bug

### Rollback Steps

**Step 1: Identify Last Good State**
```bash
helm history <release-name> -n <namespace>
```

**Step 2: Rollback**
```bash
helm rollback <release-name> <revision> -n <namespace>
```

**Step 3: Verify Rollback**
```bash
kubectl get pods -n <namespace> -l devstack_label=<label>
```

**Step 4: Fix Issue**
```
Identify what caused failure
Fix in configuration
Test before redeploying
```

### Recovery Outcome

**Success**: Service restored to working state
**Action**: Fix issue before next deployment

## Emergency Recovery Procedures

### Complete Service Outage

**Step 1: Immediate Triage**
1. Check all pod statuses
2. Check recent events
3. Identify if cluster-wide or service-specific

**Step 2: Quick Fixes**
1. Try pod restart: `kubectl delete pods -n <ns> -l devstack_label=<label>`
2. Check dependencies: databases, caches, etc.
3. Verify resources: CPU, memory, network

**Step 3: Rollback if Necessary**
```bash
helm rollback <release> -n <namespace>
```

**Step 4: Escalate**
If quick fixes don't work:
- Notify team
- Check cluster health
- Review recent changes

### Cascading Failures

**Symptoms**: Multiple services failing

**Steps**:
1. Identify root cause service
2. Fix or disable failing service
3. Restart dependent services
4. Monitor recovery

### Recovery Time Objectives

| Issue Type | Target Recovery Time |
|------------|---------------------|
| OOMKilled | < 5 minutes (auto-fix) |
| Configuration Error | < 10 minutes |
| Missing Dependency | < 15 minutes |
| Application Bug | Variable (needs code fix) |
| Image Pull Error | < 5 minutes |
| Resource Constraints | < 10 minutes |

## Post-Recovery Actions

### After Successful Recovery

1. **Document Issue**
   - What failed
   - Root cause
   - Fix applied
   - Time to recover

2. **Update Configuration**
   - Commit fixes to git
   - Update documentation
   - Add to runbook

3. **Prevent Recurrence**
   - Add validation checks
   - Update deployment checklist
   - Improve monitoring

4. **Team Communication**
   - Share learnings
   - Update procedures
   - Review incident

### Recovery Checklist

- [ ] Service fully operational
- [ ] All pods running and ready
- [ ] Resources stable
- [ ] Logs clean (no errors)
- [ ] Endpoints accessible
- [ ] Monitoring showing healthy
- [ ] Documentation updated
- [ ] Team notified

## Recovery Metrics

Track these for continuous improvement:

- Mean Time to Detect (MTTD)
- Mean Time to Resolve (MTTR)
- Recovery success rate
- Auto-fix success rate
- Escalation frequency

## Advanced Recovery Techniques

### Debug Pod Deployment

```bash
kubectl run debug -it --rm --image=busybox -n <namespace> -- sh
```

### Network Debugging

```bash
kubectl run netshoot --rm -it --image=nicolaka/netshoot -- /bin/bash
```

### Exec into Failing Pod

```bash
kubectl exec -it <pod-name> -n <namespace> -- /bin/sh
```

## Recovery Decision Tree

```
Pod Failing?
├── Yes → Check Status
│   ├── CrashLoopBackOff → Check Logs → Identify Error Type
│   ├── ImagePullBackOff → Check Image Tag → Fix & Redeploy
│   ├── OOMKilled → Increase Memory → Redeploy
│   ├── Pending → Check Scheduling → Fix Resources/Selector
│   └── Error → Check Exit Code → Debug Application
└── No → Service Healthy → Continue Monitoring
```
