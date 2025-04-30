package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// CreateQRCode returns a tool that creates QR codes in Razorpay
func CreateQRCode(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"type",
			mcpgo.Description(
				"The type of the QR Code. Currently only supports 'upi_qr'",
			),
			mcpgo.Required(),
			mcpgo.Pattern("^upi_qr$"),
		),
		mcpgo.WithString(
			"name",
			mcpgo.Description(
				"Label to identify the QR Code (e.g., 'Store Front Display')",
			),
		),
		mcpgo.WithString(
			"usage",
			mcpgo.Description(
				"Whether QR should accept single or multiple payments. "+
					"Possible values: 'single_use', 'multiple_use'",
			),
			mcpgo.Required(),
			mcpgo.Pattern("^(single_use|multiple_use)$"),
		),
		mcpgo.WithBoolean(
			"fixed_amount",
			mcpgo.Description(
				"Whether QR should accept only specific amount (true) or any "+
					"amount (false)",
			),
			mcpgo.DefaultValue(false),
		),
		mcpgo.WithNumber(
			"payment_amount",
			mcpgo.Description(
				"The specific amount allowed for transaction in smallest "+
					"currency unit",
			),
			mcpgo.Min(1),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("A brief description about the QR Code"),
		),
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description(
				"The unique identifier of the customer to link with the QR Code",
			),
		),
		mcpgo.WithNumber(
			"close_by",
			mcpgo.Description(
				"Unix timestamp at which QR Code should be automatically "+
					"closed (min 2 mins after current time)",
			),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description(
				"Key-value pairs for additional information "+
					"(max 15 pairs, 256 chars each)",
			),
			mcpgo.MaxProperties(15),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Validate required parameters
		qrType, err := RequiredParam[string](r, "type")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		usage, err := RequiredParam[string](r, "usage")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Create request payload
		qrData := map[string]interface{}{
			"type":  qrType,
			"usage": usage,
		}

		// Add optional parameters if provided
		name, err := OptionalParam[string](r, "name")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if name != "" {
			qrData["name"] = name
		}

		fixedAmount, err := OptionalParam[bool](r, "fixed_amount")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		qrData["fixed_amount"] = fixedAmount

		paymentAmount, err := OptionalInt(r, "payment_amount")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if paymentAmount > 0 {
			qrData["payment_amount"] = paymentAmount
		}

		description, err := OptionalParam[string](r, "description")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if description != "" {
			qrData["description"] = description
		}

		customerID, err := OptionalParam[string](r, "customer_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if customerID != "" {
			qrData["customer_id"] = customerID
		}

		closeBy, err := OptionalInt(r, "close_by")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if closeBy > 0 {
			qrData["close_by"] = closeBy
		}

		notes, err := OptionalParam[map[string]interface{}](r, "notes")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if notes != nil {
			qrData["notes"] = notes
		}

		// Create QR code using Razorpay SDK
		qrCode, err := client.QrCode.Create(qrData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating QR code failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(qrCode)
	}

	return mcpgo.NewTool(
		"create_qr_code",
		"Create a new QR code in Razorpay that can be used to accept UPI payments",
		parameters,
		handler,
	)
}
