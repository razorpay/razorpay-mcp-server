package razorpay

import (
	"context"
	"fmt"
	"github.com/razorpay/razorpay-go/constants"
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
			mcpgo.Description("Amount to be paid using the link in smallest currency unit(e.g., ₹300, use 30000)"), // nolint:lll
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
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// validate required parameters
		amount, err := RequiredInt(r, "amount")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		currency, err := RequiredParam[string](r, "currency")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Create request payload
		paymentLinkData := map[string]interface{}{
			"amount":   amount,
			"currency": currency,
		}

		// Add optional description if provided
		desc, err := OptionalParam[string](r, "description")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		if desc != "" {
			paymentLinkData["description"] = desc
		}

		paymentLink, err := client.PaymentLink.Create(paymentLinkData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating payment link failed: %s", err.Error())), nil // nolint:lll
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
func CreateUpiPaymentLink( // nolint:gocyclo
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
		// validate required parameters
		amount, err := RequiredInt(r, "amount")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		currency := "INR"

		// Create request payload
		paymentLinkData := map[string]interface{}{
			"amount":   amount,
			"currency": currency,
			"upi_link": "true",
		}

		// Add optional description if provided
		desc, err := OptionalParam[string](r, "description")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if desc != "" {
			paymentLinkData["description"] = desc
		}

		// Add optional accept_partial if provided
		acceptPartial, err := OptionalParam[bool](r, "accept_partial")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if acceptPartial {
			paymentLinkData["accept_partial"] = acceptPartial
		}

		// Add optional first_min_partial_amount if provided
		firstMinPartialAmount, err := OptionalInt(r, "first_min_partial_amount")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if firstMinPartialAmount > 0 {
			paymentLinkData["first_min_partial_amount"] = firstMinPartialAmount
		}

		// Add optional expire_by if provided
		expireBy, err := OptionalInt(r, "expire_by")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if expireBy > 0 {
			paymentLinkData["expire_by"] = expireBy
		}

		// Add optional reference_id if provided
		referenceID, err := OptionalParam[string](r, "reference_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if referenceID != "" {
			paymentLinkData["reference_id"] = referenceID
		}

		// Handle customer details if any are provided
		customerName, err := OptionalParam[string](r, "customer_name")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		customerEmail, err := OptionalParam[string](r, "customer_email")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		customerContact, err := OptionalParam[string](r, "customer_contact")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		if customerName != "" || customerEmail != "" || customerContact != "" {
			customer := make(map[string]interface{})
			if customerName != "" {
				customer["name"] = customerName
			}
			if customerEmail != "" {
				customer["email"] = customerEmail
			}
			if customerContact != "" {
				customer["contact"] = customerContact
			}
			paymentLinkData["customer"] = customer
		}

		// Handle notification settings if any are provided
		notify := make(map[string]interface{})

		notifySMS, err := OptionalParam[bool](r, "notify_sms")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		notifyEmail, err := OptionalParam[bool](r, "notify_email")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if notifySMS {
			notify["sms"] = notifySMS
		}
		if notifyEmail {
			notify["email"] = notifyEmail
		}
		paymentLinkData["notify"] = notify

		// Add optional reminder_enable if provided
		reminderEnable, err := OptionalParam[bool](r, "reminder_enable")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		paymentLinkData["reminder_enable"] = reminderEnable

		// Add optional notes if provided
		notes, err := OptionalParam[map[string]interface{}](r, "notes")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if len(notes) > 0 {
			paymentLinkData["notes"] = notes
		}

		// Add optional callback_url if provided
		callbackURL, err := OptionalParam[string](r, "callback_url")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if callbackURL != "" {
			paymentLinkData["callback_url"] = callbackURL

			// If callback_url is set, callback_method should be set to 'get'
			callbackMethod, err := OptionalParam[string](r, "callback_method")
			if result, err := HandleValidationError(err); result != nil {
				return result, err
			}
			if callbackMethod != "" {
				paymentLinkData["callback_method"] = callbackMethod
			} else {
				// Default to 'get' if not specified
				paymentLinkData["callback_method"] = "get"
			}
		}

		paymentLink, err := client.PaymentLink.Create(paymentLinkData, nil)
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
			mcpgo.Description("ID of the payment link to be fetched(ID should have a plink_ prefix)."), // nolint:lll
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Use the helper function to get the required parameter
		id, err := RequiredParam[string](r, "payment_link_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		paymentLink, err := client.PaymentLink.Fetch(id, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payment link failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"fetch_payment_link",
		"Fetch payment link details using it's ID."+
			"Response contains the basic details like amount, status etc",
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
		// Validate required parameters
		paymentLinkID, err := RequiredParam[string](r, "payment_link_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		medium, err := RequiredParam[string](r, "medium")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Validate medium is either "sms" or "email"
		if medium != "sms" && medium != "email" {
			return mcpgo.NewToolResultError(
				"medium must be either 'sms' or 'email'"), nil
		}

		// Call the SDK function
		response, err := client.PaymentLink.NotifyBy(paymentLinkID, medium, nil, nil)
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
		// Validate required parameters
		paymentLinkID, err := RequiredParam[string](r, "payment_link_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}

		// Create update payload
		updateData := make(map[string]interface{})

		// Add optional reference_id if provided
		referenceID, err := OptionalParam[string](r, "reference_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if referenceID != "" {
			updateData["reference_id"] = referenceID
		}

		// Add optional expire_by if provided
		expireBy, err := OptionalInt(r, "expire_by")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if expireBy > 0 {
			updateData["expire_by"] = expireBy
		}

		// Add optional reminder_enable if provided
		reminderEnable, err := OptionalParam[bool](r, "reminder_enable")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if reminderEnable {
			updateData["reminder_enable"] = reminderEnable
		}

		// Add optional accept_partial if provided
		acceptPartial, err := OptionalParam[bool](r, "accept_partial")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if acceptPartial {
			updateData["accept_partial"] = acceptPartial
		}

		// Add optional notes if provided
		notes, err := OptionalParam[map[string]interface{}](r, "notes")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if len(notes) > 0 {
			updateData["notes"] = notes
		}

		// Ensure we have at least one field to update
		if len(updateData) == 0 {
			return mcpgo.NewToolResultError(
				"at least one field to update must be provided"), nil
		}

		// Call the SDK function
		paymentLink, err := client.PaymentLink.Update(paymentLinkID, updateData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("updating payment link failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(paymentLink)
	}

	return mcpgo.NewTool(
		"update_payment_link",
		"Update an existing standard payment link with new details such as reference ID, "+
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
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Create query parameters map
		queryParams := make(map[string]interface{})

		// Add optional payment_id if provided
		paymentID, err := OptionalParam[string](r, "payment_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if paymentID != "" {
			queryParams["payment_id"] = paymentID
		}

		// Add optional reference_id if provided
		referenceID, err := OptionalParam[string](r, "reference_id")
		if result, err := HandleValidationError(err); result != nil {
			return result, err
		}
		if referenceID != "" {
			queryParams["reference_id"] = referenceID
		}

		// To fetch all payment links, we'll use the API endpoint without a specific payment link ID
		url := fmt.Sprintf("/%s%s", constants.VERSION_V1, constants.PaymentLink_URL)

		// Call the API directly using the Request object
		response, err := client.PaymentLink.Request.Get(url, queryParams, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching payment links failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"fetch_all_payment_links",
		"Fetch all payment links with optional filtering by payment ID or reference ID",
		parameters,
		handler,
	)
}
