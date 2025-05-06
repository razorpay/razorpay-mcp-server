package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// FetchPayoutByID returns a tool that fetches a payout by its ID
func FetchPayout(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payout_id",
			mcpgo.Description(
				"The unique identifier of the payout to fetch. Format: pout_ "+
					"followed by alphanumeric characters (e.g., pout_qr2726363738)",
			),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		payload := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(payload, "payout_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		payout, err := client.Payout.Fetch(
			payload["payout_id"].(string),
			nil,
			nil,
		)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payout failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(payout)
	}

	return mcpgo.NewTool(
		"fetch_payout_with_id",
		"Fetch a payout's details using its ID",
		parameters,
		handler,
	)
}

// FetchAllPayouts returns a tool that fetches all payouts
func FetchAllPayouts(
	log *slog.Logger,
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
		payload := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(payload, "account_number").
			ValidateAndAddPagination(payload)

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		payout, err := client.Payout.All(payload, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payouts failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(payout)
	}

	return mcpgo.NewTool(
		"fetch_all_payouts",
		"Fetch all payouts for a bank account number",
		parameters,
		handler,
	)
}
