# QR Codes

A scannable UPI payment code merchants display to accept payments. QR codes are created with a defined usage policy (single or multiple payments) and optionally linked to a customer or payment at creation time. Once closed, a QR code cannot accept further payments.

---

## Decisions

### D1: UPI-only QR type

**Context:** Razorpay supports multiple payment methods (card, netbanking, UPI). A QR code needs a type parameter.
**Decision:** Only `upi_qr` is supported — hardcoded as the sole allowed value via `mcpgo.Pattern("^upi_qr$")`.
**Alternatives considered:**
- **Generic QR:** Would require out-of-band method selection by the payer — UPI's collect flow makes the QR self-contained.
**Trade-offs:**
- Gained: Simplicity; UPI collect flow handles payer method selection automatically.
- Lost: Card/netbanking QR — those require redirect flows incompatible with static QR.
**Code:** `qr_codes.go:CreateQRCode()`
**Revisit if:** Razorpay introduces card-on-QR or netbanking-on-QR collect flows.

### D2: `close_qr_code` is a destructive (write) tool

**Context:** MCP tools are classified as read or write to gate agent autonomy.
**Decision:** `CloseQRCode()` is flagged as a write/destructive tool because closing a QR code is irreversible — the Razorpay API does not provide a reopen operation.
**Alternatives considered:**
- **Read-only classification:** Would allow agents to call it without user approval — dangerous given irreversibility.
**Trade-offs:**
- Gained: Agents must confirm before closing; prevents accidental payment disruption.
- Lost: Convenience for automated cleanup flows.
**Code:** `qr_codes.go:CloseQRCode()`
**Revisit if:** Razorpay API adds a reopen endpoint.

### D3: Separate fetch variants by customer_id and payment_id

**Context:** QR codes can be associated with a customer or with a specific payment at creation time; the association is immutable after creation.
**Decision:** Two distinct tools — `FetchQRCodesByCustomerID()` and `FetchQRCodesByPaymentID()` — both proxy to `client.QrCode.All()` but enforce the required filter type at the MCP layer.
**Alternatives considered:**
- **Single fetch tool with optional filters:** Would allow agents to omit both filters and return unrelated QR codes, creating ambiguity.
**Trade-offs:**
- Gained: Clear intent; agents cannot accidentally omit a required scoping parameter.
- Lost: Minor code duplication.
**Code:** `qr_codes.go:FetchQRCodesByCustomerID()`, `qr_codes.go:FetchQRCodesByPaymentID()`
**Revisit if:** Razorpay API adds a dedicated by-customer or by-payment endpoint distinct from the general list.

---

## Non-Obvious Constraints

| Rule | Why | Enforced at |
|------|-----|-------------|
| `type` must be `upi_qr` | Only UPI collect flow is QR-compatible on Razorpay | `CreateQRCode()` — regex pattern |
| `usage` is required (`single_use` or `multiple_use`) | Determines whether QR auto-closes after first payment | `CreateQRCode()` — required field |
| `payment_amount` required when `fixed_amount=true` | A fixed-amount QR with no amount is invalid | `CreateQRCode()` — runtime check |
| `close_by` must be at least 2 minutes in the future | Razorpay API rejects timestamps too close to now | Razorpay API (not enforced locally) |
| Closed QR codes cannot be reopened | Irreversible API operation; no reopen endpoint exists | Document at call site; warn agents |
| `qr_code_id` must start with `qr_` | Razorpay ID format; wrong prefix causes API 400 | Convention — not validated locally |
| `customer_id` must start with `cust_` | Razorpay customer ID format | Convention — not validated locally |
| `payment_id` must start with `pay_` | Razorpay payment ID format | Convention — not validated locally |

---

## Flow Map

### Create QR and Accept Payments

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Single-use QR (most traffic) | Agent calls create with `usage=single_use` | `qr_codes.go:CreateQRCode()` -> `client.QrCode.Create()` | DP1: fixed_amount check | QR created; auto-closes after first payment |
| Multi-use QR | Agent calls create with `usage=multiple_use` | `qr_codes.go:CreateQRCode()` -> `client.QrCode.Create()` | DP1: fixed_amount check | QR remains open until explicitly closed or `close_by` reached |
| Fetch payments for reconciliation (common) | Agent calls fetch after payments expected | `qr_codes.go:FetchPaymentsForQRCode()` -> `client.QrCode.FetchPayments()` | — | Returns paginated payment list for the QR |

### Close QR Code

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Explicit close (destructive) | Agent calls close_qr_code | `qr_codes.go:CloseQRCode()` -> `client.QrCode.Close()` | IRREVERSIBLE — confirm before calling | QR permanently closed; no further payments accepted |

**Decision Points:**
- **DP1: fixed_amount + payment_amount coupling** — If `fixed_amount=true`, `payment_amount` must also be provided. The code enforces this after general validation, not at the parameter schema level, so agents will get a runtime error rather than a schema error if they omit `payment_amount`.

---

## Service Contracts

| ID type | Format | Used by |
|---------|--------|---------|
| QR code ID | `qr_xxx` | `FetchQRCode()`, `FetchPaymentsForQRCode()`, `CloseQRCode()` |
| Customer ID | `cust_xxx` | `FetchQRCodesByCustomerID()`, `CreateQRCode()` (optional link) |
| Payment ID | `pay_xxx` | `FetchQRCodesByPaymentID()` |

**Upstream dependency:** All tools call the Razorpay API via `rzpsdk.Client.QrCode.*`. If the Razorpay API is unavailable, all QR operations fail — there is no local fallback or caching.
