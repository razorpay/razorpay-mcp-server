# repo-skill: razorpay-mcp-server

Domain knowledge for the razorpay-mcp-server — a Go-based MCP server exposing Razorpay payment APIs to AI agents via the Model Context Protocol (stdio transport).

## Loading Rules

### always-load
Load these for every task in this repo:
- `core/boundaries.md` — service purpose, what this IS vs is NOT, architecture, auth, config, toolsets overview
- `core/quick-ref.md` — how to add tools, toolsets, run tests, read/write classification

### on-mention
Load when the named entity is mentioned in the task:

| Mention | Load |
|---------|------|
| payment, capture, OTP, S2S, token, revoke | `domain/payments.md` |
| payment link, UPI link, plink | `domain/payment-links.md` |
| order, mandate, recurring | `domain/orders.md` |
| refund, rfnd | `domain/refunds.md` |
| payout, pout | `domain/payouts.md` |
| QR code, qr_code | `domain/qr-codes.md` |
| settlement, instant settlement, setl, setlod, reconciliation | `domain/settlements.md` |
| checkout integration, detect_stack, integrate_razorpay | `domain/checkout-integration.md` |
| toolset, read-only mode, tool registration, ToolsetGroup | `technical-patterns.md` |
| MCP contract, tool schema, parameter schema, error format | `integration/service-contracts.md` |
| Razorpay SDK, API error, credential, RAZORPAY_KEY | `integration/external-deps.md` |

### on-file-change
Load when modifying these paths:

| Path pattern | Load |
|-------------|------|
| `pkg/toolsets/**` | `technical-patterns.md` |
| `pkg/razorpay/tools.go`, `pkg/razorpay/server.go` | `technical-patterns.md` |
| `pkg/razorpay/payments.go`, `pkg/razorpay/tokens.go` | `domain/payments.md` |
| `pkg/razorpay/orders.go` | `domain/orders.md` |
| `pkg/razorpay/refunds.go` | `domain/refunds.md` |
| `pkg/razorpay/payouts.go` | `domain/payouts.md` |
| `pkg/razorpay/qr_codes.go` | `domain/qr-codes.md` |
| `pkg/razorpay/settlements.go` | `domain/settlements.md` |
| `pkg/razorpay/payment_links.go` | `domain/payment-links.md` |
| `pkg/razorpay/integrations/**` | `domain/checkout-integration.md` |
| `cmd/razorpay-mcp-server/**` | `core/boundaries.md`, `integration/external-deps.md` |

## Trigger phrases
"add a tool", "new toolset", "read-only mode", "paise", "capture payment", "initiate payment", "OTP flow", "token", "instant settlement", "checkout integration", "MCP tool", "razorpay-mcp"
