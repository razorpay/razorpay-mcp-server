package razorpay

import (
	"context"
	"encoding/json"
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
		amount, err := RequiredInt(r, "amount")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Get and validate currency
		currency, err := RequiredParam[string](r, "currency")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Get and validate receipt (optional)
		receipt, err := OptionalParam[string](r, "receipt")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Get and validate notes (optional)
		notes, err := OptionalParam[map[string]interface{}](r, "notes")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Get and validate partial payment options
		partialPayment, err := OptionalParam[bool](r, "partial_payment")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Create request payload
		orderData := map[string]interface{}{
			"amount":   amount,
			"currency": currency,
		}

		// Add optional receipt if provided
		if receipt != "" {
			orderData["receipt"] = receipt
		}

		// Add optional notes if provided
		if notes != nil {
			orderData["notes"] = notes
		}

		// Process partial payment if enabled
		if partialPayment {
			orderData["partial_payment"] = partialPayment

			// If partial payment is enabled, validate min amount
			minAmount, err := OptionalInt(r, "first_payment_min_amount")
			if result, err := HandleValidationError(err); result != nil {
				return result, err
			}

			// Only add first_payment_min_amount if it was provided and > 0
			if minAmount > 0 {
				orderData["first_payment_min_amount"] = minAmount
			}
		}

		// Create the order using Razorpay SDK
		order, err := client.Order.Create(orderData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating order failed: %s", err.Error()),
			), nil
		}

		// Marshal the response to JSON
		b, err := json.Marshal(order)
		if err != nil {
			return mcpgo.NewToolResultError("failed to marshal result"), nil
		}

		return mcpgo.NewToolResultText(string(b)), nil
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
		orderID, err := RequiredParam[string](r, "order_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Fetch the order using Razorpay SDK
		order, err := client.Order.Fetch(orderID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching order failed: %s", err.Error()),
			), nil
		}

		// Marshal the response to JSON
		b, err := json.Marshal(order)
		if err != nil {
			return mcpgo.NewToolResultError("failed to marshal result"), nil
		}

		return mcpgo.NewToolResultText(string(b)), nil
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
		// Prepare query parameters map
		options := make(map[string]interface{})

		// Process pagination options
		if result := AddPaginationToQueryParams(r, options); result != nil {
			return result, nil
		}

		// Process date range parameters directly
		from, err := OptionalInt(r, "from")
		if result, _ := HandleValidationError(err); result != nil {
			return result, nil
		}
		if from > 0 {
			options["from"] = from
		}

		to, err := OptionalInt(r, "to")
		if result, _ := HandleValidationError(err); result != nil {
			return result, nil
		}
		if to > 0 {
			options["to"] = to
		}

		// Process authorized status parameter directly
		authorized, err := OptionalInt(r, "authorized")
		if result, _ := HandleValidationError(err); result != nil {
			return result, nil
		}

		// Always add the authorized parameter if it was provided in the request
		options["authorized"] = authorized

		// Process receipt parameter directly
		receipt, err := OptionalParam[string](r, "receipt")
		if result, _ := HandleValidationError(err); result != nil {
			return result, nil
		}
		if receipt != "" {
			options["receipt"] = receipt
		}

		// Process expand parameters
		if result := AddExpandToQueryParams(r, options); result != nil {
			return result, nil
		}

		// Fetch all orders using Razorpay SDK
		orders, err := client.Order.All(options, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching orders failed: %s", err.Error()),
			), nil
		}

		// Convert to JSON
		return mcpgo.NewToolResultJSON(orders)
	}

	return mcpgo.NewTool(
		"fetch_all_orders",
		"Fetch all orders with optional filtering and pagination",
		parameters,
		handler,
	)
}
