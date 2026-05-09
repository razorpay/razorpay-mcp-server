//nolint:lll // descriptions contain long API field documentation strings
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
		mcpgo.WithString("bearer_token",
			mcpgo.Description("(Mandatory) OAuth access token from generate_access_token."),
			mcpgo.Required(),
		),
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
			mcpgo.Description("(Mandatory) Type of business. One of: proprietorship, individual, "+
				"partnership, private_limited, public_limited, llp, ngo, "+
				"educational_institutes, trust, society, not_yet_registered, "+
				"other, huf, government, judicial_person, local_authority, section_8_company."),
			mcpgo.Required(),
		),
		mcpgo.WithString("contact_name",
			mcpgo.Description("(Mandatory) Full name of the contact person (4-255 characters)."),
			mcpgo.Required(),
		),
		mcpgo.WithString("customer_facing_business_name",
			mcpgo.Description("(Optional) Name shown to customers. Defaults to legal_business_name."),
		),
		mcpgo.WithString("reference_id",
			mcpgo.Description("(Optional) Partner's unique reference ID (max 512 characters)."),
			mcpgo.Max(512),
		),
		mcpgo.WithString("profile_category",
			mcpgo.Description("(Optional, required if any profile field is set) Business category. "+
				"One of: ecommerce, education, financial_services, food, gaming, "+
				"government, healthcare, housing, it_and_software, logistics, "+
				"media_and_entertainment, not_for_profit, services, social, "+
				"tours_and_travel, transport, utilities, others."),
		),
		mcpgo.WithString("profile_subcategory",
			mcpgo.Description("(Optional, required if any profile field is set) Business subcategory "+
				"e.g. 'clinic', 'hospital', 'ecommerce_marketplace', 'saas'."),
		),
		mcpgo.WithString("profile_description",
			mcpgo.Description("(Optional) Short description of the business (max 255 characters)."),
			mcpgo.Max(255),
		),
		mcpgo.WithString("profile_business_model",
			mcpgo.Description("(Optional) Description of the business model e.g. 'b2c', 'b2b'."),
		),
		mcpgo.WithString("registered_street1",
			mcpgo.Description("(Optional, required when profile is set) Registered address line 1."),
			mcpgo.Max(100),
		),
		mcpgo.WithString("registered_street2",
			mcpgo.Description("(Optional, required when profile is set) Registered address line 2."),
			mcpgo.Max(100),
		),
		mcpgo.WithString("registered_city",
			mcpgo.Description("(Optional, required when profile is set) City of registered address."),
		),
		mcpgo.WithString("registered_state",
			mcpgo.Description("(Optional, required when profile is set) State e.g. 'Karnataka', 'MH'."),
		),
		mcpgo.WithNumber("registered_postal_code",
			mcpgo.Description("(Optional, required when profile is set) Postal code e.g. 560034."),
		),
		mcpgo.WithString("registered_country",
			mcpgo.Description("(Optional, required when profile is set) Country ISO code e.g. 'IN'."),
		),
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
			mcpgo.Description("(Optional) Operation address postal code e.g. 560034."),
		),
		mcpgo.WithString("operation_country",
			mcpgo.Description("(Optional) Operation address country ISO code e.g. 'IN'."),
		),
		mcpgo.WithString("pan",
			mcpgo.Description("(Optional) Business PAN card number."),
		),
		mcpgo.WithString("gst",
			mcpgo.Description("(Optional) GST number."),
		),
		mcpgo.WithString("cin",
			mcpgo.Description("(Optional) Corporate Identity Number."),
		),
		mcpgo.WithString("brand_color",
			mcpgo.Description("(Optional) Brand hex color WITHOUT # prefix e.g. 'FFFFFF'."),
			mcpgo.Pattern("^[0-9a-fA-F]{6}$"),
		),
		mcpgo.WithObject("notes",
			mcpgo.Description("(Optional) Key-value pairs for internal reference (max 15 pairs)."),
			mcpgo.MaxProperties(15),
		),
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
		mcpgo.WithArray("websites",
			mcpgo.Description("(Optional) List of website URLs where payments are accepted."),
			mcpgo.Items(map[string]interface{}{"type": "string"}),
		),
	}
}

// buildProfileBlock constructs the profile object from flat request params.
func buildProfileBlock(r *mcpgo.CallToolRequest) (map[string]interface{}, []string) {
	var warnings []string

	cat, _ := extractString(r, "profile_category", false)
	sub, _ := extractString(r, "profile_subcategory", false)
	desc, _ := extractString(r, "profile_description", false)
	bm, _ := extractString(r, "profile_business_model", false)

	rs1, _ := extractString(r, "registered_street1", false)
	rs2, _ := extractString(r, "registered_street2", false)
	rc, _ := extractString(r, "registered_city", false)
	rst, _ := extractString(r, "registered_state", false)
	rp, _ := extractNumber(r, "registered_postal_code", false)
	rco, _ := extractString(r, "registered_country", false)

	os1, _ := extractString(r, "operation_street1", false)
	os2, _ := extractString(r, "operation_street2", false)
	oc, _ := extractString(r, "operation_city", false)
	ost, _ := extractString(r, "operation_state", false)
	op, _ := extractNumber(r, "operation_postal_code", false)
	oco, _ := extractString(r, "operation_country", false)

	hasProfile := cat != nil || sub != nil || desc != nil || bm != nil || rs1 != nil
	if !hasProfile {
		return nil, nil
	}

	profile := map[string]interface{}{}
	if cat == nil {
		warnings = append(warnings,
			"MISSING: profile_category is required when profile is provided")
	} else {
		profile["category"] = *cat
	}
	if sub == nil {
		warnings = append(warnings,
			"MISSING: profile_subcategory is required when profile is provided")
	} else {
		profile["subcategory"] = *sub
	}
	if desc != nil {
		profile["description"] = *desc
	}
	if bm != nil {
		profile["business_model"] = *bm
	}

	registered, regWarnings := buildAddress("registered", rs1, rs2, rc, rst, rp, rco)
	warnings = append(warnings, regWarnings...)

	addresses := map[string]interface{}{"registered": registered}
	if os1 != nil || os2 != nil || oc != nil {
		op_ := buildOperationAddress(os1, os2, oc, ost, op, oco)
		addresses["operation"] = op_
	}
	profile["addresses"] = addresses

	return profile, warnings
}

func buildAddress(
	kind string,
	s1, s2, city, state *string,
	postal *float64,
	country *string,
) (map[string]interface{}, []string) {
	addr := map[string]interface{}{}
	var missing []string
	check := func(field string, val *string) {
		if val != nil {
			addr[field] = *val
		} else {
			missing = append(missing,
				fmt.Sprintf("%s_%s", kind, field))
		}
	}
	check("street1", s1)
	check("street2", s2)
	check("city", city)
	check("state", state)
	if postal != nil {
		addr["postal_code"] = int(*postal)
	} else {
		missing = append(missing, kind+"_postal_code")
	}
	check("country", country)

	var warnings []string
	if len(missing) > 0 {
		warnings = append(warnings,
			"MISSING: "+kind+" address fields required when profile is set: "+
				strings.Join(missing, ", "))
	}
	return addr, warnings
}

func buildOperationAddress(
	s1, s2, city, state *string,
	postal *float64,
	country *string,
) map[string]interface{} {
	addr := map[string]interface{}{}
	set := func(key string, v *string) {
		if v != nil {
			addr[key] = *v
		}
	}
	set("street1", s1)
	set("street2", s2)
	set("city", city)
	set("state", state)
	if postal != nil {
		addr["postal_code"] = int(*postal)
	}
	set("country", country)
	return addr
}

// buildLegalInfoBlock constructs the legal_info object.
func buildLegalInfoBlock(r *mcpgo.CallToolRequest) map[string]interface{} {
	pan, _ := extractString(r, "pan", false)
	gst, _ := extractString(r, "gst", false)
	cin, _ := extractString(r, "cin", false)
	if pan == nil && gst == nil && cin == nil {
		return nil
	}
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
	return legal
}

// buildContactInfoBlock constructs the contact_info object.
func buildContactInfoBlock(r *mcpgo.CallToolRequest) map[string]interface{} {
	sections := []string{"chargeback", "refund", "support", "dispute"}
	fields := []string{"email", "phone", "policy_url"}
	contactInfo := map[string]interface{}{}
	for _, section := range sections {
		obj := map[string]interface{}{}
		for _, field := range fields {
			val, _ := extractString(r, section+"_"+field, false)
			if val != nil {
				obj[field] = *val
			}
		}
		if len(obj) > 0 {
			contactInfo[section] = obj
		}
	}
	if len(contactInfo) == 0 {
		return nil
	}
	return contactInfo
}

// buildAccountPayload constructs the nested API request body from flat params.
func buildAccountPayload(
	r *mcpgo.CallToolRequest,
) (map[string]interface{}, []string, error) {
	v := newValidator(r)
	p := make(map[string]interface{})
	var warnings []string

	// Mandatory fields
	v.requireString(p, "email").
		requireString(p, "phone").
		requireString(p, "legal_business_name").
		requireString(p, "business_type").
		requireString(p, "contact_name")

	if _, _ = v.handleErrorsIfAny(); len(v.errors) > 0 {
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

	// Profile
	if profile, profileWarnings := buildProfileBlock(r); profile != nil {
		p["profile"] = profile
		warnings = append(warnings, profileWarnings...)
	}

	// Legal info
	if legal := buildLegalInfoBlock(r); legal != nil {
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
	if ci := buildContactInfoBlock(r); ci != nil {
		p["contact_info"] = ci
	}

	// Apps - websites
	websites, _ := extractStringArray(r, "websites")
	if len(websites) > 0 {
		p["apps"] = map[string]interface{}{"websites": websites}
	}

	return p, warnings, nil
}

// PreviewCreateAccount returns a tool that previews the request body
// without making the API call.
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
			sb.WriteString("\n## Validation Issues\n\n")
			for _, w := range warnings {
				sb.WriteString(fmt.Sprintf("- %s\n", w))
			}
			sb.WriteString("\nPlease provide the missing fields before proceeding.\n")
		} else {
			sb.WriteString("\nAll mandatory fields are present.\n")
		}

		sb.WriteString("\n---\n")
		sb.WriteString("Would you like to:\n")
		sb.WriteString("- **Add more fields?** Tell me what else to include.\n")
		sb.WriteString("- **Proceed?** Call `create_account` with the same fields.\n")

		return mcpgo.NewToolResultText(sb.String()), nil
	}

	return mcpgo.NewTool(
		"preview_create_account",
		"Preview the sub-merchant account creation request body before submitting. "+
			"Accepts all fields individually, assembles nested JSON, "+
			"and validates mandatory fields. "+
			"After reviewing, call create_account with the same fields to submit.",
		accountParameters(),
		handler,
	)
}

// CreateAccount returns a tool that creates a new sub-merchant account.
func CreateAccount(
	obs *observability.Observability,
	_ *rzpsdk.Client,
) mcpgo.Tool {
	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		token, err := extractString(&r, "bearer_token", true)
		if err != nil || token == nil {
			return mcpgo.NewToolResultError("bearer_token is required"), nil
		}

		payload, warnings, _ := buildAccountPayload(&r)

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

		account, err := doAccountsRequest(
			ctx, http.MethodPost, accountsBaseURL(), *token, payload,
		)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("creating account failed", err),
			), nil
		}

		return mcpgo.NewToolResultJSON(account)
	}

	return mcpgo.NewTool(
		"create_account",
		"Create a new sub-merchant account using the Razorpay Partnerships "+
			"Onboarding API (POST /v2/accounts). "+
			"Requires bearer_token from generate_access_token. "+
			"Tip: call preview_create_account first to review the payload.",
		accountParameters(),
		handler,
	)
}

// FetchAccount returns a tool to fetch a sub-merchant account's details.
func FetchAccount(
	obs *observability.Observability,
	_ *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString("bearer_token",
			mcpgo.Description("(Mandatory) OAuth access token from generate_access_token."),
			mcpgo.Required(),
		),
		mcpgo.WithString("account_id",
			mcpgo.Description("Unique identifier of the sub-merchant account. "+
				"Must start with 'acc_'."),
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

		url := fmt.Sprintf(
			"%s/%s", accountsBaseURL(), p["account_id"].(string),
		)
		account, err := doAccountsRequest(
			ctx, http.MethodGet, url, p["bearer_token"].(string), nil,
		)
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
			"Requires bearer_token from generate_access_token.",
		parameters,
		handler,
	)
}
