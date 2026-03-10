---
name: razorpay-upi-reserve-pay
description: Integrate Razorpay UPI Reserve Pay (SBMD) APIs into merchant codebases for recurring payments. Use when merchants want to add UPI Reserve Pay, single-block multiple-debit payments, authorization transactions, recurring UPI payments, or mention SBMD, UPI mandates, or reserve-and-debit workflows.
---

# Razorpay UPI Reserve Pay Integration

Guide for integrating UPI Reserve Pay (Single-Block, Multiple-Debit) into merchant applications.

## Overview

UPI Reserve Pay allows businesses to:
- Block a specific amount from customer's account with one authorization
- Debit multiple times from the reserved fund without additional approvals
- Provide frictionless repeat payment experiences

**Example**: A customer authorizes ₹2000 block. Multiple orders (₹400, ₹600) are automatically debited without PIN entry.

## Integration Workflow

**IMPORTANT**: Before implementing, you must understand the merchant's existing codebase.

### Phase 1: Discovery & Understanding

Before writing any code, explore the merchant's application:

1. **Identify existing payment flow**:
   - Where are orders created?
   - How are payments processed?
   - What payment gateways are currently integrated?
   - Is Razorpay already integrated for regular payments?

2. **Understand architecture**:
   - Backend framework (Node.js, Python, Go, Java, etc.)
   - Frontend framework (React, Vue, vanilla JS, etc.)
   - Database schema for orders/payments
   - API structure and routing patterns

3. **Locate key files**:
   - Order creation logic
   - Payment processing handlers
   - Customer management code
   - Razorpay API client initialization
   - Configuration files (API keys, etc.)

4. **Check existing patterns**:
   - How are Razorpay orders currently created?
   - Where is customer data stored?
   - How are payment callbacks handled?
   - Error handling patterns

**Action**: Use exploration tools to find these files before proceeding to implementation.

### Phase 2: Integration Planning

After understanding the codebase:

1. **Map integration points**:
   - Where to add authorization flow
   - Where to store token_id
   - Where to trigger subsequent payments
   - Webhook handling location

2. **Database changes needed**:
   - Store `token_id` with customer records
   - Track authorization status
   - Link tokens to orders/subscriptions

3. **API endpoints to add/modify**:
   - Authorization initiation endpoint
   - Token management endpoints
   - Subsequent payment triggers

### Phase 3: Implementation

Now implement the UPI Reserve Pay flow following the merchant's code patterns.

## Complete Workflow

### Step 1: Create Authorization Order

Create an order for the initial authorization transaction.

**API Endpoint**: `POST /v1/orders`

**Required Parameters**:
```json
{
  "amount": 200,
  "currency": "INR",
  "customer_id": "cust_xxx",
  "method": "upi",
  "token": {
    "max_amount": 200,
    "expire_at": 1767091469,
    "frequency": "as_presented",
    "type": "single_block_multiple_debit"
  },
  "receipt": "Receipt No. 1",
  "notes": {
    "notes_key_1": "Additional info"
  }
}
```

**Key Fields**:
- `amount`: Initial authorization amount (in paise) - typically equals max_amount
- `token.max_amount`: Maximum amount that can be debited in a single charge (in paise)
- `token.expire_at`: Unix timestamp for mandate expiry (default and max: 90 days from now)
- `token.frequency`: Must be `"as_presented"` for SBMD
- `token.type`: Must be `"single_block_multiple_debit"`

**Response**: Returns `order_id` to use in authorization payment.

**Note**: Do NOT add `force_terminal_id` parameter - it's not required for standard integration.

### Step 2: Create Authorization Payment

Initiate UPI mandate authorization via intent flow.

**API Endpoint**: `POST /v1/payments/create/json`

**Required Parameters**:
```json
{
  "amount": 200,
  "contact": "9123456780",
  "currency": "INR",
  "customer_id": "cust_xxx",
  "email": "customer@example.com",
  "method": "upi",
  "order_id": "order_xxx",
  "recurring": true,
  "upi": {
    "flow": "intent"
  }
}
```

**Response**:
```json
{
  "razorpay_payment_id": "pay_xxx",
  "next": [
    {
      "action": "intent",
      "url": "upi://mandate?pa=..."
    },
    {
      "action": "poll",
      "url": "https://api.razorpay.com/v1/payments/pay_xxx"
    }
  ]
}
```

**Critical Webhook Handling**:
- **IGNORE** `payment.failed` webhook during authorization
- **CONSUME** `token.confirmed` webhook to know mandate status
- User approves mandate in TPAP apps (PhonePe/GPay/Paytm)

### Step 2.1: Token Confirmation Webhook

After customer approves mandate, you'll receive a webhook.

**Webhook Event**: `token.confirmed`

**Payload**:
```json
{
  "event": "token.confirmed",
  "payload": {
    "token": {
      "entity": {
        "id": "token_xxx",
        "method": "upi",
        "vpa": {
          "username": "9789809650",
          "handle": "upi",
          "name": "CUSTOMER NAME",
          "status": "valid"
        },
        "recurring_details": {
          "type": "single_block_multiple_debit",
          "status": "confirmed",
          "failure_reason": null,
          "amount_blocked": 200,
          "amount_debited": 0
        },
        "max_amount": 200,
        "expired_at": 1767091469
      }
    }
  }
}
```

**Key Points**:
- `recurring_details.status = "confirmed"` indicates success
- `amount_blocked` shows the mandate limit
- Store the `token.id` for subsequent debits
- Implement timeout logic as users may not act on mandate immediately

### Step 3: Fetch Token Using Customer ID

After authorization, retrieve the token to use for subsequent payments.

**API Endpoint**: `GET /v1/customers/{customer_id}/tokens`

**Response**:
```json
{
  "entity": "collection",
  "count": 1,
  "items": [
    {
      "id": "token_xxx",
      "method": "upi",
      "vpa": {
        "username": "user",
        "handle": "upi",
        "name": "CUSTOMER NAME"
      },
      "recurring_details": {
        "type": "single_block_multiple_debit",
        "status": "confirmed",
        "failure_reason": null,
        "amount_blocked": 200,
        "amount_debited": 100
      },
      "max_amount": 200,
      "expired_at": 1767091469
    }
  ]
}
```

**Alternative Methods**:
- Fetch token by payment ID: `GET /v1/payments/{payment_id}` (includes `token_id` field)
- Fetch specific token: `GET /v1/customers/{customer_id}/tokens/{token_id}`

**Important Token Fields**:
- `id`: Token ID to use for subsequent debits
- `recurring_details.type`: Should be "single_block_multiple_debit"
- `recurring_details.status`: "confirmed", "cancellation_initiated", etc.
- `recurring_details.amount_blocked`: Total blocked amount
- `recurring_details.amount_debited`: Amount already debited
- `recurring_details.failure_reason`: Error message if any
- `max_amount`: Maximum per-transaction limit
- `expired_at`: Mandate expiry timestamp
- `used_at`: Last debit timestamp
- `vpa`: Customer's UPI ID and name

### Step 4: Create Subsequent Debit Order

Create a new order for each debit transaction.

**API Endpoint**: `POST /v1/orders`

**Required Parameters**:
```json
{
  "amount": 100,
  "currency": "INR",
  "receipt": "Receipt No. 2",
  "payment_capture": "0"
}
```

**Important**: This is a standard order, NOT an authorization order. Do not include the `token` object.

**Response**: Returns `order_id` for the debit payment.

### Step 5: Create Subsequent Payment (Debit from Mandate)

Charge the customer using the saved token.

**API Endpoint**: `POST /v1/payments/create/json`

**Required Parameters**:
```json
{
  "amount": 100,
  "currency": "INR",
  "order_id": "order_yyy",
  "customer_id": "cust_xxx",
  "token": "token_xxx",
  "recurring": "1",
  "contact": "9000090000",
  "email": "customer@example.com",
  "description": "Creating recurring payment for Customer Name",
  "notes": {
    "note_key_1": "Additional info"
  }
}
```

**Response**:
```json
{
  "razorpay_payment_id": "pay_xxx",
  "razorpay_order_id": "order_yyy",
  "razorpay_signature": "signature_xxx"
}
```

**Webhooks**:
- **Success**: `payment.authorized` webhook (status = "authorized")
- **Failure**: `payment.failed` webhook with error details

**Important**: The payment status will be "authorized" initially. You need to capture it separately if `payment_capture` was set to "0".

## Token Management

### Cancel Token (Release Funds at Bank)
**API**: `PUT /v1/customers/{customer_id}/tokens/{token_id}/cancel`

**Purpose**: Initiates mandate cancellation to release blocked funds.

**Response**:
```json
{
  "status": "cancellation_initiated"
}
```

**Webhook**: `token.cancellation_initiated` confirms the cancellation request.

**Use this** to properly cancel the mandate and unblock funds at NPCI/bank level.

## Refunds

### Create Refund
**API**: `POST /v1/payments/{payment_id}/refund`

**Required Parameters**:
```json
{
  "amount": 100,
  "speed": "normal",
  "receipt": "refund12324"
}
```

**Webhook Events**:
- `refund.created`: Triggered when refund is created
- `refund.processed`: Triggered when refund is successfully processed
- `refund.failed`: Triggered when refund processing fails

More details: https://razorpay.com/docs/webhooks/refunds

## Webhook Summary

| Event | Description | When to Use |
|-------|-------------|-------------|
| `token.confirmed` | Mandate approved by customer | Store token_id, enable subsequent debits |
| `token.cancellation_initiated` | Cancellation request acknowledged | Update token status |
| `payment.failed` (during auth) | **IGNORE THIS** | Not relevant for mandate creation |
| `payment.authorized` | Debit successful | Capture payment if needed |
| `payment.failed` (during debit) | Debit failed | Handle error, retry logic |
| `refund.created` | Refund initiated | Update order status |
| `refund.processed` | Refund completed | Confirm with customer |
| `refund.failed` | Refund failed | Handle retry |

**Note**: No webhook exists for token deletion (DELETE operation).

## Common Error Scenarios

### Authorization Errors

| Error Code | Description | Action |
|------------|-------------|--------|
| `bank_account_invalid` | Account linked to VPA is invalid | Create new mandate |
| `insufficient_funds` | Insufficient balance | Ask customer to add funds |
| `incorrect_pin` | Wrong UPI PIN entered | Retry with correct PIN |
| `invalid_vpa` | Invalid UPI ID | Retry with valid VPA |
| `payment_declined` | Declined by bank/customer | Check with customer |
| `payment_timed_out` | Payment timeout | Retry |
| `mandate_request_limit_breached` | Too many mandate requests | Wait before retrying |

### Debit Errors

| Error Code | Description | Action |
|------------|-------------|--------|
| `invalid_request` | Token expired or invalid | Create new authorization |
| `transaction_limit_exceeded` | Exceeded amount limits | Reduce amount or contact customer |
| `insufficient_funds` | Not enough balance | Wait or contact customer |

## Implementation Checklist

When implementing UPI Reserve Pay:

```
Phase 1: Discovery
- [ ] Explore and understand existing payment flow
- [ ] Identify where orders are created
- [ ] Locate Razorpay integration code
- [ ] Understand customer data structure
- [ ] Review database schema
- [ ] Map webhook handling

Phase 2: Database & Schema
- [ ] Add token_id field to customer/user table
- [ ] Add authorization status tracking
- [ ] Create migrations if needed
- [ ] Plan token-to-order linking

Phase 3: Authorization Flow
- [ ] Create customer (if not exists)
- [ ] Create authorization order with token parameters
- [ ] Call /v1/payments/create/json with recurring=true and upi.flow=intent
- [ ] Handle intent URL response (redirect to UPI app)
- [ ] IGNORE payment.failed webhook during authorization
- [ ] Implement token.confirmed webhook handler
- [ ] Fetch and store token_id in database

Phase 4: Debit Flow
- [ ] Create standard debit orders (no token object)
- [ ] Call /v1/payments/create/json with token parameter
- [ ] Handle payment.authorized webhook for success
- [ ] Handle payment.failed webhook for failures
- [ ] Implement proper error handling and retry logic

Phase 5: Testing & Integration
- [ ] Test full flow end-to-end
- [ ] Verify existing payment flows still work
- [ ] Test error scenarios
- [ ] Update API documentation
```

## Adapting to Existing Codebase

### Follow Existing Patterns

**Match their coding style**:
- Use their variable naming conventions
- Follow their error handling patterns
- Use their logging/monitoring approach
- Match their API response structure

**Reuse existing components**:
- Use their Razorpay client initialization
- Leverage existing customer lookup functions
- Use their database access patterns
- Integrate with their webhook handlers

### Common Integration Patterns

#### Pattern 1: Existing Razorpay Integration

If Razorpay is already integrated:
```
1. Find existing order creation code
2. Create similar function for authorization orders
3. Add token parameters to existing structure
4. Reuse existing payment handler patterns
```

#### Pattern 2: No Existing Razorpay

If starting fresh:
```
1. Set up Razorpay client with API keys
2. Create dedicated service/module for UPI Reserve Pay
3. Follow merchant's service architecture pattern
4. Integrate webhooks into their webhook handler
```

#### Pattern 3: Microservices Architecture

If using microservices:
```
1. Identify payment service
2. Add UPI Reserve Pay endpoints to payment service
3. Use message queue for async operations if they do
4. Follow their inter-service communication pattern
```

## Key Differences from Regular Payments

1. **Authorization order** requires `token` object with SBMD parameters
2. **Debit orders** are standard orders (no token object)
3. **Authorization payment** uses `upi.flow = "intent"` and returns intent URL
4. **Success indicator** is `token.confirmed` webhook (IGNORE `payment.failed`)
5. **Debit payment** uses `/v1/payments/create/json` with token parameter
6. **Both auth and debit** use same endpoint: `/v1/payments/create/json`


### Verification Steps

Before deploying:
- [ ] Run existing payment tests - all should pass
- [ ] Test regular payment flow - should work unchanged
- [ ] Verify database migrations don't break existing queries
- [ ] Check API endpoints - existing routes unaffected
- [ ] Test webhook handler - handles both old and new events

## Discovery Questions to Ask

When starting integration, clarify:

1. **Business Context**:
   - What is the use case? (subscriptions, quick commerce, repeat purchases)
   - How much should be blocked initially?
   - How often will debits occur?
   - Is this one-time or ongoing mandate?

2. **Technical Context**:
   - What backend language/framework are you using?
   - Is Razorpay already integrated?
   - Where is your order creation logic?
   - How do you store customer data?
   - Do you have webhook handling set up?

3. **Current Flow**:
   - Walk me through your current payment flow
   - Show me where orders are created
   - Where do you handle payment success/failure?

## Testing Tips

1. **Use test mode**: Use Razorpay test credentials (never production during development)

2. **Test authorization flow**:
   - Create customer, order, and initiate payment
   - Use real UPI app (PhonePe/GPay) to approve in test mode
   - Monitor for `token.confirmed` webhook
   - Verify token stored with correct status

3. **Test debit flow**:
   - Create debit order with amount ≤ max_amount
   - Initiate payment with saved token
   - Check `payment.authorized` webhook
   - Verify `amount_debited` increases in token

4. **Test error scenarios**:
   - Debit amount > max_amount (should fail)
   - Debit from expired token
   - Debit from cancelled token
   - Multiple concurrent debits

5. **Test cancellation**:
   - Cancel token via API
   - Verify `token.cancellation_initiated` webhook
   - Check token status updated
   - Attempt debit on cancelled token (should fail)

6. **Verify integration**:
   - Token fields stored correctly in database
   - Existing payment flows unaffected
   - All webhooks handled properly
   - Refund flow works correctly

7. **Important notes**:
   - During testing, mandate rejection option may not be available in TPAP apps
   - Implement timeout logic for pending mandate approvals
   - Test with small amounts in test mode
   - Verify polling mechanism for intent flow

## Quick Reference: Integration Steps

Once you understand the codebase, follow this sequence:

1. **Add database fields** for token_id, token_status, amount_blocked, amount_debited (all nullable)
2. **Create authorization endpoint** following existing route patterns
3. **Implement authorization order creation** with token parameters (type="single_block_multiple_debit")
4. **Call /v1/payments/create/json** with recurring=true and upi.flow=intent
5. **Add webhook handler for token.confirmed** to store token_id (IGNORE payment.failed)
6. **Implement timeout logic** for pending mandate approvals
7. **Create debit endpoint** that calls /v1/payments/create/json with token parameter
8. **Add webhook handlers** for payment.authorized and payment.failed
9. **Implement token management** (cancel to release funds, avoid delete)
10. **Add refund support** if needed
11. **Test isolation** - verify existing flows unaffected
12. **Document** new endpoints for merchant's team

## Additional Resources

- API Reference: https://razorpay.com/docs/api/payments/recurring-payments/upi-reserve-pay/
- UPI Reserve Pay Guide: https://razorpay.com/docs/payments/recurring-payments/upi-reserve-pay/
- Webhooks: https://razorpay.com/docs/webhooks/

## Related MCP Tools

If implementing as MCP tools, consider creating:
- `create_upi_reserve_authorization_order`
- `fetch_recurring_tokens`
- `create_recurring_payment`
- `cancel_recurring_token`
