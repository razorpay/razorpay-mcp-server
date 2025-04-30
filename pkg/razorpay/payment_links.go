package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// CreatePaymentLink returns a tool that creates payment links in Razorpay
func CreatePaymentLink(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Amount to be paid using the link in smallest "+
				"currency unit(e.g., ₹300, use 30000)"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("Three-letter ISO code for the currency (e.g., INR)"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("A brief description of the Payment Link "+
				"explaining the intent of the payment."),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		validationErrors := NewValidationErrors()

		amount, err := RequiredInt(r, "amount")
		if err != nil {
			validationErrors.AddErrors(err)
		}

		currency, err := RequiredParam[string](r, "currency")
		if err != nil {
			validationErrors.AddErrors(err)
		}

		desc, err := OptionalParam[string](r, "description")
		if err != nil {
			validationErrors.AddErrors(err)
		}

		if validationErrors.HasErrors() {
			return HandleValidationErrors(validationErrors)
		}

		paymentLinkData := map[string]interface{}{
			"amount":   amount,
			"currency": currency,
		}

		if desc != "" {
			paymentLinkData["description"] = desc
		}

		paymentLink, err := client.PaymentLink.Create(paymentLinkData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating payment link failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"create_payment_link",
		"Create a new payment link in Razorpay with a specified amount",
		parameters,
		handler,
	)
}

// FetchPaymentLink returns a tool that fetches payment link details using
// payment_link_id
func FetchPaymentLink(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_link_id",
			mcpgo.Description("ID of the payment link to be fetched"+
				"(ID should have a plink_ prefix)."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		validationErrors := NewValidationErrors()

		id, err := RequiredParam[string](r, "payment_link_id")
		if err != nil {
			validationErrors.AddErrors(err)
		}

		if validationErrors.HasErrors() {
			return HandleValidationErrors(validationErrors)
		}

		paymentLink, err := client.PaymentLink.Fetch(id, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payment link failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"fetch_payment_link",
		"Fetch payment link details using it's ID."+
			"Response contains the basic details like amount, status etc",
		parameters,
		handler,
	)
}
