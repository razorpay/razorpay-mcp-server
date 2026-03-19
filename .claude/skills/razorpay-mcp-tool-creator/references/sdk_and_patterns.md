# SDK & Patterns Reference

## Table of Contents
1. [URL Constants](#url-constants)
2. [SDK Client Resources & Methods](#sdk-client-resources--methods)
3. [Tool Imports](#tool-imports)
4. [Toolset → File Mapping](#toolset--file-mapping)
5. [Test Helper Signatures](#test-helper-signatures)
6. [Full Tool Example (Create + Fetch collection)](#full-tool-example)
7. [Linter Suppressions](#linter-suppressions)

---

## URL Constants

From `github.com/razorpay/razorpay-go/constants`:

| Constant            | Value                 |
|---------------------|-----------------------|
| `VERSION_V1`        | `"v1"`                |
| `VERSION_V2`        | `"v2"`                |
| `ORDER_URL`         | `"/orders"`           |
| `INVOICE_URL`       | `"/invoices"`         |
| `PAYMENT_URL`       | `"/payments"`         |
| `PaymentLink_URL`   | `"/payment_links"`    |
| `REFUND_URL`        | `"/refunds"`          |
| `CARD_URL`          | `"/cards"`            |
| `CUSTOMER_URL`      | `"/customers"`        |
| `ADDON_URL`         | `"/addons"`           |
| `TRANSFER_URL`      | `"/transfers"`        |
| `VIRTUAL_ACCOUNT_URL` | `"/virtual_accounts"` |
| `SUBSCRIPTION_URL`  | `"/subscriptions"`    |
| `PLAN_URL`          | `"/plans"`            |
| `QRCODE_URL`        | `"/payments/qr_codes"` |
| `FUND_ACCOUNT_URL`  | `"/fund_accounts"`    |
| `SETTLEMENT_URL`    | `"/settlements"`      |
| `ITEM_URL`          | `"/items"`            |
| `ACCOUNT_URL`       | `"/accounts"`         |
| `STAKEHOLDER_URL`   | `"/stakeholders"`     |
| `PRODUCT_URL`       | `"/products"`         |
| `PAYOUT_URL`        | `"/payouts"`          |

**Mock URL patterns in tests:**

```go
// Collection endpoint (no ID):
path := fmt.Sprintf("/%s%s", constants.VERSION_V1, constants.ORDER_URL)
// → "/v1/orders"

// Single resource endpoint (with ID):
pathFmt := fmt.Sprintf("/%s%s/%%s", constants.VERSION_V1, constants.ORDER_URL)
// In test case: fmt.Sprintf(pathFmt, "order_abc123")
// → "/v1/orders/order_abc123"

// Nested endpoint:
pathFmt := fmt.Sprintf("/%s%s/%%s/payments", constants.VERSION_V1, constants.ORDER_URL)
// → "/v1/orders/order_abc123/payments"
```

---

## SDK Client Resources & Methods

The `*rzpsdk.Client` has these resource accessors. Each resource has methods following the pattern below.

### Common method signatures

```go
// Fetch by ID
client.Order.Fetch(id string, queryParams, headers map[string]interface{}) (map[string]interface{}, error)

// Fetch all / list
client.Order.All(queryParams, headers map[string]interface{}) (map[string]interface{}, error)

// Create
client.Order.Create(payload, headers map[string]interface{}) (map[string]interface{}, error)

// Update
client.Order.Update(id string, payload, headers map[string]interface{}) (map[string]interface{}, error)

// Nested fetch
client.Order.Payments(orderID string, queryParams, headers map[string]interface{}) (map[string]interface{}, error)
```

Pass `nil` for unused `queryParams` and `headers`.

### Resource → client field mapping

| Resource         | Client field          | Example methods                                    |
|------------------|-----------------------|----------------------------------------------------|
| Order            | `client.Order`        | `Fetch`, `All`, `Create`, `Update`, `Payments`     |
| Payment          | `client.Payment`      | `Fetch`, `All`, `Capture`, `Update`, `FetchCardDetails` |
| Refund           | `client.Refund`       | `Fetch`, `All`, `Create`, `Update`, `Payments`     |
| Invoice          | `client.Invoice`      | `Fetch`, `All`, `Create`, `Update`, `Issue`, `Cancel` |
| PaymentLink      | `client.PaymentLink`  | `Fetch`, `All`, `Create`, `Update`, `NotifyBy`     |
| Subscription     | `client.Subscription` | `Fetch`, `All`, `Create`, `Update`, `Cancel`       |
| Customer         | `client.Customer`     | `Fetch`, `All`, `Create`, `Edit`                   |
| Transfer         | `client.Transfer`     | `Fetch`, `All`, `Create`, `Update`                 |
| VirtualAccount   | `client.VirtualAccount` | `Fetch`, `All`, `Create`, `Close`, `Payments`    |
| QrCode           | `client.QrCode`       | `Fetch`, `All`, `Create`, `Close`, `Payments`      |
| Settlement       | `client.Settlement`   | `Fetch`, `All`, `FetchRecon`, `FetchInstant`, `CreateInstant` |
| Payout           | `client.Payout`       | `Fetch`, `All`                                     |
| Plan             | `client.Plan`         | `Fetch`, `All`, `Create`                           |
| Addon            | `client.Addon`        | `Fetch`, `Delete`                                  |
| Card             | `client.Card`         | `Fetch`                                            |
| FundAccount      | `client.FundAccount`  | `Create`, `All`                                    |
| Item             | `client.Item`         | `Fetch`, `All`, `Create`, `Update`, `Delete`       |
| Webhook          | `client.Webhook`      | `Create`, `Fetch`, `All`, `Update`, `Delete`       |

---

## Tool Imports

Always include these imports in a tool file:

```go
package razorpay

import (
    "context"
    "fmt"

    rzpsdk "github.com/razorpay/razorpay-go"

    "github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
    "github.com/razorpay/razorpay-mcp-server/pkg/observability"
)
```

Drop `"fmt"` if not used. Add `"time"` if computing timestamps.

Test file imports:

```go
package razorpay

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/razorpay/razorpay-go/constants"

    "github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)
```

---

## Toolset → File Mapping

| Toolset              | File                          | Existing functions (examples)            |
|----------------------|-------------------------------|------------------------------------------|
| `payments`           | `payments.go`                 | FetchPayment, CapturePayment, UpdatePayment |
| `payment_links`      | `payment_links.go`            | FetchPaymentLink, CreatePaymentLink      |
| `orders`             | `orders.go`                   | FetchOrder, CreateOrder, UpdateOrder     |
| `refunds`            | `refunds.go`                  | FetchRefund, CreateRefund, UpdateRefund  |
| `payouts`            | `payouts.go`                  | FetchPayout, FetchAllPayouts             |
| `qr_codes`           | `qr_codes.go`                 | FetchQRCode, CreateQRCode, CloseQRCode   |
| `settlements`        | `settlements.go`              | FetchSettlement, CreateInstantSettlement |
| `checkout_integration` | `integrations/`             | DetectStack, IntegrateRazorpayCheckout   |
| new resource         | create `<resource>.go`        | —                                        |

---

## Test Helper Signatures

```go
// RazorpayToolTestCase — define test inputs and expectations
type RazorpayToolTestCase struct {
    Name           string
    Request        map[string]interface{}
    MockHttpClient func() (*http.Client, *httptest.Server)  // nil = no HTTP call
    ExpectError    bool
    ExpectedResult map[string]interface{}  // used when ExpectError == false
    ExpectedErrMsg string                  // substring checked when ExpectError == true
}

// runToolTest — executes the test case
func runToolTest(
    t *testing.T,
    tc RazorpayToolTestCase,
    toolCreator func(*observability.Observability, *rzpsdk.Client) mcpgo.Tool,
    objectType string,   // used in diff error messages, e.g. "Order", "Payment"
)

// mock.NewHTTPClient — creates a mock HTTP server
func mock.NewHTTPClient(endpoints ...mock.Endpoint) (*http.Client, *httptest.Server)

type mock.Endpoint struct {
    Path     string       // e.g. "/v1/orders/order_abc"
    Method   string       // "GET", "POST", "PATCH", "DELETE"
    Response interface{}  // map with "error" key → 400; otherwise → 200
}
```

**Error response format** (triggers 400 in mock):
```go
map[string]interface{}{
    "error": map[string]interface{}{
        "code":        "BAD_REQUEST_ERROR",
        "description": "the error message",
    },
}
```

Expected error message in test: `"<prefix>: the error message"` (uses `formatErrorMessage`).

---

## Full Tool Example

### Create + Fetch All (with pagination)

```go
// CreateRefund returns a tool to create a refund for a payment
func CreateRefund(
    obs *observability.Observability,
    client *rzpsdk.Client,
) mcpgo.Tool {
    parameters := []mcpgo.ToolParameter{
        mcpgo.WithString("payment_id",
            mcpgo.Description("ID of the payment to refund"),
            mcpgo.Required(),
        ),
        mcpgo.WithNumber("amount",
            mcpgo.Description("Amount to refund in paise. Omit for full refund."),
            mcpgo.Min(1),
        ),
        mcpgo.WithObject("notes",
            mcpgo.Description("Key-value pairs for additional information"),
            mcpgo.MaxProperties(15),
        ),
    }

    handler := func(ctx context.Context, r mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
        client, err := getClientFromContextOrDefault(ctx, client)
        if err != nil {
            return mcpgo.NewToolResultError(err.Error()), nil
        }

        payload := make(map[string]interface{})
        validator := NewValidator(&r).
            ValidateAndAddRequiredString(payload, "payment_id").
            ValidateAndAddOptionalFloat(payload, "amount").
            ValidateAndAddOptionalMap(payload, "notes")

        if result, err := validator.HandleErrorsIfAny(); result != nil {
            return result, err
        }

        paymentID := payload["payment_id"].(string)
        delete(payload, "payment_id")
        resp, err := client.Payment.Refund(paymentID, payload, nil)
        if err != nil {
            return mcpgo.NewToolResultError(
                formatErrorMessage("creating refund failed", err),
            ), nil
        }
        return mcpgo.NewToolResultJSON(resp)
    }

    return mcpgo.NewTool(
        "create_refund",
        "Create a refund for a specific payment.",
        parameters,
        handler,
    )
}

// FetchAllRefunds returns a tool to list refunds with pagination
func FetchAllRefunds(
    obs *observability.Observability,
    client *rzpsdk.Client,
) mcpgo.Tool {
    parameters := []mcpgo.ToolParameter{
        mcpgo.WithNumber("count",
            mcpgo.Description("Number of refunds to fetch (default 10, max 100)"),
            mcpgo.Min(1), mcpgo.Max(100),
        ),
        mcpgo.WithNumber("skip",
            mcpgo.Description("Number of refunds to skip"),
            mcpgo.Min(0),
        ),
        mcpgo.WithNumber("from",
            mcpgo.Description("Unix timestamp: fetch refunds created after this time"),
            mcpgo.Min(0),
        ),
        mcpgo.WithNumber("to",
            mcpgo.Description("Unix timestamp: fetch refunds created before this time"),
            mcpgo.Min(0),
        ),
    }

    handler := func(ctx context.Context, r mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
        client, err := getClientFromContextOrDefault(ctx, client)
        if err != nil {
            return mcpgo.NewToolResultError(err.Error()), nil
        }

        queryParams := make(map[string]interface{})
        validator := NewValidator(&r).
            ValidateAndAddPagination(queryParams).
            ValidateAndAddOptionalInt(queryParams, "from").
            ValidateAndAddOptionalInt(queryParams, "to")

        if result, err := validator.HandleErrorsIfAny(); result != nil {
            return result, err
        }

        resp, err := client.Refund.All(queryParams, nil)
        if err != nil {
            return mcpgo.NewToolResultError(
                formatErrorMessage("fetching refunds failed", err),
            ), nil
        }
        return mcpgo.NewToolResultJSON(resp)
    }

    return mcpgo.NewTool(
        "fetch_all_refunds",
        "Fetch all refunds with optional filtering and pagination.",
        parameters,
        handler,
    )
}
```

---

## Linter Suppressions

| Issue                     | Fix                                              |
|---------------------------|--------------------------------------------------|
| Line > 80 chars           | `//nolint:lll` at end of the long line           |
| G304 file path            | `//nolint:gosec`                                 |
| gosec string constant     | `//nolint:gosec` after the comparison            |
| errcheck on `w.Write`     | `_, _ = w.Write(...)` or handle the error        |

Golangci config enforces: errcheck, gosimple, govet, staticcheck, gocyclo (max 15), misspell, gofmt, goimports, revive, interfacebloat (max 5 methods per interface), line length 80.
