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
// using a customer_id or contact number.
// When customer_id is provided it is used directly; otherwise
// the contact number is used to create/find the customer first.
func FetchSavedPaymentMethods(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description(
				"Razorpay customer ID to fetch saved payment methods for. "+
					"Must start with 'cust_' followed by alphanumeric characters. "+
					"Example: 'cust_xxx'. "+
					"When provided, this takes priority over the contact parameter."),
		),
		mcpgo.WithString(
			"contact",
			mcpgo.Description(
				"Contact number of the customer to fetch all saved payment methods for. "+
					"For example, 9876543210 or +919876543210. "+
					"Used only when customer_id is not provided."),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddOptionalString(params, "customer_id").
			ValidateAndAddOptionalString(params, "contact")

		hasCustomerID := false
		hasContact := false
		if v, ok := params["customer_id"].(string); ok && v != "" {
			hasCustomerID = true
		}
		if v, ok := params["contact"].(string); ok && v != "" {
			hasContact = true
		}

		// At least one non-empty identifier is required when params are valid.
		if !hasCustomerID && !hasContact && !validator.HasErrors() {
			validator = validator.addError(
				fmt.Errorf(
					"either customer_id or contact must be provided"))
		}

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		var customerID string
		var customer map[string]interface{}

		if hasCustomerID {
			customerID = params["customer_id"].(string)

			url := fmt.Sprintf("/%s%s/%s",
				constants.VERSION_V1,
				constants.CUSTOMER_URL,
				customerID)
			customer, err = client.Request.Get(url, nil, nil)
			if err != nil {
				return mcpgo.NewToolResultError(
					formatErrorMessage("fetching customer failed", err)), nil
			}
		} else {
			contact := params["contact"].(string)
			customerData := map[string]interface{}{
				"contact":       contact,
				"fail_existing": "0",
			}

			customer, err = client.Customer.Create(customerData, nil)
			if err != nil {
				return mcpgo.NewToolResultError(
					formatErrorMessage(
						"creating or fetching customer failed", err)), nil
			}

			id, ok := customer["id"].(string)
			if !ok {
				return mcpgo.NewToolResultError(
					"Customer ID not found in response"), nil
			}
			customerID = id
		}

		url := fmt.Sprintf("/%s/customers/%s/tokens",
			constants.VERSION_V1, customerID)

		tokensResponse, err := client.Request.Get(url, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage(
					"fetching saved payment methods failed", err)), nil
		}

		result := map[string]interface{}{
			"customer":              customer,
			"saved_payment_methods": tokensResponse,
		}
		return mcpgo.NewToolResultJSON(result)
	}

	return mcpgo.NewTool(
		"fetch_tokens",
		"Get all saved payment methods (cards, UPI) "+
			"for a customer. Accepts either a customer_id "+
			"(preferred) or a contact number. "+
			"When customer_id is provided it is used "+
			"directly to fetch tokens; otherwise the "+
			"contact number is used to find or create "+
			"the customer first. Returns saved payment "+
			"tokens including credit/debit cards, UPI IDs,"+
			" digital wallets, and other tokenized "+
			"payment instruments.",
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

		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "customer_id").
			ValidateAndAddRequiredString(params, "token_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		customerID := params["customer_id"].(string)
		tokenID := params["token_id"].(string)

		url := fmt.Sprintf(
			"/%s%s/%s/tokens/%s/cancel",
			constants.VERSION_V1,
			constants.CUSTOMER_URL,
			customerID,
			tokenID,
		)
		response, err := client.Token.Request.Put(url, nil, nil)

		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("revoking token failed", err)), nil
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
