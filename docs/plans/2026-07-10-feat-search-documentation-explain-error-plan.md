---
title: Add search_documentation and explain_error MCP tools
type: feat
date: 2026-07-10
tdd: true
status: implemented, pr-raised
---

# Add `search_documentation` and `explain_error` tools

## Overview

Ported two tools already prototyped in the local Node.js companion server
(`razorpay-mcp-tools/`) into the official Go repo
(`razorpay/razorpay-mcp-server`):

- **`search_documentation`** — keyword-scored search across 17 indexed
  Razorpay doc sections, returns summary + runnable code example + doc URL.
- **`explain_error`** — 3-tier lookup (exact code+sub_description →
  code-only → keyword) across a registry of 56 Razorpay error codes, returns
  plain-English explanation, common causes, resolution steps, and (where
  applicable) a cross-linked guardrail rule (`guardrail_ref` +
  `guardrail_title`).

Neither tool calls the Razorpay API — both are pure local-data lookups.

## Implementation (as built)

Contrary to the `//go:embed` approach originally proposed below, the data
and logic were placed directly in `pkg/razorpay/` as hardcoded Go struct
literals, matching the existing hand-written-data convention used by
`pkg/razorpay/integrations/mobile.go` etc. — not `//go:embed`.

- **`pkg/razorpay/docs.go`** (single file, `package razorpay`) —
  - `docSection` / `errorEntry` struct types
  - `docSections []docSection` — 17 entries, hardcoded literals
  - `errorRegistryData []errorEntry` — 56 entries, hardcoded literals
  - `SearchDocumentation(obs *observability.Observability) mcpgo.Tool`
  - `ExplainError(obs *observability.Observability) mcpgo.Tool`
  - Both signatures **omit** the `client *rzpsdk.Client` param entirely
    (rather than accepting-and-ignoring it) since neither tool has any
    Razorpay API dependency — cleaner than the originally proposed
    signature-parity approach.
- **`pkg/razorpay/tools.go`** — new `documentation` toolset registered via
  `AddReadTools(SearchDocumentation(obs), ExplainError(obs))` and
  `toolsetGroup.AddToolset(documentation)`.
- **`pkg/razorpay/docs_test.go`** — 13 subtests across
  `TestSearchDocumentation` / `TestExplainError`, using direct handler
  invocation (`tool.GetHandler()(ctx, request)`) — no mocks, consistent with
  `detect_stack`/`checkout_tool`'s non-SDK test style.
- **`README.md`** — added 2 rows to the Available Tools table.

Why hardcoded literals over `//go:embed`: avoids introducing a new
file-loading pattern (JSON + embed + init-time unmarshal) into a repo where
every existing non-SDK data table (`integrations/mobile.go`, etc.) is
already hand-written Go. A one-time Go code generator (discarded after use)
converted the source JSON into struct literals to eliminate manual
transcription risk for the 56 error entries / 17 doc sections.

### Data source
- `razorpay-mcp-tools/data/docs-index.json` v1.0.0 — 17 sections.
- `razorpay-mcp-tools/data/error-registry.json` v1.3.0 — 56 errors.
- Scoring/fallback logic ported line-for-line from
  `razorpay-mcp-tools/server.js`'s `searchDocumentation`/`explainError`.

### Verification
- `go build ./...` — success
- `go vet ./...` — clean
- `gofmt -l` — clean
- `go test ./...` — 960 passed (full suite, no regressions)

## Remaining before PR
- [ ] Commit and push `feature/search-documentation-explain-error`
- [ ] Open PR against `razorpay/razorpay-mcp-server` `main`, calling out the
      hardcoded-literals-over-embed decision for reviewer visibility
