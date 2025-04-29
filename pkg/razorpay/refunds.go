package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// CreateRefund returns a tool that creates a normal refund for a payment
func CreateRefund(
	_ *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment which "+
				"needs to be refunded."),
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Payment amount in the smallest currency unit "+
				"(e.g., for ₹295, use 29500)"),
		),
		mcpgo.WithString(
			"speed",
			mcpgo.Description("The speed at which the refund is to be "+
				"processed. Default is 'normal'. For instant refunds, speed "+
				"is set as 'optimum'."),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs used to store additional "+
				"information. A maximum of 15 key-value pairs can be included."),
		),
		mcpgo.WithString(
			"receipt",
			mcpgo.Description("A unique identifier provided by you for "+
				"your internal reference."),
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

		data := make(map[string]interface{})

		var amount int
		if amountFloat, err := OptionalParam[float64](r, "amount"); err == nil {
			amount = int(amountFloat)
		}

		if speed, err := OptionalParam[string](r, "speed"); err == nil {
			data["speed"] = speed
		}

		notesType := OptionalParam[map[string]interface{}]
		if notes, err := notesType(r, "notes"); err == nil {
			data["notes"] = notes
		}

		if receipt, err := OptionalParam[string](r, "receipt"); err == nil {
			data["receipt"] = receipt
		}

		refund, err := client.Payment.Refund(paymentID, amount, data, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating refund failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(refund)
	}

	return mcpgo.NewTool(
		"create_refund",
		"Use this tool to create a normal refund for a payment. "+
			"Amount should be in the smallest currency unit "+
			"(e.g., for ₹295, use 29500)",
		parameters,
		handler,
	)
}

// FetchRefund returns a tool that fetches a refund by ID
func FetchRefund(
	_ *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"refund_id",
			mcpgo.Description(
				"Unique identifier of the refund which is to be retrieved."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		refundID, err := RequiredParam[string](r, "refund_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		refund, err := client.Refund.Fetch(refundID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching refund failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(refund)
	}

	return mcpgo.NewTool(
		"fetch_refund",
		"Use this tool to retrieve the details of a specific refund using its id.",
		parameters,
		handler,
	)
}

// UpdateRefund returns a tool that updates a refund's notes
func UpdateRefund(
	_ *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"refund_id",
			mcpgo.Description("Unique identifier of the refund which "+
				"needs to be updated."),
			mcpgo.Required(),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs used to store additional "+
				"information. A maximum of 15 key-value pairs can be included, "+
				"with each value not exceeding 256 characters."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		refundID, err := RequiredParam[string](r, "refund_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		notesType := RequiredParam[map[string]interface{}]
		notes, err := notesType(r, "notes")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		data := make(map[string]interface{})
		data["notes"] = notes

		refund, err := client.Refund.Update(refundID, data, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("updating refund failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(refund)
	}

	return mcpgo.NewTool(
		"update_refund",
		"Use this tool to update the notes for a specific refund. "+
			"Only the notes field can be modified.",
		parameters,
		handler,
	)
}

// FetchAllRefunds returns a tool that fetches all refunds with pagination
// support
func FetchAllRefunds(
	_ *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"from",
			mcpgo.Description("Unix timestamp at which the refunds were created"),
		),
		mcpgo.WithNumber(
			"to",
			mcpgo.Description("Unix timestamp till which the refunds were created"),
		),
		mcpgo.WithNumber(
			"count",
			mcpgo.Description("The number of refunds to fetch. "+
				"You can fetch a maximum of 100 refunds"),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description("The number of refunds to be skipped"),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		queryParams := make(map[string]interface{})

		if fromFloat, err := OptionalParam[float64](r, "from"); err == nil {
			queryParams["from"] = int(fromFloat)
		}

		if toFloat, err := OptionalParam[float64](r, "to"); err == nil {
			queryParams["to"] = int(toFloat)
		}

		if countFloat, err := OptionalParam[float64](r, "count"); err == nil {
			queryParams["count"] = int(countFloat)
		}

		if skipFloat, err := OptionalParam[float64](r, "skip"); err == nil {
			queryParams["skip"] = int(skipFloat)
		}

		refunds, err := client.Refund.All(queryParams, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching refunds failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(refunds)
	}

	return mcpgo.NewTool(
		"fetch_all_refunds",
		"Use this tool to retrieve details of all refunds. "+
			"By default, only the last 10 refunds are returned.",
		parameters,
		handler,
	)
}
