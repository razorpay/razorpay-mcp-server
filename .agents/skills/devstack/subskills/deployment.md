# Deployment Subskill

Autonomous deployment workflow for helmfile-based services.

## Purpose

Handles end-to-end deployment of services including pre-validation, helmfile sync, and initial health checks.

## CRITICAL DEPLOYMENT RULES

**⚠️ IMPORTANT CHANGES - READ CAREFULLY**:

1. **NO Selector Flags**: NEVER use `-l` selector flags in helmfile commands
   - ❌ WRONG: `helmfile -f helmfile.yaml -l name=service-label sync`
   - ✅ CORRECT: `helmfile -f helmfile.yaml sync`

2. **Service Selection via Uncommenting**:
   - To deploy specific service(s): UNCOMMENT only those services, COMMENT OUT all others
   - To deploy all services: UNCOMMENT all desired services
   - Helmfile deploys ALL uncommented services automatically

3. **Auto-Uncomment Required**:
   - ALWAYS check if requested services are commented
   - AUTOMATICALLY uncomment ALL services mentioned in deployment request
   - Report uncommenting action to user

4. **Image Validation Required**:
   - ALWAYS validate images via Harbor API before deployment (unless `skip_image_validation: true`)
   - Check images exist in registry
   - Verify linux/amd64 architecture support
   - BLOCK deployment if images are invalid or missing
   - WARN if images don't support amd64 (ask user to check CI workflow)

## When to Use

- Deploying new services
- Updating existing deployments
- Re-deploying after configuration changes
- Rolling out new image versions

## Workflow

### Pre-Deployment Checklist

Before deploying, you MUST complete these steps for each requested service:

1. ✅ **Locate service** in helmfile.yaml
2. ✅ **Uncomment service** if it's commented out
3. ✅ **Update image field** if user specified new commit ID
4. ✅ **Comment out** all other services not being deployed
5. ✅ **Validate** configuration files
6. ✅ **Validate images** via Harbor API (check existence and amd64 support)
7. ✅ **Deploy** using `helmfile -f helmfile.yaml sync` (no selectors)

### Phase 1: Pre-Deployment Validation

#### 1. Change to Helmfile Directory

**Important**: Read the helmfile directory path from `../config.json`:
```bash
# The skill will read config.json to get the helmfile_directory path
# Example: cd /Users/parag.dudeja/Documents/Work/rzp-repos/harbor-action-tracking/kube-manifests/helmfile
cd <helmfile_directory from config.json>
```

**Auto-Detection Workflow**:
If `auto_detect: true` in config.json:
1. Try the configured `helmfile_directory` path
2. If not found, try fallback paths relative to repo root:
   - `kube-manifests/helmfile`
   - `helmfile`
   - `../kube-manifests/helmfile`
3. Report the path being used to the user
4. If none found, ask user to configure the path

#### 2. Locate Service in helmfile.yaml

Use Grep to find the service release entry:
```bash
grep -n "name: <service>-{{ .Values.devstack_label }}" helmfile.yaml
```

**CRITICAL - Service Comment Management**:
- For ALL services mentioned in the deployment request:
  - If commented (lines starting with #) → AUTOMATICALLY uncomment it
  - Verify the service entry includes all required fields
  - Check if image tag is specified
- **CRITICAL**: For ALL services NOT mentioned in deployment request:
  - If uncommented → AUTOMATICALLY comment it out
  - This prevents deploying unintended services
  - Only services explicitly mentioned should be uncommented

**Example**:
```
User request: "deploy pg-router and asv"

Actions:
1. Find pg-router → uncomment if needed ✓
2. Find asv → uncomment if needed ✓
3. Find ALL other uncommented services → comment them out ✓
4. Result: ONLY pg-router and asv are uncommented
```

**CRITICAL - Image Update**:
- **ALWAYS check the `image:` field** in the service's values section in helmfile.yaml
- The image value is a **commit ID** (git commit hash)
- Update workflow:
  1. If user specifies image/commit: **UPDATE** the `image:` field to the new commit ID
  2. If user says "existing", "current", or "keep": **DO NOT UPDATE** - keep current image value
  3. If no image specified: **DO NOT UPDATE** - keep current image value
- Image field location in helmfile.yaml:
  ```yaml
  - name: service-{{ .Values.devstack_label }}
    namespace: service
    chart: ./charts/service
    values:
      - image: <COMMIT_ID_HERE>  # ← UPDATE THIS if new image specified
      - devstack_label: {{ .Values.devstack_label }}
      - ttl: {{ .Values.ttl }}
  ```

**Example**:
```yaml
# Before (commented)
# - name: payment-service-{{ .Values.devstack_label }}
#   namespace: payment-service
#   chart: ./charts/payment-service

# After (uncommented automatically)
- name: payment-service-{{ .Values.devstack_label }}
  namespace: payment-service
  chart: ./charts/payment-service
  values:
    - image: abc123def
    - devstack_label: {{ .Values.devstack_label }}
    - ttl: {{ .Values.ttl }}
```

#### 3. Update Image Values (If Specified)

**CRITICAL STEP - Update Commit IDs in helmfile.yaml**:

For each service in the deployment request:
1. **Read the current image value** from helmfile.yaml
2. **Determine if update is needed**:
   - User specified new commit ID → UPDATE required
   - User said "existing", "current", "keep" → NO UPDATE
   - No image mentioned → NO UPDATE
3. **Update the image field** if needed using Edit tool

**Example Updates**:

```yaml
# User request: "deploy pg-router with image abc123def"
# BEFORE:
- name: pg-router-{{ .Values.devstack_label }}
  values:
    - image: old456xyz  # ← Current image

# AFTER (UPDATE):
- name: pg-router-{{ .Values.devstack_label }}
  values:
    - image: abc123def  # ← Updated to new commit ID
```

```yaml
# User request: "deploy asv with existing image"
# BEFORE:
- name: account-service-{{ .Values.devstack_label }}
  values:
    - image: current789  # ← Keep this

# AFTER (NO CHANGE):
- name: account-service-{{ .Values.devstack_label }}
  values:
    - image: current789  # ← Unchanged, using existing
```

**Important**:
- Image values are **git commit hashes** (SHA-1, 40 characters)
- Shortened commit IDs (7-8 characters) are also valid
- **Always verify** the commit exists before deploying

#### 4. Validate Chart Configuration

Read and validate these files:

**charts/<service-name>/values.yaml**:
- Check for required fields (see [../references/config-checklist.md](../references/config-checklist.md))
- Validate resource limits exist
- Verify TTL and devstack_label placeholders

**charts/<service-name>/templates/deployment.yaml**:
- Check for liveness/readiness probes
- Validate resource requests/limits
- Verify nodeSelector and labels
- Check DNS policy configuration

**Automatic Fixes Applied**:
- Missing resource limits → Add defaults
- Missing TTL annotation → Add janitor/ttl
- Missing devstack_label → Add to labels
- Missing DNS policy → Add ClusterFirst

#### 4. Run Template Validation

**CRITICAL**: Run template validation WITHOUT selector flags (validates all uncommented services)

```bash
helmfile -f helmfile.yaml template
```

**Purpose**: Render templates to catch syntax errors before deployment for all uncommented services

**If Template Errors Found**:
- Analyze error message
- Check for:
  - Missing values
  - Syntax errors in templates
  - Invalid YAML formatting
- Apply automatic fixes if possible
- Report issues that need manual intervention

#### 5. Image Validation (Pre-Deployment Check)

**CRITICAL**: Validate all container images exist in Harbor registry and support required architecture.

**Skip Validation**: Only if `skip_image_validation: true` in config.json

**Workflow**:

1. **Extract ALL Images from Template Output**:
   ```bash
   # Parse helmfile template output to extract ALL container images
   helmfile -f helmfile.yaml template 2>&1 | grep -E "image:|Image:" | grep -v "#" | awk '{print $2}' | grep "c.rzp.io" | sed 's/"//g' | sed "s/'//g" | sort -u
   ```

   **CRITICAL**: This extracts:
   - Main service images (api, worker, etc.)
   - All worker images (notification_worker, payment_worker, etc.)
   - Init container images
   - Sidecar container images
   - Migration job images
   - **ALL images** used by the deployment

2. **Build Complete Image List**:
   - Extract all unique container image references
   - Remove quotes and clean up formatting
   - Format: `c.rzp.io/razorpay/service:tag` or `c.rzp.io/razorpay/service:commit-id`
   - **Include ALL containers** - typically 10-20 images per service
   - Example for pg-router: api, 10+ worker images, migration images

3. **Call Harbor Image Validation API**:
   ```bash
   # Build JSON array with ALL extracted images
   curl 'https://harbor-image-checker.dev.razorpay.in/check-images' \
     -H 'Content-Type: application/json' \
     -d '{
       "images": [
         "c.rzp.io/razorpay/pg-router:api_commit-abc123",
         "c.rzp.io/razorpay/pg-router:notification_worker_commit-abc123",
         "c.rzp.io/razorpay/pg-router:payment_worker_commit-abc123",
         "c.rzp.io/razorpay/pg-router:ledger_worker_commit-abc123",
         "c.rzp.io/razorpay/asv:api-commit-def456",
         "c.rzp.io/razorpay/asv:worker-commit-def456"
         ... (include ALL images from template output)
       ]
     }'
   ```

   **IMPORTANT**:
   - Validate **ALL** images extracted in step 1
   - Don't just validate the first 1-2 images
   - Typical deployments have 10-20+ images
   - Missing validation will cause ImagePullBackOff during deployment

4. **Analyze Response**:
   - Check `valid_count`, `invalid_count`, `skipped_count`
   - For each image in results:
     - `valid: false` → **BLOCK deployment** - invalid image
     - `exists: false` → **BLOCK deployment** - image not found
     - Check `architectures` array for `linux/amd64` → **WARN if missing**

5. **Validation Rules**:
   - ✅ **PASS**: `valid: true`, `exists: true`, `linux/amd64` in architectures
   - ⚠️ **WARN**: `valid: true`, `exists: true`, but NO `linux/amd64` architecture
   - ❌ **FAIL**: `valid: false` OR `exists: false`

**Example Response Handling**:

```json
{
  "total": 2,
  "valid_count": 1,
  "invalid_count": 1,
  "results": [
    {
      "image": "c.rzp.io/razorpay/pg-router:abc123",
      "valid": true,
      "exists": true,
      "architectures": ["linux/amd64", "linux/arm64"]  // ✅ PASS
    },
    {
      "image": "c.rzp.io/razorpay/api:def456",
      "valid": true,
      "exists": false,  // ❌ FAIL - image not found
      "error": "Image not found in Harbor registry"
    }
  ]
}
```

**Actions Based on Results**:

- **All images valid with amd64**:
  - ✅ Proceed to deployment
  - Report: "All images validated successfully"

- **Missing amd64 architecture**:
  - ⚠️ **WARN user**: "Image {image} exists but does not support linux/amd64 architecture"
  - **Inform user**: "Please check CI workflow and ensure build includes amd64 platform"
  - **Ask user**: Continue deployment anyway? (may fail on amd64 nodes)

- **Invalid or missing images**:
  - ❌ **BLOCK deployment**
  - **Report**: "Image validation failed for: {image}"
  - **Error details**: Show error message from API
  - **Fix instructions**:
    - If `exists: false`: "Image not found. Check commit ID and ensure CI build completed"
    - If `valid: false`: Show specific error from API response
  - **Do not proceed** with deployment until fixed

**Configuration**:

To skip image validation, update `config.json`:
```json
{
  "skip_image_validation": true
}
```

**Error Handling**:
- If API call fails (network error, timeout): **WARN** but allow deployment
- If API returns 500/error: **WARN** but allow deployment
- Only **BLOCK** if API successfully returns invalid results

### Phase 2: Deployment

#### 6. Clean Previous Deployment (Default Behavior)

**IMPORTANT**: By default, always delete the previous deployment before syncing to ensure a clean state.

**CRITICAL - Deploy Without Selector Flag**:
- Deploy using `helmfile -f helmfile.yaml` WITHOUT the `-l` selector flag
- This ensures all uncommented services in helmfile.yaml are deployed
- Only uncommented services will be deployed

```bash
# Delete existing releases (ignore failures)
helmfile -f helmfile.yaml delete || true
```

**Why delete first?**
- ✅ **Clean slate**: Ensures no stale resources from previous deployments
- ✅ **Fresh configuration**: All configs applied from scratch
- ✅ **Prevents conflicts**: Avoids issues with changed resource types
- ✅ **Hook re-execution**: DB/SQS/SNS configurators run fresh
- ✅ **Secret updates**: Secrets regenerated with latest values

**Failure handling**:
- Use `|| true` to ignore delete failures
- If release doesn't exist, delete fails gracefully
- Sync proceeds regardless of delete status

**Skip delete only if**:
- User explicitly says "deploy without deleting" or "keep existing deployment"
- User says "update existing deployment"
- User requests "incremental update"
- Configuration has `delete_before_sync: false` in config.json

**To disable delete by default**:
Edit `agent-skills/infrastructure/skills/devstack/config.json`:
```json
{
  "delete_before_sync": false
}
```

**Note**: Keeping `delete_before_sync: true` (default) is recommended for clean deployments

#### 7. Execute Deployment

**CRITICAL - No Selector Flag**:
Deploy all uncommented services without using selector flags:

```bash
# Sync all uncommented services
helmfile -f helmfile.yaml sync
```

**Complete deployment command** (executed together):
```bash
helmfile -f helmfile.yaml delete || true; \
helmfile -f helmfile.yaml sync
```

**Monitor Output For**:
- Delete completion (if release existed)
- Helm release creation status
- Any error messages
- Resource creation confirmations

**If Deployment Fails**:
- Capture error output
- Proceed to debugging phase
- Provide detailed error analysis

**Success Indicators**:
```
# Delete phase (may not exist)
release "<service>-<label>" uninstalled

# Sync phase
Installing release=<service>-<label>, chart=./charts/<service>
Release "<service>-<label>" has been installed. Happy Helming!
```

### Phase 3: Initial Health Check

#### 8. Verify Deployment Created Resources

Wait 10 seconds after deployment, then check:

```bash
# Check if pods were created
kubectl get pods -n <namespace> -l devstack_label=<label>

# Check if services were created
kubectl get svc -n <namespace> | grep <label>
```

**Expected State After Deployment**:
- Pods: ContainerCreating or Running (may take 30-60s)
- Services: Created with ClusterIP assigned
- No immediate errors in events

#### 9. Transition to Monitoring

After initial deployment:
- Wait 30-60 seconds for pods to initialize
- Proceed to [Monitoring Subskill](monitoring.md) for health verification
- If issues detected, proceed to [Debugging Subskill](debugging.md)

## Multi-Service Deployment Patterns

### Deploy All Uncommented Services (No Service Specified)

**User Request**: "/devstack deploy" OR "deploy all services"

**Actions**:
1. Search helmfile.yaml for all uncommented service entries (lines not starting with #)
2. Parse each service name and namespace
3. Present complete list to user with AskUserQuestion tool:
   ```
   Found X uncommented services in helmfile.yaml:
   - service1 (namespace: ns1) - line 123
   - service2 (namespace: ns2) - line 456
   - service3 (namespace: ns3) - line 789

   Deploy all these services?
   ```
4. If user confirms:
   - Deploy all services in a single helmfile command WITHOUT selector flags
   - Monitor all deployments
5. If user declines:
   - Ask which specific services to deploy

**Commands**:
```bash
# Deploy all uncommented services together (no selector flag)
helmfile -f helmfile.yaml delete || true
helmfile -f helmfile.yaml sync
```

**Example**:
```bash
# User runs: /devstack deploy
# Assistant finds: pg-router, account-service, api (all uncommented)
# After confirmation, deploys all at once without selector
helmfile -f helmfile.yaml delete || true && helmfile -f helmfile.yaml sync
```

### Deploy Multiple Specific Services

**User Request**: "deploy pg-router with abc123, asv with existing, api with def456"

**IMPORTANT**: Deploy all services in a SINGLE helmfile command WITHOUT selector flags!

**Actions**:
1. Parse the request to extract:
   - Service names (pg-router, asv, api)
   - Image requirements (abc123, existing, def456)
2. For each service:
   - Locate in helmfile.yaml
   - **CRITICAL**: If ANY service is commented, AUTOMATICALLY uncomment it
   - **CRITICAL**: Read current `image:` value
   - **UPDATE image field** ONLY if new commit hash provided (abc123, def456)
   - **KEEP existing image** if "existing", "current", or "keep" specified (asv)
   - Report what image will be used for each service
3. **CRITICAL**: Ensure ALL other services in helmfile.yaml are COMMENTED OUT
4. After ALL specified services are uncommented and images updated:
   ```bash
   # CORRECT: Deploy all uncommented services (no selector flag)
   helmfile -f helmfile.yaml delete || true
   helmfile -f helmfile.yaml sync
   ```

**Commands**:
```bash
# Multi-service deployment without selector flag
helmfile -f helmfile.yaml delete || true
helmfile -f helmfile.yaml sync
```

**Example**:
```yaml
# Before update
- name: pg-router-{{ .Values.devstack_label }}
  values:
    - image: old123

- name: account-service-{{ .Values.devstack_label }}
  values:
    - image: old456  # Keep this (user said "existing")

- name: api-{{ .Values.devstack_label }}
  values:
    - image: old789

# After update (only pg-router and api changed)
- name: pg-router-{{ .Values.devstack_label }}
  values:
    - image: abc123  # UPDATED

- name: account-service-{{ .Values.devstack_label }}
  values:
    - image: old456  # UNCHANGED (existing)

- name: api-{{ .Values.devstack_label }}
  values:
    - image: def456  # UPDATED
```

**Deploy Command**:
```bash
# Deploy all uncommented services (pg-router, account-service, api)
helmfile -f helmfile.yaml delete || true
helmfile -f helmfile.yaml sync
```

### Deploy Multiple Services with Same Image

**User Request**: "deploy pg-router, api, asv all with image abc123"

**Actions**:
1. Locate all three services in helmfile.yaml
2. **CRITICAL**: Uncomment pg-router, api, and asv if they are commented
3. **CRITICAL**: Comment out all other services
4. Update all three images to abc123
5. Deploy all together

**Commands**:
```bash
# Deploy all uncommented services (no selector flag)
helmfile -f helmfile.yaml delete || true
helmfile -f helmfile.yaml sync
```

## Common Deployment Patterns

### Deploy Service with Specific Image Tag

**User Request**: "Deploy payment-service with label john using image abc123"

**Actions**:
1. Find service in helmfile.yaml
2. **CRITICAL**: If commented, AUTOMATICALLY uncomment it
3. **CRITICAL**: Ensure all other services are commented out
4. Update image value to abc123
5. Validate configuration
6. Delete existing deployment (ignore failures)
7. Run helmfile sync (deploys only uncommented service)
8. Monitor deployment

**Commands**:
```bash
# Deploy only uncommented services (payment-service in this case)
helmfile -f helmfile.yaml delete || true
helmfile -f helmfile.yaml sync
```

### Deploy Previously Commented Service

**User Request**: "Deploy api-gateway with label alice"

**Actions**:
1. Find commented service in helmfile.yaml
2. **CRITICAL**: AUTOMATICALLY uncomment entire release block
3. **CRITICAL**: Ensure all other services are commented out
4. Add missing configurations (if any)
5. Validate configuration
6. Delete existing deployment (ignore failures)
7. Run helmfile sync (deploys only api-gateway)
8. Report uncommented service to user

**Commands**:
```bash
# Deploy only uncommented service (api-gateway)
helmfile -f helmfile.yaml delete || true
helmfile -f helmfile.yaml sync
```

### Redeploy After Configuration Changes

**User Request**: "Redeploy merchant-service with updated memory limits"

**Actions**:
1. Verify new configuration in values.yaml
2. **CRITICAL**: Ensure merchant-service is uncommented
3. **CRITICAL**: Ensure all other services are commented out
4. Run template validation
5. Delete existing deployment (clean slate)
6. Execute helmfile sync (deploys only merchant-service)
7. Monitor deployment
8. Verify new pods have updated resources

**Commands**:
```bash
# Deploy only uncommented service (merchant-service)
helmfile -f helmfile.yaml delete || true
helmfile -f helmfile.yaml sync
```

**Note**: Delete ensures configurator hooks (DB, SQS, SNS) run with fresh config

## Deployment Flags and Options

### Helmfile Deployment Options

**CRITICAL**: All deployments use NO selector flags. Service selection is done by uncommenting in helmfile.yaml.

```bash
# Standard deployment (DEFAULT - with delete first, no selector)
helmfile -f helmfile.yaml delete || true
helmfile -f helmfile.yaml sync

# One-liner (preferred)
helmfile -f helmfile.yaml delete || true; helmfile -f helmfile.yaml sync

# Update existing (skip delete) - only if user explicitly requests
helmfile -f helmfile.yaml sync

# Force recreate (if pods stuck in bad state)
helmfile -f helmfile.yaml delete || true
helmfile -f helmfile.yaml sync --force

# Skip validation (not recommended)
helmfile -f helmfile.yaml delete || true
helmfile -f helmfile.yaml sync --skip-deps
```

### Service Selection via Uncommenting

**CRITICAL WORKFLOW**:
1. **For deploying specific service(s)**: Uncomment ONLY the target service(s), comment out all others
2. **For deploying all services**: Ensure all desired services are uncommented
3. **Run helmfile without selector flags**: This deploys ALL uncommented services

## Error Scenarios

### Helmfile Sync Fails

**Error**: `Error: release <service> failed`

**Actions**:
1. Check helmfile output for specific error
2. Common causes:
   - Invalid YAML syntax → Fix and retry
   - Missing chart directory → Verify path
   - Invalid values → Check values.yaml
3. Apply fix and retry deployment

### Image Pull Errors During Deployment

**Error**: `Failed to pull image "c.rzp.io/razorpay/<service>:<tag>"`

**Actions**:
1. Verify image tag exists in registry
2. Check image name format
3. Suggest valid image tag
4. Update helmfile.yaml and redeploy

### Namespace Not Found

**Error**: `Error: namespace "<namespace>" not found`

**Actions**:
1. Verify namespace in helmfile.yaml matches cluster
2. Create namespace if missing:
   ```bash
   kubectl create namespace <namespace>
   ```
3. Retry deployment

## Auto-Fix Examples

### Example 1: Missing Resource Limits

**Before** (values.yaml):
```yaml
web_requests_cpu: 50m
web_requests_memory: 50Mi
# Missing limits!
```

**Auto-Fix Applied**:
```yaml
web_requests_cpu: 50m
web_requests_memory: 50Mi
web_limits_memory: 100Mi     # ADDED
# NOTE: CPU limits intentionally NOT added to prevent throttling
```

### Example 2: Commented Service

**Before** (helmfile.yaml):
```yaml
# - name: payment-service-{{ .Values.devstack_label }}
#   namespace: payment-service
#   chart: ./charts/payment-service
```

**Auto-Fix Applied**:
```yaml
- name: payment-service-{{ .Values.devstack_label }}
  namespace: payment-service
  chart: ./charts/payment-service
```

### Example 3: Missing TTL Annotation

**Before** (deployment.yaml):
```yaml
metadata:
  annotations:
    app: payment-service
  # Missing TTL!
```

**Auto-Fix Applied**:
```yaml
metadata:
  annotations:
    app: payment-service
    janitor/ttl: "{{ .Values.ttl }}"    # ADDED
```

## Best Practices

### Always Validate Before Deploy
- Run template validation
- Check configuration against checklist
- Verify image tag exists

### Use Appropriate TTL Values
- `1h` - Short-lived testing
- `8h` - Daily development work
- `forever` - Long-running environments (use sparingly)

### Set Proper Resource Limits
- Start conservative (50m CPU, 100Mi memory)
- Monitor actual usage
- Adjust based on metrics

### Label Consistently
- Use meaningful devstack labels (your username)
- Include labels in all resources
- Use labels for cleanup

## Output Examples

### Successful Deployment

```
## ✅ Deployment Successful!

### What I Did
1. ✅ Found payment-service in helmfile.yaml:892
2. ✅ Validated configuration - all checks passed
3. ✅ Deleted existing deployment (clean slate)
4. ✅ Executed helmfile sync
5. ✅ Deployment completed without errors
6. ✅ Initial health check passed

### Deployment Output
```
# Delete phase
release "payment-service-john" uninstalled

# Sync phase
Installing release=payment-service-john, chart=./charts/payment-service
Release "payment-service-john" has been installed. Happy Helming!
```

### Resources Deployed
**Helm Release**: payment-service-john (INSTALLED)
**Namespace**: payment-service

### Next Steps
Waiting 30 seconds for pods to become ready...
Will automatically monitor pod status and report health.
```

### Deployment with Auto-Fixes

```
## ⚠️ Deployment Completed with Auto-Fixes

### What I Did
1. ✅ Found api-gateway in helmfile.yaml:445
2. ⚠️ Service was commented - AUTOMATICALLY uncommented
3. ⚠️ Missing web_limits_memory - AUTOMATICALLY added: 200Mi
4. ⚠️ Missing TTL annotation - AUTOMATICALLY added
5. ✅ Template validation passed
6. ✅ Deleted existing deployment (clean slate)
7. ✅ Executed helmfile sync
8. ✅ Deployment completed

### Auto-Fixes Applied
- Uncommented service in helmfile.yaml:445-450
- Added web_limits_memory: 200Mi to values.yaml:23
- Added janitor/ttl annotation to deployment.yaml:15

### Deployment Output
```
# Delete phase
release "api-gateway-alice" uninstalled

# Sync phase
Installing release=api-gateway-alice, chart=./charts/api-gateway
Release "api-gateway-alice" has been installed. Happy Helming!
```

### Resources Deployed
**Helm Release**: api-gateway-alice (INSTALLED)
**Namespace**: api-gateway

### Next Steps
Proceeding to monitor pod health...
```

## Related Subskills

- [Validation](validation.md) - Deep-dive into configuration validation
- [Monitoring](monitoring.md) - Post-deployment health checks
- [Debugging](debugging.md) - Troubleshooting failed deployments
