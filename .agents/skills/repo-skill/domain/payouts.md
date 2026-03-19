# Payouts Domain

Payouts represent outbound money transfers from a merchant's bank account to any bank account (employees, vendors, partners). Unlike refunds — which reverse a payment back to the original payer — payouts are initiated independently of any prior payment transaction.

---

## Decisions

### D1: Read-Only Toolset (No Create/Cancel)

**Context:** Payouts API requires two safeguards absent from the standard API: a mandatory `X-Payout-Idempotency` header (caller-generated unique key) and, in many configurations, a separate API key scoped to the RazorpayX banking product. Omitting the idempotency header can cause duplicate fund transfers.

**Decision:** Expose only `fetch_payout_with_id` and `fetch_all_payouts`. No create, update, or cancel tools.

**Alternatives considered:**
- **Expose create with idempotency key param:** Rejected — an agent generating idempotency keys in an agentic loop risks retrying with the same key or generating collisions, leading to duplicate or blocked payouts.
- **Separate RazorpayX client:** Rejected for now — adds credential management complexity with no read-side benefit.

**Trade-offs:**
- Gained: no risk of unintended fund transfers from agentic workflows.
- Lost: agents cannot initiate disbursements autonomously.

**Code:** `payouts.go:FetchPayout()`, `payouts.go:FetchAllPayouts()`

**Revisit if:** A safe idempotency-key generation strategy is standardized across the MCP server, or a dedicated RazorpayX MCP client is introduced.

---

## Non-Obvious Constraints

| Rule | Why | Enforced at |
|------|-----|-------------|
| `account_number` is required for listing | Payouts are scoped to a source bank account; there is no global payout namespace in the API | `payouts.go:FetchAllPayouts()` — `ValidateAndAddRequiredString` |
| Payout ID format is `pout_xxx` | Razorpay-assigned prefix; differs from payment (`pay_`) and refund (`rfnd_`) IDs | Example in tool description: `pout_00000000000001` |
| Pagination uses `count`/`skip`, not cursor | SDK's `Payout.All()` uses offset-style pagination; no cursor token is exposed | `payouts.go:FetchAllPayouts()` — `ValidateAndAddPagination` |

---

## Flow Map

### Fetch Payout by ID

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Happy path (most traffic) | Agent provides `payout_id` | `payouts.go:FetchPayout()` -> SDK `Payout.Fetch()` | DP1: validate `payout_id` present | Payout object returned as JSON |
| Missing ID (common mistake) | Agent omits `payout_id` | `payouts.go:FetchPayout()` -> `validator.HandleErrorsIfAny()` | DP1: required field absent | Tool error returned, no API call |

**Decision Points:**
- **DP1: payout_id validation** — Required string check before any network call — prevents SDK panic on empty ID string.

### List Payouts for Account

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Happy path (most traffic) | Agent provides `account_number` | `payouts.go:FetchAllPayouts()` -> SDK `Payout.All()` | DP1: validate account_number | Paginated payout list returned |
| Missing account_number (common mistake) | Agent omits `account_number` | `payouts.go:FetchAllPayouts()` -> `validator.HandleErrorsIfAny()` | DP1: required field absent | Tool error, no API call |

**Decision Points:**
- **DP1: account_number required** — The Razorpay Payouts API mandates this filter; listing across all accounts is not supported by the upstream API.

---

## Service Contracts

**Payout ID format:** `pout_xxx` (e.g., `pout_00000000000001`)

**account_number:** Required for `fetch_all_payouts`; identifies the RazorpayX source bank account, not a payment instrument.

**Downstream:** Razorpay SDK `client.Payout.Fetch()` and `client.Payout.All()` — targets the RazorpayX banking API. If RazorpayX is down, both tools fail; no local fallback.

**Gap for human verification:** Confirm whether the production RazorpayX API uses a distinct base URL or API key from the standard Razorpay API in the SDK configuration used here.
