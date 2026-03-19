# razorpay-mcp-server

Go MCP (Model Context Protocol) server that exposes Razorpay payment APIs as tools to AI agents via stdio transport. It translates MCP tool-call requests into Razorpay Go SDK calls and returns structured results. Not a payment gateway — all payment processing happens at `api.razorpay.com`.

## Tech Stack

Go 1.24.2 · mark3labs/mcp-go v0.43.2 · razorpay-go SDK v1.4.0 · Cobra/Viper CLI · Docker (alpine)

## Quick Start

```bash
make local-build                        # build native binary to ./bin/
make test                               # run tests with race detector
make lint                               # golangci-lint
make local-run                          # run locally (set env vars first)

# Set credentials before running:
export RAZORPAY_KEY_ID=rzp_...
export RAZORPAY_KEY_SECRET=...
go run ./cmd/razorpay-mcp-server stdio --key $RAZORPAY_KEY_ID --secret $RAZORPAY_KEY_SECRET
```

## Architecture

```
CLI (Cobra/Viper) -> stdio.go -> razorpay/server.go -> razorpay/tools.go:NewToolSets()
  -> toolsets:EnableToolsets() -> toolsets:RegisterTools() -> mcpgo.StdioServer.Listen()
    -> individual tool handlers -> razorpay-go SDK -> api.razorpay.com
```

- **Owns:** MCP protocol adapter, tool registration, toolset management, read-only mode enforcement
- **Does NOT own:** payment processing, hosted `mcp.razorpay.com` endpoint, OAuth flows
- **stdout = MCP wire protocol** — NEVER log to stdout; use `--log-file` for all logging
- **Transport:** stdio only (HTTP/SSE not implemented in this repo)
- **checkout_integration** toolset tools make NO API calls — they are local code-generation helpers

## Domain Entities (8 toolsets, ~50 tools)

- **payments** — capture, initiate (S2S), OTP flow, saved tokens; token tools live here too
- **payment_links** — standard and UPI-only payment links; SMS/email notification
- **orders** — regular and mandate (recurring) orders; notes-only updates
- **refunds** — full/partial refunds; scoped to a payment or global fetch
- **payouts** — read-only; fetch by account number only; no create/cancel
- **qr_codes** — UPI QR codes; `close_qr_code` is irreversible
- **settlements** — regular (`setl_` prefix) vs instant (`setlod_` prefix); recon report
- **checkout_integration** — `detect_stack` + `integrate_razorpay_checkout` (dev-assist, no API calls)

## Key Patterns & Gotchas

- **Amounts in paise** — all money params are in paise (100 paise = ₹1); document in tool descriptions
- **Config naming mismatch** — env var `RAZORPAY_KEY_ID` maps to viper key `key` (not `razorpay_key_id`); `RAZORPAY_KEY` does NOT work
- **Empty `--toolsets` = ALL enabled** — counterintuitive; empty slice triggers `everythingOn=true`
- **Validator return pattern** — `HandleErrorsIfAny()` returns `(result, nil)`; check `result != nil`, NOT `err != nil`
- **Go error from handlers must be nil** — use `mcpgo.NewToolResultError()` for user-visible errors
- **`capture_payment` requires `status=authorized`** — two-step Razorpay settlement model
- **OTP flow is sequential** — `initiate_payment` → `resend_otp` (opt) → `submit_otp`; no OTP without payment in flight
- **Read-only enforced at startup** — write tools never registered, not filtered per-request

## Deeper Context

For detailed decisions, constraints, and flow maps:
- **Repo skill:** `.agents/skills/repo-skill/SKILL.md` — load by entity name for progressive disclosure
- **Path-scoped rules:** `.claude/rules/*.md` — toolset architecture, tool implementation, payments domain, testing
- **Nested AGENTS.md:** `pkg/razorpay/`, `pkg/toolsets/`, `pkg/mcpgo/`, `pkg/razorpay/integrations/`

## Skills Index

| Skill | Trigger |
|-------|---------|
| **repo-skill** | Any domain question — load `.agents/skills/repo-skill/SKILL.md` |
| **go-code-reviewer** | Review Go code, PRs, check for race conditions, transaction context bugs |
| **code-review** | Comprehensive Go code review (Uber Go style, error handling, concurrency) |
| **code-security** | Security review, auth changes, input validation, injection risks |
| **pre-mortem** | Pre-merge PR checks — infrastructure patterns, service contracts, observability |
| **tech-spec-reviewer** | Review tech specs, design docs, architecture proposals |
| **api-compliance** | Razorpay API council submission review |
| **semgrep-devrev-remediation** | Remediate SCA vulnerable dependencies via DevRev tickets |
| **log-volume-optimizer** | Reduce log volume, optimize Coralogix costs |
| **devstack** | Deploy/debug helmfile services, local dev setup |
| **project-planner** | Create project plans and task breakdowns from specs or design docs |

## Agent Config

| File | Agent | Purpose |
|------|-------|---------|
| `AGENTS.md` | All | This file — concise service map |
| `CLAUDE.md` | Claude Code | Symlink → AGENTS.md |
| `.claude/rules/` | Claude Code | Path-scoped domain rules (toolset-architecture, tool-implementation, payments-domain, testing) |
| `.agents/skills/repo-skill/` | All | Extracted domain knowledge with progressive disclosure |
| `.claude/settings.json` | Claude Code | Session hooks (agentfill AGENTS.md loader) |
