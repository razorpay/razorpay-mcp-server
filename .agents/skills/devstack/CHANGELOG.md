# Devstack Skill Changelog

## Version 1.0.2 (2026-02-04)

### New Feature: Devstack v2 (devstackctl) Support

**Overview**: Added comprehensive support for installing and configuring devstack v2 (devstackctl), the modern CLI for devstack operations.

#### 1. Installation Workflow

**GitHub CLI Requirement** (Private Repository):
- devstackctl binary is hosted in private `razorpay/devstack-v2` repository
- GitHub CLI (`gh`) required for authenticated download
- Installation steps:
  1. Install GitHub CLI (`brew install gh` / `apt install gh`)
  2. Authenticate with `gh auth login`
  3. Download using `gh release download`

**Platform Support**:
- macOS ARM64 (M1/M2/M3 Macs)
- macOS Intel (x86_64)
- Linux ARM64
- Linux AMD64

**Installation Command**:
```bash
# Using GitHub CLI (works with private repo)
gh release download v0.5.0 -R razorpay/devstack-v2 -p "devstackctl-darwin-arm64" -O ~/.devstack/bin/devstackctl --clobber
chmod +x ~/.devstack/bin/devstackctl
```

#### 2. Post-Deployment IngressRoute Display

After every successful `devstackctl deploy`, the skill displays all IngressRoute access URLs:

**Output Format**:
```
### Access URLs

| Type | URL | Port | Header Required |
|------|-----|------|-----------------|
| With Header | `https://pg-router.dev.razorpay.in` | 80 | `rzpctx-dev-serve-user: parag` |
| With Header | `https://pg-router.dev.razorpay.in` | 81 | `rzpctx-dev-serve-user: parag` |
| Direct Access | `https://pg-router-parag.dev.razorpay.in` | 80 | None (auto-injected) |
| Direct Access | `https://pg-router-parag.dev.razorpay.in` | 81 | None (auto-injected) |
```

**Route Types**:
- **Header-Based Routes**: Shared domain requiring `rzpctx-dev-serve-user` header
- **Direct Access Routes**: Dedicated domain with auto-injected headers

#### 3. Key Features

1. **Complete Installation Flow**
   - Step 1: Install GitHub CLI
   - Step 2: Authenticate GitHub CLI
   - Step 3: Download devstackctl binary
   - Step 4: Add to PATH
   - Step 5: Verify installation
   - Step 6: Configure devstackctl (`devstackctl init`)
   - Step 7: Setup kubectl access

2. **User Triggers**
   - "install devstack v2"
   - "setup devstack v2"
   - "setup devstackctl"
   - "install devstackctl"

3. **Troubleshooting Coverage**
   - GitHub CLI installation issues
   - Authentication failures
   - Repository access denied
   - Command not found errors
   - Permission denied issues
   - Architecture mismatch errors

#### 4. Binary Details

- Version: v0.5.0
- Repository: `razorpay/devstack-v2` (private)
- Installation Path: `~/.devstack/bin/devstackctl`
- Shell Configuration: `export PATH="$HOME/.devstack/bin:$PATH"`

### Files Updated
- `subskills/user-onboarding.md` - Complete devstack v2 installation workflow with GitHub CLI, IngressRoute display section
- `SKILL.md` - Added devstack v2 installation example, updated version
- `CHANGELOG.md` - Documented devstack v2 support

### Impact
1. Users can install devstack v2 (devstackctl) from private GitHub repository
2. GitHub CLI authentication enables secure access to private releases
3. All IngressRoute access URLs displayed after deployment
4. Clear distinction between devstack v1 (legacy) and v2 (modern)
5. Comprehensive troubleshooting for installation issues

---

## Version 1.0.1 (2026-02-02)

### Breaking Change: CPU Limits Removed

**Rationale**: CPU limits cause unnecessary throttling even when CPU is available on the node, leading to performance degradation. Removing CPU limits allows applications to burst and use available CPU resources without artificial restrictions.

**Changes Made**:
- Removed CPU limit configurations from all helm chart templates
- Updated documentation to explain why CPU limits are not used
- Modified auto-fix strategies to NOT add CPU limits
- Updated validation checklists to reflect best practice of no CPU limits

**Impact**:
- Applications will no longer be throttled by CPU limits
- Better performance during CPU bursts
- More efficient resource utilization on nodes
- CPU requests still in place to guarantee minimum CPU allocation

**Migration Guide**:
- Existing deployments: No action required, CPU limits can remain if already deployed
- New deployments: Will automatically be created without CPU limits
- To remove CPU limits from existing charts: Delete `web_limits_cpu` and `worker_limits_cpu` from values.yaml and remove from deployment templates

### Bug Fixes & Enhancements

**RBAC Permission Error Handling**
- Added error pattern detection for RBAC permission denied errors
- Pattern: `forbidden: User "<user@email.com>" cannot list resource "secrets" in API group "" in the namespace "<namespace>"`
- Auto-detection of cluster access issues
- Guided workflow for cluster access provisioning
- Integration with cluster access pipeline at deploy.razorpay.com
- Validation commands for verifying RBAC permissions
- Cluster context verification (dev-serve)

**User Experience Improvements**
- Removed incorrect "Expected Output" sections from user onboarding documentation
- Clear guidance on running cluster access pipeline
- Alternative onboarding flow for users without pipeline access
- Step-by-step validation after access provisioning
- Better error messages for permission issues

### Documentation Updates

**Helm Chart Location Clarification**
- Added critical warnings throughout documentation emphasizing that helm charts MUST be created/updated in `helmfile/charts/<application-name>` within the kube-manifests repository ONLY
- Updated onboarding subskill with prominent location requirements
- Updated helm-chart-templates reference with path specifications
- Updated path-detection reference to include chart location guidance
- Updated main SKILL.md with critical location warning

**kube-manifests Repository Cloning**
- Added instructions to clone kube-manifests repository if not present
- Updated path detection to prompt users to clone repo when not found
- Added kube-manifests cloning to user onboarding workflow
- Updated prerequisites sections to mention repository requirement

### Files Updated
- `SKILL.md` - Added critical location warning, kube-manifests prerequisite, and RBAC troubleshooting
- `CHANGELOG.md` - Consolidated version history
- `subskills/onboarding.md` - Added critical location section, kube-manifests cloning instructions, and repository check in Phase 0
- `subskills/user-onboarding.md` - Added Step 4 for cloning kube-manifests, Step 5 for config setup, enhanced cluster access troubleshooting, removed incorrect expected outputs
- `subskills/debugging.md` - Added RBAC permission denied section with automatic debug actions and report format
- `references/helm-chart-templates.md` - Removed CPU limits from deployment templates, added explanatory notes, added critical location warning
- `references/config-checklist.md` - Updated resource requirements to exclude CPU limits
- `references/auto-fix-strategies.md` - Removed CPU limit auto-fixes, added explanation
- `references/error-patterns.md` - Added RBAC permission denied pattern, validation commands, updated quick reference table
- `subskills/validation.md` - Updated validation rules to not require CPU limits
- Template examples updated to show memory limits only

### Impact
This update ensures that:
1. Users encountering RBAC permission errors get clear, actionable guidance
2. CPU limits are no longer enforced, improving application performance
3. Users are prompted to clone kube-manifests repository if it's not present
4. Helm charts are consistently created in the correct location within kube-manifests
5. New developers have clear guidance on repository setup during onboarding
6. Documentation accurately reflects actual command outputs

## Version 1.0.0 (2026-01-23)

Initial release with comprehensive helmfile deployment and debugging capabilities.

### Core Features

**1. Autonomous Deployment**
- Pre-deployment validation and auto-fixing
- Service selection via uncommenting in helmfile.yaml
- Deployment without selector flags for reliability
- Clean slate deployments (delete-before-sync by default)
- Multi-service deployment support

**2. Image Management**
- Automatic image update workflow (commit ID → helmfile.yaml)
- Pre-deployment image validation via Harbor API
- Complete image validation (all containers, workers, sidecars)
- Architecture verification (linux/amd64)

**3. Service Management**
- Auto-uncomment services mentioned in deployment request
- Auto-comment services NOT mentioned (prevents unintended deployments)
- Service discovery from helmfile.yaml

**4. Intelligent Debugging**
- Automatic pod health monitoring
- Root cause analysis for common errors
- Log analysis (current and previous containers)
- Event inspection and pattern detection
- Self-healing for fixable issues (OOMKilled, config errors)

**5. Configuration Validation**
- values.yaml validation against best practices
- deployment.yaml validation
- helmfile.yaml validation
- Auto-fix for common issues (resource limits, probes, TTL)

**6. Post-Deployment Monitoring**
- Real-time pod health status
- Resource usage tracking
- Service endpoint verification
- Detailed deployment reports with access URLs

**7. Developer Onboarding**
- Setup devstack for new developers
- Install CLI tools and configure kubectl
- First-time developer machine setup

**8. Application Onboarding**
- Helm chart creation from scratch
- Ephemeral resources setup (DB, cache, SQS, SNS)
- Secrets management configuration
- Complete deployment resource templates

**9. Devspace Code Sync**
- Live code synchronization to running pods
- Real-time log streaming for debugging
- Rapid iteration without rebuilding images
- Language-specific configurations (Golang, PHP, Node.js)

**10. Knowledge Base**
- Comprehensive FAQ with common questions
- Error pattern detection and solutions
- Auto-fix strategies documentation
- Recovery workflow guides
- Helm chart templates reference

### Subskills

1. **Deployment** - Autonomous deployment with validation and monitoring
2. **Debugging** - Intelligent troubleshooting for failing pods
3. **Validation** - Configuration validation and auto-fixing
4. **Monitoring** - Post-deployment health checks
5. **Onboarding** - Application onboarding to devstack
6. **User Onboarding** - Developer machine setup
7. **Devspace** - Live code sync and debugging

### Reference Documentation

- FAQ with TTL management, external access, NAT IPs
- Configuration checklist for best practices
- Error patterns with common errors and solutions
- Auto-fix strategies for automated remediation
- Recovery workflows for complex issues
- Path detection logic for helmfile directory
- Helm chart templates including SNS/SQS
- SNS configurator guide for pub/sub messaging

### Prerequisites

- kubectl configured with cluster access
- helmfile installed
- Helmfile directory configured (auto-detection supported)

### Configuration

Three options for helmfile directory setup:
1. Auto-detection from repository root (kube-manifests/helmfile)
2. Update config.json with custom path
3. Quick setup command for current repository

### Help Resources

- Devstack Documentation: https://alpha.razorpay.com/repo/devstack-docs
- Slack Support: #platform-devstack
- Internal FAQ and troubleshooting guides
