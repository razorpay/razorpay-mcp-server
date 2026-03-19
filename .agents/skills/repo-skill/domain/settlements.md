# Settlements

Settlements represent the transfer of collected payment funds to a merchant's bank account.
Razorpay has two distinct types: **regular settlements** (automatic, scheduled by Razorpay) and
**instant settlements** (on-demand, merchant-triggered). They are separate API resources with
different ID namespaces and response shapes.

---

## Decisions

### D1: No create tool for regular settlements

**Context:** Regular settlements are Razorpay-initiated on a daily/weekly schedule.
**Decision:** Only read tools are exposed for regular settlements (`fetch_settlement_with_id`, `fetch_all_settlements`, `fetch_settlement_recon_details`). No create tool exists.
**Alternatives considered:**
- **Expose a create endpoint:** Not applicable — Razorpay's API does not allow merchants to trigger regular settlements; they happen automatically.
**Trade-offs:**
- Gained: Simpler surface; no risk of agents triggering unintended settlements.
- Lost: Nothing — the capability does not exist on the Razorpay side.
**Code:** `settlements.go:FetchSettlement()`, `settlements.go:FetchAllSettlements()`
**Revisit if:** Razorpay adds a merchant-initiated regular settlement API.

### D2: Separate tools for regular vs instant settlements

**Context:** Regular and instant settlements are different Razorpay API resources with different ID prefixes (`setl_` vs `setlod_`), different SDK methods, and different response shapes.
**Decision:** Distinct tool functions — `FetchSettlement()` / `FetchAllSettlements()` for regular; `FetchInstantSettlement()` / `FetchAllInstantSettlements()` / `CreateInstantSettlement()` for instant.
**Alternatives considered:**
- **Single unified tool with a type flag:** Would require the agent (or user) to know which type they're dealing with upfront, and the underlying SDK calls are entirely different.
**Trade-offs:**
- Gained: Clear tool intent; prevents ID prefix confusion at the MCP tool layer.
- Lost: Slightly larger tool surface.
**Code:** `settlements.go:FetchInstantSettlement()` vs `settlements.go:FetchSettlement()`
**Revisit if:** Razorpay unifies the two resources in their API.

---

## Non-Obvious Constraints

- **ID prefix must match resource type.** Regular settlement IDs start with `setl_`; instant settlement IDs start with `setlod_`. Passing a `setl_` ID to `fetch_instant_settlement_with_id` (or vice versa) returns an API error — not a client-side validation error.

- **Recon report requires year + month (mandatory); day is optional.** `FetchSettlementRecon()` enforces `year` and `month` as required integers. There is no way to fetch a recon report without scoping it to at least a month.

- **Instant settlement minimum amount is 200 paise (₹2).** Enforced client-side via `mcpgo.Min(200)` on the `amount` field in `CreateInstantSettlement()`. Amounts below this are rejected before the API call.

- **`settle_full_balance` overrides `amount`.** When `settle_full_balance: true`, Razorpay ignores the `amount` parameter and settles the maximum eligible balance. The `amount` field is still required by the tool schema even when this flag is set.

- **Merchant eligibility for instant settlements is enforced server-side.** The MCP tool does not validate eligibility — the Razorpay API returns an error if the merchant is not enabled for on-demand settlements. Agents should surface this error as-is.

- **`fetch_all_instant_settlements` supports payout expansion.** Pass `expand: ["ondemand_payouts"]` to include payout details inline. Only `ondemand_payouts` is a valid expand value; the schema enforces this via enum.

---

## Flow Map

### Fetch Regular Settlement

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Fetch by ID (most traffic) | Agent provides `setl_` prefixed ID | `settlements.go:FetchSettlement()` -> `client.Settlement.Fetch()` | DP1: ID prefix | Settlement JSON returned |
| Fetch all with filters (common) | Agent provides optional from/to/count/skip | `settlements.go:FetchAllSettlements()` -> `client.Settlement.All()` | — | Paginated list returned |
| Fetch recon report | Agent provides year + month (+ optional day) | `settlements.go:FetchSettlementRecon()` -> `client.Settlement.Reports()` | DP2: date params required | Reconciliation report returned |

**Decision Points:**
- **DP1: ID prefix** — Regular settlement IDs must be `setl_xxx`. Using an instant settlement ID (`setlod_xxx`) here returns an API error.
- **DP2: date params required** — `year` and `month` are mandatory; the tool rejects the call before hitting the API if either is missing.

### Create Instant Settlement

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Create with explicit amount (most traffic) | Agent provides `amount` in paise | `settlements.go:CreateInstantSettlement()` -> `client.Settlement.CreateOnDemandSettlement()` | DP3: min amount; DP4: merchant eligibility | Instant settlement created, `setlod_` ID returned |
| Settle full balance (common) | Agent sets `settle_full_balance: true` | `settlements.go:CreateInstantSettlement()` -> `client.Settlement.CreateOnDemandSettlement()` | DP4: merchant eligibility | Maximum eligible amount settled |
| Ineligible merchant (rare) | Any create attempt | `settlements.go:CreateInstantSettlement()` | DP4 | API error returned as tool error |

**Decision Points:**
- **DP3: minimum amount** — Amounts below 200 paise are rejected client-side by the validator before any API call.
- **DP4: merchant eligibility** — Razorpay enforces eligibility server-side; no client-side check. Error is surfaced directly.

---

## Service Contracts

| Contract | Detail |
|----------|--------|
| Regular settlement ID format | `setl_xxx` — passed to `fetch_settlement_with_id` |
| Instant settlement ID format | `setlod_xxx` — passed to `fetch_instant_settlement_with_id` |
| Recon report scope | Requires `year` (YYYY) + `month` (MM); `day` (DD) optional |
| Instant settlement amount | In paise; minimum 200 (₹2) |
| Payout expansion | `expand: ["ondemand_payouts"]` on `fetch_all_instant_settlements` only |
| Downstream | All reads/writes go through `rzpsdk.Client.Settlement.*` — no direct HTTP calls |
