package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// CreateSubscriptionRegistrationAuthLink returns a tool that creates
// subscription registration authorization links
func CreateSubscriptionRegistrationAuthLink(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_name",
			mcpgo.Description("Name of the customer"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"customer_email",
			mcpgo.Description("Email address of the customer"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"customer_contact",
			mcpgo.Description("Contact number of the customer "+
				"(e.g., +60102102460)"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"type",
			mcpgo.Description("Type of the authorization link. Must be 'link'"),
			mcpgo.Required(),
			mcpgo.Enum("link"),
		),
		mcpgo.WithString(
			"amount",
			mcpgo.Description("Amount for the authorization transaction "+
				"in the smallest currency unit (e.g., paise for INR, "+
				"sen for MYR). Example: '100' for 1.00 MYR"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("Three-letter ISO currency code "+
				"(e.g., INR, MYR, USD)"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("Description of the registration link purpose"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"subscription_registration_method",
			mcpgo.Description("Payment method for subscription registration. "+
				"Supported values: 'card', 'emandate', 'nach'"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"subscription_registration_max_amount",
			mcpgo.Description("Maximum amount that can be charged in a single "+
				"transaction in smallest currency unit"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"subscription_registration_expire_at",
			mcpgo.Description("Unix timestamp when the subscription "+
				"registration expires"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"subscription_registration_frequency",
			mcpgo.Description("Frequency of the subscription charges. "+
				"Supported values: 'as_presented', 'daily', 'weekly', "+
				"'monthly', 'yearly'"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"receipt",
			mcpgo.Description("Receipt number for the transaction. "+
				"Must be unique"),
		),
		mcpgo.WithBoolean(
			"email_notify",
			mcpgo.Description("Whether to send email notification "+
				"to the customer. Default: true"),
		),
		mcpgo.WithBoolean(
			"sms_notify",
			mcpgo.Description("Whether to send SMS notification "+
				"to the customer. Default: true"),
		),
		mcpgo.WithNumber(
			"expire_by",
			mcpgo.Description("Unix timestamp when the auth link expires"),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for additional information. "+
				"Maximum 15 pairs, each value limited to 256 characters"),
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
		customer := make(map[string]interface{})
		subscriptionReg := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddOptionalStringToPath(
				customer, "customer_name", "name").
			ValidateAndAddOptionalStringToPath(
				customer, "customer_email", "email").
			ValidateAndAddOptionalStringToPath(
				customer, "customer_contact", "contact").
			ValidateAndAddRequiredString(payload, "type").
			ValidateAndAddRequiredString(payload, "amount").
			ValidateAndAddRequiredString(payload, "currency").
			ValidateAndAddRequiredString(payload, "description").
			ValidateAndAddOptionalStringToPath(
				subscriptionReg, "subscription_registration_method", "method").
			ValidateAndAddOptionalStringToPath(
				subscriptionReg, "subscription_registration_max_amount",
				"max_amount").
			ValidateAndAddOptionalStringToPath(
				subscriptionReg, "subscription_registration_expire_at",
				"expire_at").
			ValidateAndAddOptionalStringToPath(
				subscriptionReg, "subscription_registration_frequency",
				"frequency").
			ValidateAndAddOptionalString(payload, "receipt").
			ValidateAndAddOptionalBool(payload, "email_notify").
			ValidateAndAddOptionalBool(payload, "sms_notify").
			ValidateAndAddOptionalInt(payload, "expire_by").
			ValidateAndAddOptionalMap(payload, "notes")

		// Validate required nested fields manually
		if _, exists := customer["name"]; !exists {
			validator = validator.addError(
				fmt.Errorf("missing required parameter: customer_name"))
		}
		if _, exists := customer["email"]; !exists {
			validator = validator.addError(
				fmt.Errorf("missing required parameter: customer_email"))
		}
		if _, exists := customer["contact"]; !exists {
			validator = validator.addError(
				fmt.Errorf("missing required parameter: customer_contact"))
		}
		if _, exists := subscriptionReg["method"]; !exists {
			validator = validator.addError(fmt.Errorf(
				"missing required parameter: subscription_registration_method"))
		}
		if _, exists := subscriptionReg["max_amount"]; !exists {
			validator = validator.addError(fmt.Errorf(
				"missing required parameter: " +
					"subscription_registration_max_amount"))
		}
		if _, exists := subscriptionReg["expire_at"]; !exists {
			validator = validator.addError(fmt.Errorf(
				"missing required parameter: " +
					"subscription_registration_expire_at"))
		}
		if _, exists := subscriptionReg["frequency"]; !exists {
			validator = validator.addError(fmt.Errorf(
				"missing required parameter: " +
					"subscription_registration_frequency"))
		}

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Add customer and subscription_registration objects to payload
		payload["customer"] = customer
		payload["subscription_registration"] = subscriptionReg

		// Make API call
		url := fmt.Sprintf("/%s/subscription_registration/auth_links",
			constants.VERSION_V1)
		response, err := client.Request.Post(url, payload, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"creating subscription registration auth link failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"create_subscription_registration_auth_link",
		"Create a subscription registration authorization link for recurring "+
			"payments. Use when setting up recurring payment mandates via "+
			"card, emandate, or NACH. The link allows customers to authorize "+
			"future charges up to a specified maximum amount. "+
			"Amount is in smallest currency unit (paise for INR, sen for MYR).",
		parameters,
		handler,
	)
}
