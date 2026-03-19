# SNS Configurator Guide

Complete reference for setting up SNS (Simple Notification Service) topics with SQS subscriptions in devstack using Localstack.

## Overview

The SNS configurator enables **pub/sub messaging** for microservices in devstack environments by:
- Creating SNS topics on Localstack
- Setting up SQS queue subscriptions
- Managing topic lifecycles (auto-delete and recreate)
- Storing topic ARNs in Kubernetes secrets

## When to Use SNS Configurator

Use SNS when your application needs:

1. **Event Broadcasting (Fan-out)**
   - One event published to multiple consumers
   - Example: Order placed → [Inventory service, Email service, Analytics service]

2. **Microservices Decoupling**
   - Publishers don't know about subscribers
   - Add/remove consumers without changing publisher

3. **Cross-Service Communication**
   - Service A publishes events
   - Services B, C, D subscribe independently

4. **Event-Driven Architecture**
   - Domain events published to topics
   - Multiple bounded contexts react to events

## SNS vs SQS

| Feature | SNS | SQS |
|---------|-----|-----|
| **Pattern** | Pub/Sub | Point-to-Point |
| **Consumers** | Multiple (fan-out) | Single (or competing consumers) |
| **Use Case** | Broadcasting events | Work queues, job processing |
| **Delivery** | Push to subscribers | Pull by consumers |
| **Persistence** | No (ephemeral) | Yes (messages stored) |

**Common Combination**: SNS publishes to multiple SQS queues (best of both worlds)

## Architecture

```
┌──────────────┐
│  Publisher   │
│  Service     │
└──────┬───────┘
       │ Publish
       ▼
┌──────────────────┐
│   SNS Topic      │
│  devstack-foo    │
└──────┬───────────┘
       │ Subscribe (fan-out)
       ├─────────┬─────────┐
       ▼         ▼         ▼
  ┌────────┐ ┌────────┐ ┌────────┐
  │ SQS Q1 │ │ SQS Q2 │ │ SQS Q3 │
  └───┬────┘ └───┬────┘ └───┬────┘
      ▼          ▼          ▼
  ┌────────┐ ┌────────┐ ┌────────┐
  │Service │ │Service │ │Service │
  │   A    │ │   B    │ │   C    │
  └────────┘ └────────┘ └────────┘
```

## Setup Instructions

### Step 1: Update values.yaml

Add SNS configurator settings:

```yaml
# Enable SNS configurator
configurator:
  sns: true
  sqs: false  # Set to true if you also need direct SQS queues

# Define SNS topics and their subscriptions
topics:
  - prefix: devstack-my-service-order-placed
    secret_name: SNS_TOPICS_ORDER_PLACED_ARN
    subscriptions:
      - devstack-inventory-service-orders
      - devstack-email-service-notifications
      - devstack-analytics-events

  - prefix: devstack-my-service-user-registered
    secret_name: SNS_TOPICS_USER_REGISTERED_ARN
    subscriptions:
      - devstack-onboarding-service-users
      - devstack-marketing-service-leads
```

**Field Explanations**:
- `prefix`: Base name of the SNS topic (will be suffixed with `-<devstack_label>`)
- `secret_name`: Key name where topic ARN will be stored in Kubernetes secret
- `subscriptions`: List of SQS queue names (without devstack label suffix)

### Step 2: Create sns-configmap.yaml

Create `templates/sns-configmap.yaml`:

```yaml
{{- if .Values.configurator.sns }}
apiVersion: v1
kind: ConfigMap
data:
  app.yaml: |
    provider: localstack
    debug: true
    deleteExistingTopic: true  # Delete and recreate topics on each sync
    multipleTopics:
      {{- range $pidx, $topic := $.Values.topics }}
      - name: {{ $topic.prefix }}-{{ $.Values.devstack_label }}
        subscriptions:
          {{- range $sidx, $subscription := $topic.subscriptions }}
          - protocol: sqs
            endpoint: http://localstack.localstack.svc.cluster.local:4566/000000000000/{{ $subscription }}-{{ $.Values.devstack_label }}
            dlqEndpoint: arn:aws:sqs:ap-south-1:000000000000:{{ $subscription }}-base
          {{- end }}
      {{- end }}
metadata:
  labels:
    app: sns-{{ .Values.name }}-{{ $.Values.devstack_label }}
  name: sns-{{ .Values.name }}-{{ $.Values.devstack_label }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "2"
  namespace: sns-configurator
{{- end }}
```

**Configuration Options**:
- `provider: localstack` - AWS mock service for development
- `debug: true` - Enable detailed logging
- `deleteExistingTopic: true` - Clean slate on each deployment
- `protocol: sqs` - Subscription protocol (currently only SQS supported)
- `endpoint` - Full SQS queue URL for subscription
- `dlqEndpoint` - Dead Letter Queue ARN for failed messages

### Step 3: Create sns-configurator.yaml

Create `templates/sns-configurator.yaml`:

```yaml
{{- if .Values.configurator.sns }}
apiVersion: batch/v1
kind: Job
metadata:
  name: sns-{{ .Values.name }}-{{ $.Values.devstack_label }}
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-weight": "4"
    janitor/ttl: "{{ $.Values.ttl }}"
  namespace: sns-configurator
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 0
  template:
    metadata:
      labels:
        name: sns-{{ .Values.name }}-{{ $.Values.devstack_label }}
    spec:
      containers:
        - image: 'c.rzp.io/razorpay/kube-manifests:snsc'
          imagePullPolicy: IfNotPresent
          name: snsc
          resources:
            limits:
              cpu: 50m
              memory: 50Mi
            requests:
              cpu: 50m
              memory: 50Mi
          volumeMounts:
          - name: config-volume
            mountPath: /src/config
      imagePullSecrets:
        - name: registry
      nodeSelector:
        node.kubernetes.io/worker-configurators: ''
      volumes:
        - name: config-volume
          configMap:
            name: sns-{{ .Values.name }}-{{ $.Values.devstack_label }}
      restartPolicy: Never
{{- end }}
```

**Important Notes**:
- Hook runs `post-install` (after main deployment)
- Hook weight `4` ensures it runs AFTER SQS configurator (weight `3`)
- `backoffLimit: 0` - Don't retry on failure (manual intervention needed)
- `ttlSecondsAfterFinished: 0` - Clean up job pod immediately
- Image `snsc` contains the SNS configurator script

## Hook Execution Order

When deploying with SNS (and optionally SQS):

```
Weight 1: secret-cloner (clone base secret)
Weight 2: sns-configmap, sqs-configmap (create config)
Weight 3: sqs-configurator (create SQS queues first!)
Weight 4: sns-configurator (create topics and subscriptions)
Weight 5: sec-updater (update secrets with ARNs/URLs)
```

**Why this order?**
- SQS queues MUST exist before SNS can subscribe to them
- Topics reference queue endpoints in subscriptions
- Secrets updated last with all ARNs and URLs

## Topic ARN Storage

Topic ARNs are stored in Kubernetes secrets for application access:

**In Secret** (after deployment):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-service-john
data:
  SNS_TOPICS_ORDER_PLACED_ARN: YXJuOmF3czpzbnM6YXAtc291dGgtMTowMDAwMDAwMDAwMDA6ZGV2c3RhY2stb3JkZXItcGxhY2VkLWpvaG4=
  # Base64: arn:aws:sns:ap-south-1:000000000000:devstack-order-placed-john
```

**In Application** (environment variable):
```javascript
const topicArn = process.env.SNS_TOPICS_ORDER_PLACED_ARN;
// arn:aws:sns:ap-south-1:000000000000:devstack-order-placed-john

await sns.publish({
  TopicArn: topicArn,
  Message: JSON.stringify({ orderId: 123, status: 'placed' })
}).promise();
```

## Complete Example

### Scenario: E-commerce Order Service

Service publishes order events to SNS, consumed by multiple downstream services.

**values.yaml**:
```yaml
namespace: order-service
name: order-service
devstack_label: john

configurator:
  sns: true
  sqs: true  # Also need direct queues for async jobs

# Direct SQS queues for worker jobs
queues:
  - name: devstack-order-processing-jobs
    secretKey: ORDER_PROCESSING_QUEUE_URL

# SNS topics for event broadcasting
topics:
  # Order lifecycle events
  - prefix: devstack-order-placed
    secret_name: SNS_TOPICS_ORDER_PLACED_ARN
    subscriptions:
      - devstack-inventory-service-orders
      - devstack-email-service-order-confirmations
      - devstack-analytics-service-events

  - prefix: devstack-order-shipped
    secret_name: SNS_TOPICS_ORDER_SHIPPED_ARN
    subscriptions:
      - devstack-email-service-shipping-notifications
      - devstack-tracking-service-shipments

  - prefix: devstack-order-cancelled
    secret_name: SNS_TOPICS_ORDER_CANCELLED_ARN
    subscriptions:
      - devstack-inventory-service-orders
      - devstack-refund-service-cancellations
      - devstack-analytics-service-events
```

**Application Code** (Node.js):
```javascript
const AWS = require('aws-sdk');
const sns = new AWS.SNS({
  endpoint: 'http://localstack.localstack.svc.cluster.local:4566',
  region: 'ap-south-1'
});

// Publish order placed event
async function publishOrderPlaced(order) {
  const event = {
    eventType: 'ORDER_PLACED',
    timestamp: new Date().toISOString(),
    orderId: order.id,
    customerId: order.customerId,
    totalAmount: order.totalAmount,
    items: order.items
  };

  await sns.publish({
    TopicArn: process.env.SNS_TOPICS_ORDER_PLACED_ARN,
    Message: JSON.stringify(event),
    MessageAttributes: {
      eventType: { DataType: 'String', StringValue: 'ORDER_PLACED' },
      customerId: { DataType: 'String', StringValue: order.customerId }
    }
  }).promise();

  console.log(`Published order placed event for order ${order.id}`);
}
```

**Subscriber Services** (consume from SQS):
```javascript
// inventory-service subscribes to order events
const sqs = new AWS.SQS({
  endpoint: 'http://localstack.localstack.svc.cluster.local:4566',
  region: 'ap-south-1'
});

async function pollOrderEvents() {
  const queueUrl = 'http://localstack.localstack.svc.cluster.local:4566/000000000000/devstack-inventory-service-orders-john';

  while (true) {
    const messages = await sqs.receiveMessage({
      QueueUrl: queueUrl,
      MaxNumberOfMessages: 10,
      WaitTimeSeconds: 20
    }).promise();

    for (const message of messages.Messages || []) {
      const snsMessage = JSON.parse(message.Body);
      const event = JSON.parse(snsMessage.Message);

      console.log('Received order event:', event);

      // Process based on event type
      if (event.eventType === 'ORDER_PLACED') {
        await reserveInventory(event.items);
      } else if (event.eventType === 'ORDER_CANCELLED') {
        await releaseInventory(event.items);
      }

      // Delete message after processing
      await sqs.deleteMessage({
        QueueUrl: queueUrl,
        ReceiptHandle: message.ReceiptHandle
      }).promise();
    }
  }
}
```

## Advanced Patterns

### Pattern 1: Fan-out with Filtering

Use message attributes for selective consumption:

```yaml
topics:
  - prefix: devstack-order-events
    secret_name: SNS_TOPICS_ORDER_EVENTS_ARN
    subscriptions:
      - devstack-high-value-orders     # Filter: amount > 10000
      - devstack-all-orders            # No filter
      - devstack-premium-customers     # Filter: customer_tier = premium
```

Subscribers can filter messages based on MessageAttributes without processing.

### Pattern 2: Multi-Region (Simulated)

Multiple localstack instances for testing multi-region:

```yaml
topics:
  - prefix: devstack-global-events
    secret_name: SNS_TOPICS_GLOBAL_EVENTS_ARN
    subscriptions:
      - devstack-us-east-processor
      - devstack-eu-west-processor
      - devstack-ap-south-processor
```

### Pattern 3: Event Sourcing

Use SNS as event log with multiple consumers:

```yaml
topics:
  - prefix: devstack-domain-events
    secret_name: SNS_TOPICS_DOMAIN_EVENTS_ARN
    subscriptions:
      - devstack-event-store           # Persist all events
      - devstack-read-model-updater    # Update queries
      - devstack-notification-service  # Send notifications
      - devstack-analytics             # Real-time analytics
```

## Troubleshooting

### Issue: SNS configurator job fails

**Symptoms**:
```
Error: ResourceNotFoundException: Queue does not exist
```

**Debug**:
```bash
# Check SNS configurator logs
kubectl logs -n sns-configurator sns-order-service-john-xxxxx

# List existing SQS queues
kubectl exec -n localstack localstack-0 -- \
  aws --endpoint-url=http://localhost:4566 sqs list-queues

# Check hook execution order
kubectl get jobs -n sns-configurator
kubectl get jobs -n sqs-configurator
```

**Root Cause**: SQS subscription queues don't exist when SNS configurator runs

**Fix**:
1. Ensure SQS queues are created first (either via SQS configurator or manually)
2. Verify hook weights: SQS (weight 3) < SNS (weight 4)
3. Check queue names match exactly in both configs

### Issue: Messages published to SNS but not received

**Debug**:
```bash
# Connect to localstack
kubectl exec -it -n localstack localstack-0 -- bash

# List topics
aws --endpoint-url=http://localhost:4566 sns list-topics

# List subscriptions for a topic
aws --endpoint-url=http://localhost:4566 sns list-subscriptions-by-topic \
  --topic-arn arn:aws:sns:ap-south-1:000000000000:devstack-order-placed-john

# Check if messages are in subscription queue
aws --endpoint-url=http://localhost:4566 sqs receive-message \
  --queue-url http://localhost:4566/000000000000/devstack-inventory-service-orders-john
```

**Common Issues**:
- Subscription not confirmed (auto-confirm in localstack, but check logs)
- Wrong endpoint URL in subscription
- Messages sent to wrong topic ARN
- Subscription deleted but not recreated

### Issue: Topic ARN not in secret

**Debug**:
```bash
# Check secret contents
kubectl get secret -n order-service order-service-john -o yaml

# Check sec-updater logs
kubectl logs -n secret-cloner sec-updater-order-service-john-xxxxx
```

**Fix**:
- Verify `secret_name` field in values.yaml topics configuration
- Check sec-updater-cm.yaml includes SNS topic ARN entries
- Ensure sec-updater runs after SNS configurator (weight 5 > weight 4)

### Issue: deleteExistingTopic causing data loss

**Symptom**: Topics recreated on every deployment, losing existing messages

**Explanation**: By design! SNS topics are ephemeral in devstack

**Fix**: If you need persistent topics, set `deleteExistingTopic: false` in sns-configmap.yaml

**Trade-off**: Old subscriptions may persist if you change configuration

## Best Practices

### 1. Topic Naming

```yaml
# Good: Descriptive, includes service context
- prefix: devstack-order-service-order-placed
- prefix: devstack-inventory-low-stock-alert

# Bad: Too generic
- prefix: devstack-events
- prefix: devstack-notifications
```

### 2. Subscription Queue Naming

```yaml
# Good: Service name + purpose
subscriptions:
  - devstack-email-service-order-notifications
  - devstack-analytics-service-order-events

# Bad: Generic names
subscriptions:
  - devstack-queue1
  - devstack-consumer
```

### 3. Event Schema

Use consistent event structure:

```json
{
  "eventId": "uuid",
  "eventType": "ORDER_PLACED",
  "timestamp": "2026-01-22T10:30:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "orderId": 123,
    "customerId": 456
  }
}
```

### 4. Error Handling

```javascript
// Implement retry with exponential backoff
async function publishWithRetry(topicArn, message, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      await sns.publish({ TopicArn: topicArn, Message: message }).promise();
      return;
    } catch (err) {
      if (i === maxRetries - 1) throw err;
      await sleep(Math.pow(2, i) * 1000);
    }
  }
}
```

### 5. Local Development

```javascript
// Use environment variable to switch between local and AWS
const snsConfig = {
  region: process.env.AWS_REGION || 'ap-south-1'
};

if (process.env.APP_ENV === 'devstack') {
  snsConfig.endpoint = 'http://localstack.localstack.svc.cluster.local:4566';
}

const sns = new AWS.SNS(snsConfig);
```

### 6. Monitoring

Add logging for published events:

```javascript
await sns.publish({
  TopicArn: topicArn,
  Message: JSON.stringify(event)
}).promise();

logger.info('Published event to SNS', {
  topicArn,
  eventType: event.eventType,
  eventId: event.eventId
});
```

## Reference

### Topic ARN Format
```
arn:aws:sns:ap-south-1:000000000000:<topic-prefix>-<devstack-label>
```

Example: `arn:aws:sns:ap-south-1:000000000000:devstack-order-placed-john`

### Subscription Endpoint Format
```
http://localstack.localstack.svc.cluster.local:4566/000000000000/<queue-name>-<devstack-label>
```

Example: `http://localstack.localstack.svc.cluster.local:4566/000000000000/devstack-inventory-orders-john`

### Helm Hook Weights

| Resource | Hook Phase | Weight |
|----------|-----------|--------|
| secret-cloner | pre-install | 1 |
| sns-configmap | pre-install | 2 |
| sqs-configurator | pre-install | 3 |
| sns-configurator | **post-install** | 4 |
| sec-updater-cm | pre-install | 4 |
| sec-updater | pre-install | 5 |

### Configuration File Reference

**app.yaml** (in ConfigMap):
```yaml
provider: localstack           # AWS provider (always localstack for devstack)
debug: true                    # Enable debug logging
deleteExistingTopic: true      # Delete and recreate topics on sync
multipleTopics:                # List of topics to create
  - name: topic-name           # Topic name (with label suffix)
    subscriptions:             # List of SQS subscriptions
      - protocol: sqs          # Protocol (only sqs supported)
        endpoint: queue-url    # Full SQS queue URL
        dlqEndpoint: dlq-arn   # Dead letter queue ARN
```

## Migration from Direct SQS

If migrating from direct SQS communication to SNS:

**Before** (Point-to-Point):
```
Service A → SQS Queue 1 → Service B
         → SQS Queue 2 → Service C
```

**After** (Pub/Sub):
```
Service A → SNS Topic → SQS Queue 1 → Service B
                      → SQS Queue 2 → Service C
```

**Steps**:
1. Add SNS configurator to Service A
2. Keep existing SQS queues
3. Update Service A to publish to SNS instead of direct SQS
4. Add SNS subscriptions to existing queues
5. No changes needed in Services B/C (still consume from SQS)

**Benefits**:
- Service A doesn't know about consumers
- Easy to add new consumers without touching Service A
- Message filtering at subscription level

---

**Version**: 1.0.0
**Last Updated**: 2026-01-22
**Related**: [helm-chart-templates.md](helm-chart-templates.md), [onboarding.md](../subskills/onboarding.md)
