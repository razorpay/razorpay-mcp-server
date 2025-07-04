package razorpay

import (
	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
	"github.com/razorpay/razorpay-mcp-server/pkg/toolsets"
)

func NewToolSets(
	obs *observability.Observability,
	client *rzpsdk.Client,
	enabledToolsets []string,
	readOnly bool,
) (*toolsets.ToolsetGroup, error) {
	// Create a new toolset group
	toolsetGroup := toolsets.NewToolsetGroup(readOnly)

	// Create toolsets
	payments := toolsets.NewToolset("payments", "Razorpay Payments related tools").
		AddReadTools(
			FetchPayment(obs, client),
			FetchPaymentCardDetails(obs, client),
			FetchAllPayments(obs, client),
		).
		AddWriteTools(
			CreatePaymentWallet(obs, client),
			CreatePaymentUpiCollect(obs, client),
			CreatePaymentOrder(obs, client),
			AcceptAndProcessPayments(obs, client),
			AcceptPaymentsByChat(obs, client),
			CreatePaymentsByToken(obs, client),
			OtpGenerateForPayment(obs, client),
			OtpVerifyForPayment(obs, client),
			CapturePayment(obs, client),
			UpdatePayment(obs, client),
		)

	paymentLinks := toolsets.NewToolset(
		"payment_links",
		"Razorpay Payment Links related tools").
		AddReadTools(
			FetchPaymentLink(obs, client),
			FetchAllPaymentLinks(obs, client),
		).
		AddWriteTools(
			CreatePaymentLink(obs, client),
			CreateUpiPaymentLink(obs, client),
			ResendPaymentLinkNotification(obs, client),
			UpdatePaymentLink(obs, client),
		)

	orders := toolsets.NewToolset("orders", "Razorpay Orders related tools").
		AddReadTools(
			FetchOrder(obs, client),
			FetchAllOrders(obs, client),
			FetchOrderPayments(obs, client),
		).
		AddWriteTools(
			CreateOrder(obs, client),
			UpdateOrder(obs, client),
		)

	// refunds := toolsets.NewToolset("refunds", "Razorpay Refunds related tools").
	// 	AddReadTools(
	// 		FetchRefund(obs, client),
	// 		FetchMultipleRefundsForPayment(obs, client),
	// 		FetchSpecificRefundForPayment(obs, client),
	// 		FetchAllRefunds(obs, client),
	// 	).
	// 	AddWriteTools(
	// 		CreateRefund(obs, client),
	// 		UpdateRefund(obs, client),
	// 	)

	// payouts := toolsets.NewToolset("payouts", "Razorpay Payouts related tools").
	// 	AddReadTools(
	// 		FetchPayout(obs, client),
	// 		FetchAllPayouts(obs, client),
	// 	)

	// qrCodes := toolsets.NewToolset("qr_codes", "Razorpay QR Codes related tools").
	// 	AddReadTools(
	// 		FetchQRCode(obs, client),
	// 		FetchAllQRCodes(obs, client),
	// 		FetchQRCodesByCustomerID(obs, client),
	// 		FetchQRCodesByPaymentID(obs, client),
	// 		FetchPaymentsForQRCode(obs, client),
	// 	).
	// 	AddWriteTools(
	// 		CreateQRCode(obs, client),
	// 		CloseQRCode(obs, client),
	// 	)

	// settlements := toolsets.NewToolset("settlements",
	// 	"Razorpay Settlements related tools").
	// 	AddReadTools(
	// 		FetchSettlement(obs, client),
	// 		FetchSettlementRecon(obs, client),
	// 		FetchAllSettlements(obs, client),
	// 		FetchAllInstantSettlements(obs, client),
	// 		FetchInstantSettlement(obs, client),
	// 	).
	// 	AddWriteTools(
	// 		CreateInstantSettlement(obs, client),
	// 	)

	tokens := toolsets.NewToolset("tokens", "Razorpay tokens related tools").
		AddReadTools(
			FetchToken(obs, client),
			FetchAllTokens(obs, client),
		)

	customers := toolsets.NewToolset("customers", "Razorpay Customers related tools"). //nolint:lll
												AddWriteTools(
			CreateCustomer(obs, client),
		)

	// Add toolsets to the group
	toolsetGroup.AddToolset(payments)
	toolsetGroup.AddToolset(paymentLinks)
	toolsetGroup.AddToolset(orders)
	//toolsetGroup.AddToolset(refunds)
	//toolsetGroup.AddToolset(payouts)
	//toolsetGroup.AddToolset(qrCodes)
	//toolsetGroup.AddToolset(settlements)

	toolsetGroup.AddToolset(tokens)
	toolsetGroup.AddToolset(customers)

	// Enable the requested features
	if err := toolsetGroup.EnableToolsets(enabledToolsets); err != nil {
		return nil, err
	}

	return toolsetGroup, nil
}
