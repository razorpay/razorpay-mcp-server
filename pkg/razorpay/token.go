package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// FetchToken returns a tool that fetches token details using customer_id and token_id
func FetchToken(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description("The unique identifier of the customer"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"token_id",
			mcpgo.Description("The unique identifier of the token to fetch"),
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

		customerId := params["customer_id"].(string)
		tokenId := params["token_id"].(string)

		// Fetch token using Razorpay SDK
		token, err := client.Token.Fetch(customerId, tokenId, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching token failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(token)
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
			mcpgo.Description("The unique identifier of the customer"),
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
			ValidateAndAddRequiredString(params, "customer_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		customerId := params["customer_id"].(string)

		// Fetch all tokens for the customer using Razorpay SDK
		tokens, err := client.Token.All(customerId, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching tokens failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(tokens)
	}

	return mcpgo.NewTool(
		"fetch_all_tokens",
		"Use this tool to retrieve all tokens for a specific customer. "+
			"Tokens represent stored card details and can be used for recurring payments.",
		parameters,
		handler,
	)
}
