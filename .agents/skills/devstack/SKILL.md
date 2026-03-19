---
name: devstack
description: Deploy and debug helmfile-based services with automated configuration validation, intelligent troubleshooting, and autonomous error recovery. Use when (1) Setting up devstack for new developers, (2) Onboarding new applications to devstack, (3) Creating helm charts and ephemeral resources (DB, cache, SQS, SNS), (4) Deploying helmfile services with devstack labels, (5) Debugging pod failures and crashes, (6) Validating helmfile/helm chart configurations, (7) Troubleshooting ImagePullBackOff, CrashLoopBackOff, OOMKilled errors, (8) Auto-fixing resource limits, probes, and configuration issues, (9) Post-deployment health monitoring and recovery, (10) Setting up devspace for live code sync and debugging.
version: "1.0.2"
category: "Infrastructure"
author: "razorpay"
disable-model-invocation: true
---

# Devstack Deployment & Debug Assistant

Comprehensive toolset for deploying and debugging helmfile-based Kubernetes services with automated validation, intelligent troubleshooting, and autonomous error recovery.

## What This Skill Does

Provides end-to-end automation for helmfile deployments:

1. **Pre-Deployment Validation**
   - Configuration validation against best practices
   - Automatic fixing of common issues (resource limits, probes, TTL)
   - Template rendering validation
   - Service availability checks

2. **Autonomous Deployment**
   - Automated helmfile sync operations
   - Real-time deployment monitoring
   - Rollback on critical failures

3. **Post-Deployment Debugging**
   - Automatic pod health monitoring
   - Intelligent log analysis and pattern detection
   - Root cause identification for common errors
   - Self-healing for fixable issues (OOMKilled, config errors)

4. **Error Recovery**
   - Automatic retry with fixes applied
   - Memory limit auto-scaling
   - Configuration auto-correction
   - Detailed remediation steps for complex issues

## Prerequisites

- **kube-manifests repository cloned locally**
  ```bash
  # If not already cloned:
  git clone git@github.com:razorpay/kube-manifests.git
  cd kube-manifests/helmfile
  ```
- kubectl configured with cluster access
- helmfile installed
- Helmfile directory configured (see [Configuration](#configuration) below)

## Configuration

The devstack skill needs to know where your helmfile directory is located. You have three options:

### Option 1: Use Auto-Detection (Recommended)
The skill will automatically search for `kube-manifests/helmfile` from your repository root. No configuration needed if your project follows this structure.

### Option 2: Update config.json
Edit `agent-skills/infrastructure/skills/devstack/config.json` and set your helmfile path:

```json
{
  "helmfile_directory": "/path/to/your/kube-manifests/helmfile",
  "auto_detect": true
}
```

### Option 3: Quick Setup Command
Run this command to automatically configure the path for your current repository:

```bash
# From your repository root
echo "{\"helmfile_directory\": \"$(pwd)/kube-manifests/helmfile\", \"auto_detect\": true}" > agent-skills/infrastructure/skills/devstack/config.json
```

**Current Configuration**:
- Path: Check `agent-skills/infrastructure/skills/devstack/config.json`
- To verify: The skill will report the path it's using when you run any deployment command

**Need Help?** See the detailed [Setup Guide (SETUP.md)](SETUP.md) for:
- Step-by-step configuration instructions
- Common scenarios and solutions
- Troubleshooting path issues
- Team configuration best practices

## ⚠️ CRITICAL: Helm Chart Location

**When creating or updating helm charts, ALWAYS use the `helmfile/charts/<application-name>` directory within the kube-manifests repository.**

This is the ONLY valid location for helm charts in devstack. Never create charts in any other location.

## When to Use This Skill

Use this skill when you need to:
- **Setup devstack for a new developer** (install CLI tools, configure kubectl)
- **Setup devspace for live code sync** (sync local code to pods, stream logs)
- Onboard a new application to devstack
- Create helm charts for a new service (in `helmfile/charts/<service-name>/` only)
- Set up ephemeral databases, caches, queues, or SNS topics
- Configure SQS queues and SNS pub/sub messaging
- Configure secrets for applications
- Deploy a service to devstack environment
- **Debug with live code updates** (rapid iteration without rebuilds)
- Troubleshoot failing pods (CrashLoopBackOff, ImagePullBackOff, OOMKilled)
- Validate helmfile/helm chart configurations
- Perform health checks on deployed services
- Auto-fix common deployment issues
- Get detailed deployment reports with access URLs
- **Onboard a service as a long-running base pod** (Spinnaker pipeline via spinacode)

## Quick Start

### Setup Devstack for New Developer
```
/devstack

Setup devstack on my machine for the first time
```

Or for devstack v2:
```
/devstack

Install devstack v2
```

### Onboard a New Application
```
/devstack

Onboard new-service to devstack with ephemeral database
```

### Deploy a Single Service
```
/devstack

Deploy payment-service with devstack label john
```

### Deploy Multiple Services
```
/devstack

deploy pg-router with abc123, asv with existing image, api with def456
```

### Deploy All Uncommented Services
```
/devstack

deploy
```
(Will show all uncommented services and ask for confirmation)

### Debug Existing Deployment
```
/devstack

Debug pods in namespace payment-service with label john
```

### Validate Configuration
```
/devstack

Validate configuration for service payment-service
```

### Setup Devspace Code Sync
```
/devstack

Help me set up devspace for terminals service with label alice
```

### Create Base Pod Pipeline
```
/devstack

Create base pod pipeline for payment-service
```

## Subskills

This skill consists of specialized subskills for different aspects of helmfile management:

### 1. [Deployment](subskills/deployment.md)
**When to use**: Deploying new services or updating existing ones
- Pre-deployment validation
- Helmfile sync execution
- Service discovery and uncommenting
- Image tag verification

### 2. [Debugging](subskills/debugging.md)
**When to use**: Troubleshooting failing or crashing pods
- Pod status analysis (CrashLoopBackOff, ImagePullBackOff, OOMKilled, etc.)
- Log analysis (current and previous containers)
- Event inspection
- Root cause identification

### 3. [Validation](subskills/validation.md)
**When to use**: Checking configurations before deployment
- values.yaml validation
- deployment.yaml validation
- helmfile.yaml validation
- Best practices compliance

### 4. [Monitoring](subskills/monitoring.md)
**When to use**: Post-deployment health checks and ongoing monitoring
- Pod health status
- Resource usage tracking
- Service endpoint verification
- Real-time log streaming

### 5. [Onboarding](subskills/onboarding.md)
**When to use**: Onboarding new applications to devstack
- Helm chart creation from scratch
- Setting up ephemeral databases, caches, queues, and SNS topics
- Configuring secrets management
- Creating deployment, service, and ingress resources
- End-to-end application onboarding workflow

### 6. [User Onboarding](subskills/user-onboarding.md)
**When to use**: Setting up devstack for a new developer
- Installing devstack v2 (devstackctl) or v1 (legacy) CLI tools
- Configuring kubectl context for devstack cluster
- Setting up shell environment and PATH
- Verifying installation and cluster access
- First-time developer setup
- Supports both devstack v2 (modern) and v1 (legacy) installations

### 7. [Devspace Code Sync](subskills/devspace.md)
**When to use**: Live code synchronization and debugging
- Setting up devspace for local development
- Syncing local code changes to running pods
- Real-time log streaming for debugging
- Rapid iteration without rebuilding images
- Language-specific configurations (Golang, PHP, Node.js)

### 8. [Base Pod Readiness](subskills/base-pod-readiness.md)
**When to use**: Checking if a service's helm chart is ready for base pod deployment
- Validates `devstack_label` label on all resources
- Validates `janitor/ttl` annotation on all resources
- Detects ephemeral resources that must be excluded for base pods
- Validates ingress route URLs (no `-base` suffix for base pods)
- Validates no `injectheader` middleware on base pod ingress
- Checks replica counts for all deployments (≥2 required); outputs override string for pipeline
- Auto-fixes structural issues and raises kube-manifests PR; emits structured result block for pipeline subskill

### 9. [Base Pod Pipeline](subskills/base-pod-pipeline.md)
**When to use**: Creating a Spinnaker pipeline to deploy a service as a long-running base pod on devstack
- Invokes base-pod-validation and reads its structured result block
- Builds `default_overrides` dynamically (base values + replica overrides from validation)
- Generates the `deploy-to-devserve.json` pipeline config in spinacode
- Creates spinacode PR; links to kube-manifests PR if chart fixes were needed
- Handles repo cloning for both spinacode and kube-manifests if not found locally

## Complete Workflow Example

```
User: "Deploy payment-service with label john"

Assistant (Autonomous Flow):

Phase 1: Pre-Deployment Validation ✅
- Located service in helmfile.yaml (line 892)
- Automatically uncommented service
- Validated configuration
- Fixed missing web_limits_memory: 200Mi
- Template validation passed

Phase 2: Clean Deployment ✅
- Deleted existing deployment (clean slate)
- Deployed via helmfile sync
- No errors during deployment

Phase 3: Post-Deployment Monitoring ⏳
- Waiting 30s for pods to start...
- Checking pod status...
- Status: CrashLoopBackOff ❌

Phase 4: Autonomous Debugging 🔍
- Retrieved pod logs
- Identified: Database connection error
- Root Cause: MySQL service unavailable
- Provided fix instructions

Phase 5: Final Report 📋
✅ Configuration fixes applied
✅ Deployment successful
❌ Pods failing due to DB dependency
💡 Next steps provided with commands
```

## How It Works

### Autonomous Behavior

The skill operates autonomously:
- **No permission needed** for diagnostic commands (logs, events, describe)
- **Auto-fixes simple issues** (resource limits, TTL, probes)
- **Only asks for input** when truly ambiguous (e.g., which service to deploy)

### Thoroughness

For every deployment:
- Check pod status after deployment
- Get logs for every non-running pod
- Check events for every problem pod
- Analyze root cause before suggesting fixes

### Clean Deployments (Default)

By default, all deployments use delete-before-sync for clean slate:
- **Always deletes** existing deployment before sync (configurable)
- **Ignores failures** - delete errors won't block deployment
- **Re-runs hooks** - DB/SQS/SNS configurators execute fresh
- **Fresh secrets** - Secrets regenerated with latest values
- **Prevents conflicts** - Avoids issues from resource type changes

**Benefits**:
- ✅ No stale resources from previous deployments
- ✅ All configurations applied from scratch
- ✅ Configurator hooks re-execute with current config
- ✅ Eliminates "it works on my machine" issues

**To skip delete**: User must explicitly say "update existing" or set `delete_before_sync: false` in config.json

### Clarity

All reports include:
- Clear status indicators (✅ ❌ ⚠️)
- Explanation of "why" not just "what"
- Exact commands to run for verification
- File paths with line numbers

## Reference Documentation

The skill uses comprehensive reference guides:

1. **[FAQ](references/faq.md)** - Frequently asked questions and answers
2. **[Configuration Checklist](references/config-checklist.md)** - Required fields and best practices
3. **[Error Patterns](references/error-patterns.md)** - Common errors and solutions
4. **[Auto-Fix Strategies](references/auto-fix-strategies.md)** - What gets fixed automatically
5. **[Recovery Workflows](references/recovery-workflows.md)** - Step-by-step recovery procedures
6. **[Path Detection](references/path-detection.md)** - Helmfile directory detection logic
7. **[Helm Chart Templates](references/helm-chart-templates.md)** - Complete template reference including SNS/SQS
8. **[SNS Configurator Guide](references/sns-configurator-guide.md)** - Complete guide for SNS topics and pub/sub messaging

## Common Use Cases

### Deploy Single Service
```
Deploy api-gateway with label alice
```
Auto-handles: validation, deployment, health checks, troubleshooting

### Deploy Multiple Services Together
```
deploy pg-router with abc123, asv with existing, api with def456
```
Auto-handles:
- Parse multiple service specifications
- Update only specified images
- Deploy all services in single helmfile command
- Monitor all deployments together

### Deploy All Uncommented Services
```
/devstack deploy
```
Auto-handles:
- Find all uncommented services in helmfile.yaml
- Present list to user for confirmation
- Deploy all services together
- Monitor all deployments

### Debug Crashing Pods
```
Why are pods crashing in namespace payment-service with label bob?
```
Auto-retrieves: status, events, logs, root cause analysis

### Fix OOMKilled Pods
```
Pods are being OOMKilled in user-service namespace with label charlie
```
Auto-applies: memory limit increase, redeploy, monitor

### Validate Before Deploy
```
Validate configuration for merchant-service before deploying
```
Auto-checks: values.yaml, deployment.yaml, helmfile.yaml

## Output Structure

All operations provide structured reports:

```
## 🔍 Deployment Status: [SUCCESS/FAILED/PARTIAL]

### What I Did
1. ✅ Action completed
2. ❌ Action failed

### Issues Found
**Issue Name**
- Root Cause: Explanation
- Evidence: Logs/Events
- Fix Applied: Auto-fix OR Fix Needed: Manual steps

### Resources Deployed
**Pods**: pod-name (STATUS)
**Services**: service-name (TYPE - IP:PORT)
**Access URLs**: http://...

### Next Steps
- [ ] Step 1
- [ ] Step 2

### Debug Commands
```bash
# Commands for manual verification
```
```

## Edge Cases Handled

- **Multiple pods failing**: Checks each individually
- **Partial deployments**: Reports working vs failing pods
- **Long startup times**: Waits up to 5 minutes with progress checks
- **Missing secrets**: Provides exact creation commands
- **Configuration errors**: Auto-fixes or provides detailed steps

## Limitations

- Cannot auto-create secrets (security restriction)
- Cannot fix application-level bugs
- Requires kubectl access to cluster
- Limited to helmfile-based deployments

## Frequently Asked Questions

For detailed answers to common questions, see the [FAQ Reference](references/faq.md).

**Quick answers to top questions:**

**Q: How to keep deployments up for more than 72 hours?**
- Whitelist your label in `devstack-config/whitelist_labels.yaml`
- Restart `ttl-validation-webhook` after PR merge
- Ensure `devstack_label` label is present on all resources

**Q: How to make an endpoint available externally?**
- Create IngressRoute with `kubernetes.io/ingress.class: traefik-external`
- Use domain pattern: `*.ext.dev.razorpay.in`
- No Route53 needed, but endpoint will be publicly accessible

**Q: What are devstack NAT IPs?**
- `52.66.76.68`, `52.66.95.207`, `13.127.201.109`
- Use for IP whitelisting external services

**Q: How do I debug ImagePullBackOff?**
- Verify image exists via Harbor API
- Check image tag in helmfile.yaml
- Ensure commit ID is correct

**Q: Can I run multiple services with the same label?**
- Yes, deploy all services with the same `devstack_label`
- They share the same lifecycle and cleanup

**More questions?** See the [complete FAQ](references/faq.md)

## Troubleshooting

**Getting RBAC permission errors?**
- Error: `forbidden: User "..." cannot list resource "secrets"`
- Solution: Run cluster access pipeline at https://deploy.razorpay.com/#/applications/devserve-infra/executions
- Pipeline Name: "Cluster access"
- Wait 2-5 minutes and verify with: `kubectl auth can-i list secrets -n app`
- See [User Onboarding](subskills/user-onboarding.md) for detailed steps

**Skill not auto-fixing issues?**
- Check if issue is in auto-fix list (see references/auto-fix-strategies.md)
- Complex issues require manual intervention

**Need more detailed logs?**
- Specify: "Show me detailed logs for pod X"
- Increase tail lines: "Get last 200 lines of logs"

**Want to skip validation?**
- Not recommended, but specify: "Deploy without validation"

## Version

**Current**: 1.0.2 (2026-02-04)

See [CHANGELOG.md](CHANGELOG.md) for version history.

## Additional Resources

- [Deployment Subskill](subskills/deployment.md) - Detailed deployment workflows
- [Debugging Subskill](subskills/debugging.md) - Comprehensive debugging guide
- [Validation Subskill](subskills/validation.md) - Configuration validation details
- [Monitoring Subskill](subskills/monitoring.md) - Monitoring and health check commands
- [Devspace Code Sync](subskills/devspace.md) - Live code synchronization and debugging guide

## Need More Help?

**Devstack Documentation:**
- 📖 https://alpha.razorpay.com/repo/devstack-docs
- Comprehensive guides, examples, and best practices

**Slack Support:**
- 💬 Channel: `#platform-devstack`
- Get help from the devstack team
- Ask questions, share feedback, report issues

**Internal Resources:**
- Check the [FAQ](references/faq.md) for common questions
- Review [Error Patterns](references/error-patterns.md) for troubleshooting
- Use this `/devstack` skill for automated assistance
