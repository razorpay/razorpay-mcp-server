# Quick Reference — razorpay-mcp-server

## 1. Add a New Tool

1. Create a function in `pkg/razorpay/{entity}.go` that returns `mcpgo.Tool`:
   - Declare `parameters []mcpgo.ToolParameter` using `mcpgo.WithString()`, `mcpgo.WithNumber()`, etc.
   - Build a handler `func(ctx, r mcpgo.CallToolRequest) (*mcpgo.ToolResult, error)`.
   - Return `mcpgo.NewTool(name, description, parameters, handler)`.
2. Register in `tools.go:NewToolSets()` by calling `.AddReadTools(YourTool(obs, client))` or `.AddWriteTools(YourTool(obs, client))` on the relevant toolset.

**Read vs Write classification:**
- `AddReadTools()` → sets `WithReadOnlyHintAnnotation(true)` + `WithDestructiveHintAnnotation(false)` at registration time (`toolsets.go:RegisterTools()`).
- `AddWriteTools()` → sets `WithReadOnlyHintAnnotation(false)` + `WithDestructiveHintAnnotation(true)`.
- The global `--read-only` flag causes `toolsets.go:AddWriteTools()` to silently drop all write tools at startup — no error, they are simply never registered.

## 2. Add a New Toolset

1. In `tools.go:NewToolSets()`, call `toolsets.NewToolset(name, description)` and chain `.AddReadTools()` / `.AddWriteTools()`.
2. Register with `toolsetGroup.AddToolset(myToolset)`.
3. No server.go change needed — `toolsetGroup.RegisterTools(s)` iterates all registered toolsets.

**Non-obvious:** `NewToolset()` defaults `Enabled: false`. A toolset that is created but not added to the group, or added but never enabled, is silently skipped. Enabling happens via `--toolsets` flag or by passing an empty slice (which enables all via `everythingOn` path in `toolsets.go:EnableToolsets()`).

## 3. Validate Tool Parameters

Use the fluent `Validator` from `tools_params.go`:

```
validator := NewValidator(&r).
    ValidateAndAddRequiredString(params, "payment_id").
    ValidateAndAddOptionalString(params, "description")

if result, err := validator.HandleErrorsIfAny(); result != nil {
    return result, err
}
```

- Required params: error returned if missing, collected and returned as one message.
- Optional params: absent or nil values are skipped without error.
- `HandleErrorsIfAny()` returns a `ToolResultError` (not a Go error) — the handler always returns `nil` for the Go `error` on validation failures.

## 4. Run Locally

```bash
make local-run
# or directly:
go run ./cmd/razorpay-mcp-server stdio --key KEY --secret SECRET
```

Enable specific toolsets:
```bash
go run ./cmd/razorpay-mcp-server stdio --key KEY --secret SECRET --toolsets payments,orders,refunds
```

Read-only mode (skips all write tools):
```bash
go run ./cmd/razorpay-mcp-server stdio --key KEY --secret SECRET --read-only
```

## 5. Run Tests

```bash
make test                # go test -race ./...
make test-coverage       # generates coverage.html for pkg/...
```

## 6. Mock Testing Pattern

Use `RazorpayToolTestCase` + `runToolTest()` from `test_helpers.go`:

```go
tests := []RazorpayToolTestCase{
    {
        Name:    "successful fetch",
        Request: map[string]interface{}{"payment_id": "pay_abc"},
        MockHttpClient: func() (*http.Client, *httptest.Server) {
            return mock.NewHTTPClient(mock.Endpoint{
                Path:     "/v1/payments/pay_abc",
                Method:   "GET",
                Response: map[string]interface{}{"id": "pay_abc"},
            })
        },
        ExpectError:    false,
        ExpectedResult: map[string]interface{}{"id": "pay_abc"},
    },
    {
        Name:           "missing param",
        Request:        map[string]interface{}{},
        MockHttpClient: nil, // skip HTTP mock for validation-only cases
        ExpectError:    true,
        ExpectedErrMsg: "missing required parameter: payment_id",
    },
}
for _, tc := range tests {
    t.Run(tc.Name, func(t *testing.T) {
        runToolTest(t, tc, FetchPayment, "Payment")
    })
}
```

**Non-obvious:** `newMockRzpClient()` in `test_helpers.go` sets `BaseURL` and `HTTPClient` on `rzpMockClient.Order.Request` — this is a shared `Request` object used by all resource types in the SDK, so mocking one resource mocks all of them.
