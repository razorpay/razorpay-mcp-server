# Devstack Frequently Asked Questions (FAQ)

Common questions and answers about using devstack for development and deployment.

## Table of Contents

1. [TTL and Label Management](#ttl-and-label-management)
2. [External Access and Networking](#external-access-and-networking)
3. [Deployment and Configuration](#deployment-and-configuration)
4. [Troubleshooting](#troubleshooting)

---

## TTL and Label Management

### Q: How to keep deployments/labels up for more than 72 hours?

**Answer:**

By default, devstack service deployments can be kept running for **72 hours maximum**. This applies to non-base pods only. To keep a label up for more than 72 hours, the label must be **whitelisted**.

**Steps to Whitelist a Label:**

1. **Update the Whitelist Configuration**

   Add your label to the whitelist file:
   ```
   https://github.com/razorpay/kube-manifests/blob/master/devstack-config/whitelist_labels.yaml
   ```

2. **Create Pull Request**

   Submit a PR with your label added to `whitelist_labels.yaml`

3. **Restart TTL Validation Webhook**

   After the PR is merged, restart the webhook:
   ```bash
   kubectl rollout restart deployment ttl-validation-webhook -n kube-system
   ```

4. **Ensure devstack_label is Present**

   All deployments for this label **MUST** have the `devstack_label` label:

   **Required Label Configuration:**
   ```yaml
   metadata:
     annotations:
       "helm.sh/hook": pre-install,pre-upgrade
       "helm.sh/hook-weight": "7"
       janitor/ttl: "{{ .Values.ttl }}"
     labels:
       bu: {{ .Values.bu }}
       name: {{ .Values.name }}-{{ .Values.devstack_label }}

       # ------------------- CRITICAL: This label MUST be present --------------------- #
       devstack_label: {{ .Values.devstack_label }}
       # ------------------------------------------------------------------------------ #

       {{ if eq .Values.devstack_label "base" }}
       velero.io/include-in-backup: "true"
       {{ end }}
     name: {{ .Values.name }}-{{ .Values.devstack_label }}
     namespace: {{ .Values.namespace }}
   ```

**Important Notes:**
- Whitelisting is required for labels that need to run longer than 72 hours
- The `devstack_label` label must be present on ALL resources (deployments, services, configmaps, etc.)
- Base pods are permanent and don't require whitelisting
- Non-whitelisted labels will be automatically cleaned up after 72 hours by the janitor

**Example Workflow:**
```bash
# 1. Add label to whitelist_labels.yaml
echo "  - my-long-running-label" >> devstack-config/whitelist_labels.yaml

# 2. Commit and push
git add devstack-config/whitelist_labels.yaml
git commit -m "Whitelist label: my-long-running-label"
git push origin feature/whitelist-label

# 3. After PR merge, restart webhook
kubectl rollout restart deployment ttl-validation-webhook -n kube-system

# 4. Verify webhook restarted
kubectl get pods -n kube-system -l app=ttl-validation-webhook

# 5. Deploy with whitelisted label
/devstack deploy my-service with label my-long-running-label
```

---

## External Access and Networking

### Q: How to make an endpoint available externally?

**Answer:**

To make a service endpoint accessible from outside the cluster (internet), you need to:

1. **Onboard the route to Edge**
2. **Create an external IngressRoute**

**Detailed Steps:**

#### Step 1: Onboard to Edge

Follow the edge onboarding documentation:
```
https://docs.google.com/document/d/1mA-LCh-mezKMZz6lykAIHy00ErsuZVq_N1PIwsbF2eQ/edit?usp=sharing
```

#### Step 2: Create External IngressRoute

Create an IngressRoute that points to `traefik-external`:

**Critical Configuration Requirements:**
- ✅ `kubernetes.io/ingress.class: traefik-external` annotation
- ✅ Domains set as `*.ext.dev.razorpay.in`
- ✅ Route53 entries are **NOT required**
- ⚠️ **No IP whitelisting** - exposed domains are publicly accessible from internet

**Example External IngressRoute:**

```yaml
kind: IngressRoute
apiVersion: traefik.containo.us/v1alpha1
metadata:
  annotations:
    kubernetes.io/ingress.class: traefik-external  # ← REQUIRED for external access
    janitor/ttl: "{{ .Values.ttl }}"
  name: {{ .Values.name }}-edge-ext-{{ .Values.devstack_label }}
  namespace: {{ .Values.namespace }}
spec:
  entryPoints:
    - http
  routes:
    # Route 1: Direct label-specific access
    - kind: Rule
      match: Host(`{{ .Values.name }}-{{ .Values.devstack_label }}.ext.dev.razorpay.in`)
      services:
        - name: '{{ .Values.name }}-{{ .Values.devstack_label }}-edge-redirect'
          port: 8000
      middlewares:
        - name: edge-injectheader-{{ .Values.devstack_label }}

    # Route 2: Access via rzpctx header
    - kind: Rule
      match: Host(`{{ .Values.name }}.ext.dev.razorpay.in`) && Headers(`rzpctx-dev-serve-user`,`{{ .Values.devstack_label }}`)
      services:
        - name: '{{ .Values.name }}-{{ .Values.devstack_label }}-edge-redirect'
          port: 8000

    # Route 3: Fallback to base deployment
    - kind: Rule
      match: Host(`{{ .Values.name }}.ext.dev.razorpay.in`)
      services:
        - name: '{{ .Values.name }}-base-edge-redirect'
          port: 8000
```

**Access Patterns:**

1. **Label-Specific URL:**
   ```
   https://myservice-alice.ext.dev.razorpay.in/api/endpoint
   ```

2. **Base URL with Header:**
   ```bash
   curl -H "rzpctx-dev-serve-user: alice" \
        https://myservice.ext.dev.razorpay.in/api/endpoint
   ```

3. **Base URL (fallback):**
   ```
   https://myservice.ext.dev.razorpay.in/api/endpoint
   # Falls back to base deployment
   ```

**Security Considerations:**

⚠️ **IMPORTANT**:
- External LB has **NO IP whitelisting**
- Once exposed, the endpoint is **publicly accessible from the internet**
- Only expose endpoints that are safe for public access
- Consider adding authentication/authorization at the application level
- Use internal ingress for internal-only services

**Internal vs External Ingress:**

| Feature | Internal Ingress | External Ingress |
|---------|-----------------|------------------|
| Annotation | `traefik-internal` | `traefik-external` |
| Domain | `*.dev.razorpay.in` | `*.ext.dev.razorpay.in` |
| Access | Within cluster/VPN | Public internet |
| IP Whitelist | Yes (VPN/office) | No |
| Use Case | Internal services | Public APIs, webhooks |

**Example: Internal Ingress (VPN-only)**

```yaml
kind: IngressRoute
apiVersion: traefik.containo.us/v1alpha1
metadata:
  annotations:
    kubernetes.io/ingress.class: traefik-internal  # ← Internal only
  name: {{ .Values.name }}-internal-{{ .Values.devstack_label }}
spec:
  routes:
    - match: Host(`{{ .Values.name }}-{{ .Values.devstack_label }}.dev.razorpay.in`)
      services:
        - name: {{ .Values.name }}-{{ .Values.devstack_label }}
          port: 8080
```

---

### Q: What are devstack NAT IPs?

**Answer:**

Devstack uses the following **NAT IP addresses** for outbound traffic:

```
52.66.76.68
52.66.95.207
13.127.201.109
```

**When to Use NAT IPs:**

1. **IP Whitelisting External Services**
   - If your service needs to connect to external APIs that require IP whitelisting
   - Whitelist all three NAT IPs for redundancy

2. **Firewall Rules**
   - Configure firewall rules to allow traffic from these IPs
   - Used for outbound connections from devstack pods

3. **Third-Party Integration**
   - Share these IPs with third-party services for whitelisting
   - Examples: Payment gateways, external APIs, webhooks

**Example Use Cases:**

**Use Case 1: Whitelist for Payment Gateway**
```
Service: PaymentGateway
Whitelist these IPs:
- 52.66.76.68
- 52.66.95.207
- 13.127.201.109

Reason: Devstack pods will make API calls to payment gateway
```

**Use Case 2: Database IP Whitelist**
```bash
# If connecting to external database, whitelist NAT IPs
# Example: RDS security group
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxx \
  --protocol tcp \
  --port 3306 \
  --cidr 52.66.76.68/32

aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxx \
  --protocol tcp \
  --port 3306 \
  --cidr 52.66.95.207/32

aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxx \
  --protocol tcp \
  --port 3306 \
  --cidr 13.127.201.109/32
```

**Important Notes:**
- All outbound traffic from devstack pods uses these NAT IPs
- IPs are shared across all devstack deployments
- Changes to NAT IPs are rare but communicated in advance
- Always whitelist all three IPs for high availability

---

## Deployment and Configuration

### Q: What is the default TTL for devstack deployments?

**Answer:**

The default TTL (Time To Live) depends on the label:

- **Base deployments** (`devstack_label: base`): Permanent (no TTL)
- **Non-base deployments**: 72 hours (3 days)
- **Whitelisted labels**: Can be extended beyond 72 hours

**TTL Values:**

```yaml
# Common TTL values in helmfile.yaml
ttl: 1h    # 1 hour
ttl: 8h    # 8 hours (typical dev work)
ttl: 24h   # 1 day
ttl: 72h   # 3 days (maximum for non-whitelisted)
ttl: forever  # Only for whitelisted labels or base
```

**How TTL Works:**

1. **Janitor Service** monitors all pods with `janitor/ttl` annotation
2. **After TTL expires**, resources are automatically deleted
3. **Whitelisted labels** can use longer TTLs or `forever`
4. **Base deployments** are never cleaned up

**Example:**
```yaml
metadata:
  annotations:
    janitor/ttl: "{{ .Values.ttl }}"  # Templated from helmfile
```

---

### Q: How do I debug "ImagePullBackOff" errors?

**Answer:**

See the [Debugging Subskill](../subskills/debugging.md) for comprehensive troubleshooting.

**Quick Fix:**

1. **Verify image exists:**
   ```bash
   # Check image in Harbor
   curl 'https://harbor-image-checker.dev.razorpay.in/check-images' \
     -H 'Content-Type: application/json' \
     -d '{"images": ["c.rzp.io/razorpay/service:commit-id"]}'
   ```

2. **Check image pull secrets:**
   ```bash
   kubectl get secrets -n <namespace>
   kubectl get pod <pod-name> -n <namespace> -o yaml | grep imagePullSecrets
   ```

3. **Verify image tag in helmfile.yaml:**
   ```yaml
   values:
     - image: <correct-commit-id>  # Ensure this is correct
   ```

4. **Redeploy with correct image:**
   ```bash
   # Update helmfile.yaml with correct image
   /devstack deploy service with image <correct-commit-id>
   ```

---

### Q: Can I use custom domains for devstack services?

**Answer:**

**No, custom domains are not supported** on devstack. You must use the provided domain patterns:

**Internal Access (VPN):**
```
<service-name>-<label>.dev.razorpay.in
```

**External Access (Public):**
```
<service-name>-<label>.ext.dev.razorpay.in
```

**If you need a custom domain:**
- Deploy to staging or production environment
- Use Route53 to configure custom domains
- Devstack is for development only, not production-like custom domains

---

### Q: How do I access logs for my service?

**Answer:**

**Option 1: Using kubectl**
```bash
# Get pod name
kubectl get pods -n <namespace> -l devstack_label=<your-label>

# Stream logs
kubectl logs -f <pod-name> -n <namespace>

# Get logs from previous container (if crashed)
kubectl logs <pod-name> -n <namespace> --previous
```

**Option 2: Using devspace (recommended for development)**
```bash
cd ~/repos/<service>
devspace dev --no-warn
# Logs stream automatically
```

**Option 3: Using monitoring subskill**
```bash
/devstack
Show me logs for service in namespace <namespace> with label <label>
```

---

### Q: Can I run multiple services with the same label?

**Answer:**

**Yes**, you can run multiple services with the same devstack label. This is common for microservices deployments.

**Example:**
```yaml
# helmfile.yaml - all uncommented, same label
- name: api-{{ .Values.devstack_label }}
  namespace: api

- name: worker-{{ .Values.devstack_label }}
  namespace: worker

- name: scheduler-{{ .Values.devstack_label }}
  namespace: scheduler
```

**Deploy all together:**
```bash
/devstack deploy all services with label alice
```

**Benefits:**
- Related services share the same lifecycle
- Easy cleanup (delete all with same label)
- Consistent environment across microservices

---

### Q: How do I connect to an ephemeral database?

**Answer:**

Ephemeral databases are automatically created when you deploy with `ephemeral_db: true`.

**Connection Details:**

1. **Database is accessible within the cluster:**
   ```
   Host: <service>-<label>-db.<namespace>.svc.cluster.local
   Port: 3306 (MySQL) or 5432 (PostgreSQL)
   Username: From secret
   Password: From secret
   Database: <service>-<label>
   ```

2. **Connection string format:**
   ```bash
   # MySQL
   mysql://<USERNAME>:<PASSWORD>@service-label-db.namespace.svc.cluster.local:3306/database

   # PostgreSQL
   postgresql://<USERNAME>:<PASSWORD>@service-label-db.namespace.svc.cluster.local:5432/database
   ```

3. **Access from local machine (port-forward):**
   ```bash
   kubectl port-forward svc/<service>-<label>-db -n <namespace> 3306:3306

   # Connect locally
   mysql -h 127.0.0.1 -P 3306 -u <user> -p
   ```

4. **Get credentials from secret:**
   ```bash
   kubectl get secret <service>-<label>-db-secret -n <namespace> -o yaml
   ```

---

### Q: What happens when I redeploy a service?

**Answer:**

When you redeploy (using `delete_before_sync: true`, which is default):

1. **Existing deployment is deleted:**
   ```bash
   helmfile delete || true  # Ignore failures
   ```

2. **All resources are removed:**
   - Pods are terminated
   - Services are deleted
   - ConfigMaps and Secrets are recreated
   - Persistent data (if any) may be lost

3. **Fresh deployment is created:**
   ```bash
   helmfile sync  # Creates new resources
   ```

4. **Hooks re-execute:**
   - Database configurators run again
   - SQS/SNS setup runs again
   - Secrets are regenerated

**Important:**
- **Ephemeral databases** are recreated (data is lost)
- **Custom databases** persist if not deleted
- **Persistent volumes** persist if configured
- **Downtime** occurs during redeployment

**To update without deletion:**
```json
// config.json
{
  "delete_before_sync": false
}
```

Then deploy:
```bash
/devstack deploy service update existing
```

---

### Q: How do I share my devstack deployment with others?

**Answer:**

**Option 1: Share the URL (Internal)**
```
https://<service>-<label>.dev.razorpay.in

Example: https://api-alice.dev.razorpay.in
```
- Accessible to anyone on VPN
- No additional configuration needed

**Option 2: Share with rzpctx Header (External)**
```bash
curl -H "rzpctx-dev-serve-user: alice" \
     https://api.ext.dev.razorpay.in/endpoint
```
- Works for external routes
- Recipients use the header to reach your label

**Option 3: Create Demo Label**
```bash
# Deploy with a shared label name
/devstack deploy api with label demo

# Share: https://api-demo.dev.razorpay.in
```

**Security Notes:**
- Internal URLs require VPN access
- External URLs are publicly accessible
- Don't expose sensitive services externally
- Consider authentication for shared deployments

---

## Troubleshooting

### Q: My pod is stuck in "Pending" state. What should I do?

**Answer:**

**Common Causes:**

1. **Insufficient Resources**
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   # Look for: "Insufficient cpu" or "Insufficient memory"
   ```

   **Fix:** Reduce resource requests in values.yaml:
   ```yaml
   web_requests_cpu: 50m      # Reduce from higher value
   web_requests_memory: 50Mi  # Reduce from higher value
   ```

2. **Node Selector Mismatch**
   ```bash
   kubectl describe pod <pod-name> -n <namespace>
   # Look for: "No nodes available to schedule pods"
   ```

   **Fix:** Check nodeSelector in deployment.yaml:
   ```yaml
   nodeSelector:
     devstack: "true"  # Ensure this label exists on nodes
   ```

3. **PVC Not Bound**
   ```bash
   kubectl get pvc -n <namespace>
   # Check if PVC is "Bound" or "Pending"
   ```

   **Fix:** Check storage class and availability

**Debug Commands:**
```bash
# Get detailed pod status
kubectl describe pod <pod-name> -n <namespace>

# Check events
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# Check node resources
kubectl top nodes
```

---

### Q: How do I clean up old deployments?

**Answer:**

**Option 1: Let TTL Cleanup Happen Automatically**
- Janitor service cleans up after TTL expires
- Wait for automatic cleanup (up to 72 hours)

**Option 2: Manual Cleanup**
```bash
# Delete specific service
helmfile -f helmfile.yaml delete

# Or use kubectl
kubectl delete all -l devstack_label=<label> -n <namespace>

# Delete all resources including secrets
kubectl delete all,secret,configmap,pvc -l devstack_label=<label> -n <namespace>
```

**Option 3: Clean Specific Namespace**
```bash
# Delete all pods with label
kubectl delete pods -l devstack_label=<label> -n <namespace>

# Delete deployment
kubectl delete deployment <service>-<label> -n <namespace>
```

**Bulk Cleanup:**
```bash
# Delete all non-base deployments in namespace
kubectl delete all -l 'devstack_label!=base' -n <namespace>
```

**Important:**
- Be careful with deletion commands
- Verify label before deleting
- Backup important data before cleanup
- Cannot recover deleted ephemeral resources

---

### Q: Why is my service getting OOMKilled?

**Answer:**

**OOMKilled** (Out Of Memory Killed) means the container exceeded its memory limit.

**Immediate Fix:**

1. **Check current memory limit:**
   ```bash
   kubectl describe pod <pod-name> -n <namespace> | grep -A 5 Limits
   ```

2. **Increase memory limit in values.yaml:**
   ```yaml
   # Before
   web_limits_memory: 100Mi

   # After
   web_limits_memory: 200Mi  # Double or increase as needed
   ```

3. **Redeploy:**
   ```bash
   /devstack deploy service with label <label>
   ```

**Long-term Solutions:**

1. **Monitor Memory Usage:**
   ```bash
   kubectl top pod <pod-name> -n <namespace>
   ```

2. **Profile Application:**
   - Check for memory leaks
   - Optimize memory-intensive operations
   - Use memory profiling tools

3. **Set Appropriate Limits:**
   ```yaml
   # Conservative
   web_requests_memory: 50Mi
   web_limits_memory: 100Mi

   # Medium
   web_requests_memory: 100Mi
   web_limits_memory: 200Mi

   # High Memory Service
   web_requests_memory: 500Mi
   web_limits_memory: 1Gi
   ```

**See Also:** [Debugging Subskill](../subskills/debugging.md) - OOMKilled section

---

### Q: How do I know if my deployment was successful?

**Answer:**

**Check Pod Status:**
```bash
kubectl get pods -n <namespace> -l devstack_label=<label>

# Should show: Running (1/1 Ready)
```

**Check Service:**
```bash
kubectl get svc -n <namespace> -l devstack_label=<label>

# Should show: ClusterIP assigned
```

**Check Ingress:**
```bash
kubectl get ingress -n <namespace> -l devstack_label=<label>

# Should show: Hosts configured
```

**Test Endpoint:**
```bash
curl https://<service>-<label>.dev.razorpay.in/health
# or
curl https://<service>-<label>.ext.dev.razorpay.in/health
```

**Check Logs:**
```bash
kubectl logs <pod-name> -n <namespace>
# Look for startup messages, no errors
```

**Success Indicators:**
- ✅ Pod status: Running
- ✅ Ready: 1/1
- ✅ Restarts: 0 (or low number)
- ✅ Service has ClusterIP
- ✅ Endpoint responds
- ✅ No error logs

---

## Additional Resources

- [Deployment Subskill](../subskills/deployment.md)
- [Debugging Subskill](../subskills/debugging.md)
- [Validation Subskill](../subskills/validation.md)
- [Monitoring Subskill](../subskills/monitoring.md)
- [Devspace Code Sync](../subskills/devspace.md)
- [Error Patterns Reference](error-patterns.md)
- [Configuration Checklist](config-checklist.md)

---

## Need More Help?

If your question isn't answered here:

1. **Check Devstack Documentation:**
   ```
   https://alpha.razorpay.com/repo/devstack-docs
   ```
   Comprehensive documentation, guides, and examples

2. **Reach Out on Slack:**
   - Channel: `#platform-devstack`
   - Get help from the devstack team
   - Ask questions and share feedback

3. **Internal Resources:**
   - Check the [Debugging Subskill](../subskills/debugging.md) for troubleshooting
   - Review [Error Patterns](error-patterns.md) for common issues
   - Use `/devstack` skill to ask specific questions

4. **Debug Commands:**
   - Check pod events: `kubectl describe pod <pod-name> -n <namespace>`
   - Review pod logs: `kubectl logs <pod-name> -n <namespace>`
   - Monitor pod status: `kubectl get pods -n <namespace> -l devstack_label=<label>`

---

**Last Updated:** 2026-01-23
**Version:** 1.8.0
