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
		payload := make(map[string]interface{})
		data := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(payload, "payment_id").
			ValidateAndAddRequiredFloat(payload, "amount").
			ValidateAndAddOptionalString(data, "speed").
			ValidateAndAddOptionalString(data, "receipt").
			ValidateAndAddOptionalMap(data, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		refund, err := client.Payment.Refund(
			payload["payment_id"].(string),
			int(payload["amount"].(float64)), data, nil)
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
		payload := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(payload, "refund_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		refund, err := client.Refund.Fetch(payload["refund_id"].(string), nil, nil)
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
		payload := make(map[string]interface{})
		data := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(payload, "refund_id").
			ValidateAndAddRequiredMap(data, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		refund, err := client.Refund.Update(payload["refund_id"].(string), data, nil)
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
