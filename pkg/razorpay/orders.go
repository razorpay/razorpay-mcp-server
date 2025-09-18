package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// CreateOrder returns a tool that creates new orders in Razorpay
func CreateOrder(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Payment amount in the smallest currency sub-unit. "+
				"REQUIRED for all orders. Examples: for ₹295, use 29500; "+
				"for $2.95, use 295. Minimum amount is 100 (₹1.00 for INR). "+
				"For mandate orders, this is typically the same as token.max_amount."),
			mcpgo.Required(),
			mcpgo.Min(100), // Minimum amount is 100 (1.00 in currency)
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("ISO 4217 currency code. REQUIRED for all orders. "+
				"Must be 3 uppercase letters. Examples: 'INR' for Indian Rupees, "+
				"'USD' for US Dollars, 'SGD' for Singapore Dollars. "+
				"For mandate orders in India, typically use 'INR'."),
			mcpgo.Required(),
			mcpgo.Pattern("^[A-Z]{3}$"), // ISO currency codes are 3 uppercase letters
		),
		mcpgo.WithString(
			"receipt",
			mcpgo.Description("Optional receipt number for internal reference. "+
				"Must be unique across all orders. Maximum 40 characters. "+
				"Example: 'Receipt No. 1', 'ORDER_123', 'INV-2024-001'."),
			mcpgo.Max(40),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Optional key-value pairs for additional information. "+
				"Maximum 15 pairs, each key and value limited to 256 characters. "+
				"Example: {\"customer_name\": \"John Doe\", "+
				"\"product\": \"Premium Plan\", "+
				"\"notes_key_1\": \"Tea, Earl Grey, Hot\"}."),
			mcpgo.MaxProperties(15),
		),
		mcpgo.WithBoolean(
			"partial_payment",
			mcpgo.Description("Whether the customer can make partial payments"),
			mcpgo.DefaultValue(false),
		),
		mcpgo.WithNumber(
			"first_payment_min_amount",
			mcpgo.Description("Minimum amount for first partial "+
				"payment (only if partial_payment is true)"),
			mcpgo.Min(100),
		),
		mcpgo.WithArray(
			"transfers",
			mcpgo.Description("Array of transfer objects for distributing "+
				"payment amounts among multiple linked accounts. Each transfer "+
				"object should contain: account (linked account ID), amount "+
				"(in currency subunits), currency (ISO code), and optional fields "+
				"like notes, linked_account_notes, on_hold, on_hold_until"),
		),
		mcpgo.WithString(
			"method",
			mcpgo.Description("Payment method for mandate orders. "+
				"REQUIRED for mandate orders. Must be 'upi' when using "+
				"token.type='single_block_multiple_debit'. This field is used "+
				"only for mandate/recurring payment orders."),
		),
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description("Customer ID for mandate orders. "+
				"REQUIRED for mandate orders. Must start with 'cust_' followed by "+
				"alphanumeric characters. Example: 'cust_RFqtp3IMdXrZP7'. "+
				"This identifies the customer for recurring payments."),
		),
		mcpgo.WithObject(
			"token",
			mcpgo.Description("Token object for mandate orders. "+
				"REQUIRED for mandate orders. Must contain: max_amount "+
				"(positive number, maximum debit amount), frequency "+
				"(as_presented/monthly/one_time/yearly/weekly/daily), "+
				"type='single_block_multiple_debit' (only supported type), "+
				"and optionally expire_at (Unix timestamp, defaults to today+60days). "+
				"Example: {\"max_amount\": 100, \"frequency\": \"as_presented\", "+
				"\"type\": \"single_block_multiple_debit\"}"),
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
			ValidateAndAddRequiredFloat(payload, "amount").
			ValidateAndAddRequiredString(payload, "currency").
			ValidateAndAddOptionalString(payload, "receipt").
			ValidateAndAddOptionalMap(payload, "notes").
			ValidateAndAddOptionalBool(payload, "partial_payment").
			ValidateAndAddOptionalArray(payload, "transfers").
			ValidateAndAddOptionalString(payload, "method").
			ValidateAndAddOptionalString(payload, "customer_id").
			ValidateAndAddToken(payload, "token")

		// Add first_payment_min_amount only if partial_payment is true
		if payload["partial_payment"] == true {
			validator.ValidateAndAddOptionalFloat(payload, "first_payment_min_amount")
		}

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		order, err := client.Order.Create(payload, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating order failed: %s", err.Error()),
			), nil
		}

		return mcpgo.NewToolResultJSON(order)
	}

	return mcpgo.NewTool(
		"create_order",
		"Create a new order in Razorpay. Supports both regular orders and "+
			"mandate orders. "+
			"\n\nFor REGULAR ORDERS: Provide amount, currency, and optional "+
			"receipt/notes. "+
			"\n\nFor MANDATE ORDERS (recurring payments): You MUST provide ALL "+
			"of these fields: "+
			"amount, currency, method='upi', customer_id (starts with 'cust_'), "+
			"and token object. "+
			"\n\nThe token object is required for mandate orders and must contain: "+
			"max_amount (positive number), frequency "+
			"(as_presented/monthly/one_time/yearly/weekly/daily), "+
			"type='single_block_multiple_debit', and optionally expire_at "+
			"(defaults to today+60days). "+
			"\n\nIMPORTANT: When token.type is 'single_block_multiple_debit', "+
			"the method MUST be 'upi'. "+
			"\n\nExample mandate order payload: "+
			`{"amount": 100, "currency": "INR", "method": "upi", `+
			`"customer_id": "cust_abc123", `+
			`"token": {"max_amount": 100, "frequency": "as_presented", `+
			`"type": "single_block_multiple_debit"}, `+
			`"receipt": "Receipt No. 1", "notes": {"key": "value"}}`,
		parameters,
		handler,
	)
}

// FetchOrder returns a tool to fetch order details by ID
func FetchOrder(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"order_id",
			mcpgo.Description("Unique identifier of the order to be retrieved"),
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
			ValidateAndAddRequiredString(payload, "order_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		order, err := client.Order.Fetch(payload["order_id"].(string), nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching order failed: %s", err.Error()),
			), nil
		}

		return mcpgo.NewToolResultJSON(order)
	}

	return mcpgo.NewTool(
		"fetch_order",
		"Fetch an order's details using its ID",
		parameters,
		handler,
	)
}

// FetchAllOrders returns a tool to fetch all orders with optional filtering
func FetchAllOrders(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"count",
			mcpgo.Description("Number of orders to be fetched "+
				"(default: 10, max: 100)"),
			mcpgo.Min(1),
			mcpgo.Max(100),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description("Number of orders to be skipped (default: 0)"),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"from",
			mcpgo.Description("Timestamp (in Unix format) from when "+
				"the orders should be fetched"),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"to",
			mcpgo.Description("Timestamp (in Unix format) up till "+
				"when orders are to be fetched"),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"authorized",
			mcpgo.Description("Filter orders based on payment authorization status. "+
				"Values: 0 (orders with unauthorized payments), "+
				"1 (orders with authorized payments)"),
			mcpgo.Min(0),
			mcpgo.Max(1),
		),
		mcpgo.WithString(
			"receipt",
			mcpgo.Description("Filter orders that contain the "+
				"provided value for receipt"),
		),
		mcpgo.WithArray(
			"expand",
			mcpgo.Description("Used to retrieve additional information. "+
				"Supported values: payments, payments.card, transfers, virtual_account"),
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

		queryParams := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddPagination(queryParams).
			ValidateAndAddOptionalInt(queryParams, "from").
			ValidateAndAddOptionalInt(queryParams, "to").
			ValidateAndAddOptionalInt(queryParams, "authorized").
			ValidateAndAddOptionalString(queryParams, "receipt").
			ValidateAndAddExpand(queryParams)

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		orders, err := client.Order.All(queryParams, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching orders failed: %s", err.Error()),
			), nil
		}

		return mcpgo.NewToolResultJSON(orders)
	}

	return mcpgo.NewTool(
		"fetch_all_orders",
		"Fetch all orders with optional filtering and pagination",
		parameters,
		handler,
	)
}

// FetchOrderPayments returns a tool to fetch all payments for a specific order
func FetchOrderPayments(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"order_id",
			mcpgo.Description(
				"Unique identifier of the order for which payments should"+
					" be retrieved. Order id should start with `order_`"),
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

		orderPaymentsReq := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(orderPaymentsReq, "order_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Fetch payments for the order using Razorpay SDK
		// Note: Using the Order.Payments method from SDK
		orderID := orderPaymentsReq["order_id"].(string)
		payments, err := client.Order.Payments(orderID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"fetching payments for order failed: %s",
					err.Error(),
				),
			), nil
		}

		// Return the result as JSON
		return mcpgo.NewToolResultJSON(payments)
	}

	return mcpgo.NewTool(
		"fetch_order_payments",
		"Fetch all payments made for a specific order in Razorpay",
		parameters,
		handler,
	)
}

// UpdateOrder returns a tool to update an order
// only the order's notes can be updated
func UpdateOrder(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"order_id",
			mcpgo.Description("Unique identifier of the order which "+
				"needs to be updated. ID should have an order_ prefix."),
			mcpgo.Required(),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs used to store additional "+
				"information about the order. A maximum of 15 key-value pairs "+
				"can be included, with each value not exceeding 256 characters."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		orderUpdateReq := make(map[string]interface{})
		data := make(map[string]interface{})

		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(orderUpdateReq, "order_id").
			ValidateAndAddRequiredMap(orderUpdateReq, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		data["notes"] = orderUpdateReq["notes"]
		orderID := orderUpdateReq["order_id"].(string)

		order, err := client.Order.Update(orderID, data, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("updating order failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(order)
	}

	return mcpgo.NewTool(
		"update_order",
		"Use this tool to update the notes for a specific order. "+
			"Only the notes field can be modified.",
		parameters,
		handler,
	)
}
