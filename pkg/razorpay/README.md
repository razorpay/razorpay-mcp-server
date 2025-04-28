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

Each tool follows this pattern:

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
        "tool_id",
        "A description of the tool. NOTE: Add any exceptions/rules if relevant for the LLMs.",
        parameters,
        handler,
    )
}
```

### Parameter Definition

Define parameters using the mcpgo helpers:

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

Use the helper functions for parameter validation:

```go
// Required parameters
id, err := RequiredParam[string](r, "id")
if result, err := HandleValidationError(err); result != nil {
    return result, err
}

// Optional parameters
description, err := OptionalParam[string](r, "description")
if result, err := HandleValidationError(err); result != nil {
    return result, err
}

// Required integers
amount, err := RequiredInt(r, "amount")
if result, err := HandleValidationError(err); result != nil {
    return result, err
}

// Optional integers
limit, err := OptionalInt(r, "limit")
if result, err := HandleValidationError(err); result != nil {
    return result, err
}
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
        id, err := RequiredParam[string](r, "id")
        if result, err := HandleValidationError(err); result != nil {
            return result, err
        }

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
        // Required parameters
        amount, err := RequiredInt(r, "amount")
        if result, err := HandleValidationError(err); result != nil {
            return result, err
        }
        
        currency, err := RequiredParam[string](r, "currency")
        if result, err := HandleValidationError(err); result != nil {
            return result, err
        }

        // Create request payload
        data := map[string]interface{}{
            "amount": amount,
            "currency": currency,
        }

        // Optional parameters
        description, err := OptionalParam[string](r, "description")
        if result, err := HandleValidationError(err); result != nil {
            return result, err
        }
        
        if description != "" {
            data["description"] = description
        }

        // Call the API
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

Add your tool to the appropriate section in `server.go`:

```go
// RegisterTools adds all available tools to the server
func (s *Server) RegisterTools() {
    // payments tools
    s.server.AddTools(
        FetchPayment(s.log, s.client),
        // Add your payment tool here
    )
    
    // orders tools
    s.server.AddTools(
        CreateOrder(s.log, s.client),
        FetchOrder(s.log, s.client),
        // Add your order tool here
    )
    
    // payment links tools
    s.server.AddTools(
        CreatePaymentLink(s.log, s.client),
        FetchPaymentLink(s.log, s.client),
        // Add your payment link tool here
    )
    
    // Add a new section for your resource if needed
}
```

### Updating Documentation

After adding a new tool, Update the "Available Tools" section in the README.md in the root of the repository

### Best Practices

1. **Consistent Naming**: Use consistent naming patterns:
   - Fetch methods: `fetch_resource`
   - Create methods: `create_resource`
   - FetchAll methods: `fetch_all_resources`

2. **Error Handling**: Always provide clear error messages

3. **Validation**: Always validate required parameters

4. **Documentation**: Describe all the parameters clearly for the LLMs to understand.

5. **Organization**: Add tools to the appropriate file based on resource type

6. **Testing**: Test your tool with different parameter combinations 