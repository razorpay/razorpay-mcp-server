//nolint:lll // File contains embedded documentation and error data with long lines
package razorpay

import (
	"context"
	"sort"
	"strings"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// docSection represents a single documentation entry used by
// SearchDocumentation to answer "how do I implement X" queries.
type docSection struct {
	ID            string
	Topic         string
	Title         string
	Keywords      []string
	Summary       string
	CodeExample   string
	DocURL        string
	GuardrailRefs []string
}

// errorEntry represents a single known Razorpay error used by
// ExplainError to answer "what does this error mean" queries.
type errorEntry struct {
	Code            string
	SubDescription  string
	Title           string
	Explanation     string
	CommonCauses    []string
	ResolutionSteps []string
	DocURL          string
	IsRetriable     bool
	GuardrailRef    string
	GuardrailTitle  string
}

var docSections = []docSection{
	{
		ID:            "doc-001",
		Topic:         "payment-verification",
		Title:         "Server-side payment verification via webhook",
		Keywords:      []string{"webhook", "verify", "verification", "signature", "payment.captured", "server-side", "fulfill", "validateWebhookSignature"},
		Summary:       "Payments must be verified server-side via webhook signature validation before fulfilling orders. Never trust the client-side handler callback as confirmation of payment.",
		CodeExample:   "// Express.js webhook handler\n// IMPORTANT: register this BEFORE express.json() middleware\napp.post('/webhooks/razorpay',\n  express.raw({ type: 'application/json' }),\n  (req, res) => {\n    const signature = req.headers['x-razorpay-signature'];\n\n    const isValid = Razorpay.validateWebhookSignature(\n      req.body.toString(),   // raw Buffer, not parsed JSON\n      signature,\n      process.env.RAZORPAY_WEBHOOK_SECRET\n    );\n\n    if (!isValid) {\n      return res.status(400).json({ error: 'Invalid signature' });\n    }\n\n    const event = JSON.parse(req.body);\n\n    if (event.event === 'payment.captured') {\n      const payment = event.payload.payment.entity;\n      fulfillOrder(payment.order_id);  // safe to fulfill here\n    }\n\n    res.json({ status: 'ok' });\n  }\n);",
		DocURL:        "https://razorpay.com/docs/webhooks/validate-test/",
		GuardrailRefs: []string{"RZP-001", "RZP-004"},
	},
	{
		ID:            "doc-002",
		Topic:         "order-creation",
		Title:         "Creating an order before initiating payment",
		Keywords:      []string{"order", "create order", "orders api", "receipt", "order_id", "paise", "before payment"},
		Summary:       "A Razorpay Order must be created server-side before opening Checkout. The order_id is passed to Checkout to link the payment. Amount must be in paise (INR × 100).",
		CodeExample:   "// Server-side order creation (Node.js)\nconst Razorpay = require('razorpay');\nconst razorpay = new Razorpay({\n  key_id: process.env.RAZORPAY_KEY_ID,\n  key_secret: process.env.RAZORPAY_KEY_SECRET,  // server-only!\n});\n\napp.post('/api/create-order', async (req, res) => {\n  const { productId } = req.body;\n  // Get amount from server-side catalog — never trust client amount\n  const amount = PRODUCTS[productId].price * 100;  // rupees → paise\n\n  const order = await razorpay.orders.create({\n    amount,                        // in paise\n    currency: 'INR',\n    receipt: `rcpt_${Date.now()}`, // unique receipt ID\n    notes: { productId }\n  });\n\n  res.json({\n    orderId: order.id,\n    amount: order.amount,\n    keyId: process.env.RAZORPAY_KEY_ID  // key_id is public, key_secret is NOT\n  });\n});\n\n// Frontend: pass order_id to Checkout\nconst order = await fetch('/api/create-order', { ... }).then(r => r.json());\nconst rzp = new Razorpay({\n  key: order.keyId,\n  order_id: order.orderId,  // mandatory\n  amount: order.amount,\n});",
		DocURL:        "https://razorpay.com/docs/api/orders/#create-an-order",
		GuardrailRefs: []string{"RZP-002", "RZP-013"},
	},
	{
		ID:            "doc-003",
		Topic:         "checkout-sdk",
		Title:         "Razorpay Checkout.js — correct integration",
		Keywords:      []string{"checkout", "checkout.js", "open checkout", "payment form", "razorpay options", "handler", "payment.failed", "cdn"},
		Summary:       "Load Checkout.js only from the official CDN. Pass order_id, key_id (not secret), and handle both success (via webhook) and failure (payment.failed handler).",
		CodeExample:   "<!-- Load ONLY from official CDN -->\n<script src=\"https://checkout.razorpay.com/v1/checkout.js\"></script>\n\n<script>\nasync function openCheckout(productId) {\n  // Step 1: Create order server-side\n  const order = await fetch('/api/create-order', {\n    method: 'POST',\n    headers: { 'Content-Type': 'application/json' },\n    body: JSON.stringify({ productId })\n  }).then(r => r.json());\n\n  // Step 2: Open Checkout\n  const rzp = new Razorpay({\n    key: order.keyId,           // key_id only — never key_secret!\n    order_id: order.orderId,    // from server\n    amount: order.amount,       // in paise, from server\n    currency: 'INR',\n    name: 'Your Company',\n    handler: function(response) {\n      // Step 3: Do NOT fulfill here — show pending state\n      // Real confirmation comes via payment.captured webhook\n      showPendingState('Verifying payment...');\n    },\n    prefill: { name, email, contact: '+91XXXXXXXXXX' }\n  });\n\n  // Step 4: Always handle failures\n  rzp.on('payment.failed', function(response) {\n    showError(response.error.description);\n    enableRetry();\n  });\n\n  rzp.open();\n}\n</script>",
		DocURL:        "https://razorpay.com/docs/payments/payment-gateway/web-integration/standard/",
		GuardrailRefs: []string{"RZP-009", "RZP-012", "RZP-013", "RZP-014"},
	},
	{
		ID:            "doc-004",
		Topic:         "api-keys",
		Title:         "API key security — key_id vs key_secret",
		Keywords:      []string{"api key", "key_id", "key_secret", "secret", "environment variable", "test key", "live key", "rzp_test", "rzp_live"},
		Summary:       "key_id (rzp_test_... or rzp_live_...) is public and safe to send to the browser. key_secret must NEVER appear in frontend code. Store key_secret only in server-side environment variables.",
		CodeExample:   "// ✅ CORRECT: key_secret only on server\n// server.js\nconst razorpay = new Razorpay({\n  key_id: process.env.RAZORPAY_KEY_ID,\n  key_secret: process.env.RAZORPAY_KEY_SECRET,  // server-only\n});\n\n// Send only key_id to frontend (it's public)\nres.json({ orderId: order.id, keyId: process.env.RAZORPAY_KEY_ID });\n\n// ❌ WRONG: key_secret in frontend\nconst KEY_SECRET = 'your_secret_here';  // NEVER do this\n\n// .env file (never commit to git)\nRAZORPAY_KEY_ID=rzp_test_xxxxxxxxxxxx\nRAZORPAY_KEY_SECRET=xxxxxxxxxxxxxxxxxxxx\nRAZORPAY_WEBHOOK_SECRET=xxxxxxxxxxxxxxxxxxxx",
		DocURL:        "https://razorpay.com/docs/api/authentication/",
		GuardrailRefs: []string{"RZP-009", "RZP-010", "RZP-011"},
	},
	{
		ID:            "doc-005",
		Topic:         "upi-autopay",
		Title:         "UPI AutoPay — recurring payments setup",
		Keywords:      []string{"upi", "autopay", "upi autopay", "recurring", "mandate", "subscription", "afa", "15000", "15k"},
		Summary:       "UPI AutoPay mandates > ₹15,000 require AFA (Additional Factor Authentication) per RBI rules. Listen for subscription.authenticated webhook, check afa_required field, and redirect to AFA URL if needed.",
		CodeExample:   "// 1. Create subscription\nconst sub = await razorpay.subscriptions.create({\n  plan_id: 'plan_xxx',\n  total_count: 12,\n  quantity: 1,\n});\n\n// 2. Open Checkout with subscription_id\nconst rzp = new Razorpay({\n  key: process.env.RAZORPAY_KEY_ID,\n  subscription_id: sub.id,  // not order_id\n  name: 'Your Company',\n  handler: function(response) {\n    showPendingState('Setting up mandate...');\n  },\n});\nrzp.open();\n\n// 3. Webhook: handle AFA for amounts > ₹15,000\nif (event.event === 'subscription.authenticated') {\n  const sub = event.payload.subscription.entity;\n  if (sub.afa_required === true) {\n    // redirect customer to complete AFA\n    redirectCustomerToAFAUrl(sub.afa_url);\n  }\n}\n\n// 4. Webhook: subscription charged\nif (event.event === 'subscription.charged') {\n  const payment = event.payload.payment.entity;\n  fulfillSubscriptionPeriod(payment.subscription_id);\n}\n\n// 5. CRITICAL: handle halted subscriptions\nif (event.event === 'subscription.halted') {\n  notifyCustomerToUpdatePaymentMethod(sub.customer_id);\n}",
		DocURL:        "https://razorpay.com/docs/payments/recurring-payments/upi/autopay/",
		GuardrailRefs: []string{"RZP-005", "RZP-015", "RZP-016"},
	},
	{
		ID:            "doc-006",
		Topic:         "nach",
		Title:         "NACH mandate — bank account recurring payments",
		Keywords:      []string{"nach", "enach", "bank debit", "mandate", "bank account", "pre-debit", "notification", "24 hours", "bank verification"},
		Summary:       "NACH mandates require a 24-48h bank verification period after creation, and a mandatory 24h pre-debit notification before each debit. Cannot debit immediately after mandate creation.",
		CodeExample:   "// NACH mandate flow\n\n// 1. Create subscription with method: nach\nconst sub = await razorpay.subscriptions.create({\n  plan_id: 'plan_xxx',\n  total_count: 60,\n  customer_notify: 1,\n  addons: [],\n  notify_info: {\n    notify_phone: '+91XXXXXXXXXX',\n    notify_email: 'customer@email.com',\n  }\n});\n\n// 2. Listen for mandate authenticated (24-48h after creation)\nif (event.event === 'subscription.authenticated') {\n  // Bank has verified the mandate — safe to proceed\n  updateMandateStatus(sub.id, 'active');\n}\n\n// 3. Send pre-debit notification (MANDATORY — 24h before debit)\n// This is your responsibility to implement\nawait sendPreDebitNotification({\n  customer_email: customer.email,\n  debit_date: nextDebitDate,\n  amount: planAmount,\n  mandate_id: sub.id,\n});\n\n// 4. Debit happens — listen for result\nif (event.event === 'subscription.charged') {\n  // Success\n} else if (event.event === 'subscription.halted') {\n  // Debit failed — notify customer, retry logic needed\n}",
		DocURL:        "https://razorpay.com/docs/payments/recurring-payments/nach/",
		GuardrailRefs: []string{"RZP-006", "RZP-015"},
	},
	{
		ID:            "doc-007",
		Topic:         "upi-collect",
		Title:         "UPI Collect — handling PENDING state",
		Keywords:      []string{"upi", "collect", "vpa", "upi id", "pending", "async", "status"},
		Summary:       "UPI Collect payments go through a PENDING state immediately after the customer submits their VPA. This is normal and expected — the payment is not failed. Webhook confirmation arrives 1-5 minutes later.",
		CodeExample:   "// UPI payment flow\n\n// 1. Create order and open Checkout (same as standard)\nconst rzp = new Razorpay({\n  key: keyId,\n  order_id: orderId,\n  amount: amountInPaise,\n  method: 'upi',  // or let customer choose\n  handler: function(response) {\n    // ⚠️ Payment is PENDING here — NOT confirmed\n    // status will be 'pending' or 'created', not 'captured'\n    showPendingState('Payment pending — awaiting UPI confirmation...');\n  },\n});\n\n// 2. Poll for status (fallback if webhook is delayed)\nasync function pollPaymentStatus(paymentId, maxAttempts = 10) {\n  for (let i = 0; i < maxAttempts; i++) {\n    await sleep(i === 0 ? 3000 : 5000);  // wait before first poll\n    const payment = await razorpay.payments.fetch(paymentId);\n\n    if (payment.status === 'captured') return 'success';\n    if (payment.status === 'failed') return 'failed';\n    // 'pending' or 'created' → keep polling\n  }\n  return 'timeout';\n}\n\n// 3. Webhook: payment.captured arrives async (source of truth)\nif (event.event === 'payment.captured') {\n  fulfillOrder(event.payload.payment.entity.order_id);\n}",
		DocURL:        "https://razorpay.com/docs/payments/payment-methods/upi/",
		GuardrailRefs: []string{"RZP-007", "RZP-001"},
	},
	{
		ID:            "doc-008",
		Topic:         "upi-intent",
		Title:         "UPI Intent vs UPI QR — choosing the right flow",
		Keywords:      []string{"upi intent", "upi qr", "mobile", "desktop", "deep link", "qr code", "upi flow"},
		Summary:       "UPI Intent works on mobile (deep-links to UPI apps like PhonePe, GPay). UPI QR works on desktop (customer scans with their phone). Use Intent for mobile, QR for desktop.",
		CodeExample:   "// Detect device and choose flow\nconst isMobile = /Android|iPhone|iPad/.test(navigator.userAgent);\n\nconst rzp = new Razorpay({\n  key: keyId,\n  order_id: orderId,\n  amount: amountInPaise,\n\n  // Recommend UPI method based on device\n  config: {\n    display: {\n      blocks: {\n        upi: { name: 'UPI', instruments: [\n          isMobile\n            ? { method: 'upi', flows: ['intent'] }  // mobile: intent\n            : { method: 'upi', flows: ['qr'] }        // desktop: QR\n        ]}\n      },\n      sequence: ['block.upi'],\n      preferences: { show_default_blocks: false }\n    }\n  }\n});\nrzp.open();",
		DocURL:        "https://razorpay.com/docs/payments/payment-methods/upi/",
		GuardrailRefs: []string{"RZP-008"},
	},
	{
		ID:            "doc-009",
		Topic:         "route-api",
		Title:         "Route API — marketplace payment splits",
		Keywords:      []string{"route", "transfer", "split", "marketplace", "linked account", "platform fee", "vendor", "partner"},
		Summary:       "Route API lets you split payments between your platform and vendor/seller accounts. Transfers must happen AFTER payment.captured webhook — never from client callback. Route must be enabled on your account.",
		CodeExample:   "// Route API — initiate transfer after payment.captured webhook\n\nif (event.event === 'payment.captured') {\n  const payment = event.payload.payment.entity;\n  const order = await getOrderFromDB(payment.order_id);\n\n  // Initiate transfer ONLY after payment is captured\n  await razorpay.payments.transfer(payment.id, {\n    transfers: [\n      {\n        account: 'acc_linked_vendor_id',    // vendor's linked account\n        amount: order.vendorAmount,          // in paise\n        currency: 'INR',\n        on_hold: false,\n        linked_account_notes: ['vendor_order_id'],\n        notes: { order_id: order.id }\n      }\n    ]\n  });\n}\n\n// Listen for transfer confirmation\nif (event.event === 'transfer.processed') {\n  const transfer = event.payload.transfer.entity;\n  updateTransferStatus(transfer.id, 'completed');\n}\n\nif (event.event === 'transfer.failed') {\n  // Handle failed transfer — retry or alert\n  alertTransferFailure(transfer.id);\n}",
		DocURL:        "https://razorpay.com/docs/payments/route/",
		GuardrailRefs: []string{"RZP-017", "RZP-018"},
	},
	{
		ID:            "doc-010",
		Topic:         "webhook-setup",
		Title:         "Setting up and testing webhooks",
		Keywords:      []string{"webhook", "setup", "configure", "events", "endpoint", "test webhook", "ngrok", "local", "dashboard"},
		Summary:       "Configure webhook endpoints in the Razorpay Dashboard. For local testing, use ngrok or the Razorpay CLI (rzp listen) to forward webhooks to localhost.",
		CodeExample:   "// 1. In Dashboard: Settings > Webhooks > Add New Webhook\n// URL: https://your-domain.com/webhooks/razorpay\n// Select events: payment.captured, payment.failed, subscription.charged, etc.\n\n// 2. For local development — using Razorpay CLI\n// rzp listen --forward-to localhost:3000/webhooks/razorpay\n\n// 3. Webhook handler setup (Express)\n// CRITICAL: raw body middleware MUST come before express.json()\napp.use('/webhooks/razorpay', express.raw({ type: 'application/json' }));\napp.use(express.json());  // this comes AFTER for all other routes\n\napp.post('/webhooks/razorpay', (req, res) => {\n  const isValid = Razorpay.validateWebhookSignature(\n    req.body.toString(),\n    req.headers['x-razorpay-signature'],\n    process.env.RAZORPAY_WEBHOOK_SECRET\n  );\n  if (!isValid) return res.status(400).json({ error: 'Invalid signature' });\n\n  const event = JSON.parse(req.body);\n  // handle event...\n  res.json({ status: 'ok' });\n});",
		DocURL:        "https://razorpay.com/docs/webhooks/",
		GuardrailRefs: []string{"RZP-001", "RZP-004"},
	},
	{
		ID:            "doc-011",
		Topic:         "payment-link",
		Title:         "Payment Links — no-code payment collection",
		Keywords:      []string{"payment link", "payment_links", "link", "share link", "no-code", "invoice", "expire_by", "sms", "email"},
		Summary:       "Payment Links allow you to collect payments without a website or checkout integration. Create via API or Dashboard, share via SMS/email. Always set expire_by for time-sensitive payments.",
		CodeExample:   "// Create a payment link\nconst link = await razorpay.paymentLink.create({\n  amount: 99900,              // in paise — ₹999\n  currency: 'INR',\n  accept_partial: false,\n  expire_by: Math.floor(Date.now() / 1000) + (7 * 24 * 3600), // 7 days\n  reference_id: `order_${Date.now()}`,\n  description: 'Payment for Order #1234',\n  customer: {\n    name: 'Arjun Mehta',\n    email: 'arjun@example.com',\n    contact: '+919876543210',\n  },\n  notify: { sms: true, email: true },\n  reminder_enable: true,\n  callback_url: 'https://your-domain.com/payment-success',\n  callback_method: 'get',\n});\n\nconsole.log(link.short_url);  // Share this with customer\n\n// Webhook: listen for payment\nif (event.event === 'payment_link.paid') {\n  const pl = event.payload.payment_link.entity;\n  fulfillOrder(pl.reference_id);\n}",
		DocURL:        "https://razorpay.com/docs/payments/payment-links/",
		GuardrailRefs: nil,
	},
	{
		ID:            "doc-012",
		Topic:         "refunds",
		Title:         "Processing refunds",
		Keywords:      []string{"refund", "refund payment", "partial refund", "full refund", "instant refund", "refund.processed", "refund status"},
		Summary:       "Refunds are initiated via API and confirmed asynchronously via webhook. UPI/bank refunds take 5-7 business days. Always use refund.processed webhook as the source of truth, not the API response.",
		CodeExample:   "// Full refund\nconst refund = await razorpay.payments.refund(paymentId, {\n  speed: 'normal',  // or 'optimum' for instant (additional fee)\n  notes: { reason: 'Customer request' },\n});\n\n// Partial refund\nconst partialRefund = await razorpay.payments.refund(paymentId, {\n  amount: 50000,    // partial amount in paise\n  speed: 'normal',\n});\n\n// Webhook: refund status (source of truth)\nif (event.event === 'refund.processed') {\n  const refund = event.payload.refund.entity;\n  markRefundComplete(refund.payment_id, refund.id);\n}\n\nif (event.event === 'refund.failed') {\n  const refund = event.payload.refund.entity;\n  // Initiate alternative refund or manual bank transfer\n  alertRefundFailure(refund.id);\n}\n\n// IMPORTANT: API response status 'pending' is normal for UPI/bank refunds\n// Do NOT use API response as final status — wait for webhook",
		DocURL:        "https://razorpay.com/docs/api/refunds/",
		GuardrailRefs: nil,
	},
	{
		ID:            "doc-013",
		Topic:         "subscriptions",
		Title:         "Subscriptions API — complete setup",
		Keywords:      []string{"subscription", "plan", "recurring", "billing", "interval", "total_count", "subscription.charged", "subscription.halted"},
		Summary:       "Razorpay Subscriptions require: Create Plan → Create Subscription → Authenticate via Checkout → Handle webhooks (charged + halted). Use interval (not billing_cycle_anchor) for India.",
		CodeExample:   "// Step 1: Create plan\nconst plan = await razorpay.plans.create({\n  period: 'monthly',\n  interval: 1,\n  item: {\n    name: 'Pro License',\n    amount: 99900,  // ₹999/month in paise\n    currency: 'INR',\n  }\n});\n\n// Step 2: Create subscription\nconst sub = await razorpay.subscriptions.create({\n  plan_id: plan.id,\n  total_count: 12,  // 12 months\n  quantity: 1,\n  // Use interval for India — NOT billing_cycle_anchor\n  // billing_cycle_anchor causes issues with India payment rails\n});\n\n// Step 3: Checkout with subscription_id (not order_id)\nconst rzp = new Razorpay({\n  key: keyId,\n  subscription_id: sub.id,  // ← subscription_id, not order_id\n  handler: function(response) {\n    showPendingState('Setting up subscription...');\n  },\n});\nrzp.open();\n\n// Step 4: Webhooks\n// subscription.authenticated → mandate setup complete\n// subscription.charged → billing success → fulfill\n// subscription.halted → billing failed → notify customer\n// subscription.cancelled → customer cancelled",
		DocURL:        "https://razorpay.com/docs/payments/subscriptions/",
		GuardrailRefs: []string{"RZP-015", "RZP-016"},
	},
	{
		ID:            "doc-014",
		Topic:         "test-integration",
		Title:         "Testing your Razorpay integration",
		Keywords:      []string{"test", "testing", "test card", "test upi", "test credentials", "sandbox", "test mode", "4111"},
		Summary:       "Use test API keys (rzp_test_...) and test card numbers to verify your integration without real money movement.",
		CodeExample:   "// Test credentials\n// Card: 4111 1111 1111 1111  (Visa)\n//       5267 3181 8797 5449  (Mastercard)\n// Expiry: any future date (e.g., 12/26)\n// CVV: any 3 digits\n// OTP: any 4 digits (e.g., 1234)\n\n// Test UPI ID: success@razorpay\n// Test UPI ID for failure: failure@razorpay\n\n// Test Net Banking: any bank, any credentials\n\n// Test keys\nRAZORPAY_KEY_ID=rzp_test_xxxxxxxxxxxx   # starts with rzp_test_\nRAZORPAY_KEY_SECRET=xxxxxxxxxxxxxxxxxxxxxx\n\n// NEVER use test keys in production\n// NEVER use live keys in development",
		DocURL:        "https://razorpay.com/docs/payments/payment-gateway/test-payment/",
		GuardrailRefs: []string{"RZP-010"},
	},
	{
		ID:            "doc-015",
		Topic:         "error-handling",
		Title:         "Handling payment failures — payment.failed handler",
		Keywords:      []string{"payment failed", "payment.failed", "error", "handler", "failure", "retry", "decline"},
		Summary:       "Always implement the payment.failed event handler on the Razorpay Checkout instance. Without it, customers see a blank screen when payment fails and cannot retry.",
		CodeExample:   "const rzp = new Razorpay(options);\n\n// ✅ REQUIRED: payment.failed handler\nrzp.on('payment.failed', function(response) {\n  console.error('Payment failed:', response.error);\n\n  // response.error fields:\n  // .code          — error code (e.g., 'BAD_REQUEST_ERROR')\n  // .description   — human-readable description\n  // .source        — 'customer' or 'business'\n  // .step          — where in flow it failed\n  // .reason        — machine-readable reason\n  // .metadata.payment_id  — payment ID for retry\n  // .metadata.order_id    — order ID\n\n  showErrorMessage(response.error.description);\n  enableRetryButton();\n\n  // Log for debugging\n  logPaymentFailure({\n    paymentId: response.error.metadata.payment_id,\n    reason: response.error.reason,\n    code: response.error.code,\n  });\n});\n\nrzp.open();",
		DocURL:        "https://razorpay.com/docs/payments/payment-gateway/web-integration/standard/#step-5-handle-payment-success-and-failure",
		GuardrailRefs: []string{"RZP-014"},
	},
	{
		ID:            "doc-016",
		Topic:         "capture-payment",
		Title:         "Capturing payments — auto vs manual",
		Keywords:      []string{"capture", "auto capture", "manual capture", "authorized", "payment capture", "capture payment"},
		Summary:       "Most Razorpay accounts use auto-capture (enabled by default). With auto-capture, payments go directly from authorized to captured — no manual capture call needed. Only disable auto-capture if you need to review orders before charging.",
		CodeExample:   "// Check if manual capture is needed\n// Dashboard > Settings > Payment Capture\n\n// If manual capture is enabled:\napp.post('/webhooks/razorpay', async (req, res) => {\n  const event = JSON.parse(req.body);\n\n  if (event.event === 'payment.authorized') {\n    const payment = event.payload.payment.entity;\n    // Review order before capturing\n    const order = await getOrder(payment.order_id);\n    if (order.isValid) {\n      await razorpay.payments.capture(payment.id, payment.amount);\n    } else {\n      await razorpay.payments.refund(payment.id);  // refund unauthorized order\n    }\n  }\n});\n\n// With auto-capture (most common):\n// Listen only for payment.captured — no capture call needed\nif (event.event === 'payment.captured') {\n  fulfillOrder(payment.order_id);\n}",
		DocURL:        "https://razorpay.com/docs/api/payments/#capture-a-payment",
		GuardrailRefs: []string{"RZP-003"},
	},
	{
		ID:            "doc-017",
		Topic:         "first-integration",
		Title:         "First-time Razorpay integration — complete quickstart",
		Keywords:      []string{"getting started", "quickstart", "integrate", "first time", "setup", "how to", "start", "begin"},
		Summary:       "Complete 5-step integration: install SDK → create order (server) → open Checkout (client) → verify via webhook (server) → handle failures.",
		CodeExample:   "// Complete integration in 5 steps\n\n// Step 1: Install\n// npm install razorpay dotenv\n\n// Step 2: Server — create order (server.js)\nconst razorpay = new Razorpay({\n  key_id: process.env.RAZORPAY_KEY_ID,\n  key_secret: process.env.RAZORPAY_KEY_SECRET,\n});\n\napp.post('/api/create-order', async (req, res) => {\n  const order = await razorpay.orders.create({\n    amount: 99900,   // ₹999 in paise\n    currency: 'INR',\n    receipt: `rcpt_${Date.now()}`,\n  });\n  res.json({ orderId: order.id, amount: order.amount,\n             keyId: process.env.RAZORPAY_KEY_ID });\n});\n\n// Step 3: Client — open Checkout (index.html)\nconst order = await fetch('/api/create-order', {method:'POST'}).then(r=>r.json());\nconst rzp = new Razorpay({\n  key: order.keyId, order_id: order.orderId, amount: order.amount,\n  handler: (r) => showPending('Verifying...'),\n});\nrzp.on('payment.failed', (r) => showError(r.error.description));\nrzp.open();\n\n// Step 4: Server — webhook verification\napp.post('/webhooks', express.raw({type:'application/json'}), (req, res) => {\n  Razorpay.validateWebhookSignature(req.body.toString(),\n    req.headers['x-razorpay-signature'], process.env.RAZORPAY_WEBHOOK_SECRET);\n  const event = JSON.parse(req.body);\n  if (event.event === 'payment.captured') fulfillOrder(event.payload.payment.entity);\n  res.json({status:'ok'});\n});\n\n// Step 5: Test card: 4111 1111 1111 1111 | any exp | any CVV | OTP: 1234",
		DocURL:        "https://razorpay.com/docs/payments/payment-gateway/",
		GuardrailRefs: []string{"RZP-001", "RZP-002", "RZP-009", "RZP-013", "RZP-014"},
	},
}

var errorRegistryData = []errorEntry{
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "order_id is required",
		Title:           "Missing Order ID",
		Explanation:     "Payment cannot be initiated or captured without linking it to a Razorpay Order. You must create an Order via the Orders API before opening Checkout.",
		CommonCauses:    []string{"Opening Razorpay Checkout without creating a server-side order first", "Not passing order_id in the Checkout options object", "Calling POST /v1/payments/{id}/capture without an order_id"},
		ResolutionSteps: []string{"Create an order first: POST /v1/orders with amount (in paise), currency, receipt", "Pass the returned order.id as order_id in your Checkout options", "Never skip order creation — it is mandatory for all standard payment flows"},
		DocURL:          "https://razorpay.com/docs/api/orders/#create-an-order",
		IsRetriable:     false,
		GuardrailRef:    "RZP-002",
		GuardrailTitle:  "Create an order via /v1/orders before initiating payment",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "The payment has already been captured",
		Title:           "Payment Already Captured",
		Explanation:     "You attempted to capture a payment that has already been captured. Each payment can only be captured once.",
		CommonCauses:    []string{"Duplicate webhook delivery causing double-capture attempt", "Race condition between webhook handler and polling fallback", "Not implementing idempotency checks before capture"},
		ResolutionSteps: []string{"Check payment status via GET /v1/payments/{id} before attempting capture", "Implement idempotency: store payment_id and skip if already processed", "Use webhook signature validation to deduplicate webhook events"},
		DocURL:          "https://razorpay.com/docs/api/payments/#capture-a-payment",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Invalid key_id",
		Title:           "Invalid API Key ID",
		Explanation:     "The key_id provided does not match any active Razorpay account. Check that you're using the correct key for the environment (test vs live).",
		CommonCauses:    []string{"Using a test key (rzp_test_...) in production or vice versa", "Typo or truncation in the key_id value", "Using a deactivated or deleted API key"},
		ResolutionSteps: []string{"Verify key format: test keys start with rzp_test_, live keys with rzp_live_", "Copy the key directly from Dashboard > Settings > API Keys", "Ensure environment variables are loaded correctly (.env file, dotenv config)"},
		DocURL:          "https://razorpay.com/docs/api/authentication/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-010",
		GuardrailTitle:  "Use test keys (rzp_test_) in development/staging, live keys (rzp_live_) in production only",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "The api key and secret provided are invalid",
		Title:           "Invalid API Key or Secret",
		Explanation:     "Authentication failed. The key_id and key_secret combination is not valid.",
		CommonCauses:    []string{"key_secret exposed in frontend code and rotated by security scan", "Regenerated keys in Dashboard but old keys still in code", "Extra whitespace or newline characters in the secret"},
		ResolutionSteps: []string{"Never put key_secret in frontend/client-side code", "Regenerate keys in Dashboard if exposed: Settings > API Keys > Regenerate", "Store key_secret only in server-side environment variables", "Trim whitespace when reading from environment: process.env.KEY_SECRET.trim()"},
		DocURL:          "https://razorpay.com/docs/api/authentication/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-009",
		GuardrailTitle:  "Never expose key_secret in frontend, browser, or mobile code",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Payment signature verification failed",
		Title:           "Webhook Signature Verification Failed",
		Explanation:     "The HMAC-SHA256 signature in the x-razorpay-signature header does not match the computed signature. The webhook payload may have been tampered with, or the verification is being done on the wrong data.",
		CommonCauses:    []string{"Verifying signature on parsed JSON instead of raw request body (Buffer)", "Using the wrong secret — payment verification secret vs webhook secret", "Body parsing middleware consuming the raw body before signature check", "Incorrect webhook secret copied from Dashboard"},
		ResolutionSteps: []string{"Use raw body for verification: express.raw({ type: 'application/json' }) BEFORE express.json()", "Register the raw body route before any JSON parsing middleware", "Verify using Razorpay.validateWebhookSignature(rawBody.toString(), signature, webhookSecret)", "Webhook secret is different from key_secret — get it from Dashboard > Webhooks"},
		DocURL:          "https://razorpay.com/docs/webhooks/validate-test/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-004",
		GuardrailTitle:  "Use raw (unparsed) request body for webhook signature validation",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "amount is required",
		Title:           "Missing Amount",
		Explanation:     "The amount field is required when creating an order or payment.",
		CommonCauses:    []string{"Passing amount in rupees instead of paise (e.g., 999 instead of 99900)", "Amount field missing from the request body", "Amount is zero or undefined"},
		ResolutionSteps: []string{"Always pass amount in paise: ₹999 = 99900, ₹4999 = 499900", "Minimum amount is ₹1 (100 paise)", "Convert: const amountInPaise = Math.round(amountInRupees * 100)"},
		DocURL:          "https://razorpay.com/docs/api/orders/#create-an-order",
		IsRetriable:     false,
		GuardrailRef:    "RZP-013",
		GuardrailTitle:  "Always pass amount in paise (rupees × 100), not rupees",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Invalid amount. Amount should be in paise",
		Title:           "Amount Not in Paise",
		Explanation:     "Razorpay requires all amounts in the smallest currency unit (paise for INR). Passing ₹999 as 999 instead of 99900 will result in a ₹9.99 charge.",
		CommonCauses:    []string{"AI-generated code that passes rupee amounts directly", "Copying code from Stripe which uses cents but not converting for INR", "Decimal amounts (e.g., 999.50) — amounts must be integers"},
		ResolutionSteps: []string{"Multiply rupees by 100: amount_in_paise = rupee_amount * 100", "Always use integers — round if necessary: Math.round(price * 100)", "Store and pass amounts as integers throughout your system"},
		DocURL:          "https://razorpay.com/docs/payments/payment-gateway/web-integration/standard/#step-3-display-checkout",
		IsRetriable:     false,
		GuardrailRef:    "RZP-013",
		GuardrailTitle:  "Always pass amount in paise (rupees × 100), not rupees",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Route is not enabled for your account",
		Title:           "Route API Not Enabled",
		Explanation:     "The Route API for marketplace payments/transfers is a separate product that must be enabled on your account before use.",
		CommonCauses:    []string{"Attempting to create transfers without requesting Route access", "Using Route in test mode before enabling in Dashboard", "Account type doesn't support Route (requires specific business category)"},
		ResolutionSteps: []string{"Enable Route: Dashboard > Settings > Payment Route > Enable", "Contact support if the option is not visible — Route requires account review", "Test in test mode first after enabling"},
		DocURL:          "https://razorpay.com/docs/payments/route/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-017",
		GuardrailTitle:  "Check Route product enablement before initiating transfers — it requires explicit Razorpay approval",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "The payment is in a state where transfers cannot be initiated",
		Title:           "Transfer Before Payment Captured",
		Explanation:     "Transfers can only be initiated after a payment reaches the 'captured' state via webhook. Attempting to transfer before capture will fail.",
		CommonCauses:    []string{"Initiating transfer from the client-side payment handler (not webhook)", "Calling transfer API in payment.authorized webhook instead of payment.captured", "Race condition between capture and transfer"},
		ResolutionSteps: []string{"Only initiate transfers from within the payment.captured webhook handler", "Never initiate transfers from client-side callbacks", "Ensure your webhook endpoint receives and processes payment.captured events"},
		DocURL:          "https://razorpay.com/docs/payments/route/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-018",
		GuardrailTitle:  "Initiate Route transfers only after payment.captured webhook — never from client-side callback",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "International payments are not enabled for your account",
		Title:           "International Payments Not Enabled",
		Explanation:     "Accepting international cards or foreign currency payments requires explicit enablement on your Razorpay account.",
		CommonCauses:    []string{"Passing currency: 'USD' without enabling multi-currency", "Customer using an international card on an account without international payments", "Using test key — some international payment features require live key testing"},
		ResolutionSteps: []string{"Enable international payments: Dashboard > Settings > Checkout > International Payments", "For multi-currency: Dashboard > Settings > Multi-currency > Enable", "Both test and live keys need separate enablement"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/international-payments/",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "Payment declined by bank",
		Title:           "Payment Declined by Issuing Bank",
		Explanation:     "The customer's bank declined the payment. This is a terminal failure — the specific reason is known to the bank but not always disclosed to merchants.",
		CommonCauses:    []string{"Insufficient funds", "Card blocked for online transactions", "Bank's fraud prevention rules triggered", "Incorrect CVV or card details"},
		ResolutionSteps: []string{"Show the customer a clear error message and retry option", "Suggest alternative payment methods (UPI, NetBanking)", "Implement the payment.failed webhook handler to catch this event", "Do NOT automatically retry without customer action — this may flag fraud"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-014",
		GuardrailTitle:  "Always implement a payment.failed handler in Razorpay checkout",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "UPI transaction declined",
		Title:           "UPI Payment Declined",
		Explanation:     "The UPI payment was declined. This could be due to incorrect UPI PIN, insufficient balance, or the UPI app blocking the transaction.",
		CommonCauses:    []string{"Incorrect UPI PIN entered by customer", "Daily UPI transaction limit exceeded", "Bank server timeout or unavailability", "Invalid VPA (UPI ID)"},
		ResolutionSteps: []string{"Implement payment.failed webhook handler — UPI failures arrive asynchronously", "Do NOT treat UPI status 'pending' as failed — it's normal for UPI to be pending for 1-5 minutes", "Show retry option with different payment method", "For UPI Collect: set a polling fallback with GET /v1/payments/{id} with exponential backoff"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/upi/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-007",
		GuardrailTitle:  "Handle UPI PENDING state — do not treat it as failure or success",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "subscription_id is required for subscription payments",
		Title:           "Missing Subscription ID",
		Explanation:     "For subscription-based recurring payments, the subscription_id must be passed in Checkout options. This links the payment to the subscription plan.",
		CommonCauses:    []string{"Passing plan_id instead of subscription_id", "Creating a plan but not creating a subscription before opening Checkout", "Confusing subscription.id with plan.id"},
		ResolutionSteps: []string{"Flow: Create Plan → Create Subscription → pass subscription.id to Checkout", "POST /v1/subscriptions to create subscription, use returned id", "Checkout option: { subscription_id: sub.id } (not plan_id)"},
		DocURL:          "https://razorpay.com/docs/payments/subscriptions/",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "The subscription is not in a state to accept payments",
		Title:           "Subscription Not Active",
		Explanation:     "Payment was attempted on a subscription that is not in the 'created' or 'authenticated' state. Subscriptions in 'halted', 'cancelled', or 'expired' states cannot accept new payments.",
		CommonCauses:    []string{"Subscription.halted not handled — debit failed and subscription is now halted", "Subscription expired (billing_cycle_anchor issue for India)", "Cancelled subscription — customer or merchant cancelled it"},
		ResolutionSteps: []string{"Listen to subscription.halted webhook and notify the customer to update payment method", "Use interval (not billing_cycle_anchor) for India-based subscription schedules", "Check subscription status: GET /v1/subscriptions/{id} before initiating payment"},
		DocURL:          "https://razorpay.com/docs/payments/subscriptions/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-015",
		GuardrailTitle:  "Handle both subscription.charged AND subscription.halted webhooks",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "AFA (Additional Factor Authentication) required for this transaction",
		Title:           "AFA Required for High-Value UPI AutoPay",
		Explanation:     "UPI AutoPay mandates with amounts above ₹15,000 require Additional Factor Authentication (AFA) per RBI mandate. The customer must complete an extra authentication step via their bank/UPI app.",
		CommonCauses:    []string{"Not handling the afa_required field in subscription.authenticated webhook", "Not implementing redirect_url for AFA flow in Checkout options", "Assuming UPI AutoPay > ₹15K works like regular < ₹15K flow"},
		ResolutionSteps: []string{"Check subscription.authenticated webhook for afa_required: true", "If afa_required, redirect customer to the afa_url provided in the webhook", "Customer completes authentication in their bank app/UPI app", "Listen for subscription.authenticated with afa_completed: true before considering mandate active"},
		DocURL:          "https://razorpay.com/docs/payments/recurring-payments/upi/autopay-afa/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-005",
		GuardrailTitle:  "UPI AutoPay mandates above ₹15,000 require AFA redirect — handle it explicitly",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "NACH mandate requires pre-debit notification",
		Title:           "NACH Pre-Debit Notification Required",
		Explanation:     "RBI requires merchants to send customers a pre-debit notification at least 24 hours before each NACH (National Automated Clearing House) debit. Skipping this is a compliance violation.",
		CommonCauses:    []string{"Assuming NACH debits automatically like UPI AutoPay", "Not implementing the pre-debit notification step in the subscription flow", "Debiting too soon after mandate creation (bank verification takes 24-48h)"},
		ResolutionSteps: []string{"Send pre-debit notification 24h before each planned debit", "Use Razorpay's notification API or your own email/SMS system", "NACH mandate bank verification: wait for subscription.authenticated webhook (takes 24-48h)", "Do not attempt debit before receiving subscription.authenticated webhook"},
		DocURL:          "https://razorpay.com/docs/payments/recurring-payments/nach/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-006",
		GuardrailTitle:  "NACH mandates require 24-hour pre-debit notification — implement before each debit",
	},
	{
		Code:            "SERVER_ERROR",
		SubDescription:  "An error occurred on our servers",
		Title:           "Razorpay Server Error",
		Explanation:     "Razorpay's servers encountered an unexpected error. This is a temporary issue on Razorpay's side.",
		CommonCauses:    []string{"Temporary service degradation", "High traffic period", "Dependent service (bank, payment network) unavailable"},
		ResolutionSteps: []string{"Retry the request after a short delay (exponential backoff: 1s, 2s, 4s)", "Check Razorpay status page: status.razorpay.com", "For payment status: poll GET /v1/payments/{id} rather than retrying the payment"},
		DocURL:          "https://razorpay.com/docs/api/errors/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-ERR-006",
		GuardrailTitle:  "Implement a circuit breaker for Razorpay API calls — degrade gracefully during outages",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "checkout.js must be loaded from the official Razorpay CDN",
		Title:           "Unofficial Checkout.js URL",
		Explanation:     "Razorpay Checkout must be loaded only from https://checkout.razorpay.com/v1/checkout.js. Loading from any other source (cached, self-hosted, CDN mirror) is a security risk and may result in stale SDK versions.",
		CommonCauses:    []string{"Bundling checkout.js into your webpack/vite build", "Loading from a cached version or alternate URL", "npm-installing checkout.js instead of using the CDN script tag"},
		ResolutionSteps: []string{"Use ONLY: <script src=\"https://checkout.razorpay.com/v1/checkout.js\"></script>", "Never bundle, copy, or self-host this file", "Load it as a <script> tag in HTML, not as an npm import"},
		DocURL:          "https://razorpay.com/docs/payments/payment-gateway/web-integration/standard/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-012",
		GuardrailTitle:  "Load checkout.js only from the official Razorpay CDN URL",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "contact is required",
		Title:           "Missing Customer Contact",
		Explanation:     "The customer's phone number (contact) is required in Checkout prefill options for certain payment methods (UPI, Net Banking).",
		CommonCauses:    []string{"Omitting prefill object from Checkout options", "Passing contact without the country code prefix (+91)"},
		ResolutionSteps: []string{"Add prefill to Checkout options: { name, email, contact: '+91XXXXXXXXXX' }", "Contact must include country code: +91 for India", "Can be pre-filled from your user database to improve checkout UX"},
		DocURL:          "https://razorpay.com/docs/payments/payment-gateway/web-integration/standard/#step-3-display-checkout",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Transfer amount exceeds payment amount",
		Title:           "Route Transfer Amount Exceeds Payment",
		Explanation:     "When using Route API, the total amount transferred to linked accounts cannot exceed the original payment amount.",
		CommonCauses:    []string{"Miscalculating platform fee vs transfer amount", "Attempting multiple transfers that sum to more than the payment", "Not accounting for Razorpay processing fees in transfer calculation"},
		ResolutionSteps: []string{"Total transfers ≤ payment amount after Razorpay fees", "Calculate: transfer_amount = payment_amount - platform_commission - razorpay_fee", "Use the linked_account_notes to document the split"},
		DocURL:          "https://razorpay.com/docs/payments/route/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-018",
		GuardrailTitle:  "Initiate Route transfers only after payment.captured webhook — never from client-side callback",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "currency is not supported",
		Title:           "Unsupported Currency",
		Explanation:     "The currency code passed is not supported by your account or for the selected payment method.",
		CommonCauses:    []string{"Passing 'USD' without enabling multi-currency on the account", "Using an invalid ISO 4217 currency code", "Payment method doesn't support the requested currency (UPI is INR only)"},
		ResolutionSteps: []string{"Domestic Indian payments: always use currency: 'INR'", "International: enable multi-currency in Dashboard first", "UPI, NACH, NetBanking: INR only — cannot be used for foreign currency payments"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/international-payments/",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "receipt must be unique",
		Title:           "Duplicate Receipt ID",
		Explanation:     "The receipt field must be unique across all orders in your account. You're trying to create an order with a receipt that already exists.",
		CommonCauses:    []string{"Using a non-unique receipt like 'order_1' or a static test value", "Not generating a unique receipt per order (UUID or timestamp-based)", "Retry logic creating the same order twice with the same receipt", "Hard-coded receipt value in development"},
		ResolutionSteps: []string{"Generate unique receipts: `receipt_${uuid()}` or `receipt_${Date.now()}`", "Use your internal order ID as the receipt for easy correlation", "On retry, use the same receipt (idempotency) or generate a new unique one"},
		DocURL:          "https://razorpay.com/docs/api/orders/#create-an-order",
		IsRetriable:     true,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Payment link has expired",
		Title:           "Payment Link Expired",
		Explanation:     "The payment link has passed its expiry date and can no longer accept payments.",
		CommonCauses:    []string{"Not setting expire_by when creating the link (defaults to 6 months)", "Customer accessing an old link", "Sending links without an expiry and customer paying late"},
		ResolutionSteps: []string{"Create a new payment link: POST /v1/payment_links", "Always set expire_by for time-sensitive transactions", "Listen for payment_link.expired webhook to notify customers proactively"},
		DocURL:          "https://razorpay.com/docs/payments/payment-links/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-PL-001",
		GuardrailTitle:  "Always set expire_by on payment links — links are permanently active without it",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Linked account does not exist",
		Title:           "Route Linked Account Not Found",
		Explanation:     "The linked account ID passed in a Route transfer does not exist or belongs to a different Razorpay account.",
		CommonCauses:    []string{"Using a test mode linked account ID in live mode (or vice versa)", "Linked account was deactivated", "Typo in the linked account ID"},
		ResolutionSteps: []string{"Verify linked account: GET /v1/beta/accounts/{id}", "Test and live linked accounts are separate — use matching keys", "Linked accounts must be created by the platform account, not independently"},
		DocURL:          "https://razorpay.com/docs/payments/route/linked-accounts/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-017",
		GuardrailTitle:  "Check Route product enablement before initiating transfers — it requires explicit Razorpay approval",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "The payment is not in capturable state",
		Title:           "Payment Cannot Be Captured",
		Explanation:     "Capture was attempted on a payment that is not in the 'authorized' state. Only authorized payments can be captured.",
		CommonCauses:    []string{"Trying to capture a payment that has already been captured", "Trying to capture a payment that failed or was refunded", "Auto-capture is enabled on the account — no manual capture needed"},
		ResolutionSteps: []string{"Check payment status: GET /v1/payments/{id} — status must be 'authorized'", "For most accounts, auto-capture is ON — you don't need to call capture", "Check Dashboard > Settings > Payment Capture to see your capture setting"},
		DocURL:          "https://razorpay.com/docs/api/payments/#capture-a-payment",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Payment signature verification failed",
		Title:           "Webhook Signature Mismatch",
		Explanation:     "The HMAC-SHA256 signature in X-Razorpay-Signature does not match. Either the webhook secret is wrong, or you are passing a parsed/modified body instead of the raw request body.",
		CommonCauses:    []string{"Using req.body (JSON parsed) instead of raw buffer for HMAC computation", "Webhook secret in your code doesn't match the secret set in Razorpay Dashboard", "Body parsing middleware (express.json) consuming the raw body before webhook handler", "Copy-paste error in the webhook secret — trailing whitespace or partial copy"},
		ResolutionSteps: []string{"Use express.raw() or save rawBody in a middleware BEFORE express.json()", "Verify webhook secret matches exactly: Dashboard > Webhooks > Edit > Show Secret", "Use Razorpay SDK: razorpay.webhooks.validateWebhookSignature(body, signature, secret)", "Never modify or re-serialize the body before verification"},
		DocURL:          "https://razorpay.com/docs/webhooks/validate-test/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-004",
		GuardrailTitle:  "Use raw (unparsed) request body for webhook signature validation",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "The refund amount is greater than the payment amount",
		Title:           "Refund Exceeds Payment Amount",
		Explanation:     "The refund amount requested is greater than what was originally charged or what remains refundable after prior partial refunds.",
		CommonCauses:    []string{"Amount passed in rupees instead of paise (500 instead of 50000)", "Multiple partial refunds that cumulatively exceed original payment", "Not checking payment.amount_refunded before issuing another partial refund"},
		ResolutionSteps: []string{"Always pass refund amount in paise (multiply rupees by 100)", "Fetch payment first: const p = await rzp.payments.fetch(id)", "Max refundable = p.amount - (p.amount_refunded || 0)", "Validate before calling: if (requestedPaise > maxRefundable) throw error"},
		DocURL:          "https://razorpay.com/docs/api/refunds/#create-a-refund",
		IsRetriable:     false,
		GuardrailRef:    "RZP-REF-003",
		GuardrailTitle:  "Partial refund amount must be in paise and must not exceed the captured amount",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Payment is in authorized state, cannot be refunded",
		Title:           "Cannot Refund Authorized Payment",
		Explanation:     "Refunds can only be issued on captured payments. An authorized payment has not yet been captured — you must capture it first, or let the authorization expire.",
		CommonCauses:    []string{"Issuing a refund on a payment that was authorized but not captured", "Manual capture accounts where capture has not been called yet"},
		ResolutionSteps: []string{"Check payment.status — must be 'captured' before refunding", "If you don't want to charge: let authorization expire (5-7 days) — no refund needed", "If captured: rzp.payments.refund(paymentId, { amount: ... })"},
		DocURL:          "https://razorpay.com/docs/api/refunds/#create-a-refund",
		IsRetriable:     false,
		GuardrailRef:    "RZP-REF-001",
		GuardrailTitle:  "Verify payment status is captured before attempting a refund",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Payment link has already been paid",
		Title:           "Payment Link Already Paid",
		Explanation:     "An attempt was made to pay a payment link that has already been successfully paid.",
		CommonCauses:    []string{"Customer clicked a shared link multiple times", "Duplicate webhook causing duplicate fulfillment attempt", "Testing with the same link after it was already paid"},
		ResolutionSteps: []string{"Check payment_link.status via GET /v1/payment_links/{id} — status will be 'paid'", "Implement idempotency: check db for existing fulfillment before processing", "Use payment_link.paid webhook and deduplicate on payload.payment_link.entity.id"},
		DocURL:          "https://razorpay.com/docs/api/payment-links/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-PL-007",
		GuardrailTitle:  "Cancelling a payment link does not refund payments already made on it",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "expire_by must be greater than current time",
		Title:           "Payment Link Expiry in the Past",
		Explanation:     "The expire_by timestamp provided is in the past. You must set expire_by to a future Unix timestamp.",
		CommonCauses:    []string{"Using Date.now() (milliseconds) instead of Unix timestamp (seconds)", "Passing a date string instead of a Unix timestamp integer", "Off-by-one: subtracting instead of adding seconds to current time"},
		ResolutionSteps: []string{"Use: Math.floor(Date.now() / 1000) + (days * 86400)", "Do NOT use Date.now() directly — it's in milliseconds, Razorpay expects seconds", "Minimum: at least 15 minutes in the future; recommended: 7 days for invoices"},
		DocURL:          "https://razorpay.com/docs/api/payment-links/#create-payment-link",
		IsRetriable:     false,
		GuardrailRef:    "RZP-PL-001",
		GuardrailTitle:  "Always set expire_by on payment links — links are permanently active without it",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "VPA is not valid",
		Title:           "Invalid UPI VPA (UPI ID)",
		Explanation:     "The UPI Virtual Payment Address (VPA) provided does not exist or is not registered on the UPI network.",
		CommonCauses:    []string{"Typo in the UPI ID (user@okaxis, user@oksbi, etc.)", "UPI ID was deactivated or changed", "Testing with a fake/dummy UPI ID"},
		ResolutionSteps: []string{"Validate UPI ID format: GET /v1/payments/validate/vpa?vpa=user@okaxis", "Show inline validation before payment to catch typos early", "For test mode, use test UPI IDs from Razorpay test account dashboard"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/upi/",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "The transfer account is not linked",
		Title:           "Linked Account Not Active",
		Explanation:     "The Route transfer destination account exists but is not fully KYC-verified and active. Transfers can only go to fully onboarded linked accounts.",
		CommonCauses:    []string{"Linked account created but KYC not completed by the sub-merchant", "Linked account stakeholder details incomplete", "Account suspended due to compliance review"},
		ResolutionSteps: []string{"Check account status: GET /v2/accounts/{accountId}", "Share the KYC link with your sub-merchant to complete onboarding", "Contact Razorpay support if account appears stuck in review"},
		DocURL:          "https://razorpay.com/docs/api/partners/linked-account/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-RT-001",
		GuardrailTitle:  "Verify Route product is enabled on your account before initiating any transfer",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Transfer amount must be less than or equal to remaining payment amount",
		Title:           "Route Transfer Exceeds Available Amount",
		Explanation:     "The total amount being transferred via Route exceeds the original payment amount minus any prior transfers or platform fees.",
		CommonCauses:    []string{"Summing transfers in rupees instead of paise", "Not accounting for the platform fee and/or GST", "Multiple transfers initiated (duplicate webhook handling)"},
		ResolutionSteps: []string{"All amounts in paise. Platform commission + all transfers must ≤ payment.amount", "Use payment.amount_transferred to check what's already been transferred", "Implement idempotency on transfer initiation to prevent duplicates"},
		DocURL:          "https://razorpay.com/docs/api/payments/route/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-RT-005",
		GuardrailTitle:  "Collect platform fee via on_hold or transfer math — never via a separate charge",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "Transaction not permitted to cardholder",
		Title:           "Transaction Type Blocked by Bank",
		Explanation:     "The cardholder's bank has blocked this specific transaction type. The card exists and has funds, but the bank's rules prevent this category of transaction (e-commerce, international, recurring, etc.).",
		CommonCauses:    []string{"E-commerce/online transactions not enabled on the card (common for new debit cards)", "International transactions disabled (card is domestic-only)", "Recurring/mandate payments not enabled on the card", "Corporate card with merchant category code (MCC) restrictions"},
		ResolutionSteps: []string{"Ask the customer to enable online transactions: Bank App > Card Settings > Enable Online Transactions", "Suggest alternative payment method: UPI or NetBanking (not card-restricted)", "Do NOT retry automatically — the card will keep declining until customer unblocks it", "Log with payment.failed webhook: reason='transaction_not_permitted'"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/cards/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-014",
		GuardrailTitle:  "Always implement a payment.failed handler in Razorpay checkout",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "Do not honor",
		Title:           "Bank Generic Decline (Do Not Honor)",
		Explanation:     "'Do Not Honor' is a bank's generic decline code when the bank blocks the transaction without disclosing the exact reason. It covers fraud flags, risk limits, and policy restrictions.",
		CommonCauses:    []string{"Bank's fraud scoring flagged the transaction (unusual merchant, amount, or location)", "Card has a standing block due to prior dispute or missed payment", "Daily/monthly spend limit reached on the card", "Bank system anomaly — transient internal block"},
		ResolutionSteps: []string{"Ask customer to try an alternative payment method (UPI, different card, NetBanking)", "If retry is needed, wait at least 10 minutes before re-attempting", "Customer can call their bank to unblock the card for the specific merchant", "Do NOT auto-retry — repeated declined attempts may further flag the card"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/cards/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-014",
		GuardrailTitle:  "Always implement a payment.failed handler in Razorpay checkout",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "Card expired",
		Title:           "Expired Card",
		Explanation:     "The customer's card has passed its expiry date. The bank rejects all transactions on expired cards regardless of available balance.",
		CommonCauses:    []string{"Customer using an old/expired card saved in their browser autofill", "Saved card in your system is past expiry", "Customer entered the wrong expiry date"},
		ResolutionSteps: []string{"Show clear error: 'Your card has expired. Please use a different card.'", "Prompt customer to enter a new/renewed card or use UPI", "If you store saved cards: periodically check expiry and prompt customers to update", "Remove expired cards from your saved-card store to prevent this in future"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/cards/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-014",
		GuardrailTitle:  "Always implement a payment.failed handler in Razorpay checkout",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "Invalid CVV",
		Title:           "Invalid Card CVV",
		Explanation:     "The CVV (Card Verification Value) entered by the customer does not match what the bank has on record for the card.",
		CommonCauses:    []string{"Customer entered wrong 3-digit CVV from the back of the card", "For Amex: 4-digit CVV on the front (not back) entered incorrectly", "Customer using a virtual card with a different CVV"},
		ResolutionSteps: []string{"Show error: 'Incorrect CVV. Please check the 3-digit code on the back of your card.'", "Allow customer to retry — this is user error, not a code/integration issue", "Do NOT log or store CVV anywhere — it is PCI-DSS prohibited", "After 3 consecutive CVV failures, suggest switching to UPI"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/cards/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-014",
		GuardrailTitle:  "Always implement a payment.failed handler in Razorpay checkout",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "Card reported lost or stolen",
		Title:           "Lost or Stolen Card",
		Explanation:     "The card used for payment has been reported as lost or stolen by the cardholder to their bank. The bank blocks all transactions as a fraud prevention measure.",
		CommonCauses:    []string{"Customer's physical card was lost/stolen and blocked by their bank", "Fraudster attempting to use a card that the owner has already blocked", "Bank proactively blocked the card after detecting fraud on it"},
		ResolutionSteps: []string{"Do NOT retry — this card is permanently blocked", "Suggest customer use an alternative payment method", "Flag the transaction in your system for review if the customer insists on retry", "Contact Razorpay support if you see a pattern of lost/stolen card attempts"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/cards/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-014",
		GuardrailTitle:  "Always implement a payment.failed handler in Razorpay checkout",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "3DS authentication failed",
		Title:           "3D Secure Authentication Failed",
		Explanation:     "The payment failed at the 3D Secure (OTP/bank authentication) step. The customer either failed authentication or the bank's 3DS service was unavailable.",
		CommonCauses:    []string{"Customer entered wrong OTP from bank SMS", "Customer closed the 3DS authentication popup/redirect before completing", "Bank's 3DS server was temporarily unavailable", "Authentication timed out (customer took too long to enter OTP)"},
		ResolutionSteps: []string{"Allow customer to retry payment — 3DS failure is retriable", "Show: 'Authentication failed. Please try again and complete the OTP step.'", "Implement payment.failed webhook handler — 3DS failures arrive here", "If bank 3DS server is consistently down, suggest UPI/NetBanking as fallback"},
		DocURL:          "https://razorpay.com/docs/payments/payment-gateway/3ds/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-014",
		GuardrailTitle:  "Always implement a payment.failed handler in Razorpay checkout",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "Network timeout — payment gateway unreachable",
		Title:           "Payment Gateway Timeout",
		Explanation:     "The request to Razorpay's payment processing infrastructure timed out. This is a transient infrastructure issue.",
		CommonCauses:    []string{"Razorpay infrastructure temporarily overloaded", "Network connectivity issue between your server and Razorpay", "Your server's outbound request timeout is too short (<10s)"},
		ResolutionSteps: []string{"Retry with exponential backoff: 1s, 2s, 4s delays (max 3 retries)", "Use idempotency key to prevent duplicate operations on retry", "Set request timeout to at least 30 seconds for Razorpay API calls", "Check https://status.razorpay.com for active incidents"},
		DocURL:          "https://razorpay.com/docs/api/errors/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-ERR-002",
		GuardrailTitle:  "Retry GATEWAY_ERROR with exponential backoff — these are transient",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "Issuing bank is temporarily unavailable",
		Title:           "Bank Server Down (Transient)",
		Explanation:     "The customer's bank is temporarily unreachable. This is a transient error — the bank's payment processing servers are down or overloaded.",
		CommonCauses:    []string{"Bank maintenance window (common on weekend nights in India)", "Bank infrastructure outage", "High traffic to bank (salary day, festival period)"},
		ResolutionSteps: []string{"Retry after 2-5 minutes", "Offer customer an alternative payment method (UPI, wallet, different bank)", "This is NOT a code bug — do not file a bug report without checking bank status"},
		DocURL:          "https://razorpay.com/docs/api/errors/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-ERR-002",
		GuardrailTitle:  "Retry GATEWAY_ERROR with exponential backoff — these are transient",
	},
	{
		Code:            "GATEWAY_ERROR",
		SubDescription:  "UPI PSP is temporarily unavailable",
		Title:           "UPI PSP Server Down",
		Explanation:     "The UPI Payment Service Provider (PSP) app server is temporarily unavailable. Common for specific UPI apps (GPay, PhonePe, Paytm) during outages.",
		CommonCauses:    []string{"Specific UPI PSP infrastructure outage (not all UPI affected)", "NPCI switch maintenance", "High volume during festivals/salary dates"},
		ResolutionSteps: []string{"Retry after 2-3 minutes", "If customer uses specific UPI app, suggest trying another (BHIM, bank's own UPI)", "Offer non-UPI fallback: netbanking, cards", "Check NPCI / individual PSP status pages"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/upi/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-ERR-002",
		GuardrailTitle:  "Retry GATEWAY_ERROR with exponential backoff — these are transient",
	},
	{
		Code:            "SERVER_ERROR",
		SubDescription:  "Internal server error",
		Title:           "Razorpay Internal Server Error",
		Explanation:     "Razorpay's servers encountered an unexpected error. This is not caused by your request — it's a server-side issue on Razorpay's infrastructure.",
		CommonCauses:    []string{"Razorpay infrastructure incident", "Deployment issues on Razorpay's end", "Rare edge cases in request processing"},
		ResolutionSteps: []string{"Check https://status.razorpay.com for active incidents", "Retry with exponential backoff (1s, 2s, 4s) — up to 3 attempts", "If persisting for >10 minutes, contact Razorpay support with request ID", "Implement circuit breaker to prevent cascading failures"},
		DocURL:          "https://razorpay.com/docs/api/errors/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-ERR-006",
		GuardrailTitle:  "Implement a circuit breaker for Razorpay API calls — degrade gracefully during outages",
	},
	{
		Code:            "SERVER_ERROR",
		SubDescription:  "Service temporarily unavailable",
		Title:           "Razorpay Service Unavailable (503)",
		Explanation:     "Razorpay's service is temporarily unavailable, typically due to maintenance or high load. Requests should be retried after a short wait.",
		CommonCauses:    []string{"Scheduled maintenance window", "Razorpay deployment in progress", "Traffic spike causing temporary capacity issues"},
		ResolutionSteps: []string{"Implement circuit breaker — stop hitting the API if >50% requests fail", "Retry after 30 seconds with exponential backoff", "Subscribe to status.razorpay.com for maintenance notifications", "Show user: 'Payment system temporarily unavailable, please try in a moment'"},
		DocURL:          "https://razorpay.com/docs/api/errors/",
		IsRetriable:     true,
		GuardrailRef:    "RZP-ERR-006",
		GuardrailTitle:  "Implement a circuit breaker for Razorpay API calls — degrade gracefully during outages",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "The customer's card does not support international transactions",
		Title:           "Card International Transactions Disabled",
		Explanation:     "The customer's card is not enabled for international/online transactions. This is a bank-level restriction on the card, not a Razorpay configuration issue.",
		CommonCauses:    []string{"Customer's bank has disabled online/international transactions by default", "Card is a debit card with online transactions disabled", "Customer needs to enable online transactions via their bank app/net banking"},
		ResolutionSteps: []string{"Ask customer to enable online transactions: Bank App > Card Settings > Enable Online Transactions", "Offer alternative: UPI, netbanking, wallet (not card-restricted)", "This is NOT fixable on your end — it's the customer's bank setting"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/cards/",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Insufficient funds in the account",
		Title:           "Insufficient Funds",
		Explanation:     "The customer's bank account or card does not have sufficient balance to complete the payment.",
		CommonCauses:    []string{"Customer account/card balance is lower than the payment amount", "Daily transaction limit reached on the card/account"},
		ResolutionSteps: []string{"Ask customer to try a different card or payment method", "Suggest UPI from a different bank account if card is low", "Show user-friendly message: 'Insufficient balance — please try another payment method'", "Do NOT retry automatically — customer must take action"},
		DocURL:          "https://razorpay.com/docs/api/errors/",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "OTP verification failed",
		Title:           "OTP Verification Failed",
		Explanation:     "The customer entered an incorrect OTP during 3DS/bank authentication. The payment attempt failed at the authentication step.",
		CommonCauses:    []string{"Customer entered wrong OTP", "OTP expired (typically 30-60 seconds validity)", "Customer entered previous OTP for a different transaction"},
		ResolutionSteps: []string{"Allow customer to retry payment — this is user error, not a code bug", "Show user: 'Incorrect OTP. Please request a new OTP and try again.'", "Do not auto-retry server-side — user must re-initiate the checkout flow", "Log payment.failed with reason='invalid_otp' for analytics"},
		DocURL:          "https://razorpay.com/docs/api/errors/",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "The plan_id is not valid",
		Title:           "Invalid Subscription Plan ID",
		Explanation:     "The plan_id provided when creating a subscription does not exist or belongs to a different account.",
		CommonCauses:    []string{"Plan created in test mode but subscription created in live mode (or vice versa)", "Plan ID from a different Razorpay account used", "Plan was deleted after being referenced in code"},
		ResolutionSteps: []string{"Verify plan exists: GET /v1/plans/{id}", "Ensure plan and subscription are created with the same API key (test/live consistency)", "Re-create the plan if it was deleted, update the plan_id in your config"},
		DocURL:          "https://razorpay.com/docs/api/subscriptions/#create-a-plan",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "webhook URL is not reachable",
		Title:           "Webhook URL Not Reachable",
		Explanation:     "Razorpay cannot reach your webhook URL. The URL returned a non-2xx response or timed out during the test ping.",
		CommonCauses:    []string{"Webhook URL pointing to localhost (not reachable from Razorpay's servers)", "Webhook URL is correct but endpoint doesn't exist (404)", "Server requires authentication that Razorpay doesn't provide", "SSL certificate error on your server"},
		ResolutionSteps: []string{"For local development, use ngrok: ngrok http 3000", "Ensure the webhook endpoint returns HTTP 200 within 5 seconds", "Check your server logs for incoming requests from Razorpay IPs", "Test via Dashboard: Webhooks > Test Webhook"},
		DocURL:          "https://razorpay.com/docs/webhooks/",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "The given key is disabled",
		Title:           "API Key Disabled",
		Explanation:     "The API key being used has been disabled in the Razorpay Dashboard, either manually or due to a security event.",
		CommonCauses:    []string{"Key was manually disabled after a security concern", "Account was temporarily suspended", "Key was part of a rotation and the old key was disabled"},
		ResolutionSteps: []string{"Go to Dashboard > Settings > API Keys — check if key is disabled", "Generate a new key if the old one is disabled", "Never use the same key in multiple environments simultaneously", "Rotate keys immediately if any exposure is suspected"},
		DocURL:          "https://razorpay.com/docs/api/authentication/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-011",
		GuardrailTitle:  "Rotate key_secret immediately if exposed — treat exposure as a security incident",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Method not allowed for this merchant",
		Title:           "Payment Method Not Enabled",
		Explanation:     "The payment method you're trying to use (e.g., EMI, specific wallet, international cards) is not enabled for your Razorpay account.",
		CommonCauses:    []string{"EMI not activated (requires separate activation for cards EMI)", "Specific wallet (Mobikwik, Airtel Money) not enabled", "International payments not enabled", "Newer payment method requires explicit activation"},
		ResolutionSteps: []string{"Check Dashboard > Payment Methods — see which methods are active", "Contact Razorpay support or your account manager to enable the method", "Filter checkout options to only show enabled methods to avoid this error"},
		DocURL:          "https://razorpay.com/docs/payments/payment-methods/",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "notes must have a maximum of 15 key-value pairs",
		Title:           "Too Many Notes Fields",
		Explanation:     "The notes object can have a maximum of 15 key-value pairs. You're passing more than 15 keys in the notes field.",
		CommonCauses:    []string{"Passing a large metadata object directly as notes", "Combining too many tracking fields into notes", "Iterating and adding all request fields to notes"},
		ResolutionSteps: []string{"Limit notes to the most important 15 fields for reconciliation", "Consolidate related fields: { order_meta: JSON.stringify({ field1, field2, ... }) }", "Store extra metadata in your own database, not in Razorpay notes"},
		DocURL:          "https://razorpay.com/docs/api/orders/#create-an-order",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "currency INR is required for subscription",
		Title:           "Subscription Requires INR Currency",
		Explanation:     "Razorpay Subscriptions only support INR. International currency subscriptions are not supported.",
		CommonCauses:    []string{"Trying to create a subscription for USD, EUR, or other foreign currency", "Multi-currency checkout configuration applied to subscription flows"},
		ResolutionSteps: []string{"Use currency: 'INR' for all subscription-related API calls", "For international customers: collect in INR and handle conversion externally", "For foreign currency recurring: consider Stripe or a cross-border solution"},
		DocURL:          "https://razorpay.com/docs/api/subscriptions/",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "Payment link is cancelled",
		Title:           "Payment Made on Cancelled Link",
		Explanation:     "An attempt was made to pay a payment link that has been cancelled. Cancelled links cannot accept new payments.",
		CommonCauses:    []string{"Link was cancelled (rzp.paymentLink.cancel) but URL was still shared with customer", "Customer bookmarked the link before it was cancelled", "Testing with an old cancelled link"},
		ResolutionSteps: []string{"Create a new payment link if payment is still needed", "When cancelling a link, notify the customer and provide a new link", "Check link status before sharing: GET /v1/payment_links/{id}"},
		DocURL:          "https://razorpay.com/docs/api/payment-links/",
		IsRetriable:     false,
		GuardrailRef:    "RZP-PL-007",
		GuardrailTitle:  "Cancelling a payment link does not refund payments already made on it",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "amount should be equal to the order amount",
		Title:           "Capture Amount Mismatch",
		Explanation:     "The amount you're passing to the capture API does not match the order amount. Razorpay requires the capture amount to equal the authorized amount.",
		CommonCauses:    []string{"Passing a partial amount to capture (partial capture is not supported for cards)", "Amount in rupees instead of paise passed to capture", "Discounted amount passed without creating a new order"},
		ResolutionSteps: []string{"Pass the exact authorized amount in paise: payment.amount", "For partial charges: do not capture, instead create a new order for the correct amount", "If the price changed, cancel the authorized payment and create a new order"},
		DocURL:          "https://razorpay.com/docs/api/payments/#capture-a-payment",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
	{
		Code:            "BAD_REQUEST_ERROR",
		SubDescription:  "customer phone number is not valid",
		Title:           "Invalid Customer Phone Number",
		Explanation:     "The customer contact number provided is not a valid Indian mobile number. Razorpay requires a valid 10-digit Indian mobile number with country code +91.",
		CommonCauses:    []string{"Phone number without country code (9876543210 instead of +919876543210)", "Using a landline number instead of mobile", "Non-Indian phone number format", "Including spaces or dashes in the number"},
		ResolutionSteps: []string{"Format as: +91XXXXXXXXXX (10 digits after +91, no spaces)", "Validate before API call: /^\\+91[6-9]\\d{9}$/.test(contact)", "Strip spaces and dashes: contact.replace(/[\\s-]/g, '')"},
		DocURL:          "https://razorpay.com/docs/api/customers/#create-a-customer",
		IsRetriable:     false,
		GuardrailRef:    "",
		GuardrailTitle:  "",
	},
}

// searchResultItem is a single documentation match returned by
// SearchDocumentation.
type searchResultItem struct {
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	Excerpt       string   `json:"excerpt"`
	Relevance     int      `json:"relevance"`
	GuardrailRefs []string `json:"guardrail_refs,omitempty"`
}

// codeExampleItem is a single runnable code example returned by
// SearchDocumentation.
type codeExampleItem struct {
	Title       string `json:"title"`
	Language    string `json:"language"`
	Code        string `json:"code"`
	Description string `json:"description"`
	DocURL      string `json:"doc_url"`
}

// searchDocumentationOutput is the response shape for SearchDocumentation.
type searchDocumentationOutput struct {
	Query        string             `json:"query,omitempty"`
	Results      []searchResultItem `json:"results"`
	CodeExamples []codeExampleItem  `json:"code_examples,omitempty"`
	Message      string             `json:"message,omitempty"`
}

// explainErrorOutput is the response shape for ExplainError.
type explainErrorOutput struct {
	ErrorCode        string   `json:"error_code"`
	ErrorDescription string   `json:"error_description,omitempty"`
	Title            string   `json:"title"`
	Explanation      string   `json:"explanation"`
	CommonCauses     []string `json:"common_causes"`
	ResolutionSteps  []string `json:"resolution_steps"`
	DocURL           string   `json:"doc_url"`
	IsRetriable      bool     `json:"is_retriable"`
	GuardrailRef     *string  `json:"guardrail_ref"`
	GuardrailTitle   *string  `json:"guardrail_title"`
}

// scoredDocSection pairs a docSection with its computed relevance score
// during SearchDocumentation ranking.
type scoredDocSection struct {
	section docSection
	score   int
}

// SearchDocumentation returns a tool that searches Razorpay documentation
// and returns the most relevant sections along with runnable code examples.
// This is a static-data helper tool: it makes no Razorpay API calls.
func SearchDocumentation(obs *observability.Observability) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"query",
			mcpgo.Description(
				"Natural language query. E.g. \"how to set up UPI AutoPay\" "+
					"or \"webhook signature verification\""),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"language",
			mcpgo.Description(
				"Programming language for code examples. E.g. \"node\", "+
					"\"python\", \"php\". Defaults to node."),
		),
		mcpgo.WithString(
			"topic",
			mcpgo.Description("Optional topic filter."),
			mcpgo.Enum(
				"payment-verification",
				"order-creation",
				"checkout-sdk",
				"api-keys",
				"upi-autopay",
				"nach",
				"upi-collect",
				"upi-intent",
				"route-api",
				"webhook-setup",
				"payment-link",
				"refunds",
				"subscriptions",
				"test-integration",
				"error-handling",
				"capture-payment",
				"first-integration",
			),
		),
	}

	handler := func(
		ctx context.Context, r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		args, ok := r.Arguments.(map[string]interface{})
		if !ok {
			return mcpgo.NewToolResultError("Invalid arguments"), nil
		}

		query, _ := args["query"].(string)
		if strings.TrimSpace(query) == "" {
			return mcpgo.NewToolResultError(
				"Invalid arguments: query is required"), nil
		}
		language, _ := args["language"].(string)
		topic, _ := args["topic"].(string)

		output := searchDocumentationForQuery(query, language, topic)

		return mcpgo.NewToolResultJSON(output)
	}

	return mcpgo.NewTool(
		"search_documentation",
		"Search Razorpay documentation and return relevant sections with "+
			"runnable code examples. Use this when a developer asks how to "+
			"implement a Razorpay feature, instead of relying on training "+
			"data. Returns the top 3 most relevant doc sections with code "+
			"examples.",
		parameters,
		handler,
	)
}

// searchDocumentationForQuery ports the scoring/ranking logic from the
// razorpay-mcp-tools Node.js prototype's searchDocumentation function.
func searchDocumentationForQuery(
	query string, language string, topic string,
) searchDocumentationOutput {
	q := strings.ToLower(query)

	var queryWords []string
	for _, w := range strings.Fields(q) {
		if len(w) > 2 {
			queryWords = append(queryWords, w)
		}
	}

	scored := make([]scoredDocSection, 0, len(docSections))
	for _, section := range docSections {
		score := 0
		if topic != "" && section.Topic == topic {
			score += 10
		}
		for _, word := range queryWords {
			if strings.Contains(section.Topic, word) {
				score += 3
			}
			if strings.Contains(strings.ToLower(section.Title), word) {
				score += 2
			}
			if strings.Contains(strings.ToLower(section.Summary), word) {
				score++
			}
			for _, kw := range section.Keywords {
				if strings.Contains(strings.ToLower(kw), word) {
					score += 2
					break
				}
			}
		}
		for _, kw := range section.Keywords {
			if strings.Contains(q, strings.ToLower(kw)) {
				score += 3
			}
		}
		if score > 0 {
			scored = append(scored, scoredDocSection{section: section, score: score})
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	if len(scored) > 3 {
		scored = scored[:3]
	}

	if len(scored) == 0 {
		return searchDocumentationOutput{
			Results: []searchResultItem{},
			Message: "No documentation found for \"" + query + "\". Try " +
				"keywords like: webhook, order, checkout, upi, subscription, " +
				"refund, route, api-key",
		}
	}

	var langNote string
	if language != "" && language != "node" {
		langNote = "\n// Note: Example shown in Node.js. For " + language +
			", see: " + scored[0].section.DocURL
	}

	effectiveLanguage := language
	if effectiveLanguage == "" {
		effectiveLanguage = "node"
	}

	results := make([]searchResultItem, 0, len(scored))
	codeExamples := make([]codeExampleItem, 0, len(scored))
	for _, s := range scored {
		relevance := 0
		for _, kw := range s.section.Keywords {
			if strings.Contains(q, strings.ToLower(kw)) {
				relevance++
			}
		}
		results = append(results, searchResultItem{
			Title:         s.section.Title,
			URL:           s.section.DocURL,
			Excerpt:       s.section.Summary,
			Relevance:     relevance,
			GuardrailRefs: s.section.GuardrailRefs,
		})
		codeExamples = append(codeExamples, codeExampleItem{
			Title:       s.section.Title,
			Language:    effectiveLanguage,
			Code:        s.section.CodeExample + langNote,
			Description: s.section.Summary,
			DocURL:      s.section.DocURL,
		})
	}

	return searchDocumentationOutput{
		Query:        query,
		Results:      results,
		CodeExamples: codeExamples,
	}
}

// ExplainError returns a tool that takes a Razorpay error code and optional
// description and returns a plain-English explanation with common causes
// and step-by-step resolution. This is a static-data helper tool: it makes
// no Razorpay API calls.
func ExplainError(obs *observability.Observability) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"error_code",
			mcpgo.Description(
				"Razorpay error code. E.g. \"BAD_REQUEST_ERROR\", "+
					"\"GATEWAY_ERROR\", \"SERVER_ERROR\""),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"error_description",
			mcpgo.Description(
				"The error description or message. E.g. \"order_id is "+
					"required\" or \"Payment signature verification failed\""),
		),
		mcpgo.WithString(
			"sub_description",
			mcpgo.Description(
				"The sub-description or reason field from the error "+
					"response. Same as error_description if only one is "+
					"available."),
		),
	}

	handler := func(
		ctx context.Context, r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		args, ok := r.Arguments.(map[string]interface{})
		if !ok {
			return mcpgo.NewToolResultError("Invalid arguments"), nil
		}

		errorCode, _ := args["error_code"].(string)
		if strings.TrimSpace(errorCode) == "" {
			return mcpgo.NewToolResultError(
				"Invalid arguments: error_code is required"), nil
		}
		errorDescription, _ := args["error_description"].(string)
		subDescription, _ := args["sub_description"].(string)

		output := explainErrorForCode(errorCode, errorDescription, subDescription)

		return mcpgo.NewToolResultJSON(output)
	}

	return mcpgo.NewTool(
		"explain_error",
		"Takes a Razorpay error code and optional description, returns a "+
			"plain-English explanation with common causes and step-by-step "+
			"resolution. Use this when a developer pastes a Razorpay error "+
			"instead of giving a generic response.",
		parameters,
		handler,
	)
}

// explainErrorForCode ports the fallback lookup logic from the
// razorpay-mcp-tools Node.js prototype's explainError function.
func explainErrorForCode(
	errorCode string, errorDescription string, subDescription string,
) explainErrorOutput {
	descToMatch := strings.ToLower(subDescription)
	if descToMatch == "" {
		descToMatch = strings.ToLower(errorDescription)
	}

	var match *errorEntry

	for i := range errorRegistryData {
		e := &errorRegistryData[i]
		if e.Code != errorCode {
			continue
		}
		subLower := strings.ToLower(e.SubDescription)
		if descToMatch == "" ||
			strings.Contains(subLower, descToMatch) ||
			strings.Contains(descToMatch, subLower) {
			match = e
			break
		}
	}

	if match == nil && errorCode != "" {
		for i := range errorRegistryData {
			e := &errorRegistryData[i]
			if e.Code == errorCode {
				match = e
				break
			}
		}
	}

	if match == nil && descToMatch != "" {
		var words []string
		for _, w := range strings.Fields(descToMatch) {
			if len(w) > 3 {
				words = append(words, w)
			}
		}
	outer:
		for i := range errorRegistryData {
			e := &errorRegistryData[i]
			for _, w := range words {
				if strings.Contains(strings.ToLower(e.SubDescription), w) ||
					strings.Contains(strings.ToLower(e.Explanation), w) ||
					strings.Contains(strings.ToLower(e.Title), w) {
					match = e
					break outer
				}
			}
		}
	}

	if match == nil {
		fallbackDesc := errorDescription
		if fallbackDesc == "" {
			fallbackDesc = subDescription
		}
		return explainErrorOutput{
			ErrorCode:        errorCode,
			ErrorDescription: fallbackDesc,
			Title:            "Error Not Found in Registry",
			Explanation: "The error code \"" + errorCode + "\" with " +
				"description \"" + fallbackDesc + "\" is not in the local " +
				"registry. Check the Razorpay API error docs for details.",
			CommonCauses: []string{
				"Check the error description for hints",
				"Verify request parameters against API docs",
			},
			ResolutionSteps: []string{
				"Review the API documentation for this endpoint",
				"Check the Razorpay support portal",
			},
			DocURL:         "https://razorpay.com/docs/api/errors/",
			IsRetriable:    false,
			GuardrailRef:   nil,
			GuardrailTitle: nil,
		}
	}

	var guardrailRef, guardrailTitle *string
	if match.GuardrailRef != "" {
		guardrailRef = &match.GuardrailRef
	}
	if match.GuardrailTitle != "" {
		guardrailTitle = &match.GuardrailTitle
	}

	return explainErrorOutput{
		ErrorCode:        match.Code,
		ErrorDescription: match.SubDescription,
		Title:            match.Title,
		Explanation:      match.Explanation,
		CommonCauses:     match.CommonCauses,
		ResolutionSteps:  match.ResolutionSteps,
		DocURL:           match.DocURL,
		IsRetriable:      match.IsRetriable,
		GuardrailRef:     guardrailRef,
		GuardrailTitle:   guardrailTitle,
	}
}
