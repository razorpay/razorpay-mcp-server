# Payment Links

A Payment Link is a shareable URL that lets merchants collect payments without building a checkout UI. Links can target all payment methods (standard) or UPI-only. Both types share a `plink_xxx` ID prefix and the same Razorpay API resource, but they behave differently at the gateway layer.

---

## Decisions

### D1: Two separate create tools for standard vs UPI links

**Context:** UPI links accept only INR and cannot use `accept_partial`. Standard links support multiple currencies and partial payments.
**Decision:** Expose `create_payment_link` and `payment_link_upi_create` as distinct tools rather than a single tool with a flag.
**Alternatives considered:**
- **Single create tool with `upi_link` boolean flag:** Rejected — callers would need to know which parameters are invalid for UPI, leading to silent failures or confusing error messages.
**Trade-offs:**
- Gained: Each tool has a self-consistent parameter set with no invalid combinations.
- Lost: Some parameter duplication between the two handlers.
**Code:** `payment_links.go:CreatePaymentLink()`, `payment_links.go:CreateUpiPaymentLink()`
**Revisit if:** Razorpay SDK introduces a first-class UPI link type that validates parameters server-side.

### D2: `payment_link_notify` is a write tool despite being a notification

**Context:** MCP tools are classified read or write based on whether they trigger external side effects.
**Decision:** Mark as write — calling `NotifyBy` dispatches an SMS or email immediately; it is not idempotent and costs the merchant a notification credit.
**Code:** `payment_links.go:ResendPaymentLinkNotification()`
**Revisit if:** Razorpay adds a dry-run or preview mode for notifications.

---

## Non-Obvious Constraints

| Constraint | Rule | Why | Enforced at |
|-----------|------|-----|-------------|
| Amount in paise | Integer, minimum 100 | Razorpay API rejects sub-paisa amounts; 100 paise = ₹1 | `payment_links.go:CreatePaymentLink()` via `mcpgo.Min(100)` |
| UPI links: INR only | `currency` must be `INR` | UPI rails are India-only; non-INR rejected by gateway | Razorpay API (not enforced client-side) |
| UPI links: no partial payments | `accept_partial` is ignored/rejected for UPI links | UPI does not support partial-collect flows | Razorpay API; tool description warns callers |
| Notification medium enum | `medium` must be `sms` or `email` | Razorpay `NotifyBy` API accepts only these two channels | `payment_links.go:ResendPaymentLinkNotification()` via `mcpgo.Enum("sms","email")` |
| Update requires at least one field | Request is rejected if `plUpdateReq` is empty | Sending an empty PATCH is a no-op that wastes an API call | `payment_links.go:UpdatePaymentLink()` explicit guard |
| Updatable fields only | Only `reference_id`, `expire_by`, `reminder_enable`, `accept_partial`, `notes` can be patched | Razorpay API does not allow mutating amount or customer after creation | `payment_links.go:UpdatePaymentLink()` parameter list |
| `fetch_all_payment_links` filters are additive | `payment_id` and `reference_id` are both passed to the API if provided | The Razorpay API applies them as AND filters; the MCP layer does not enforce mutual exclusion — use one or the other to avoid empty results | `payment_links.go:FetchAllPaymentLinks()` |

---

## Flow Map

### Create → Notify → Fetch

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Standard link creation (most traffic) | Agent calls `create_payment_link` | `payment_links.go:CreatePaymentLink()` -> `client.PaymentLink.Create()` | DP1: UPI or standard? | Link with `plink_xxx` ID returned |
| UPI link creation (common) | Agent calls `payment_link_upi_create` | `payment_links.go:CreateUpiPaymentLink()` -> injects `upi_link=true` -> `client.PaymentLink.Create()` | DP1 | UPI-restricted link returned |
| Notify after creation (common) | Agent calls `payment_link_notify` | `payment_links.go:ResendPaymentLinkNotification()` -> `client.PaymentLink.NotifyBy()` | DP2: medium sms or email | Notification dispatched; side effect is irreversible |
| Fetch single link (most traffic) | Agent calls `fetch_payment_link` | `payment_links.go:FetchPaymentLink()` -> `client.PaymentLink.Fetch()` | — | Full link object returned |

**Decision Points:**
- **DP1: Standard vs UPI** — Which create tool to call. Why: UPI links require `upi_link=true` in the API body; standard links must not send that flag, or the gateway treats them as UPI.
- **DP2: Notification medium** — `sms` or `email`. Why: Razorpay routes through different channels (telecom vs SMTP); an invalid value results in a 400 from the API.

### Update Flow

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Update mutable fields (common) | Agent calls `update_payment_link` | `payment_links.go:UpdatePaymentLink()` -> `client.PaymentLink.Update()` | DP3: any field provided? | Updated link returned |
| No-op guard (rare) | Agent sends empty update | `payment_links.go:UpdatePaymentLink()` empty-map check | DP3 fails | Error returned before API call |

**Decision Points:**
- **DP3: Empty update guard** — If no optional fields are populated, the tool returns an error locally rather than sending an empty PATCH. Why: preserves API quota and produces a clearer error message than the Razorpay API would return.

---

## Service Contracts

| Contract | Detail |
|----------|--------|
| Payment link ID format | `plink_xxx` — callers must supply this prefix; the SDK does not add it |
| Amount unit | Paise (integer). Minimum 100 (= ₹1) |
| Notification medium enum | `sms` \| `email` — no other values accepted |
| UPI link flag | Injected as `upi_link="true"` (string) in the API body by `CreateUpiPaymentLink()` |
| Downstream dependency | All tools call `client.PaymentLink.*` methods from `razorpay-go` SDK; if Razorpay API is down, all tools fail with the SDK error surfaced as a tool error result |
