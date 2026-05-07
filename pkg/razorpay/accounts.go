package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// CreateAccount returns a tool that creates a new sub-merchant account
// under a partner using the Razorpay Partnerships Onboarding API (POST /v2/accounts).
func CreateAccount(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"email",
			mcpgo.Description("Business email address of the sub-merchant. "+
				"Must be a valid email format. If provided, must not already exist in Razorpay."),
		),
		mcpgo.WithString(
			"phone",
			mcpgo.Description("Business phone number of the sub-merchant. Required. "+
				"Must be a valid phone number (8–15 digits, with or without country code)."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"legal_business_name",
			mcpgo.Description("Legal business name of the sub-merchant. "+
				"Must be a safe alphanumeric string."),
		),
		mcpgo.WithString(
			"business_type",
			mcpgo.Description("Type of business. Accepted values: "+
				"proprietorship, individual, partnership, private_limited, public_limited, "+
				"llp, ngo, educational_institutes, trust, society, not_yet_registered, "+
				"other, huf, government, judicial_person, local_authority, section_8_company"),
		),
		mcpgo.WithString(
			"contact_name",
			mcpgo.Description("Full name of the contact person at the sub-merchant "+
				"(max 255 characters). Allowed: letters, spaces, comma, @, #, hyphen, period, %, /"),
			mcpgo.Max(255),
		),
		mcpgo.WithString(
			"customer_facing_business_name",
			mcpgo.Description("Name shown to customers on payment pages and billing labels. "+
				"Defaults to legal_business_name if not set."),
		),
		mcpgo.WithString(
			"reference_id",
			mcpgo.Description("Partner's own unique reference ID for this sub-merchant "+
				"(max 512 characters)."),
			mcpgo.Max(512),
		),
		mcpgo.WithString(
			"type",
			mcpgo.Description("Account type. Only accepted value: route. "+
				"Use this for route/marketplace partner flows."),
		),
		mcpgo.WithObject(
			"profile",
			mcpgo.Description("Business profile details. Fields: "+
				"category (required, one of: ecommerce, education, financial_services, food, gaming, "+
				"government, healthcare, housing, it_and_software, logistics, media_and_entertainment, "+
				"not_for_profit, services, social, tours_and_travel, transport, utilities, others), "+
				"subcategory (required, e.g. 'ecommerce_marketplace', 'saas', 'grocery' — see docs for full list), "+
				"description (string, max 255 chars, letters/spaces/comma/@/#/hyphen/period/%%/slash only), "+
				"addresses (object with 'registered' (required) and 'operation' (optional) — "+
				"each address has: street1 (required, max 100 chars), street2 (required, max 100 chars), "+
				"city (required), state (required, valid Indian state code e.g. 'MH', 'KA'), "+
				"postal_code (required), country (required))"),
		),
		mcpgo.WithObject(
			"legal_info",
			mcpgo.Description("Legal identifiers of the business. Fields: "+
				"pan (string, business PAN card number), "+
				"gst (string, GST number), "+
				"cin (string, Corporate Identity Number)"),
		),
		mcpgo.WithObject(
			"brand",
			mcpgo.Description("Branding details. Fields: "+
				"color (string, 6-character hex color code WITHOUT # prefix, e.g. 'FF5733' or '3b5998')"),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for partner's internal reference "+
				"(max 15 pairs, 512 characters per value)."),
			mcpgo.MaxProperties(15),
		),
		mcpgo.WithObject(
			"contact_info",
			mcpgo.Description("Contact details for customer-facing communication. "+
				"Sub-fields: chargeback, refund, support, dispute — "+
				"each is an object with: email (required, valid email), "+
				"phone (optional, valid phone number), policy_url (optional, string)"),
		),
		mcpgo.WithObject(
			"apps",
			mcpgo.Description("Apps/platforms where the sub-merchant accepts payments. "+
				"Fields: "+
				"websites (array of URL strings), "+
				"android (array of objects with 'url' (required) and 'name' (required), max 100 entries), "+
				"ios (array of objects with 'url' (required) and 'name' (required), max 100 entries)"),
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
			ValidateAndAddRequiredString(payload, "phone").
			ValidateAndAddOptionalString(payload, "email").
			ValidateAndAddOptionalString(payload, "legal_business_name").
			ValidateAndAddOptionalString(payload, "business_type").
			ValidateAndAddOptionalString(payload, "contact_name").
			ValidateAndAddOptionalString(payload, "customer_facing_business_name").
			ValidateAndAddOptionalString(payload, "reference_id").
			ValidateAndAddOptionalString(payload, "type").
			ValidateAndAddOptionalMap(payload, "profile").
			ValidateAndAddOptionalMap(payload, "legal_info").
			ValidateAndAddOptionalMap(payload, "brand").
			ValidateAndAddOptionalMap(payload, "notes").
			ValidateAndAddOptionalMap(payload, "contact_info").
			ValidateAndAddOptionalMap(payload, "apps")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		account, err := client.Account.Create(payload, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating account failed: %s", err.Error()),
			), nil
		}

		return mcpgo.NewToolResultJSON(account)
	}

	return mcpgo.NewTool(
		"create_account",
		"Create a new sub-merchant account under the partner using the "+
			"Razorpay Partnerships Onboarding API (POST /v2/accounts). "+
			"\n\nOnly 'phone' is required. All other fields are optional but recommended for faster KYC activation. "+
			"\n\nKey fields: email (must not already exist in Razorpay), phone (required), "+
			"legal_business_name, business_type (proprietorship/individual/partnership/private_limited/"+
			"public_limited/llp/ngo/educational_institutes/trust/society/not_yet_registered/other/huf/"+
			"government/judicial_person/local_authority/section_8_company), contact_name. "+
			"\n\nProfile requires category + subcategory + registered address (street1, street2, city, state, postal_code, country). "+
			"\n\nBrand color must be a 6-char hex WITHOUT # prefix (e.g. 'FF5733'). "+
			"\n\nThe created account starts in 'created' status. Use the product configuration API "+
			"to request product activation for this account.",
		parameters,
		handler,
	)
}

// FetchAccount returns a tool to fetch a sub-merchant account's details by ID
// using the Razorpay Partnerships Onboarding API (GET /v2/accounts/:account_id).
func FetchAccount(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"account_id",
			mcpgo.Description("Unique identifier of the sub-merchant account to fetch. "+
				"Must start with 'acc_' (e.g. acc_GP4lfNA0iIMn5B)."),
			mcpgo.Required(),
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
			ValidateAndAddRequiredString(payload, "account_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		account, err := client.Account.Fetch(payload["account_id"].(string), nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("fetching account failed", err),
			), nil
		}

		return mcpgo.NewToolResultJSON(account)
	}

	return mcpgo.NewTool(
		"fetch_account",
		"Fetch a sub-merchant account's details by its ID using the "+
			"Razorpay Partnerships Onboarding API (GET /v2/accounts/:account_id). "+
			"\n\nReturns full account details including status (created/under_review/"+
			"needs_clarification/activated/rejected), activation timestamps, "+
			"profile, legal info, and contact details.",
		parameters,
		handler,
	)
}
