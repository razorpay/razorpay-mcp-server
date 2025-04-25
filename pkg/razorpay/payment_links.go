package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"

	rzpsdk "github.com/razorpay/razorpay-go"
)

// CreatePaymentLink returns a tool that creates payment links in Razorpay
func CreatePaymentLink(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Amount to be paid using the link in smallest currency unit(e.g., ₹300, use 30000)"), // nolint:lll
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("Three-letter ISO code for the currency (e.g., INR)"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("A brief description of the Payment Link explaining the intent of the payment."), // nolint:lll
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// validate required parameters
		amount, err := RequiredInt(r, "amount")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}
		currency, err := RequiredParam[string](r, "currency")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		// Create request payload
		paymentLinkData := map[string]interface{}{
			"amount":   amount,
			"currency": currency,
		}

		// Add optional description if provided
		desc, err := OptionalParam[string](r, "description")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		if desc != "" {
			paymentLinkData["description"] = desc
		}

		paymentLink, err := client.PaymentLink.Create(paymentLinkData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("pl create failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"payment_link.create",
		"Create a new standard payment link in Razorpay with a specified amount",
		parameters,
		handler,
	)
}

// CreateUpiPaymentLink returns a tool that creates payment links in Razorpay
func CreateUpiPaymentLink(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Amount to be paid using the link in smallest currency unit(e.g., ₹300, use 30000), Only accepted currency is INR"), // nolint:lll
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("A brief description of the Payment Link explaining the intent of the payment."), // nolint:lll
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// validate required parameters
		amount, err := RequiredInt(r, "amount")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		currency := "INR"

		// Create request payload
		paymentLinkData := map[string]interface{}{
			"amount":   amount,
			"currency": currency,
			"upi_link": "true",
		}

		// Add optional description if provided
		desc, err := OptionalParam[string](r, "description")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		if desc != "" {
			paymentLinkData["description"] = desc
		}

		paymentLink, err := client.PaymentLink.Create(paymentLinkData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("upi pl create failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"payment_link_upi.create",
		"Create a new UPI payment link(Intent link) in Razorpay with a specified amount.", // nolint:lll
		parameters,
		handler,
	)
}

// FetchPaymentLink
// returns a tool that fetches payment link details using payment_link_id
func FetchPaymentLink(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_link_id",
			mcpgo.Description("ID of the payment link to be fetched(ID should have a plink_ prefix)."), // nolint:lll
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Use the helper function to get the required parameter
		id, err := RequiredParam[string](r, "payment_link_id")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		paymentLink, err := client.PaymentLink.Fetch(id, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payment link failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"payment_link.fetch",
		"Fetch payment link details of a standard payment link or a UPI Payment link using it's ID."+ // nolint:lll
			"Response contains the basic details like amount, status etc",
		parameters,
		handler,
	)
}
