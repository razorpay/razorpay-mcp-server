# SQS Queue Infrastructure Checks

## Overview

Validates AWS SQS queue patterns to prevent message loss, duplicate processing, and visibility timeout issues.

**Load when:** PR modifies SQS queue consumers, producers, or queue configuration

**Total Checks:** 6

**Severity Distribution:**
- 🚨 Critical: 3
- ⚠️ High: 2
- 📋 Medium: 1

---

## Check 1: Dead Letter Queue (DLQ) Configuration 🚨 CRITICAL

### What to Check

SQS queues must have DLQ configured to capture messages that fail processing repeatedly.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Queue without DLQ
queueURL, err := sqs.CreateQueue(&sqs.CreateQueueInput{
    QueueName: aws.String("payment-notifications"),
    Attributes: map[string]*string{
        "VisibilityTimeout": aws.String("30"),
        "MessageRetentionPeriod": aws.String("86400"),
        // ❌ No RedrivePolicy - failed messages lost after max receives!
    },
})
```

**Problem:**
- Messages failing 5+ times are deleted
- No way to investigate failures
- Data loss

### Good Pattern ✅

```go
// CORRECT: Queue with DLQ configuration
// Step 1: Create DLQ
dlqURL, _ := sqs.CreateQueue(&sqs.CreateQueueInput{
    QueueName: aws.String("payment-notifications-dlq"),
    Attributes: map[string]*string{
        "MessageRetentionPeriod": aws.String("1209600"),  // 14 days
    },
})

// Get DLQ ARN
dlqARN, _ := sqs.GetQueueAttributes(&sqs.GetQueueAttributesInput{
    QueueUrl: dlqURL,
    AttributeNames: []*string{aws.String("QueueArn")},
})

// Step 2: Create main queue with DLQ
redrivePolicy := fmt.Sprintf(`{
    "deadLetterTargetArn": "%s",
    "maxReceiveCount": 3
}`, *dlqARN.Attributes["QueueArn"])

queueURL, _ := sqs.CreateQueue(&sqs.CreateQueueInput{
    QueueName: aws.String("payment-notifications"),
    Attributes: map[string]*string{
        "VisibilityTimeout":      aws.String("30"),
        "MessageRetentionPeriod": aws.String("86400"),
        "RedrivePolicy":          aws.String(redrivePolicy),  // ✅ DLQ configured
    },
})
```

### Detection Strategy

```bash
# Find SQS queue creation
grep -n "CreateQueue" internal/sqs/*.go pkg/queue/*.go

# For each CreateQueue, verify:
# - RedrivePolicy attribute exists
# - maxReceiveCount set (typically 3-5)
# - DLQ created before main queue
```

### Flag Conditions

Flag if:
- `CreateQueue` without `RedrivePolicy` attribute
- `maxReceiveCount` not set or too high (>10)
- No DLQ queue creation code
- Production queue without DLQ

### Severity

🚨 **Critical** - Message loss, no failure investigation

### Reference

Pattern seen in payments services for notification queues

---

## Check 2: Message Deletion After Processing 🚨 CRITICAL

### What to Check

Messages must be explicitly deleted after successful processing, not before.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Delete before processing
func ConsumeMessages(queueURL string) {
    result, _ := sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
        QueueUrl:            aws.String(queueURL),
        MaxNumberOfMessages: aws.Int64(10),
    })

    for _, message := range result.Messages {
        // ❌ Delete BEFORE processing - if processing fails, message is lost!
        sqs.DeleteMessage(&sqs.DeleteMessageInput{
            QueueUrl:      aws.String(queueURL),
            ReceiptHandle: message.ReceiptHandle,
        })

        processMessage(message)  // If this fails, message is gone!
    }
}

// ANTI-PATTERN: Delete without error check
func ConsumeMessages2(queueURL string) {
    for _, message := range messages {
        processMessage(message)

        // ❌ Always deletes, even if processing failed
        sqs.DeleteMessage(&sqs.DeleteMessageInput{
            QueueUrl:      aws.String(queueURL),
            ReceiptHandle: message.ReceiptHandle,
        })
    }
}
```

**Problem:**
- Message deleted before processing
- Processing errors cause message loss
- No retry mechanism

### Good Pattern ✅

```go
// CORRECT: Delete only after successful processing
func ConsumeMessages(queueURL string) {
    result, _ := sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
        QueueUrl:            aws.String(queueURL),
        MaxNumberOfMessages: aws.Int64(10),
        WaitTimeSeconds:     aws.Int64(20),  // Long polling
    })

    for _, message := range result.Messages {
        // Process first
        err := processMessage(message)

        if err != nil {
            // ✅ Don't delete - let message become visible again
            logger.Error(ctx, "message_processing_failed",
                "error", err,
                "messageId", *message.MessageId)
            continue
        }

        // ✅ Delete only after successful processing
        _, err = sqs.DeleteMessage(&sqs.DeleteMessageInput{
            QueueUrl:      aws.String(queueURL),
            ReceiptHandle: message.ReceiptHandle,
        })

        if err != nil {
            logger.Error(ctx, "message_delete_failed", "error", err)
            // Message will be reprocessed (idempotent handler needed!)
        }
    }
}
```

### Severity

🚨 **Critical** - Message loss on processing failures

---

## Check 3: Visibility Timeout Management 🚨 CRITICAL

### What to Check

Visibility timeout must be longer than processing time to prevent duplicate processing.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Short visibility timeout
queueURL, _ := sqs.CreateQueue(&sqs.CreateQueueInput{
    QueueName: aws.String("heavy-processing-queue"),
    Attributes: map[string]*string{
        "VisibilityTimeout": aws.String("30"),  // ❌ 30 seconds
    },
})

// Consumer takes 60 seconds to process
func processMessage(msg *sqs.Message) error {
    time.Sleep(60 * time.Second)  // Heavy processing
    // Message becomes visible again at 30s!
    // Another consumer picks it up → Duplicate processing!
    return nil
}
```

**Problem:**
- Message reappears while still processing
- Multiple consumers process same message
- Duplicate transactions, double charges, etc.

### Good Pattern ✅

```go
// PATTERN 1: Set visibility timeout > max processing time
queueURL, _ := sqs.CreateQueue(&sqs.CreateQueueInput{
    QueueName: aws.String("heavy-processing-queue"),
    Attributes: map[string]*string{
        "VisibilityTimeout": aws.String("300"),  // ✅ 5 minutes (> max 3 min processing)
    },
})

// PATTERN 2: Extend visibility timeout during processing
func processLongRunningMessage(msg *sqs.Message, queueURL string) error {
    // Start processing
    go extendVisibilityPeriodically(msg, queueURL)

    // Long processing
    err := heavyProcessing(msg)

    return err
}

func extendVisibilityPeriodically(msg *sqs.Message, queueURL string) {
    ticker := time.NewTicker(20 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        // ✅ Extend visibility timeout while processing
        _, err := sqs.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
            QueueUrl:          aws.String(queueURL),
            ReceiptHandle:     msg.ReceiptHandle,
            VisibilityTimeout: aws.Int64(60),  // Extend by 60 more seconds
        })

        if err != nil {
            logger.Warn(ctx, "failed_to_extend_visibility", "error", err)
            break
        }
    }
}

// PATTERN 3: Idempotent message handler (best practice)
func processMessage(msg *sqs.Message) error {
    messageId := *msg.MessageId

    // ✅ Check if already processed
    if isProcessed(messageId) {
        logger.Info(ctx, "message_already_processed", "id", messageId)
        return nil
    }

    // Process with distributed lock
    lock := acquireLock(messageId)
    defer lock.Release()

    err := doProcessing(msg)
    if err != nil {
        return err
    }

    // Mark as processed
    markAsProcessed(messageId)
    return nil
}
```

### Detection Strategy

Look for visibility timeout configuration and compare with actual processing time.

### Severity

🚨 **Critical** - Duplicate processing, data corruption

---

## Check 4: Long Polling for Receive Messages ⚠️ HIGH

### What to Check

Use long polling (WaitTimeSeconds > 0) to reduce empty responses and costs.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Short polling (default)
func pollQueue(queueURL string) {
    for {
        result, _ := sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
            QueueUrl:            aws.String(queueURL),
            MaxNumberOfMessages: aws.Int64(10),
            // ❌ No WaitTimeSeconds - short polling!
        })

        // Makes API call immediately even if queue is empty
        // High API costs, high latency
        time.Sleep(1 * time.Second)  // ❌ Manual delay
    }
}
```

**Problem:**
- Many empty responses
- High API request costs
- Unnecessary CPU usage

### Good Pattern ✅

```go
// CORRECT: Long polling
func pollQueue(queueURL string) {
    for {
        result, err := sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
            QueueUrl:            aws.String(queueURL),
            MaxNumberOfMessages: aws.Int64(10),
            WaitTimeSeconds:     aws.Int64(20),  // ✅ Long polling (max 20)
            AttributeNames: []*string{
                aws.String("ApproximateReceiveCount"),
            },
        })

        if err != nil {
            logger.Error(ctx, "receive_message_failed", "error", err)
            time.Sleep(5 * time.Second)  // Backoff on error
            continue
        }

        // Process messages
        for _, message := range result.Messages {
            processMessage(message)
        }

        // No sleep needed - WaitTimeSeconds handles it
    }
}
```

### Severity

⚠️ **High** - High costs, poor performance

---

## Check 5: Batch Processing for Performance ⚠️ HIGH

### What to Check

Use batch operations (SendMessageBatch, DeleteMessageBatch) for efficiency.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Send messages one by one
func sendNotifications(notifications []Notification) {
    for _, notif := range notifications {
        // ❌ 100 notifications = 100 API calls!
        sqs.SendMessage(&sqs.SendMessageInput{
            QueueUrl:    aws.String(queueURL),
            MessageBody: aws.String(notif.ToJSON()),
        })
    }
}

// ANTI-PATTERN: Delete one by one
func deleteProcessedMessages(messages []*sqs.Message) {
    for _, msg := range messages {
        // ❌ 100 messages = 100 API calls!
        sqs.DeleteMessage(&sqs.DeleteMessageInput{
            QueueUrl:      aws.String(queueURL),
            ReceiptHandle: msg.ReceiptHandle,
        })
    }
}
```

**Problem:**
- High API costs
- Slow throughput
- Rate limiting issues

### Good Pattern ✅

```go
// CORRECT: Batch send
func sendNotifications(notifications []Notification) error {
    const batchSize = 10  // SQS max batch size

    for i := 0; i < len(notifications); i += batchSize {
        end := min(i+batchSize, len(notifications))
        batch := notifications[i:end]

        entries := make([]*sqs.SendMessageBatchRequestEntry, len(batch))
        for j, notif := range batch {
            entries[j] = &sqs.SendMessageBatchRequestEntry{
                Id:          aws.String(fmt.Sprintf("msg-%d", j)),
                MessageBody: aws.String(notif.ToJSON()),
            }
        }

        // ✅ Send 10 messages in one API call
        result, err := sqs.SendMessageBatch(&sqs.SendMessageBatchInput{
            QueueUrl: aws.String(queueURL),
            Entries:  entries,
        })

        if err != nil {
            return err
        }

        // Check for partial failures
        if len(result.Failed) > 0 {
            logger.Warn(ctx, "batch_send_partial_failure",
                "failed_count", len(result.Failed))
        }
    }

    return nil
}

// CORRECT: Batch delete
func deleteMessages(messages []*sqs.Message) {
    const batchSize = 10

    for i := 0; i < len(messages); i += batchSize {
        end := min(i+batchSize, len(messages))
        batch := messages[i:end]

        entries := make([]*sqs.DeleteMessageBatchRequestEntry, len(batch))
        for j, msg := range batch {
            entries[j] = &sqs.DeleteMessageBatchRequestEntry{
                Id:            aws.String(fmt.Sprintf("msg-%d", j)),
                ReceiptHandle: msg.ReceiptHandle,
            }
        }

        // ✅ Delete 10 messages in one API call
        sqs.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
            QueueUrl: aws.String(queueURL),
            Entries:  entries,
        })
    }
}
```

### Severity

⚠️ **High** - High costs, poor throughput

---

## Check 6: Error Handling and Retry Backoff 📋 MEDIUM

### What to Check

SQS API errors should have exponential backoff retry.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No error handling
func pollQueue(queueURL string) {
    for {
        result, _ := sqs.ReceiveMessage(input)  // ❌ Error ignored
        for _, msg := range result.Messages {
            processMessage(msg)
        }
    }
}

// ANTI-PATTERN: Constant retry interval
func pollQueueWithRetry(queueURL string) {
    for {
        result, err := sqs.ReceiveMessage(input)
        if err != nil {
            time.Sleep(1 * time.Second)  // ❌ Same delay every time
            continue
        }
        // Process
    }
}
```

### Good Pattern ✅

```go
// CORRECT: Exponential backoff
func pollQueue(queueURL string) {
    backoff := 1 * time.Second
    const maxBackoff = 60 * time.Second

    for {
        result, err := sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
            QueueUrl:        aws.String(queueURL),
            WaitTimeSeconds: aws.Int64(20),
        })

        if err != nil {
            logger.Error(ctx, "sqs_receive_error", "error", err)

            // ✅ Exponential backoff
            time.Sleep(backoff)
            backoff = min(backoff*2, maxBackoff)
            continue
        }

        // ✅ Reset backoff on success
        backoff = 1 * time.Second

        for _, message := range result.Messages {
            processMessage(message)
        }
    }
}
```

### Severity

📋 **Medium** - Poor error resilience

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | DLQ configuration | 🚨 Critical | Message loss |
| 2 | Delete after processing | 🚨 Critical | Message loss |
| 3 | Visibility timeout | 🚨 Critical | Duplicate processing |
| 4 | Long polling | ⚠️ High | High costs |
| 5 | Batch operations | ⚠️ High | Poor throughput |
| 6 | Retry backoff | 📋 Medium | Poor resilience |

---

## How to Apply

**For each file matching** `internal/sqs/*`, `pkg/queue/*`, AWS SQS code:

1. Check DLQ configuration in queue creation
2. Verify message deletion happens after processing
3. Check visibility timeout > max processing time
4. Verify long polling (WaitTimeSeconds = 20)
5. Look for batch operations
6. Check error handling with backoff

**Example output:**

```
📁 File: internal/sqs/consumer.go

🚨 Check #1 Failed: No DLQ configured (Line 23)
   Code: CreateQueue without RedrivePolicy
   Fix: Add DLQ with maxReceiveCount: 3

🚨 Check #2 Failed: Message deleted before processing (Line 67)
   Code: DeleteMessage before processMessage()
   Fix: Move DeleteMessage after successful processing

⚠️  Check #4 Failed: No long polling (Line 45)
   Code: ReceiveMessage without WaitTimeSeconds
   Fix: Add WaitTimeSeconds: 20

✅ Check #3 Passed: Visibility timeout appropriate (300s)
✅ Check #5 Passed: Using SendMessageBatch
✅ Check #6 Passed: Exponential backoff on errors
```
