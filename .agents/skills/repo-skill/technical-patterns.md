# Technical Patterns — razorpay-mcp-server

Non-obvious patterns only. Agents read the code for schemas, signatures, and step-by-step logic.

---

## 1. Toolset Registration Flow

**Sequence:** `NewToolSets()` builds domain toolsets → `NewToolsetGroup()` holds all → `EnableToolsets()` activates selected → `RegisterTools()` adds active tools to MCP server.

**Gotcha — all toolsets start disabled:** `toolsets.go:NewToolset()` always initializes `Enabled: false`. No toolset is active until `EnableToolsets()` runs. Forgetting to call `EnableToolsets()` before `RegisterTools()` silently registers zero tools — no error, no warning.

**Gotcha — empty slice means all enabled:** `toolsets.go:EnableToolsets()` treats an empty `names` slice as "enable everything" (sets `everythingOn = true`). A non-empty slice enables only named toolsets. This inverts normal Go conventions — callers passing `nil` or `[]string{}` get all toolsets, not zero.

**Gotcha — readOnly propagates at group level:** `toolsets.go:AddToolset()` stamps `ts.readOnly = true` when `ToolsetGroup.readOnly` is set. Write tools added *after* group creation are still silently dropped via `AddWriteTools()` guard, but write tools added *before* group addition are discarded at `RegisterTools()` time. The `--read-only` flag is enforced structurally, not at request time.

---

## 2. Read/Write Annotation System

**What it is:** MCP protocol-level hints to AI clients, set in `tool.go:toMCPServerTool()`.

- Read tools: `WithReadOnlyHintAnnotation(true)` + `WithDestructiveHintAnnotation(false)` + `WithOpenWorldHintAnnotation(false)`
- Write tools: `WithReadOnlyHintAnnotation(false)` + `WithDestructiveHintAnnotation(true)` + `WithOpenWorldHintAnnotation(false)`

**Why `OpenWorld = false`:** Signals to AI clients that tools operate on known, bounded data (a specific Razorpay account), not open-ended web searches. This affects how clients like Claude present tool results.

**Gotcha — annotations are set during RegisterTools, not at tool creation:** `tool.go:SetReadOnly()` is called inside `toolsets.go:RegisterTools()`, after the tool is already built. The `isReadOnly` field on `mark3labsToolImpl` starts as `false` regardless of which list (readTools/writeTools) the tool was added to.

---

## 3. SDK Client Propagation via Context

**Pattern:** Client is injected into `context.Context` at startup (stdio.go) and extracted inside each tool handler via `contextkey.go:ClientFromContext()`.

**Why context, not constructor injection:** Tool handlers are registered as closures at startup but called later per-request. The context approach avoids threading the client through every intermediate MCP framework layer.

**Gotcha — no nil guard in tool handlers:** `contextkey.go:ClientFromContext()` returns `nil` if no client is found. Tool handlers type-assert the result directly. If context is missing the client (e.g., middleware strips it), handlers panic at the type assertion — not a graceful error. Tests avoid this by passing `rzpsdk.Client` directly via `runToolTest()`, bypassing context entirely.

---

## 4. Parameter Validation Pattern (Validator)

**Non-obvious return convention:** `tools_params.go:HandleErrorsIfAny()` returns `(*ToolResult, error)` where the error return is **always nil**. Validation failures come back as a non-nil `*ToolResult` with `IsError: true`, not as a Go `error`.

**Correct caller pattern:**
```go
result, err := validator.HandleErrorsIfAny()
if err != nil { return nil, err }   // never triggers
if result != nil { return result, nil }  // this catches validation failures
```

Checking only `err` silently swallows all validation failures and proceeds with invalid params. This is intentional: MCP tool results must always be `(*ToolResult, nil)` — returning a Go error bubbles up as a protocol-level failure, not a user-visible tool error.

**Gotcha — expand[] serialization:** `tools_params.go:ValidateAndAddExpand()` writes the same key `"expand[]"` multiple times into the params map. Only the last value survives in a `map[string]interface{}`. This works because the underlying Razorpay SDK re-serializes the map into query params, but it means expand with multiple values may silently send only one. Needs human verification.

---

## 5. Hook-Based Lifecycle Logging

**Why hooks write to file, not stdout:** stdout is the MCP wire protocol (JSON-RPC). Any non-protocol bytes written to stdout corrupt the session. The observability logger (configured in stdio mode) writes to a file or stderr — never stdout. `server.go:SetupHooks()` registers five hooks: `BeforeAny`, `OnSuccess`, `OnError`, `BeforeCallTool`, `AfterCallTool`.

**Special case — ToolsList truncation:** `server.go:SetupHooks()` `OnSuccess` hook detects `mcp.MethodToolsList` and logs only tool names, not full JSON schemas. Tool schemas can be large (hundreds of lines); logging them floods observability files on every client connection.

---

## 6. Mock HTTP Testing Pattern

**How it works:** `test_helpers.go:newMockRzpClient()` creates a real `rzpsdk.Client`, then patches `rzpMockClient.Order.Request` — the single `Request` object shared by reference across all resource types (payments, orders, refunds, etc.) in the SDK client struct. One patch covers all API resources.

**Gotcha — MockHttpClient is a factory, not a client:** `RazorpayToolTestCase.MockHttpClient` is `func() (*http.Client, *httptest.Server)` — called once per test case inside `newMockRzpClient()`. The test server is created fresh per case and must be closed by the caller (deferred in `runToolTest()`). Reusing a closed server across test cases causes silent failures.

**Gotcha — tool handlers bypass context in tests:** `test_helpers.go:runToolTest()` calls `tool.GetHandler()(context.Background(), request)` with a plain background context — no client in context. Tools under test must accept the client as a constructor argument (not from context) or tests will panic. This is why all tool constructors take `*rzpsdk.Client` explicitly.
