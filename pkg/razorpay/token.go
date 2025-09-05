package razorpay

import (
	"context"
	"fmt"
	"strings"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// FetchToken returns a tool that fetches token details using customer_id
// and token_id
func FetchToken(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description("The unique identifier of the customer. "+
				"Must start with 'cust_'"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"token_id",
			mcpgo.Description("The unique identifier of the token to fetch. "+
				"Must start with 'token_'"),
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
		params := make(map[string]interface{})

		// Validate required customer_id parameter
		customerIDValue, err := extractValueGeneric[string](&r, "customer_id", true)
		if err != nil {
			validator = validator.addError(err)
		} else if customerIDValue != nil && *customerIDValue == "" {
			validator = validator.addError(
				fmt.Errorf("missing required parameter: customer_id"))
		} else if customerIDValue != nil {
			// Validate customer_id format
			if !strings.HasPrefix(*customerIDValue, "cust_") {
				validator = validator.addError(
					fmt.Errorf("customer_id must start with 'cust_'"))
			} else {
				params["customer_id"] = *customerIDValue
			}
		}

		// Validate required token_id parameter
		tokenIDValue, err := extractValueGeneric[string](&r, "token_id", true)
		if err != nil {
			validator = validator.addError(err)
		} else if tokenIDValue != nil && *tokenIDValue == "" {
			validator = validator.addError(
				fmt.Errorf("missing required parameter: token_id"))
		} else if tokenIDValue != nil {
			// Validate token_id format
			if !strings.HasPrefix(*tokenIDValue, "token_") {
				validator = validator.addError(
					fmt.Errorf("token_id must start with 'token_'"))
			} else {
				params["token_id"] = *tokenIDValue
			}
		}

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		customerID := params["customer_id"].(string)
		tokenID := params["token_id"].(string)

		// Create the API endpoint URL
		url := fmt.Sprintf("/%s/customers/%s/tokens/%s",
			constants.VERSION_V1, customerID, tokenID)

		// Make the API request
		response, err := client.Request.Get(url, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("Failed to fetch token: %v", err)), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"fetch_token",
		"Use this tool to retrieve the details of a specific token "+
			"using customer_id and token_id. Tokens are used for storing "+
			"card details securely and for recurring payments.",
		parameters,
		handler,
	)
}

// FetchAllTokens returns a tool that fetches all tokens for a customer
func FetchAllTokens(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description("The unique identifier of the customer. "+
				"Must start with 'cust_'"),
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
		params := make(map[string]interface{})

		// Validate required customer_id parameter
		customerIDValue, err := extractValueGeneric[string](&r, "customer_id", true)
		if err != nil {
			validator = validator.addError(err)
		} else if customerIDValue != nil && *customerIDValue == "" {
			validator = validator.addError(
				fmt.Errorf("missing required parameter: customer_id"))
		} else if customerIDValue != nil {
			// Validate customer_id format
			if !strings.HasPrefix(*customerIDValue, "cust_") {
				validator = validator.addError(
					fmt.Errorf("customer_id must start with 'cust_'"))
			} else {
				params["customer_id"] = *customerIDValue
			}
		}

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		customerID := params["customer_id"].(string)

		// Create the API endpoint URL
		url := fmt.Sprintf("/%s/customers/%s/tokens",
			constants.VERSION_V1, customerID)

		// Make the API request
		response, err := client.Request.Get(url, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("Failed to fetch tokens: %v", err)), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"fetch_all_tokens",
		"Use this tool to retrieve all tokens for a specific customer. "+
			"Tokens represent stored card details and can be used for "+
			"recurring payments.",
		parameters,
		handler,
	)
}
