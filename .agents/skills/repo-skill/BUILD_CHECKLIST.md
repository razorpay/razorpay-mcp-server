# BUILD_CHECKLIST.md — razorpay-mcp-server

Agent-readiness extraction checklist. Each task produces one output file.
Mark items `[x]` as completed. Run `/agent-ready:pickup` to resume.

---

## Wave 1 — Core (no dependencies, run first)

- [x] `core/boundaries.md` — Service purpose, deployment modes (local stdio vs hosted mcp.razorpay.com), what this repo is and is NOT, architecture overview (toolset group → toolsets → tools → Razorpay SDK), auth mechanisms (Basic Auth only in this repo), config options (env vars, CLI flags, YAML file), read-only mode semantics
- [x] `core/quick-ref.md` — Common developer operations: how to add a new tool, how to add a toolset, how to run tests, how to run locally, how read/write classification works, how to enable/disable specific toolsets, how to run in read-only mode

---

## Wave 2 — Domain (all run in parallel after Wave 1)

- [x] `domain/payments.md` — Payments toolset: read/write tool split, S2S JSON v1 flow for `initiate_payment`, OTP flow (resend/submit sequence), amount-in-paisa constraint, authorized-status requirement for capture, token management (fetch_tokens/revoke_token belong here), saved payment methods via contact number, flow map for payment lifecycle
- [x] `domain/payment-links.md` — Payment links toolset: standard vs UPI payment link distinction, notification channels (SMS vs email), update constraints, flow map
- [x] `domain/orders.md` — Orders toolset: regular orders vs mandate orders (recurring payments) distinction in `create_order`, notes-only update constraint, relationship between order and payment (fetch_order_payments), flow map
- [x] `domain/refunds.md` — Refunds toolset: refund scoping (by payment vs global), `fetch_multiple_refunds_for_payment` vs `fetch_specific_refund_for_payment` distinction, default-10-refunds behavior, amount in paise, notes-only update constraint, flow map
- [x] `domain/payouts.md` — Payouts toolset: read-only by design (no create/update), fetch by account number constraint, how payouts differ from refunds (bank account vs payment reversal), flow map
- [x] `domain/qr-codes.md` — QR codes toolset: UPI-specific QR codes, close operation is irreversible, multi-filter fetch patterns (by customer, by payment), fetch payments via QR, flow map
- [x] `domain/settlements.md` — Settlements toolset: regular vs instant settlements distinction, `setlod_` prefix for instant settlement IDs, reconciliation report time-period requirement, create_instant_settlement constraints, flow map
- [x] `domain/checkout-integration.md` — Checkout integration toolset: `detect_stack` + `integrate_razorpay_checkout` are developer-assist tools (not payment ops), code generation scope (backend routes + frontend + verification), supported tech stacks, when to use this vs direct API tools

---

## Wave 3 — Technical (parallel with Wave 2)

- [x] `technical-patterns.md` — Non-obvious infrastructure patterns: toolset registration flow (NewToolSets → ToolsetGroup → RegisterTools), read/write tool annotation system (ReadOnlyHintAnnotation + DestructiveHintAnnotation), global read-only mode skipping write tools, toolset filtering via `--toolsets` flag, Razorpay SDK client propagation via context (contextkey pattern), hook-based lifecycle logging (BeforeAny/OnSuccess/OnError/BeforeCallTool/AfterCallTool), parameter validation via fluent Validator pattern, error surfacing format (formatErrorMessage), mock HTTP server pattern for testing, version embedding via LDFLAGS

---

## Wave 4 — Integration (after Wave 2)

- [x] `integration/service-contracts.md` — MCP protocol surface: tool naming conventions, parameter schema patterns (required vs optional, enum constraints, min/max), error response format (IsError flag + text), toolset capability declaration, how AI agents discover and invoke tools, read/write annotations as MCP hints
- [x] `integration/external-deps.md` — Razorpay API dependency: SDK version (razorpay-go v1.4.0), credential configuration (env/flag/YAML priority order), User-Agent header pattern, how API errors propagate to MCP error responses, hosted remote endpoint (mcp.razorpay.com) vs local stdio distinction, Basic Auth encoding for remote endpoint

---

## Summary

| Wave | Tasks | Output Files |
|------|-------|-------------|
| Wave 1 - Core | 2 | core/boundaries.md, core/quick-ref.md |
| Wave 2 - Domain | 8 | domain/{payments,payment-links,orders,refunds,payouts,qr-codes,settlements,checkout-integration}.md |
| Wave 3 - Technical | 1 | technical-patterns.md |
| Wave 4 - Integration | 2 | integration/service-contracts.md, integration/external-deps.md |
| **Total** | **13** | **13 files** |
