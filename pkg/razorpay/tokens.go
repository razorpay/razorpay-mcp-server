package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// FetchSavedPaymentMethods returns a tool that fetches saved cards
// using contact number
func FetchSavedPaymentMethods(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"contact",
			mcpgo.Description(
				"Contact number of the customer to fetch all saved payment methods for. "+
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
		customerData := map[string]interface{}{
			"contact":       contact,
			"fail_existing": "0", // Get existing customer if exists
		}

		// Create/get customer using Razorpay SDK
		customer, err := client.Customer.Create(customerData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"Failed to create/fetch customer with contact %s: %v", contact, err,
				)), nil
		}

		customerID, ok := customer["id"].(string)
		if !ok {
			return mcpgo.NewToolResultError("Customer ID not found in response"), nil
		}

		tokensURL := fmt.Sprintf("/%s/customers/%s/tokens",
			constants.VERSION_V1, customerID)

		// Make the API request to get tokens
		tokensResponse, err := client.Request.Get(tokensURL, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"Failed to fetch saved payment methods for customer %s: %v",
					customerID,
					err,
				)), nil
		}

		// Build the base URL for balances
		balancesURL := fmt.Sprintf("/%s/customers/%s/balances",
			constants.VERSION_V1, customerID)

		queryParams := map[string]interface{}{"wallet[]": "amazonpay"}

		balancesResponse, err := client.Request.Get(balancesURL, queryParams, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"Failed to fetch saved payment methods for customer %s: %v",
					customerID,
					err,
				)), nil
		}

		result := map[string]interface{}{
			"customer":              customer,
			"saved_payment_methods": tokensResponse,
			"wallet_balances":       balancesResponse,
		}
		return mcpgo.NewToolResultJSON(result)
	}

	return mcpgo.NewTool(
		"fetch_tokens",
		"Get all saved payment methods (cards, UPI)"+
			" for a contact number. "+
			"This tool first finds or creates a"+
			" customer with the given contact number, "+
			"then fetches all saved payment tokens "+
			"associated with that customer including "+
			"credit/debit cards, UPI IDs, digital wallets,"+
			" and other tokenized payment instruments.",
		parameters,
		handler,
	)
}

// RevokeToken returns a tool that revokes a saved payment token
func RevokeToken(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description(
				"Customer ID for which the token should be revoked. "+
					"Must start with 'cust_' followed by alphanumeric characters. "+
					"Example: 'cust_xxx'"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"token_id",
			mcpgo.Description(
				"Token ID of the saved payment method to be revoked. "+
					"Must start with 'token_' followed by alphanumeric characters. "+
					"Example: 'token_xxx'"),
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

		// Validate required customer_id parameter
		customerIDValue, err := extractValueGeneric[string](&r, "customer_id", true)
		if err != nil {
			validator = validator.addError(err)
		} else if customerIDValue == nil || *customerIDValue == "" {
			validator = validator.addError(
				fmt.Errorf("missing required parameter: customer_id"))
		}
		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}
		customerID := *customerIDValue

		// Validate required token_id parameter
		tokenIDValue, err := extractValueGeneric[string](&r, "token_id", true)
		if err != nil {
			validator = validator.addError(err)
		} else if tokenIDValue == nil || *tokenIDValue == "" {
			validator = validator.addError(
				fmt.Errorf("missing required parameter: token_id"))
		}
		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}
		tokenID := *tokenIDValue

		revokeURL := fmt.Sprintf(
			"/%s%s/%s/tokens/%s/cancel",
			constants.VERSION_V1,
			constants.CUSTOMER_URL,
			customerID,
			tokenID,
		)
		response, err := client.Token.Request.Put(revokeURL, nil, nil)

		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"Failed to revoke token %s for customer %s: %v",
					tokenID,
					customerID,
					err,
				)), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"revoke_token",
		"Revoke a saved payment method (token) for a customer. "+
			"This tool revokes the specified token "+
			"associated with the given customer ID. "+
			"Once revoked, the token cannot be used for future payments.",
		parameters,
		handler,
	)
}
