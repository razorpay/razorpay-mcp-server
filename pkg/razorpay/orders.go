package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// CreateOrder returns a tool that creates new orders in Razorpay
func CreateOrder(
	_ *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Payment amount in the smallest "+
				"currency sub-unit (e.g., for ₹295, use 29500)"),
			mcpgo.Required(),
			mcpgo.Min(100), // Minimum amount is 100 (1.00 in currency)
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("ISO code for the currency "+
				"(e.g., INR, USD, SGD)"),
			mcpgo.Required(),
			mcpgo.Pattern("^[A-Z]{3}$"), // ISO currency codes are 3 uppercase letters
		),
		mcpgo.WithString(
			"receipt",
			mcpgo.Description("Receipt number for internal "+
				"reference (max 40 chars, must be unique)"),
			mcpgo.Max(40),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for additional "+
				"information (max 15 pairs, 256 chars each)"),
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
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		validationErrors := NewValidationErrors()

		amount, err := RequiredInt(r, "amount")
		if err != nil {
			validationErrors.AddErrors(err)
		}

		currency, err := RequiredParam[string](r, "currency")
		if err != nil {
			validationErrors.AddErrors(err)
		}

		receipt, err := OptionalParam[string](r, "receipt")
		if err != nil {
			validationErrors.AddErrors(err)
		}

		notes, err := OptionalParam[map[string]interface{}](r, "notes")
		if err != nil {
			validationErrors.AddErrors(err)
		}

		partialPayment, err := OptionalParam[bool](r, "partial_payment")
		if err != nil {
			validationErrors.AddErrors(err)
		}

		var minAmount int
		if partialPayment {
			minAmount, err = OptionalInt(r, "first_payment_min_amount")
			if err != nil {
				validationErrors.AddErrors(err)
			}
		}

		if validationErrors.HasErrors() {
			return HandleValidationErrors(validationErrors)
		}

		orderData := map[string]interface{}{
			"amount":   amount,
			"currency": currency,
		}

		if receipt != "" {
			orderData["receipt"] = receipt
		}

		if notes != nil {
			orderData["notes"] = notes
		}

		if partialPayment {
			orderData["partial_payment"] = partialPayment

			if minAmount > 0 {
				orderData["first_payment_min_amount"] = int(minAmount)
			}
		}

		order, err := client.Order.Create(orderData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating order failed: %s", err.Error()),
			), nil
		}

		return mcpgo.NewToolResultJSON(order)
	}

	return mcpgo.NewTool(
		"create_order",
		"Create a new order in Razorpay",
		parameters,
		handler,
	)
}

// FetchOrder returns a tool to fetch order details by ID
func FetchOrder(
	_ *slog.Logger,
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
		validationErrors := NewValidationErrors()

		orderID, err := RequiredParam[string](r, "order_id")
		if err != nil {
			validationErrors.AddErrors(err)
		}

		if validationErrors.HasErrors() {
			return HandleValidationErrors(validationErrors)
		}

		order, err := client.Order.Fetch(orderID, nil, nil)
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
	_ *slog.Logger,
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
		validationErrors := NewValidationErrors()
		options := make(map[string]interface{})

		// Add pagination parameters (count and skip)
		validationErrors.AddErrors(
			AddPaginationToQueryParams(r, options)...)

		// Validate from and to parameters
		from, err := OptionalInt(r, "from")
		if err != nil {
			validationErrors.AddErrors(err)
		} else if from > 0 {
			options["from"] = int(from)
		}

		to, err := OptionalInt(r, "to")
		if err != nil {
			validationErrors.AddErrors(err)
		} else if to > 0 {
			options["to"] = int(to)
		}

		// Validate authorized parameter
		authorized, err := OptionalInt(r, "authorized")
		if err != nil {
			validationErrors.AddErrors(err)
		} else {
			options["authorized"] = int(authorized)
		}

		// Validate receipt parameter
		receipt, err := OptionalParam[string](r, "receipt")
		if err != nil {
			validationErrors.AddErrors(err)
		} else if receipt != "" {
			options["receipt"] = receipt
		}

		// Add expand parameters
		validationErrors.AddErrors(
			AddExpandToQueryParams(r, options)...)

		if validationErrors.HasErrors() {
			return HandleValidationErrors(validationErrors)
		}

		orders, err := client.Order.All(options, nil)
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
