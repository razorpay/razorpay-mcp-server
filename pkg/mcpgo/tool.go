package mcpgo

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolHandler handles tool calls
type ToolHandler func(
	ctx context.Context,
	request CallToolRequest) (*ToolResult, error)

// CallToolRequest represents a request to call a tool
type CallToolRequest struct {
	Name      string
	Arguments map[string]interface{}
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	Text    string
	IsError bool
	Content []interface{}
}

// Tool represents a tool that can be added to the server
type Tool interface {
	// internal method to convert to mcp's ServerTool
	toMCPServerTool() server.ServerTool
}

// ToolParameter represents a parameter for a tool
type ToolParameter struct {
	Name        string
	Description string
	Required    bool
	Schema      map[string]interface{}
}

// WithString creates a string parameter
func WithString(name, description string, required bool) ToolParameter {
	return ToolParameter{
		Name:        name,
		Description: description,
		Required:    required,
		Schema:      map[string]interface{}{"type": "string"},
	}
}

// WithNumber creates a number parameter
func WithNumber(name, description string, required bool) ToolParameter {
	return ToolParameter{
		Name:        name,
		Description: description,
		Required:    required,
		Schema:      map[string]interface{}{"type": "number"},
	}
}

// WithBoolean creates a boolean parameter
func WithBoolean(name, description string, required bool) ToolParameter {
	return ToolParameter{
		Name:        name,
		Description: description,
		Required:    required,
		Schema:      map[string]interface{}{"type": "boolean"},
	}
}

// WithObject creates an object parameter
func WithObject(name, description string, required bool) ToolParameter {
	return ToolParameter{
		Name:        name,
		Description: description,
		Required:    required,
		Schema:      map[string]interface{}{"type": "object"},
	}
}

// WithArray creates an array parameter
func WithArray(name, description string, required bool) ToolParameter {
	return ToolParameter{
		Name:        name,
		Description: description,
		Required:    required,
		Schema:      map[string]interface{}{"type": "array"},
	}
}

// mark3labsToolImpl implements the Tool interface
type mark3labsToolImpl struct {
	name        string
	description string
	handler     ToolHandler
	parameters  []ToolParameter
}

// NewTool creates a new tool with the given
// name, description, parameters and handler
func NewTool(
	name,
	description string,
	parameters []ToolParameter,
	handler ToolHandler) *mark3labsToolImpl {
	return &mark3labsToolImpl{
		name:        name,
		description: description,
		handler:     handler,
		parameters:  parameters,
	}
}

// toMCPServerTool converts our Tool to mcp's ServerTool
func (t *mark3labsToolImpl) toMCPServerTool() server.ServerTool {
	// Create the mcp tool with appropriate options
	var toolOpts []mcp.ToolOption

	// Add description
	toolOpts = append(toolOpts, mcp.WithDescription(t.description))

	// Add parameters with their schemas
	for _, param := range t.parameters {
		propOpts := []mcp.PropertyOption{}

		// Add description to property
		if param.Description != "" {
			propOpts = append(propOpts, mcp.Description(param.Description))
		}

		// Add required flag
		if param.Required {
			propOpts = append(propOpts, mcp.Required())
		}

		// Get the type from the schema
		schemaType, ok := param.Schema["type"].(string)
		if !ok {
			// Default to string if type is missing or not a string
			schemaType = "string"
		}

		// Use the appropriate function based on schema type
		switch schemaType {
		case "string":
			toolOpts = append(toolOpts, mcp.WithString(param.Name, propOpts...))
		case "number", "integer":
			toolOpts = append(toolOpts, mcp.WithNumber(param.Name, propOpts...))
		case "boolean":
			toolOpts = append(toolOpts, mcp.WithBoolean(param.Name, propOpts...))
		case "object":
			toolOpts = append(toolOpts, mcp.WithObject(param.Name, propOpts...))
		case "array":
			toolOpts = append(toolOpts, mcp.WithArray(param.Name, propOpts...))
		default:
			// Unknown type, default to string
			toolOpts = append(toolOpts, mcp.WithString(param.Name, propOpts...))
		}
	}

	// Create the tool with all options
	tool := mcp.NewTool(t.name, toolOpts...)

	// Create the handler
	handlerFunc := func(
		ctx context.Context,
		req mcp.CallToolRequest,
	) (*mcp.CallToolResult, error) {
		// Convert mcp request to our request
		ourReq := CallToolRequest{
			Name:      req.Params.Name,
			Arguments: req.Params.Arguments,
		}

		// Call our handler
		result, err := t.handler(ctx, ourReq)
		if err != nil {
			return nil, err
		}

		// Convert our result to mcp result
		var mcpResult *mcp.CallToolResult
		if result.IsError {
			mcpResult = mcp.NewToolResultError(result.Text)
		} else {
			mcpResult = mcp.NewToolResultText(result.Text)
		}

		return mcpResult, nil
	}

	return server.ServerTool{
		Tool:    tool,
		Handler: handlerFunc,
	}
}

// NewToolResultJSON creates a new tool result with JSON content
func NewToolResultJSON(data interface{}) (*ToolResult, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &ToolResult{
		Text:    string(jsonBytes),
		IsError: false,
		Content: nil,
	}, nil
}

// NewToolResultText creates a new tool result with text content
func NewToolResultText(text string) *ToolResult {
	return &ToolResult{
		Text:    text,
		IsError: false,
		Content: nil,
	}
}

// NewToolResultError creates a new tool result with an error
func NewToolResultError(text string) *ToolResult {
	return &ToolResult{
		Text:    text,
		IsError: true,
		Content: nil,
	}
}
