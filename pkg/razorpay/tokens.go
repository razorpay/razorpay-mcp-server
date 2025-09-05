package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// FetchSavedCardsWithContact returns a tool that fetches saved cards using contact number
func FetchSavedCardsWithContact(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"contact",
			mcpgo.Description("Contact number of the customer to fetch all saved payment methods for. "+
				"For example, 9876543210 or +919876543210"),
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

		validator := NewValidator(&r)

		// Validate required contact parameter
		contactValue, err := extractValueGeneric[string](&r, "contact", true)
		if err != nil {
			validator = validator.addError(err)
		} else if contactValue == nil || *contactValue == "" {
			validator = validator.addError(
				fmt.Errorf("missing required parameter: contact"))
		}

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		contact := *contactValue

		// First, try to create a customer with fail_existing=0 to get existing customer
		customerData := map[string]interface{}{
			"name":          "Customer_" + contact, // Temporary name
			"contact":       contact,
			"fail_existing": "0", // Get existing customer if exists
		}

		// Create/get customer using Razorpay SDK
		customer, err := client.Customer.Create(customerData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("Failed to create/fetch customer with contact %s: %v", contact, err)), nil
		}

		// Extract customer ID from the response
		customerID, ok := customer["id"].(string)
		if !ok {
			return mcpgo.NewToolResultError("Customer ID not found in response"), nil
		}

		// Now fetch all tokens for this customer
		url := fmt.Sprintf("/%s/customers/%s/tokens",
			constants.VERSION_V1, customerID)

		// Make the API request to get tokens
		tokensResponse, err := client.Request.Get(url, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("Failed to fetch saved payment methods for customer %s: %v", customerID, err)), nil
		}

		// Create a combined response with customer info and all saved payment methods
		result := map[string]interface{}{
			"customer":              customer,
			"saved_payment_methods": tokensResponse,
		}

		return mcpgo.NewToolResultJSON(result)
	}

	return mcpgo.NewTool(
		"fetch_saved_cards_with_contact",
		"Get all saved payment methods (cards, UPI, wallets, etc.) for a contact number. "+
			"This tool first finds or creates a customer with the given contact number, "+
			"then fetches all saved payment tokens associated with that customer including "+
			"credit/debit cards, UPI IDs, digital wallets, and other tokenized payment instruments.",
		parameters,
		handler,
	)
}
