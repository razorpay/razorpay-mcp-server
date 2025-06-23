package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// FetchPayoutByID returns a tool that fetches a payout by its ID
func FetchPayout(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payout_id",
			mcpgo.Description(
				"The unique identifier of the payout. For example, 'pout_00000000000001'",
			),
			mcpgo.Required(),
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

		FetchPayoutOptions := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(FetchPayoutOptions, "payout_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		payout, err := client.Payout.Fetch(
			FetchPayoutOptions["payout_id"].(string),
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
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"account_number",
			mcpgo.Description("The account from which the payouts were done."+
				"For example, 7878780080316316"),
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"count",
			mcpgo.Description("Number of payouts to be fetched. Default value is 10."+
				"Maximum value is 100. This can be used for pagination,"+
				"in combination with the skip parameter"),
			mcpgo.Min(1),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description("Numbers of payouts to be skipped. Default value is 0."+
				"This can be used for pagination, in combination with count"),
			mcpgo.Min(0),
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

		FetchAllPayoutsOptions := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(FetchAllPayoutsOptions, "account_number").
			ValidateAndAddPagination(FetchAllPayoutsOptions)

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		payout, err := client.Payout.All(FetchAllPayoutsOptions, nil)
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
