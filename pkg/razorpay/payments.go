package razorpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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
			mcpgo.Description("Customer's email address. If not provided, "+
				"will generate a dummy email using contact number"),
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
			ValidateAndAddOptionalString(params, "email").
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

		// Handle email - generate dummy if not provided
		email := ""
		if e, exists := params["email"]; exists && e != "" {
			email = e.(string)
		} else {
			// Generate dummy email using contact number if available
			if contact, exists := params["contact"]; exists && contact != "" {
				email = contact.(string) + "@mcp.razorpay.com"
			} else {
				email = "user@mcp.razorpay.com"
			}
		}

		// Prepare payment data for the S2S JSON v1 flow
		paymentData := map[string]interface{}{
			"amount":   params["amount"],
			"currency": currency,
			"method":   method,
			"token_id": params["token_id"],
			"order_id": params["order_id"],
			"email":    email,
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
			"amount, order_id, and token_id. Supports optional parameters "+
			"like email, contact, CVV, IP, user agent, description, and notes. "+
			"If email is not provided, generates a dummy email using contact number. "+
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
			"url",
			mcpgo.Description("The URL to hit for OTP generation. "+
				"This should be a valid HTTP/HTTPS URL from the payment response."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Get client from context or use default
		_, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "url")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		url := params["url"].(string)

		obs.Logger.Infof(ctx, "Making POST request for OTP generation", "url", url)

		// Create HTTP request using the SDK's authentication
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to create HTTP request: %s", err.Error())), nil
		}

		// Get authentication credentials from environment or use test defaults
		keyID := os.Getenv("RAZORPAY_KEY_ID")
		keySecret := os.Getenv("RAZORPAY_KEY_SECRET")
		if keyID == "" || keySecret == "" {
			// Use test credentials for testing
			keyID = "rzp_test_key"
			keySecret = "rzp_test_secret"
		}
		req.SetBasicAuth(keyID, keySecret)
		req.Header.Set("Content-Type", "application/json")

		// Make POST request to the provided URL
		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to make POST request: %s", err.Error())), nil
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to read response body: %s", err.Error())), nil
		}

		// Parse JSON response if possible
		var responseData interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			// If it's not JSON, return as string
			responseData = string(body)
		}

		// Prepare response
		response := map[string]interface{}{
			"url":           url,
			"status_code":   resp.StatusCode,
			"response_data": responseData,
			"content_type":  resp.Header.Get("Content-Type"),
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			response["status"] = "success"
			response["message"] = "OTP sent successfully. Please enter the OTP " +
				"received on your mobile number to complete the payment."

			// Extract otp_submit URL from JSON response for next step
			otpSubmitURL := extractOtpSubmitURL(responseData)
			if otpSubmitURL != "" {
				response["otp_submit_url"] = otpSubmitURL
				response["next_step"] = fmt.Sprintf("Use 'verify_otp' tool with "+
					"the OTP code received from user and url %s to complete payment.",
					otpSubmitURL)
				response["next_tool"] = "verify_otp"
				response["next_tool_params"] = map[string]interface{}{
					"url":        otpSubmitURL,
					"otp_string": "{OTP_CODE_FROM_USER}",
				}
			}
		} else {
			response["status"] = "error"
			response["message"] = fmt.Sprintf("OTP generation failed with "+
				"status code: %d", resp.StatusCode)
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"send_otp",
		"Sends OTP for payment authentication. Makes a POST request to the "+
			"provided URL for OTP generation and returns the response with "+
			"next step instructions.", //nolint:lll
		parameters,
		handler,
	)
}

// VerifyOtp returns a tool that verifies OTP for payment completion
func VerifyOtp(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"url",
			mcpgo.Description("The URL to hit for OTP verification. "+
				"This should be a valid HTTP/HTTPS URL from the OTP generation response."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"otp_string",
			mcpgo.Description("The OTP code to verify. "+
				"This will be sent in the POST body."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Get client from context or use default
		_, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "url").
			ValidateAndAddRequiredString(params, "otp_string")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		url := params["url"].(string)
		otpString := params["otp_string"].(string)

		obs.Logger.Infof(ctx, "Making POST request for OTP verification", "url", url)

		// Prepare request body with OTP data
		requestBody := map[string]interface{}{
			"otp": otpString,
		}

		// Convert request body to JSON
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to marshal request body: %s", err.Error())), nil
		}

		// Create HTTP request with OTP data
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to create HTTP request: %s", err.Error())), nil
		}

		// Get authentication credentials from environment or use test defaults
		keyID := os.Getenv("RAZORPAY_KEY_ID")
		keySecret := os.Getenv("RAZORPAY_KEY_SECRET")
		if keyID == "" || keySecret == "" {
			// Use test credentials for testing
			keyID = "rzp_test_key"
			keySecret = "rzp_test_secret"
		}
		req.SetBasicAuth(keyID, keySecret)
		req.Header.Set("Content-Type", "application/json")

		// Make POST request to the provided URL with OTP data
		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to make POST request: %s", err.Error())), nil
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to read response body: %s", err.Error())), nil
		}

		// Parse JSON response if possible
		var responseData interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			// If it's not JSON, return as string
			responseData = string(body)
		}

		// Prepare response
		response := map[string]interface{}{
			"url":           url,
			"otp_sent":      otpString,
			"status_code":   resp.StatusCode,
			"response_data": responseData,
			"content_type":  resp.Header.Get("Content-Type"),
			"request_body":  requestBody,
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			response["status"] = "success"
			response["message"] = "OTP verification completed successfully. " +
				"Payment has been processed."
		} else {
			response["status"] = "error"
			response["message"] = fmt.Sprintf("OTP verification failed with "+
				"status code: %d", resp.StatusCode)
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"verify_otp",
		"Verifies OTP for payment completion. Makes a POST request to the "+
			"provided URL with the OTP code and returns the response.", //nolint:lll
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
