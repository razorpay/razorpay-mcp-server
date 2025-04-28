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
