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
			FetchPaymentCardDetails(log, client),
			FetchAllPayments(log, client),
		).
		AddWriteTools(
			CapturePayment(log, client),
			UpdatePayment(log, client),
		)

	paymentLinks := toolsets.NewToolset(
		"payment_links",
		"Razorpay Payment Links related tools").
		AddReadTools(
			FetchPaymentLink(log, client),
			FetchAllPaymentLinks(log, client),
		).
		AddWriteTools(
			CreatePaymentLink(log, client),
			CreateUpiPaymentLink(log, client),
			ResendPaymentLinkNotification(log, client),
			UpdatePaymentLink(log, client),
		)

	orders := toolsets.NewToolset("orders", "Razorpay Orders related tools").
		AddReadTools(
			FetchOrder(log, client),
			FetchAllOrders(log, client),
			FetchOrderPayments(log, client),
		).
		AddWriteTools(
			CreateOrder(log, client),
			UpdateOrder(log, client),
		)

	refunds := toolsets.NewToolset("refunds", "Razorpay Refunds related tools").
		AddReadTools(
			FetchRefund(log, client),
			FetchMultipleRefundsForPayment(log, client),
			FetchSpecificRefundForPayment(log, client),
			FetchAllRefunds(log, client),
		).
		AddWriteTools(
			CreateRefund(log, client),
			UpdateRefund(log, client),
		)

	qrCodes := toolsets.NewToolset("qr_codes", "Razorpay QR Codes related tools").
		AddReadTools(
			FetchQRCode(log, client),
			FetchAllQRCodes(log, client),
			FetchQRCodesByCustomerID(log, client),
			FetchQRCodesByPaymentID(log, client),
			FetchPaymentsForQRCode(log, client),
		).
		AddWriteTools(
			CreateQRCode(log, client),
			CloseQRCode(log, client),
		)

	settlements := toolsets.NewToolset("settlements", "Razorpay Settlements related tools"). // nolint:lll
													AddReadTools(
			FetchSettlement(log, client),
			FetchSettlementRecon(log, client),
			FetchAllSettlements(log, client),
			FetchAllInstantSettlements(log, client),
			FetchInstantSettlement(log, client),
		).
		AddWriteTools(
			CreateInstantSettlement(log, client),
		)

	// Add toolsets to the group
	toolsetGroup.AddToolset(payments)
	toolsetGroup.AddToolset(paymentLinks)
	toolsetGroup.AddToolset(orders)
	toolsetGroup.AddToolset(refunds)
	toolsetGroup.AddToolset(qrCodes)
	toolsetGroup.AddToolset(settlements)

	// Enable the requested features
	if err := toolsetGroup.EnableToolsets(enabledToolsets); err != nil {
		return nil, err
	}

	return toolsetGroup, nil
}
