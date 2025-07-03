package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// CreateCustomer returns a tool that creates a new customer in Razorpay
func CreateCustomer(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"contact",
			mcpgo.Description("Contact number (mobile number) of the customer"),
		),
		mcpgo.WithString(
			"fail_existing",
			mcpgo.Description("Set to '0' to return existing customer if already exists instead of failing"),
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

		customerData := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddOptionalString(customerData, "contact").
			ValidateAndAddOptionalString(customerData, "fail_existing")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Create customer using Razorpay SDK
		customer, err := client.Customer.Create(customerData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating customer failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(customer)
	}

	return mcpgo.NewTool(
		"create_customer",
		"Create a new customer in Razorpay with contact details and optional fail_existing flag", //nolint:lll
		parameters,
		handler,
	)
}