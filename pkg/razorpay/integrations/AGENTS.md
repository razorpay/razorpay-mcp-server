# pkg/razorpay/integrations — Checkout Integration Tools

Purpose: Developer-assist tools that help agents integrate Razorpay Checkout into a merchant's codebase. These tools generate code snippets and detect the tech stack. They do NOT execute payments or call any Razorpay API.

## Key Files

| File | Role |
|------|------|
| `checkout_tool.go` | `IntegrateRazorpayCheckout()` — generates frontend + backend integration code |
| `detect_stack_tool.go` | `DetectStack()` — infers tech stack from file/dependency info supplied by the agent |
| `backend_go.go` / `backend_node.go` / `backend_python.go` / `backend_java.go` / `backend_other.go` | Backend-specific code templates per language |
| `frontend_templates.go` | Frontend HTML/JS snippet templates |
| `mobile.go` | Mobile (Android/iOS) integration templates |
| `helpers.go` | Shared template rendering utilities |
| `types.go` | `Stack`, `IntegrationConfig`, and related types |

## Critical Rules

- **These tools make zero API calls.** No `rzpsdk.Client` usage. No calls to `api.razorpay.com`. Errors from these tools are template or logic errors, not network errors.
- **`DetectStack` is passive — it cannot read the filesystem itself.** The agent must supply file names and dependency content as tool parameters. The tool infers the stack from what the agent provides; it does not scan the user's project directly.
- **Go detection defaults to gin if the router cannot be determined.** When Go is detected but no known router is identified from the provided dependency info, the generated backend snippet uses gin. The agent should verify with the user if gin is not the actual router.
- **`checkout_integration` toolset is disabled when `--toolsets` is used without naming it.** If the user passes `--toolsets payments,orders`, the `checkout_integration` toolset is silently excluded. The agent must name it explicitly (e.g., `--toolsets payments,checkout_integration`) or omit `--toolsets` entirely (which enables all toolsets).
- **Generated code is a starting template, not production-ready.** Snippets use placeholder values for `key_id` and `order_id`. The agent should communicate this to the user.

## Common Operations

### Understand a generated snippet
Read the relevant `backend_{lang}.go` or `frontend_templates.go` file for the template source. The `types.go` `IntegrationConfig` struct maps tool parameters to template variables.

### Add a new backend language
1. Add a `backend_{lang}.go` file with a template function matching the pattern in existing backend files.
2. Register the new language in `detect_stack_tool.go:DetectStack()` detection logic.
3. Wire the template into `checkout_tool.go:IntegrateRazorpayCheckout()` dispatch.

## Load for Context

- `.agents/skills/repo-skill/domain/checkout-integration.md` — detection logic decisions, Go-defaults-to-gin constraint, toolset visibility rules
