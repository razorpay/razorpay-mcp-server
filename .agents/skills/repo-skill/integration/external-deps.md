# External Dependencies

## Razorpay API (razorpay-go v1.4.0)

The only external runtime dependency is the Razorpay REST API, consumed via the official Go SDK.

**SDK:** `github.com/razorpay/razorpay-go v1.4.0`

**Client init:** `stdio.go` calls `rzpsdk.NewClient(key, secret)` once at startup — Basic Auth, shared for all tool calls in the process lifetime. The single client instance is passed into `razorpay.NewRzpMcpServer()` and stored in context via `contextkey.WithClient()` / `contextkey.ClientFromContext()`.

**User-Agent:** `"razorpay-mcp{version}/stdio"` — set via `client.SetUserAgent()` immediately after init. The `version` variable is injected at build time via `-ldflags`.

**Base URL:** Default Razorpay API endpoint (`api.razorpay.com`). No custom base URL override exists in this server.

---

## Credential Configuration (Priority Order)

Viper resolves credentials in this order (highest priority first):

1. CLI flags: `--key` / `--secret`
2. Environment variables: `RAZORPAY_KEY_ID` / `RAZORPAY_KEY_SECRET`
3. YAML config file: `~/.razorpay-mcp-server` (keys: `key`, `secret`)

**Non-obvious naming mismatch:** `main.go` maps `RAZORPAY_KEY_ID` → viper key `key` and `RAZORPAY_KEY_SECRET` → viper key `secret` via explicit `viper.BindEnv()` calls. The env var names do NOT match the viper keys. Agents or tooling that assume `AUTOMATICENV` would resolve `RAZORPAY_KEY_ID` as viper key `razorpay_key_id` will get no value — the explicit `BindEnv` is required and intentional.

---

## API Error Propagation

| Step | Code | Behavior |
|------|------|----------|
| SDK call fails | tool handler | SDK returns Go `error` |
| Format message | `tools_params.go:formatErrorMessage()` | Produces `"<prefix>: <err.Error()>"` or `"<prefix>: resource does not exist"` when err is nil |
| Return to MCP client | `mcpgo/tool.go:NewToolResultError()` | Sets `IsError=true` in MCP result |

No retry logic, no circuit breaker, no timeout configuration at the MCP layer. The SDK applies its own HTTP timeout defaults. Tool returns error result immediately on any API failure.

---

## Local stdio vs Hosted Remote

| Aspect | Local stdio (this codebase) | Hosted (mcp.razorpay.com) |
|--------|-----------------------------|---------------------------|
| Auth | Credentials in env vars or CLI flags | `Authorization: Basic <base64(key:secret)>` header |
| Tool availability | All enabled toolsets | Subset — no write tools for some categories |
| Implementation | This repo | Separate Razorpay service, not in this codebase |

This codebase does NOT implement the hosted endpoint. That is a separate Razorpay-managed service.

---

## Gaps Needing Human Verification

- SDK HTTP timeout defaults are set inside `razorpay-go` — not visible in this repo. Verify in `razorpay-go v1.4.0` source if timeout behavior matters.
- The exact subset of tools available on the hosted endpoint is documented only in README and may drift — verify against `mcp.razorpay.com` if completeness matters.
