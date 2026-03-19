---
description: "Guardrails for implementing MCP tool handlers in razorpay-mcp-server"
globs: ["pkg/razorpay/*.go", "pkg/mcpgo/*.go"]
---

- NEVER write to stdout inside any tool handler or hook — stdout is the MCP wire protocol (JSON-RPC); any non-protocol bytes corrupt the session. All logging goes to the file logger via `--log-file`.
- `tools_params.go:HandleErrorsIfAny()` returns `(*ToolResult, error)` where error is ALWAYS nil — check `result != nil`, NOT `err != nil`; checking only `err` silently swallows all validation errors.
- The Go `error` return from a `ToolHandler` must always be `nil` — return user-visible errors as `mcpgo.NewToolResultError()` inside `ToolResult`; non-nil Go error escalates to a protocol-level MCP failure.
- Retrieve the Razorpay SDK client from the constructor argument — tool constructors accept `*rzpsdk.Client` explicitly; do not use `contextkey.ClientFromContext(ctx)` in handlers (causes nil-pointer panic in tests where context has no client).
- Format API error messages with `tools_params.go:formatErrorMessage("action phrase", err)` — pattern: `"fetching payment failed: <sdk error>"`.
- All amount parameters representing money must be in paise (100 paise = ₹1) — document this unit explicitly in the tool description.
- Tool descriptions are the primary interface contract for AI clients — state required preconditions, unit conventions (paise), and irreversibility warnings directly in the description string.
- `checkout_integration` toolset tools make NO Razorpay API calls — they are local code-generation helpers; do not add API calls to them.
