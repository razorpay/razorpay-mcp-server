# pkg/toolsets — Toolset Management

Purpose: Groups tools into named toolsets, enforces read-only mode, and controls which toolsets are active at startup. All tool-to-server registration flows through this package.

## Key Files

| File | Role |
|------|------|
| `toolsets.go` | `Toolset`, `ToolsetGroup`, `NewToolset()`, `EnableToolsets()`, `RegisterTools()`, `AddToolset()` |

## Critical Rules

- **`NewToolset()` defaults to `Enabled: false`.** A toolset that is never explicitly enabled is silently skipped by `RegisterTools()`. No error, no warning — zero tools registered for that toolset.
- **Empty `--toolsets` slice enables ALL toolsets.** `EnableToolsets()` treats `nil` or `[]string{}` as "enable everything" (`everythingOn = true`). This inverts normal Go conventions — callers expecting zero toolsets from an empty input will get all of them.
- **Invalid toolset name is a hard startup error.** Passing an unrecognized name to `--toolsets` aborts server startup — it does not silently skip the unknown name.
- **`AddWriteTools()` is a no-op when `readOnly=true`.** Read-only mode is stamped at `AddToolset()` time. Write tools are dropped structurally at startup, not filtered per-request. There is no runtime override.
- **Do not add tools directly to the MCP server.** All tools must go through `AddReadTools()` / `AddWriteTools()` on a `Toolset`, then the toolset must be passed to `ToolsetGroup.AddToolset()`. Direct server registration bypasses read-only enforcement and annotation setting.
- **Annotations are set during `RegisterTools()`, not at tool creation.** `tool.go:SetReadOnly()` is called inside `RegisterTools()`. The `isReadOnly` flag on a tool is `false` until that moment regardless of which list it was added to.

## Common Operations

### Add a new toolset
1. In `tools.go:NewToolSets()`, call `toolsets.NewToolset(name, description)` and chain `.AddReadTools()` / `.AddWriteTools()`.
2. Call `toolsetGroup.AddToolset(myToolset)`.
3. No `server.go` changes needed — `RegisterTools()` iterates all toolsets in the group.

### Enable toolsets in tests
Call `EnableToolsets([]string{"payments", "orders"})` explicitly before `RegisterTools()`, or pass an empty slice to enable all. Never skip `EnableToolsets()` — `RegisterTools()` with all-disabled toolsets registers nothing.

## Load for Context

- `.agents/skills/repo-skill/technical-patterns.md` — toolset registration gotchas, read-only propagation, annotation timing
- `.agents/skills/repo-skill/core/quick-ref.md` — add-toolset step-by-step
