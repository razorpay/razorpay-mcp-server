package razorpay

import (
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/toolsets"
)

func NewToolSets(
	log *slog.Logger,
	client *rzpsdk.Client,
	enabledToolsets []string,
	readOnly bool,
) (*toolsets.ToolsetGroup, error) {
	// Create a new toolset group
	toolsetGroup := toolsets.NewToolsetGroup(readOnly)

	// Create toolsets
	payments := toolsets.NewToolset("payments", "Razorpay Payments related tools").
		AddReadTools(
			FetchPayment(log, client.Payment),
		)

	paymentLinks := toolsets.NewToolset(
		"payment_links",
		"Razorpay Payment Links related tools").
		AddReadTools(
			FetchPaymentLink(log, client.PaymentLink),
		).
		AddWriteTools(
			CreatePaymentLink(log, client.PaymentLink),
		)

	orders := toolsets.NewToolset("orders", "Razorpay Orders related tools").
		AddReadTools(
			FetchOrder(log, client.Order),
		).
		AddWriteTools(
			CreateOrder(log, client.Order),
		)

	// Add toolsets to the group
	toolsetGroup.AddToolset(payments)
	toolsetGroup.AddToolset(paymentLinks)
	toolsetGroup.AddToolset(orders)

	// Enable the requested features
	if err := toolsetGroup.EnableToolsets(enabledToolsets); err != nil {
		return nil, err
	}

	return toolsetGroup, nil
}
