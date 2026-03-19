# Monitoring Subskill

Post-deployment health monitoring and ongoing status checks for helmfile deployments.

## Purpose

Provides continuous health monitoring, resource tracking, and access verification for deployed services.

## When to Use

- After deployment to verify health
- Ongoing health checks for running services
- Resource usage monitoring
- Access URL verification
- Troubleshooting intermittent issues

## Health Check Workflow

### Post-Deployment Monitoring

**Timing**: Start 30-60 seconds after deployment

#### Phase 1: Initial Status Check

```bash
# Get pod status
kubectl get pods -n <namespace> -l devstack_label=<label>
```

**Expected States**:
- `Running` (1/1) → Healthy ✅
- `Running` (0/1) → Not ready, check probes ⚠️
- `ContainerCreating` → Still starting, wait ⏳
- `Pending` → Check scheduling → [Debugging](debugging.md)
- `CrashLoopBackOff` → Check logs → [Debugging](debugging.md)
- `ImagePullBackOff` → Check image → [Debugging](debugging.md)

**Status Indicators**:
```
NAME                                    READY   STATUS    RESTARTS   AGE
payment-service-john-web-abc123         1/1     Running   0          45s  ✅
payment-service-john-worker-def456      0/1     Running   0          45s  ⚠️
```

#### Phase 2: Readiness Verification

If pods show `Running 0/1`:

```bash
# Check why not ready
kubectl describe pod <pod-name> -n <namespace> | grep -A 10 "Conditions:"
```

**Common Reasons**:
- Readiness probe failing
- Application still initializing
- Dependencies not ready

**Actions**:
- Wait up to 2 minutes for readiness
- Check logs if not ready after 2 minutes
- Verify probe configuration

#### Phase 3: Service Verification

```bash
# Check services
kubectl get svc -n <namespace> | grep <devstack-label>

# Check endpoints (verify service routes to pods)
kubectl get endpoints -n <namespace> | grep <devstack-label>
```

**Expected**:
```
NAME                      TYPE        CLUSTER-IP       PORT(S)
payment-service-john      ClusterIP   172.20.45.67     80/TCP

NAME                      ENDPOINTS
payment-service-john      10.244.1.5:8080,10.244.2.3:8080
```

**If No Endpoints**:
- Service selector doesn't match pod labels
- Pods not ready yet
- Service misconfigured

#### Phase 4: Ingress/Route Verification

```bash
# Check ingress routes
kubectl get ingressroute -n <namespace> | grep <devstack-label>

# Get ingress details
kubectl describe ingressroute <route-name> -n <namespace>
```

**Extract Access URL**:
```
Host: payment-service-john.devstack.example.com
```

## Ongoing Monitoring

### Pod Health Status

**Check Current Status**:
```bash
kubectl get pods -n <namespace> -l devstack_label=<label> -o wide
```

**Additional Info with `-o wide`**:
- Node placement
- IP address
- Nominated node

**Watch Mode** (real-time updates):
```bash
kubectl get pods -n <namespace> -l devstack_label=<label> --watch
```

### Resource Usage Monitoring

**Pod Resource Usage** (requires metrics-server):
```bash
kubectl top pod -n <namespace> -l devstack_label=<label>
```

**Output**:
```
NAME                                CPU(cores)   MEMORY(bytes)
payment-service-john-web-abc123     45m          82Mi
payment-service-john-worker-def456  65m          156Mi
```

**Analysis**:
- Compare CPU usage vs limits
- Compare memory usage vs limits
- Identify if approaching limits

**Example Analysis**:
```
Resource Usage Analysis:

payment-service-john-web:
  CPU: 45m / 100m (45% of limit) ✅
  Memory: 82Mi / 100Mi (82% of limit) ⚠️

  Recommendation: Consider increasing memory limit to 200Mi
  Current usage close to limit may cause OOMKills under load.
```

### Log Monitoring

**Current Logs**:
```bash
kubectl logs <pod-name> -n <namespace> --tail=50
```

**Follow Logs** (real-time):
```bash
kubectl logs <pod-name> -n <namespace> -f
```

**Filter Logs** (find errors):
```bash
kubectl logs <pod-name> -n <namespace> --tail=100 | grep -i error
kubectl logs <pod-name> -n <namespace> --tail=100 | grep -i warning
kubectl logs <pod-name> -n <namespace> --tail=100 | grep -i exception
```

**Multiple Containers**:
```bash
# Specific container
kubectl logs <pod-name> -c <container-name> -n <namespace>

# All containers
kubectl logs <pod-name> --all-containers=true -n <namespace>
```

### Event Monitoring

**Recent Events**:
```bash
kubectl get events -n <namespace> --sort-by='.lastTimestamp' | tail -20
```

**Pod-Specific Events**:
```bash
kubectl get events -n <namespace> --field-selector involvedObject.name=<pod-name> --sort-by='.lastTimestamp'
```

**Watch Events** (real-time):
```bash
kubectl get events -n <namespace> --watch
```

**Event Types to Watch**:
- `Warning` events → Potential issues
- `BackOff` → Container restart issues
- `Unhealthy` → Probe failures
- `FailedScheduling` → Scheduling problems

## Health Check Intervals

### Continuous Monitoring Schedule

**First 5 Minutes** (critical startup period):
- Check every 30 seconds
- Watch for CrashLoopBackOff
- Verify readiness

**5-30 Minutes** (stabilization period):
- Check every 2 minutes
- Monitor resource usage
- Watch for patterns

**After 30 Minutes** (steady state):
- Check every 10-15 minutes
- Look for anomalies
- Monitor trends

## Access Verification

### Service Endpoints

**Internal Access** (from within cluster):
```
http://<service-name>.<namespace>.svc.cluster.local:<port>
```

Example:
```
http://payment-service-john.payment-service.svc.cluster.local:80
```

**Verify from Another Pod**:
```bash
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://payment-service-john.payment-service.svc.cluster.local/health
```

### Ingress URLs

**External Access**:
```
https://<service-name>-<label>.<domain>
```

Example:
```
https://payment-service-john.devstack.example.com
```

**Test Endpoints**:
```bash
# Health check
curl https://payment-service-john.devstack.example.com/health

# Readiness check
curl https://payment-service-john.devstack.example.com/ready

# API endpoint (if applicable)
curl https://payment-service-john.devstack.example.com/api/v1/status
```

## Monitoring Report Format

### Healthy Deployment Report

```
## ✅ Health Check: HEALTHY

### Pod Status
**payment-service-john-web-abc123**
- Status: Running (1/1)
- Age: 5m23s
- Restarts: 0
- Node: node-03

**payment-service-john-worker-def456**
- Status: Running (1/1)
- Age: 5m23s
- Restarts: 0
- Node: node-05

### Resource Usage
**Web Container**:
- CPU: 45m / 100m (45%)
- Memory: 82Mi / 100Mi (82%) ⚠️
- Recommendation: Increase memory limit

**Worker Container**:
- CPU: 65m / 150m (43%)
- Memory: 145Mi / 256Mi (56%) ✅

### Services
**payment-service-john**:
- Type: ClusterIP
- IP: 172.20.45.67
- Endpoints: 2 pods connected ✅

### Access URLs
- Internal: http://payment-service-john.payment-service.svc.cluster.local
- External: https://payment-service-john.devstack.example.com

### Recent Activity
- No errors in last 10 minutes ✅
- No warnings in last 10 minutes ✅
- Probes passing consistently ✅

### Next Check
Scheduled in 15 minutes (or on-demand)
```

### Unhealthy Deployment Report

```
## ⚠️ Health Check: DEGRADED

### Issues Detected

**Issue 1: High Memory Usage**
- Pod: payment-service-john-web-abc123
- Memory: 95Mi / 100Mi (95%) ❌
- Status: Risk of OOMKill
- Action: Increasing memory limit to 200Mi

**Issue 2: Readiness Probe Failing**
- Pod: payment-service-john-worker-def456
- Status: Running (0/1)
- Probe: HTTP GET /ready timeout
- Action: Investigating logs...

### Pod Status
**payment-service-john-web-abc123**: Running but high memory ⚠️
**payment-service-john-worker-def456**: Running but not ready ❌

### Immediate Actions
1. ⏳ Increasing web memory limit
2. 🔍 Checking worker logs for readiness issues
3. ⏳ Waiting 2 minutes for stabilization

### Will Monitor
- Memory usage after limit increase
- Worker readiness status
- Overall stability
```

## Advanced Monitoring

### Port Forwarding for Local Testing

```bash
# Forward service port to local machine
kubectl port-forward -n <namespace> svc/<service-name> 8080:80

# Test locally
curl http://localhost:8080/health
```

### Execute Commands in Pod

```bash
# Get shell access
kubectl exec -it <pod-name> -n <namespace> -- /bin/sh

# Run specific command
kubectl exec <pod-name> -n <namespace> -- ps aux
kubectl exec <pod-name> -n <namespace> -- netstat -tlpn
```

### Network Debugging

**Check DNS Resolution**:
```bash
kubectl exec <pod-name> -n <namespace> -- nslookup mysql-base
```

**Check Connectivity**:
```bash
kubectl exec <pod-name> -n <namespace> -- nc -zv mysql-base 3306
```

**Check Routes**:
```bash
kubectl exec <pod-name> -n <namespace> -- ip route
```

## Monitoring Commands Reference

```bash
# Quick status check
kubectl get pods -n <namespace> -l devstack_label=<label>

# Detailed info
kubectl describe pod <pod-name> -n <namespace>

# Resource usage
kubectl top pod <pod-name> -n <namespace>

# Logs (last 50 lines)
kubectl logs <pod-name> -n <namespace> --tail=50

# Follow logs
kubectl logs <pod-name> -n <namespace> -f

# Previous logs (if crashed)
kubectl logs <pod-name> -n <namespace> --previous

# Events
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# Services
kubectl get svc -n <namespace> | grep <label>

# Endpoints
kubectl get endpoints -n <namespace> | grep <label>

# Ingress routes
kubectl get ingressroute -n <namespace> | grep <label>

# Wide output (includes node, IP)
kubectl get pods -n <namespace> -l devstack_label=<label> -o wide

# JSON output (for scripting)
kubectl get pods -n <namespace> -l devstack_label=<label> -o json

# Watch mode (real-time)
kubectl get pods -n <namespace> -l devstack_label=<label> --watch
```

## Alerting Patterns

### When to Alert

**Critical** (immediate action):
- Pod in CrashLoopBackOff
- OOMKilled events
- All replicas down
- Service endpoints empty

**Warning** (monitor closely):
- Memory usage > 90%
- CPU usage > 80%
- Readiness probe failing
- High restart count (> 5)

**Info** (for awareness):
- Memory usage > 70%
- Successful scaling events
- Configuration updates

## Related Subskills

- [Deployment](deployment.md) - Triggers monitoring after deployment
- [Debugging](debugging.md) - Called when issues detected
- [Validation](validation.md) - Prevents issues before deployment
