# Refund

A Refund is a reversal of a captured payment, either full or partial. It is always scoped to exactly one payment and cannot exist independently.

---

## Decisions

### D1: Two separate fetch tools for payment-scoped refunds

**Context:** A payment can have multiple refunds. Callers need either the full list or one specific refund.
**Decision:** Expose `fetch_multiple_refunds_for_payment` (list all refunds for a payment) and `fetch_specific_refund_for_payment` (one refund by refund_id within a payment) as distinct tools.
**Alternatives considered:**
- **Single tool with optional refund_id:** Collapsed into one call — rejected because the two use cases have different required parameters and different Razorpay API endpoints (`Payment.FetchMultipleRefund` vs `Payment.FetchRefund`).
**Trade-offs:**
- Gained: Clearer intent per tool; no ambiguous optional-parameter behaviour.
- Lost: Slightly more surface area for callers to learn.
**Code:** `refunds.go:FetchMultipleRefundsForPayment()`, `refunds.go:FetchSpecificRefundForPayment()`

### D2: Amount in paise, minimum 100

**Context:** Razorpay's API operates in the smallest currency unit (paise for INR). Keeping the same unit across payments and refunds avoids conversion bugs.
**Decision:** Refund amount is always in paise; minimum enforced at 100 (₹1).
**Code:** `refunds.go:CreateRefund()`
**Revisit if:** Razorpay adds multi-currency support with different smallest units.

### D3: Default page size of 10 is a Razorpay API default, not an MCP choice

**Context:** Both `fetch_all_refunds` and `fetch_multiple_refunds_for_payment` say "last 10 returned by default."
**Decision:** No override applied — the MCP server passes through whatever the Razorpay SDK returns; the default comes from the upstream API.
**Code:** `refunds.go:FetchAllRefunds()`, `refunds.go:FetchMultipleRefundsForPayment()`

---

## Non-Obvious Constraints

- **Partial vs full refund:** `amount` is required in the current tool definition (min 100 paise). A full refund requires passing the full original payment amount — the API does not infer "full refund" from an omitted amount in this MCP tool.
- **Update is notes-only:** `update_refund` accepts only `notes`; no other field (amount, speed, receipt) can be changed after creation. Attempting to pass other fields will not error but will be silently ignored by the upstream API.
- **fetch_specific_refund_for_payment requires both IDs:** `payment_id` + `refund_id` are both required. Using only `refund_id` on `client.Payment.FetchRefund` would fail; use `fetch_refund` (global) if only `refund_id` is known.
- **Max 100 per page:** Both `fetch_all_refunds` and `fetch_multiple_refunds_for_payment` cap at 100 records per call; use `skip`/`count` for pagination.
- **ID prefixes are enforced by convention, not validated in MCP code:** `rfnd_` for refund IDs, `pay_` for payment IDs. The Razorpay API will reject mismatched IDs.
- **Speed field:** Only `normal` and `optimum` are valid values; `optimum` triggers instant refund logic on Razorpay's side.

---

## Flow Map

### Create Refund

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Partial refund (most traffic) | Caller provides payment_id + amount < original | `refunds.go:CreateRefund()` -> `client.Payment.Refund()` | DP1: amount >= 100? | Refund object returned |
| Full refund | Caller provides payment_id + full original amount | `refunds.go:CreateRefund()` -> `client.Payment.Refund()` | DP1 | Refund object returned |
| Speed: optimum | Caller sets speed=optimum | `refunds.go:CreateRefund()` -> `client.Payment.Refund()` | DP2: speed value passed to API | Instant refund initiated |

**Decision Points:**
- **DP1: Amount validation** — MCP enforces min 100 via `mcpgo.Min(100)`; upper bound (must not exceed payment amount) enforced by Razorpay API, not MCP.
- **DP2: Speed field** — Optional; absence defaults to `normal` on the Razorpay side.

### Fetch Refunds for a Payment

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| List all refunds for payment (common) | Caller has payment_id, wants all refunds | `refunds.go:FetchMultipleRefundsForPayment()` -> `client.Payment.FetchMultipleRefund()` | DP1: pagination params | Paginated list (default last 10) |
| Fetch one specific refund (common) | Caller has both payment_id and refund_id | `refunds.go:FetchSpecificRefundForPayment()` -> `client.Payment.FetchRefund()` | DP2: both IDs required | Single refund object |
| Fetch refund globally (rare) | Caller has only refund_id, no payment_id | `refunds.go:FetchRefund()` -> `client.Refund.Fetch()` | — | Single refund object |

**Decision Points:**
- **DP1: Pagination** — `count` and `skip` are optional; without them, API returns last 10.
- **DP2: Both IDs required** — The payment-scoped endpoint (`Payment.FetchRefund`) requires both; use the global `FetchRefund` tool if payment_id is unknown.

---

## Service Contracts

**ID formats:**
- Refund ID: `rfnd_xxx` — used in `fetch_refund`, `update_refund`, `fetch_specific_refund_for_payment`
- Payment ID: `pay_xxx` — used in `create_refund`, `fetch_multiple_refunds_for_payment`, `fetch_specific_refund_for_payment`

**Amount:** Always in paise (integer). Minimum 100.

**Upstream dependency (Razorpay API):** All tools delegate to `rzpsdk.Client`. If the Razorpay API is down, all refund tools return a tool error via `refunds.go:formatErrorMessage()`. No local state is maintained; there is no retry logic in the MCP layer.

**What callers must know:** A refund belongs to exactly one payment. The relationship is permanent — a refund cannot be reassigned to a different payment after creation.
