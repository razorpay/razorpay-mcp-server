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

// FetchSettlementRecon returns a tool that fetches settlement reconciliation reports
func FetchSettlementRecon(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"year",
			mcpgo.Description("Year for which the settlement report is requested (YYYY format)"),
			mcpgo.Required(),
			mcpgo.Pattern("^[0-9]{4}$"),
		),
		mcpgo.WithString(
			"month",
			mcpgo.Description("Month for which the settlement report is requested (MM format)"),
			mcpgo.Required(),
			mcpgo.Pattern("^[0-9]{1,2}$"),
		),
		mcpgo.WithString(
			"day",
			mcpgo.Description("Optional: Day for which the settlement report is requested (DD format)"),
			mcpgo.Pattern("^[0-9]{1,2}$"),
		),
		mcpgo.WithString(
			"count",
			mcpgo.Description("Optional: Number of records to fetch (default: 10, max: 100)"),
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
				fmt.Sprintf("fetching settlement reconciliation report failed: %s", err.Error())), nil
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
