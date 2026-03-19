# Base Pod Readiness

## Purpose

Validates that a service's helm chart in `kube-manifests/helmfile/charts/<service>/` meets all requirements for base pod deployment. Automatically fixes structural issues and raises a PR to kube-manifests. Emits a structured result block consumed by the [Base Pod Pipeline](base-pod-pipeline.md) subskill.

## When to Use

- Before creating a Spinnaker base pod pipeline for a service
- To check if an existing helm chart is base-pod ready
- When base pod deployment is failing due to configuration issues

## Prerequisites

- `kube-manifests` repo available locally (auto-cloned if missing)
- `gh` CLI authenticated for PR creation
- Helm chart must already exist (created via [Onboarding](onboarding.md) subskill)

---

## Readiness Checks

There are 6 checks. Checks 1–5 result in auto-fixes and a kube-manifests PR if issues are found. Check 6 produces a pipeline override string (no file changes).

### Check 1: devstack_label Label

Every Kubernetes resource in `templates/` must have `devstack_label` in its `.metadata.labels` block.

**Inspect**: All `.yaml` files under `helmfile/charts/<service>/templates/` — look for `metadata:` → `labels:` sections.

**Fail condition**: Any resource missing `devstack_label: ...` in its labels.

**Fix**:
```yaml
# In labels block of each resource:
devstack_label: "{{ .Values.devstack_label }}"
```

---

### Check 2: janitor/ttl Annotation

Every Kubernetes resource must have `janitor/ttl` in its `.metadata.annotations` block.

**Inspect**: All `.yaml` files under `templates/` — look for `metadata:` → `annotations:` sections.

**Fail condition**: Any resource missing `janitor/ttl: ...` in its annotations.

**Fix**:
```yaml
# In annotations block of each resource:
janitor/ttl: "{{ .Values.ttl }}"
```

---

### Check 3: No Ephemeral Resources for Base Pods

Ephemeral infrastructure (databases, caches, queues, secrets) must not run when `devstack_label == "base"`. The following template files are ephemeral-only and must be conditionally excluded:

- `db-configurator.yaml`, `db-configmap.yaml`
- `cache-configurator.yaml`, `cache-configmap.yaml`
- `sqs-configurator.yaml`, `sqs-configmap.yaml`
- `sns-configurator.yaml`, `sns-configmap.yaml`
- `secret-cloner.yaml`, `sec-updater.yaml`, `sec-updater-cm.yaml`

**Inspect**: Check if any of the above files exist and are NOT already wrapped in a `devstack_label != "base"` condition.

**Fail condition**: Any of the above files exist without a guard.

**Fix**: Wrap the **entire contents** of each such file with:
```yaml
{{- if ne .Values.devstack_label "base" }}
<existing file contents>
{{- end }}
```

---

### Check 4: Ingress Route URLs Must Not Contain `-base`

For base pods, ingress hostnames must NOT append `-base` to the service name. Only ephemeral deployments should use the `-<devstack_label>` suffix pattern.

**Inspect**: All ingress/edge YAML files (`edge.yaml`, `ingress.yaml`, `ingressroute.yaml`, or any file containing `IngressRoute`) — look for `Host(...)` rules using `{{ .Values.devstack_label }}` in the hostname.

**Fail condition**: Hostname templates of the form:
```
Host(`{{ .Values.name }}-...-{{ .Values.devstack_label }}.dev.razorpay.in`)
```
that have no conditional guard for base pods.

**Fix**: Replace with a conditional hostname:
```yaml
{{- if eq .Values.devstack_label "base" }}
Host(`{{ .Values.name }}<suffix>.dev.razorpay.in`)
{{- else }}
Host(`{{ .Values.name }}<suffix>-{{ .Values.devstack_label }}.dev.razorpay.in`)
{{- end }}
```
Where `<suffix>` is whatever was between the service name and the devstack_label (e.g., `-web`, `-edge`, `-graphql-edge`).

---

### Check 5: No injectheader Middleware on Base Pod Ingress

The `injectheader` middleware injects a `rzpctx-dev-serve-user` header for ephemeral routing. Base pods must not use it.

**Inspect**: All ingress/edge YAML files — look for any `Middleware` resource with `injectheader` in its name or spec, and any `middlewares:` references to it in route definitions.

**Fail condition**: `injectheader` Middleware resource or reference exists without a `ne .Values.devstack_label "base"` guard.

**Fix**: Wrap the Middleware resource definition and all route `middlewares:` references that include it:
```yaml
{{- if ne .Values.devstack_label "base" }}
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: injectheader-...
spec:
  headers:
    customRequestHeaders:
      rzpctx-dev-serve-user: {{ .Values.devstack_label }}
{{- end }}
```
And in route specs:
```yaml
{{- if ne .Values.devstack_label "base" }}
middlewares:
  - name: injectheader-...
{{- end }}
```

---

### Check 6: Replica Counts (Override Only — No File Changes)

Every web and worker deployment must run with ≥ 2 replicas when deployed as a base pod. This is enforced via Spinnaker `default_overrides`, not via chart changes.

**Inspect**: Parse `values.yaml` for all replica-related fields:
- `web_replicas`, `web_base_replicas`
- `worker_replicas`, `worker_base_replicas`
- Any other `*_replicas` fields

**Logic for each replica field**:
1. If a `*_base_replicas` variant exists, use that value for base pod effective count.
2. Otherwise use the plain `*_replicas` value.
3. If effective count < 2 → record that the plain `*_replicas` field needs an override to `2`.

**Output**: Comma-separated override string, e.g.:
```
web_replicas=2,worker_replicas=2
```
This is included in the Structured Result Block — no changes are made to any file.

---

## Workflow

### Phase 1: Locate Chart

```bash
# Auto-detect kube-manifests repo
find ~ -maxdepth 4 -name "kube-manifests" -type d 2>/dev/null | head -1
```

- If not found → clone: `git clone git@github.com:razorpay/kube-manifests.git ~/razorpay/kube-manifests`
- Check `helmfile/charts/<service>/` exists in the repo
- If chart directory does not exist → **BLOCKED**: report error and stop
  ```
  ❌ Chart not found: helmfile/charts/<service>/
  💡 First onboard the service using the /devstack Onboarding subskill
  ```

### Phase 2: Run All 6 Checks

Read all files under `helmfile/charts/<service>/templates/` and `values.yaml`.

For each check, report:
- ✅ PASS — requirement already met
- ❌ NEEDS FIX — will be auto-fixed
- ⚠️ OVERRIDE NEEDED — check 6 only, no file change

### Phase 3: Apply Auto-Fixes (Checks 1–5)

If any checks 1–5 failed, apply the fixes described above in-place.

After fixing, re-read the changed files to confirm the fix is syntactically valid Go template YAML before committing.

### Phase 4: Create kube-manifests PR (If Files Changed)

If any files were modified in Phase 3:

```bash
cd <kube-manifests-root>
git checkout -b base-pod-ready/<service-name>
git add helmfile/charts/<service>/
git commit -m "feat(<service>): make helm chart base-pod ready

- Add devstack_label label to all resources
- Add janitor/ttl annotation to all resources
- Guard ephemeral resources from base pod deployments
- Fix ingress route hostnames (no -base suffix for base pods)
- Remove injectheader middleware for base pod ingress"

git push origin base-pod-ready/<service-name>
gh pr create \
  --repo razorpay/kube-manifests \
  --title "feat(<service>): make helm chart base-pod ready" \
  --body "..." \
  --base master
```

**PR body** should list:
- Each issue found (check name, affected file, line reference)
- The fix applied
- Note that replica count is handled via Spinnaker overrides (not chart changes)

### Phase 5: Emit Structured Result Block

Always emit this block at the end — it is consumed by the [Base Pod Pipeline](base-pod-pipeline.md) subskill:

````
## Readiness Result — base-pod-readiness

checks:
  devstack_label:      <PASS|FIXED>
  janitor_ttl:         <PASS|FIXED>
  ephemeral_resources: <PASS|FIXED>
  ingress_url:         <PASS|FIXED>
  injectheader:        <PASS|FIXED>
  replica_counts:      <PASS|OVERRIDE_NEEDED>

replica_overrides: "<comma-separated overrides or empty string>"
kube_manifests_pr: "<PR URL or none>"
overall: <READY|BLOCKED>
````

**`overall` values**:
- `READY` — chart exists; proceed with pipeline creation (replica overrides and/or PR raised is fine)
- `BLOCKED` — chart does not exist; pipeline cannot be created

---

## Output Report

```
## 🔍 Base Pod Readiness: <service-name>

### Checks

| Check                | Status        | Details                              |
|----------------------|---------------|--------------------------------------|
| devstack_label label | ✅ PASS       |                                      |
| janitor/ttl          | ❌ FIXED      | Added to deployment.yaml, svc.yaml   |
| Ephemeral resources  | ❌ FIXED      | Wrapped db-configurator.yaml         |
| Ingress URL          | ✅ PASS       |                                      |
| injectheader         | ❌ FIXED      | Guarded in edge.yaml                 |
| Replica counts       | ⚠️ OVERRIDE   | web_replicas=1 → override to 2       |

### Fixes Applied
- `templates/svc.yaml`: Added janitor/ttl annotation
- `templates/deployment.yaml`: Added janitor/ttl annotation
- `templates/db-configurator.yaml`: Wrapped in ne devstack_label "base" guard
- `templates/edge.yaml`: Removed injectheader middleware for base pods

### kube-manifests PR
✅ PR raised: https://github.com/razorpay/kube-manifests/pull/<N>

### Pipeline Overrides
⚠️ Replica overrides required: web_replicas=2
   (Will be applied via Spinnaker default_overrides)

### Result
✅ Chart is READY for base pod pipeline creation
```

---

## Edge Cases

- **Resource has no `labels:` block at all**: Create the block, then add `devstack_label`.
- **Resource has no `annotations:` block at all**: Create the block, then add `janitor/ttl`.
- **Ingress files use different hostname patterns**: Match any pattern containing `{{ .Values.devstack_label }}` in a `Host(...)` expression.
- **All checks pass**: No PR is created, proceed directly to pipeline creation.
- **`*_base_replicas` value is already ≥ 2**: No override needed for that deployment.

---

## Related Subskills

- [Onboarding](onboarding.md) — Create the helm chart before validation
- [Base Pod Pipeline](base-pod-pipeline.md) — Consumes this subskill's output to create the Spinnaker pipeline
- [Validation](validation.md) — General chart validation for ephemeral deployments
