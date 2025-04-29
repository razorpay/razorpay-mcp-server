package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// FetchPayment returns a tool that fetches payment details using payment_id
func FetchPayment(
	log *slog.Logger,
	client *rzpsdk.Client,
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

		payment, err := client.Payment.Fetch(paymentID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payment failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(payment)
	}

	return mcpgo.NewTool(
		"fetch_payment",
		"Use this tool to retrieve the details of a specific payment using its id. "+
			"Amount returned is in paisa",
		parameters,
		handler,
	)
}

// FetchMultipleRefundsForPayment returns a tool that fetches multiple refunds
// for a payment
func FetchMultipleRefundsForPayment(
	_ *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment for which "+
				"refunds are to be retrieved."),
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"from",
			mcpgo.Description("Unix timestamp at which the refunds were created."),
		),
		mcpgo.WithNumber(
			"to",
			mcpgo.Description("Unix timestamp till which the refunds were created."),
		),
		mcpgo.WithNumber(
			"count",
			mcpgo.Description("The number of refunds to fetch for the payment."),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description("The number of refunds to be skipped for the payment."),
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

		queryParams := make(map[string]interface{})

		from, err := OptionalInt(r, "from")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if from > 0 {
			queryParams["from"] = from
		}

		to, err := OptionalInt(r, "to")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if to > 0 {
			queryParams["to"] = to
		}

		count, err := OptionalInt(r, "count")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if count > 0 {
			queryParams["count"] = count
		}

		skip, err := OptionalInt(r, "skip")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if skip > 0 {
			queryParams["skip"] = skip
		}

		refunds, err := client.Payment.FetchMultipleRefund(
			paymentID, queryParams, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching multiple refunds failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(refunds)
	}

	return mcpgo.NewTool(
		"fetch_multiple_refunds_for_payment",
		"Use this tool to retrieve multiple refunds for a payment. "+
			"By default, only the last 10 refunds are returned.",
		parameters,
		handler,
	)
}

// FetchSpecificRefundForPayment returns a tool that fetches a specific refund
// for a payment
func FetchSpecificRefundForPayment(
	_ *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment for which "+
				"the refund has been made."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"refund_id",
			mcpgo.Description("Unique identifier of the refund to be retrieved."),
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

		refundID, err := RequiredParam[string](r, "refund_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		refund, err := client.Payment.FetchRefund(paymentID, refundID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching specific refund for payment failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(refund)
	}

	return mcpgo.NewTool(
		"fetch_specific_refund_for_payment",
		"Use this tool to retrieve details of a specific refund made for a payment.",
		parameters,
		handler,
	)
}
