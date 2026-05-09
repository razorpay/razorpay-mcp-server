package onboardingapis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// accountParameters returns the full flat parameter list shared by both
// preview_create_account and create_account.
func accountParameters() []mcpgo.ToolParameter {
	return []mcpgo.ToolParameter{
		// --- Auth ---
		mcpgo.WithString("bearer_token",
			mcpgo.Description("(Mandatory) OAuth access token obtained from generate_access_token. "+
				"Used as 'Authorization: Bearer <token>'."),
			mcpgo.Required(),
		),

		// --- Mandatory fields ---
		mcpgo.WithString("email",
			mcpgo.Description("(Mandatory) Business email address of the sub-merchant."),
			mcpgo.Required(),
		),
		mcpgo.WithString("phone",
			mcpgo.Description("(Mandatory) Business phone number, 8-15 digits, without country code."),
			mcpgo.Required(),
		),
		mcpgo.WithString("legal_business_name",
			mcpgo.Description("(Mandatory) Registered legal business name (4-200 characters)."),
			mcpgo.Required(),
		),
		mcpgo.WithString("business_type",
			mcpgo.Description("(Mandatory) Type of business. One of: proprietorship, individual, partnership, "+
				"private_limited, public_limited, llp, ngo, educational_institutes, trust, society, "+
				"not_yet_registered, other, huf, government, judicial_person, local_authority, section_8_company."),
			mcpgo.Required(),
		),
		mcpgo.WithString("contact_name",
			mcpgo.Description("(Mandatory) Full name of the contact person (4-255 characters)."),
			mcpgo.Required(),
		),

		// --- Top-level optional fields ---
		mcpgo.WithString("customer_facing_business_name",
			mcpgo.Description("(Optional) Name shown to customers on payment pages. Defaults to legal_business_name."),
		),
		mcpgo.WithString("reference_id",
			mcpgo.Description("(Optional) Partner's own unique reference ID for this sub-merchant (max 512 characters)."),
			mcpgo.Max(512),
		),

		// --- Profile fields ---
		mcpgo.WithString("profile_category",
			mcpgo.Description("(Optional, required if any profile field is set) Business category. One of: "+
				"ecommerce, education, financial_services, food, gaming, government, healthcare, housing, "+
				"it_and_software, logistics, media_and_entertainment, not_for_profit, services, social, "+
				"tours_and_travel, transport, utilities, others."),
		),
		mcpgo.WithString("profile_subcategory",
			mcpgo.Description("(Optional, required if any profile field is set) Business subcategory e.g. "+
				"'clinic', 'hospital', 'ecommerce_marketplace', 'saas', 'grocery'. See Razorpay docs for full list."),
		),
		mcpgo.WithString("profile_description",
			mcpgo.Description("(Optional) Short description of the business (max 255 characters)."),
			mcpgo.Max(255),
		),
		mcpgo.WithString("profile_business_model",
			mcpgo.Description("(Optional) Description of the business model e.g. 'b2c', 'b2b'."),
		),

		// --- Registered address (required when profile is set) ---
		mcpgo.WithString("registered_street1",
			mcpgo.Description("(Optional, required when profile is set) Registered address line 1 (max 100 chars)."),
			mcpgo.Max(100),
		),
		mcpgo.WithString("registered_street2",
			mcpgo.Description("(Optional, required when profile is set) Registered address line 2 (max 100 chars)."),
			mcpgo.Max(100),
		),
		mcpgo.WithString("registered_city",
			mcpgo.Description("(Optional, required when profile is set) City of registered address."),
		),
		mcpgo.WithString("registered_state",
			mcpgo.Description("(Optional, required when profile is set) State of registered address e.g. 'Karnataka', 'MH'."),
		),
		mcpgo.WithNumber("registered_postal_code",
			mcpgo.Description("(Optional, required when profile is set) Postal code as integer e.g. 560034."),
		),
		mcpgo.WithString("registered_country",
			mcpgo.Description("(Optional, required when profile is set) Country ISO code e.g. 'IN'."),
		),

		// --- Operation address (optional even when profile is set) ---
		mcpgo.WithString("operation_street1",
			mcpgo.Description("(Optional) Operation address line 1 (max 100 chars)."),
			mcpgo.Max(100),
		),
		mcpgo.WithString("operation_street2",
			mcpgo.Description("(Optional) Operation address line 2 (max 100 chars)."),
			mcpgo.Max(100),
		),
		mcpgo.WithString("operation_city",
			mcpgo.Description("(Optional) City of operation address."),
		),
		mcpgo.WithString("operation_state",
			mcpgo.Description("(Optional) State of operation address e.g. 'Karnataka', 'MH'."),
		),
		mcpgo.WithNumber("operation_postal_code",
			mcpgo.Description("(Optional) Operation address postal code as integer e.g. 560034."),
		),
		mcpgo.WithString("operation_country",
			mcpgo.Description("(Optional) Operation address country ISO code e.g. 'IN'."),
		),

		// --- Legal info ---
		mcpgo.WithString("pan",
			mcpgo.Description("(Optional) Business PAN card number."),
		),
		mcpgo.WithString("gst",
			mcpgo.Description("(Optional) GST number."),
		),
		mcpgo.WithString("cin",
			mcpgo.Description("(Optional) Corporate Identity Number."),
		),

		// --- Brand ---
		mcpgo.WithString("brand_color",
			mcpgo.Description("(Optional) Brand hex color WITHOUT # prefix e.g. 'FFFFFF', 'FF5733'."),
			mcpgo.Pattern("^[0-9a-fA-F]{6}$"),
		),

		// --- Notes ---
		mcpgo.WithObject("notes",
			mcpgo.Description("(Optional) Key-value pairs for internal reference (max 15 pairs)."),
			mcpgo.MaxProperties(15),
		),

		// --- Contact info ---
		mcpgo.WithString("chargeback_email",
			mcpgo.Description("(Optional) Email for chargeback communication."),
		),
		mcpgo.WithString("chargeback_phone",
			mcpgo.Description("(Optional) Phone for chargeback communication."),
		),
		mcpgo.WithString("chargeback_policy_url",
			mcpgo.Description("(Optional) Policy URL for chargeback."),
		),
		mcpgo.WithString("refund_email",
			mcpgo.Description("(Optional) Email for refund communication."),
		),
		mcpgo.WithString("refund_phone",
			mcpgo.Description("(Optional) Phone for refund communication."),
		),
		mcpgo.WithString("refund_policy_url",
			mcpgo.Description("(Optional) Policy URL for refunds."),
		),
		mcpgo.WithString("support_email",
			mcpgo.Description("(Optional) Email for customer support."),
		),
		mcpgo.WithString("support_phone",
			mcpgo.Description("(Optional) Phone for customer support."),
		),
		mcpgo.WithString("support_policy_url",
			mcpgo.Description("(Optional) Policy URL for support."),
		),
		mcpgo.WithString("dispute_email",
			mcpgo.Description("(Optional) Email for dispute communication."),
		),
		mcpgo.WithString("dispute_phone",
			mcpgo.Description("(Optional) Phone for dispute communication."),
		),
		mcpgo.WithString("dispute_policy_url",
			mcpgo.Description("(Optional) Policy URL for disputes."),
		),

		// --- Apps ---
		mcpgo.WithArray("websites",
			mcpgo.Description("(Optional) List of website URLs where payments are accepted."),
			mcpgo.Items(map[string]interface{}{"type": "string"}),
		),
	}
}

// buildAccountPayload constructs the nested API request body from flat params
// and returns validation warnings for missing mandatory/conditional fields.
func buildAccountPayload(r *mcpgo.CallToolRequest) (map[string]interface{}, []string, error) {
	v := newValidator(r)
	p := make(map[string]interface{})
	var warnings []string

	// Mandatory fields
	v.requireString(p, "email").
		requireString(p, "phone").
		requireString(p, "legal_business_name").
		requireString(p, "business_type").
		requireString(p, "contact_name")

	if result, _ := v.handleErrorsIfAny(); result != nil {
		// Collect mandatory missing fields as warnings for preview,
		// errors for create.
		for _, e := range v.errors {
			warnings = append(warnings, "MISSING MANDATORY: "+e.Error())
		}
	}

	// Optional top-level strings
	opt := newValidator(r)
	optP := make(map[string]interface{})
	opt.optionalString(optP, "customer_facing_business_name").
		optionalString(optP, "reference_id")
	for k, val := range optP {
		p[k] = val
	}

	// Profile block
	profileCategory, _ := extractString(r, "profile_category", false)
	profileSubcategory, _ := extractString(r, "profile_subcategory", false)
	profileDescription, _ := extractString(r, "profile_description", false)
	profileBusinessModel, _ := extractString(r, "profile_business_model", false)

	regStreet1, _ := extractString(r, "registered_street1", false)
	regStreet2, _ := extractString(r, "registered_street2", false)
	regCity, _ := extractString(r, "registered_city", false)
	regState, _ := extractString(r, "registered_state", false)
	regPostal, _ := extractNumber(r, "registered_postal_code", false)
	regCountry, _ := extractString(r, "registered_country", false)

	opStreet1, _ := extractString(r, "operation_street1", false)
	opStreet2, _ := extractString(r, "operation_street2", false)
	opCity, _ := extractString(r, "operation_city", false)
	opState, _ := extractString(r, "operation_state", false)
	opPostal, _ := extractNumber(r, "operation_postal_code", false)
	opCountry, _ := extractString(r, "operation_country", false)

	hasProfile := profileCategory != nil || profileSubcategory != nil ||
		profileDescription != nil || profileBusinessModel != nil ||
		regStreet1 != nil

	if hasProfile {
		profile := map[string]interface{}{}

		if profileCategory == nil {
			warnings = append(warnings, "MISSING: profile_category is required when profile is provided")
		} else {
			profile["category"] = *profileCategory
		}
		if profileSubcategory == nil {
			warnings = append(warnings, "MISSING: profile_subcategory is required when profile is provided")
		} else {
			profile["subcategory"] = *profileSubcategory
		}
		if profileDescription != nil {
			profile["description"] = *profileDescription
		}
		if profileBusinessModel != nil {
			profile["business_model"] = *profileBusinessModel
		}

		// Registered address (required when profile is set)
		registered := map[string]interface{}{}
		missingReg := []string{}
		if regStreet1 != nil {
			registered["street1"] = *regStreet1
		} else {
			missingReg = append(missingReg, "registered_street1")
		}
		if regStreet2 != nil {
			registered["street2"] = *regStreet2
		} else {
			missingReg = append(missingReg, "registered_street2")
		}
		if regCity != nil {
			registered["city"] = *regCity
		} else {
			missingReg = append(missingReg, "registered_city")
		}
		if regState != nil {
			registered["state"] = *regState
		} else {
			missingReg = append(missingReg, "registered_state")
		}
		if regPostal != nil {
			registered["postal_code"] = int(*regPostal)
		} else {
			missingReg = append(missingReg, "registered_postal_code")
		}
		if regCountry != nil {
			registered["country"] = *regCountry
		} else {
			missingReg = append(missingReg, "registered_country")
		}
		if len(missingReg) > 0 {
			warnings = append(warnings, "MISSING: registered address fields required when profile is set: "+strings.Join(missingReg, ", "))
		}

		addresses := map[string]interface{}{"registered": registered}

		// Operation address (fully optional)
		hasOp := opStreet1 != nil || opStreet2 != nil || opCity != nil
		if hasOp {
			op := map[string]interface{}{}
			if opStreet1 != nil {
				op["street1"] = *opStreet1
			}
			if opStreet2 != nil {
				op["street2"] = *opStreet2
			}
			if opCity != nil {
				op["city"] = *opCity
			}
			if opState != nil {
				op["state"] = *opState
			}
			if opPostal != nil {
				op["postal_code"] = int(*opPostal)
			}
			if opCountry != nil {
				op["country"] = *opCountry
			}
			addresses["operation"] = op
		}

		profile["addresses"] = addresses
		p["profile"] = profile
	}

	// Legal info
	pan, _ := extractString(r, "pan", false)
	gst, _ := extractString(r, "gst", false)
	cin, _ := extractString(r, "cin", false)
	if pan != nil || gst != nil || cin != nil {
		legal := map[string]interface{}{}
		if pan != nil {
			legal["pan"] = *pan
		}
		if gst != nil {
			legal["gst"] = *gst
		}
		if cin != nil {
			legal["cin"] = *cin
		}
		p["legal_info"] = legal
	}

	// Brand
	brandColor, _ := extractString(r, "brand_color", false)
	if brandColor != nil {
		p["brand"] = map[string]interface{}{"color": *brandColor}
	}

	// Notes
	notesV := newValidator(r)
	notesP := make(map[string]interface{})
	notesV.optionalMap(notesP, "notes")
	if n, ok := notesP["notes"]; ok {
		p["notes"] = n
	}

	// Contact info
	contactSections := map[string]map[string]*string{
		"chargeback": {
			"email":      mustExtractString(r, "chargeback_email"),
			"phone":      mustExtractString(r, "chargeback_phone"),
			"policy_url": mustExtractString(r, "chargeback_policy_url"),
		},
		"refund": {
			"email":      mustExtractString(r, "refund_email"),
			"phone":      mustExtractString(r, "refund_phone"),
			"policy_url": mustExtractString(r, "refund_policy_url"),
		},
		"support": {
			"email":      mustExtractString(r, "support_email"),
			"phone":      mustExtractString(r, "support_phone"),
			"policy_url": mustExtractString(r, "support_policy_url"),
		},
		"dispute": {
			"email":      mustExtractString(r, "dispute_email"),
			"phone":      mustExtractString(r, "dispute_phone"),
			"policy_url": mustExtractString(r, "dispute_policy_url"),
		},
	}
	contactInfo := map[string]interface{}{}
	for section, fields := range contactSections {
		obj := map[string]interface{}{}
		for field, val := range fields {
			if val != nil {
				obj[field] = *val
			}
		}
		if len(obj) > 0 {
			contactInfo[section] = obj
		}
	}
	if len(contactInfo) > 0 {
		p["contact_info"] = contactInfo
	}

	// Apps - websites
	websites, _ := extractStringArray(r, "websites")
	if len(websites) > 0 {
		p["apps"] = map[string]interface{}{"websites": websites}
	}

	return p, warnings, nil
}

// mustExtractString extracts an optional string, returning nil on error.
func mustExtractString(r *mcpgo.CallToolRequest, name string) *string {
	val, _ := extractString(r, name, false)
	return val
}

// PreviewCreateAccount returns a tool that constructs and previews the
// create account request body without making the API call.
func PreviewCreateAccount(
	obs *observability.Observability,
	_ *rzpsdk.Client,
) mcpgo.Tool {
	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		payload, warnings, _ := buildAccountPayload(&r)

		prettyJSON, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("constructing request preview failed", err),
			), nil
		}

		var sb strings.Builder
		sb.WriteString("## Constructed Request Body\n\n")
		sb.WriteString("```json\n")
		sb.WriteString(string(prettyJSON))
		sb.WriteString("\n```\n")

		if len(warnings) > 0 {
			sb.WriteString("\n## ⚠️ Validation Issues\n\n")
			for _, w := range warnings {
				sb.WriteString(fmt.Sprintf("- %s\n", w))
			}
			sb.WriteString("\nPlease provide the missing fields before proceeding.\n")
		} else {
			sb.WriteString("\n✅ All mandatory fields are present.\n")
		}

		sb.WriteString("\n---\n")
		sb.WriteString("Would you like to:\n")
		sb.WriteString("- **Add more fields?** Tell me what else you'd like to include.\n")
		sb.WriteString("- **Proceed?** Call `create_account` with the same fields to submit.\n")

		return mcpgo.NewToolResultText(sb.String()), nil
	}

	return mcpgo.NewTool(
		"preview_create_account",
		"Preview the sub-merchant account creation request body before submitting. "+
			"Accepts all fields individually (flat), assembles the nested JSON, validates mandatory fields, "+
			"and returns the constructed payload for review. "+
			"\n\nMandatory fields: email, phone, legal_business_name, business_type, contact_name. "+
			"\n\nIf profile fields are provided, registered address fields are also required. "+
			"\n\nAfter reviewing the preview, call create_account with the same fields to submit.",
		accountParameters(),
		handler,
	)
}

// CreateAccount returns a tool that creates a new sub-merchant account
// using the Razorpay Partnerships Onboarding API (POST /v2/accounts).
func CreateAccount(
	obs *observability.Observability,
	_ *rzpsdk.Client,
) mcpgo.Tool {
	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Extract bearer token
		token, err := extractString(&r, "bearer_token", true)
		if err != nil || token == nil {
			return mcpgo.NewToolResultError("bearer_token is required"), nil
		}

		payload, warnings, _ := buildAccountPayload(&r)

		// Block on mandatory field errors
		var mandatoryErrors []string
		for _, w := range warnings {
			if strings.HasPrefix(w, "MISSING MANDATORY") {
				mandatoryErrors = append(mandatoryErrors, w)
			}
		}
		if len(mandatoryErrors) > 0 {
			return mcpgo.NewToolResultError(
				"Cannot create account — mandatory fields missing:\n- " +
					strings.Join(mandatoryErrors, "\n- "),
			), nil
		}

		account, err := doAccountsRequest(ctx, http.MethodPost, accountsBaseURL(), *token, payload)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("creating account failed", err),
			), nil
		}

		return mcpgo.NewToolResultJSON(account)
	}

	return mcpgo.NewTool(
		"create_account",
		"Create a new sub-merchant account under the partner using the "+
			"Razorpay Partnerships Onboarding API (POST /v2/accounts). "+
			"\n\nRequires a bearer_token from generate_access_token. "+
			"\n\nMandatory: bearer_token, email, phone, legal_business_name, business_type, contact_name. "+
			"\n\nTip: call preview_create_account first to review the constructed payload before submitting.",
		accountParameters(),
		handler,
	)
}

// FetchAccount returns a tool to fetch a sub-merchant account's details by ID.
func FetchAccount(
	obs *observability.Observability,
	_ *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"bearer_token",
			mcpgo.Description("(Mandatory) OAuth access token obtained from generate_access_token."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"account_id",
			mcpgo.Description("Unique identifier of the sub-merchant account. Must start with 'acc_'."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		p := make(map[string]interface{})
		v := newValidator(&r).
			requireString(p, "bearer_token").
			requireString(p, "account_id")
		if result, err := v.handleErrorsIfAny(); result != nil {
			return result, err
		}

		url := fmt.Sprintf("%s/%s", accountsBaseURL(), p["account_id"].(string))
		account, err := doAccountsRequest(ctx, http.MethodGet, url, p["bearer_token"].(string), nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("fetching account failed", err),
			), nil
		}

		return mcpgo.NewToolResultJSON(account)
	}

	return mcpgo.NewTool(
		"fetch_account",
		"Fetch a sub-merchant account's details by ID "+
			"(GET /v2/accounts/:account_id). "+
			"\n\nRequires a bearer_token from generate_access_token. "+
			"\n\nReturns status (created/under_review/needs_clarification/activated/rejected), "+
			"profile, legal info, and contact details.",
		parameters,
		handler,
	)
}
