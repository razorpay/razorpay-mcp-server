package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// CreateRegistrationLink creates a registration link
// (auth link) for subscription registration
func CreateRegistrationLink(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"type",
			mcpgo.Description(
				"Type of registration link. "+
					"Use 'link'."),
			mcpgo.Required(),
		),
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description(
				"Amount in the smallest currency "+
					"unit (e.g., paise for INR)."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description(
				"Three-letter ISO currency code "+
					"(e.g., INR, MYR)."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description(
				"Brief description of the "+
					"registration link."),
			mcpgo.Required(),
		),
		mcpgo.WithObject(
			"subscription_registration",
			mcpgo.Description(
				"Subscription registration details."+
					" Must include 'method' (card,"+
					" emandate, nach, upi). May"+
					" include 'max_amount',"+
					" 'expire_at' (Unix timestamp),"+
					" and 'frequency' (as_presented,"+
					" monthly, weekly, yearly,"+
					" daily)."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"customer_name",
			mcpgo.Description(
				"Name of the customer."),
		),
		mcpgo.WithString(
			"customer_email",
			mcpgo.Description(
				"Email address of the customer."),
		),
		mcpgo.WithString(
			"customer_contact",
			mcpgo.Description(
				"Contact number of the customer."),
		),
		mcpgo.WithString(
			"receipt",
			mcpgo.Description(
				"Unique receipt identifier "+
					"provided by the merchant."),
		),
		mcpgo.WithBoolean(
			"email_notify",
			mcpgo.Description(
				"Send email notification. "+
					"Default: true"),
		),
		mcpgo.WithBoolean(
			"sms_notify",
			mcpgo.Description(
				"Send SMS notification. "+
					"Default: true"),
		),
		mcpgo.WithNumber(
			"expire_by",
			mcpgo.Description(
				"Unix timestamp when the "+
					"registration link expires."),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description(
				"Key-value pairs for additional "+
					"info. Max 15 pairs, each up "+
					"to 256 characters."),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		client, err := getClientFromContextOrDefault(
			ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(
				err.Error()), nil
		}

		payload := make(map[string]interface{})
		customer := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(
				payload, "type").
			ValidateAndAddRequiredInt(
				payload, "amount").
			ValidateAndAddRequiredString(
				payload, "currency").
			ValidateAndAddRequiredString(
				payload, "description").
			ValidateAndAddRequiredMap(
				payload,
				"subscription_registration").
			ValidateAndAddOptionalStringToPath(
				customer,
				"customer_name", "name").
			ValidateAndAddOptionalStringToPath(
				customer,
				"customer_email", "email").
			ValidateAndAddOptionalStringToPath(
				customer,
				"customer_contact", "contact").
			ValidateAndAddOptionalString(
				payload, "receipt").
			ValidateAndAddOptionalBool(
				payload, "email_notify").
			ValidateAndAddOptionalBool(
				payload, "sms_notify").
			ValidateAndAddOptionalInt(
				payload, "expire_by").
			ValidateAndAddOptionalMap(
				payload, "notes")

		if result, err :=
			validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		if len(customer) > 0 {
			payload["customer"] = customer
		}

		url := fmt.Sprintf(
			"/%s/subscription_registration/auth_links",
			constants.VERSION_V1)

		response, err := client.Invoice.Request.Post(
			url, payload, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"creating registration link "+
						"failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"create_registration_link",
		"Create a registration link (auth link) "+
			"for subscription registration in "+
			"Razorpay to set up recurring payments "+
			"via card, emandate, NACH, or UPI.",
		parameters,
		handler,
	)
}
