package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// CreateBill creates a new bill (invoice) in Razorpay
func CreateBill(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"type",
			mcpgo.Description(
				"Type of the document. Use 'invoice' for bills."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description(
				"Unique identifier of the customer "+
					"the bill is created for. "+
					"Either customer_id or customer "+
					"details (name/email/contact) "+
					"must be provided."),
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
		mcpgo.WithArray(
			"line_items",
			mcpgo.Description(
				"Array of line items in the bill. "+
					"Each item must have 'name' and "+
					"'amount' (in paise, smallest "+
					"currency unit). Optional fields: "+
					"'description', 'quantity', 'unit', "+
					"'unit_amount', 'item_id'."),
			mcpgo.Items(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the line item",
					},
					"amount": map[string]interface{}{
						"type":        "number",
						"description": "Amount in paise (smallest currency unit). For INR: 100 paise = ₹1", //nolint:lll
						"minimum":     1,
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Description of the line item",
					},
					"quantity": map[string]interface{}{
						"type":        "number",
						"description": "Quantity of the item",
						"minimum":     1,
					},
					"unit": map[string]interface{}{
						"type":        "string",
						"description": "Unit of the item (e.g., 'kg', 'pcs')",
					},
				},
				"required": []interface{}{"name", "amount"},
			}),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description(
				"Brief description of the bill."),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description(
				"Three-letter ISO currency code "+
					"(e.g., INR, USD)."),
			mcpgo.Pattern("^[A-Z]{3}$"),
		),
		mcpgo.WithNumber(
			"draft",
			mcpgo.Description(
				"Set to 1 to save bill as draft. "+
					"Set to 0 to issue immediately. "+
					"Default: 0"),
			mcpgo.Min(0),
			mcpgo.Max(1),
		),
		mcpgo.WithNumber(
			"date",
			mcpgo.Description(
				"Unix timestamp for the bill date."),
		),
		mcpgo.WithNumber(
			"due_date",
			mcpgo.Description(
				"Unix timestamp for payment due date."),
		),
		mcpgo.WithBoolean(
			"partial_payment",
			mcpgo.Description(
				"Whether the customer can make "+
					"partial payments. Default: false"),
		),
		mcpgo.WithString(
			"terms",
			mcpgo.Description(
				"Terms and conditions for the bill."),
		),
		mcpgo.WithBoolean(
			"email_notify",
			mcpgo.Description(
				"Send email notification to customer. "+
					"Default: true"),
		),
		mcpgo.WithBoolean(
			"sms_notify",
			mcpgo.Description(
				"Send SMS notification to customer. "+
					"Default: true"),
		),
		mcpgo.WithNumber(
			"expire_by",
			mcpgo.Description(
				"Unix timestamp when the bill expires."),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description(
				"Key-value pairs for additional "+
					"information. Max 15 pairs, "+
					"each up to 256 characters."),
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

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(payload, "type").
			ValidateAndAddOptionalString(payload, "customer_id").
			ValidateAndAddOptionalStringToPath(
				customer, "customer_name", "name").
			ValidateAndAddOptionalStringToPath(
				customer, "customer_email", "email").
			ValidateAndAddOptionalStringToPath(
				customer, "customer_contact", "contact").
			ValidateAndAddOptionalArray(payload, "line_items").
			ValidateAndAddOptionalString(payload, "description").
			ValidateAndAddOptionalString(payload, "currency").
			ValidateAndAddOptionalInt(payload, "draft").
			ValidateAndAddOptionalInt(payload, "date").
			ValidateAndAddOptionalInt(payload, "due_date").
			ValidateAndAddOptionalBool(payload, "partial_payment").
			ValidateAndAddOptionalString(payload, "terms").
			ValidateAndAddOptionalBool(payload, "email_notify").
			ValidateAndAddOptionalBool(payload, "sms_notify").
			ValidateAndAddOptionalInt(payload, "expire_by").
			ValidateAndAddOptionalMap(payload, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		if len(customer) > 0 {
			payload["customer"] = customer
		}

		bill, err := client.Invoice.Create(payload, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"creating bill failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(bill)
	}

	return mcpgo.NewTool(
		"create_bill",
		"Create a new bill (invoice) in Razorpay. "+
			"Use when you need to generate a bill for a "+
			"customer with itemised line items. "+
			"Provide customer_id or customer details "+
			"(name/email/contact) along with line_items. "+
			"Set draft=1 to save without issuing; "+
			"omit or set draft=0 to issue immediately. "+
			"Amounts in line_items are in paise "+
			"(100 paise = ₹1 for INR).",
		parameters,
		handler,
	)
}

// FetchBill fetches a specific bill by its ID
func FetchBill(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"bill_id",
			mcpgo.Description(
				"Unique identifier of the bill to "+
					"retrieve. Starts with 'inv_'."),
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
			ValidateAndAddRequiredString(payload, "bill_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		billID := payload["bill_id"].(string)
		bill, err := client.Invoice.Fetch(billID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"fetching bill failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(bill)
	}

	return mcpgo.NewTool(
		"fetch_bill",
		"Fetch details of a specific bill by its ID. "+
			"Use when you need bill status, amount, "+
			"customer details, or payment info. "+
			"The bill ID starts with 'inv_'.",
		parameters,
		handler,
	)
}

// FetchAllBills fetches all bills with optional filtering
func FetchAllBills(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"type",
			mcpgo.Description(
				"Filter by document type. "+
					"Use 'invoice' for bills. "+
					"Default: 'invoice'"),
			mcpgo.DefaultValue("invoice"),
		),
		mcpgo.WithNumber(
			"count",
			mcpgo.Description(
				"Number of bills to fetch. "+
					"Default: 10, Max: 100"),
			mcpgo.Min(1),
			mcpgo.Max(100),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description(
				"Number of bills to skip. "+
					"Default: 0"),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"from",
			mcpgo.Description(
				"Unix timestamp. Fetch bills "+
					"created on or after this time."),
		),
		mcpgo.WithNumber(
			"to",
			mcpgo.Description(
				"Unix timestamp. Fetch bills "+
					"created on or before this time."),
		),
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description(
				"Filter bills by customer ID."),
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

		queryParams := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddOptionalString(queryParams, "type").
			ValidateAndAddPagination(queryParams).
			ValidateAndAddOptionalInt(queryParams, "from").
			ValidateAndAddOptionalInt(queryParams, "to").
			ValidateAndAddOptionalString(queryParams, "customer_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Default type to "invoice" for bills
		if _, ok := queryParams["type"]; !ok {
			queryParams["type"] = "invoice"
		}

		bills, err := client.Invoice.All(queryParams, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"fetching bills failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(bills)
	}

	return mcpgo.NewTool(
		"fetch_all_bills",
		"Fetch all bills with optional filtering and pagination. "+
			"Use when you need a list of bills for a customer "+
			"or within a date range. "+
			"Defaults to fetching bills of type 'invoice'. "+
			"Supports pagination via count and skip.",
		parameters,
		handler,
	)
}

// UpdateBill updates a bill's notes
func UpdateBill(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"bill_id",
			mcpgo.Description(
				"Unique identifier of the bill to "+
					"update. Starts with 'inv_'."),
			mcpgo.Required(),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description(
				"Key-value pairs for additional "+
					"information. Max 15 pairs, "+
					"each up to 256 characters."),
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
		data := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(payload, "bill_id").
			ValidateAndAddRequiredMap(payload, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		data["notes"] = payload["notes"]
		billID := payload["bill_id"].(string)

		bill, err := client.Invoice.Update(billID, data, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"updating bill failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(bill)
	}

	return mcpgo.NewTool(
		"update_bill",
		"Update the notes of an existing bill. "+
			"Only the notes field can be modified after creation. "+
			"The bill must not be in a terminal state "+
			"(cancelled or expired). "+
			"The bill ID starts with 'inv_'.",
		parameters,
		handler,
	)
}

// IssueBill issues a draft bill
func IssueBill(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"bill_id",
			mcpgo.Description(
				"Unique identifier of the draft "+
					"bill to issue. Starts with 'inv_'."),
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
			ValidateAndAddRequiredString(payload, "bill_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		billID := payload["bill_id"].(string)
		bill, err := client.Invoice.Issue(billID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"issuing bill failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(bill)
	}

	return mcpgo.NewTool(
		"issue_bill",
		"Issue a draft bill to make it active and notify the customer. "+
			"Only works on bills with status 'draft'. "+
			"Use after create_bill with draft=1 when ready to send. "+
			"The bill ID starts with 'inv_'.",
		parameters,
		handler,
	)
}

// CancelBill cancels an issued bill
func CancelBill(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"bill_id",
			mcpgo.Description(
				"Unique identifier of the bill to "+
					"cancel. Starts with 'inv_'."),
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
			ValidateAndAddRequiredString(payload, "bill_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		billID := payload["bill_id"].(string)
		bill, err := client.Invoice.Cancel(billID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"cancelling bill failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(bill)
	}

	return mcpgo.NewTool(
		"cancel_bill",
		"Cancel an issued bill in Razorpay. "+
			"Only works on bills with status 'issued'. "+
			"Cannot be reversed — use with caution. "+
			"The bill ID starts with 'inv_'.",
		parameters,
		handler,
	)
}

// DeleteBill deletes a draft bill
func DeleteBill(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"bill_id",
			mcpgo.Description(
				"Unique identifier of the draft "+
					"bill to delete. Starts with 'inv_'."),
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
			ValidateAndAddRequiredString(payload, "bill_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		billID := payload["bill_id"].(string)
		result, err := client.Invoice.Delete(billID, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"deleting bill failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(result)
	}

	return mcpgo.NewTool(
		"delete_bill",
		"Delete a draft bill permanently. "+
			"Only works on bills with status 'draft'. "+
			"This action cannot be undone. "+
			"Use cancel_bill for issued bills instead. "+
			"The bill ID starts with 'inv_'.",
		parameters,
		handler,
	)
}

// SendBillNotification sends a notification for a bill via email or SMS
func SendBillNotification(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"bill_id",
			mcpgo.Description(
				"Unique identifier of the bill. "+
					"Starts with 'inv_'."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"medium",
			mcpgo.Description(
				"Notification channel. "+
					"Use 'sms' or 'email'."),
			mcpgo.Required(),
			mcpgo.Enum("sms", "email"),
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
			ValidateAndAddRequiredString(payload, "bill_id").
			ValidateAndAddRequiredString(payload, "medium")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		billID := payload["bill_id"].(string)
		medium := payload["medium"].(string)

		url := fmt.Sprintf(
			"/%s%s/%s/notify_by/%s",
			constants.VERSION_V1,
			constants.INVOICE_URL,
			billID,
			medium,
		)

		response, err := client.Invoice.Request.Post(url, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf(
					"sending bill notification failed: %s",
					err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"send_bill_notification",
		"Send a notification for a bill to the customer via SMS or email. "+
			"Use when you need to resend a bill notification. "+
			"The bill must be in 'issued' status. "+
			"medium must be 'sms' or 'email'. "+
			"The bill ID starts with 'inv_'.",
		parameters,
		handler,
	)
}
