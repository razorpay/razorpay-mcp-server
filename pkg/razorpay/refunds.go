package razorpay

import (
	"context"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// refundIdempotencyHeader carries the caller-supplied idempotency key on
// the create-refund request.
// https://razorpay.com/docs/api/refunds/normal-refunds-idempotent/
const refundIdempotencyHeader = "X-Refund-Idempotency"

// idempotencyKeyPattern is the key format Razorpay documents:
// at least 10 characters of letters, numbers, hyphens and underscores.
const idempotencyKeyPattern = `^[A-Za-z0-9_-]{10,}$`

// CreateRefund returns a tool that creates a normal refund for a payment
func CreateRefund(
	obs *observability.Observability,
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
			mcpgo.Description("Payment amount in paise (smallest currency unit). "+
				"For INR: 100 paise = ₹1. Example: for ₹295, use 29500"),
			mcpgo.Required(),
			mcpgo.Min(100), // Minimum amount is 100 (1.00 in currency)
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
		mcpgo.WithString(
			"idempotency_key",
			mcpgo.Description("Makes this refund request retryable. "+
				"Retrying with the same idempotency_key and an identical "+
				"request body returns the original refund instead of "+
				"creating a second one. Reuse the same value for every "+
				"retry of the same refund; a new value creates a separate "+
				"refund. At least 10 characters; letters, numbers, "+
				"hyphens and underscores only."),
			mcpgo.Pattern(idempotencyKeyPattern),
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

		payload := make(map[string]interface{})
		data := make(map[string]interface{})
		options := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(payload, "payment_id").
			ValidateAndAddRequiredFloat(payload, "amount").
			ValidateAndAddOptionalString(data, "speed").
			ValidateAndAddOptionalString(data, "receipt").
			ValidateAndAddOptionalMap(data, "notes").
			ValidateAndAddOptionalString(options, "idempotency_key")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// extraHeaders stays nil when no key is supplied, so the request is
		// unchanged for every existing caller.
		var extraHeaders map[string]string
		if key, ok := options["idempotency_key"].(string); ok && key != "" {
			extraHeaders = map[string]string{refundIdempotencyHeader: key}
		}

		refund, err := client.Payment.Refund(
			payload["payment_id"].(string),
			int(payload["amount"].(float64)), data, extraHeaders)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("creating refund failed", err)), nil
		}

		return mcpgo.NewToolResultJSON(refund)
	}

	return mcpgo.NewTool(
		"create_refund",
		"Use this tool to create a normal refund for a payment. "+
			"Amount should be in paise (smallest currency unit). "+
			"For INR: 100 paise = ₹1. Example: for ₹295, use 29500",
		parameters,
		handler,
	)
}

// FetchRefund returns a tool that fetches a refund by ID
func FetchRefund(
	obs *observability.Observability,
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
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		payload := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(payload, "refund_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		refund, err := client.Refund.Fetch(payload["refund_id"].(string), nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("fetching refund failed", err)), nil
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
	obs *observability.Observability,
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
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

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
				formatErrorMessage("updating refund failed", err)), nil
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

// FetchMultipleRefundsForPayment returns a tool that fetches multiple refunds
// for a payment
func FetchMultipleRefundsForPayment(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment for which "+
				"refunds are to be retrieved. ID should have a pay_ prefix."),
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
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		fetchReq := make(map[string]interface{})
		fetchOptions := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(fetchReq, "payment_id").
			ValidateAndAddOptionalInt(fetchOptions, "from").
			ValidateAndAddOptionalInt(fetchOptions, "to").
			ValidateAndAddPagination(fetchOptions)

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		refunds, err := client.Payment.FetchMultipleRefund(
			fetchReq["payment_id"].(string), fetchOptions, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("fetching multiple refunds failed", err)), nil
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
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment for which "+
				"the refund has been made. ID should have a pay_ prefix."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"refund_id",
			mcpgo.Description("Unique identifier of the refund to be retrieved. "+
				"ID should have a rfnd_ prefix."),
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
			ValidateAndAddRequiredString(params, "payment_id").
			ValidateAndAddRequiredString(params, "refund_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		refund, err := client.Payment.FetchRefund(
			params["payment_id"].(string),
			params["refund_id"].(string),
			nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("fetching specific refund for payment failed", err)), nil
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

// FetchAllRefunds returns a tool that fetches all refunds with pagination
// support
func FetchAllRefunds(
	obs *observability.Observability,
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
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		queryParams := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddOptionalInt(queryParams, "from").
			ValidateAndAddOptionalInt(queryParams, "to").
			ValidateAndAddPagination(queryParams)

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		refunds, err := client.Refund.All(queryParams, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("fetching refunds failed", err)), nil
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
