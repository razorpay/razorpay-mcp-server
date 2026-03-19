# Service Boundaries

## Purpose

This repository is an MCP (Model Context Protocol) server that exposes Razorpay payment APIs as tools to AI agents and LLM clients. It translates MCP tool-call requests into Razorpay Go SDK calls and returns structured results.

## What This Is / Is Not

**IS:**
- A protocol adapter: MCP tool calls -> Razorpay Go SDK -> `api.razorpay.com`
- A tool registry: groups Razorpay API operations into named toolsets, enforces read-only mode
- A local binary (stdio mode) or a client to the hosted remote endpoint

**IS NOT:**
- A payment gateway — all actual payment processing happens at `api.razorpay.com`
- A webhook handler — no inbound event processing
- The hosted `mcp.razorpay.com` endpoint — that is a separate Razorpay-operated service; this repo only connects to it as a client via `mcp-remote`
- A multi-transport server — only stdio transport is implemented in this repo (SSE/HTTP are not)

## Architecture

```
CLI (Cobra/Viper)
  -> stdio.go:runStdioServer()
    -> razorpay/server.go:NewRzpMcpServer()      # builds MCP server with hooks
      -> razorpay/tools.go:NewToolSets()          # creates ToolsetGroup
        -> toolsets/toolsets.go:EnableToolsets()  # enables named or all toolsets
        -> toolsets/toolsets.go:RegisterTools()   # registers read/write tools per mode
      -> mcpgo.StdioServer.Listen()               # stdin/stdout MCP protocol loop
        -> individual tool handlers               # call Razorpay Go SDK
          -> api.razorpay.com
```

**Non-obvious:** `checkout_integration` toolset tools (`DetectStack`, `IntegrateRazorpayCheckout`) do NOT call `api.razorpay.com` — they are local code generation helpers with no external API calls.

## Auth & Configuration

**Auth:** Basic Auth only (Razorpay Key ID + Key Secret). Credentials are passed to `rzpsdk.NewClient()` and embedded in every outbound SDK call. No OAuth, no token refresh.

**Config resolution order (highest to lowest priority):**
1. CLI flags (`--key`, `--secret`, `--log-file`, `--toolsets`, `--read-only`)
2. Environment variables (`RAZORPAY_KEY_ID`, `RAZORPAY_KEY_SECRET`)
3. YAML config file (`~/.razorpay-mcp-server.yaml`)

**Non-obvious constraint:** `RAZORPAY_KEY_ID` maps to the internal viper key `key`; `RAZORPAY_KEY_SECRET` maps to `secret`. Standard `RAZORPAY_KEY` env var name does NOT work — see `main.go:init()`.

**Read-only mode** (`--read-only`): propagates through `ToolsetGroup` at construction time via `toolsets.go:AddToolset()`. Write tools are excluded from `AddWriteTools()` and never registered — this is enforced at startup, not per-request.

**Toolset selective loading** (`--toolsets`): passing an empty slice enables ALL toolsets (`everythingOn = true`). Passing an invalid toolset name returns a hard error at startup — the server will not start.

## Toolsets

8 toolsets registered in `tools.go:NewToolSets()`:

| Toolset | Read Tools | Write Tools | Notes |
|---------|-----------|-------------|-------|
| `payments` | 3 | 7 | Includes token tools (FetchSavedPaymentMethods, RevokeToken) — non-obvious grouping |
| `payment_links` | 2 | 4 | |
| `orders` | 3 | 2 | |
| `refunds` | 4 | 2 | |
| `payouts` | 2 | 0 | Read-only by nature; no write tools defined |
| `qr_codes` | 5 | 2 | |
| `settlements` | 5 | 1 | |
| `checkout_integration` | 2 | 0 | No API calls; local code generation only |

**Non-obvious:** Token management tools (`FetchSavedPaymentMethods`, `RevokeToken`) live inside the `payments` toolset despite being token operations. Enabling `payments` is required to access them.

**Non-obvious:** `create_refund` and `close_qr_code` are NOT supported on the remote hosted server (`mcp.razorpay.com`) per README — they work only in local stdio mode.

## Deployment Modes

### Mode 1: Local stdio (this repo's binary)

Process runs as a subprocess of the MCP client (e.g., Claude Desktop, Cursor). Communication over stdin/stdout. Credentials supplied via env vars, CLI flags, or YAML config.

```json
{
  "command": "/path/to/razorpay-mcp-server",
  "args": ["stdio"],
  "env": { "RAZORPAY_KEY_ID": "...", "RAZORPAY_KEY_SECRET": "..." }
}
```

**Non-obvious:** Logs MUST go to a file (`--log-file`), never to stdout — stdout is the MCP protocol stream. Writing logs to stdout corrupts the MCP session.

### Mode 2: Hosted remote (mcp.razorpay.com)

This repo is NOT deployed as the remote server. The client connects via `npx mcp-remote` with a Basic Auth header encoded as `Base64(key_id:key_secret)`. The hosted endpoint has its own restrictions (some write tools unsupported).

```json
{
  "command": "npx",
  "args": ["mcp-remote", "https://mcp.razorpay.com/mcp", "--header", "Authorization:${AUTH_HEADER}"],
  "env": { "AUTH_HEADER": "Basic <Base64(key:secret)>" }
}
```
