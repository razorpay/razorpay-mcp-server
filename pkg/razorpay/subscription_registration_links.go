package razorpay

import (
	"context"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// CreateSubscriptionRegistrationLink returns a tool that creates a subscription
// registration auth link used to collect payment method authorization from
// customers for recurring payments.
func CreateSubscriptionRegistrationLink(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithObject(
			"customer",
			mcpgo.Description("Customer details. Must contain: name (string), "+
				"email (string), contact (string with country code, e.g. +919876543210)."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"type",
			mcpgo.Description("Type of auth mechanism. Use 'link' to send "+
				"a registration link to the customer."),
			mcpgo.Required(),
			mcpgo.Enum("link"),
		),
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Authorization amount in smallest currency subunit "+
				"(e.g. paise for INR). For INR: 100 paise = ₹1."),
			mcpgo.Required(),
			mcpgo.Min(0),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("Three-letter ISO currency code (e.g. INR, MYR)."),
			mcpgo.Required(),
			mcpgo.Pattern("^[A-Z]{3}$"),
		),
		mcpgo.WithObject(
			"subscription_registration",
			mcpgo.Description("Subscription registration details. Must contain: "+
				"method (string, e.g. 'card', 'emandate', 'nach'), "+
				"max_amount (number, maximum debit amount in smallest currency unit), "+
				"expire_at (Unix timestamp when the mandate expires), "+
				"frequency (string: 'as_presented', 'monthly', 'yearly', etc.)."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("Description of the subscription registration link."),
		),
		mcpgo.WithString(
			"receipt",
			mcpgo.Description("Receipt number for internal reference (max 40 chars)."),
			mcpgo.Max(40),
		),
		mcpgo.WithBoolean(
			"email_notify",
			mcpgo.Description("Send email notification to the customer."),
		),
		mcpgo.WithBoolean(
			"sms_notify",
			mcpgo.Description("Send SMS notification to the customer."),
		),
		mcpgo.WithNumber(
			"expire_by",
			mcpgo.Description("Unix timestamp after which the link expires."),
			mcpgo.Min(0),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for additional information "+
				"(max 15 pairs, 256 chars each)."),
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

		payload := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredMap(payload, "customer").
			ValidateAndAddRequiredString(payload, "type").
			ValidateAndAddRequiredFloat(payload, "amount").
			ValidateAndAddRequiredString(payload, "currency").
			ValidateAndAddRequiredMap(payload, "subscription_registration").
			ValidateAndAddOptionalString(payload, "description").
			ValidateAndAddOptionalString(payload, "receipt").
			ValidateAndAddOptionalBool(payload, "email_notify").
			ValidateAndAddOptionalBool(payload, "sms_notify").
			ValidateAndAddOptionalInt(payload, "expire_by").
			ValidateAndAddOptionalMap(payload, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		resp, err := client.Post( //nolint:lll
			"/v1/subscription_registration/auth_links", payload, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("creating subscription registration link failed", err),
			), nil
		}

		return mcpgo.NewToolResultJSON(resp)
	}

	return mcpgo.NewTool(
		"create_subscription_registration_link",
		"Create a subscription registration auth link to collect payment method "+
			"authorization from customers for recurring payments. "+
			"The link is sent to the customer via SMS or email and allows them "+
			"to authorize a mandate (e.g. card, emandate, nach) for future debits.",
		parameters,
		handler,
	)
}
