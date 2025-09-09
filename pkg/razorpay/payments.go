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

// InitiatePayment returns a tool that initiates a payment using order_id
// and token
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
			"token",
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
			mcpgo.Description("Customer's email address (optional)"),
		),
		mcpgo.WithString(
			"contact",
			mcpgo.Description("Customer's phone number"),
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
			ValidateAndAddRequiredString(params, "token").
			ValidateAndAddRequiredString(params, "order_id").
			ValidateAndAddOptionalString(params, "email").
			ValidateAndAddOptionalString(params, "contact")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Set default currency
		currency := "INR"
		if c, exists := params["currency"]; exists && c != "" {
			currency = c.(string)
		}

		// Prepare payment data for the S2S JSON v1 flow
		paymentData := map[string]interface{}{
			"amount":   params["amount"],
			"currency": currency,
			"order_id": params["order_id"],
			"token":    params["token"],
		}

		if contact, exists := params["contact"]; exists && contact != "" {
			paymentData["contact"] = contact
		}

		// Add optional parameters if provided
		if email, exists := params["email"]; exists && email != "" {
			paymentData["email"] = email
		} else {
			paymentData["email"] = paymentData["contact"].(string) + "@mcp.razorpay.com"
		}

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
				response["next_step_instruction"] = "Call the 'send_otp' tool next " +
					"in the process to complete the payment authentication."
			}
		} else {
			response["next_step_instruction"] = "Payment initiated successfully. " +
				"If OTP is required, call the 'send_otp' tool next in the process."
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"initiate_payment",
		"Initiate a payment using the S2S JSON v1 flow with required parameters: "+
			"amount, order_id, and token. Supports optional parameters "+
			"like email and contact. "+
			"Returns payment details including next action steps if required.", //nolint:lll
		parameters,
		handler,
	)
}

// SendOtp returns a tool that sends OTP for payment authentication
func SendOtp(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment for which "+
				"OTP needs to be generated. Must start with 'pay_'"),
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

		paymentID := params["payment_id"].(string)

		// Generate OTP using Razorpay SDK
		otpResponse, err := client.Payment.OtpGenerate(paymentID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("OTP generation failed: %s", err.Error())), nil
		}

		// Prepare response
		response := map[string]interface{}{
			"payment_id":    paymentID,
			"status":        "success",
			"message":       "OTP sent successfully. Please enter the OTP received on your mobile number to complete the payment.", //nolint:lll
			"response_data": otpResponse,
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"send_otp",
		"Generate and send an OTP to the customer's registered mobile number "+
			"for payment authentication.",
		parameters,
		handler,
	)
}

// ResendOtp returns a tool that sends OTP for payment authentication
func ResendOtp(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment for which "+
				"OTP needs to be generated. Must start with 'pay_'"),
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

		paymentID := params["payment_id"].(string)

		// Resend OTP using Razorpay SDK
		otpResponse, err := client.Payment.OtpResend(paymentID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("OTP resend failed: %s", err.Error())), nil
		}

		// Extract OTP submit URL from response
		otpSubmitURL := extractOtpSubmitURL(otpResponse)

		// Prepare response
		response := map[string]interface{}{
			"payment_id":    paymentID,
			"status":        "success",
			"message":       "OTP sent successfully. Please enter the OTP received on your mobile number to complete the payment.", //nolint:lll
			"response_data": otpResponse,
		}

		// Add next step instructions if OTP submit URL is available
		if otpSubmitURL != "" {
			response["otp_submit_url"] = otpSubmitURL
			response["next_step"] = fmt.Sprintf(
				"Use 'submit_otp' tool with the OTP code received from user and url %s to complete payment.", //nolint:lll
				otpSubmitURL)
			response["next_tool"] = "submit_otp"
			response["next_tool_params"] = map[string]interface{}{
				"url":        otpSubmitURL,
				"otp_string": "{OTP_CODE_FROM_USER}",
			}
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"resend_otp",
		"Resend OTP to the customer's registered mobile number if the previous "+
			"OTP was not received or has expired.",
		parameters,
		handler,
	)
}

// SubmitOtp returns a tool that submits OTP for payment verification
func SubmitOtp(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"otp_string",
			mcpgo.Description("OTP string received from the user"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Unique identifier of the payment for which "+
				"OTP needs to be submitted. Must start with 'pay_'"),
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
			ValidateAndAddRequiredString(params, "otp_string").
			ValidateAndAddRequiredString(params, "payment_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		paymentID := params["payment_id"].(string)
		data := map[string]interface{}{
			"otp": params["otp_string"].(string),
		}
		otpResponse, err := client.Payment.OtpSubmit(paymentID, data, nil)

		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("OTP verification failed: %s", err.Error())), nil
		}

		// Prepare response
		response := map[string]interface{}{
			"payment_id":    paymentID,
			"status":        "success",
			"message":       "OTP verified successfully.",
			"response_data": otpResponse,
		}
		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"submit_otp",
		"Verify and submit the OTP received by the customer to complete "+
			"the payment authentication process.",
		parameters,
		handler,
	)
}

// extractOtpSubmitURL extracts the OTP submit URL from the payment response
func extractOtpSubmitURL(responseData interface{}) string {
	jsonData, ok := responseData.(map[string]interface{})
	if !ok {
		return ""
	}

	nextArray, exists := jsonData["next"]
	if !exists || nextArray == nil {
		return ""
	}

	nextSlice, ok := nextArray.([]interface{})
	if !ok {
		return ""
	}

	for _, item := range nextSlice {
		nextItem, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		action, exists := nextItem["action"]
		if !exists || action != "otp_submit" {
			continue
		}

		submitURL, exists := nextItem["url"]
		if exists && submitURL != nil {
			return submitURL.(string)
		}
	}

	return ""
}
