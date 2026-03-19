# Base Pod Pipeline

## Purpose

Creates a Spinnaker pipeline in the `spinacode` repository to deploy a service as a long-running base pod on devstack. Base pods run the production commit, are monitored, and are deployed exclusively via Spinnaker for audit trail.

Orchestrates the full workflow:
1. Validates helm chart readiness (via [Base Pod Readiness](base-pod-readiness.md))
2. Reads validation output to build dynamic pipeline overrides
3. Generates `deploy-to-devserve.json` pipeline config
4. Raises a PR to `razorpay/spinacode`

## When to Use

- Setting up a new service to run as a base pod on devstack
- The service's helm chart already exists in `kube-manifests/helmfile/charts/<service>/`

## Prerequisites

- `kube-manifests` repo available locally (auto-cloned if missing)
- `spinacode` repo available locally (auto-cloned if missing)
- `gh` CLI authenticated for PR creation
- `uuidgen` or Python available for UUID generation

---

## Inputs

Collect the following from the user before proceeding. Show defaults where available:

| Field | Description | Default |
|---|---|---|
| `service_name` | Service to onboard (must match helm chart directory name) | — |
| `github_repo_name` | GitHub repository name | Same as `service_name` |
| `namespace` | Kubernetes namespace | Same as `service_name` |
| `commit_txt_host` | Hostname for service | `<service>.dev.razorpay.in` |
| `slack_channel` | Slack channel for deployment notifications | `tech_deployments` |
| `slack_group_id` | Slack group ID for `@here`-style mentions (e.g., `S12345678`) | — |
| `slack_group_handle` | Slack group handle for display (e.g., `team-payments`) | — |

Ask for all fields in a single prompt. Confirm defaults with user before proceeding.

---

## Workflow

### Phase 1: Collect Inputs

Prompt the user for the fields in the table above. Present defaults inline. Example prompt:

```
To create the base pod pipeline for <service>, I need a few details:

Required:
- service_name: [provided]
- slack_group_id: (e.g. S12345678)
- slack_group_handle: (e.g. team-payments)

Using defaults (confirm or override):
- github_repo_name: <service>
- namespace: <service>
- service_host: <service>.dev.razorpay.in
- slack_channel: tech_deployments
```

---

### Phase 2: Invoke Base Pod Validation

Run the [Base Pod Readiness](base-pod-readiness.md) subskill for the service.

**After validation completes, read the Structured Result Block** — specifically these fields:

```
replica_overrides: "<value>"      # e.g., "web_replicas=2,worker_replicas=2" or ""
kube_manifests_pr: "<value>"      # e.g., PR URL or "none"
overall: <READY|BLOCKED>
```

- If `overall = BLOCKED` → **abort**. The chart does not exist; direct user to run `/devstack Onboarding` first.
- If `overall = READY` → proceed with pipeline creation.

---

### Phase 3: Build `default_overrides` String

Construct the `default_overrides` value for the pipeline JSON:

**Base (always included)**:
```
devstack_label=base,ttl=forever,image=${ parameters.image_tag }
```

**If `replica_overrides` is non-empty** (from the validation result block), append:
```
,<replica_overrides>
```

**Examples**:

| replica_overrides from validation | Resulting default_overrides |
|---|---|
| `""` (empty) | `devstack_label=base,ttl=forever,image=${ parameters.image_tag }` |
| `"web_replicas=2"` | `devstack_label=base,ttl=forever,image=${ parameters.image_tag },web_replicas=2` |
| `"web_replicas=2,worker_replicas=2"` | `devstack_label=base,ttl=forever,image=${ parameters.image_tag },web_replicas=2,worker_replicas=2` |

---

### Phase 4: Locate / Clone spinacode Repo

```bash
# Auto-detect
find ~ -maxdepth 4 -name "spinacode" -type d 2>/dev/null | head -1
```

If not found:
```bash
git clone git@github.com:razorpay/spinacode.git ~/razorpay/spinacode
```

---

### Phase 5: Check for Existing Pipeline

```bash
ls <spinacode-root>/v3/<service-name>/dev-serve/mum-rspl/deploy-to-devserve.json
```

If the file already exists:
```
⚠️ Pipeline already exists: v3/<service>/dev-serve/mum-rspl/deploy-to-devserve.json
Do you want to overwrite it? (yes/no)
```

If user says no → abort with instructions to review the existing file manually.

---

### Phase 6: Generate Pipeline JSON

**Generate UUID**:
```bash
python3 -c "import uuid; print(uuid.uuid4())"
# or
uuidgen | tr '[:upper:]' '[:lower:]'
```

**Create directory and write file**:
```bash
mkdir -p <spinacode-root>/v3/<service-name>/dev-serve/mum-rspl/
```

Write the following JSON to `v3/<service-name>/dev-serve/mum-rspl/deploy-to-devserve.json`, substituting all placeholders:

```json
{
    "application": "devstack-mum-rspl-<service_name>",
    "exclude": [],
    "id": "<generated-uuid>",
    "index": 0,
    "keepWaitingPipelines": false,
    "limitConcurrent": true,
    "locked": {
        "ui": true,
        "allowUnlockUi": false,
        "description": "Note: Pipeline Templates govern this pipeline. UI edits are blocked for consistency."
    },
    "name": "Deploy base pods on DevStack",
    "notifications": [],
    "parameterConfig": [],
    "stages": [],
    "type": "templatedPipeline",
    "schema": "v2",
    "template": {
        "artifactAccount": "front50ArtifactCredentials",
        "reference": "spinnaker://75cb9c2f-bb26-42e2-91f8-2b16c5445a3b:latest",
        "type": "front50/pipelineTemplate"
    },
    "triggers": [],
    "variables": {
        "application": "devstack-mum-rspl-<service_name>",
        "commit_txt_host": "<service_host>",
        "default_overrides": "<default_overrides_string>",
        "github_repo_name": "<github_repo_name>",
        "helm_chart_overrides_file": "razorpay/kube-manifests/contents/helmfile/charts/<service_name>/values.yaml",
        "helm_chart_path_prefix": "helmfile/<service_name>/<service_name>-1-",
        "kube_manifests_bucket_names": "{\"Mumbai\":[{\"value\":\"rzp-kube-manifests\"}]}",
        "logs_host": "https://razorpay-non-prod.app.coralogix.in",
        "namespace": "<namespace>",
        "pre_deploy_notification_text": "<!subteam^<slack_group_id>|<slack_group_handle>>\nDevstack base <service_name> Deployment initiated by ${ trigger.user }\n Old Commit: ${ old_commit }\n New Commit : ${ parameters.image_tag }",
        "region": "Mumbai",
        "service_name": "<service_name>",
        "slack_channel": "<slack_channel>",
        "sleeve_host": "http://spin-sleeve.spinnaker:8080",
        "user_groups_mentions": "<!subteam^<slack_group_id>|<slack_group_handle>>"
    }
}
```

**Substitution map**:

| Placeholder | Value |
|---|---|
| `<service_name>` | User-provided `service_name` |
| `<generated-uuid>` | UUID generated in this phase |
| `<commit_txt_host>` | User-provided `service_host` |
| `<default_overrides_string>` | String built in Phase 3 |
| `<github_repo_name>` | User-provided `github_repo_name` |
| `<namespace>` | User-provided `namespace` |
| `<slack_group_id>` | User-provided `slack_group_id` |
| `<slack_group_handle>` | User-provided `slack_group_handle` |
| `<slack_channel>` | User-provided `slack_channel` |

---

### Phase 7: Create spinacode PR

```bash
cd <spinacode-root>
git checkout -b devstack-base-pod/<service-name>
git add v3/<service-name>/
git commit -m "feat(<service-name>): add devstack base pod pipeline"
git push origin devstack-base-pod/<service-name>
gh pr create \
  --repo razorpay/spinacode \
  --title "feat(<service-name>): add devstack base pod Spinnaker pipeline" \
  --body "$(cat <<'EOF'
## Summary

Adds a Spinnaker pipeline to deploy `<service-name>` as a base pod on devstack (dev-serve / Mumbai).

## Pipeline Details

- **Spinnaker application**: `devstack-mum-rspl-<service-name>`
- **Namespace**: `<namespace>`
- **Helm chart**: `helmfile/charts/<service-name>/values.yaml`
- **default_overrides**: `<default_overrides_string>`
- **Slack channel**: `<slack_channel>`

## Replica Overrides

<IF replica_overrides non-empty>
⚠️ The helm chart has deployments with <2 replicas. The following overrides are applied via `default_overrides` to ensure base pod stability:
`<replica_overrides>`

<ELSE>
✅ All deployments have ≥2 replicas configured. No replica overrides needed.
</IF>

## Related PRs

<IF kube_manifests_pr != "none">
🔗 kube-manifests PR (chart base-pod readiness fixes): <kube_manifests_pr>
<ELSE>
✅ Helm chart was already base-pod ready. No kube-manifests changes needed.
</IF>

## Test Plan

- [ ] Trigger the pipeline in Spinnaker UI after merging
- [ ] Verify base pods come up in namespace `<namespace>` with label `devstack_label=base`
- [ ] Confirm TTL is `forever` and pods persist across janitor cycles
EOF
)" \
  --base master
```

---

## Output Report

```
## ✅ Base Pod Pipeline Created: <service-name>

### What I Did
1. ✅ Collected inputs from user
2. ✅ Validated helm chart (base-pod-readiness)
3. ✅ Built default_overrides with replica overrides: <value or "none needed">
4. ✅ Generated deploy-to-devserve.json
5. ✅ Created spinacode PR

### Pipeline Configuration
- Application: devstack-mum-rspl-<service>
- Namespace: <namespace>
- default_overrides: <value>

### Pull Requests
- spinacode PR: <URL>
- kube-manifests PR: <URL or "none">

### Next Steps
- [ ] Get spinacode PR reviewed and merged
- [ ] If kube-manifests PR was raised, get it merged first
- [ ] After both PRs are merged, trigger the pipeline in Spinnaker:
      https://deploy.razorpay.com/#/applications/devstack-mum-rspl-<service>/executions
- [ ] Verify base pods are running: kubectl get pods -n <namespace> -l devstack_label=base
```

---

## Error Cases

| Error | Cause | Resolution |
|---|---|---|
| `BLOCKED: chart not found` | Helm chart doesn't exist | Run `/devstack Onboarding` first |
| `spinacode clone failed` | No SSH access to GitHub | Ensure SSH key is set up for razorpay org |
| `Pipeline already exists` | Duplicate creation attempt | Review existing file; overwrite only if intentional |
| `gh pr create failed` | Not authenticated or no fork | Run `gh auth login` and verify repo access |

---

## Related Subskills

- [Base Pod Readiness](base-pod-readiness.md) — Called by this subskill; validates chart and provides replica overrides
- [Onboarding](onboarding.md) — Creates the helm chart before this pipeline can be set up
- [Deployment](deployment.md) — For ephemeral pod deployments (not base pods)
