package razorpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

// CreatePaymentOrder returns a tool that creates a new order for payment processing
func CreatePaymentOrder(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Payment amount in the smallest currency sub-unit "+
				"(e.g., for ₹295, use 29500)"),
			mcpgo.Required(),
			mcpgo.Min(100),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("ISO code for the currency (e.g., INR, USD, SGD)"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"receipt",
			mcpgo.Description("Receipt number for internal reference (max 40 chars, must be unique)"),
		),
		mcpgo.WithBoolean(
			"partial_payment",
			mcpgo.Description("Whether the customer can make partial payments (default: false)"),
		),
		mcpgo.WithNumber(
			"first_payment_min_amount",
			mcpgo.Description("Minimum amount for first partial payment (only if partial_payment is true)"),
			mcpgo.Min(100),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for additional information "+
				"(max 15 pairs, 256 chars each)"),
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

		orderReq := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredInt(orderReq, "amount").
			ValidateAndAddRequiredString(orderReq, "currency").
			ValidateAndAddOptionalString(orderReq, "receipt").
			ValidateAndAddOptionalBool(orderReq, "partial_payment").
			ValidateAndAddOptionalInt(orderReq, "first_payment_min_amount").
			ValidateAndAddOptionalMap(orderReq, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Create the order
		order, err := client.Order.Create(orderReq, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating order failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(order)
	}

	return mcpgo.NewTool(
		"create_payment_order",
		"Use this tool to create a new order for payment processing. "+
			"Orders are required before accepting payments in Razorpay.", //nolint:lll
		parameters,
		handler,
	)
}

// AcceptAndProcessPayments returns a tool that accepts and processes payments by fetching customer card details
func AcceptAndProcessPayments(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description("Unique identifier of the customer whose card details need to be fetched. "+
				"Customer ID should start with 'cust_'"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"order_id",
			mcpgo.Description("Unique identifier of the order for which payment is being processed. "+
				"Order ID should start with 'order_'"),
		),
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Payment amount in the smallest currency sub-unit "+
				"(e.g., for ₹295, use 29500)"),
			mcpgo.Min(100),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("ISO code for the currency (e.g., INR, USD, SGD)"),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("A brief description of the payment being processed"),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for additional information "+
				"(max 15 pairs, 256 chars each)"),
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
			ValidateAndAddRequiredString(params, "customer_id").
			ValidateAndAddOptionalString(params, "order_id").
			ValidateAndAddOptionalInt(params, "amount").
			ValidateAndAddOptionalString(params, "currency").
			ValidateAndAddOptionalString(params, "description").
			ValidateAndAddOptionalMap(params, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		customerId := params["customer_id"].(string)

		// Step 1: Fetch customer details to get card information
		customer, err := client.Customer.Fetch(customerId, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching customer failed: %s", err.Error())), nil
		}

		// Step 2: Fetch saved tokens for the customer
		// Using Razorpay Go SDK Token.All() method to get all tokens for the customer
		tokenOptions := map[string]interface{}{
			"customer_id": customerId,
		}

		tokens, err := client.Token.All(customerId, tokenOptions, nil)
		if err != nil {
			// Log the error but don't fail the entire operation
			// Some customers might not have any saved tokens
			obs.Logger.Infof(ctx, "Failed to fetch customer tokens", "customer_id", customerId, "error", err.Error())
		}

		// Prepare response with customer information, tokens, and next steps
		response := map[string]interface{}{
			"customer_details": customer,
			"status":           "customer_fetched",
			"message":          "Customer details retrieved successfully. Card information is available through saved payment methods.",
			"next_steps": []interface{}{
				"Use customer tokens to fetch saved payment methods",
				"Create payment using saved cards or collect new card details",
				"Process payment against the order",
			},
		}

		// Add token information if available
		if tokens != nil {
			response["saved_tokens"] = tokens
			if tokenCount, ok := tokens["count"].(int); ok && tokenCount > 0 {
				response["message"] = fmt.Sprintf("Customer details retrieved successfully. Found %d saved payment method(s).", tokenCount)
				response["next_steps"] = []interface{}{
					"Use saved tokens for quick payment processing",
					"Create payment using existing saved cards",
					"Process payment against the order",
				}
			} else {
				response["message"] = "Customer details retrieved successfully. No saved payment methods found."
				response["next_steps"] = []interface{}{
					"Collect new card details from customer",
					"Create payment with new card information",
					"Process payment against the order",
				}
			}
		}

		// Add optional parameters to response if provided
		if orderId, exists := params["order_id"]; exists {
			response["order_id"] = orderId
		}
		if amount, exists := params["amount"]; exists {
			response["amount"] = amount
		}
		if currency, exists := params["currency"]; exists {
			response["currency"] = currency
		}
		if description, exists := params["description"]; exists {
			response["description"] = description
		}
		if notes, exists := params["notes"]; exists {
			response["notes"] = notes
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"accept_and_process_payments",
		"Use this tool to accept and process payments by fetching customer card details. "+
			"This tool retrieves customer information and prepares for payment processing.", //nolint:lll
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

// CreatePaymentsByToken returns a tool to create a payment using a saved token
//
//nolint:lll
func CreatePaymentsByToken(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"token_id",
			mcpgo.Description("Unique identifier of the saved token to use for payment. "+
				"Token ID should start with 'token_'"),
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Payment amount in the smallest currency sub-unit "+
				"(e.g., for ₹295, use 29500)"),
			mcpgo.Required(),
			mcpgo.Min(100),
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

		// Debug: Log that the handler is being called
		obs.Logger.Infof(ctx, "CreatePaymentsByToken handler called")

		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "token_id").
			ValidateAndAddRequiredInt(params, "amount")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Prepare payment data for token-based payment
		paymentData := map[string]interface{}{
			"amount":   params["amount"],
			"currency": "INR",
			"contact":  9015410084,
			"email":    "gaurav.kumar@example.com",
			"method":   "card",
			"token":    params["token_id"],
		}

		// Log paymentData in JSON format before calling the endpoint
		if jsonBytes, jsonErr := json.Marshal(paymentData); jsonErr == nil {
			obs.Logger.Infof(ctx, "=== PAYMENT DATA JSON ===")
			obs.Logger.Infof(ctx, string(jsonBytes))
			obs.Logger.Infof(ctx, "========================")
		}

		// Create payment using Razorpay SDK
		// Using CreatePaymentJson method from the official SDK
		// Reference: https://github.com/razorpay/razorpay-go/blob/f24483c59e84302bc4908cb474a2258c539b8af7/resources/payment.go#L75
		payment, err := client.Payment.CreatePaymentJson(paymentData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating payment with token failed: %s", err.Error())), nil
		}

		// Extract payment ID from the response - try both direct and nested paths
		var paymentID string
		if id, exists := payment["razorpay_payment_id"]; exists && id != nil {
			paymentID = id.(string)
		} else if paymentDetails, exists := payment["payment_details"]; exists && paymentDetails != nil {
			if details, ok := paymentDetails.(map[string]interface{}); ok {
				if id, exists := details["razorpay_payment_id"]; exists && id != nil {
					paymentID = id.(string)
				}
			}
		}

		// Extract redirect URL from the next array where action equals "otp_generate"
		// Look in both direct path and nested payment_details path
		var otpRedirectURL string

		// First try direct path: payment["next"]
		if nextArray, exists := payment["next"]; exists && nextArray != nil {
			if nextSlice, ok := nextArray.([]interface{}); ok {
				for _, item := range nextSlice {
					if nextItem, ok := item.(map[string]interface{}); ok {
						if action, exists := nextItem["action"]; exists && action == "otp_generate" {
							if url, exists := nextItem["url"]; exists && url != nil {
								otpRedirectURL = url.(string)
								break
							}
						}
					}
				}
			}
		}

		// If not found, try nested path: payment["payment_details"]["next"]
		if otpRedirectURL == "" {
			if paymentDetails, exists := payment["payment_details"]; exists && paymentDetails != nil {
				if details, ok := paymentDetails.(map[string]interface{}); ok {
					if nextArray, exists := details["next"]; exists && nextArray != nil {
						if nextSlice, ok := nextArray.([]interface{}); ok {
							for _, item := range nextSlice {
								if nextItem, ok := item.(map[string]interface{}); ok {
									if action, exists := nextItem["action"]; exists && action == "otp_generate" {
										if url, exists := nextItem["url"]; exists && url != nil {
											otpRedirectURL = url.(string)
											break
										}
									}
								}
							}
						}
					}
				}
			}
		}

		// Prepare response with payment and OTP information
		response := map[string]interface{}{
			"payment_details": payment,
			"payment_id":      paymentID,
			"status":          "payment_created",
			"message":         "Payment created successfully using saved token.",
		}

		response["current_step"] = "create_payments_by_token"
		response["status"] = "payment_created"

		// Add redirect URL if found
		if otpRedirectURL != "" {
			response["otpRedirectURL"] = otpRedirectURL
			response["message"] = "Payment created successfully using saved token. Redirect URL available for authentication."

			// Signal automatic next tool execution
			response["auto_execute_next"] = true
			response["next_tool"] = "otp_generate_for_payment"
			response["next_tool_params"] = map[string]interface{}{
				"url": otpRedirectURL,
			}
			response["next_step"] = "Auto-executing otp_generate_for_payment tool..."
		} else {
			response["message"] = "Payment created successfully using saved token. No additional authentication required."
			response["auto_execute_next"] = false
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"create_payments_by_token",
		"Create a new payment using a saved token. This tool accepts a token_id and amount and "+
			"creates a payment using the saved payment method associated with that token.", //nolint:lll
		parameters,
		handler,
	)
}

// OtpGenerateForPayment returns a tool that makes a POST request to generate OTP for payment
func OtpGenerateForPayment(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"url",
			mcpgo.Description("The URL to hit for OTP generation. This should be a valid HTTP/HTTPS URL."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "url")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		url := params["url"].(string)

		obs.Logger.Infof(ctx, "Making POST request for OTP generation", "url", url)

		// Create HTTP request with authorization header
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to create HTTP request: %s", err.Error())), nil
		}

		// Add Authorization header
		req.Header.Set("Authorization", "Basic cnpwX2xpdmVfNU13MFdHQ21OOE1NRDQ6bzN2THN4akNvcGxhZE9PWTcyN0dhOFl3")
		req.Header.Set("Content-Type", "application/json")

		// Make POST request to the provided URL
		client := &http.Client{}
		resp, err := client.Do(req)
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
			"message":       "Please enter the OTP received on your mobile number to complete the payment...",
			"content_type":  resp.Header.Get("Content-Type"),
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			response["status"] = "success"

			// Extract otp_submit URL from JSON response
			var otpSubmitURL string
			if jsonData, ok := responseData.(map[string]interface{}); ok {
				if nextArray, exists := jsonData["next"]; exists && nextArray != nil {
					if nextSlice, ok := nextArray.([]interface{}); ok {
						for _, item := range nextSlice {
							if nextItem, ok := item.(map[string]interface{}); ok {
								if action, exists := nextItem["action"]; exists && action == "otp_submit" {
									if submitURL, exists := nextItem["url"]; exists && submitURL != nil {
										otpSubmitURL = submitURL.(string)
										break
									}
								}
							}
						}
					}
				}
			}

			// Set next step with extracted otp_submit URL
			if otpSubmitURL != "" {
				response["otp_submit_url"] = otpSubmitURL
				response["next_step"] = fmt.Sprintf("Use 'otp_verify_for_payment' tool with the OTP code received from user and url %s to complete payment.", otpSubmitURL)
				response["auto_execute_next"] = true
				response["next_tool"] = "otp_verify_for_payment"
				response["next_tool_params"] = map[string]interface{}{
					"url":        otpSubmitURL,
					"otp_string": "{OTP_CODE_FROM_USER}",
				}
			} else {
				response["auto_execute_next"] = false
			}
		} else {
			response["status"] = "error"
			response["message"] = fmt.Sprintf("POST request failed with status code: %d", resp.StatusCode)
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"otp_generate_for_payment",
		"Makes a POST request to the provided URL for OTP generation. "+
			"This tool hits the specified URL and returns the response.", //nolint:lll
		parameters,
		handler,
	)
}

// OtpVerifyForPayment returns a tool that makes a POST request to verify OTP for payment
func OtpVerifyForPayment(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"url",
			mcpgo.Description("The URL to hit for OTP verification. This should be a valid HTTP/HTTPS URL."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"otp_string",
			mcpgo.Description("The OTP code to verify. This will be sent in the POST body."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
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

		// Create HTTP request with authorization header and OTP data
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to create HTTP request: %s", err.Error())), nil
		}

		// Add Authorization header
		req.Header.Set("Authorization", "Basic cnpwX2xpdmVfNU13MFdHQ21OOE1NRDQ6bzN2THN4akNvcGxhZE9PWTcyN0dhOFl3")
		req.Header.Set("Content-Type", "application/json")

		// Make POST request to the provided URL with OTP data
		client := &http.Client{}
		resp, err := client.Do(req)
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
			"message":       "OTP verification POST request completed successfully",
			"content_type":  resp.Header.Get("Content-Type"),
			"request_body":  requestBody,
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			response["status"] = "success"
		} else {
			response["status"] = "error"
			response["message"] = fmt.Sprintf("OTP verification failed with status code: %d", resp.StatusCode)
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"otp_verify_for_payment",
		"Makes a POST request to the provided URL for OTP verification. "+
			"This tool sends the OTP string in the request body as {\"otp\":\"otp_string\"} and returns the response.", //nolint:lll
		parameters,
		handler,
	)
}

// AcceptPaymentsByChat returns a tool that accepts mobile number and amount,
// displays a dummy card and asks for user confirmation
func AcceptPaymentsByChat(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"mobile_number",
			mcpgo.Description("Mobile number of the customer for payment processing"),
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Payment amount in INR (e.g., for ₹300, use 300)"),
			mcpgo.Required(),
			mcpgo.Min(1),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "mobile_number").
			ValidateAndAddRequiredInt(params, "amount")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		mobileNumber := params["mobile_number"].(string)
		amount := params["amount"].(int64)

		obs.Logger.Infof(ctx, "AcceptPaymentsByChat called",
			"mobile_number", mobileNumber, "amount", amount)

		// Step 1: Create customer using the mobile number
		customerData := map[string]interface{}{
			"contact":       mobileNumber,
			"fail_existing": "0", // Return existing customer if already exists
		}

		customer, err := client.Customer.Create(customerData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating customer failed: %s", err.Error())), nil
		}

		// Extract customer ID from response
		customerId, ok := customer["id"].(string)
		if !ok {
			return mcpgo.NewToolResultError("failed to extract customer ID from response"), nil
		}

		obs.Logger.Infof(ctx, "Customer created/retrieved", "customer_id", customerId)

		// Step 2: Fetch all tokens for the customer
		tokens, err := client.Token.All(customerId, nil, nil)
		if err != nil {
			obs.Logger.Infof(ctx, "Failed to fetch tokens for customer", "customer_id", customerId, "error", err.Error())
			// Don't fail the entire operation, just log the error
			tokens = map[string]interface{}{
				"count": 0,
				"items": []interface{}{},
				"error": err.Error(),
			}
		}

		// Extract token ID from tokens response
		var tokenId string
		var nextSteps []string

		if tokenItems, ok := tokens["items"]; ok {
			if itemsArray, ok := tokenItems.([]interface{}); ok && len(itemsArray) > 0 {
				if firstToken, ok := itemsArray[0].(map[string]interface{}); ok {
					if id, ok := firstToken["id"].(string); ok {
						tokenId = id
					}
				}
			}
		}

		// Prepare response with customer and token information
		response := map[string]interface{}{
			"mobile_number":    mobileNumber,
			"amount_inr":       amount,
			"amount_paisa":     amount * 100, // Convert to paisa for API usage
			"customer_details": customer,
			"customer_id":      customerId,
			"tokens_response":  tokens,
			"status":           "customer_and_tokens_fetched",
			"message": fmt.Sprintf("Payment Details:\n- Mobile: %s\n- Amount: ₹%d\n- Customer ID: %s\n- Tokens Found: %v\n\nCustomer and token details retrieved successfully.",
				mobileNumber, amount, customerId, tokens),
			"confirmation_required": true,
			"auto_execute_next":     false,
		}

		// Add next tool parameters based on token availability
		if tokenId != "" {
			response["next_tool_on_card"] = "create_payments_by_token"
			response["next_tool_on_wallet"] = "create_payment_wallet"
			response["next_tool_params"] = map[string]interface{}{
				//"token_id": tokenId,
				"token_id": "token_QoWfObL2yAS5Yk",
				"amount":   amount * 100, // Convert to paisa for API
			}
			nextSteps = []string{
				"User should respond with 'card' or 'wallet' to continue",
				"User can respond with 'no' or 'cancel' to cancel",
				fmt.Sprintf("Upon 'card' confirmation, create_payments_by_token will be executed with %s", tokenId),
				fmt.Sprintf("Upon 'wallet' confirmation, create_payment_wallet will be executed with %s", tokenId),
			}
		} else {
			nextSteps = []string{
				"No saved payment tokens found for this customer",
				"Customer needs to provide new payment details",
				"Consider collecting card details or other payment methods",
			}
		}

		response["next_steps"] = nextSteps

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"acceptpaymentsBychat",
		"Create a payment request by accepting mobile number and amount in INR. "+
			"and asks for user confirmation to proceed.", //nolint:lll
		parameters,
		handler,
	)
}

// CreatePayment returns a tool that creates a payment with wallet using the CreatePaymentJson function
func CreatePaymentWallet(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Payment amount in the smallest currency sub-unit "+
				"(e.g., for ₹295, use 29500)"),
			mcpgo.Required(),
			mcpgo.Min(100), // Minimum amount is 100 (1.00 in currency)
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("ISO code for the currency to be taken as INR"),
			mcpgo.Required(),
			mcpgo.Pattern("^[A-Z]{3}$"), // ISO currency codes are 3 uppercase letters
		),
		mcpgo.WithString(
			"method",
			mcpgo.Description("Payment method to be taken as wallet"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"wallet",
			mcpgo.Description("wallet for the payment to be taken as amazonpay"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"email",
			mcpgo.Description("Customer's email address to be taken as kushalsalecha@yahoo.com"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"contact",
			mcpgo.Description("Customer's contact number"),
			mcpgo.Required(),
		),
		// Card-specific parameters (optional)

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

		paymentData := make(map[string]interface{})
		cardData := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredFloat(paymentData, "amount").
			ValidateAndAddRequiredString(paymentData, "currency").
			ValidateAndAddRequiredString(paymentData, "method").
			ValidateAndAddRequiredString(paymentData, "wallet").
			ValidateAndAddRequiredString(paymentData, "email").
			ValidateAndAddRequiredString(paymentData, "contact")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Validate card method requirements
		//method := paymentData["method"].(string)
		// if method == "card" {
		// 	// Check if all required card fields are present
		// 	requiredCardFields := []string{"token"}
		// 	for _, field := range requiredCardFields {
		// 		if _, exists := cardData[field]; !exists {
		// 			return mcpgo.NewToolResultError(
		// 				fmt.Sprintf("%s is required when method is 'card'", field)), nil
		// 		}
		// 	}

		// 	// Add card data to payment data
		// 	card := make(map[string]interface{})
		// 	card["number"] = cardData["card_number"]
		// 	card["expiry_month"] = cardData["card_expiry_month"]
		// 	card["expiry_year"] = cardData["card_expiry_year"]
		// 	card["cvv"] = cardData["card_cvv"]
		// 	card["name"] = cardData["card_name"]

		// 	paymentData["card"] = card
		// }

		// Remove card fields from the top level as they are now nested under 'card'
		for key := range cardData {
			delete(paymentData, key)
		}

		// Create payment using Razorpay SDK's CreatePaymentJson
		payment, err := client.Payment.CreatePaymentJson(paymentData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating payment failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(payment)
	}

	return mcpgo.NewTool(
		"create_payment",
		"Create a payment using Razorpay's CreatePaymentJson function. Supports various payment methods including cards and wallets.", //nolint:lll
		parameters,
		handler,
	)
}
