# MCPGO Package

The `mcpgo` package provides an abstraction layer over the `github.com/mark3labs/mcp-go` library. Its purpose is to isolate this external dependency from the rest of the application by wrapping all necessary functionality within clean interfaces.

## Purpose

This package was created to isolate the `mark3labs/mcp-go` dependency for several key reasons:

1. **Dependency Isolation**: Confine all `mark3labs/mcp-go` imports to this package, ensuring the rest of the application does not directly depend on this external library.

2. **Official MCP GO SDK and Future Compatibility**: Prepare for the eventual release of an official MCP SDK by creating a clean abstraction layer that can be updated to use the official SDK when it becomes available. The official SDK is currently under development (see [Official MCP Go SDK discussion](https://github.com/orgs/modelcontextprotocol/discussions/224#discussioncomment-12927030)).

3. **Simplified API**: Provide a more focused, application-specific API that only exposes the functionality needed by our application.

4. **Error Handling**: Implement proper error handling patterns rather than relying on panics, making the application more robust.

## Components

The package contains several core components:

- **Server**: An interface representing an MCP server, with the `mark3labsImpl` providing the current implementation.
- **Tool**: Interface for defining MCP tools that can be registered with the server.
- **TransportServer**: Interface for different transport mechanisms (stdio, TCP).
- **ToolResult/ToolParameter**: Structures for handling tool calls and results.

## Parameter Helper Functions

The package provides convenience functions for creating tool parameters:

- `WithString(name, description string, required bool)`: Creates a string parameter
- `WithNumber(name, description string, required bool)`: Creates a number parameter
- `WithBoolean(name, description string, required bool)`: Creates a boolean parameter
- `WithObject(name, description string, required bool)`: Creates an object parameter
- `WithArray(name, description string, required bool)`: Creates an array parameter

## Tool Result Helper Functions

The package also provides functions for creating tool results:

- `NewToolResultText(text string)`: Creates a text result
- `NewToolResultJSON(data interface{})`: Creates a JSON result
- `NewToolResultError(text string)`: Creates an error result

## Usage Example

```go
// Create a server
server := mcpgo.NewServer(
    "my-server",
    "1.0.0",
    mcpgo.WithLogging(),
    mcpgo.WithToolCapabilities(true),
)

// Create a tool
tool := mcpgo.NewTool(
    "my_tool",
    "Description of my tool",
    []mcpgo.ToolParameter{
        mcpgo.WithString(
            "param1",
            mcpgo.Description("Description of param1"),
            mcpgo.Required(),
        ),
    },
    func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
        // Extract parameter value
        param1Value, ok := req.Arguments["param1"]
        if !ok {
            return mcpgo.NewToolResultError("Missing required parameter: param1"), nil
        }
        
        // Process and return result
        return mcpgo.NewToolResultText("Result: " + param1Value.(string)), nil
    },
)

// Add tool to server
server.AddTools(tool)

// Create and run a stdio server
stdioServer, err := mcpgo.NewStdioServer(server)
if err != nil {
    log.Fatalf("Failed to create stdio server: %v", err)
}
err = stdioServer.Listen(context.Background(), os.Stdin, os.Stdout)
if err != nil {
    log.Fatalf("Server error: %v", err)
}
```

## Real-world Example

Here's how we use this package in the Razorpay MCP server to create a payment fetching tool:

```go
// FetchPayment returns a tool that fetches payment details using payment_id
func FetchPayment(
    log *slog.Logger,
    client *rzpsdk.Client,
) mcpgo.Tool {
    parameters := []mcpgo.ToolParameter{
        mcpgo.WithString(
            "payment_id",
            mcpgo.Description("payment_id is unique identifier of the payment to be retrieved."),
            mcpgo.Required(),
        ),
    }

    handler := func(
        ctx context.Context,
        r mcpgo.CallToolRequest,
    ) (*mcpgo.ToolResult, error) {
        arg, ok := r.Arguments["payment_id"]
        if !ok {
            return mcpgo.NewToolResultError(
                "payment id is a required field"), nil
        }
        id, ok := arg.(string)
        if !ok {
            return mcpgo.NewToolResultError(
                "payment id is expected to be a string"), nil
        }

        payment, err := client.Payment.Fetch(id, nil, nil)
        if err != nil {
            return mcpgo.NewToolResultError(
                fmt.Sprintf("fetching payment failed: %s", err.Error())), nil
        }

        return mcpgo.NewToolResultJSON(payment)
    }

    return mcpgo.NewTool(
        "fetch_payment",
        "fetch payment details using payment id.",
        parameters,
        handler,
    )
}
```

## Design Principles

1. **Minimal Interface Exposure**: The interfaces defined in this package include only methods that are actually used by our application.

2. **Proper Error Handling**: Functions return errors instead of panicking, allowing for graceful error handling throughout the application.

3. **Implementation Hiding**: The implementation details using `mark3labs/mcp-go` are hidden behind clean interfaces, making future transitions easier.

4. **Naming Clarity**: All implementation types are prefixed with `mark3labs` to clearly indicate they are specifically tied to the current library being used.

## Directory Structure

```
pkg/mcpgo/
├── server.go       # Server interface and implementation
├── transport.go    # TransportServer interface
├── stdio.go        # StdioServer implementation 
├── tool.go         # Tool interfaces and implementation
└── README.md       # This file
``` 