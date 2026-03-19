---
description: "Guardrails for testing tool handlers in razorpay-mcp-server"
globs: ["pkg/razorpay/*_test.go", "pkg/razorpay/test_helpers.go", "pkg/razorpay/mock/**"]
---

- Use `RazorpayToolTestCase` + `runToolTest()` from `test_helpers.go` for all tool handler tests — do not write custom test loops or call tool handlers directly.
- `MockHttpClient` in test cases is a factory function `func() (*http.Client, *httptest.Server)` — called once per test case; server is closed by `runToolTest()` via defer.
- `test_helpers.go:newMockRzpClient()` patches the shared `Request` object — one patch covers ALL SDK resource types; no per-resource setup needed.
- Mock servers must return HTTP 400 + JSON error body for error test cases — non-JSON body produces a different error string than expected.
- Use `go-test/deep`'s `deep.Equal()` for comparing `ToolResult` values — not `reflect.DeepEqual()` or testify's `assert.Equal()`.
- Every new tool must test: (1) success path, (2) Razorpay API error, (3) validation error (missing/invalid required param).
- Tool handlers under test receive `context.Background()` with no SDK client — tools must accept `*rzpsdk.Client` as constructor arg, not retrieve from context.
- Mock HTTP response bodies must be valid JSON matching Razorpay API response shape — SDK unmarshals directly; mismatched shape causes silent zero-value fields.
