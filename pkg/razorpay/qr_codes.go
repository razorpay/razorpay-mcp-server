package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// CreateQRCode returns a tool that creates QR codes in Razorpay
func CreateQRCode(
	obs *observability.Observability,
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
			mcpgo.Enum("single_use", "multiple_use"),
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
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		qrData := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(qrData, "type").
			ValidateAndAddRequiredString(qrData, "usage").
			ValidateAndAddOptionalString(qrData, "name").
			ValidateAndAddOptionalBool(qrData, "fixed_amount").
			ValidateAndAddOptionalFloat(qrData, "payment_amount").
			ValidateAndAddOptionalString(qrData, "description").
			ValidateAndAddOptionalString(qrData, "customer_id").
			ValidateAndAddOptionalFloat(qrData, "close_by").
			ValidateAndAddOptionalMap(qrData, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Check if fixed_amount is true, then payment_amount is required
		if fixedAmount, exists := qrData["fixed_amount"]; exists &&
			fixedAmount.(bool) {
			if _, exists := qrData["payment_amount"]; !exists {
				return mcpgo.NewToolResultError(
					"payment_amount is required when fixed_amount is true"), nil
			}
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

// FetchQRCode returns a tool that fetches a specific QR code by ID
func FetchQRCode(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"qr_code_id",
			mcpgo.Description(
				"Unique identifier of the QR Code to be retrieved"+
					"The QR code id should start with 'qr_'",
			),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})
		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "qr_code_id")
		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}
		qrCodeID := params["qr_code_id"].(string)

		// Fetch QR code by ID using Razorpay SDK
		qrCode, err := client.QrCode.Fetch(qrCodeID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching QR code failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(qrCode)
	}

	return mcpgo.NewTool(
		"fetch_qr_code",
		"Fetch a QR code's details using it's ID",
		parameters,
		handler,
	)
}

// FetchAllQRCodes returns a tool that fetches all QR codes
// with pagination support
func FetchAllQRCodes(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"from",
			mcpgo.Description(
				"Unix timestamp, in seconds, from when QR Codes are to be retrieved",
			),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"to",
			mcpgo.Description(
				"Unix timestamp, in seconds, till when QR Codes are to be retrieved",
			),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"count",
			mcpgo.Description(
				"Number of QR Codes to be retrieved (default: 10, max: 100)",
			),
			mcpgo.Min(1),
			mcpgo.Max(100),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description(
				"Number of QR Codes to be skipped (default: 0)",
			),
			mcpgo.Min(0),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		fetchQROptions := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddOptionalInt(fetchQROptions, "from").
			ValidateAndAddOptionalInt(fetchQROptions, "to").
			ValidateAndAddPagination(fetchQROptions)

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Fetch QR codes using Razorpay SDK
		qrCodes, err := client.QrCode.All(fetchQROptions, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching QR codes failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(qrCodes)
	}

	return mcpgo.NewTool(
		"fetch_all_qr_codes",
		"Fetch all QR codes with optional filtering and pagination",
		parameters,
		handler,
	)
}

// FetchQRCodesByCustomerID returns a tool that fetches QR codes
// for a specific customer ID
func FetchQRCodesByCustomerID(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description(
				"The unique identifier of the customer",
			),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		fetchQROptions := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(fetchQROptions, "customer_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Fetch QR codes by customer ID using Razorpay SDK
		qrCodes, err := client.QrCode.All(fetchQROptions, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching QR codes failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(qrCodes)
	}

	return mcpgo.NewTool(
		"fetch_qr_codes_by_customer_id",
		"Fetch all QR codes for a specific customer",
		parameters,
		handler,
	)
}

// FetchQRCodesByPaymentID returns a tool that fetches QR codes
// for a specific payment ID
func FetchQRCodesByPaymentID(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description(
				"The unique identifier of the payment"+
					"The payment id always should start with 'pay_'",
			),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		fetchQROptions := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(fetchQROptions, "payment_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Fetch QR codes by payment ID using Razorpay SDK
		qrCodes, err := client.QrCode.All(fetchQROptions, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching QR codes failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(qrCodes)
	}

	return mcpgo.NewTool(
		"fetch_qr_codes_by_payment_id",
		"Fetch all QR codes for a specific payment",
		parameters,
		handler,
	)
}

// FetchPaymentsForQRCode returns a tool that fetches payments made on a QR code
func FetchPaymentsForQRCode(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"qr_code_id",
			mcpgo.Description(
				"The unique identifier of the QR Code to fetch payments for"+
					"The QR code id should start with 'qr_'",
			),
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"from",
			mcpgo.Description(
				"Unix timestamp, in seconds, from when payments are to be retrieved",
			),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"to",
			mcpgo.Description(
				"Unix timestamp, in seconds, till when payments are to be fetched",
			),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"count",
			mcpgo.Description(
				"Number of payments to be fetched (default: 10, max: 100)",
			),
			mcpgo.Min(1),
			mcpgo.Max(100),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description(
				"Number of records to be skipped while fetching the payments",
			),
			mcpgo.Min(0),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})
		fetchQROptions := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "qr_code_id").
			ValidateAndAddOptionalInt(fetchQROptions, "from").
			ValidateAndAddOptionalInt(fetchQROptions, "to").
			ValidateAndAddOptionalInt(fetchQROptions, "count").
			ValidateAndAddOptionalInt(fetchQROptions, "skip")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		qrCodeID := params["qr_code_id"].(string)

		// Fetch payments for QR code using Razorpay SDK
		payments, err := client.QrCode.FetchPayments(qrCodeID, fetchQROptions, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payments for QR code failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(payments)
	}

	return mcpgo.NewTool(
		"fetch_payments_for_qr_code",
		"Fetch all payments made on a QR code",
		parameters,
		handler,
	)
}

// CloseQRCode returns a tool that closes a specific QR code
func CloseQRCode(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"qr_code_id",
			mcpgo.Description(
				"Unique identifier of the QR Code to be closed"+
					"The QR code id should start with 'qr_'",
			),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})
		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "qr_code_id")
		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}
		qrCodeID := params["qr_code_id"].(string)

		// Close QR code by ID using Razorpay SDK
		qrCode, err := client.QrCode.Close(qrCodeID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("closing QR code failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(qrCode)
	}

	return mcpgo.NewTool(
		"close_qr_code",
		"Close a QR Code that's no longer needed",
		parameters,
		handler,
	)
}
