package razorpay

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"

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

// extractNextActions extracts all available actions from the payment response
func extractNextActions(
	payment map[string]interface{},
) []map[string]interface{} {
	var actions []map[string]interface{}
	if nextArray, exists := payment["next"]; exists && nextArray != nil {
		if nextSlice, ok := nextArray.([]interface{}); ok {
			for _, item := range nextSlice {
				if nextItem, ok := item.(map[string]interface{}); ok {
					actions = append(actions, nextItem)
				}
			}
		}
	}
	return actions
}

// OTPResponse represents the response from OTP generation API

// sendOtp sends an OTP to the customer and returns the response
func sendOtp(otpUrl string) error {
	if otpUrl == "" {
		return fmt.Errorf("OTP URL is empty")
	}
	// Validate URL is safe and from Razorpay domain for security
	parsedURL, err := url.Parse(otpUrl)
	if err != nil {
		return fmt.Errorf("invalid OTP URL: %s", err.Error())
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("OTP URL must use HTTPS")
	}

	if !strings.Contains(parsedURL.Host, "razorpay.com") {
		return fmt.Errorf("OTP URL must be from Razorpay domain")
	}

	// Create a secure HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", otpUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to create OTP request: %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("OTP generation failed: %s", err.Error())
	}
	defer resp.Body.Close()

	// Validate HTTP response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("OTP generation failed with HTTP status: %d",
			resp.StatusCode)
	}
	return nil
}

// buildInitiatePaymentResponse constructs the response for initiate payment
func buildInitiatePaymentResponse(
	payment map[string]interface{},
	paymentID string,
	actions []map[string]interface{},
) (map[string]interface{}, string) {
	response := map[string]interface{}{
		"razorpay_payment_id": paymentID,
		"payment_details":     payment,
		"status":              "payment_initiated",
		"message": "Payment initiated successfully using " +
			"S2S JSON v1 flow",
	}
	otpUrl := ""

	if len(actions) > 0 {
		response["available_actions"] = actions

		// Add guidance based on available actions
		var actionTypes []string
		hasOTP := false
		hasRedirect := false
		hasUPICollect := false
		hasUPIIntent := false

		for _, action := range actions {
			if actionType, exists := action["action"]; exists {
				actionStr := actionType.(string)
				actionTypes = append(actionTypes, actionStr)
				if actionStr == "otp_generate" {
					hasOTP = true
					otpUrl = action["url"].(string)
				}

				if actionStr == "redirect" {
					hasRedirect = true
				}

				if actionStr == "upi_collect" {
					hasUPICollect = true
				}

				if actionStr == "upi_intent" {
					hasUPIIntent = true
				}
			}
		}

		switch {
		case hasOTP:
			response["message"] = "Payment initiated. OTP authentication is " +
				"available. " +
				"Use the 'submit_otp' tool to submit OTP received by the customer " +
				"for authentication."
			addNextStepInstructions(response, paymentID)
		case hasRedirect:
			response["message"] = "Payment initiated. Redirect authentication is " +
				"available. Use the redirect URL provided in available_actions."
		case hasUPICollect:
			response["message"] = fmt.Sprintf(
				"Payment initiated. Available actions: %v", actionTypes)
		case hasUPIIntent:
			response["message"] = fmt.Sprintf(
				"Payment initiated. Available actions: %v", actionTypes)
		default:
			response["message"] = fmt.Sprintf(
				"Payment initiated. Available actions: %v", actionTypes)
		}
	} else {
		addFallbackNextStepInstructions(response, paymentID)
	}

	return response, otpUrl
}

// addNextStepInstructions adds next step guidance to the response
func addNextStepInstructions(
	response map[string]interface{},
	paymentID string,
) {
	if paymentID != "" {
		response["next_step"] = "Use 'resend_otp' to regenerate OTP or " +
			"'submit_otp' to proceed to enter OTP."
		response["next_tool"] = "resend_otp"
		response["next_tool_params"] = map[string]interface{}{
			"payment_id": paymentID,
		}
	}
}

// addFallbackNextStepInstructions adds fallback next step guidance
func addFallbackNextStepInstructions(
	response map[string]interface{},
	paymentID string,
) {
	if paymentID != "" {
		response["next_step"] = "Use 'resend_otp' to regenerate OTP or " +
			"'submit_otp' to proceed to enter OTP if " +
			"OTP authentication is required."
		response["next_tool"] = "resend_otp"
		response["next_tool_params"] = map[string]interface{}{
			"payment_id": paymentID,
		}
	}
}

// addContactAndEmailToPaymentData adds contact and email to payment data
func addContactAndEmailToPaymentData(
	paymentData map[string]interface{},
	params map[string]interface{},
) {
	// Add contact if provided
	if contact, exists := params["contact"]; exists && contact != "" {
		paymentData["contact"] = contact
	}

	// Add email if provided, otherwise generate from contact
	if email, exists := params["email"]; exists && email != "" {
		paymentData["email"] = email
	} else if contact, exists := paymentData["contact"]; exists && contact != "" {
		paymentData["email"] = contact.(string) + "@mcp.razorpay.com"
	}
}

// addAdditionalPaymentParameters adds additional parameters for UPI collect
// and other flows
func addAdditionalPaymentParameters(
	paymentData map[string]interface{},
	params map[string]interface{},
) {
	// Note: customer_id is now handled explicitly in buildPaymentData

	// Add method if provided
	if method, exists := params["method"]; exists && method != "" {
		paymentData["method"] = method
	}

	// Add save if provided
	if save, exists := params["save"]; exists {
		paymentData["save"] = save
	}

	// Add recurring if provided
	if recurring, exists := params["recurring"]; exists {
		paymentData["recurring"] = recurring
	}

	// Add UPI parameters if provided
	if upiParams, exists := params["upi"]; exists && upiParams != nil {
		if upiMap, ok := upiParams.(map[string]interface{}); ok {
			paymentData["upi"] = upiMap
		}
	}

	// Add wallet parameter if provided (for wallet payments)
	if wallet, exists := params["wallet"]; exists && wallet != "" {
		paymentData["wallet"] = wallet
	}
}

// processUPIParameters handles VPA and UPI intent parameter processing
func processUPIParameters(params map[string]interface{}) {
	vpa, hasVPA := params["vpa"]
	upiIntent, hasUPIIntent := params["upi_intent"]

	// Handle VPA parameter (UPI collect flow)
	if hasVPA && vpa != "" {
		// Set method to UPI
		params["method"] = "upi"
		// Set UPI parameters for collect flow
		params["upi"] = map[string]interface{}{
			"flow":        "collect",
			"expiry_time": "6",
			"vpa":         vpa,
		}
	}

	// Handle UPI intent parameter (UPI intent flow)
	if hasUPIIntent && upiIntent == true {
		// Set method to UPI
		params["method"] = "upi"
		// Set UPI parameters for intent flow
		params["upi"] = map[string]interface{}{
			"flow": "intent",
		}
	}
}

// createOrGetCustomer creates or gets a customer if contact is provided
func createOrGetCustomer(
	client *rzpsdk.Client,
	params map[string]interface{},
) (map[string]interface{}, error) {
	contactValue, exists := params["contact"]
	if !exists || contactValue == "" {
		return nil, nil
	}

	contact := contactValue.(string)
	customerData := map[string]interface{}{
		"contact":       contact,
		"fail_existing": "0", // Get existing customer if exists
	}

	// Create/get customer using Razorpay SDK
	customer, err := client.Customer.Create(customerData, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create/fetch customer with contact %s: %v",
			contact,
			err,
		)
	}
	return customer, nil
}

// buildPaymentData constructs the payment data for the API call
func buildPaymentData(
	params map[string]interface{},
	currency string,
	customerId string,
) *map[string]interface{} {
	paymentData := map[string]interface{}{
		"amount":   params["amount"],
		"currency": currency,
		"order_id": params["order_id"],
	}
	if customerId != "" {
		paymentData["customer_id"] = customerId
	}

	// Add token if provided (required for saved payment methods,
	// optional for UPI collect)
	if token, exists := params["token"]; exists && token != "" {
		paymentData["token"] = token
	}

	// Add contact and email parameters
	addContactAndEmailToPaymentData(paymentData, params)

	// Add additional parameters for UPI collect and other flows
	addAdditionalPaymentParameters(paymentData, params)

	// Add force_terminal_id if provided (for single block multiple debit orders)
	if terminalID, exists := params["force_terminal_id"]; exists &&
		terminalID != "" {
		paymentData["force_terminal_id"] = terminalID
	}

	return &paymentData
}

// processPaymentResult processes the payment creation result
func processPaymentResult(
	payment map[string]interface{},
	params map[string]interface{},
) (map[string]interface{}, error) {
	// Extract payment ID and next actions from the response
	paymentID := extractPaymentID(payment)
	actions := extractNextActions(payment)

	// Add token query parameter to authenticate url for amazopay wallet payments
	if wallet, ok := params["wallet"].(string); ok && wallet == "amazonpay" {
		for i, action := range actions {
			if actionType, exists := action["action"]; exists && actionType == "authenticate" {
				if authURL, exists := action["url"]; exists && authURL != nil {
					if token, exists := params["token"]; exists && token != nil {
						tokenStr := token.(string)
						tokenID := strings.TrimPrefix(tokenStr, "token_")

						// Add new field with token query param (keep original URL intact)
						fullURL := fmt.Sprintf("%s?token=%s", authURL.(string), tokenID)
						actions[i]["url_with_token"] = fullURL
					}
				}
				break
			}
		}
	}

	// Build structured response using the helper function
	response, otpUrl := buildInitiatePaymentResponse(payment, paymentID, actions)

	// Only send OTP if there's an OTP URL
	if otpUrl != "" {
		err := sendOtp(otpUrl)
		if err != nil {
			return nil, fmt.Errorf("OTP generation failed: %s", err.Error())
		}
	}

	return response, nil
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
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description("Customer ID for the payment. "+
				"Must start with 'cust_'"),
		),
		mcpgo.WithBoolean(
			"save",
			mcpgo.Description("Whether to save the payment method for future use"),
		),
		mcpgo.WithString(
			"vpa",
			mcpgo.Description("Virtual Payment Address (VPA) for UPI payment. "+
				"When provided, automatically sets method='upi' and UPI parameters "+
				"with flow='collect' and expiry_time='6' (e.g., '9876543210@ptsbi')"),
		),
		mcpgo.WithBoolean(
			"upi_intent",
			mcpgo.Description("Enable UPI intent flow. "+
				"When set to true, automatically sets method='upi' and UPI parameters "+
				"with flow='intent'. The API will return a UPI URL in the response."),
		),
		mcpgo.WithBoolean(
			"recurring",
			mcpgo.Description("Set this to true for recurring payments like "+
				"single block multiple debit."),
		),
		mcpgo.WithString(
			"force_terminal_id",
			mcpgo.Description("Terminal ID to be passed in case of single block "+
				"multiple debit order."),
		),
		mcpgo.WithString(
			"wallet",
			mcpgo.Description("Wallet provider for wallet payments "+
				"(e.g., 'amazonpay', 'phonepe', 'paytm'). "+
				"Use with method='wallet' for wallet-based payments."),
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
			ValidateAndAddOptionalString(params, "token").
			ValidateAndAddRequiredString(params, "order_id").
			ValidateAndAddOptionalString(params, "email").
			ValidateAndAddOptionalString(params, "contact").
			ValidateAndAddOptionalString(params, "customer_id").
			ValidateAndAddOptionalBool(params, "save").
			ValidateAndAddOptionalString(params, "vpa").
			ValidateAndAddOptionalBool(params, "upi_intent").
			ValidateAndAddOptionalBool(params, "recurring").
			ValidateAndAddOptionalString(params, "force_terminal_id").
			ValidateAndAddOptionalString(params, "wallet")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Set default currency
		currency := "INR"
		if c, exists := params["currency"]; exists && c != "" {
			currency = c.(string)
		}

		if params["method"] == "wallet" && params["wallet"] == "amazonpay" {
			balancesURL := fmt.Sprintf("/%s/customers/%s/balances",
				constants.VERSION_V1, params["customer_id"])

			queryParams := map[string]interface{}{"wallet[]": "amazonpay"}

			balancesResponse, err := client.Request.Get(balancesURL, queryParams, nil)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("failed to fetch wallet balances: %s", err.Error())), nil
			}
			if int(balancesResponse["items"].([]interface{})[0].(map[string]interface{})["amount"].(float64))*100 < params["amount"].(int) {
				return mcpgo.NewToolResultError(fmt.Sprintf("wallet balance is less than the payment amount: %s, please recharge amazopay wallet", balancesResponse["items"].([]interface{})[0].(map[string]interface{})["amount"])), nil
			}
		}

		// Process UPI parameters (VPA for collect flow, upi_intent for intent flow)
		processUPIParameters(params)

		// Handle customer ID
		var customerID string
		if custID, exists := params["customer_id"]; exists && custID != "" {
			customerID = custID.(string)
		} else {
			// Create or get customer if contact is provided
			customer, err := createOrGetCustomer(client, params)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			if customer != nil {
				if id, ok := customer["id"].(string); ok {
					customerID = id
				}
			}
		}

		// Build payment data
		paymentDataPtr := buildPaymentData(params, currency, customerID)
		paymentData := *paymentDataPtr

		// Create payment using Razorpay SDK's CreatePaymentJson method
		// This follows the S2S JSON v1 flow:
		// https://api.razorpay.com/v1/payments/create/json
		payment, err := client.Payment.CreatePaymentJson(paymentData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("initiating payment failed: %s", err.Error())), nil
		}

		// Process payment result
		response, err := processPaymentResult(payment, params)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"initiate_payment",
		"Initiate a payment using the S2S JSON v1 flow. "+
			"Required parameters: amount and order_id. "+
			"For saved payment methods, provide token. "+
			"For UPI collect flow, provide 'vpa' parameter "+
			"which automatically sets UPI with flow='collect' and expiry_time='6'. "+
			"For UPI intent flow, set 'upi_intent=true' parameter "+
			"which automatically sets UPI with flow='intent' and API returns UPI URL. "+
			"For wallet payments, provide 'wallet' parameter (e.g., 'amazonpay') "+
			"Supports additional parameters like customer_id, email, "+
			"contact, save, and recurring. "+
			"Returns payment details including next action steps if required.",
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
			"payment_id": paymentID,
			"status":     "success",
			"message": "OTP sent successfully. Please enter the OTP received on your " +
				"mobile number to complete the payment.",
			"response_data": otpResponse,
		}

		// Add next step instructions if OTP submit URL is available
		if otpSubmitURL != "" {
			response["otp_submit_url"] = otpSubmitURL
			response["next_step"] = "Use 'submit_otp' tool with the OTP code received " +
				"from user to complete payment authentication."
			response["next_tool"] = "submit_otp"
			response["next_tool_params"] = map[string]interface{}{
				"payment_id": paymentID,
				"otp_string": "{OTP_CODE_FROM_USER}",
			}
		} else {
			response["next_step"] = "Use 'submit_otp' tool with the OTP code received " +
				"from user to complete payment authentication."
			response["next_tool"] = "submit_otp"
			response["next_tool_params"] = map[string]interface{}{
				"payment_id": paymentID,
				"otp_string": "{OTP_CODE_FROM_USER}",
			}
		}

		result, err := mcpgo.NewToolResultJSON(response)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("JSON marshal error: %v", err)), nil
		}
		return result, nil
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
		result, err := mcpgo.NewToolResultJSON(response)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("JSON marshal error: %v", err)), nil
		}
		return result, nil
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
			if urlStr, ok := submitURL.(string); ok {
				return urlStr
			}
		}
	}

	return ""
}
