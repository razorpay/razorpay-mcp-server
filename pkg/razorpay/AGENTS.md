# pkg/razorpay — MCP Tool Handlers

Purpose: All MCP tool handler functions and toolset registration for every Razorpay resource. This is the primary working directory for adding, modifying, or debugging tools.

## Key Files

| File | Role |
|------|------|
| `payments.go` | Payment lifecycle tools: fetch, capture, initiate, OTP flows |
| `orders.go` | Order create, update, fetch, fetch-payments |
| `refunds.go` | Refund create, fetch, update, all-refunds |
| `payment_links.go` | Payment link CRUD tools |
| `payouts.go` | Payout fetch tools (read-only by nature) |
| `qr_codes.go` | QR code create, fetch, close, fetch-payments |
| `settlements.go` | Settlement fetch, reports, on-demand |
| `tokens.go` | `FetchSavedPaymentMethods`, `RevokeToken` — grouped inside `payments` toolset |
| `tools.go` | `NewToolSets()` — central toolset registry; registers all toolsets |
| `server.go` | `NewRzpMcpServer()` — builds MCP server, wires hooks, injects SDK client into context |
| `tools_params.go` | `Validator`, `formatErrorMessage()`, `ValidateAndAddExpand()` |
| `test_helpers.go` | `RazorpayToolTestCase`, `runToolTest()`, `newMockRzpClient()` |

## Critical Rules

- **NEVER write to stdout.** stdout is the MCP wire protocol. Any non-protocol bytes corrupt the session. Use the logger (file-based) for all diagnostics.
- **Amounts are in paise.** INR 100 paise = ₹1. Minimum 100 (= ₹1). Applies to payments, orders, refunds, and any tool accepting `amount`.
- **Use `formatErrorMessage()` for all API errors.** Produces `"<action> failed: <reason>"`. Empty API errors become `"resource does not exist"`.
- **`Validator.HandleErrorsIfAny()` returns `(result, nil)`.** The Go error return is always nil. Check `result != nil` to detect validation failures — checking only `err` silently swallows errors.
- **`initiate_payment` uses S2S JSON v1 only.** No redirect or browser checkout. Required because agents cannot render browser sessions.
- **Token tools live in the `payments` toolset.** `FetchSavedPaymentMethods` and `RevokeToken` are in `tokens.go` but registered under the `payments` toolset in `tools.go`.

## Common Operations

### Add a new tool
1. Write a function in `{entity}.go` returning `mcpgo.Tool` — declare parameters, build handler, call `mcpgo.NewTool()`.
2. Register in `tools.go:NewToolSets()`: call `.AddReadTools(YourTool(...))` or `.AddWriteTools(YourTool(...))` on the relevant toolset.
3. No changes to `server.go` needed — `RegisterTools()` iterates all toolsets automatically.

### Test a tool
Use `RazorpayToolTestCase` + `runToolTest()` from `test_helpers.go`. Pass `MockHttpClient` as a factory (`func() (*http.Client, *httptest.Server)`). For validation-only cases, set `MockHttpClient: nil`.

### Validate parameters
```go
validator := NewValidator(&r).
    ValidateAndAddRequiredString(params, "payment_id").
    ValidateAndAddOptionalString(params, "description")
if result, err := validator.HandleErrorsIfAny(); result != nil {
    return result, err
}
```

## Load for Context

- `.agents/skills/repo-skill/core/quick-ref.md` — add-tool and test patterns
- `.agents/skills/repo-skill/domain/payments.md` — payment constraints, OTP flow, token rules
- `.agents/skills/repo-skill/domain/orders.md` — order constraints, mandate rules
- `.agents/skills/repo-skill/technical-patterns.md` — Validator gotcha, mock testing gotcha
