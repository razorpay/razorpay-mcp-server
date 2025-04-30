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
			FetchPayment(log, client),
		)

	paymentLinks := toolsets.NewToolset(
		"payment_links",
		"Razorpay Payment Links related tools").
		AddReadTools(
			FetchPaymentLink(log, client),
		).
		AddWriteTools(
			CreatePaymentLink(log, client),
		)

	orders := toolsets.NewToolset("orders", "Razorpay Orders related tools").
		AddReadTools(
			FetchOrder(log, client),
			FetchAllOrders(log, client),
		).
		AddWriteTools(
			CreateOrder(log, client),
		)

	refunds := toolsets.NewToolset("refunds", "Razorpay Refunds related tools").
		AddReadTools(
			FetchRefund(log, client),
		).
		AddWriteTools(
			CreateRefund(log, client),
			UpdateRefund(log, client),
		)

	// Add toolsets to the group
	toolsetGroup.AddToolset(payments)
	toolsetGroup.AddToolset(paymentLinks)
	toolsetGroup.AddToolset(orders)
	toolsetGroup.AddToolset(refunds)

	// Enable the requested features
	if err := toolsetGroup.EnableToolsets(enabledToolsets); err != nil {
		return nil, err
	}

	return toolsetGroup, nil
}
