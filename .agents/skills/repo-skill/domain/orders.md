# Orders Domain

An Order is Razorpay's pre-payment intent object. A customer cannot make a payment without a corresponding order — the order_id is the anchor linking payment attempts back to the merchant's intent.

---

## Decisions

### D1: Single `create_order` tool for both regular and mandate orders

**Context:** Razorpay's API uses a single `/orders` endpoint for both one-time and recurring (mandate) orders, but mandate orders require an entirely different parameter set.
**Decision:** Expose one `create_order` tool with optional mandate-specific fields rather than two separate tools.
**Alternatives considered:**
- **Separate `create_mandate_order` tool:** Cleaner parameter contract per flow — rejected because it duplicates the underlying API call and forces callers to choose before understanding their context.
**Trade-offs:**
- Gained: One tool for all order creation; mirrors the Razorpay API shape.
- Lost: Tool description must carry significant disambiguation burden; easy to call with missing mandate fields.
**Code:** `orders.go:CreateOrder()`
**Revisit if:** Mandate order flows become complex enough that separate tooling improves agent success rates.

---

## Non-Obvious Constraints

- **Mandate order required fields** — `method`, `customer_id`, and `token` are optional in the schema but are all required when creating a mandate (recurring) order. The only currently supported token type is `single_block_multiple_debit`, and when that type is used, `method` MUST be `upi`. Enforcement is at the Razorpay API layer, not in this server.
  Code: `orders.go:CreateOrder()`

- **`update_order` is notes-only** — Razorpay's API does not allow modifying amount, currency, receipt, or status after creation. The `UpdateOrder` handler explicitly extracts only `notes` from the validated payload and discards everything else.
  Code: `orders.go:UpdateOrder()`

- **`first_payment_min_amount` is conditional** — This field is only forwarded to the Razorpay API when `partial_payment` is `true`. Passing it on a full-payment order is silently ignored by the handler.
  Code: `orders.go:CreateOrder()`

- **`fetch_order_payments` returns all payment attempts** — Including failed, cancelled, and pending payments, not just successful captures. Callers reconciling revenue must filter by payment status.
  Code: `orders.go:FetchOrderPayments()`

- **Order ID format** — Must be prefixed `order_` (e.g., `order_xxx`). Validated by description convention, not by a regex guard in this server; malformed IDs will error at the Razorpay API layer.

- **Notes constraints** — Max 15 key-value pairs; each value max 256 characters. Enforced via `mcpgo.MaxProperties(15)` on the schema, not at runtime in the handler.

---

## Flow Map

### Create Order and Pay

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Regular order (most traffic) | Agent calls `create_order` with amount + currency | `orders.go:CreateOrder()` -> `client.Order.Create()` | DP1: partial payment flag | Order created; order_id returned for payment step |
| Mandate order (recurring) | Agent calls `create_order` with method + customer_id + token | `orders.go:CreateOrder()` -> `client.Order.Create()` | DP2: token type must match method | Mandate order created; customer linked for recurring debits |
| Fetch payments for order | Agent calls `fetch_order_payments` after payment attempt | `orders.go:FetchOrderPayments()` -> `client.Order.Payments()` | None | All payment attempts returned, including failed ones |

**Decision Points:**
- **DP1: partial_payment flag** — When `true`, `first_payment_min_amount` is conditionally forwarded. Why: Razorpay rejects `first_payment_min_amount` on non-partial orders.
- **DP2: token.type vs method coupling** — `single_block_multiple_debit` requires `method=upi`. Why: UPI is the only Razorpay-supported channel for this mandate type.

---

## Service Contracts

**order_id format:** `order_xxx` (Razorpay-assigned, prefix `order_`)
**Amount:** Integer, in paise (smallest currency sub-unit); minimum 100 (= 1.00 INR)
**Currency:** ISO 4217 three-letter uppercase code (e.g., `INR`, `USD`); validated by pattern `^[A-Z]{3}$`
**Notes:** Key-value map, max 15 keys, values max 256 chars
**Receipt:** Merchant-assigned, max 40 chars, must be unique per merchant account
**Token expire_at:** Unix timestamp; defaults to today + 60 days if omitted

**What breaks if Razorpay API is down:** All order tools return errors — there is no local state or caching. Order creation is a hard dependency on the upstream API.
