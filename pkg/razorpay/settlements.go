package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// FetchSettlement returns a tool that fetches a settlement by ID
func FetchSettlement(
	log *slog.Logger,
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
		settlementID, err := RequiredParam[string](r, "settlement_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

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
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"year",
			mcpgo.Description("Year for which the settlement report is "+
				"requested (YYYY format)"),
			mcpgo.Required(),
			mcpgo.Pattern("^[0-9]{4}$"),
		),
		mcpgo.WithString(
			"month",
			mcpgo.Description("Month for which the settlement report is "+
				"requested (MM format)"),
			mcpgo.Required(),
			mcpgo.Pattern("^[0-9]{1,2}$"),
		),
		mcpgo.WithString(
			"day",
			mcpgo.Description("Optional: Day for which the settlement report is "+
				"requested (DD format)"),
			mcpgo.Pattern("^[0-9]{1,2}$"),
		),
		mcpgo.WithString(
			"count",
			mcpgo.Description("Optional: Number of records to fetch "+
				"(default: 10, max: 100)"),
			mcpgo.Pattern("^[0-9]+$"),
		),
		mcpgo.WithString(
			"skip",
			mcpgo.Description("Optional: Number of records to skip for pagination"),
			mcpgo.Pattern("^[0-9]+$"),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		year, err := RequiredParam[string](r, "year")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		month, err := RequiredParam[string](r, "month")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Create query parameters
		queryParams := map[string]interface{}{
			"year":  year,
			"month": month,
		}

		// Add optional day parameter if provided
		day, err := OptionalParam[string](r, "day")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if day != "" {
			queryParams["day"] = day
		}

		// Add optional count parameter if provided
		count, err := OptionalParam[string](r, "count")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if count != "" {
			queryParams["count"] = count
		}

		// Add optional skip parameter if provided
		skip, err := OptionalParam[string](r, "skip")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if skip != "" {
			queryParams["skip"] = skip
		}

		report, err := client.Settlement.Reports(queryParams, nil)
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
	log *slog.Logger,
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
		// Create query parameters map
		options := make(map[string]interface{})

		// Handle pagination parameters
		if result := AddPaginationToQueryParams(r, options); result != nil {
			return result, nil
		}

		// Handle date range parameters
		from, err := OptionalInt(r, "from")
		if result, _ := HandleValidationError(err); result != nil {
			return result, nil
		}
		if from > 0 {
			options["from"] = from
		}

		to, err := OptionalInt(r, "to")
		if result, _ := HandleValidationError(err); result != nil {
			return result, nil
		}
		if to > 0 {
			options["to"] = to
		}

		// Fetch all settlements using Razorpay SDK
		settlements, err := client.Settlement.All(options, nil)
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
