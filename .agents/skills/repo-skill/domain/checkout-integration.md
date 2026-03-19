# Checkout Integration Toolset

A developer-assist toolset that generates Razorpay Standard Checkout integration code for a target application. It does NOT call Razorpay APIs at runtime and has no side effects — all output is locally generated code.

## Decisions

### D1: Separate toolset from payment operations

**Context:** The MCP server exposes both operational tools (create order, capture payment) and developer-assist tools (generate integration code). Mixing them into the same toolset would expose code-generation tools during payment workflows and vice versa.
**Decision:** `checkout_integration` is a distinct named toolset, registered separately in `tools.go:NewToolSets()`, so it can be enabled or disabled independently via `--toolsets`.
**Alternatives considered:**
- **Merge into payments toolset:** Rejected because the use case is entirely different — one is runtime payment execution, the other is one-time developer onboarding.
**Trade-offs:**
- Gained: operators can expose only code-generation tools to developer-facing agents without exposing payment write operations.
- Lost: developers must remember to include `checkout_integration` explicitly when using `--toolsets`.
**Code:** `tools.go:NewToolSets()`
**Revisit if:** The toolset grows to include runtime API calls, making the read-only classification inaccurate.

### D2: detect_stack must run before integrate_razorpay_checkout

**Context:** `integrate_razorpay_checkout` branches on `language`, `backendFramework`, and `frontendFramework` to select among 20+ code templates. These values cannot be defaulted meaningfully — a wrong framework generates unusable code.
**Decision:** The tool description for `integrate_razorpay_checkout` hard-codes the instruction "ALWAYS call detect_stack first" and "Do NOT ask the user for these values." `detect_stack` resolves them from file and manifest hints provided by the calling agent.
**Alternatives considered:**
- **Ask user for language/framework:** Rejected because this breaks the zero-friction integration goal; users shouldn't need to know framework names.
- **Auto-detect inside integrate_razorpay_checkout:** Rejected because the checkout tool receives no filesystem access — detection requires an agent-side file listing step.
**Trade-offs:**
- Gained: agent always has accurate stack info; no user friction.
- Lost: two-tool round trip required; `detect_stack` must be called first every time.
**Code:** `checkout_tool.go:IntegrateRazorpayCheckout()`, `detect_stack_tool.go:DetectStack()`
**Revisit if:** MCP adds filesystem resource access, allowing single-tool detection and generation.

### D3: Both tools registered as read-only

**Context:** Both tools generate code locally and make no external API calls or mutations.
**Decision:** Both `IntegrateRazorpayCheckout` and `DetectStack` are added via `AddReadTools()` in `tools.go:NewToolSets()`. The toolset framework marks read tools with `SetReadOnly(true)` and always registers them regardless of `--read-only` mode.
**Trade-offs:**
- Gained: tools remain available in locked-down read-only deployments.
- Lost: none — there is no write behavior to gate.
**Code:** `toolsets.go:RegisterTools()`, `tools.go:NewToolSets()`

## Non-Obvious Constraints

- **`integrate_razorpay_checkout` returns complete, apply-ready code** — backend route(s), frontend JS/HTML, and payment verification in one response. The AI instruction in the tool description says to apply all returned files without prompting the user for additional steps. Agents must not present code as a preview.
- **`detect_stack` does not scan the filesystem autonomously** — it receives a `files` list and dependency manifest contents that the calling agent must read and pass in. Passing an empty `files` list returns `language: "unknown"` with `confidence: 0.1` — the integration will fail.
- **`checkout_integration` toolset is opt-out, not opt-in, when `--toolsets` is omitted** — passing an empty `enabledToolsets` slice triggers `everythingOn` in `toolsets.go:EnableToolsets()`, enabling all toolsets including this one. Passing any explicit toolset list disables `checkout_integration` unless it is named explicitly.
- **Go detection defaults to `gin`** — if `go.mod` is present but contains neither echo nor fiber imports, `detectProjectStack()` returns `framework: "gin"` regardless of actual router. Verify before applying generated code.

## Flow Map

### Detect Stack and Generate Integration Code

| Flow | Trigger | Key Functions | Decision Points | Outcome |
|------|---------|---------------|-----------------|---------|
| Happy path (most traffic) | Agent lists project files and reads dependency manifest | `detect_stack_tool.go:DetectStack()` -> `detect_stack_tool.go:detectProjectStack()` -> `checkout_tool.go:IntegrateRazorpayCheckout()` | DP1: Stack identified with high confidence | Complete integration code returned for detected framework |
| Unknown stack (rare) | Agent passes empty or unrecognized file list | `detect_stack_tool.go:detectProjectStack()` | DP1: No manifest matched | Returns `language: "unknown"`, confidence 0.1 — integration call will produce express/vanilla fallback |
| Mobile app (common for React Native / Flutter) | pubspec.yaml or react-native dep detected | `detect_stack_tool.go:detectProjectStack()` | DP2: Mobile path | `IsFullStack: false` returned; `integrate_razorpay_checkout` uses mobile-specific template (`mobile.go`) |

**Decision Points:**
- **DP1: Manifest priority order** — `detectProjectStack()` checks pubspec.yaml -> go.mod -> Cargo.toml -> pom.xml -> composer.json -> Gemfile -> .csproj -> requirements.txt -> package.json in that order and returns on first match. A project with both go.mod and package.json will always be detected as Go. Why: language specificity reduces ambiguity; Go projects rarely need Node frontend code generation.
- **DP2: React Native early return** — if `react-native` or `expo` is found in package.json deps, `detectProjectStack()` returns immediately with `Framework: "react-native"` and `IsFullStack: false` before any backend framework detection. Why: React Native projects have no separate backend route to generate.

## When to Use vs. Direct API Tools

| Goal | Use |
|------|-----|
| Help a developer add Razorpay checkout to their app | `detect_stack` then `integrate_razorpay_checkout` |
| Create a payment order at runtime | `orders` toolset — `CreateOrder()` |
| Capture or fetch payment details | `payments` toolset — `CapturePayment()`, `FetchPayment()` |
| Issue a refund | `refunds` toolset — `CreateRefund()` |
