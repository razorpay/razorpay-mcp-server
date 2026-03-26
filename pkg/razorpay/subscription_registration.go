package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// CreateSubscriptionRegistrationAuthLink returns a tool that creates a
// subscription registration auth link (emandate / recurring payment
// registration link) via the Razorpay API.
func CreateSubscriptionRegistrationAuthLink(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		// Customer fields
		mcpgo.WithString(
			"customer_name",
			mcpgo.Description("Full name of the customer."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"customer_email",
			mcpgo.Description("Email address of the customer."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"customer_contact",
			mcpgo.Description("Contact/phone number of the customer (e.g. +919876543210)."),
			mcpgo.Required(),
		),
		// Top-level fields
		mcpgo.WithString(
			"type",
			mcpgo.Description("Type of the auth link. Must be 'link'."),
			mcpgo.Required(),
			mcpgo.Enum("link"),
		),
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Amount to be charged (in smallest currency unit, e.g. paise for INR). Example: 100 means ₹1."), // nolint:lll
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("Three-letter ISO currency code (e.g. INR, MYR)."),
			mcpgo.Required(),
		),
		// subscription_registration sub-object: method is required
		mcpgo.WithString(
			"method",
			mcpgo.Description("Payment method for the registration. Supported: 'card', 'emandate', 'nach', 'upi'."), // nolint:lll
			mcpgo.Required(),
			mcpgo.Enum("card", "emandate", "nach", "upi"),
		),
		// Optional top-level fields
		mcpgo.WithString(
			"description",
			mcpgo.Description("A short description about the registration link."),
		),
		mcpgo.WithString(
			"receipt",
			mcpgo.Description("A unique identifier for the receipt. Max 40 characters."),
		),
		mcpgo.WithBoolean(
			"email_notify",
			mcpgo.Description("Send an email notification to the customer. Default: true."),
		),
		mcpgo.WithBoolean(
			"sms_notify",
			mcpgo.Description("Send an SMS notification to the customer. Default: true."),
		),
		mcpgo.WithNumber(
			"expire_by",
			mcpgo.Description("Unix timestamp after which the auth link expires."),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for storing additional information. Max 15 pairs, each value up to 256 characters."), // nolint:lll
		),
		// Optional subscription_registration sub-object fields
		mcpgo.WithNumber(
			"max_amount",
			mcpgo.Description("Maximum amount (in smallest currency unit) that can be charged in a single debit."), // nolint:lll
		),
		mcpgo.WithNumber(
			"expire_at",
			mcpgo.Description("Unix timestamp after which the subscription registration token expires."), // nolint:lll
		),
		mcpgo.WithString(
			"frequency",
			mcpgo.Description("Frequency of the recurring charge. Supported: 'as_presented', 'monthly', 'weekly', 'daily'."), // nolint:lll
			mcpgo.Enum("as_presented", "monthly", "weekly", "daily"),
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

		reqBody := make(map[string]interface{})
		customerFields := make(map[string]interface{})
		subscriptionReg := make(map[string]interface{})

		// Validate required top-level and subscription_registration fields
		validator := NewValidator(&r).
			ValidateAndAddRequiredString(reqBody, "type").
			ValidateAndAddRequiredInt(reqBody, "amount").
			ValidateAndAddRequiredString(reqBody, "currency").
			ValidateAndAddRequiredString(subscriptionReg, "method").
			// Required customer fields (extracted with full param names, remapped below)
			ValidateAndAddRequiredString(customerFields, "customer_name").
			ValidateAndAddRequiredString(customerFields, "customer_email").
			ValidateAndAddRequiredString(customerFields, "customer_contact").
			// Optional top-level fields
			ValidateAndAddOptionalString(reqBody, "description").
			ValidateAndAddOptionalString(reqBody, "receipt").
			ValidateAndAddOptionalBool(reqBody, "email_notify").
			ValidateAndAddOptionalBool(reqBody, "sms_notify").
			ValidateAndAddOptionalInt(reqBody, "expire_by").
			ValidateAndAddOptionalMap(reqBody, "notes").
			// Optional subscription_registration fields
			ValidateAndAddOptionalIntToPath(subscriptionReg, "max_amount", "max_amount").
			ValidateAndAddOptionalIntToPath(subscriptionReg, "expire_at", "expire_at").
			ValidateAndAddOptionalStringToPath(subscriptionReg, "frequency", "frequency")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Remap customer fields (strip the "customer_" prefix for the API payload)
		if name, ok := customerFields["customer_name"]; ok {
			customer := map[string]interface{}{
				"name":    name,
				"email":   customerFields["customer_email"],
				"contact": customerFields["customer_contact"],
			}
			reqBody["customer"] = customer
		}

		reqBody["subscription_registration"] = subscriptionReg

		url := fmt.Sprintf("/%s/subscription_registration/auth_links", constants.VERSION_V1)
		response, err := client.Request.Post(url, reqBody, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating subscription registration auth link failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"create_subscription_registration_auth_link",
		"Create a subscription registration auth link (recurring payment / emandate "+
			"registration link) in Razorpay. The link is sent to the customer via SMS/email "+
			"and allows them to register their payment method for future recurring charges.",
		parameters,
		handler,
	)
}
