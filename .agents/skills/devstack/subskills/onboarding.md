# Application Onboarding Subskill

Comprehensive workflow for onboarding new applications to the devstack ecosystem.

## ⚠️ CRITICAL: Helm Chart Location

**ALWAYS create/update helm charts in `helmfile/charts/<application-name>` within the kube-manifests repository ONLY.**

- ✅ CORRECT: `<kube-manifests-repo>/helmfile/charts/<service-name>/`
- ❌ WRONG: `charts/<service-name>/`
- ❌ WRONG: `<any-other-repo>/charts/<service-name>/`
- ❌ WRONG: Any path outside `helmfile/charts/` in kube-manifests repo

**This is the ONLY valid location for helm charts in the devstack ecosystem.**

## Purpose

Guide developers through the complete process of onboarding a new application to devstack, from helm chart creation to deployment and monitoring.

## When to Use

- Onboarding a new microservice to devstack
- Setting up ephemeral environments for a new application
- Creating helm charts for kubernetes deployment
- Configuring databases, caches, and queues for an application
- Setting up secrets and monitoring for a new service

## Prerequisites

- **kube-manifests repository cloned locally** (if not cloned, see below)
- Access to kube-manifests repository
- kubectl configured with devstack cluster access
- Understanding of the application's runtime requirements (DB, cache, queues, etc.)
- Application container image available in `c.rzp.io/razorpay/<service-name>`
- Helmfile directory configured (see [../SKILL.md#configuration](../SKILL.md#configuration))

### Clone kube-manifests Repository

If you don't have the kube-manifests repository cloned locally:

```bash
# Clone the repository
git clone git@github.com:razorpay/kube-manifests.git

# Navigate to helmfile directory
cd kube-manifests/helmfile

# Verify helmfile.yaml exists
ls helmfile.yaml
```

Update the devstack config to point to this location:
```bash
# From your current working directory, update config
echo "{\"helmfile_directory\": \"$(pwd)/kube-manifests/helmfile\", \"auto_detect\": true}" > claude-skills/infrastructure/skills/devstack/config.json
```

## Onboarding Workflow

### Phase 0: Path Detection and Repository Check

**Before starting**, ensure the kube-manifests repository is cloned and helmfile directory path is configured:

#### Step 1: Check if kube-manifests repository exists

```bash
# Check if kube-manifests repo is cloned
if [ ! -d "kube-manifests" ] && [ ! -d "../kube-manifests" ]; then
  echo "⚠️ kube-manifests repository not found"
  echo "Please clone it first:"
  echo "  git clone git@github.com:razorpay/kube-manifests.git"
  exit 1
fi
```

#### Step 2: Verify helmfile directory

```bash
# Read helmfile directory from config.json (see references/path-detection.md)
# The skill will automatically detect the path and report it to you
```

For the examples below, replace `<HELMFILE_DIR>` with your actual helmfile directory path from config.json (typically `<path>/kube-manifests/helmfile`).

### Phase 1: Repository Setup

1. **Navigate to helmfile directory in kube-manifests repo**
   ```bash
   # Use the path from config.json
   cd <HELMFILE_DIR>

   # Verify you're in the right place
   ls helmfile.yaml

   # Create a branch for onboarding
   git checkout -b onboard-<service-name>
   ```

2. **Create helm chart directory structure in helmfile/charts/<service-name>**
   ```bash
   # CRITICAL: Must be inside helmfile/charts/ directory
   cd helmfile/charts
   mkdir <service-name>
   cd <service-name>
   mkdir templates
   touch Chart.yaml values.yaml
   cd templates
   touch NOTES.txt deployment.yaml svc.yaml preview-url.yaml
   ```

   **Final directory structure:**
   ```
   <kube-manifests-repo>/helmfile/charts/<service-name>/
   ├── Chart.yaml
   ├── values.yaml
   └── templates/
       ├── NOTES.txt
       ├── deployment.yaml
       ├── svc.yaml
       └── preview-url.yaml
   ```

### Phase 2: Chart Configuration

#### Chart.yaml

Create basic chart metadata:

```yaml
apiVersion: v2
name: <service-name>
description: <service-name> helmchart
type: application
version: 0.1.0
appVersion: 1.16.0
```

#### values.yaml

Configure service parameters. **Critical fields**:

```yaml
# Application Configuration
app_env: dev
namespace: <service-name>
name: <service-name>
bu: platform  # Business unit

# Image Configuration
image_base: c.rzp.io/razorpay/<service-name>
image_pull_policy: IfNotPresent

# Resource Limits (REQUIRED)
web_requests_cpu: 50m
web_requests_memory: 200Mi
web_limits_memory: 500Mi

# Deployment Configuration
replicas: 1
service_port: 80
container_port: 9400  # Application server port

# Node Placement
node_selector: node.kubernetes.io/worker-generic
dns_policy: ClusterFirst

# Secrets
secret_name: <service-name>

# Base Pod Configuration (for persistent base deployment)
base:
  replicas: 2
  node_selector: node.kubernetes.io/worker-generic-base

# Ephemeral Resources (Configure as needed)
ephemeral_db: true
ephemeral_cache: false
ephemeral_sqs: false
ephemeral_sns: false
```

**Optional Database Configuration** (if ephemeral_db: true):

```yaml
database:
  type: mysql  # or postgres
  name: <service-name>
  namespace: <service-name>
  username: <service-name>
  password: <generated-password>
  requests_cpu: 50m
  requests_memory: 50Mi
  dns_policy: ClusterFirst
  node_selector: node.kubernetes.io/worker-database
  version: ""  # Defaults to latest
  attach_volume: false
  volume_size: ""
```

**Optional Cache Configuration** (if ephemeral_cache: true):

```yaml
cache:
  namespace: <service-name>
  type: redis
  requests_cpu: 50m
  requests_memory: 50Mi
  dns_policy: ClusterFirst
  node_selector: node.kubernetes.io/worker-database-graviton
  version: "6.0"
```

#### NOTES.txt

Template for post-deployment access information:

```txt
*****************HURRRAAAYYYYY******************
Thank you for installing {{ .Chart.Name }}.

This installation of yours can be accessed on
URL :  https://{{ .Values.ingress }}
Header : "rzpctx-dev-serve-user": "{{ .Values.devstack_label }}"
OR

URL : https://{{ .Values.name }}-{{ .Values.devstack_label }}.dev.razorpay.in

For serving through your local code from this installation, please follow the devspace doc
PS: Also remember to run helmfile delete once you are done.
************************************************
```

### Phase 3: Create Kubernetes Resources

#### deployment.yaml

**Critical Requirements**:
- Suffix deployment name with `{{ .Values.devstack_label }}`
- Include mandatory annotations: `janitor/ttl`
- Include mandatory labels: `bu`, `name`, `devstack_label`
- Configure resource limits (CPU + Memory)
- Add readiness and liveness probes
- Mount secrets properly

**Minimum Template**:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    janitor/ttl: "{{ .Values.ttl }}"
  labels:
    bu: {{ .Values.bu }}
    name: {{ .Values.name }}-{{ .Values.devstack_label }}
    devstack_label: {{ .Values.devstack_label }}
    {{ if eq .Values.devstack_label "base" }}
    velero.io/include-in-backup: "true"
    protected: "true"
    {{ end }}
  name: {{ .Values.name }}-{{ .Values.devstack_label }}
  namespace: {{ .Values.namespace }}
spec:
  replicas: {{ .Values.replicas }}
  selector:
    matchLabels:
      name: {{ .Values.name }}-{{ .Values.devstack_label }}
  template:
    metadata:
      annotations:
        prometheus.io/path: /metrics
        prometheus.io/port: "{{ .Values.container_port }}"
        prometheus.io/scrape: "true"
      labels:
        bu: {{ .Values.bu }}
        name: {{ .Values.name }}-{{ .Values.devstack_label }}
        devstack_label: {{ .Values.devstack_label }}
    spec:
      containers:
        - name: web
          image: {{ .Values.image_base }}:{{ .Values.image }}
          imagePullPolicy: {{ .Values.image_pull_policy }}
          ports:
            - containerPort: {{ .Values.container_port }}
          env:
            - name: APP_ENV
              value: {{ .Values.app_env }}
          envFrom:
            - secretRef:
                name: {{ .Values.name }}
                optional: false
          {{ if or .Values.ephemeral_db .Values.ephemeral_sqs .Values.ephemeral_cache }}
            - secretRef:
                name: {{ .Values.name }}-{{ .Values.devstack_label }}
                optional: false
          {{ end }}
          livenessProbe:
            httpGet:
              path: /health
              port: {{ .Values.container_port }}
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health
              port: {{ .Values.container_port }}
            initialDelaySeconds: 10
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          resources:
            requests:
              cpu: {{ .Values.web_requests_cpu }}
              memory: {{ .Values.web_requests_memory }}
            limits:
              memory: {{ .Values.web_limits_memory }}
      dnsPolicy: {{ .Values.dns_policy }}
      nodeSelector:
        {{ if eq .Values.devstack_label "base" }}
          {{ .Values.base.node_selector }}: ""
        {{ else }}
          {{ .Values.node_selector }}: ""
        {{ end }}
```

#### svc.yaml

Create ClusterIP service:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ .Values.name }}-{{ .Values.devstack_label }}
  namespace: {{ .Values.namespace }}
  annotations:
    janitor/ttl: "{{ .Values.ttl }}"
  labels:
    devstack_label: {{ .Values.devstack_label }}
    {{ if eq .Values.devstack_label "base" }}
    velero.io/include-in-backup: "true"
    protected: "true"
    {{ end }}
spec:
  ports:
    - port: {{ .Values.service_port }}
      protocol: TCP
      targetPort: {{ .Values.container_port }}
  selector:
    name: {{ .Values.name }}-{{ .Values.devstack_label }}
  type: ClusterIP
```

#### preview-url.yaml

Create IngressRoute for external access. Supports two access patterns:
1. `service-name.dev.razorpay.in` + header `rzpctx-dev-serve-user: <label>`
2. `service-name-<label>.dev.razorpay.in`

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  labels:
    devstack_label: {{ .Values.devstack_label }}
    {{ if eq .Values.devstack_label "base" }}
    velero.io/include-in-backup: "true"
    protected: "true"
    {{ end }}
  annotations:
    janitor/ttl: "{{ .Values.ttl }}"
  name: injectheader-{{ .Values.devstack_label }}
  namespace: {{ .Values.namespace }}
spec:
  headers:
    customRequestHeaders:
      rzpctx-dev-serve-user: {{ .Values.devstack_label }}
---
kind: IngressRoute
apiVersion: traefik.containo.us/v1alpha1
metadata:
  labels:
    devstack_label: {{ .Values.devstack_label }}
    {{ if eq .Values.devstack_label "base" }}
    velero.io/include-in-backup: "true"
    protected: "true"
    {{ end }}
  annotations:
    kubernetes.io/ingress.class: traefik-concierge
    janitor/ttl: "{{ .Values.ttl }}"
  name: {{ .Values.name }}-{{ .Values.devstack_label }}
  namespace: {{ .Values.namespace }}
spec:
  entryPoints:
    - http
  routes:
    - kind: Rule
      match: Host(`{{ .Values.name }}.dev.razorpay.in`) && Headers(`rzpctx-dev-serve-user`,`{{ .Values.devstack_label }}`)
      services:
        - name: '{{ .Values.name }}-{{ .Values.devstack_label }}'
          port: {{ .Values.service_port }}
    - kind: Rule
      match: Host(`{{ .Values.name }}-{{ .Values.devstack_label }}.dev.razorpay.in`)
      services:
        - name: '{{ .Values.name }}-{{ .Values.devstack_label }}'
          port: {{ .Values.service_port }}
      middlewares:
        - name: injectheader-{{ .Values.devstack_label }}
```

### Phase 4: Configure Ephemeral Resources

#### Ephemeral Database (Optional)

If `ephemeral_db: true`, create these files:

**db-configmap.yaml**:

```yaml
{{- if .Values.ephemeral_db }}
apiVersion: v1
kind: ConfigMap
data:
  app.yaml: |
    name: {{ .Values.database.type }}-{{ .Values.devstack_label }}
    type: {{ .Values.database.type }}
    imageTag: {{ .Values.database.version | quote }}
    namespace: {{ .Values.database.namespace }}
    ttl: {{ .Values.ttl }}
    requestsCpu: {{ .Values.database.requests_cpu }}
    requestsMemory: {{ .Values.database.requests_memory }}
    dnsPolicy: {{ .Values.database.dns_policy }}
    nodeSelector: {{ .Values.database.node_selector }}
    rootPassword: {{ randAlphaNum 12 | lower }}
    attachVolume: {{ .Values.database.attach_volume | default false }}
    volumeSize: {{ .Values.database.volume_size | default "" }}
    databases:
      - dbName: {{ .Values.database.name }}
        username: {{ .Values.database.username }}
        password: {{ .Values.database.password }}
        seeding: false
        snapshotPath: ""
        configKey: db
metadata:
  labels:
    app: dbc-{{ .Values.name }}-{{ .Values.devstack_label }}
  name: dbc-{{ .Values.name }}-{{ .Values.devstack_label }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "2"
  namespace: db-configurator
{{- end }}
```

**db-configurator.yaml**:

```yaml
{{- if .Values.ephemeral_db }}
apiVersion: batch/v1
kind: Job
metadata:
  name: dbc-{{ .Values.name }}-{{ .Values.devstack_label }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "3"
    janitor/ttl: "{{ .Values.ttl }}"
  namespace: db-configurator
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 60
  template:
    spec:
      containers:
        - image: 'c.rzp.io/razorpay/kube-manifests:dbc'
          imagePullPolicy: Always
          name: dbc
          resources:
            limits:
              cpu: 200m
              memory: 500Mi
            requests:
              cpu: 100m
              memory: 150Mi
          volumeMounts:
            - name: config-volume
              mountPath: /src/config
      nodeSelector:
        node.kubernetes.io/worker-configurators: ''
      volumes:
        - name: config-volume
          configMap:
            name: dbc-{{ .Values.name }}-{{ .Values.devstack_label }}
      restartPolicy: Never
{{- end }}
```

**Database Connection Details**:
- Host: `<database-type>-<label>.<namespace>.svc.cluster.local`
- Port: 3306 (MySQL) or 5432 (Postgres)
- Database: As configured in configmap
- Username/Password: As configured in configmap

#### Ephemeral Cache (Optional)

Similar structure to database. Create `cache-configmap.yaml` and `cache-configurator.yaml`.

**Cache Connection**: `redis-<label>.<namespace>.svc.cluster.local:6379`

#### Ephemeral SQS/Queues (Optional)

For applications requiring async queues:

**sqs-configmap.yaml**: Define queue names and secret keys
**sqs-configurator.yaml**: Job to provision queues on localstack

**Queue URL Format**: `http://localstack.localstack.svc.cluster.local:4566/000000000000/<queue-name>-<label>`

**Example values.yaml configuration**:
```yaml
configurator:
  sqs: true

queues:
  - name: devstack-my-service-jobs
    secretKey: JOBS_QUEUE_URL
  - name: devstack-my-service-tasks
    secretKey: TASKS_QUEUE_URL
```

See [helm-chart-templates.md](../references/helm-chart-templates.md#sqs-configmapyaml) for complete templates.

#### Ephemeral SNS/Topics (Optional)

For applications requiring pub/sub messaging with SNS topics:

**sns-configmap.yaml**: Define SNS topics and their SQS subscriptions
**sns-configurator.yaml**: Job to provision topics and subscriptions on localstack

**Topic ARN Format**: `arn:aws:sns:ap-south-1:000000000000:<topic-prefix>-<label>`

**Subscription Endpoint Format**: `http://localstack.localstack.svc.cluster.local:4566/000000000000/<subscription-queue>-<label>`

**Example values.yaml configuration**:
```yaml
configurator:
  sns: true

topics:
  - prefix: devstack-my-service-event-processed
    secret_name: SNS_TOPICS_EVENT_PROCESSED_NAME
    subscriptions:
      - devstack-consumer-service-process-event
      - devstack-analytics-service-track-event
  - prefix: devstack-my-service-notification-sent
    secret_name: SNS_TOPICS_NOTIFICATION_SENT_NAME
    subscriptions:
      - devstack-audit-service-log-notification
```

**Key Features**:
- **Multiple topics**: Define multiple SNS topics with different subscriptions
- **SQS subscriptions**: Each topic can have multiple SQS queue subscriptions
- **Auto-cleanup**: Topics are deleted and recreated on each sync (controlled by `deleteExistingTopic`)
- **Debug mode**: Enabled by default for better troubleshooting
- **Dead Letter Queue**: Supports DLQ endpoints for failed messages

**Hook Execution Order**:
1. `pre-install,pre-upgrade` with weight `2` - Creates SNS ConfigMap
2. `post-install,post-upgrade` with weight `4` - Executes SNS configurator job

See [helm-chart-templates.md](../references/helm-chart-templates.md#sns-configmapyaml) for complete templates.

#### Using Both SQS and SNS (Optional)

You can enable both configurators for applications that need both queues and topics:

```yaml
configurator:
  sqs: true
  sns: true

# SQS queues
queues:
  - name: devstack-my-service-jobs
    secretKey: JOBS_QUEUE_URL

# SNS topics that publish to SQS queues
topics:
  - prefix: devstack-my-service-events
    secret_name: SNS_TOPICS_EVENTS_NAME
    subscriptions:
      - devstack-my-service-jobs  # Can subscribe to your own SQS queues
      - devstack-other-service-consumer
```

**Common Use Cases**:
- **Event-driven architecture**: SNS topics for events, SQS for processing
- **Fan-out pattern**: One SNS topic publishes to multiple SQS consumers
- **Microservices communication**: SNS for pub/sub, SQS for work queues
- **Cross-service integration**: SNS topics consumed by multiple services

### Phase 5: Secrets Management

#### Base Secrets

1. **Add secrets to credstash**:
   - URL: https://credstash-ui.concierge.stage.razorpay.in/dist/
   - Table: `kubestash-dev-serve`
   - Key format: `<namespace>/<secret-name>/<secret-key>`
   - Example: `pg-router/pg-router-live/DB_HOST`

2. **Secrets sync automatically** within 5 minutes via kubestash job

3. **Verify secret creation**:
   ```bash
   kubectl get secret -n <namespace> <secret-name>
   ```

#### Ephemeral Secrets

For label-specific overrides (e.g., ephemeral DB credentials):

**secret-cloner.yaml**: Clones base secret

```yaml
{{- if .Values.ephemeral_db }}
apiVersion: batch/v1
kind: Job
metadata:
  name: sec-{{ .Values.name }}-{{ .Values.devstack_label }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "1"
    janitor/ttl: "{{ .Values.ttl }}"
  namespace: secret-cloner
spec:
  backoffLimit: 0
  template:
    spec:
      containers:
        - env:
            - name: ACTION
              value: clone
            - name: NAMESPACE
              value: '{{ .Values.namespace }}'
            - name: SECRETNAME
              value: '{{ .Values.secret_name }}'
            - name: SECRETSUFFIX
              value: '{{ .Values.devstack_label }}'
          image: 'c.rzp.io/razorpay/kube-manifests:sec'
          name: sec
      restartPolicy: OnFailure
{{- end }}
```

**sec-updater-cm.yaml**: Define keys to override

```yaml
{{- if .Values.ephemeral_db }}
apiVersion: v1
kind: ConfigMap
data:
  app.yaml: |
    updateEntries:
      s1:
        key: DB_HOST
        value: {{ .Values.database.type }}-{{ .Values.devstack_label }}.{{ .Values.database.namespace }}.svc.cluster.local
      s2:
        key: DB_NAME
        value: {{ .Values.database.name }}
      s3:
        key: DB_USERNAME
        value: {{ .Values.database.username }}
      s4:
        key: DB_PASSWORD
        value: {{ .Values.database.password }}
    action: update
    secretName: {{ .Values.secret_name }}-{{ .Values.devstack_label }}
    namespace: {{ .Values.namespace }}
metadata:
  name: sec-updater-{{ .Values.name }}-{{ .Values.devstack_label }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "4"
  namespace: secret-cloner
{{- end }}
```

**sec-updater.yaml**: Job to update secrets

### Phase 6: Add Service to Helmfile

1. **Edit helmfile.yaml in kube-manifests repo**:
   ```bash
   cd <kube-manifests-repo>/helmfile
   vim helmfile.yaml
   ```

2. **Add service entry**:
   ```yaml
   - name: <service-name>-{{ .Values.devstack_label }}
     namespace: <service-name>
     chart: ./charts/<service-name>
     values:
       - image: <commit-hash>
       - devstack_label: {{ .Values.devstack_label }}
       - ttl: {{ .Values.ttl }}
       - namespace: <service-name>
       - secret: {{ .Values.secret }}
   ```

### Phase 7: Validation & Deployment

1. **Validate chart syntax in helmfile/charts/**:
   ```bash
   cd <kube-manifests-repo>/helmfile
   helm lint charts/<service-name>/
   ```

2. **Validate template rendering**:
   ```bash
   helmfile -f helmfile.yaml -l name=<service-name>-<label> template
   ```

3. **Deploy**:
   ```bash
   helmfile -f helmfile.yaml -l name=<service-name>-<label> sync
   ```

4. **Verify deployment**:
   ```bash
   kubectl get pods -n <service-name> -l devstack_label=<label>
   kubectl get svc -n <service-name> -l devstack_label=<label>
   ```

### Phase 8: Setup Monitoring & Logging

#### Logging

Logs automatically pushed to Coralogix if printed to stdout.

**Access logs**: https://razorpay-non-prod.app.coralogix.in/#/query-new/archive-logs

#### Monitoring

Ensure prometheus annotations in deployment:

```yaml
template:
  metadata:
    annotations:
      prometheus.io/path: /metrics
      prometheus.io/port: "<container-port>"
      prometheus.io/scrape: "true"
```

### Phase 9: Deploy Base Pod via Spinnaker (Optional)

For persistent base deployments, use Spinnaker V3 pipeline:

1. **Create pipeline config** in spinnacode repo
2. **Configure variables**:
   - `namespace`: Service namespace
   - `helm_chart_path_prefix`: S3 location
   - `default_overrides`: `devstack_label=base,ttl=forever`
   - `helm_release_name_override`: `<service-name>-base`

3. **Execute deployment** at deploy.razorpay.com

## Validation Checklist

Before deployment, verify:

- [ ] Chart.yaml has correct name and version
- [ ] values.yaml has all required fields
- [ ] Resource limits configured (CPU + Memory)
- [ ] Mandatory annotations present: `janitor/ttl`
- [ ] Mandatory labels present: `bu`, `name`, `devstack_label`
- [ ] Probes configured (liveness + readiness)
- [ ] Secrets mounted correctly
- [ ] Service targetPort matches container port
- [ ] IngressRoute configured for external access
- [ ] Ephemeral resources configured if needed
- [ ] Base secrets added to credstash
- [ ] Service added to helmfile.yaml
- [ ] Template renders without errors
- [ ] Monitoring annotations present

## Common Issues

**Issue**: Template rendering fails
**Fix**: Run `helm lint charts/<service-name>/` to identify syntax errors

**Issue**: Secrets not found
**Fix**: Verify secret exists: `kubectl get secret -n <namespace> <secret-name>`

**Issue**: Pods fail to start (ImagePullBackOff)
**Fix**: Verify image exists in registry and imagePullPolicy is correct

**Issue**: Database connection fails
**Fix**: Check ephemeral DB is running: `kubectl get pods -n <namespace> -l app=<database-type>-<label>`

**Issue**: Service not accessible
**Fix**: Verify IngressRoute and Service are created: `kubectl get ingressroute,svc -n <namespace>`

## Related Documentation

- [Deployment Subskill](deployment.md) - Deploy onboarded applications
- [Debugging Subskill](debugging.md) - Debug deployment issues
- [Validation Subskill](validation.md) - Validate configurations
- [Config Checklist](../references/config-checklist.md) - Complete configuration reference

## Automation Tools

**Go Foundation V2**: Auto-generates manifests for Golang applications
- Documentation: https://idocs.razorpay.com/platform/dev-productivity/go-foundation/v2/#devstack
- Use for Golang services to skip manual chart creation

## Adding Ephemeral Resources to Existing Applications

If an application is already onboarded to devstack but needs ephemeral resources (database, cache, queues) added:

### Step 1: Identify Existing Chart

1. **Locate the chart in helmfile/charts/ directory**:
   ```bash
   # Navigate to kube-manifests repo
   cd <kube-manifests-repo>

   # Verify chart exists in correct location
   ls helmfile/charts/<service-name>/
   ```

2. **Read current configuration**:
   ```bash
   cat helmfile/charts/<service-name>/values.yaml
   cat helmfile/charts/<service-name>/templates/deployment.yaml
   ```

### Step 2: Update values.yaml

Add ephemeral resource flags and configuration:

```yaml
# Add to existing values.yaml

# Enable ephemeral database
ephemeral_db: true

database:
  type: mysql  # or postgres
  name: <service-name>
  namespace: <service-name>
  username: <service-name>
  password: <auto-generated-password>
  requests_cpu: 50m
  requests_memory: 50Mi
  dns_policy: ClusterFirst
  node_selector: node.kubernetes.io/worker-database
  version: ""

# Or enable ephemeral cache
ephemeral_cache: true

cache:
  namespace: <service-name>
  type: redis
  requests_cpu: 50m
  requests_memory: 50Mi
  dns_policy: ClusterFirst
  node_selector: node.kubernetes.io/worker-database-graviton
  version: "6.0"
```

### Step 3: Create New Template Files

Add these files to `helmfile/charts/<service-name>/templates/`:

**For Ephemeral Database**:
- `db-configmap.yaml` (see helm-chart-templates.md)
- `db-configurator.yaml` (see helm-chart-templates.md)

**For Ephemeral Cache**:
- `cache-configmap.yaml` (similar to db-configmap.yaml)
- `cache-configurator.yaml` (similar to db-configurator.yaml)

**For Secret Management**:
- `secret-cloner.yaml` (if not already present)
- `sec-updater-cm.yaml` (update with DB/cache credentials)
- `sec-updater.yaml` (if not already present)

### Step 4: Update Existing deployment.yaml

Modify the secret mounting section to include ephemeral secrets:

**Find this section**:
```yaml
envFrom:
  - secretRef:
      name: {{ .Values.name }}
      optional: false
```

**Replace with**:
```yaml
envFrom:
  - secretRef:
      name: {{ .Values.name }}
      optional: false
{{ if or .Values.ephemeral_db .Values.ephemeral_cache .Values.ephemeral_sqs }}
  - secretRef:
      name: {{ .Values.name }}-{{ .Values.devstack_label }}
      optional: false
{{ end }}
```

### Step 5: Configure Secret Updates

Update `sec-updater-cm.yaml` to include database/cache connection details:

```yaml
updateEntries:
  s1:
    key: DB_HOST
    value: {{ .Values.database.type }}-{{ .Values.devstack_label }}.{{ .Values.database.namespace }}.svc.cluster.local
  s2:
    key: DB_NAME
    value: {{ .Values.database.name }}
  s3:
    key: DB_USERNAME
    value: {{ .Values.database.username }}
  s4:
    key: DB_PASSWORD
    value: {{ .Values.database.password }}
  s5:
    key: DB_PORT
    value: "3306"  # or 5432 for postgres
```

### Step 6: Add Base Secrets

1. **Add database credentials to credstash**:
   - URL: https://credstash-ui.concierge.stage.razorpay.in/dist/
   - Table: `kubestash-dev-serve`
   - Keys to add:
     - `<namespace>/<secret-name>/DB_HOST` (will be overridden for ephemeral)
     - `<namespace>/<secret-name>/DB_NAME` (will be overridden for ephemeral)
     - `<namespace>/<secret-name>/DB_USERNAME` (will be overridden for ephemeral)
     - `<namespace>/<secret-name>/DB_PASSWORD` (will be overridden for ephemeral)
     - `<namespace>/<secret-name>/DB_PORT` (will be overridden for ephemeral)

2. **Wait for sync** (5 minutes) or verify:
   ```bash
   kubectl get secret -n <namespace> <secret-name>
   ```

### Step 7: Test Changes

1. **Validate template rendering**:
   ```bash
   cd <kube-manifests-repo>/helmfile
   helmfile -f helmfile.yaml -l name=<service>-<label> template
   ```

2. **Deploy and verify**:
   ```bash
   helmfile -f helmfile.yaml -l name=<service>-<label> sync
   kubectl get pods -n <namespace> -l devstack_label=<label>
   kubectl get pods -n <namespace> -l app=<database-type>-<label>  # Check DB pod
   ```

3. **Check application logs**:
   ```bash
   kubectl logs -f <app-pod-name> -n <namespace>
   # Should show successful database connection
   ```

### Common Modifications

#### Adding Only Database

Files to create/modify:
- ✅ `values.yaml` - Add `ephemeral_db: true` and `database:` config
- ✅ `templates/db-configmap.yaml` - NEW
- ✅ `templates/db-configurator.yaml` - NEW
- ✅ `templates/deployment.yaml` - Modify `envFrom` section
- ✅ `templates/secret-cloner.yaml` - Create if missing
- ✅ `templates/sec-updater-cm.yaml` - Create/update with DB credentials
- ✅ `templates/sec-updater.yaml` - Create if missing

#### Adding Only Cache

Files to create/modify:
- ✅ `values.yaml` - Add `ephemeral_cache: true` and `cache:` config
- ✅ `templates/cache-configmap.yaml` - NEW
- ✅ `templates/cache-configurator.yaml` - NEW
- ✅ `templates/deployment.yaml` - Modify `envFrom` section
- ✅ `templates/secret-cloner.yaml` - Create if missing
- ✅ `templates/sec-updater-cm.yaml` - Create/update with Redis host
- ✅ `templates/sec-updater.yaml` - Create if missing

#### Adding SQS Queues

Files to create/modify:
- ✅ `values.yaml` - Add `configurator.sqs: true` and `queues:` config
- ✅ `templates/sqs-configmap.yaml` - NEW
- ✅ `templates/sqs-configurator.yaml` - NEW
- ✅ `templates/secret-cloner.yaml` - Create if missing (for queue URLs in secrets)
- ✅ `templates/sec-updater.yaml` - Create if missing

**Note**: SQS configurator automatically updates the secret with queue URLs using the `secretKey` field.

#### Adding SNS Topics

Files to create/modify:
- ✅ `values.yaml` - Add `configurator.sns: true` and `topics:` config
- ✅ `templates/sns-configmap.yaml` - NEW
- ✅ `templates/sns-configurator.yaml` - NEW
- ✅ `templates/secret-cloner.yaml` - Create if missing (for topic ARNs in secrets)
- ✅ `templates/sec-updater.yaml` - Create if missing

**Important Considerations**:
- SNS topics publish to SQS queues, so ensure subscription queues exist
- Hook weight is `4` (runs after SQS configurator which has weight `3`)
- Each topic can have multiple SQS subscriptions (fan-out pattern)
- Topic ARNs are stored in secrets using the `secret_name` field

**Hook Execution Order for SNS**:
1. `pre-install` weight `2` - SNS ConfigMap created
2. `post-install` weight `4` - SNS configurator runs (after SQS)

#### Adding Multiple Resources

When adding multiple resources (database + cache + SQS + SNS):
- Create all resource-specific templates
- Update `sec-updater-cm.yaml` with ALL connection details
- Ensure conditional in deployment.yaml includes all: `{{ if or .Values.ephemeral_db .Values.ephemeral_cache .Values.ephemeral_sqs }}`
- For SNS+SQS: Create SQS queues first (subscription endpoints), then SNS topics
- Consider hook weights to ensure proper order: DB/Cache/SQS (weight 3) → SNS (weight 4)

### Troubleshooting

**Issue**: Ephemeral database pod not starting

**Debug**:
```bash
kubectl get pods -n db-configurator -l app=dbc-<service>-<label>
kubectl logs -n db-configurator dbc-<service>-<label>-xxxxx
kubectl describe job -n db-configurator dbc-<service>-<label>
```

**Issue**: Application can't connect to database

**Fix**: Check secret was updated correctly:
```bash
kubectl get secret -n <namespace> <service>-<label> -o yaml
# Verify DB_HOST, DB_NAME, DB_USERNAME, DB_PASSWORD are present
```

**Issue**: Helm hook failures

**Fix**: Check hook weights are correct:
```
secret-cloner: hook-weight: "1"
db-configmap: hook-weight: "2"
db-configurator: hook-weight: "3"
sqs-configmap: hook-weight: "2"
sqs-configurator: hook-weight: "3"
sns-configmap: hook-weight: "2"
sns-configurator: hook-weight: "4"
sec-updater-cm: hook-weight: "4"
sec-updater: hook-weight: "5"
```

**Issue**: SNS configurator fails with subscription errors

**Debug**:
```bash
kubectl get pods -n sns-configurator -l name=sns-<service>-<label>
kubectl logs -n sns-configurator sns-<service>-<label>-xxxxx
```

**Common Causes**:
- SQS queue (subscription endpoint) doesn't exist
- Queue name mismatch in subscription endpoint
- SNS configurator ran before SQS configurator (check hook weights)

**Fix**:
1. Verify SQS queues exist:
   ```bash
   aws --endpoint-url=http://localstack.localstack.svc.cluster.local:4566 sqs list-queues
   ```
2. Check subscription queue names match in both configs
3. Ensure SNS hook weight (4) is greater than SQS hook weight (3)

**Issue**: SNS topics not receiving messages

**Debug**:
```bash
# Check topic exists
aws --endpoint-url=http://localstack.localstack.svc.cluster.local:4566 sns list-topics

# Check subscriptions
aws --endpoint-url=http://localstack.localstack.svc.cluster.local:4566 sns list-subscriptions-by-topic --topic-arn <arn>

# Check if messages are going to DLQ
aws --endpoint-url=http://localstack.localstack.svc.cluster.local:4566 sqs receive-message --queue-url <dlq-url>
```

**Fix**:
- Verify subscription endpoints are correct in sns-configmap.yaml
- Check protocol is set to `sqs`
- Ensure application is publishing to correct topic ARN

### Quick Checklist

When adding ephemeral resources to existing app:

- [ ] Update `values.yaml` with ephemeral config
- [ ] Create resource configmap (db/cache/sqs/sns)
- [ ] Create resource configurator job
- [ ] Update `deployment.yaml` to mount ephemeral secret (if using db/cache/sqs)
- [ ] Create/update `secret-cloner.yaml`
- [ ] Create/update `sec-updater-cm.yaml` with connection details
- [ ] Create/update `sec-updater.yaml`
- [ ] For SNS: Ensure subscription SQS queues exist before creating topics
- [ ] Add base secrets to credstash (if needed)
- [ ] Validate template rendering
- [ ] Deploy and test connectivity

## Best Practices

1. **Start minimal**: Begin with basic deployment + service, add ephemeral resources as needed
2. **Use ephemeral resources for development**: Keep RDS/Redis for base/prod only
3. **Set appropriate TTLs**: `1h` for quick testing, `8h` for active development, `forever` for base
4. **Resource limits**: Start conservative (50m CPU, 200Mi memory), scale as needed
5. **Probes**: Use simple `/health` endpoints, avoid complex dependency checks in liveness
6. **Secrets**: Never hardcode secrets in charts, always use credstash
7. **Node selectors**: Use `worker-generic` for compute, `worker-database` for databases
8. **Naming**: Keep deployment names consistent with production (suffix with label only)

## Quick Reference Commands

```bash
# Navigate to kube-manifests repo
cd <kube-manifests-repo>/helmfile

# Create chart structure (ALWAYS in helmfile/charts/)
mkdir -p charts/<service>/templates

# Validate chart
helm lint charts/<service>/

# Test template rendering
helmfile -f helmfile.yaml -l name=<service>-<label> template

# Deploy
helmfile -f helmfile.yaml -l name=<service>-<label> sync

# Check status
kubectl get pods,svc,ingressroute -n <namespace> -l devstack_label=<label>

# View logs
kubectl logs -f <pod-name> -n <namespace>

# Delete deployment
helmfile -f helmfile.yaml -l name=<service>-<label> delete
```
