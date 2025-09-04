package razorpay

import (
	"context"
	"fmt"
	"strings"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// CreateCustomer returns a tool that creates a new customer with basic details
func CreateCustomer(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"name",
			mcpgo.Description("Name of the customer. Maximum 50 characters."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"contact",
			mcpgo.Description("Contact number of the customer. "+
				"For example, +11234567890"),
		),
		mcpgo.WithString(
			"email",
			mcpgo.Description("Email address of the customer. "+
				"For example, john.smith@example.com"),
		),
		mcpgo.WithString(
			"fail_existing",
			mcpgo.Description("Parameter to determine the action if the customer "+
				"already exists. Possible values: '0' (default) - Get existing customer, "+
				"'1' - Fail if customer exists"),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for additional information about "+
				"the customer. Maximum 15 key-value pairs, 256 characters each."),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		// Create request payload
		customerData := make(map[string]interface{})

		validator := NewValidator(&r)

		// Validate required name parameter with empty string check
		if nameValue, err := extractValueGeneric[string](&r, "name", true); err != nil {
			validator = validator.addError(err)
		} else if nameValue != nil && *nameValue == "" {
			validator = validator.addError(fmt.Errorf("missing required parameter: name"))
		} else if nameValue != nil {
			customerData["name"] = *nameValue
		}

		validator = validator.
			ValidateAndAddOptionalString(customerData, "contact").
			ValidateAndAddOptionalString(customerData, "email").
			ValidateAndAddOptionalString(customerData, "fail_existing").
			ValidateAndAddOptionalMap(customerData, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Create the customer using Razorpay SDK
		customer, err := client.Customer.Create(customerData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("creating customer failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(customer)
	}

	return mcpgo.NewTool(
		"create_customer",
		"Use this endpoint to create or add a customer with basic details "+
			"such as name and contact details.",
		parameters,
		handler,
	)
}

// FetchCustomer returns a tool that fetches customer details using customer_id
func FetchCustomer(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description("Unique identifier of the customer to be retrieved. "+
				"Must start with 'cust_'"),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})

		validator := NewValidator(&r)

		// Validate required customer_id parameter with empty string check
		if customerIdValue, err := extractValueGeneric[string](&r, "customer_id", true); err != nil {
			validator = validator.addError(err)
		} else if customerIdValue != nil && *customerIdValue == "" {
			validator = validator.addError(fmt.Errorf("missing required parameter: customer_id"))
		} else if customerIdValue != nil {
			params["customer_id"] = *customerIdValue
		}

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		customerId := params["customer_id"].(string)

		// Fetch the customer using Razorpay SDK
		customer, err := client.Customer.Fetch(customerId, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching customer failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(customer)
	}

	return mcpgo.NewTool(
		"fetch_customer",
		"Use this tool to retrieve the details of a specific customer using its id.",
		parameters,
		handler,
	)
}

// EditCustomer returns a tool that updates customer details
func EditCustomer(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description("Unique identifier of the customer to be updated. "+
				"Must start with 'cust_'"),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"name",
			mcpgo.Description("Updated name of the customer. Maximum 50 characters."),
		),
		mcpgo.WithString(
			"contact",
			mcpgo.Description("Updated contact number of the customer. "+
				"For example, +11234567890"),
		),
		mcpgo.WithString(
			"email",
			mcpgo.Description("Updated email address of the customer. "+
				"For example, john.smith@example.com"),
		),
		mcpgo.WithObject(
			"notes",
			mcpgo.Description("Key-value pairs for additional information about "+
				"the customer. Maximum 15 key-value pairs, 256 characters each."),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		params := make(map[string]interface{})
		customerUpdateData := make(map[string]interface{})

		validator := NewValidator(&r)

		// Validate required customer_id parameter with empty string check
		if customerIdValue, err := extractValueGeneric[string](&r, "customer_id", true); err != nil {
			validator = validator.addError(err)
		} else if customerIdValue != nil && *customerIdValue == "" {
			validator = validator.addError(fmt.Errorf("missing required parameter: customer_id"))
		} else if customerIdValue != nil {
			params["customer_id"] = *customerIdValue
		}

		validator = validator.
			ValidateAndAddOptionalString(customerUpdateData, "name").
			ValidateAndAddOptionalString(customerUpdateData, "contact").
			ValidateAndAddOptionalString(customerUpdateData, "email").
			ValidateAndAddOptionalMap(customerUpdateData, "notes")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		customerId := params["customer_id"].(string)

		// Update the customer using Razorpay SDK
		customer, err := client.Customer.Edit(customerId, customerUpdateData, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("updating customer failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(customer)
	}

	return mcpgo.NewTool(
		"edit_customer",
		"Use this tool to update customer details such as name, contact, email, and notes.",
		parameters,
		handler,
	)
}

// FetchAllCustomers returns a tool that fetches multiple customers with filtering and pagination
func FetchAllCustomers(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		// Pagination parameters
		mcpgo.WithNumber(
			"count",
			mcpgo.Description("Number of customers to fetch "+
				"(default: 10, max: 100)"),
			mcpgo.Min(1),
			mcpgo.Max(100),
		),
		mcpgo.WithNumber(
			"skip",
			mcpgo.Description("Number of customers to skip (default: 0)"),
			mcpgo.Min(0),
		),
		// Time range filters
		mcpgo.WithNumber(
			"from",
			mcpgo.Description("Unix timestamp (in seconds) from when "+
				"customers are to be fetched"),
			mcpgo.Min(0),
		),
		mcpgo.WithNumber(
			"to",
			mcpgo.Description("Unix timestamp (in seconds) up till when "+
				"customers are to be fetched"),
			mcpgo.Min(0),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		// Get client from context or use default
		client, err := getClientFromContextOrDefault(ctx, client)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		// Create query parameters map
		customerListOptions := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddPagination(customerListOptions).
			ValidateAndAddOptionalInt(customerListOptions, "from").
			ValidateAndAddOptionalInt(customerListOptions, "to")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		// Fetch all customers using Razorpay SDK
		customers, err := client.Customer.All(customerListOptions, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("fetching customers failed: %s", err.Error())), nil
		}

		return mcpgo.NewToolResultJSON(customers)
	}

	return mcpgo.NewTool(
		"fetch_all_customers",
		"Fetch all customers with optional filtering and pagination",
		parameters,
		handler,
	)
}

// FetchCustomerTokens creates a tool to fetch all tokens for a specific customer
func FetchCustomerTokens(obs *observability.Observability, client *rzpsdk.Client) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"customer_id",
			mcpgo.Description("The unique identifier of the customer for whom tokens are to be retrieved. Must start with 'cust_'"),
			mcpgo.Required(),
		),
	}

	handler := func(ctx context.Context, r mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
		validator := NewValidator(&r)
		params := make(map[string]interface{})

		// Validate required customer_id parameter with empty string check
		if customerIDValue, err := extractValueGeneric[string](&r, "customer_id", true); err != nil {
			validator = validator.addError(err)
		} else if customerIDValue != nil && *customerIDValue == "" {
			validator = validator.addError(fmt.Errorf("missing required parameter: customer_id"))
		} else if customerIDValue != nil {
			// Validate customer_id format
			if !strings.HasPrefix(*customerIDValue, "cust_") {
				validator = validator.addError(fmt.Errorf("customer_id must start with 'cust_'"))
			} else {
				params["customer_id"] = *customerIDValue
			}
		}

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		customerID := params["customer_id"].(string)

		// Create the API endpoint URL
		url := fmt.Sprintf("/%s/customers/%s/tokens", constants.VERSION_V1, customerID)

		// Make the API request
		response, err := client.Request.Get(url, nil, nil)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("Failed to fetch customer tokens: %v", err)), nil
		}

		return mcpgo.NewToolResultJSON(response)
	}

	return mcpgo.NewTool(
		"fetch_customer_tokens",
		"Fetch all tokens for a specific customer. Returns active tokens that can be used for recurring payments",
		parameters,
		handler,
	)
}
