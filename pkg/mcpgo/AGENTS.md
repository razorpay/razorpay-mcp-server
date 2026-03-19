# pkg/mcpgo — MCP Abstraction Layer

Purpose: Thin abstraction over `mark3labs/mcp-go`. Provides the server, transport, tool types, and hook lifecycle used by the rest of the codebase. Isolates the upstream MCP library from tool handler code.

## Key Files

| File | Role |
|------|------|
| `server.go` | `Server` interface, `SetupHooks()`, `NewMcpServer()` |
| `tool.go` | `Tool`, `ToolParameter`, `toMCPServerTool()`, `SetReadOnly()` |
| `stdio.go` | `StdioServer`, `Listen()` — runs the stdin/stdout protocol loop |
| `transport.go` | Transport interface wrapping `mark3labs/mcp-go` transports |

## Critical Rules

- **stdout is the MCP wire protocol — NEVER log to stdout.** The `StdioServer` reads/writes JSON-RPC over stdin/stdout. Any non-protocol byte written to stdout (fmt.Println, log.Println, etc.) corrupts the session irreversibly for that connection.
- **Go error return from tool handlers must always be `nil`.** Returning a non-nil Go `error` from a handler propagates as a protocol-level MCP failure, not a user-visible tool error. Use `mcpgo.NewToolResultError()` to surface errors to the caller.
- **Use `mcpgo.NewToolResultError()` for all tool-level errors.** This produces a `*ToolResult` with `IsError: true` that the MCP client displays as a failed tool call, while keeping the MCP session alive.
- **Annotations are applied in `tool.go:SetReadOnly()`**, called by `toolsets.go:RegisterTools()` — not at tool construction time. A tool's `isReadOnly` field is `false` until `RegisterTools()` runs.
- **`OpenWorld = false` on all tools.** Signals to AI clients that tools operate on bounded, account-scoped data — not open-ended searches. This affects how clients present tool results.

## Hook Lifecycle

Hooks are registered in `server.go:SetupHooks()` and fire in this order per tool call:

```
BeforeAny -> BeforeCallTool -> [handler executes] -> AfterCallTool -> OnSuccess | OnError
```

- `OnSuccess` for `MethodToolsList` logs only tool names, not full schemas — schemas are large and flood log files on every client connection.
- All hooks write to the observability logger (file or stderr), never to stdout.

## Common Operations

### Return an error from a tool handler
```go
return mcpgo.NewToolResultError("fetching payment failed: " + err.Error()), nil
```
Never return `(nil, err)` — that escalates to a protocol error.

### Return a successful JSON result
```go
return mcpgo.NewToolResultJSON(responseData), nil
```

## Load for Context

- `.agents/skills/repo-skill/technical-patterns.md` — hook lifecycle details, SDK client propagation via context, stdout constraint rationale
