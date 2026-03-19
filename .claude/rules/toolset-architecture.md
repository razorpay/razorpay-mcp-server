---
description: "Guardrails for toolset and tool registration in razorpay-mcp-server"
globs: ["pkg/toolsets/**", "pkg/razorpay/tools.go", "pkg/razorpay/server.go"]
---

- `toolsets.go:NewToolset()` always initializes `Enabled: false` — a toolset never passed to `EnableToolsets()` silently registers zero tools with no error.
- Passing an empty slice to `EnableToolsets()` enables ALL toolsets (`everythingOn = true`), not zero — inverse of normal Go conventions; do not rely on empty slice to mean "none".
- Passing an invalid toolset name to `EnableToolsets()` causes a hard startup error — validate names against `tools.go:NewToolSets()`.
- Read-only tools go in `AddReadTools()`, write/mutating tools in `AddWriteTools()` — mixing them breaks `--read-only` mode (`AddWriteTools()` silently no-ops when `readOnly=true`).
- Both MCP hint annotations must be set consistently: read = `ReadOnlyHintAnnotation=true` + `DestructiveHintAnnotation=false`; write = inverse — AI clients use these to decide whether to prompt for confirmation.
- `OpenWorldHintAnnotation` must always be `false` — tools operate on a bounded Razorpay account, not the open web.
- Never register tools directly on the MCP server — always go through `Toolset.AddReadTools/AddWriteTools` then `RegisterTools()`.
- Annotations are stamped during `toolsets.go:RegisterTools()` via `tool.SetReadOnly()`, not at tool creation — `isReadOnly` is false until `RegisterTools()` runs.
- When adding a new toolset, register it in both `tools.go:NewToolSets()` AND ensure `EnableToolsets()` can match its name — omitting either silently makes the toolset unreachable.
- `--read-only` is enforced structurally at startup (write tools never registered), not per-request — there is no runtime check inside handlers.
