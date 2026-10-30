package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/integrations"
	"github.com/razorpay/razorpay-mcp-server/pkg/region"
	"github.com/razorpay/razorpay-mcp-server/pkg/toolsets"
)

type toolsetConfig struct {
	profile region.Profile
}

// ToolsetOption customises NewToolSets.
type ToolsetOption func(*toolsetConfig)

// WithRegionProfile sets the region the tools are built for (default India).
func WithRegionProfile(p region.Profile) ToolsetOption {
	return func(c *toolsetConfig) {
		c.profile = p
	}
}

func NewToolSets(
	obs *observability.Observability,
	client *rzpsdk.Client,
	enabledToolsets []string,
	readOnly bool,
	opts ...ToolsetOption,
) (*toolsets.ToolsetGroup, error) {
	cfg := toolsetConfig{profile: region.India()}
	for _, opt := range opts {
		opt(&cfg)
	}

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
			CapturePayment(obs, client),
			UpdatePayment(obs, client),
			InitiatePayment(obs, client),
			ResendOtp(obs, client),
			SubmitOtp(obs, client),
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

	refunds := toolsets.NewToolset("refunds", "Razorpay Refunds related tools").
		AddReadTools(
			FetchRefund(obs, client),
			FetchMultipleRefundsForPayment(obs, client),
			FetchSpecificRefundForPayment(obs, client),
			FetchAllRefunds(obs, client),
		).
		AddWriteTools(
			CreateRefund(obs, client),
			UpdateRefund(obs, client),
		)

	payouts := toolsets.NewToolset("payouts", "Razorpay Payouts related tools").
		AddReadTools(
			FetchPayout(obs, client),
			FetchAllPayouts(obs, client),
		)

	qrCodes := toolsets.NewToolset("qr_codes", "Razorpay QR Codes related tools").
		AddReadTools(
			FetchQRCode(obs, client),
			FetchAllQRCodes(obs, client),
			FetchQRCodesByCustomerID(obs, client),
			FetchQRCodesByPaymentID(obs, client),
			FetchPaymentsForQRCode(obs, client),
		).
		AddWriteTools(
			CreateQRCode(obs, client),
			CloseQRCode(obs, client),
		)

	registrationLinks := toolsets.NewToolset(
		"registration_links",
		"Razorpay Registration Links related tools").
		AddWriteTools(
			CreateRegistrationLink(obs, client),
		)

	settlements := toolsets.NewToolset("settlements",
		"Razorpay Settlements related tools").
		AddReadTools(
			FetchSettlement(obs, client),
			FetchSettlementRecon(obs, client),
			FetchAllSettlements(obs, client),
			FetchAllInstantSettlements(obs, client),
			FetchInstantSettlement(obs, client),
		).
		AddWriteTools(
			CreateInstantSettlement(obs, client),
		)

	// Add the single custom tool to an existing toolset
	payments.AddWriteTools(FetchSavedPaymentMethods(obs, client)).
		AddWriteTools(RevokeToken(obs, client))

	// Checkout Integration toolset - helps integrate Razorpay checkout
	checkoutIntegration := toolsets.NewToolset(
		"checkout_integration",
		"Tools to help developers integrate Razorpay Standard Checkout").
		AddReadTools(
			integrations.IntegrateRazorpayCheckout(obs, client),
			integrations.DetectStack(obs, client),
		)

	// Add toolsets to the group
	toolsetGroup.AddToolset(checkoutIntegration)
	toolsetGroup.AddToolset(payments)
	toolsetGroup.AddToolset(paymentLinks)
	toolsetGroup.AddToolset(orders)
	toolsetGroup.AddToolset(refunds)
	toolsetGroup.AddToolset(payouts)
	toolsetGroup.AddToolset(qrCodes)
	toolsetGroup.AddToolset(settlements)
	toolsetGroup.AddToolset(registrationLinks)

	if err := applyRegionProfile(
		toolsetGroup, cfg.profile, enabledToolsets); err != nil {
		return nil, err
	}
	logHiddenRegionTools(obs, toolsetGroup, cfg.profile, enabledToolsets)

	// Enable the requested features
	if err := toolsetGroup.EnableToolsets(enabledToolsets); err != nil {
		return nil, err
	}

	return toolsetGroup, nil
}

// applyRegionProfile localizes every tool and keeps only the profile's tools.
func applyRegionProfile(
	tg *toolsets.ToolsetGroup,
	p region.Profile,
	enabledToolsets []string,
) error {
	known := map[string]bool{}
	for _, tool := range tg.AllTools() {
		if l, ok := tool.(interface{ Localize(func(string) string) }); ok {
			l.Localize(p.Localize)
		}
		if n, ok := tool.(interface{ GetName() string }); ok {
			known[n.GetName()] = true
		}
	}

	if p.AllowsAll() {
		return nil
	}
	for _, name := range p.Tools {
		if !known[name] {
			return fmt.Errorf("region %s: unknown tool %q", p.CountryCode, name)
		}
	}
	tg.SetToolFilter(p.AllowsTool)

	for _, name := range enabledToolsets {
		if _, ok := tg.Toolsets[name]; ok && tg.AllowedToolCount(name) == 0 {
			return fmt.Errorf("toolset %q has no tools available in region %s",
				name, p.CountryCode)
		}
	}
	return nil
}

// logHiddenRegionTools logs tools a requested toolset loses to the region.
func logHiddenRegionTools(
	obs *observability.Observability,
	tg *toolsets.ToolsetGroup,
	p region.Profile,
	enabledToolsets []string,
) {
	if obs == nil || obs.Logger == nil {
		return
	}
	for _, name := range enabledToolsets {
		if hidden := tg.HiddenTools(name); len(hidden) > 0 {
			obs.Logger.Infof(context.Background(),
				"tools not available in region", "region", p.CountryCode,
				"toolset", name, "tools", hidden)
		}
	}
}
