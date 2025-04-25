package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"

	rzpsdk "github.com/razorpay/razorpay-go"
)

// FetchPayment returns a tool that fetches payment details using payment_id
func FetchPayment(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			"payment_id is unique identifier of the payment to be retrieved.",
			true,
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		arg, ok := r.Arguments["payment_id"]
		if !ok {
			return mcpgo.NewToolResultError(
				"payment id is a required field"), nil
		}
		id, ok := arg.(string)
		if !ok {
			return mcpgo.NewToolResultError(
				"payment id is expected to be a string"), nil
		}

		payment, err := client.Payment.Fetch(id, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payment failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(payment)
	}

	return mcpgo.NewTool(
		"fetch_payment",
		"fetch payment details using payment id.",
		parameters,
		handler,
	)
}
