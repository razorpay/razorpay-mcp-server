package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// CreatePaymentLink returns a tool that creates payment links in Razorpay
func CreatePaymentLink(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Amount to be paid using the link in smallest "+
				"currency unit(e.g., ₹300, use 30000)"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("Three-letter ISO code for the currency (e.g., INR)"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("A brief description of the Payment Link explaining the intent of the payment."), // nolint:lll
		),
		mcpgo.WithBoolean(
			"accept_partial",
			mcpgo.Description("Indicates whether customers can make partial payments using the Payment Link. Default: false"), // nolint:lll
		),
		mcpgo.WithNumber(
			"first_min_partial_amount",
			mcpgo.Description("Minimum amount that must be paid by the customer as the first partial payment. Default value is 100."), // nolint:lll
		),
		mcpgo.WithNumber(
			"expire_by",
			mcpgo.Description("Timestamp, in Unix, when the Payment Link will expire. By default, a Payment Link will be valid for six months."), // nolint:lll
		),
		mcpgo.WithString(
			"reference_id",
			mcpgo.Description("Reference number tagged to a Payment Link. Must be unique for each Payment Link. Max 40 characters."), // nolint:lll
		),
		mcpgo.WithString(
			"customer_name",
			mcpgo.Description("Name of the customer."),
		),
		mcpgo.WithString(
			"customer_email",
			mcpgo.Description("Email address of the customer."),
		),
		mcpgo.WithString(
			"customer_contact",
			mcpgo.Description("Contact number of the customer."),
		),
		mcpgo.WithBoolean(
			"notify_sms",
			mcpgo.Description("Send SMS notifications for the Payment Link."),
		),
		mcpgo.WithBoolean(
			"notify_email",
			mcpgo.Description("Send email notifications for the Payment Link."),
		),
		mcpgo.WithBoolean(
			"reminder_enable",
			mcpgo.Description("Enable payment reminders for the Payment Link."),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs that can be used to store additional information. Maximum 15 pairs, each value limited to 256 characters."), // nolint:lll
		),
		mcpgo.WithString(
			"callback_url",
			mcpgo.Description("If specified, adds a redirect URL to the Payment Link. Customer will be redirected here after payment."), // nolint:lll
		),
		mcpgo.WithString(
			"callback_method",
			mcpgo.Description("HTTP method for callback redirection. "+
				"Must be 'get' if callback_url is set."),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Create a parameters map to collect validated parameters
		plCreateReq := make(map[string]interface{})
		customer := make(map[string]interface{})
		notify := make(map[string]interface{})
		// Validate all parameters with fluent validator
		validator := NewValidator(&r).
			ValidateAndAddRequiredInt(plCreateReq, "amount").
			ValidateAndAddRequiredString(plCreateReq, "currency").
			ValidateAndAddOptionalString(plCreateReq, "description").
			ValidateAndAddOptionalBool(plCreateReq, "accept_partial").
			ValidateAndAddOptionalInt(plCreateReq, "first_min_partial_amount").
			ValidateAndAddOptionalInt(plCreateReq, "expire_by").
			ValidateAndAddOptionalString(plCreateReq, "reference_id").
			ValidateAndAddOptionalStringTo(customer, "customer_name", "name").
			ValidateAndAddOptionalStringTo(customer, "customer_email", "email").
			ValidateAndAddOptionalStringTo(customer, "customer_contact", "contact").
			ValidateAndAddOptionalBoolTo(notify, "notify_sms", "sms").
			ValidateAndAddOptionalBoolTo(notify, "notify_email", "email").
			ValidateAndAddOptionalBool(plCreateReq, "reminder_enable").
			ValidateAndAddOptionalMap(plCreateReq, "notes").
			ValidateAndAddOptionalString(plCreateReq, "callback_url").
			ValidateAndAddOptionalString(plCreateReq, "callback_method")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Handle customer details
		if len(customer) > 0 {
			plCreateReq["customer"] = customer
		}

		// Handle notification settings
		if len(notify) > 0 {
			plCreateReq["notify"] = notify
		}

		// Create the payment link
		paymentLink, err := client.PaymentLink.Create(plCreateReq, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating payment link failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"create_payment_link",
		"Create a new standard payment link in Razorpay with a specified amount",
		parameters,
		handler,
	)
}

// CreateUpiPaymentLink returns a tool that creates payment links in Razorpay
func CreateUpiPaymentLink(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithNumber(
			"amount",
			mcpgo.Description("Amount to be paid using the link in smallest currency unit(e.g., ₹300, use 30000), Only accepted currency is INR"), // nolint:lll
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"currency",
			mcpgo.Description("Three-letter ISO code for the currency (e.g., INR). UPI links are only supported in INR"), // nolint:lll
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"description",
			mcpgo.Description("A brief description of the Payment Link explaining the intent of the payment."), // nolint:lll
		),
		mcpgo.WithBoolean(
			"accept_partial",
			mcpgo.Description("Indicates whether customers can make partial payments using the Payment Link. Default: false"), // nolint:lll
		),
		mcpgo.WithNumber(
			"first_min_partial_amount",
			mcpgo.Description("Minimum amount that must be paid by the customer as the first partial payment. Default value is 100."), // nolint:lll
		),
		mcpgo.WithNumber(
			"expire_by",
			mcpgo.Description("Timestamp, in Unix, when the Payment Link will expire. By default, a Payment Link will be valid for six months."), // nolint:lll
		),
		mcpgo.WithString(
			"reference_id",
			mcpgo.Description("Reference number tagged to a Payment Link. Must be unique for each Payment Link. Max 40 characters."), // nolint:lll
		),
		mcpgo.WithString(
			"customer_name",
			mcpgo.Description("Name of the customer."),
		),
		mcpgo.WithString(
			"customer_email",
			mcpgo.Description("Email address of the customer."),
		),
		mcpgo.WithString(
			"customer_contact",
			mcpgo.Description("Contact number of the customer."),
		),
		mcpgo.WithBoolean(
			"notify_sms",
			mcpgo.Description("Send SMS notifications for the Payment Link."),
		),
		mcpgo.WithBoolean(
			"notify_email",
			mcpgo.Description("Send email notifications for the Payment Link."),
		),
		mcpgo.WithBoolean(
			"reminder_enable",
			mcpgo.Description("Enable payment reminders for the Payment Link."),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs that can be used to store additional information. Maximum 15 pairs, each value limited to 256 characters."), // nolint:lll
		),
		mcpgo.WithString(
			"callback_url",
			mcpgo.Description("If specified, adds a redirect URL to the Payment Link. Customer will be redirected here after payment."), // nolint:lll
		),
		mcpgo.WithString(
			"callback_method",
			mcpgo.Description("HTTP method for callback redirection. "+
				"Must be 'get' if callback_url is set."),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Create a parameters map to collect validated parameters
		upiPlCreateReq := make(map[string]interface{})
		customer := make(map[string]interface{})
		notify := make(map[string]interface{})
		// Validate all parameters with fluent validator
		validator := NewValidator(&r).
			ValidateAndAddRequiredInt(upiPlCreateReq, "amount").
			ValidateAndAddRequiredString(upiPlCreateReq, "currency").
			ValidateAndAddOptionalString(upiPlCreateReq, "description").
			ValidateAndAddOptionalBool(upiPlCreateReq, "accept_partial").
			ValidateAndAddOptionalInt(upiPlCreateReq, "first_min_partial_amount").
			ValidateAndAddOptionalInt(upiPlCreateReq, "expire_by").
			ValidateAndAddOptionalString(upiPlCreateReq, "reference_id").
			ValidateAndAddOptionalStringTo(customer, "customer_name", "name").
			ValidateAndAddOptionalStringTo(customer, "customer_email", "email").
			ValidateAndAddOptionalStringTo(customer, "customer_contact", "contact").
			ValidateAndAddOptionalBoolTo(notify, "notify_sms", "sms").
			ValidateAndAddOptionalBoolTo(notify, "notify_email", "email").
			ValidateAndAddOptionalBool(upiPlCreateReq, "reminder_enable").
			ValidateAndAddOptionalMap(upiPlCreateReq, "notes").
			ValidateAndAddOptionalString(upiPlCreateReq, "callback_url").
			ValidateAndAddOptionalString(upiPlCreateReq, "callback_method")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Add the required UPI payment link parameters
		upiPlCreateReq["upi_link"] = "true"

		// Handle customer details
		if len(customer) > 0 {
			upiPlCreateReq["customer"] = customer
		}

		// Handle notification settings
		if len(notify) > 0 {
			upiPlCreateReq["notify"] = notify
		}

		// Create the payment link
		paymentLink, err := client.PaymentLink.Create(upiPlCreateReq, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("upi pl create failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"payment_link_upi.create",
		"Create a new UPI payment link in Razorpay with a specified amount and additional options.", // nolint:lll
		parameters,
		handler,
	)
}

// FetchPaymentLink returns a tool that fetches payment link details using
// payment_link_id
func FetchPaymentLink(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_link_id",
			mcpgo.Description("ID of the payment link to be fetched"+
				"(ID should have a plink_ prefix)."),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		fields := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(fields, "payment_link_id")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		paymentLinkId := fields["payment_link_id"].(string)

		paymentLink, err := client.PaymentLink.Fetch(paymentLinkId, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payment link failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"fetch_payment_link",
		"Fetch payment link details using it's ID. "+
			"Response contains the basic details like amount, status etc. "+
			"The link could be of any type(standard or UPI)",
		parameters,
		handler,
	)
}

// ResendPaymentLinkNotification returns a tool that sends/resends notifications
// for a payment link via email or SMS
func ResendPaymentLinkNotification(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_link_id",
			mcpgo.Description("ID of the payment link for which to send notification "+
				"(ID should have a plink_ prefix)."), // nolint:lll
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"medium",
			mcpgo.Description("Medium through which to send the notification. "+
				"Must be either 'sms' or 'email'."), // nolint:lll
			mcpgo.Required(),
			mcpgo.Enum("sms", "email"),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		fields := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(fields, "payment_link_id").
			ValidateAndAddRequiredString(fields, "medium")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		paymentLinkId := fields["payment_link_id"].(string)
		medium := fields["medium"].(string)

		// Call the SDK function
		response, err := client.PaymentLink.NotifyBy(paymentLinkId, medium, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("sending notification failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"payment_link_notify",
		"Send or resend notification for a payment link via SMS or email.", // nolint:lll
		parameters,
		handler,
	)
}

// UpdatePaymentLink returns a tool that updates an existing payment link
func UpdatePaymentLink(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_link_id",
			mcpgo.Description("ID of the payment link to update "+
				"(ID should have a plink_ prefix)."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"reference_id",
			mcpgo.Description("Adds a unique reference number to the payment link."),
		),
		mcpgo.WithNumber(
			"expire_by",
			mcpgo.Description("Timestamp, in Unix format, when the payment link "+
				"should expire."),
		),
		mcpgo.WithBoolean(
			"reminder_enable",
			mcpgo.Description("Enable or disable reminders for the payment link."),
		),
		mcpgo.WithBoolean(
			"accept_partial",
			mcpgo.Description("Allow customers to make partial payments. "+
				"Not allowed with UPI payment links."),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for additional information. "+
				"Maximum 15 pairs, each value limited to 256 characters."),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		plUpdateReq := make(map[string]interface{})
		otherFields := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(otherFields, "payment_link_id").
			ValidateAndAddOptionalString(plUpdateReq, "reference_id").
			ValidateAndAddOptionalInt(plUpdateReq, "expire_by").
			ValidateAndAddOptionalBool(plUpdateReq, "reminder_enable").
			ValidateAndAddOptionalBool(plUpdateReq, "accept_partial").
			ValidateAndAddOptionalMap(plUpdateReq, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		paymentLinkId := otherFields["payment_link_id"].(string)

		// Ensure we have at least one field to update
		if len(plUpdateReq) == 0 {
			return mcpgo.NewToolResultError(
				"at least one field to update must be provided"), nil
		}

		// Call the SDK function
		paymentLink, err := client.PaymentLink.Update(paymentLinkId, plUpdateReq, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("updating payment link failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"update_payment_link",
		"Update any existing standard or UPI payment link with new details such as reference ID, "+ // nolint:lll
			"expiry date, or notes.",
		parameters,
		handler,
	)
}

// FetchAllPaymentLinks returns a tool that fetches all payment links with optional filtering
func FetchAllPaymentLinks(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"payment_id",
			mcpgo.Description("Optional: Filter by payment ID associated with payment links"),
		),
		mcpgo.WithString(
			"reference_id",
			mcpgo.Description("Optional: Filter by reference ID used when creating payment links"),
		),
		mcpgo.WithNumber(
			"upi_link",
			mcpgo.Description("Optional: Filter only upi links. "+
				"Value should be 1 if you want only upi links, 0 for only standard links"+
				"If not provided, all types of links will be returned"),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		plListReq := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddOptionalString(plListReq, "payment_id").
			ValidateAndAddOptionalString(plListReq, "reference_id").
			ValidateAndAddOptionalInt(plListReq, "upi_link")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Call the API directly using the Request object
		response, err := client.PaymentLink.All(plListReq, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payment links failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"fetch_all_payment_links",
		"Fetch all payment links with optional filtering by payment ID or reference ID."+ // nolint:lll
			"You can specify the upi_link parameter to filter by link type.",
		parameters,
		handler,
	)
}
