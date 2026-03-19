# Payments Domain

A payment in Razorpay represents money movement from customer to merchant. It passes through `created` → `authorized` → `captured` (or `failed`/`refunded`). This MCP server exposes the full payment lifecycle plus token management as a single toolset.

---

## Decisions

### D1: Token tools live in the payments toolset, not a separate tokens toolset

**Context:** Tokens (saved payment methods) only have business meaning in the context of initiating a payment — a token is the mechanism for charging a returning customer without re-collecting card details.
**Decision:** `fetch_tokens` and `revoke_token` are registered in the payments toolset alongside `initiate_payment`.
**Alternatives considered:**
- **Separate tokens toolset:** Clean separation by resource type — rejected because agents working on payment flows would need to load two toolsets; the coupling is intentional.
**Trade-offs:**
- Gained: Single toolset covers the complete "returning customer payment" flow.
- Lost: Tokens are not independently browsable without loading the payments toolset.
**Code:** `tokens.go:FetchSavedPaymentMethods()`, `tokens.go:RevokeToken()`
**Revisit if:** Token management grows to cover non-payment use cases (e.g., mandate-only tokens).

### D2: `initiate_payment` uses S2S JSON v1 flow

**Context:** Razorpay offers redirect-based checkout (customer leaves merchant site) and S2S (server-to-server) flows. Agents cannot render browser redirects or host a checkout page.
**Decision:** All payment initiation uses `client.Payment.CreatePaymentJson()` — a direct API call that returns the payment object and a `next` array of actions. No browser redirect is issued.
**Alternatives considered:**
- **Standard checkout:** Requires a browser session — agents cannot participate.
- **Redirect flow:** Returns a URL for the customer to complete payment; unsuitable when the caller is an LLM agent.
**Trade-offs:**
- Gained: Fully programmable; agent controls the entire flow including OTP submission.
- Lost: Some payment methods that only support redirect (e.g., certain netbanking banks) are not supported via this path.
**Code:** `payments.go:createPaymentWithParams()`, `payments.go:InitiatePayment()`
**Revisit if:** Razorpay introduces a new S2S v2 API.

### D3: Amounts are in paise (smallest currency unit)

**Context:** Razorpay's API natively uses the smallest currency sub-unit (paise for INR, cents for USD) to avoid floating-point precision issues. This matches the Razorpay SDK contract.
**Decision:** All amount parameters are integers in paise. INR 100 paise = ₹1. Minimum enforced at 100 paise (₹1).
**Alternatives considered:**
- **Rupees with decimal:** Would require float handling; API rejects it.
**Trade-offs:**
- Gained: No rounding errors; matches Razorpay API directly.
- Lost: Non-obvious to developers unfamiliar with Indian payment APIs (100 for ₹1 looks like ₹100).
**Code:** `payments.go:CapturePayment()`, `payments.go:InitiatePayment()`
**Revisit if:** Multi-currency support requires per-currency sub-unit awareness.

---

## Non-Obvious Constraints

- **`capture_payment` only works on `authorized` payments.** Razorpay's two-step auth model: gateway authorizes (holds funds) but does not settle until merchant explicitly captures. Attempting capture on `captured`, `failed`, or `created` status returns an API error. The capture amount must equal the authorized amount.
- **`fetch_tokens` looks up by phone number, not customer_id.** The tool internally calls `Customer.Create` with `fail_existing=0` to resolve the phone to a customer, then fetches that customer's tokens. Callers do not need to know the `cust_` ID upfront. `tokens.go:FetchSavedPaymentMethods()`
- **`revoke_token` is irreversible.** The `/cancel` endpoint permanently invalidates the token. There is no restore path. `tokens.go:RevokeToken()`
- **`fetch_payment_card_details` only works for card-method payments.** The Razorpay API returns an error for UPI, netbanking, and wallet payments. `payments.go:FetchPaymentCardDetails()`
- **OTP flow is strictly sequential.** You must call `initiate_payment` first to obtain a `payment_id`. Only then can `resend_otp` or `submit_otp` be called. There is no OTP without a payment in flight.
- **Auto-generated email when contact is provided but email is omitted.** `payments.go:addContactAndEmailToPaymentData()` synthesizes `{contact}@mcp.razorpay.com`. Razorpay's API requires an email; this fallback prevents errors but produces a non-real address.
- **Non-INR recurring payments use a different SDK method.** `createPaymentWithParams()` switches to `client.Payment.CreateRecurringPayment()` when `recurring=true` + `token` is set + `currency != INR`. INR recurring payments still use `CreatePaymentJson`.

---

## Flow Map

### Payment Lifecycle (Authorize → Capture)

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Happy path (most traffic) | Agent calls `initiate_payment` then `capture_payment` | `payments.go:InitiatePayment()` -> `payments.go:CapturePayment()` | DP1: status must be `authorized` before capture | Payment captured, funds settled |
| OTP required (common for saved cards) | `initiate_payment` returns `otp_generate` action | `payments.go:processPaymentResult()` -> `payments.go:sendOtp()` -> `payments.go:SubmitOtp()` | DP2: `next` array contains `otp_generate` | OTP sent automatically; agent calls `submit_otp` to complete |
| OTP resend (occasional) | Customer did not receive OTP | `payments.go:ResendOtp()` | — | New OTP sent; agent re-calls `submit_otp` |
| Payment fails (rare) | Gateway decline | `payments.go:InitiatePayment()` | — | Error returned; no payment_id to capture |

**Decision Points:**
- **DP1: Authorized-only capture** — Razorpay enforces two-step settlement; capture on non-authorized payments returns a 400 error from the API.
- **DP2: OTP auto-trigger** — When `initiate_payment` detects `otp_generate` in the `next` array, `sendOtp()` is called automatically (server-to-server POST to the Razorpay OTP URL). The agent does not need to trigger OTP separately; it only needs to collect the OTP from the user and call `submit_otp`.

### S2S Initiate Sub-flows

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Saved payment method (common) | `token` param provided | `payments.go:buildPaymentData()` -> `payments.go:createPaymentWithParams()` | DP3: customer resolved from contact or customer_id | Payment created using stored card/UPI |
| UPI collect (common) | `vpa` param provided | `payments.go:processUPIParameters()` | DP4: sets method=upi, flow=collect, expiry=6min | Collect request sent to customer's UPI app |
| UPI intent (less common) | `upi_intent=true` | `payments.go:processUPIParameters()` | DP4: sets method=upi, flow=intent | API returns UPI intent URL for deep-link |

**Decision Points:**
- **DP3: Customer resolution** — If `customer_id` is absent, the tool calls `Customer.Create` with `fail_existing=0` to find or create a customer from the `contact` number. This avoids requiring callers to pre-fetch customer IDs.
- **DP4: UPI param auto-expansion** — `vpa` and `upi_intent` are convenience parameters. The tool expands them into the nested `upi` object that the Razorpay API requires, so callers do not need to know the API's UPI object structure.

### Token Management Flow

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| List tokens (common) | Agent calls `fetch_tokens` with phone | `tokens.go:FetchSavedPaymentMethods()` -> `Customer.Create` -> `GET /v1/customers/{id}/tokens` | DP5: phone resolves to customer | Returns customer + all saved payment methods |
| Revoke token (occasional) | Agent calls `revoke_token` | `tokens.go:RevokeToken()` -> `PUT /v1/customers/{id}/tokens/{token_id}/cancel` | — | Token permanently cancelled |

**Decision Points:**
- **DP5: Phone-to-customer resolution** — `fetch_tokens` always resolves the phone number first. This is necessary because the Razorpay token API is scoped to `customer_id`, not phone number, but agents and end-users think in terms of phone numbers.

---

## Service Contracts

**Input contracts:**
- `payment_id`: string prefixed `pay_` (Razorpay-assigned)
- `customer_id`: string prefixed `cust_`
- `token_id` / `token`: string prefixed `token_`
- `order_id`: string prefixed `order_`
- `amount`: integer in paise, minimum 100
- `currency`: ISO 4217 code (default `INR`)
- `contact`: E.164 or local format phone number

**Output contract:** All tools return JSON serialized from the Razorpay SDK response, wrapped as MCP text content via `mcpgo.NewToolResultJSON()`.

**Error format:** `"<action> failed: <api error message>"` — e.g., `"fetching payment failed: payment not found"`. Empty API errors become `"<action> failed: resource does not exist"` via `tools_params.go:formatErrorMessage()`.

**What breaks if Razorpay API is down:** All tools return tool-level errors (not MCP protocol errors). The MCP session continues; individual tool calls fail gracefully.
