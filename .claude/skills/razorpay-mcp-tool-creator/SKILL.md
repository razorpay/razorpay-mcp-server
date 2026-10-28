---
name: razorpay-mcp-tool-creator
description: >
  Creates new MCP tools for the razorpay-mcp-server codebase from a Razorpay
  API curl command or documentation URL. Handles the full lifecycle: parse the
  API spec, generate the Go tool following project conventions, write tests,
  run the linter and fix issues, then open a GitHub PR. Use when the user
  provides a Razorpay API curl snippet or docs link and wants it turned into
  an MCP tool, or says things like "add a tool for X Razorpay API", "implement
  the fetch invoice tool", "create an MCP tool from this curl", etc.
---

# Razorpay MCP Tool Creator

## Repo Context

- **Repo root**: `/Users/jatin.g/go/src/github.com/razorpay/razorpay-mcp-server`
- **Tool files**: `pkg/razorpay/<resource>.go` (e.g. `payments.go`, `orders.go`)
- **Test files**: `pkg/razorpay/<resource>_test.go`
- **Registration**: `pkg/razorpay/tools.go` → `NewToolSets()`
- **Commands**: `make lint`, `make test`, `gh pr create`

## Workflow

### Step 1 — Parse the API source

**If given a curl**, extract: HTTP method, URL path, query/body params and their types.

**If given a docs URL**, fetch the page and extract: endpoint path, HTTP method, each request parameter (name, type, required/optional, description), and the response shape.

Determine:
- **Resource** (e.g. `Payment`, `Refund`) → SDK accessor: `client.Payment`, `client.Refund`
- **Tool name** in `snake_case` (e.g. `fetch_payment`, `create_refund`)
- **Toolset**: payments, orders, refunds, payouts, qr_codes, settlements, payment_links
- **Read or write**: GET/HEAD → read; POST/PATCH/PUT/DELETE → write

Read `references/sdk_and_patterns.md` for SDK method signatures, URL constants, and complete code examples.

### Step 2 — Find or create the target file

Check whether a file for the resource exists:
```
pkg/razorpay/payments.go   → add function here
pkg/razorpay/invoices.go   → create if needed (package razorpay)
```

### Step 3 — Implement the tool function

```go
// FetchInvoice returns a tool to fetch an invoice by ID
func FetchInvoice(
    obs *observability.Observability,
    client *rzpsdk.Client,
) mcpgo.Tool {
    parameters := []mcpgo.ToolParameter{
        mcpgo.WithString("invoice_id",
            mcpgo.Description("Unique identifier of the invoice"),
            mcpgo.Required(),
        ),
    }

    handler := func(ctx context.Context, r mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
        client, err := getClientFromContextOrDefault(ctx, client)
        if err != nil {
            return mcpgo.NewToolResultError(err.Error()), nil
        }

        payload := make(map[string]interface{})
        validator := NewValidator(&r).
            ValidateAndAddRequiredString(payload, "invoice_id")

        if result, err := validator.HandleErrorsIfAny(); result != nil {
            return result, err
        }

        resp, err := client.Invoice.Fetch(payload["invoice_id"].(string), nil, nil)
        if err != nil {
            return mcpgo.NewToolResultError(
                formatErrorMessage("fetching invoice failed", err),
            ), nil
        }
        return mcpgo.NewToolResultJSON(resp)
    }

    return mcpgo.NewTool("fetch_invoice", "Fetch an invoice by its ID.", parameters, handler)
}
```

**Validator cheat-sheet** (chain these on `NewValidator(&r)`):
- `ValidateAndAddRequiredString(payload, "key")` / `ValidateAndAddOptionalString`
- `ValidateAndAddRequiredFloat(payload, "key")` / `ValidateAndAddOptionalFloat`
- `ValidateAndAddRequiredInt(payload, "key")` / `ValidateAndAddOptionalInt`
- `ValidateAndAddRequiredBool(payload, "key")` / `ValidateAndAddOptionalBool`
- `ValidateAndAddRequiredMap(payload, "key")` / `ValidateAndAddOptionalMap`
- `ValidateAndAddOptionalArray(payload, "key")`
- `ValidateAndAddPagination(queryParams)` → adds `count` + `skip`
- `ValidateAndAddExpand(queryParams)` → handles `expand[]`

**Parameter builders** (`mcpgo.With*`): `WithString`, `WithNumber`, `WithBoolean`, `WithObject`, `WithArray`

**Property options**: `Required()`, `Description("...")`, `Min(n)`, `Max(n)`, `Pattern("regex")`, `DefaultValue(v)`, `Enum(vals...)`, `Items(schema)`, `MaxProperties(n)`

### Step 4 — Register in tools.go

```go
// In NewToolSets(), add to the right toolset:
invoices := toolsets.NewToolset("invoices", "Razorpay Invoices related tools").
    AddReadTools(FetchInvoice(obs, client)).
    AddWriteTools(CreateInvoice(obs, client))
toolsetGroup.AddToolset(invoices)

// Or append to an existing toolset:
payments.AddReadTools(FetchPaymentMethods(obs, client))
```

### Step 5 — Write tests

In `pkg/razorpay/<resource>_test.go`, add a `Test_<FuncName>` function using `RazorpayToolTestCase` and `runToolTest`.

**Required test cases:**
1. Success with all parameters
2. Success with required params only (if optional params exist)
3. Missing required parameter → `"missing required parameter: <name>"`
4. Wrong type for a parameter → `"invalid parameter type: <name>"`
5. API error response → error message with the right prefix

```go
func Test_FetchInvoice(t *testing.T) {
    pathFmt := fmt.Sprintf("/%s%s/%%s", constants.VERSION_V1, constants.INVOICE_URL)

    successResp := map[string]interface{}{"id": "inv_xxx", "status": "draft"}
    errorResp   := map[string]interface{}{"error": map[string]interface{}{
        "code": "BAD_REQUEST_ERROR", "description": "invoice not found"}}

    tests := []RazorpayToolTestCase{
        {
            Name:    "successful invoice fetch",
            Request: map[string]interface{}{"invoice_id": "inv_xxx"},
            MockHttpClient: func() (*http.Client, *httptest.Server) {
                return mock.NewHTTPClient(mock.Endpoint{
                    Path: fmt.Sprintf(pathFmt, "inv_xxx"),
                    Method: "GET", Response: successResp,
                })
            },
            ExpectError: false, ExpectedResult: successResp,
        },
        {
            Name: "missing invoice_id",
            Request: map[string]interface{}{},
            MockHttpClient: nil,
            ExpectError: true, ExpectedErrMsg: "missing required parameter: invoice_id",
        },
        {
            Name:    "invoice not found",
            Request: map[string]interface{}{"invoice_id": "inv_bad"},
            MockHttpClient: func() (*http.Client, *httptest.Server) {
                return mock.NewHTTPClient(mock.Endpoint{
                    Path: fmt.Sprintf(pathFmt, "inv_bad"),
                    Method: "GET", Response: errorResp,
                })
            },
            ExpectError: true, ExpectedErrMsg: "fetching invoice failed: invoice not found",
        },
    }
    for _, tc := range tests {
        t.Run(tc.Name, func(t *testing.T) {
            runToolTest(t, tc, FetchInvoice, "Invoice")
        })
    }
}
```

For collection endpoints (no ID in path):
```go
path := fmt.Sprintf("/%s%s", constants.VERSION_V1, constants.INVOICE_URL)
```

### Step 6 — Run tests

```bash
go test -race ./...
```

Fix any failures before continuing.

### Step 7 — Run linter and fix issues

```bash
golangci-lint run ./...
```

Common fixes:
- **line >80 chars**: break string or add `//nolint:lll` at end of the line
- **errcheck**: capture and handle all returned errors
- **unused var/import**: remove it

Repeat until `make lint` exits cleanly.

### Step 8 — Create PR

```bash
git checkout -b feat/add-<tool-name>-tool
git add pkg/razorpay/<resource>.go pkg/razorpay/<resource>_test.go pkg/razorpay/tools.go
git commit -m "feat: add <tool_name> MCP tool for Razorpay <Resource> API"
gh pr create \
  --title "feat: add <tool_name> MCP tool" \
  --body "$(cat <<'EOF'
## Summary
- Adds `<tool_name>` MCP tool wrapping the Razorpay <Resource> <Action> API
- Registered as a <read|write> tool in the `<toolset>` toolset
- Tests cover success, validation errors, and API error cases

## Test plan
- [ ] `go test -race ./...` passes
- [ ] `golangci-lint run ./...` passes
- [ ] Tool appears in MCP server tool list

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

## Reference

See `references/sdk_and_patterns.md` for SDK resource methods, URL constants, and the full test helper signatures.
