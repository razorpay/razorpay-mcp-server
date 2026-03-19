# Infrastructure Security

## When This Applies

- Writing Dockerfiles or container configurations
- Deploying to Kubernetes
- Configuring CI/CD pipelines
- Setting up cloud infrastructure (AWS, GCP, Azure)
- Managing dependencies and packages (npm, pip, etc.)
- Writing Infrastructure as Code (Terraform, CloudFormation)
- Configuring build and deployment processes

## Key Rules

### MUST Do (Containers)

- Run containers as non-root user
- Use minimal base images (Alpine, distroless)
- Pin image versions to specific digests, not just tags
- Drop all capabilities and add only what's needed
- Set filesystem to read-only where possible
- Scan images for vulnerabilities before deployment
- Never expose Docker socket to containers
- Use multi-stage builds to minimize final image size

### MUST Do (CI/CD)

- Never hardcode secrets in code or CI config files
- Use secrets management (Vault, AWS Secrets Manager)
- Require code review and approval before merge
- Sign commits and verify signatures
- Use protected branches for main/production
- Run security scanning (SAST, DAST, dependency scanning)
- Require manual approval for production deployments
- Log all pipeline activities

### MUST NOT Do

- Run containers with `--privileged` flag
- Use `latest` tag for production images
- Expose Docker daemon TCP socket
- Store secrets in environment variables visible in logs
- Allow auto-merge without review
- Skip dependency vulnerability scanning
- Use shared credentials across pipelines

### SHOULD Do

- Use rootless Docker/Podman when possible
- Implement network policies in Kubernetes
- Use Pod Security Standards (Restricted level)
- Scan for misconfigurations with tools like Trivy, Checkov
- Generate and verify SBOMs
- Use short-lived, scoped credentials
- Implement supply chain security (SLSA)

## Safe Patterns

### Secure Dockerfile

```dockerfile
# SECURITY: Use specific version, not :latest
FROM python:3.11-slim-bookworm@sha256:abc123... AS builder

WORKDIR /app

# SECURITY: Copy only requirements first for layer caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# SECURITY: Multi-stage build - final image is minimal
FROM python:3.11-slim-bookworm@sha256:abc123...

# SECURITY: Create non-root user
RUN groupadd -r appgroup && useradd -r -g appgroup appuser

WORKDIR /app

# SECURITY: Copy only necessary files from builder
COPY --from=builder /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages
COPY --chown=appuser:appgroup . .

# SECURITY: Switch to non-root user
USER appuser

# SECURITY: Don't run as PID 1 - use tini or dumb-init
ENTRYPOINT ["tini", "--"]
CMD ["python", "app.py"]
```

### Secure Docker Run

```bash
# SECURITY: Run with security best practices
docker run \
    --user 1000:1000 \            # Non-root user
    --cap-drop=ALL \              # Drop all capabilities
    --cap-add=NET_BIND_SERVICE \  # Add only what's needed
    --read-only \                 # Read-only filesystem
    --tmpfs /tmp \                # Writable temp only
    --security-opt=no-new-privileges:true \
    --network=app-network \       # Isolated network
    --memory=512m \               # Memory limit
    --cpus=1 \                    # CPU limit
    myapp:v1.2.3
```

### Secure Docker Compose

```yaml
version: '3.8'
services:
  app:
    image: myapp:v1.2.3@sha256:abc123...  # SECURITY: Pinned digest
    user: "1000:1000"                       # SECURITY: Non-root
    read_only: true                         # SECURITY: Read-only FS
    security_opt:
      - no-new-privileges:true              # SECURITY: No privilege escalation
    cap_drop:
      - ALL                                 # SECURITY: Drop all capabilities
    tmpfs:
      - /tmp
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1'
    # SECURITY: Never mount docker.sock
    # volumes:
    #   - /var/run/docker.sock:/var/run/docker.sock  # DANGER!
```

### Secure Kubernetes Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: secure-app
spec:
  # SECURITY: Don't use default service account
  serviceAccountName: app-service-account
  automountServiceAccountToken: false  # SECURITY: Disable if not needed
  
  securityContext:
    runAsNonRoot: true               # SECURITY: Enforce non-root
    runAsUser: 1000
    runAsGroup: 1000
    fsGroup: 1000
    seccompProfile:
      type: RuntimeDefault           # SECURITY: Enable seccomp
  
  containers:
    - name: app
      image: myapp:v1.2.3@sha256:abc123...  # SECURITY: Pinned digest
      
      securityContext:
        allowPrivilegeEscalation: false     # SECURITY: No escalation
        readOnlyRootFilesystem: true        # SECURITY: Read-only
        capabilities:
          drop:
            - ALL                           # SECURITY: Drop all caps
        
      resources:
        limits:
          memory: "512Mi"
          cpu: "1"
        requests:
          memory: "256Mi"
          cpu: "500m"
      
      # SECURITY: Health checks
      livenessProbe:
        httpGet:
          path: /health
          port: 8080
        initialDelaySeconds: 10
      
      readinessProbe:
        httpGet:
          path: /ready
          port: 8080
```

### Secure Network Policy (Kubernetes)

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: app-network-policy
spec:
  podSelector:
    matchLabels:
      app: myapp
  policyTypes:
    - Ingress
    - Egress
  ingress:
    # SECURITY: Only allow traffic from specific sources
    - from:
        - podSelector:
            matchLabels:
              app: frontend
      ports:
        - port: 8080
  egress:
    # SECURITY: Only allow traffic to specific destinations
    - to:
        - podSelector:
            matchLabels:
              app: database
      ports:
        - port: 5432
```

### Secure CI/CD (GitHub Actions)

```yaml
name: Secure Build

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

# SECURITY: Minimal permissions
permissions:
  contents: read
  packages: write

jobs:
  build:
    runs-on: ubuntu-latest
    
    steps:
      # SECURITY: Pin actions to specific versions or SHA
      - uses: actions/checkout@v4
      
      # SECURITY: Dependency scanning
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@0.28.0
        with:
          scan-type: 'fs'
          scan-ref: '.'
          severity: 'CRITICAL,HIGH'
          exit-code: '1'
      
      # SECURITY: SAST scanning
      - name: Run CodeQL analysis
        uses: github/codeql-action/analyze@v3
      
      # SECURITY: Build with pinned action versions
      - name: Build image
        run: docker build -t myapp:${{ github.sha }} .
      
      # SECURITY: Scan built image
      - name: Scan image
        uses: aquasecurity/trivy-action@0.28.0
        with:
          image-ref: myapp:${{ github.sha }}
          exit-code: '1'
          severity: 'CRITICAL,HIGH'
  
  deploy:
    needs: build
    runs-on: ubuntu-latest
    # SECURITY: Require manual approval for production
    environment: production
    
    steps:
      - name: Deploy
        env:
          # SECURITY: Use GitHub Secrets, never hardcode
          API_KEY: ${{ secrets.DEPLOY_API_KEY }}
        run: |
          # Deploy script
```

### Secure Dependency Management

```bash
# SECURITY: Use lockfiles and verify integrity
# Python
pip install --require-hashes -r requirements.txt

# Node.js - use npm ci (not npm install) for reproducible builds
npm ci --ignore-scripts  # SECURITY: Disable postinstall scripts

# Go
go mod verify
```

```json
// package.json - SECURITY: Pin exact versions
{
  "dependencies": {
    "express": "4.18.2",
    "lodash": "4.17.21"
  }
}
```

```txt
# requirements.txt - SECURITY: Pin with hashes
Flask==2.3.3 \
    --hash=sha256:...
```

### Infrastructure as Code Security (Terraform)

```hcl
# SECURITY: S3 bucket with security best practices
resource "aws_s3_bucket" "secure_bucket" {
  bucket = "my-secure-bucket"
}

# SECURITY: Block public access
resource "aws_s3_bucket_public_access_block" "secure_bucket" {
  bucket = aws_s3_bucket.secure_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# SECURITY: Enable encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "secure_bucket" {
  bucket = aws_s3_bucket.secure_bucket.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
  }
}

# SECURITY: Enable versioning
resource "aws_s3_bucket_versioning" "secure_bucket" {
  bucket = aws_s3_bucket.secure_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}
```

## Unsafe Patterns (Flag These)

### Docker Anti-patterns
```dockerfile
# DANGER: Running as root
FROM ubuntu:latest
# No USER directive = runs as root

# DANGER: Using :latest tag
FROM python:latest  # VULNERABLE - unpredictable

# DANGER: Copying secrets
COPY .env /app/.env  # VULNERABLE
ENV API_KEY=secret123  # VULNERABLE
```

```bash
# DANGER: Privileged container
docker run --privileged myapp  # VULNERABLE

# DANGER: Exposing docker socket
docker run -v /var/run/docker.sock:/var/run/docker.sock myapp  # VULNERABLE
```

### Kubernetes Anti-patterns
```yaml
# DANGER: Running as root with privilege escalation
securityContext:
  runAsUser: 0                    # VULNERABLE - root
  allowPrivilegeEscalation: true  # VULNERABLE
  privileged: true                # VULNERABLE
```

### CI/CD Anti-patterns
```yaml
# DANGER: Hardcoded secrets
env:
  API_KEY: "sk-secret123"  # VULNERABLE

# DANGER: Auto-merge without review
on:
  push:
    branches: [main]
# No pull_request review required
```

## Security Checklist

After configuring infrastructure:

- [ ] Containers run as non-root user
- [ ] All capabilities dropped, only necessary ones added
- [ ] Images pinned to specific digests
- [ ] Docker socket not exposed to containers
- [ ] Kubernetes pods use restricted security context
- [ ] Network policies configured
- [ ] Secrets stored in secrets manager, not in code
- [ ] CI/CD requires code review before merge
- [ ] Dependency scanning enabled in pipeline
- [ ] Container image scanning before deployment
- [ ] Production deployments require manual approval

## Source References

- [Docker Security Cheat Sheet](../../security-guidelines/Docker_Security_Cheat_Sheet.md)
- [Kubernetes Security Cheat Sheet](../../security-guidelines/Kubernetes_Security_Cheat_Sheet.md)
- [CI/CD Security Cheat Sheet](../../security-guidelines/CI_CD_Security_Cheat_Sheet.md)
- [Secure Cloud Architecture Cheat Sheet](../../security-guidelines/Secure_Cloud_Architecture_Cheat_Sheet.md)
- [Infrastructure as Code Security Cheat Sheet](../../security-guidelines/Infrastructure_as_Code_Security_Cheat_Sheet.md)
- [NPM Security Cheat Sheet](../../security-guidelines/NPM_Security_Cheat_Sheet.md)
- [Software Supply Chain Security Cheat Sheet](../../security-guidelines/Software_Supply_Chain_Security_Cheat_Sheet.md)
- [Vulnerable Dependency Management Cheat Sheet](../../security-guidelines/Vulnerable_Dependency_Management_Cheat_Sheet.md)

