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

// extractPaymentID extracts the payment ID from the payment response
func extractPaymentID(payment map[string]interface{}) string {
	if id, exists := payment["razorpay_payment_id"]; exists && id != nil {
		return id.(string)
	}
	return ""
}

// extractNextActionDetails extracts action and URL from the payment response
func extractNextActionDetails(payment map[string]interface{}) (string, string) {
	var action, otpURL string
	if nextArray, exists := payment["next"]; exists && nextArray != nil {
		if nextSlice, ok := nextArray.([]interface{}); ok {
			for _, item := range nextSlice {
				if nextItem, ok := item.(map[string]interface{}); ok {
					if actionVal, exists := nextItem["action"]; exists {
						action = actionVal.(string)
						if url, exists := nextItem["url"]; exists && url != nil {
							otpURL = url.(string)
						}
						break
					}
				}
			}
		}
	}
	return action, otpURL
}

// addOptionalPaymentParameters adds optional parameters to payment data
func addOptionalPaymentParameters(
	paymentData map[string]interface{},
	params map[string]interface{},
) {
	if contact, exists := params["contact"]; exists && contact != "" {
		paymentData["contact"] = contact
	}
	if cvv, exists := params["cvv"]; exists && cvv != "" {
		// CVV goes inside card object for saved card payments
		paymentData["card"] = map[string]interface{}{
			"cvv": cvv,
		}
	}
	if ip, exists := params["ip"]; exists && ip != "" {
		paymentData["ip"] = ip
	}
	if userAgent, exists := params["user_agent"]; exists && userAgent != "" {
		paymentData["user_agent"] = userAgent
	}
	if description, exists := params["description"]; exists && description != "" {
		paymentData["description"] = description
	}
	if notes, exists := params["notes"]; exists && notes != nil {
		paymentData["notes"] = notes
	}
}

// InitiatePayment returns a tool that initiates a payment using order_id
// and token_id
// This implements the S2S JSON v1 flow for creating payments
func InitiatePayment(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Payment amount in the smallest currency sub-unit "+
				"(e.g., for ₹100, use 10000)"),
			mcpgo.Required(),
			mcpgo.Min(100),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("Currency code for the payment. Default is 'INR'"),
		),
		mcpgo.WithString(
			"method",
			mcpgo.Description("Payment method to use. "+
				"Options: 'card', 'upi'. Default is 'card'"),
		),
		mcpgo.WithString(
			"token_id",
			mcpgo.Description("Token ID of the saved payment method. "+
				"Must start with 'token_'"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"order_id",
			mcpgo.Description("Order ID for which the payment is being initiated. "+
				"Must start with 'order_'"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"email",
			mcpgo.Description("Customer's email address"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"contact",
			mcpgo.Description("Customer's phone number"),
		),
		mcpgo.WithString(
			"cvv",
			mcpgo.Description("CVV for card payments when using saved cards"),
		),
		mcpgo.WithString(
			"ip",
			mcpgo.Description("Customer's IP address"),
		),
		mcpgo.WithString(
			"user_agent",
			mcpgo.Description("Customer's browser user agent"),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("Description of the payment"),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for additional information"),
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
			ValidateAndAddRequiredInt(params, "amount").
			ValidateAndAddOptionalString(params, "currency").
			ValidateAndAddOptionalString(params, "method").
			ValidateAndAddRequiredString(params, "token_id").
			ValidateAndAddRequiredString(params, "order_id").
			ValidateAndAddRequiredString(params, "email").
			ValidateAndAddOptionalString(params, "contact").
			ValidateAndAddOptionalString(params, "cvv").
			ValidateAndAddOptionalString(params, "ip").
			ValidateAndAddOptionalString(params, "user_agent").
			ValidateAndAddOptionalString(params, "description").
			ValidateAndAddOptionalMap(params, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Set default values
		currency := "INR"
		if c, exists := params["currency"]; exists && c != "" {
			currency = c.(string)
		}

		method := "card"
		if m, exists := params["method"]; exists && m != "" {
			method = m.(string)
		}

		// Prepare payment data for the S2S JSON v1 flow
		paymentData := map[string]interface{}{
			"amount":   params["amount"],
			"currency": currency,
			"method":   method,
			"token_id": params["token_id"],
			"order_id": params["order_id"],
			"email":    params["email"], // Required parameter
		}

		// Add optional parameters if provided
		addOptionalPaymentParameters(paymentData, params)

		// Create payment using Razorpay SDK's CreatePaymentJson method
		// This follows the S2S JSON v1 flow:
		// https://api.razorpay.com/v1/payments/create/json
		payment, err := client.Payment.CreatePaymentJson(paymentData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("initiating payment failed: %s", err.Error())), nil
		}

		// Extract payment ID and next action details
		paymentID := extractPaymentID(payment)
		action, otpURL := extractNextActionDetails(payment)

		// Prepare response
		response := map[string]interface{}{
			"razorpay_payment_id": paymentID,
			"payment_details":     payment,
			"status":              "payment_initiated",
			"message":             "Payment initiated successfully using S2S JSON v1 flow", //nolint:lll
		}

		// Add action and URL if available
		if action != "" {
			response["action"] = action
			if otpURL != "" {
				response["url"] = otpURL
				response["message"] = fmt.Sprintf(
					"Payment initiated. Next action: %s. "+
						"Use the provided URL for next step.", action)
			}
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"initiate_payment",
		"Initiate a payment using the S2S JSON v1 flow with required parameters: "+
			"amount, order_id, token_id, and email. Supports optional parameters "+
			"like contact, CVV, IP, user agent, description, and notes. "+
			"Returns payment details including next action steps if required.", //nolint:lll
		parameters,
		handler,
	)
}
