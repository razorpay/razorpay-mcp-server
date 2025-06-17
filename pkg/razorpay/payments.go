package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// FetchPayment returns a tool that fetches payment details using payment_id
func FetchPayment(
	obs *observability.Observability,
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
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "payment_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		paymentId := params["payment_id"].(string)

		payment, err := client.Payment.Fetch(paymentId, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payment failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(payment)
	}

	return mcpgo.NewTool(
		"fetch_payment",
		"Use this tool to retrieve the details of a specific payment "+
			"using its id. Amount returned is in paisa",
		parameters,
		handler,
	)
}

// FetchPaymentCardDetails returns a tool that fetches card details
// for a payment
func FetchPaymentCardDetails(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment for which "+
				"you want to retrieve card details. Must start with 'pay_'"),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "payment_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		paymentId := params["payment_id"].(string)

		cardDetails, err := client.Payment.FetchCardDetails(
			paymentId, nil, nil)

		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching card details failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(cardDetails)
	}

	return mcpgo.NewTool(
		"fetch_payment_card_details",
		"Use this tool to retrieve the details of the card used to make a payment. "+
			"Only works for payments made using a card.",
		parameters,
		handler,
	)
}

// UpdatePayment returns a tool that updates the notes for a payment
func UpdatePayment(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment to be updated. "+
				"Must start with 'pay_'"),
			mcpgo.Required(),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs that can be used to store additional "+
				"information about the payment. Values must be strings or integers."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})
		paymentUpdateReq := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "payment_id").
			ValidateAndAddRequiredMap(paymentUpdateReq, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		paymentId := params["payment_id"].(string)

		// Update the payment
		updatedPayment, err := client.Payment.Edit(paymentId, paymentUpdateReq, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("updating payment failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(updatedPayment)
	}

	return mcpgo.NewTool(
		"update_payment",
		"Use this tool to update the notes field of a payment. Notes are "+
			"key-value pairs that can be used to store additional information.", //nolint:lll
		parameters,
		handler,
	)
}

// CapturePayment returns a tool that captures an authorized payment
func CapturePayment(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment to be captured. Should start with 'pay_'"), //nolint:lll
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("The amount to be captured (in paisa). "+
				"Should be equal to the authorized amount"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("ISO code of the currency in which the payment "+
				"was made (e.g., INR)"),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})
		paymentCaptureReq := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "payment_id").
			ValidateAndAddRequiredInt(params, "amount").
			ValidateAndAddRequiredString(paymentCaptureReq, "currency")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		paymentId := params["payment_id"].(string)
		amount := int(params["amount"].(int64))

		// Capture the payment
		payment, err := client.Payment.Capture(
			paymentId,
			amount,
			paymentCaptureReq,
			nil,
		)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("capturing payment failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(payment)
	}

	return mcpgo.NewTool(
		"capture_payment",
		"Use this tool to capture a previously authorized payment. Only payments with 'authorized' status can be captured", //nolint:lll
		parameters,
		handler,
	)
}

// FetchAllPayments returns a tool to fetch multiple payments with filtering and pagination
//
//nolint:lll
func FetchAllPayments(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		// Pagination parameters
		mcpgo.WithNumber(
			"count",
			mcpgo.Description("Number of payments to fetch "+
				"(default: 10, max: 100)"),
			mcpgo.Min(1),
			mcpgo.Max(100),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description("Number of payments to skip (default: 0)"),
			mcpgo.Min(0),
		),
		// Time range filters
		mcpgo.WithNumber(
			"from",
			mcpgo.Description("Unix timestamp (in seconds) from when "+
				"payments are to be fetched"),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"to",
			mcpgo.Description("Unix timestamp (in seconds) up till when "+
				"payments are to be fetched"),
			mcpgo.Min(0),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		// Create query parameters map
		paymentListOptions := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddPagination(paymentListOptions).
			ValidateAndAddOptionalInt(paymentListOptions, "from").
			ValidateAndAddOptionalInt(paymentListOptions, "to")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Fetch all payments using Razorpay SDK
		payments, err := client.Payment.All(paymentListOptions, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payments failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(payments)
	}

	return mcpgo.NewTool(
		"fetch_all_payments",
		"Fetch all payments with optional filtering and pagination",
		parameters,
		handler,
	)
}
