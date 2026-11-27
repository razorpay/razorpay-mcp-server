---
name: razorpay-upi-reserve-pay
description: Integrate Razorpay UPI Reserve Pay (SBMD) APIs into merchant codebases for recurring payments. Use when merchants want to add UPI Reserve Pay, single-block multiple-debit payments, authorization transactions, recurring UPI payments, or mention SBMD, UPI mandates, or reserve-and-debit workflows.
---

# Razorpay UPI Reserve Pay Integration

Guide for integrating UPI Reserve Pay (Single-Block, Multiple-Debit) into merchant applications using the Razorpay SDK.

**Note**: This guide uses the official Razorpay SDK for all integrations. All examples show SDK method calls rather than direct API endpoints for better reliability, type safety, and maintainability.

**⚠️ IMPORTANT - Currency Units**: All amounts in Razorpay API are specified in **paise** (smallest currency unit), not rupees. **100 paise = ₹1**. Always multiply rupee amounts by 100 before passing to the API.

## Overview

UPI Reserve Pay allows businesses to:
- Block a specific amount from customer's account with one authorization
- Debit multiple times from the reserved fund without additional approvals
- Provide frictionless repeat payment experiences

**Example**: A customer authorizes ₹20 block (amount = 2000 paise). Multiple orders of ₹4 (400 paise) and ₹6 (600 paise) are automatically debited without PIN entry.

## 💰 Currency Handling: Paise vs Rupees

**CRITICAL**: All amounts in Razorpay API must be in **paise**, not rupees.

### Conversion Reference

| Rupees | Paise (API Value) |
|--------|-------------------|
| ₹1     | 100               |
| ₹10    | 1,000             |
| ₹100   | 10,000            |
| ₹500   | 50,000            |
| ₹1,000 | 1,00,000          |

### Code Examples

**Convert rupees to paise before API calls:**

```javascript
// JavaScript/Node.js
const amountInRupees = 20;
const amountInPaise = amountInRupees * 100;  // 2000

const order = await client.orders.create({
    amount: amountInPaise,  // Pass 2000, not 20
    currency: "INR"
});
```

```python
# Python
amount_in_rupees = 20
amount_in_paise = amount_in_rupees * 100  # 2000

order = client.order.create({
    "amount": amount_in_paise,  # Pass 2000, not 20
    "currency": "INR"
})
```

```go
// Go
amountInRupees := 20
amountInPaise := amountInRupees * 100  // 2000

order, err := client.Order.Create(map[string]interface{}{
    "amount": amountInPaise,  // Pass 2000, not 20
    "currency": "INR",
}, nil)
```

**Convert paise to rupees for display:**

```javascript
// JavaScript
const amountInPaise = 2000;
const amountInRupees = amountInPaise / 100;  // 20
console.log(`Amount: ₹${amountInRupees}`);  // "Amount: ₹20"
```

```python
# Python
amount_in_paise = 2000
amount_in_rupees = amount_in_paise / 100  # 20
print(f"Amount: ₹{amount_in_rupees}")  # "Amount: ₹20"
```

```go
// Go
amountInPaise := 2000
amountInRupees := float64(amountInPaise) / 100  // 20.0
fmt.Printf("Amount: ₹%.2f\n", amountInRupees)  // "Amount: ₹20.00"
```

### Common Mistakes

❌ **Wrong**: Passing rupee values directly
```javascript
const order = await client.orders.create({
    amount: 20,  // This will create order for ₹0.20, not ₹20!
    currency: "INR"
});
```

✅ **Correct**: Converting to paise first
```javascript
const amountInRupees = 20;
const order = await client.orders.create({
    amount: amountInRupees * 100,  // 2000 paise = ₹20
    currency: "INR"
});
```

## ⚠️ CRITICAL: Correct SDK Method Names by Language

**The method names vary significantly across SDKs.** Use these EXACT method names:

### Node.js (Razorpay SDK v2.9.6+)

**Authorization Payment (Step 2):**
```javascript
const payment = await razorpay.payments.createPaymentJson({
  amount, currency, order_id, customer_id, email, contact,
  method: "upi", recurring: "1", upi: { flow: "intent" }
});
```

**Debit Payment (Step 5):**
```javascript
const payment = await razorpay.payments.createRecurringPayment({
  amount, currency, order_id, customer_id, token, recurring: "1",
  email, contact, description
});
```

❌ **Common Mistakes:**
- `razorpay.payments.create()` - Does NOT exist
- `razorpay.payments.createRecurring()` - Wrong method name
- `Payment.CreateRecurring()` - Go SDK syntax, not Node.js

### Python SDK

**Authorization Payment (Step 2):**
```python
payment = client.payment.create_json({
    "amount": amount, "currency": "INR", "order_id": order_id,
    "customer_id": customer_id, "email": email, "contact": contact,
    "method": "upi", "recurring": "1", "upi": {"flow": "intent"}
})
```

**Debit Payment (Step 5):**
```python
payment = client.payment.create_recurring({
    "amount": amount, "currency": "INR", "order_id": order_id,
    "customer_id": customer_id, "token": token_id, "recurring": "1",
    "email": email, "contact": contact
})
```

### Go SDK

**Authorization Payment (Step 2):**
```go
payment, err := client.Payment.CreateRecurring(razorpay.Payment{
    Amount: amount, Currency: "INR", OrderId: orderId,
    CustomerId: customerId, Email: email, Contact: contact,
    Method: "upi", Recurring: "1", UPI: map[string]interface{}{"flow": "intent"},
}, nil)
```

**Debit Payment (Step 5):**
```go
payment, err := client.Payment.CreateRecurring(razorpay.Payment{
    Amount: amount, Currency: "INR", OrderId: orderId,
    CustomerId: customerId, Token: tokenId, Recurring: "1",
    Email: email, Contact: contact,
}, nil)
```

### PHP SDK

**Authorization Payment (Step 2):**
```php
$payment = $client->payment->createJson([
    'amount' => $amount, 'currency' => 'INR', 'order_id' => $orderId,
    'customer_id' => $customerId, 'email' => $email, 'contact' => $contact,
    'method' => 'upi', 'recurring' => '1', 'upi' => ['flow' => 'intent']
]);
```

**Debit Payment (Step 5):**
```php
$payment = $client->payment->createRecurring([
    'amount' => $amount, 'currency' => 'INR', 'order_id' => $orderId,
    'customer_id' => $customerId, 'token' => $tokenId, 'recurring' => '1',
    'email' => $email, 'contact' => $contact
]);
```

### Java SDK

**Authorization Payment (Step 2):**
```java
JSONObject paymentRequest = new JSONObject();
paymentRequest.put("amount", amount);
paymentRequest.put("currency", "INR");
paymentRequest.put("order_id", orderId);
paymentRequest.put("customer_id", customerId);
paymentRequest.put("email", email);
paymentRequest.put("contact", contact);
paymentRequest.put("method", "upi");
paymentRequest.put("recurring", "1");
paymentRequest.put("upi", new JSONObject().put("flow", "intent"));

Payment payment = client.payments.createJson(paymentRequest);
```

**Debit Payment (Step 5):**
```java
JSONObject paymentRequest = new JSONObject();
paymentRequest.put("amount", amount);
paymentRequest.put("currency", "INR");
paymentRequest.put("order_id", orderId);
paymentRequest.put("customer_id", customerId);
paymentRequest.put("token", tokenId);
paymentRequest.put("recurring", "1");
paymentRequest.put("email", email);
paymentRequest.put("contact", contact);

Payment payment = client.payments.createRecurring(paymentRequest);
```

## 🔑 Key Takeaways

1. **Method names differ across languages** - Always check SDK documentation
2. **Node.js uses camelCase** - `createPaymentJson()`, `createRecurringPayment()`
3. **Python uses snake_case** - `create_json()`, `create_recurring()`
4. **Go/PHP/Java** - Check respective SDK docs for exact method names
5. **The generic `CreateRecurring()` in this guide** is a placeholder - use language-specific names above

## SDK Setup

This guide uses the official Razorpay SDK for all API interactions. Before implementing UPI Reserve Pay, ensure the Razorpay SDK is initialized:

**Go SDK**:
```go
import (
    razorpay "github.com/razorpay/razorpay-go"
)

client := razorpay.NewClient(keyID, keySecret)
```

**Node.js SDK**:
```javascript
const Razorpay = require('razorpay');

const client = new Razorpay({
    key_id: 'YOUR_KEY_ID',
    key_secret: 'YOUR_KEY_SECRET'
});
```

**Python SDK**:
```python
import razorpay

client = razorpay.Client(auth=("YOUR_KEY_ID", "YOUR_KEY_SECRET"))
```

**PHP SDK**:
```php
use Razorpay\Api\Api;

$client = new Api($keyId, $keySecret);
```

**Java SDK**:
```java
import com.razorpay.RazorpayClient;

RazorpayClient client = new RazorpayClient("YOUR_KEY_ID", "YOUR_KEY_SECRET");
```

All code examples in this guide use Go SDK syntax, but the concepts apply to all SDKs. Refer to [Razorpay SDK documentation](https://razorpay.com/docs/api/) for language-specific implementations.

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
   - Razorpay SDK client initialization
   - Configuration files (API keys, etc.)

4. **Check existing patterns**:
   - How are Razorpay orders currently created using the SDK?
   - Where is the Razorpay SDK client initialized?
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

Create an order for the initial authorization transaction using the Razorpay SDK.

**SDK Method**: `client.Order.Create()`

**Required Parameters**:
```go
data := map[string]interface{}{
    "amount": 200,              // Amount in paise (200 paise = ₹2.00)
    "currency": "INR",
    "customer_id": "cust_xxx",
    "method": "upi",
    "token": map[string]interface{}{
        "max_amount": 200,      // Max amount per debit in paise (200 paise = ₹2.00)
        "expire_at": 1767091469,
        "frequency": "as_presented",
        "type": "single_block_multiple_debit",
    },
    "receipt": "Receipt No. 1",
    "notes": map[string]interface{}{
        "notes_key_1": "Additional info",
    },
}

order, err := client.Order.Create(data, nil)
```

**Key Fields**:
- `amount`: Initial authorization amount **in paise** (e.g., 200 paise = ₹2.00) - typically equals max_amount
- `token.max_amount`: Maximum amount that can be debited in a single charge **in paise** (e.g., 200 paise = ₹2.00)
- `token.expire_at`: Unix timestamp for mandate expiry (default and max: 90 days from now)
- `token.frequency`: Must be `"as_presented"` for SBMD
- `token.type`: Must be `"single_block_multiple_debit"`

**Response**: Returns order object with `order_id` to use in authorization payment.

**Note**: Do NOT add `force_terminal_id` parameter - it's not required for standard integration.

### Step 2: Create Authorization Payment

Initiate UPI mandate authorization via intent flow using the Razorpay SDK.

**⚠️ IMPORTANT:** The SDK method name varies by language. See "Correct SDK Method Names by Language" section above for your language-specific method.

**Generic SDK Method (See above for your language)**: Authorization payment creation method

**Required Parameters**:
```go
data := map[string]interface{}{
    "amount": 200,                      // Amount in paise (200 paise = ₹2.00)
    "contact": "9123456780",
    "currency": "INR",
    "customer_id": "cust_xxx",
    "email": "customer@example.com",
    "method": "upi",
    "order_id": "order_xxx",
    "recurring": true,
    "upi": map[string]interface{}{
        "flow": "intent",
    },
}

payment, err := client.Payment.CreateRecurring(data, nil)
```

**Response**:
```go
{
    "razorpay_payment_id": "pay_xxx",
    "next": []interface{}{
        map[string]interface{}{
            "action": "intent",
            "url": "upi://mandate?pa=...",
        },
        map[string]interface{}{
            "action": "poll",
            "url": "https://api.razorpay.com/v1/payments/pay_xxx",
        },
    },
}
```

**Critical Webhook Handling**:
- **IGNORE** `payment.failed` webhook during authorization
- **CONSUME** `token.confirmed` webhook to know mandate status
- User approves mandate in TPAP apps (PhonePe/GPay/Paytm)

### Step 2.1: Handle Intent URL Based on Platform

After receiving the payment response with intent URL, handle it differently based on the platform:

#### For Web Browser / Website

Convert the intent URL to a QR code that customers can scan with their UPI app.

**Implementation Options**:

**Option 1: Using QR Code Library (Recommended)**

**Node.js**:
```javascript
const QRCode = require('qrcode');

// Get intent URL from payment response
const intentUrl = payment.next.find(action => action.action === 'intent')?.url;

// Generate QR code as Data URL
const qrCodeDataUrl = await QRCode.toDataURL(intentUrl, {
    width: 300,
    margin: 2,
    color: {
        dark: '#000000',
        light: '#FFFFFF'
    }
});

// Send to frontend
res.json({
    razorpay_payment_id: payment.razorpay_payment_id,
    qr_code: qrCodeDataUrl,
    intent_url: intentUrl
});
```

**Python**:
```python
import qrcode
import io
import base64

# Get intent URL from payment response
intent_url = next((action['url'] for action in payment['next'] if action['action'] == 'intent'), None)

# Generate QR code
qr = qrcode.QRCode(version=1, box_size=10, border=2)
qr.add_data(intent_url)
qr.make(fit=True)

# Create image
img = qr.make_image(fill_color="black", back_color="white")

# Convert to base64
buffer = io.BytesIO()
img.save(buffer, format='PNG')
qr_code_base64 = base64.b64encode(buffer.getvalue()).decode()

# Return to frontend
return {
    "razorpay_payment_id": payment["razorpay_payment_id"],
    "qr_code": f"data:image/png;base64,{qr_code_base64}",
    "intent_url": intent_url
}
```

**Go**:
```go
import (
    "encoding/base64"
    "github.com/skip2/go-qrcode"
)

// Get intent URL from payment response
var intentUrl string
for _, action := range payment["next"].([]interface{}) {
    actionMap := action.(map[string]interface{})
    if actionMap["action"] == "intent" {
        intentUrl = actionMap["url"].(string)
        break
    }
}

// Generate QR code
qrCode, err := qrcode.Encode(intentUrl, qrcode.Medium, 300)
if err != nil {
    return err
}

// Convert to base64
qrCodeBase64 := base64.StdEncoding.EncodeToString(qrCode)

// Return to frontend
response := map[string]interface{}{
    "razorpay_payment_id": payment["razorpay_payment_id"],
    "qr_code": "data:image/png;base64," + qrCodeBase64,
    "intent_url": intentUrl,
}
```

**Frontend Display**:
```html
<div class="qr-container">
    <h3>Scan QR Code to Authorize UPI Mandate</h3>
    <img src="{{ qr_code }}" alt="UPI QR Code" />
    <p>Scan this QR code with PhonePe, Google Pay, or Paytm to approve the mandate</p>
</div>
```

#### For Mobile App

Open the TPAP app directly using the intent URL.

**Android (Java/Kotlin)**:

```kotlin
// Get intent URL from payment response
val intentUrl = payment.next
    .find { it.action == "intent" }
    ?.url ?: return

// Open TPAP app
try {
    val intent = Intent(Intent.ACTION_VIEW, Uri.parse(intentUrl))
    intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
    startActivity(intent)
} catch (e: ActivityNotFoundException) {
    // Handle case where no UPI app is installed
    Toast.makeText(
        context,
        "No UPI app found. Please install PhonePe, Google Pay, or Paytm",
        Toast.LENGTH_LONG
    ).show()
}
```

**Android (React Native)**:
```javascript
import { Linking, Alert } from 'react-native';

// Get intent URL from payment response
const intentUrl = payment.next.find(action => action.action === 'intent')?.url;

// Open TPAP app
const openTPAPApp = async (url) => {
    try {
        const supported = await Linking.canOpenURL(url);
        if (supported) {
            await Linking.openURL(url);
        } else {
            Alert.alert(
                'No UPI App Found',
                'Please install PhonePe, Google Pay, or Paytm to continue'
            );
        }
    } catch (error) {
        console.error('Error opening UPI app:', error);
        Alert.alert('Error', 'Failed to open UPI app');
    }
};

await openTPAPApp(intentUrl);
```

**iOS (Swift)**:
```swift
// Get intent URL from payment response
guard let intentUrl = payment.next
    .first(where: { $0.action == "intent" })?
    .url else { return }

// Open TPAP app
guard let url = URL(string: intentUrl) else { return }

if UIApplication.shared.canOpenURL(url) {
    UIApplication.shared.open(url, options: [:]) { success in
        if !success {
            // Handle failure
            print("Failed to open UPI app")
        }
    }
} else {
    // Show alert - no UPI app installed
    let alert = UIAlertController(
        title: "No UPI App Found",
        message: "Please install PhonePe, Google Pay, or Paytm",
        preferredStyle: .alert
    )
    alert.addAction(UIAlertAction(title: "OK", style: .default))
    present(alert, animated: true)
}
```

**Flutter**:
```dart
import 'package:url_launcher/url_launcher.dart';

// Get intent URL from payment response
final intentUrl = payment['next']
    .firstWhere((action) => action['action'] == 'intent')['url'];

// Open TPAP app
Future<void> openTPAPApp(String url) async {
  final uri = Uri.parse(url);
  
  if (await canLaunchUrl(uri)) {
    await launchUrl(uri, mode: LaunchMode.externalApplication);
  } else {
    // Show error dialog
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('No UPI App Found'),
        content: Text('Please install PhonePe, Google Pay, or Paytm'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text('OK'),
          ),
        ],
      ),
    );
  }
}

await openTPAPApp(intentUrl);
```

**Important Notes**:
- **Web**: Display QR code and poll for payment status using the poll URL
- **Mobile**: Deep link directly opens the installed TPAP app
- **Polling**: Implement polling mechanism to check authorization status
- **Timeout**: Set appropriate timeout (2-5 minutes) for mandate approval
- **Fallback**: Handle cases where no UPI app is installed on mobile

### Step 2.2: Token Confirmation Webhook

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
- `amount_blocked` shows the mandate limit **in paise** (e.g., 200 = ₹2.00)
- `amount_debited` shows amount already used **in paise**
- Store the `token.id` for subsequent debits
- Implement timeout logic as users may not act on mandate immediately

### Step 3: Fetch Token Using Customer ID

After authorization, retrieve the token to use for subsequent payments using the Razorpay SDK.

**SDK Method**: `client.Customer.FetchAllTokens()`

**Code Example**:
```go
tokens, err := client.Customer.FetchAllTokens(customerID, nil)
```

**Response**:
```go
{
    "entity": "collection",
    "count": 1,
    "items": []interface{}{
        map[string]interface{}{
            "id": "token_xxx",
            "method": "upi",
            "vpa": map[string]interface{}{
                "username": "user",
                "handle": "upi",
                "name": "CUSTOMER NAME",
            },
            "recurring_details": map[string]interface{}{
                "type": "single_block_multiple_debit",
                "status": "confirmed",
                "failure_reason": nil,
                "amount_blocked": 200,
                "amount_debited": 100,
            },
            "max_amount": 200,
            "expired_at": 1767091469,
        },
    },
}
```

**Alternative SDK Methods**:
- Fetch token by payment ID: `client.Payment.Fetch(paymentID, nil)` (includes `token_id` field)
- Fetch specific token: `client.Customer.FetchToken(customerID, tokenID, nil)`

**Important Token Fields**:
- `id`: Token ID to use for subsequent debits
- `recurring_details.type`: Should be "single_block_multiple_debit"
- `recurring_details.status`: "confirmed", "cancellation_initiated", etc.
- `recurring_details.amount_blocked`: Total blocked amount **in paise**
- `recurring_details.amount_debited`: Amount already debited **in paise**
- `recurring_details.failure_reason`: Error message if any
- `max_amount`: Maximum per-transaction limit **in paise**
- `expired_at`: Mandate expiry timestamp
- `used_at`: Last debit timestamp
- `vpa`: Customer's UPI ID and name

### Step 4: Create Subsequent Debit Order

Create a new order for each debit transaction using the Razorpay SDK.

**SDK Method**: `client.Order.Create()`

**Required Parameters**:
```go
data := map[string]interface{}{
    "amount": 100,              // Amount in paise (100 paise = ₹1.00)
    "currency": "INR",
    "receipt": "Receipt No. 2",
    "payment_capture": 0,
}

order, err := client.Order.Create(data, nil)
```

**Important**: 
- This is a standard order, NOT an authorization order. Do not include the `token` object.
- The `amount` must be **in paise** and ≤ `max_amount` specified during authorization

**Response**: Returns order object with `order_id` for the debit payment.

### Step 5: Create Subsequent Payment (Debit from Mandate)

Charge the customer using the saved token with the Razorpay SDK.

**⚠️ IMPORTANT:** The SDK method name varies by language. See "Correct SDK Method Names by Language" section above for your language-specific method.

**Generic SDK Method (See above for your language)**: Recurring payment creation method

**Required Parameters**:
```go
data := map[string]interface{}{
    "amount": 100,                      // Amount in paise (100 paise = ₹1.00)
    "currency": "INR",
    "order_id": "order_yyy",
    "customer_id": "cust_xxx",
    "token": "token_xxx",
    "recurring": "1",
    "contact": "9000090000",
    "email": "customer@example.com",
    "description": "Creating recurring payment for Customer Name",
    "notes": map[string]interface{}{
        "note_key_1": "Additional info",
    },
}

payment, err := client.Payment.CreateRecurring(data, nil)
```

**Response**:
```go
{
    "razorpay_payment_id": "pay_xxx",
    "razorpay_order_id": "order_yyy",
    "razorpay_signature": "signature_xxx",
}
```

**Webhooks**:
- **Success**: `payment.authorized` webhook (status = "authorized")
- **Failure**: `payment.failed` webhook with error details

**Important**: The payment status will be "authorized" initially. You need to capture it separately if `payment_capture` was set to "0".

## Token Management

### Cancel Token (Release Funds at Bank)

**SDK Method**: `client.Customer.CancelToken()`

**Purpose**: Initiates mandate cancellation to release blocked funds.

**Code Example**:
```go
token, err := client.Customer.CancelToken(customerID, tokenID, nil)
```

**Response**:
```go
{
    "status": "cancellation_initiated",
}
```

**Webhook**: `token.cancellation_initiated` confirms the cancellation request.

**Use this** to properly cancel the mandate and unblock funds at NPCI/bank level.

## Refunds

### Create Refund

**SDK Method**: `client.Payment.Refund()`

**Required Parameters**:
```go
data := map[string]interface{}{
    "amount": 100,              // Refund amount in paise (100 paise = ₹1.00)
    "speed": "normal",
    "receipt": "refund12324",
}

refund, err := client.Payment.Refund(paymentID, data, nil)
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

## SDK Error Handling

The Razorpay SDK returns errors that should be handled appropriately:

**Go SDK Error Handling**:
```go
order, err := client.Order.Create(data, nil)
if err != nil {
    // Handle error - check error type
    // SDK errors contain status code and error details
    log.Printf("Order creation failed: %v", err)
    return err
}
```

**Common SDK Error Patterns**:
- **Authentication errors**: Invalid API keys - check client initialization
- **Validation errors**: Missing or invalid parameters - verify data structure
- **API errors**: Rate limits, server errors - implement retry logic
- **Network errors**: Timeouts, connection issues - handle gracefully

**Best Practices**:
- Always check for errors after SDK method calls
- Log errors with context for debugging
- Return user-friendly error messages
- Implement retry logic for transient failures
- Handle webhook failures gracefully

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
- [ ] Create customer using SDK (if not exists): client.Customer.Create()
- [ ] Create authorization order with token parameters: client.Order.Create()
- [ ] Call client.Payment.CreateRecurring() with recurring=true and upi.flow=intent
- [ ] Handle intent URL response:
  - [ ] For Web: Generate QR code from intent URL (using qrcode library)
  - [ ] For Mobile: Implement deep link to open TPAP app (PhonePe/GPay/Paytm)
- [ ] Implement polling mechanism for payment status
- [ ] IGNORE payment.failed webhook during authorization
- [ ] Implement token.confirmed webhook handler
- [ ] Fetch and store token_id in database using client.Customer.FetchAllTokens()

Phase 4: Debit Flow
- [ ] Create standard debit orders (no token object): client.Order.Create()
- [ ] Call client.Payment.CreateRecurring() with token parameter
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
- Use their Razorpay SDK client initialization
- Leverage existing SDK method patterns
- Leverage existing customer lookup functions
- Use their database access patterns
- Integrate with their webhook handlers

### Common Integration Patterns

#### Pattern 1: Existing Razorpay SDK Integration

If Razorpay SDK is already integrated:
```
1. Find existing order creation code using client.Order.Create()
2. Create similar function for authorization orders with token parameters
3. Use client.Payment.CreateRecurring() for both auth and debit flows
4. Reuse existing payment handler patterns and webhook processing
```

#### Pattern 2: No Existing Razorpay SDK

If starting fresh:
```
1. Initialize Razorpay SDK client with API keys:
   client := razorpay.NewClient(keyID, keySecret)
2. Create dedicated service/module for UPI Reserve Pay
3. Follow merchant's service architecture pattern
4. Integrate webhooks into their webhook handler
```

#### Pattern 3: Microservices Architecture

If using microservices:
```
1. Identify payment service
2. Initialize Razorpay SDK client in payment service
3. Add UPI Reserve Pay endpoints using SDK methods
4. Use message queue for async operations if they do
5. Follow their inter-service communication pattern
```

## Key Differences from Regular Payments

1. **Authorization order** requires `token` object with SBMD parameters
2. **Debit orders** are standard orders (no token object)
3. **Authorization payment** uses `upi.flow = "intent"` and returns intent URL
4. **Success indicator** is `token.confirmed` webhook (IGNORE `payment.failed`)
5. **Debit payment** uses `client.Payment.CreateRecurring()` with token parameter
6. **Both auth and debit** use same SDK method: `client.Payment.CreateRecurring()`
7. **All amounts must be in paise** - multiply rupee values by 100


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
   - Is Razorpay SDK already integrated?
   - Where is the Razorpay SDK client initialized?
   - Where is your order creation logic?
   - How do you store customer data?
   - Do you have webhook handling set up?

3. **Current Flow**:
   - Walk me through your current payment flow
   - Show me where orders are created
   - Where do you handle payment success/failure?

## Testing Tips

1. **Use test mode**: Initialize Razorpay SDK with test credentials (never production during development)
   ```go
   client := razorpay.NewClient(testKeyID, testKeySecret)
   ```

2. **Test authorization flow**:
   - Create customer using `client.Customer.Create()`
   - Create order using `client.Order.Create()` with token parameters
   - Initiate payment using `client.Payment.CreateRecurring()`
   - Test platform-specific intent handling:
     - **Web**: Verify QR code generated correctly and can be scanned
     - **Mobile**: Verify deep link opens TPAP app correctly
   - Use real UPI app (PhonePe/GPay) to approve in test mode
   - Test polling mechanism works correctly
   - Monitor for `token.confirmed` webhook
   - Verify token stored with correct status using `client.Customer.FetchAllTokens()`

3. **Test debit flow**:
   - Create debit order using `client.Order.Create()` with amount ≤ max_amount (in paise)
   - Initiate payment using `client.Payment.CreateRecurring()` with saved token
   - Check `payment.authorized` webhook
   - Verify `amount_debited` increases in token using `client.Customer.FetchToken()`

4. **Test error scenarios**:
   - Debit amount > max_amount using SDK (should fail gracefully) - remember amounts are in paise
   - Debit from expired token
   - Debit from cancelled token
   - Multiple concurrent debits
   - Handle SDK error responses properly

5. **Test cancellation**:
   - Cancel token using `client.Customer.CancelToken()`
   - Verify `token.cancellation_initiated` webhook
   - Check token status updated
   - Attempt debit on cancelled token (should fail)

6. **Verify integration**:
   - Token fields stored correctly in database
   - Existing payment flows unaffected
   - All webhooks handled properly
   - Refund flow works correctly

7. **Important notes**:
   - **All amounts must be in paise**: ₹1 = 100 paise, ₹10 = 1000 paise
   - During testing, mandate rejection option may not be available in TPAP apps
   - Implement timeout logic for pending mandate approvals
   - Test with small amounts in test mode (e.g., 100 paise = ₹1)
   - Verify polling mechanism for intent flow

## Quick Reference: Integration Steps

Once you understand the codebase, follow this sequence:

1. **Initialize Razorpay SDK client** if not already done: `client := razorpay.NewClient(keyID, keySecret)`
2. **Add database fields** for token_id, token_status, amount_blocked, amount_debited (all nullable)
3. **Create authorization endpoint** following existing route patterns
4. **Implement authorization order creation** using `client.Order.Create()` with token parameters (type="single_block_multiple_debit")
5. **Call client.Payment.CreateRecurring()** with recurring=true and upi.flow=intent
6. **Handle intent URL based on platform**:
   - For Web: Generate QR code from intent URL using QR library
   - For Mobile: Implement deep linking to open TPAP app directly
7. **Implement polling mechanism** to check payment status
8. **Add webhook handler for token.confirmed** to store token_id (IGNORE payment.failed)
9. **Implement timeout logic** for pending mandate approvals
10. **Create debit endpoint** that calls `client.Payment.CreateRecurring()` with token parameter
11. **Add webhook handlers** for payment.authorized and payment.failed
12. **Implement token management** using `client.Customer.CancelToken()` to release funds
13. **Add refund support** using `client.Payment.Refund()` if needed
14. **Test isolation** - verify existing flows unaffected
15. **Document** new endpoints for merchant's team

## Additional Resources

### Documentation
- UPI Reserve Pay Guide: https://razorpay.com/docs/payments/recurring-payments/upi-reserve-pay/
- API Reference: https://razorpay.com/docs/api/payments/recurring-payments/upi-reserve-pay/
- Webhooks: https://razorpay.com/docs/webhooks/

### SDK Documentation
- Go SDK: https://github.com/razorpay/razorpay-go
- Node.js SDK: https://github.com/razorpay/razorpay-node
- Python SDK: https://github.com/razorpay/razorpay-python
- PHP SDK: https://github.com/razorpay/razorpay-php
- Java SDK: https://github.com/razorpay/razorpay-java

## Related MCP Tools

If implementing as MCP tools use below tools:
- `create_upi_reserve_authorization_order` - Wraps `client.Order.Create()` with token parameters
- `create_upi_reserve_authorization_payment` - Wraps `client.Payment.CreateRecurring()` for mandate auth
- `fetch_recurring_tokens` - Wraps `client.Customer.FetchAllTokens()` or `client.Customer.FetchToken()`
- `create_recurring_payment` - Wraps `client.Payment.CreateRecurring()` with token parameter
- `cancel_recurring_token` - Wraps `client.Customer.CancelToken()`

Each tool should:
- Initialize or reuse existing Razorpay SDK client
- Handle SDK method calls and error responses
- Return structured data suitable for MCP protocol