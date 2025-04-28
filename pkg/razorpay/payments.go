package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// PaymentClient defines the interface for payment operations
//
//nolint:iface
type PaymentClient interface {
	Fetch(
		id string,
		data map[string]interface{},
		extraHeaders map[string]string,
	) (map[string]interface{}, error)
}

// FetchPayment returns a tool that fetches payment details using payment_id
func FetchPayment(
	log *slog.Logger,
	paymentClient PaymentClient,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("payment_id is unique identifier "+
				"of the payment to be retrieved."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		paymentID, err := RequiredParam[string](r, "payment_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		payment, err := paymentClient.Fetch(paymentID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payment failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(payment)
	}

	return mcpgo.NewTool(
		"fetch_payment",
		"Use this tool to retrieve the details of a specific payment using its id. Amount returned is in paisa", //nolint:lll
		parameters,
		handler,
	)
}
