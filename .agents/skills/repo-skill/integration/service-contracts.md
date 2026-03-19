# Service Contracts — razorpay-mcp-server

The MCP server exposes Razorpay API operations as MCP tools consumed by AI agent clients (Claude, Cursor, etc.).

## Tool Naming Convention

Pattern: `{verb}_{resource}` or `{verb}_{qualifier}_{resource}`

Examples: `fetch_payment`, `fetch_all_payments`, `fetch_specific_refund_for_payment`, `fetch_payment_card_details`

## Toolsets Exposed

| Toolset Name          | Read Tools                                      | Write Tools                                    |
|-----------------------|-------------------------------------------------|------------------------------------------------|
| `payments`            | fetch_payment, fetch_payment_card_details, fetch_all_payments | capture_payment, update_payment, initiate_payment, resend_otp, submit_otp, fetch_saved_payment_methods, revoke_token |
| `payment_links`       | fetch_payment_link, fetch_all_payment_links     | create_payment_link, create_upi_payment_link, resend_payment_link_notification, update_payment_link |
| `orders`              | fetch_order, fetch_all_orders, fetch_order_payments | create_order, update_order                 |
| `refunds`             | fetch_refund, fetch_multiple_refunds_for_payment, fetch_specific_refund_for_payment, fetch_all_refunds | create_refund, update_refund |
| `payouts`             | fetch_payout, fetch_all_payouts                 | (none)                                         |
| `qr_codes`            | fetch_qr_code, fetch_all_qr_codes, fetch_qr_codes_by_customer_id, fetch_qr_codes_by_payment_id, fetch_payments_for_qr_code | create_qr_code, close_qr_code |
| `settlements`         | fetch_settlement, fetch_settlement_recon, fetch_all_settlements, fetch_all_instant_settlements, fetch_instant_settlement | create_instant_settlement |
| `checkout_integration`| integrate_razorpay_checkout, detect_stack       | (none)                                         |

Agents can request a subset via `--toolsets` flag at startup. Empty `--toolsets` enables all. See `toolsets.go:EnableToolsets()`.

## Parameter Schema Patterns

Parameters are JSON Schema objects surfaced through the MCP protocol — agents discover them from the tool schema directly.

- **Required**: `mcpgo.Required()` — sets `"required": true` in JSON Schema. Agent must provide or tool returns validation error.
- **Optional with default**: No `Required()` call + optional `mcpgo.DefaultValue()`. Handler checks existence and falls back to hardcoded default (e.g., currency defaults to `"INR"` in `payments.go:InitiatePayment()`).
- **Enum constraints**: `mcpgo.Enum(...)` — emits `"enum": [...]` in schema. Only string enums are passed through to mcp-go. See `tool.go:addEnumOptions()`.
- **Numeric constraints**: `mcpgo.Min()` / `mcpgo.Max()` — emit `"minimum"` / `"maximum"` for numbers; `"minLength"` / `"maxLength"` for strings; `"minItems"` / `"maxItems"` for arrays. Type-dispatched in `tool.go:Min()`.
- **String pattern constraints**: `mcpgo.Pattern()` — emits `"pattern"` regex in schema, only valid on string type.

## Error Response Format

The Go `error` return from `ToolHandler` is **always nil**. All errors are returned in `ToolResult`, never as Go errors. See `tool.go:toMCPServerTool()`.

**Validation error** (before API call):
```
Validation errors:
- missing required parameter: payment_id
- invalid parameter type: amount
```
Produced by `tools_params.go:HandleErrorsIfAny()` — `IsError: true`.

**API error** (after API call):
```
fetching payment failed: <razorpay SDK error message>
```
Pattern: `"<action verb phrase>: <sdk error>"`. Produced by `tools_params.go:formatErrorMessage()` — `IsError: true`.

**Success**: `ToolResult.IsError = false`, `ToolResult.Text` is a JSON string of the Razorpay API response object. Produced by `tool.go:NewToolResultJSON()`.

## Read/Write MCP Annotations

Annotations are set in `toolsets.go:RegisterTools()` via `tool.SetReadOnly()`, which flows into `tool.go:toMCPServerTool()`:

- **Read tools**: `ReadOnlyHintAnnotation=true`, `DestructiveHintAnnotation=false`, `OpenWorldHintAnnotation=false` — AI clients may invoke automatically without confirmation.
- **Write tools**: `ReadOnlyHintAnnotation=false`, `DestructiveHintAnnotation=true`, `OpenWorldHintAnnotation=false` — AI clients should prompt for user confirmation before invoking.

**Global `--read-only` mode**: Write tools are **never registered** at all (not just annotated differently). `toolsets.go:AddWriteTools()` silently drops tools when `readOnly=true`. The toolset's write tool list remains empty.

## Toolset Filtering Gotcha

`toolsets.go:EnableToolsets()` with an empty slice enables **all** toolsets (not zero). An agent passing `--toolsets ""` gets everything; there is no way to start with zero tools except by not passing any recognized toolset name.
