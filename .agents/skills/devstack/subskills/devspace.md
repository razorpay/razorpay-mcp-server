# Devspace Code Sync Subskill

Live code synchronization and debugging for devstack deployments using devspace.

## Purpose

Enables real-time local code synchronization to pods running in the Kubernetes cluster, allowing developers to:
- Edit code locally and see changes immediately in the pod
- Stream pod logs to local terminal for debugging
- Skip the build-push-deploy cycle during development
- Debug applications with live code updates

## When to Use

- Rapid local development with immediate feedback
- Debugging issues in the cluster environment
- Testing code changes without rebuilding images
- Iterating quickly on features or bug fixes

## Prerequisites

Before setting up devspace, ensure:

1. **Application-Side Changes Complete**
   - Repository has been configured for devspace (one-time setup)
   - `devspace.yaml` file exists in the repository root
   - Application supports hot-reload or auto-restart (optional but recommended)

2. **Services Deployed on Devstack**
   - You have deployed services on devstack with a **unique devstack_label**
   - Pods are running and healthy
   - You know the namespace and service name

3. **Devspace Binary Installed**
   - If you ran the Devstack Onboarding script, devspace is already installed
   - Otherwise, install devspace CLI manually

## Workflow

### Step 1: Verify Prerequisites

#### 1.1. Check Devspace Installation

```bash
# Verify devspace is installed
devspace version

# Expected output: devspace version X.X.X
```

**If not installed**:
- Run the Devstack Onboarding script, or
- Install manually: [devspace installation guide](https://devspace.sh/docs/getting-started/installation)

#### 1.2. Verify Deployment Exists

```bash
# Check if your service is deployed
kubectl get pods -n <namespace> -l devstack_label=<your-label>

# Should show running pods
```

### Step 2: Configure devspace.yaml

#### 2.1. Locate devspace.yaml

```bash
# Navigate to repository root
cd <path_to_your_repo>

# Verify devspace.yaml exists
ls -la devspace.yaml
```

**If devspace.yaml doesn't exist**:
- Copy from reference examples (see References section below)
- Create using `devspace init` command

#### 2.2. Update Variables Section

**CRITICAL**: Update the `vars` section in `devspace.yaml` with your deployment details:

**Required Variables**:
- `devstack_label`: Your unique devstack label (same as in helmfile.yaml)
- `namespace`: Kubernetes namespace where service is deployed
- `app`: Application/service name

**Example Configuration**:

```yaml
# devspace.yaml
version: v2beta1
name: my-service

vars:
  devstack_label: parag     # ← YOUR devstack label from helmfile
  namespace: terminals      # ← Service namespace
  app: terminals            # ← Application name

deployments:
  - name: ${app}-${devstack_label}
    kubectl:
      manifests:
        - kube-manifests/deployment.yaml

dev:
  terminals-${devstack_label}:
    labelSelector:
      app: ${app}
      devstack_label: ${devstack_label}
    namespace: ${namespace}

    # Code sync configuration
    sync:
      - path: ./
        excludePaths:
          - .git/
          - .devspace/
          - vendor/
          - node_modules/
        uploadExcludePaths:
          - Dockerfile*
          - .dockerignore
        downloadExcludePaths:
          - '*'

    # Port forwarding (optional)
    ports:
      - port: "8080:8080"

    # Open terminal in container (optional)
    terminal:
      enabled: true

    # Log streaming
    logs:
      enabled: true
```

#### 2.3. Verify Configuration

**Check these fields match your deployment**:
1. `devstack_label` matches the label in helmfile.yaml
2. `namespace` matches the service namespace
3. `app` matches the service name
4. `labelSelector` matches pod labels

**Validation Command**:
```bash
# Ensure devspace can find your pod
kubectl get pods -n <namespace> -l app=<app>,devstack_label=<label>
```

### Step 3: Start Devspace Code Sync

#### 3.1. Navigate to Repository Root

```bash
cd <path_to_your_repo>

# Example for terminals service
cd ~/repos/terminals/
```

#### 3.2. Start Devspace Development Mode

```bash
# Start devspace with code sync
devspace dev --no-warn
```

**What This Does**:
1. ✅ Connects to the remote pod
2. ✅ Syncs local files to the container
3. ✅ Streams pod logs to your terminal
4. ✅ Watches for file changes and auto-syncs
5. ✅ Opens terminal session in container (if configured)

#### 3.3. Expected Output

```
info Using namespace 'terminals'
info Using kube context 'devstack'
done Loaded config from devspace.yaml

dev:terminals-parag Waiting for pod to become ready...
dev:terminals-parag Selected pod terminals-parag-7d8f9c5b4-xk2m9

dev:terminals-parag Sync: Initializing sync...
dev:terminals-parag Sync: Scanning local and remote file trees...
dev:terminals-parag Sync: Starting bi-directional sync...
dev:terminals-parag Sync: Successfully started sync

dev:terminals-parag Logs: Streaming logs...
[Pod logs appear here...]
```

### Step 4: Development Workflow

#### 4.1. Make Code Changes Locally

```bash
# Edit files in your local repository
vim src/main.go
# or use your IDE
```

**Devspace will automatically**:
- Detect file changes
- Sync changed files to the pod
- Show sync progress in terminal

#### 4.2. Monitor Logs

- Pod logs stream to your terminal in real-time
- See application output, errors, and debug messages
- Use logs to verify your changes are working

#### 4.3. Test Changes

Changes are **immediately available** in the pod:
- For interpreted languages (PHP, Python, Node.js): Changes take effect immediately
- For compiled languages (Go): May need to restart process (see Golang Notes below)

### Step 5: Stop Devspace

To stop code sync and log streaming:

```bash
# Press Ctrl+C in the terminal running devspace dev
^C

# Devspace will:
# - Stop file sync
# - Stop log streaming
# - Exit gracefully
```

**Pod continues running** after devspace stops - only the sync connection ends.

## Language-Specific Notes

### Golang Applications

**IMPORTANT**: Golang requires special handling for dependencies:

**Adding New Dependencies**:
```bash
# 1. Add dependency to go.mod locally
go get github.com/some/package

# 2. MUST push commit to GitHub before using
git add go.mod go.sum
git commit -m "Add new dependency"
git push origin feature-branch

# 3. Rebuild vendor packages (triggers on pod)
# The pod needs to fetch the new dependency from GitHub
```

**Why?**
- Devspace syncs files, but vendor packages are built from GitHub
- New dependencies in `go.mod` require rebuilding vendor
- Pod pulls dependencies from GitHub, not from local sync

**Workaround**:
- For rapid iteration, vendor the dependency locally and sync the entire vendor directory
- Or use `go mod vendor` and include vendor in sync paths

### PHP Applications

**Hot Reload**: PHP changes take effect immediately (no restart needed)

```yaml
# devspace.yaml for PHP
sync:
  - path: ./
    excludePaths:
      - .git/
      - vendor/  # Don't sync vendor, use composer install in pod
```

**Composer Dependencies**:
```bash
# Run composer install inside the pod
devspace enter
composer install
```

### Node.js Applications

**Hot Reload**: Use nodemon or similar for auto-restart

```yaml
# devspace.yaml for Node.js
sync:
  - path: ./
    excludePaths:
      - node_modules/
      - .git/
```

**NPM Dependencies**:
```bash
# Install dependencies in pod
devspace enter
npm install
```

## Common Use Cases

### Use Case 1: Quick Bug Fix

```bash
# 1. Deploy service if not already deployed
/devstack deploy my-service with label alice

# 2. Navigate to repository
cd ~/repos/my-service

# 3. Start devspace
devspace dev --no-warn

# 4. Edit code locally (fix appears in pod immediately)
vim src/handler.php

# 5. Test via API calls
curl http://my-service.devstack.local/api/test

# 6. Monitor logs in real-time
# [Logs stream in terminal]

# 7. Done - press Ctrl+C to stop sync
```

### Use Case 2: Feature Development

```bash
# 1. Start devspace
devspace dev --no-warn

# 2. Make incremental changes
# - Edit files
# - See changes in pod
# - Check logs
# - Iterate quickly

# 3. When satisfied, commit and push
git add .
git commit -m "New feature"
git push

# 4. Rebuild image for final deployment
# (devspace was just for rapid iteration)
```

### Use Case 3: Debugging Production Issues

```bash
# 1. Deploy with production-like config
/devstack deploy service with label debug-prod

# 2. Start devspace to stream logs
devspace dev --no-warn

# 3. Reproduce issue
curl http://service/api/problematic-endpoint

# 4. See detailed logs in real-time
# 5. Make fixes locally, test immediately
# 6. Verify fix works before committing
```

## Configuration Examples

### Example 1: Golang Service (Terminals)

```yaml
# devspace.yaml for terminals service
version: v2beta1
name: terminals

vars:
  devstack_label: parag
  namespace: terminals
  app: terminals

dev:
  terminals-${devstack_label}:
    labelSelector:
      app: ${app}
      devstack_label: ${devstack_label}
    namespace: ${namespace}

    sync:
      - path: ./
        excludePaths:
          - .git/
          - .devspace/
          - vendor/
          - tmp/

    logs:
      enabled: true

    terminal:
      enabled: true
```

**Reference**: [Terminals devspace.yaml](https://github.com/razorpay/terminals/blob/master/devspace.yaml)

### Example 2: PHP Service (API)

```yaml
# devspace.yaml for api service
version: v2beta1
name: api

vars:
  devstack_label: john
  namespace: api
  app: api

dev:
  api-${devstack_label}:
    labelSelector:
      app: ${app}
      devstack_label: ${devstack_label}
    namespace: ${namespace}

    sync:
      - path: ./app
        containerPath: /var/www/html/app
      - path: ./config
        containerPath: /var/www/html/config

    ports:
      - port: "8080:80"

    logs:
      enabled: true
```

**Reference**: [API devspace.yaml](https://github.com/razorpay/api/blob/master/devspace.yaml)

## Troubleshooting

### Devspace Can't Find Pod

**Error**: `No pod found matching selector`

**Fix**:
1. Verify pod is running:
   ```bash
   kubectl get pods -n <namespace> -l devstack_label=<label>
   ```
2. Check `vars` in devspace.yaml match deployment
3. Ensure `labelSelector` matches pod labels

### Sync Not Working

**Error**: Files not syncing to pod

**Fix**:
1. Check `excludePaths` - ensure you're not excluding needed files
2. Verify file permissions
3. Restart devspace:
   ```bash
   devspace purge
   devspace dev --no-warn
   ```

### Golang Dependency Issues

**Error**: `cannot find package`

**Fix**:
1. **Push go.mod changes to GitHub first**
2. Rebuild vendor in pod:
   ```bash
   devspace enter
   go mod vendor
   ```

### Connection Drops

**Error**: Devspace disconnects frequently

**Fix**:
1. Check network connection
2. Increase timeout:
   ```yaml
   dev:
     timeout: 600  # 10 minutes
   ```
3. Check pod health (may be restarting)

### Port Forward Fails

**Error**: Cannot bind to port

**Fix**:
1. Port already in use locally:
   ```bash
   lsof -ti:8080 | xargs kill -9
   ```
2. Update port in devspace.yaml:
   ```yaml
   ports:
     - port: "8081:8080"  # Use different local port
   ```

## Best Practices

### 1. Use Unique Labels
- Always use your personal devstack_label (e.g., your name)
- Prevents conflicts with other developers
- Easier to track your deployments

### 2. Exclude Unnecessary Files
```yaml
excludePaths:
  - .git/
  - .devspace/
  - node_modules/
  - vendor/
  - tmp/
  - *.log
```

### 3. Commit Before Using Devspace
- Commit your work before starting devspace
- Easier to track what changed during sync session
- Can revert if sync causes issues

### 4. Don't Rely on Devspace for Final Testing
- Devspace is for rapid iteration
- Always rebuild and redeploy for final testing
- Image used in production should match actual build

### 5. Monitor Resource Usage
- Devspace sync uses network bandwidth
- Syncing large files repeatedly can be slow
- Use `excludePaths` to skip large directories

### 6. Use Terminal Session
```yaml
terminal:
  enabled: true
```
Allows running commands inside pod while sync is active

## Limitations

- **Cannot sync binary files efficiently**: Large binaries should be excluded
- **Network dependent**: Requires stable connection to cluster
- **Not for production**: Only use in development environments
- **Language constraints**:
  - Golang requires GitHub push for new dependencies
  - Compiled languages may need manual rebuild/restart
- **Resource intensive**: Watching many files can consume resources

## Advanced Configuration

### Selective Sync Paths

```yaml
sync:
  - path: ./src
    containerPath: /app/src
  - path: ./config
    containerPath: /app/config
```

### Bidirectional Sync

```yaml
sync:
  - path: ./
    downloadExcludePaths:
      - '*'  # Don't download anything from container
```

### Multiple Containers

```yaml
dev:
  app-${devstack_label}:
    labelSelector:
      app: ${app}
    container: app-container  # Specify container name

  worker-${devstack_label}:
    labelSelector:
      app: ${app}
    container: worker-container
```

### Custom Commands on Sync

```yaml
dev:
  app-${devstack_label}:
    sync:
      - path: ./

    # Run command when sync starts
    command: ["npm", "run", "dev"]
```

## References

### Official Documentation
- [Devspace Documentation](https://devspace.sh/docs)
- [Devspace Configuration Reference](https://devspace.sh/docs/configuration/reference)

### Example Repositories
- [Terminals Service (Golang)](https://github.com/razorpay/terminals/blob/master/devspace.yaml)
- [API Service (PHP)](https://github.com/razorpay/api/blob/master/devspace.yaml)

### Related Skills
- [Deployment Subskill](deployment.md) - Deploy services first
- [Debugging Subskill](debugging.md) - Debug pod issues
- [Monitoring Subskill](monitoring.md) - Monitor pod health

## Quick Reference

### Start Devspace
```bash
cd <repo>
devspace dev --no-warn
```

### Stop Devspace
```bash
Ctrl+C
```

### Enter Pod Terminal
```bash
devspace enter
```

### Purge Devspace Cache
```bash
devspace purge
```

### Update Configuration
```bash
vim devspace.yaml
# Update vars section
devspace dev --no-warn
```

### Check Status
```bash
devspace list deployments
devspace list sync
```
