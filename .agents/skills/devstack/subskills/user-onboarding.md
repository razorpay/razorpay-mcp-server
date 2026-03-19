# User Onboarding Subskill

Quick onboarding workflow for new developers setting up devstack on their local machine.

## Purpose

Installs and configures devstack CLI tools, clones required repositories, and sets up the environment for developers to start deploying and debugging services.

## When to Use

- New developer joining the team
- Setting up devstack on a new machine
- Resetting devstack environment after issues
- Updating devstack tools to latest version

## Prerequisites

- macOS or Linux machine
- Internet connection
- Terminal access
- curl installed
- GitHub CLI (`gh`) - Required for devstack v2 (will be installed during setup)
- Access to razorpay/devstack-v2 private repository (for devstack v2)

## Onboarding Workflow

Choose between devstack v1 (legacy) or devstack v2 (recommended) installation:

- **Devstack v2 (devstackctl)**: Modern CLI with improved features and performance
- **Devstack v1 (legacy)**: Original devstack installation

### Option A: Install Devstack v2 (devstackctl) - Recommended

Devstack v2 provides a modern CLI experience with improved deployment workflows and better debugging capabilities.

#### Step 1: Install GitHub CLI (Required)

**IMPORTANT**: The devstackctl binary is hosted in a private GitHub repository. You need GitHub CLI (`gh`) to download it.

**Check if GitHub CLI is installed**:
```bash
which gh && gh --version
```

**Install GitHub CLI**:

For macOS (using Homebrew):
```bash
brew install gh
```

For Linux (Debian/Ubuntu):
```bash
sudo apt install gh
```

For Linux (Fedora):
```bash
sudo dnf install gh
```

For other platforms, see: https://github.com/cli/cli#installation

#### Step 2: Authenticate GitHub CLI

**CRITICAL**: You must authenticate with GitHub to access the private razorpay/devstack-v2 repository.

```bash
# Start authentication flow
gh auth login
```

**Authentication Steps**:
1. Select `GitHub.com`
2. Select `HTTPS` as preferred protocol
3. Select `Login with a web browser` (recommended)
4. Copy the one-time code shown in terminal
5. Press Enter to open browser
6. Paste the code and authorize GitHub CLI
7. Return to terminal

**Verify Authentication**:
```bash
# Check auth status
gh auth status

# Verify access to devstack-v2 repo
gh repo view razorpay/devstack-v2 --json name -q '.name'
# Should output: devstack-v2
```

**If you see "Could not resolve to a Repository"**:
- Ensure you have access to the razorpay/devstack-v2 repository
- Contact your team lead to request repository access

#### Step 3: Download and Install devstackctl

**Using GitHub CLI (Recommended)**:

```bash
# Create devstack bin directory
mkdir -p ~/.devstack/bin

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64) ARCH="arm64" ;;
esac

# Download binary using GitHub CLI
gh release download v0.5.0 -R razorpay/devstack-v2 -p "devstackctl-${OS}-${ARCH}" -O ~/.devstack/bin/devstackctl --clobber

# Make executable
chmod +x ~/.devstack/bin/devstackctl

echo "✅ devstackctl installed successfully"
```

**Manual Installation by Platform**:

For macOS ARM64 (M1/M2/M3):
```bash
mkdir -p ~/.devstack/bin
gh release download v0.5.0 -R razorpay/devstack-v2 -p "devstackctl-darwin-arm64" -O ~/.devstack/bin/devstackctl --clobber
chmod +x ~/.devstack/bin/devstackctl
```

For macOS Intel:
```bash
mkdir -p ~/.devstack/bin
gh release download v0.5.0 -R razorpay/devstack-v2 -p "devstackctl-darwin-amd64" -O ~/.devstack/bin/devstackctl --clobber
chmod +x ~/.devstack/bin/devstackctl
```

For Linux ARM64:
```bash
mkdir -p ~/.devstack/bin
gh release download v0.5.0 -R razorpay/devstack-v2 -p "devstackctl-linux-arm64" -O ~/.devstack/bin/devstackctl --clobber
chmod +x ~/.devstack/bin/devstackctl
```

For Linux AMD64:
```bash
mkdir -p ~/.devstack/bin
gh release download v0.5.0 -R razorpay/devstack-v2 -p "devstackctl-linux-amd64" -O ~/.devstack/bin/devstackctl --clobber
chmod +x ~/.devstack/bin/devstackctl
```

**What this does**:
- Creates `~/.devstack/bin` directory for devstack binaries
- Downloads the devstackctl binary from the private GitHub repository
- Makes the binary executable

#### Step 4: Add to PATH

Add the devstack bin directory to your shell PATH:

**For Bash** (add to `~/.bashrc` or `~/.bash_profile`):
```bash
# Add devstack v2 to PATH
export PATH="$HOME/.devstack/bin:$PATH"
```

**For Zsh** (add to `~/.zshrc`):
```bash
# Add devstack v2 to PATH
export PATH="$HOME/.devstack/bin:$PATH"
```

**Apply changes**:
```bash
# For bash
source ~/.bashrc

# For zsh
source ~/.zshrc
```

#### Step 5: Verify Installation

Check that devstackctl is properly installed and accessible:

```bash
# Check if devstackctl is in PATH
which devstackctl
# Should show: /Users/<your-username>/.devstack/bin/devstackctl

# Check version
devstackctl version
# Should show version information

# Test basic command
devstackctl --help
# Should display help information
```

**Expected Output**:
```
✅ devstackctl found at: /Users/yourname/.devstack/bin/devstackctl
✅ Version: v0.5.0
✅ Ready to use!
```

#### Step 6: Configure devstackctl

Initialize devstack v2 configuration:

```bash
# Initialize devstack configuration
devstackctl init

# Set your devstack label
devstackctl config set label <your-username>

# Verify configuration
devstackctl config show
```

#### Step 7: Setup kubectl Access

Devstackctl requires kubectl access to the devstack cluster:

```bash
# Configure kubectl context for devstack
devstackctl setup kubectl

# Verify cluster access
kubectl config current-context
# Should show: dev-serve or devstack-cluster

# Test cluster connectivity
kubectl get nodes
```

### Option B: Install Devstack v1 (Legacy)

For teams still using the legacy devstack installation.

#### Step 1: Run Devstack v1 Installer

Execute the devstack installation script:

```bash
sh -euo pipefail -c "$(curl 'https://get-devstack.dev.razorpay.in/')"
```

**What this does**:
- Downloads the latest devstack CLI tools
- Installs kubectl configuration for devstack cluster
- Sets up necessary aliases and functions
- Configures shell environment

#### Step 2: Source Shell Configuration

Load the devstack environment into your current shell:

```bash
source ~/.devstack/shrc
```

**What this does**:
- Loads devstack functions and aliases
- Sets up kubectl context for devstack cluster
- Configures environment variables
- Enables devstack commands in current terminal

#### Step 3: Verify v1 Installation

Check that devstack is properly configured:

```bash
# Check kubectl context
kubectl config current-context
# Should show: devstack-cluster or similar

# Check cluster connectivity
kubectl get nodes

# Check helmfile is available
helmfile --version
```

## Common Steps (Required for Both v1 and v2)

### Step 1: Clone kube-manifests Repository

**CRITICAL**: You need the kube-manifests repository to create and deploy helm charts.

```bash
# Navigate to your work directory
cd ~/Documents/Work/rzp-repos  # or your preferred location

# Clone kube-manifests repository
git clone git@github.com:razorpay/kube-manifests.git

# Navigate to helmfile directory
cd kube-manifests/helmfile

# Verify helmfile.yaml exists
ls helmfile.yaml
# Should output: helmfile.yaml
```

**What this does**:
- Clones the central kube-manifests repository
- Gives you access to all helm charts in `helmfile/charts/`
- Provides the helmfile.yaml for deployments

**Expected Output**:
```
Cloning into 'kube-manifests'...
remote: Enumerating objects: 15000, done.
remote: Counting objects: 100% (500/500), done.
✅ Repository cloned successfully
```

### Step 2: Configure Devstack Skill Path

Update the devstack skill configuration to point to your kube-manifests location:

```bash
# From your working directory, update the config
# Replace the path with your actual kube-manifests location
echo '{"helmfile_directory": "'$(pwd)'/kube-manifests/helmfile", "auto_detect": true}' > claude-skills/infrastructure/skills/devstack/config.json
```

**Verify configuration**:
```bash
cat claude-skills/infrastructure/skills/devstack/config.json
# Should show your helmfile directory path
```

## Post-Installation Setup

### Configure Your Devstack Label

Set your personal devstack label (usually your username):

```bash
# This should be done automatically by the installer
# But verify it's set correctly
echo $DEVSTACK_LABEL
```

If not set, add to your shell profile:
```bash
export DEVSTACK_LABEL="your-username"
```

### Test Deployment

Try deploying a simple service to verify everything works:

```bash
# Navigate to helmfile directory
cd ~/Documents/Work/rzp-repos/kube-manifests/helmfile  # or your kube-manifests location

# List available services
helmfile list

# Deploy a test service (uncomment a service in helmfile.yaml first)
helmfile -f helmfile.yaml -l name=test-service-$DEVSTACK_LABEL sync
```

## Troubleshooting

### Devstack v2 (devstackctl) Issues

#### Issue: devstackctl command not found

**Error**:
```
bash: devstackctl: command not found
```

**Fix**:
1. Verify the binary is in the correct location:
   ```bash
   ls -la ~/.devstack/bin/devstackctl
   ```

2. Check if PATH is set correctly:
   ```bash
   echo $PATH | grep -q ".devstack/bin" && echo "✅ PATH configured" || echo "❌ PATH not configured"
   ```

3. Add to PATH manually:
   ```bash
   export PATH="$HOME/.devstack/bin:$PATH"

   # Make permanent by adding to shell profile
   echo 'export PATH="$HOME/.devstack/bin:$PATH"' >> ~/.bashrc  # or ~/.zshrc
   source ~/.bashrc  # or source ~/.zshrc
   ```

#### Issue: Download fails with 404 error

**Error**:
```
curl: (22) The requested URL returned error: 404
```

**Fix**:
1. Verify the version exists: Check https://github.com/razorpay/devstack-v2/releases
2. Update the version in the download URL to the latest available
3. Ensure you're using the correct architecture (darwin-arm64, darwin-amd64, linux-arm64, linux-amd64)

**Check your architecture**:
```bash
uname -s  # Shows OS (Darwin for macOS, Linux for Linux)
uname -m  # Shows architecture (arm64, x86_64)
```

#### Issue: Permission denied when executing devstackctl

**Error**:
```
bash: /Users/username/.devstack/bin/devstackctl: Permission denied
```

**Fix**:
```bash
# Make the binary executable
chmod +x ~/.devstack/bin/devstackctl

# Verify permissions
ls -l ~/.devstack/bin/devstackctl
# Should show: -rwxr-xr-x (executable permission)
```

#### Issue: Binary architecture mismatch

**Error**:
```
bad CPU type in executable
# or
cannot execute binary file: Exec format error
```

**Fix**:
Download the correct binary for your architecture:

```bash
# Check your architecture
uname -m

# For arm64 (M1/M2/M3 Mac):
curl -L https://github.com/razorpay/devstack-v2/releases/download/v0.5.0/devstackctl-darwin-arm64 -o ~/.devstack/bin/devstackctl

# For x86_64 (Intel Mac):
curl -L https://github.com/razorpay/devstack-v2/releases/download/v0.5.0/devstackctl-darwin-amd64 -o ~/.devstack/bin/devstackctl

# Make executable
chmod +x ~/.devstack/bin/devstackctl
```

#### Issue: devstackctl version shows older version

**Fix**:
```bash
# Remove old binary
rm ~/.devstack/bin/devstackctl

# Download latest version (update v0.5.0 to latest)
curl -L https://github.com/razorpay/devstack-v2/releases/download/v0.5.0/devstackctl-darwin-arm64 -o ~/.devstack/bin/devstackctl
chmod +x ~/.devstack/bin/devstackctl

# Verify new version
devstackctl version
```

### Devstack v1 (Legacy) Issues

#### Issue: Installation script fails to download

**Error**:
```
curl: (6) Could not resolve host: get-devstack.dev.razorpay.in
```

**Fix**:
1. Check internet connection
2. Verify VPN is connected (if required)
3. Try again after a few minutes

#### Issue: kubectl context not set (v1)

**Error**:
```
error: current-context is not set
```

**Fix**:
```bash
# Re-source the shell config
source ~/.devstack/shrc

# Or manually set context
kubectl config use-context devstack-cluster
```

#### Issue: Permission denied errors (v1)

**Error**:
```
Permission denied: ~/.devstack/shrc
```

**Fix**:
```bash
# Fix permissions
chmod +x ~/.devstack/shrc
chmod -R u+w ~/.devstack/
```

#### Issue: Command not found after sourcing (v1)

**Error**:
```
bash: devstack: command not found
```

**Fix**:
1. Verify installation completed successfully
2. Re-run the installer
3. Check shell profile is loading devstack:
   ```bash
   grep devstack ~/.bashrc  # or ~/.zshrc
   ```

### Common Issues (Both v1 and v2)

#### Issue: Cluster access denied

**Error**:
```
Error from server (Forbidden): nodes is forbidden
forbidden: User "dev@razorpay.com" cannot list resource "secrets" in API group "" in the namespace "app"
```

**Root Cause**: RBAC permissions not provisioned for dev-serve cluster

**Fix**:

**Option 1: Run Cluster Access Pipeline (Recommended)**
1. Go to https://deploy.razorpay.com/#/applications/devserve-infra/executions
2. Find pipeline named "Cluster access"
3. Click "Start Manual Execution"
4. Wait for pipeline to complete (~2-5 minutes)
5. Verify access:
   ```bash
   kubectl get secrets -n app
   kubectl get pods --all-namespaces
   ```

**Option 2: Request Onboarding**
1. Contact your team lead or DevOps team
2. Request completion of onboarding flow for your user
3. This will provision all necessary RBAC permissions
4. Wait for confirmation
5. Verify access as shown above

**Validation Steps**:
```bash
# Check current cluster context
kubectl config current-context
# Should be: dev-serve

# Test permissions
kubectl auth can-i list secrets -n app
kubectl auth can-i get pods --all-namespaces
kubectl auth can-i create deployments -n <namespace>

# All should return "yes" after provisioning

# Verify kubeconfig credentials
kubectl config view
```

**If Still Experiencing Issues**:
1. Verify you're on the correct cluster:
   ```bash
   kubectl config get-contexts
   kubectl config use-context dev-serve
   ```
2. Check if your user is in the correct LDAP/AD group
3. Contact DevOps team for manual RBAC provisioning
4. Ensure pipeline completed successfully without errors

## Persistent Shell Configuration

To make devstack available in all new terminal sessions:

### For Devstack v2 (devstackctl)

**Bash** (`~/.bashrc`):
```bash
# Devstack v2 (devstackctl) - Add binary to PATH
export PATH="$HOME/.devstack/bin:$PATH"
```

**Zsh** (`~/.zshrc`):
```bash
# Devstack v2 (devstackctl) - Add binary to PATH
export PATH="$HOME/.devstack/bin:$PATH"
```

### For Devstack v1 (Legacy)

**Bash** (`~/.bashrc`):
```bash
# Devstack v1 environment
if [ -f ~/.devstack/shrc ]; then
    source ~/.devstack/shrc
fi
```

**Zsh** (`~/.zshrc`):
```bash
# Devstack v1 environment
if [ -f ~/.devstack/shrc ]; then
    source ~/.devstack/shrc
fi
```

After adding, reload your shell:
```bash
source ~/.bashrc  # or source ~/.zshrc
```

## What Gets Installed

### Devstack v2 (devstackctl)

1. **devstackctl Binary**
   - Modern CLI tool for devstack operations
   - Single binary with all functionality
   - Location: `~/.devstack/bin/devstackctl`

2. **kubectl Configuration** (via devstackctl setup)
   - Context: devstack cluster
   - Credentials: User-specific access
   - Location: `~/.kube/config`

3. **Configuration**
   - User preferences and defaults
   - Stored in devstackctl config format

### Devstack v1 (Legacy)

1. **kubectl Configuration**
   - Context: devstack cluster
   - Credentials: User-specific access
   - Location: `~/.kube/config`

2. **Devstack CLI Tools**
   - Helper functions for deployment
   - Aliases for common operations
   - Location: `~/.devstack/`

3. **Helmfile Configuration**
   - Default values for TTL, labels
   - Service templates
   - Location: Varies by team setup

4. **Shell Functions**
   - `devstack-deploy` - Quick deployment wrapper
   - `devstack-logs` - Stream logs from pods
   - `devstack-cleanup` - Delete your deployments

## Quick Reference Commands

### Devstack v2 Commands

After installation, these commands become available:

```bash
# Check devstackctl version
devstackctl version

# Show configuration
devstackctl config show

# Deploy a service
devstackctl deploy <service-name> --label <your-label>

# List deployments
devstackctl list --label <your-label>

# Get logs
devstackctl logs <service-name> --label <your-label>

# Cleanup resources
devstackctl cleanup --label <your-label>

# Get help
devstackctl --help
```

### Post-Deployment: Display All IngressRoute Access URLs (Devstack v2)

**CRITICAL**: After every successful devstackctl deployment, ALWAYS fetch and display ALL IngressRoute rules so users know how to access their service.

**Command to Fetch All IngressRoutes**:
```bash
# Get all IngressRoute specs for the deployed service
kubectl get ingressroute -n <namespace> -l devstack_label=<label> --context dev-serve -o yaml
```

**Example IngressRoute Spec**:
```yaml
spec:
  entryPoints:
  - http
  routes:
  - kind: Rule
    match: Host(`pg-router.dev.razorpay.in`) && Headers (`rzpctx-dev-serve-user`,`parag`)
    middlewares: []
    services:
    - name: pg-router-parag
      port: 80
  - kind: Rule
    match: Host(`pg-router.dev.razorpay.in`) && Headers (`rzpctx-dev-serve-user`,`parag`)
    services:
    - name: pg-router-parag
      port: 81
  - kind: Rule
    match: Host(`pg-router-parag.dev.razorpay.in`)
    middlewares:
    - name: injectheader-parag
    services:
    - name: pg-router-parag
      port: 80
  - kind: Rule
    match: Host(`pg-router-parag.dev.razorpay.in`)
    middlewares:
    - name: injectheader-parag
    services:
    - name: pg-router-parag
      port: 81
```

**Required Output Format**:

After deployment, display ALL routes in this format:

```
### Access URLs

| Type | URL | Port | Header Required |
|------|-----|------|-----------------|
| With Header | `https://pg-router.dev.razorpay.in` | 80 | `rzpctx-dev-serve-user: parag` |
| With Header | `https://pg-router.dev.razorpay.in` | 81 | `rzpctx-dev-serve-user: parag` |
| Direct Access | `https://pg-router-parag.dev.razorpay.in` | 80 | None (auto-injected) |
| Direct Access | `https://pg-router-parag.dev.razorpay.in` | 81 | None (auto-injected) |

### Quick Access Examples

# Using header-based routing (port 80)
curl -H "rzpctx-dev-serve-user: parag" https://pg-router.dev.razorpay.in/health

# Using header-based routing (port 81)
curl -H "rzpctx-dev-serve-user: parag" https://pg-router.dev.razorpay.in:81/health

# Using direct URL (no header needed)
curl https://pg-router-parag.dev.razorpay.in/health
```

**Route Types**:

1. **Header-Based Routes** (`Host(...) && Headers(...)`):
   - Shared domain: `<service>.dev.razorpay.in`
   - Requires header: `rzpctx-dev-serve-user: <label>`
   - Multiple users share the same domain

2. **Direct Access Routes** (`Host(<service>-<label>.dev.razorpay.in)`):
   - Dedicated domain: `<service>-<label>.dev.razorpay.in`
   - No header required (middleware auto-injects)
   - Easier for browser testing

**IMPORTANT**: Parse ALL routes from `spec.routes[]` and show:
- Every unique host
- Every port
- Whether header is required or not

### Devstack v1 Commands

After installation, these commands become available:

```bash
# Get your current devstack label
echo $DEVSTACK_LABEL

# Switch to devstack kubectl context
kubectl config use-context devstack-cluster

# List your deployments
kubectl get pods -A -l devstack_label=$DEVSTACK_LABEL

# Clean up your resources
kubectl delete all -A -l devstack_label=$DEVSTACK_LABEL
```

## Next Steps

After successful onboarding:

1. **Clone your service repository**
   ```bash
   git clone <your-service-repo>
   ```

2. **Deploy your first service**
   - See [deployment.md](deployment.md) for deployment workflow

3. **Learn debugging**
   - See [debugging.md](debugging.md) for troubleshooting

4. **Understand helm charts**
   - See [onboarding.md](onboarding.md) for creating charts

## Security Notes

- **Never share your kubeconfig**: Contains your personal credentials
- **Use TTLs appropriately**: Resources auto-cleanup based on TTL
- **Label everything**: Use your devstack label for all resources
- **Don't deploy to base**: Only deploy to your personal label

## Automation Script

For automated setup in CI/CD or scripts:

```bash
#!/bin/bash
set -euo pipefail

echo "🚀 Installing devstack..."

# Run installer
sh -euo pipefail -c "$(curl 'https://get-devstack.dev.razorpay.in/')"

# Source environment
source ~/.devstack/shrc

# Verify installation
if kubectl config current-context | grep -q devstack; then
    echo "✅ Devstack installed successfully"
    echo "Current context: $(kubectl config current-context)"
    echo "Devstack label: $DEVSTACK_LABEL"
else
    echo "❌ Installation failed"
    exit 1
fi
```

## Uninstallation

### Uninstall Devstack v2

To remove devstack v2 from your machine:

```bash
# Remove devstackctl binary
rm -f ~/.devstack/bin/devstackctl

# Remove devstack directory (if no other devstack files)
rm -rf ~/.devstack

# Remove kubectl context
kubectl config delete-context devstack-cluster

# Remove from shell profile
# Edit ~/.bashrc or ~/.zshrc and remove the PATH export line:
# export PATH="$HOME/.devstack/bin:$PATH"
```

### Uninstall Devstack v1

To remove devstack v1 from your machine:

```bash
# Remove devstack directory
rm -rf ~/.devstack

# Remove kubectl context
kubectl config delete-context devstack-cluster

# Remove from shell profile
# Edit ~/.bashrc or ~/.zshrc and remove devstack section
```

## Support

If you encounter issues:

1. Check this troubleshooting guide
2. Ask in #devstack-support Slack channel
3. Contact DevOps team
4. Check internal wiki for latest updates

## Version Information

The installer always fetches the latest version from `get-devstack.dev.razorpay.in`.

To check your current version:
```bash
devstack --version 2>/dev/null || echo "Version info not available"
```

---

**Version**: 1.0.0
**Last Updated**: 2026-01-22
**Related**: [deployment.md](deployment.md), [debugging.md](debugging.md)
