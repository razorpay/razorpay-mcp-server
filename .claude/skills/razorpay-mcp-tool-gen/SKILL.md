---
name: razorpay-mcp-tool-gen
description: >-
  Generate Razorpay MCP server tool implementations from API documentation URLs,
  curl commands, or request/response examples. Produces Go tool code, unit tests,
  registration, and README updates. Use when the user provides a Razorpay API doc
  link, curl example, or request/response payload and wants a new MCP tool created.
---

# Razorpay MCP Tool Generator

Generate complete MCP tool implementations from Razorpay API docs, curl commands, or request/response examples.

## Input Formats

The user may provide any of:

1. **Razorpay docs URL** — `https://razorpay.com/docs/api/...`
2. **curl command** — extract method, path, headers, body, and response
3. **Request/response JSON** — raw payload and expected response
4. **SDK function signature** — Go SDK function to call

If the **tool name** is not provided or cannot be inferred, **ask the user** before proceeding.

## Workflow

```
Task Progress:
- [ ] Step 1: Parse input and extract API contract
- [ ] Step 2: Determine tool name and resource file
- [ ] Step 3: Implement tool function
- [ ] Step 4: Register tool in tools.go
- [ ] Step 5: Write unit tests
- [ ] Step 6: Update README.md Available Tools table
- [ ] Step 7: Run linter — fix errors
- [ ] Step 8: Run tests — fix errors
- [ ] Step 9: Create branch, commit, and open PR (if gh available)
```

### Step 1: Parse Input

**From docs URL**: Fetch the page and extract endpoint path, HTTP method, required/optional params with types and descriptions, and example response.

**From curl**: Parse the method (`-X`), URL path, headers (`-H`), request body (`-d`), and note the response if provided.

**From request/response JSON**: Infer required vs optional fields, types, and construct parameter definitions.

Map each parameter to a validator type:

| JSON type | Parameter helper | Validator method |
|-----------|-----------------|------------------|
| string | `mcpgo.WithString` | `ValidateAndAddRequiredString` / `ValidateAndAddOptionalString` |
| number (int) | `mcpgo.WithNumber` | `ValidateAndAddRequiredInt` / `ValidateAndAddOptionalInt` |
| number (float) | `mcpgo.WithNumber` | `ValidateAndAddRequiredFloat` / `ValidateAndAddOptionalFloat` |
| boolean | `mcpgo.WithBoolean` | `ValidateAndAddRequiredBool` / `ValidateAndAddOptionalBool` |
| object | `mcpgo.WithObject` | `ValidateAndAddRequiredMap` / `ValidateAndAddOptionalMap` |
| array | `mcpgo.WithArray` + **`mcpgo.Items(...)`** (required) | `ValidateAndAddRequiredArray` / `ValidateAndAddOptionalArray` |

**Array parameters MUST include `mcpgo.Items(...)`** with the element schema.
Without it the published MCP schema is invalid and clients cannot use the tool.

```go
mcpgo.WithArray(
    "tags",
    mcpgo.Description("List of tag strings"),
    mcpgo.Items(map[string]interface{}{"type": "string"}),
)
```

For nested objects (e.g., `customer.name` flattened to `customer_name`), use `ValidateAndAddOptionalStringToPath`.

### Step 2: Determine Tool Name and File

Naming conventions:
- Fetch single: `fetch_{resource}` → `Fetch{Resource}`
- Fetch list: `fetch_all_{resources}` → `FetchAll{Resources}`
- Create: `create_{resource}` → `Create{Resource}`
- Update: `update_{resource}` → `Update{Resource}`

Place the tool in `pkg/razorpay/{resource_type}.go`. Create a new file only if the resource type doesn't already exist.

### Step 3: Implement Tool

Follow this exact structure:

```go
func ToolName(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		// Required params first, then optional
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		payload := make(map[string]interface{})
		validator := NewValidator(&r).
			ValidateAndAddRequiredString(payload, "id")
			// chain more validators...

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		response, err := client.Resource.Method(/* args */)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("operation failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool("tool_name", "description", parameters, handler)
}
```

#### Writing LLM-Friendly Tool Descriptions

The description string in `mcpgo.NewTool` is what LLMs read to decide which tool to call. A bad description means the tool gets ignored or misused. Every description **must** answer three questions:

1. **What** does this tool do?
2. **When** should an LLM pick this tool? (trigger conditions, prerequisites)
3. **What** constraints or gotchas should the LLM know? (units, required states, return format)

**Structure** (2-4 sentences):

```
[Action verb] + [what it does] + [key context].
[When to use / prerequisites]. [Constraints, units, or return format].
```

**Bad examples** (too vague, no context):

```go
// Vague — LLM doesn't know when to pick this
"Fetch an order's details using its ID"

// No constraints — LLM won't know about paise
"Create a new standard payment link in Razorpay with a specified amount"

// Missing trigger context
"Fetch all QR codes with optional filtering and pagination"
```

**Good examples** (actionable, LLM knows when/why):

```go
// Clear what, when, and constraints
"Retrieve details of a specific payment by its ID. " +
	"Use when you need payment status, method, or amount. " +
	"Amount returned is in paise (100 paise = ₹1)."

// Prerequisites + constraints
"Capture a previously authorized payment to settle funds. " +
	"Only works on payments with status 'authorized'. " +
	"Amount in paise; partial capture supported."

// Trigger context + return format
"Get all saved payment methods (cards, UPI) " +
	"for a contact number. " +
	"Use when the user wants to pay with a saved method. " +
	"Returns token IDs that can be used with initiate_payment."

// When to use + what makes it different
"Create a UPI-specific payment link (generates a UPI QR). " +
	"Use instead of create_payment_link when the customer " +
	"will pay via UPI. Only supports INR currency."
```

**Rules:**
- Start with an action verb (Fetch, Create, Update, Capture, Revoke)
- Mention units if amounts are involved (`paise`, `smallest currency unit`)
- Mention prerequisite states (`status must be 'authorized'`)
- If similar tools exist, explain when to pick THIS one over others
- If the tool returns data used by other tools, say which ones
- Keep it under 3-4 sentences; break long lines with `+` concatenation

Required imports:

```go
import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)
```

Parameter options: `mcpgo.Required()`, `mcpgo.Description(...)`, `mcpgo.Min(n)`, `mcpgo.Max(n)`, `mcpgo.Pattern(re)`, `mcpgo.Enum(values...)`, `mcpgo.DefaultValue(v)`.

### Step 4: Register Tool

In `pkg/razorpay/tools.go`, add to the appropriate toolset:
- Read-only tools → `AddReadTools(...)`
- Mutating tools → `AddWriteTools(...)`

If the resource type is new, create a new toolset:

```go
newResource := toolsets.NewToolset("resource_name", "Description").
	AddReadTools(FetchResource(obs, client)).
	AddWriteTools(CreateResource(obs, client))
toolsetGroup.AddToolset(newResource)
```

### Step 5: Write Unit Tests

Create or append to `pkg/razorpay/{resource_type}_test.go`.

Required imports:

```go
import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)
```

Mock path construction using SDK constants:

```go
// Base path (POST/list endpoints)
basePath := fmt.Sprintf("/%s%s",
	constants.VERSION_V1, constants.RESOURCE_URL)

// Path with ID (GET/PATCH single resource)
pathFmt := fmt.Sprintf("/%s%s/%%s",
	constants.VERSION_V1, constants.RESOURCE_URL)
```

**CRITICAL**: Never add query parameters to mock paths. The mock server uses Gorilla Mux and does not validate query params.

Required test cases:

| Case | MockHttpClient | ExpectError | Notes |
|------|---------------|-------------|-------|
| Success with all params | Returns success response | `false` | Use docs example payload/response |
| Missing required param (one per required param) | `nil` | `true` | `ExpectedErrMsg: "missing required parameter: X"` |
| Multiple validation errors | `nil` | `true` | Missing + wrong types |
| API error | Returns error response | `true` | Test SDK error handling |

Test structure:

```go
func Test_ToolName(t *testing.T) {
	apiPath := fmt.Sprintf("/%s%s",
		constants.VERSION_V1, constants.RESOURCE_URL)

	successResponse := map[string]interface{}{
		// from docs
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "error message",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name:    "successful operation",
			Request: map[string]interface{}{/* all params */},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path: apiPath, Method: "POST",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		// ... negative cases
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, ToolName, "Resource")
		})
	}
}
```

### Step 6: Update README

Add a row to the Available Tools table in the root `README.md`:

```
| `tool_name` | Brief description | [Resource](docs_url) | ✅ or ❌ |
```

### Step 7: Run Linter

```bash
golangci-lint run ./pkg/razorpay/...
```

Key rules to watch:
- **lll**: 80 char max line length. Use `// nolint:lll` or string concatenation for long lines.
- **goimports**: Local prefix `github.com/razorpay/razorpay-mcp-server`. Group: stdlib → external → local.
- **errcheck**, **bodyclose**, **gosec**: Handle all errors, close response bodies.

Fix all errors. Retry up to 2 times.

### Step 8: Run Tests

```bash
go test ./pkg/razorpay/... -v -count=1
```

All tests must pass. Fix failures and retry up to 2 times.

### Step 9: Create Branch, Commit, and Open PR

This step runs **only if `gh` CLI is available**. Check with `gh --version`. If unavailable, skip and inform the user.

**Branch naming** (per CONTRIBUTING.md): `{username}/feat-{tool_name}`

Detect the GitHub username:

```bash
gh api user --jq '.login'
```

**Full flow:**

```bash
# 1. Ensure we're on a clean main
git checkout main
git pull origin main

# 2. Create feature branch
git checkout -b USERNAME/feat-TOOL_NAME

# 3. Stage all changes
git add pkg/razorpay/{resource}.go \
       pkg/razorpay/{resource}_test.go \
       pkg/razorpay/tools.go \
       README.md

# 4. Commit with conventional format
git commit -m "feat: add TOOL_NAME tool"

# 5. Push and create PR
git push -u origin HEAD
gh pr create \
  --title "feat: add TOOL_NAME tool" \
  --body "$(cat <<'EOF'
## Summary
- Add `tool_name` tool for [brief description]
- Implements [HTTP method] [endpoint path]
- Includes unit tests with full coverage

## Test plan
- [ ] `go test ./pkg/razorpay/... -v -run Test_ToolName`
- [ ] `golangci-lint run ./pkg/razorpay/...`
- [ ] Manual verification with MCP client (optional)

## References
- API docs: [link]
EOF
)"
```

**Commit message format** (from CONTRIBUTING.md):
- `feat:` — new tool
- `fix:` — bug fix
- `ref:` — refactoring
- `test:` — test-only changes
- `chore:` — config, CI, review comments

After creating the PR, **return the PR URL** to the user.

If the repo is a fork, `gh` will automatically offer to create the PR against the upstream `razorpay/razorpay-mcp-server` repository.

## SDK Constants Reference

From `github.com/razorpay/razorpay-go/constants`:

```
VERSION_V1       = "v1"
ORDER_URL        = "/orders"
PAYMENT_URL      = "/payments"
PaymentLink_URL  = "/payment_links"
REFUND_URL       = "/refunds"
CUSTOMER_URL     = "/customers"
QRCODE_URL       = "/payments/qr_codes"
SETTLEMENT_URL   = "/settlements"
PAYOUT_URL       = "/payouts"
```

If the constant doesn't exist for your resource, check the SDK source or use a string literal.

## Completion Checklist

Include this at the end of every implementation:

- [ ] Tool function implemented in `pkg/razorpay/{resource}.go`
- [ ] Tool registered in `pkg/razorpay/tools.go`
- [ ] Unit tests in `pkg/razorpay/{resource}_test.go` (success + all negative cases)
- [ ] README.md Available Tools table updated
- [ ] Linter passes (`golangci-lint run`)
- [ ] Tests pass (`go test ./...`)
- [ ] Branch created, committed, PR opened (if `gh` available) — link returned to user
