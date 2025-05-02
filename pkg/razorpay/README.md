# Razorpay MCP Server Tools

This package contains tools for interacting with the Razorpay API via the Model Context Protocol (MCP).

## Creating New API Tools

This guide explains how to add new Razorpay API tools to the MCP server.

### Quick Start

1. Locate the API documentation at https://razorpay.com/docs/api/
2. Identify the equivalent function call for the API in the razorpay go sdk.
3. Create a new tool function in the appropriate file (or create a new file for a new resource type). Add validations for mandatory fields and call the sdk
5. Register the tool in `server.go`
6. Update "Available Tools" section in the main README.md

### Tool Structure

Add the tool definition inside pkg/razorpay's resource file. You can define a new tool using this following template:

```go
// ToolName returns a tool that [description of what it does]
func ToolName(
    log *slog.Logger,
    client *rzpsdk.Client,
) mcpgo.Tool {
    parameters := []mcpgo.ToolParameter{
        // Parameters defined here
    }

    handler := func(
        ctx context.Context,
        r mcpgo.CallToolRequest,
    ) (*mcpgo.ToolResult, error) {
        // Parameter validation
        // API call
        // Response handling
        return mcpgo.NewToolResultJSON(response)
    }

    return mcpgo.NewTool(
        "tool_name",
        "A description of the tool. NOTE: Add any exceptions/rules if relevant for the LLMs.",
        parameters,
        handler,
    )
}
```

Tool Naming Conventions:
   - Fetch methods: `fetch_resource`
   - Create methods: `create_resource`
   - FetchAll methods: `fetch_all_resources`

### Parameter Definition

Define parameters using the mcpgo helpers. This would include the type, name, description of the parameter and also specifying if the parameter required or not.

```go
// Required parameters
mcpgo.WithString(
    "parameter_name",
    mcpgo.Description("Description of the parameter"),
    mcpgo.Required(),
)

// Optional parameters
mcpgo.WithNumber(
    "amount",
    mcpgo.Description("Amount in smallest currency unit"),
)
```

Available parameter types:
- `WithString`: For string values
- `WithNumber`: For numeric values
- `WithBoolean`: For boolean values
- `WithObject`: For nested objects

### Parameter Validation

Inside the handler function, use the fluent validator pattern for parameter validation. This provides cleaner, more readable code through method chaining:

```go
// Create a new validator
v := NewValidator(&r)

// Create a map for API request parameters
payload := make(map[string]interface{})

// Validate and add parameters to the payload with method chaining
v.ValidateAndAddRequiredString(payload, "id").
  ValidateAndAddOptionalString(payload, "description").
  ValidateAndAddRequiredInt(payload, "amount").
  ValidateAndAddOptionalInt(payload, "limit")

// Validate and add common parameters
v.ValidateAndAddPagination(payload).
  ValidateAndAddExpand(payload)

// Check for validation errors
if result, err := validator.HandleErrorsIfAny(); result != nil {
	return result, err
}

// Proceed with API call using validated parameters in payload
```

### Example: GET Endpoint

```go
// FetchResource returns a tool that fetches a resource by ID
func FetchResource(
    log *slog.Logger,
    client *rzpsdk.Client,
) mcpgo.Tool {
    parameters := []mcpgo.ToolParameter{
        mcpgo.WithString(
            "id",
            mcpgo.Description("Unique identifier of the resource"),
            mcpgo.Required(),
        ),
    }

    handler := func(
        ctx context.Context,
        r mcpgo.CallToolRequest,
    ) (*mcpgo.ToolResult, error) {
        // Create validator and a payload map
        payload := make(map[string]interface{})
        v := NewValidator(&r).
            ValidateAndAddRequiredString(payload, "id")
        
        // Check for validation errors
        if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

        // Extract validated ID and make API call
        id := payload["id"].(string)
        resource, err := client.Resource.Fetch(id, nil, nil)
        if err != nil {
            return mcpgo.NewToolResultError(
                fmt.Sprintf("fetching resource failed: %s", err.Error())), nil
        }

        return mcpgo.NewToolResultJSON(resource)
    }

    return mcpgo.NewTool(
        "fetch_resource",
        "Fetch a resource from Razorpay by ID",
        parameters,
        handler,
    )
}
```

### Example: POST Endpoint

```go
// CreateResource returns a tool that creates a new resource
func CreateResource(
    log *slog.Logger,
    client *rzpsdk.Client,
) mcpgo.Tool {
    parameters := []mcpgo.ToolParameter{
        mcpgo.WithNumber(
            "amount",
            mcpgo.Description("Amount in smallest currency unit"),
            mcpgo.Required(),
        ),
        mcpgo.WithString(
            "currency",
            mcpgo.Description("Three-letter ISO code for the currency"),
            mcpgo.Required(),
        ),
        mcpgo.WithString(
            "description",
            mcpgo.Description("Brief description of the resource"),
        ),
    }

    handler := func(
        ctx context.Context,
        r mcpgo.CallToolRequest,
    ) (*mcpgo.ToolResult, error) {
        // Create payload map and validator
        data := make(map[string]interface{})
        v := NewValidator(&r).
            ValidateAndAddRequiredInt(data, "amount").
            ValidateAndAddRequiredString(data, "currency").
            ValidateAndAddOptionalString(data, "description")
        
        // Check for validation errors
        if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

        // Call the API with validated data
        resource, err := client.Resource.Create(data, nil)
        if err != nil {
            return mcpgo.NewToolResultError(
                fmt.Sprintf("creating resource failed: %s", err.Error())), nil
        }

        return mcpgo.NewToolResultJSON(resource)
    }

    return mcpgo.NewTool(
        "create_resource",
        "Create a new resource in Razorpay",
        parameters,
        handler,
    )
}
```

### Registering Tools

Add your tool to the appropriate toolset in the `NewToolSets` function in [`pkg/razorpay/tools.go`](tools.go):

```go
// NewToolSets creates and configures all available toolsets
func NewToolSets(
    log *slog.Logger,
    client *rzpsdk.Client,
    enabledToolsets []string,
    readOnly bool,
) (*toolsets.ToolsetGroup, error) {
    // Create a new toolset group
    toolsetGroup := toolsets.NewToolsetGroup(readOnly)

    // Create toolsets
    payments := toolsets.NewToolset("payments", "Razorpay Payments related tools").
        AddReadTools(
            FetchPayment(log, client),
            // Add your read-only payment tool here
        ).
        AddWriteTools(
            // Add your write payment tool here
        )

    paymentLinks := toolsets.NewToolset(
        "payment_links",
        "Razorpay Payment Links related tools").
        AddReadTools(
            FetchPaymentLink(log, client),
            // Add your read-only payment link tool here
        ).
        AddWriteTools(
            CreatePaymentLink(log, client),
            // Add your write payment link tool here
        )

    orders := toolsets.NewToolset("orders", "Razorpay Orders related tools").
        AddReadTools(
            FetchOrder(log, client),
            // Add your read-only order tool here
        ).
        AddWriteTools(
            CreateOrder(log, client),
            // Add your write order tool here
        )

    // If adding a new resource type, create a new toolset:
    /*
    newResource := toolsets.NewToolset("new_resource", "Razorpay New Resource related tools").
        AddReadTools(
            FetchNewResource(log, client),
        ).
        AddWriteTools(
            CreateNewResource(log, client),
        )
    toolsetGroup.AddToolset(newResource)
    */

    // Add toolsets to the group
    toolsetGroup.AddToolset(payments)
    toolsetGroup.AddToolset(paymentLinks)
    toolsetGroup.AddToolset(orders)

    return toolsetGroup, nil
}
```

Tools are organized into toolsets by resource type, and each toolset has separate collections for read-only tools (`AddReadTools`) and write tools (`AddWriteTools`). This allows the server to enable/disable write operations when in read-only mode.

### Writing Unit Tests

All new tools should have unit tests to verify their behavior. We use a standard pattern for testing tools:

```go
func Test_ToolName(t *testing.T) {
    // Define API path that needs to be mocked
    apiPathFmt := fmt.Sprintf(
        "/%s%s/%%s",
		constants.VERSION_V1,
        constants.PAYMENT_URL,
    )
    
    // Define mock responses
    successResponse := map[string]interface{}{
        "id": "resource_123",
        "amount": float64(1000),
        "currency": "INR",
        // Other expected fields
    }
    
    // Define test cases
    tests := []RazorpayToolTestCase{
        {
            Name: "successful case with all parameters",
            Request: map[string]interface{}{
                "key1": "value1",
                "key2": float64(1000),
                // All parameters for a complete request
            },
            MockHttpClient: func() (*http.Client, *httptest.Server) {
                return mock.NewHTTPClient(
                    mock.Endpoint{
                        Path:     fmt.Sprintf(apiPathFmt, "path_params") // or just apiPath. DO NOT add query params here.
                        Method:   "POST", // or "GET" for fetch operations
                        Response: successResponse,
                    },
                )
            },
            ExpectError:    false,
            ExpectedResult: successResponse,
        },
        {
            Name: "missing required parameter",
            Request: map[string]interface{}{
                // Missing a required parameter
            },
            MockHttpClient: nil, // No HTTP client needed for validation errors
            ExpectError:    true,
            ExpectedErrMsg: "missing required parameter: param1",
        },
        {
            Name: "multiple validation errors",
            Request: map[string]interface{}{
                // Missing required parameters and/or including invalid types
                "optional_param": "invalid_type", // Wrong type for a parameter
            },
            MockHttpClient: nil, // No HTTP client needed for validation errors
            ExpectError:    true,
            ExpectedErrMsg: "Validation errors:\n- missing required parameter: param1\n- invalid parameter type: optional_param",
        },
        // Additional test cases for other scenarios
    }
    
    // Run the tests
    for _, tc := range tests {
        t.Run(tc.Name, func(t *testing.T) {
            runToolTest(t, tc, ToolFunction, "Resource Name")
        })
    }
}
```

#### Best Practices while writing UTs for a new Tool

1. **Test Coverage**: At minimum, include:
   - One positive test case with all parameters (required and optional)
   - One negative test case for each required parameter
   - Any edge cases specific to your tool

2. **Mock HTTP Responses**: Use the `mock.NewHTTPClient` function to create mock HTTP responses for Razorpay API calls.

3. **Validation Errors**: For parameter validation errors, you don't need to mock HTTP responses as these errors are caught before the API call.

4. **Test API Errors**: Include at least one test for API-level errors (like invalid currency, not found, etc.).

5. **Naming Convention**: Use `Test_FunctionName` format for test functions.

6. Use the resource URLs from [Razorpay Go sdk constants](https://github.com/razorpay/razorpay-go/blob/master/constants/url.go) to specify the apiPath to be mocked.

See [`payment_links_test.go`](payment_links_test.go) for a complete example of tool tests.

### Updating Documentation

After adding a new tool, Update the "Available Tools" section in the README.md in the root of the repository

### Best Practices

1. **Consistent Naming**: Use consistent naming patterns:
   - Fetch methods: `fetch_resource`
   - Create methods: `create_resource`
   - FetchAll methods: `fetch_all_resources`

2. **Error Handling**: Always provide clear error messages

3. **Validation**: Always validate required parameters and collect all validation errors before returning using fluent validator pattern.
   - Use the `NewValidator` to create a validator
   - Chain validation methods (`ValidateAndAddRequiredString`, etc.)
   - Return formatted errors with `HandleErrorsIfAny()`

4. **Documentation**: Describe all the parameters clearly for the LLMs to understand.

5. **Organization**: Add tools to the appropriate file based on resource type

6. **Testing**: Test your tool with different parameter combinations 