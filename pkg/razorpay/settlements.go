package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// FetchSettlement returns a tool that fetches a settlement by ID
func FetchSettlement(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"settlement_id",
			mcpgo.Description("The ID of the settlement to fetch."+
				"ID starts with the 'setl_'"),
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

		// Create a parameters map to collect validated parameters
		fetchSettlementOptions := make(map[string]interface{})

		// Validate using fluent validator
		validator := NewValidator(&r).
			ValidateAndAddRequiredString(fetchSettlementOptions, "settlement_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		settlementID := fetchSettlementOptions["settlement_id"].(string)
		settlement, err := client.Settlement.Fetch(settlementID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching settlement failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(settlement)
	}

	return mcpgo.NewTool(
		"fetch_settlement_with_id",
		"Fetch details of a specific settlement using its ID",
		parameters,
		handler,
	)
}

// FetchSettlementRecon returns a tool that fetches settlement
// reconciliation reports
func FetchSettlementRecon(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"year",
			mcpgo.Description("Year for which the settlement report is "+
				"requested (YYYY format)"),
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"month",
			mcpgo.Description("Month for which the settlement report is "+
				"requested (MM format)"),
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"day",
			mcpgo.Description("Optional: Day for which the settlement report is "+
				"requested (DD format)"),
		),
		mcpgo.WithNumber(
			"count",
			mcpgo.Description("Optional: Number of records to fetch "+
				"(default: 10, max: 100)"),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description("Optional: Number of records to skip for pagination"),
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

		// Create a parameters map to collect validated parameters
		fetchReconOptions := make(map[string]interface{})

		// Validate using fluent validator
		validator := NewValidator(&r).
			ValidateAndAddRequiredInt(fetchReconOptions, "year").
			ValidateAndAddRequiredInt(fetchReconOptions, "month").
			ValidateAndAddOptionalInt(fetchReconOptions, "day").
			ValidateAndAddPagination(fetchReconOptions)

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		report, err := client.Settlement.Reports(fetchReconOptions, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching settlement reconciliation report failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(report)
	}

	return mcpgo.NewTool(
		"fetch_settlement_recon_details",
		"Fetch settlement reconciliation report for a specific time period",
		parameters,
		handler,
	)
}

// FetchAllSettlements returns a tool to fetch multiple settlements with
// filtering and pagination
func FetchAllSettlements(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		// Pagination parameters
		mcpgo.WithNumber(
			"count",
			mcpgo.Description("Number of settlement records to fetch "+
				"(default: 10, max: 100)"),
			mcpgo.Min(1),
			mcpgo.Max(100),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description("Number of settlement records to skip (default: 0)"),
			mcpgo.Min(0),
		),
		// Time range filters
		mcpgo.WithNumber(
			"from",
			mcpgo.Description("Unix timestamp (in seconds) from when "+
				"settlements are to be fetched"),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"to",
			mcpgo.Description("Unix timestamp (in seconds) up till when "+
				"settlements are to be fetched"),
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

		// Create parameters map to collect validated parameters
		fetchAllSettlementsOptions := make(map[string]interface{})

		// Validate using fluent validator
		validator := NewValidator(&r).
			ValidateAndAddPagination(fetchAllSettlementsOptions).
			ValidateAndAddOptionalInt(fetchAllSettlementsOptions, "from").
			ValidateAndAddOptionalInt(fetchAllSettlementsOptions, "to")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Fetch all settlements using Razorpay SDK
		settlements, err := client.Settlement.All(fetchAllSettlementsOptions, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching settlements failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(settlements)
	}

	return mcpgo.NewTool(
		"fetch_all_settlements",
		"Fetch all settlements with optional filtering and pagination",
		parameters,
		handler,
	)
}

// CreateInstantSettlement returns a tool that creates an instant settlement
func CreateInstantSettlement(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("The amount you want to get settled instantly in amount in the smallest "+ //nolint:lll
				"currency sub-unit (e.g., for ₹295, use 29500)"),
			mcpgo.Required(),
			mcpgo.Min(200), // Minimum amount is 200 (₹2)
		),
		mcpgo.WithBoolean(
			"settle_full_balance",
			mcpgo.Description("If true, Razorpay will settle the maximum amount "+
				"possible and ignore amount parameter"),
			mcpgo.DefaultValue(false),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("Custom note for the instant settlement."),
			mcpgo.Max(30),
			mcpgo.Pattern("^[a-zA-Z0-9 ]*$"),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for additional information. "+
				"Max 15 pairs, 256 chars each"),
			mcpgo.MaxProperties(15),
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

		// Create parameters map to collect validated parameters
		createInstantSettlementReq := make(map[string]interface{})

		// Validate using fluent validator
		validator := NewValidator(&r).
			ValidateAndAddRequiredInt(createInstantSettlementReq, "amount").
			ValidateAndAddOptionalBool(createInstantSettlementReq, "settle_full_balance"). // nolint:lll
			ValidateAndAddOptionalString(createInstantSettlementReq, "description").
			ValidateAndAddOptionalMap(createInstantSettlementReq, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Create the instant settlement
		settlement, err := client.Settlement.CreateOnDemandSettlement(
			createInstantSettlementReq, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating instant settlement failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(settlement)
	}

	return mcpgo.NewTool(
		"create_instant_settlement",
		"Create an instant settlement to get funds transferred to your bank account", // nolint:lll
		parameters,
		handler,
	)
}

// FetchAllInstantSettlements returns a tool to fetch all instant settlements
// with filtering and pagination
func FetchAllInstantSettlements(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		// Pagination parameters
		mcpgo.WithNumber(
			"count",
			mcpgo.Description("Number of instant settlement records to fetch "+
				"(default: 10, max: 100)"),
			mcpgo.Min(1),
			mcpgo.Max(100),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description("Number of instant settlement records to skip (default: 0)"), //nolint:lll
			mcpgo.Min(0),
		),
		// Time range filters
		mcpgo.WithNumber(
			"from",
			mcpgo.Description("Unix timestamp (in seconds) from when "+
				"instant settlements are to be fetched"),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"to",
			mcpgo.Description("Unix timestamp (in seconds) up till when "+
				"instant settlements are to be fetched"),
			mcpgo.Min(0),
		),
		// Expand parameter for payout details
		mcpgo.WithArray(
			"expand",
			mcpgo.Description("Pass this if you want to fetch payout details "+
				"as part of the response for all instant settlements. "+
				"Supported values: ondemand_payouts"),
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

		// Create parameters map to collect validated parameters
		options := make(map[string]interface{})

		// Validate using fluent validator
		validator := NewValidator(&r).
			ValidateAndAddPagination(options).
			ValidateAndAddExpand(options).
			ValidateAndAddOptionalInt(options, "from").
			ValidateAndAddOptionalInt(options, "to")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Fetch all instant settlements using Razorpay SDK
		settlements, err := client.Settlement.FetchAllOnDemandSettlement(options, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching instant settlements failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(settlements)
	}

	return mcpgo.NewTool(
		"fetch_all_instant_settlements",
		"Fetch all instant settlements with optional filtering, pagination, and payout details", //nolint:lll
		parameters,
		handler,
	)
}

// FetchInstantSettlement returns a tool that fetches instant settlement by ID
func FetchInstantSettlement(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"settlement_id",
			mcpgo.Description("The ID of the instant settlement to fetch. "+
				"ID starts with 'setlod_'"),
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

		// Create parameters map to collect validated parameters
		params := make(map[string]interface{})

		// Validate using fluent validator
		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "settlement_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		settlementID := params["settlement_id"].(string)

		// Fetch the instant settlement by ID using SDK
		settlement, err := client.Settlement.FetchOnDemandSettlementById(
			settlementID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching instant settlement failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(settlement)
	}

	return mcpgo.NewTool(
		"fetch_instant_settlement_with_id",
		"Fetch details of a specific instant settlement using its ID",
		parameters,
		handler,
	)
}
