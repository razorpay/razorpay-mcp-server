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

		validator := NewValidator(&r)

		customerIDValue, _ := extractValueGeneric[string](
			&r, "customer_id", false)
		contactValue, _ := extractValueGeneric[string](
			&r, "contact", false)

		hasCustomerID := customerIDValue != nil && *customerIDValue != ""
		hasContact := contactValue != nil && *contactValue != ""

		if !hasCustomerID && !hasContact {
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
			customerID = *customerIDValue

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
			contact := *contactValue
			customerData := map[string]interface{}{
				"contact":       contact,
				"fail_existing": "0",
			}

			customer, err = client.Customer.Create(customerData, nil)
			if err != nil {
				return mcpgo.NewToolResultError(
					formatErrorMessage("creating customer failed", err)), nil
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
				formatErrorMessage("fetching tokens failed", err)), nil
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
