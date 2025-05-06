package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// FetchPayoutByID returns a tool that fetches a payout by its ID
func FetchPayoutByID(
	_ *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payout_id",
			mcpgo.Description("The unique identifier of the payout to fetch"),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		validator := NewValidator(&r)
		params := make(map[string]interface{})

		// Validate required parameters
		validator.ValidateAndAddRequiredString(params, "payout_id")
		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Fetch the payout using Razorpay SDK
		result, err := client.Payout.Fetch(params["payout_id"].(string), nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payout failed: %s", err.Error()),
			), nil
		}

		toolResult, err := mcpgo.NewToolResultJSON(result)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to marshal result: %s", err.Error()),
			), nil
		}
		return toolResult, nil
	}

	return mcpgo.NewTool(
		"fetch_payout_by_id",
		"Fetch a payout's details using its ID",
		parameters,
		handler,
	)
}

// FetchAllPayouts returns a tool that fetches all payouts
func FetchAllPayouts(
	_ *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"account_number",
			mcpgo.Description("The account number to fetch payouts for"),
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"count",
			mcpgo.Description("Number of payouts to fetch (default: 10)"),
			mcpgo.Min(1),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description("Number of payouts to skip (for pagination)"),
			mcpgo.Min(0),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		validator := NewValidator(&r)
		params := make(map[string]interface{})

		// Validate required and optional parameters
		validator.ValidateAndAddRequiredString(params, "account_number")
		validator.ValidateAndAddPagination(params)
		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Fetch payouts using Razorpay SDK
		result, err := client.Payout.All(params, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payouts failed: %s", err.Error()),
			), nil
		}

		toolResult, err := mcpgo.NewToolResultJSON(result)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to marshal result: %s", err.Error()),
			), nil
		}
		return toolResult, nil
	}

	return mcpgo.NewTool(
		"fetch_all_payouts",
		"Fetch all payouts for an account",
		parameters,
		handler,
	)
}
