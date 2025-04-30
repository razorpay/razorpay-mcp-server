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
				"needs to be refunded. ID should have a pay_ prefix."),
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

		var amount int64
		amount, err = OptionalInt(r, "amount")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		speed, err := OptionalParam[string](r, "speed")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if speed != "" {
			data["speed"] = speed
		}

		notes, err := OptionalParam[map[string]interface{}](r, "notes")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if notes != nil {
			data["notes"] = notes
		}

		receipt, err := OptionalParam[string](r, "receipt")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if receipt != "" {
			data["receipt"] = receipt
		}

		refund, err := client.Payment.Refund(paymentID, int(amount), data, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating refund failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(refund)
	}

	return mcpgo.NewTool(
		"create_refund",
		"Use this tool to create a normal refund for a payment. "+
			"Amount should be in the smallest currency unit (e.g., for ₹295, use 29500)",
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
				"Unique identifier of the refund which is to be retrieved. "+
					"ID should have a rfnd_ prefix."),
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
				"needs to be updated. ID should have a rfnd_ prefix."),
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
