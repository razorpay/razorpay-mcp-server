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
	Arguments any
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

	// GetHandler internal method for fetching the underlying handler
	GetHandler() ToolHandler
}

// PropertyOption represents a customization option for
// a parameter's schema
type PropertyOption func(schema map[string]interface{})

// Min sets the minimum value for a number parameter or
// minimum length for a string
func Min(value float64) PropertyOption {
	return func(schema map[string]interface{}) {
		propType, ok := schema["type"].(string)
		if !ok {
			return
		}

		switch propType {
		case "number", "integer":
			schema["minimum"] = value
		case "string":
			schema["minLength"] = int(value)
		case "array":
			schema["minItems"] = int(value)
		}
	}
}

// Max sets the maximum value for a number parameter or
// maximum length for a string
func Max(value float64) PropertyOption {
	return func(schema map[string]interface{}) {
		propType, ok := schema["type"].(string)
		if !ok {
			return
		}

		switch propType {
		case "number", "integer":
			schema["maximum"] = value
		case "string":
			schema["maxLength"] = int(value)
		case "array":
			schema["maxItems"] = int(value)
		}
	}
}

// Pattern sets a regex pattern for string validation
func Pattern(pattern string) PropertyOption {
	return func(schema map[string]interface{}) {
		propType, ok := schema["type"].(string)
		if !ok || propType != "string" {
			return
		}
		schema["pattern"] = pattern
	}
}

// Enum sets allowed values for a parameter
func Enum(values ...interface{}) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["enum"] = values
	}
}

// DefaultValue sets a default value for a parameter
func DefaultValue(value interface{}) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["default"] = value
	}
}

// MaxProperties sets the maximum number of properties for an object
func MaxProperties(max int) PropertyOption {
	return func(schema map[string]interface{}) {
		propType, ok := schema["type"].(string)
		if !ok || propType != "object" {
			return
		}
		schema["maxProperties"] = max
	}
}

// MinProperties sets the minimum number of properties for an object
func MinProperties(min int) PropertyOption {
	return func(schema map[string]interface{}) {
		propType, ok := schema["type"].(string)
		if !ok || propType != "object" {
			return
		}
		schema["minProperties"] = min
	}
}

// Required sets the tool parameter as required.
// When a parameter is marked as required, the client must provide a value
// for this parameter or the tool call will fail with an error.
func Required() PropertyOption {
	return func(schema map[string]interface{}) {
		schema["required"] = true
	}
}

// Description sets the description for the tool parameter.
// The description should explain the purpose of the parameter, expected format,
// and any relevant constraints.
func Description(desc string) PropertyOption {
	return func(schema map[string]interface{}) {
		schema["description"] = desc
	}
}

// ToolParameter represents a parameter for a tool
type ToolParameter struct {
	Name   string
	Schema map[string]interface{}
}

// applyPropertyOptions applies the given property options to
// the parameter schema
func (p *ToolParameter) applyPropertyOptions(opts ...PropertyOption) {
	for _, opt := range opts {
		opt(p.Schema)
	}
}

// WithString creates a string parameter with optional property options
func WithString(name string, opts ...PropertyOption) ToolParameter {
	param := ToolParameter{
		Name:   name,
		Schema: map[string]interface{}{"type": "string"},
	}
	param.applyPropertyOptions(opts...)
	return param
}

// WithNumber creates a number parameter with optional property options
func WithNumber(name string, opts ...PropertyOption) ToolParameter {
	param := ToolParameter{
		Name:   name,
		Schema: map[string]interface{}{"type": "number"},
	}
	param.applyPropertyOptions(opts...)
	return param
}

// WithBoolean creates a boolean parameter with optional property options
func WithBoolean(name string, opts ...PropertyOption) ToolParameter {
	param := ToolParameter{
		Name:   name,
		Schema: map[string]interface{}{"type": "boolean"},
	}
	param.applyPropertyOptions(opts...)
	return param
}

// WithObject creates an object parameter with optional property options
func WithObject(name string, opts ...PropertyOption) ToolParameter {
	param := ToolParameter{
		Name:   name,
		Schema: map[string]interface{}{"type": "object"},
	}
	param.applyPropertyOptions(opts...)
	return param
}

// WithArray creates an array parameter with optional property options
func WithArray(name string, opts ...PropertyOption) ToolParameter {
	param := ToolParameter{
		Name:   name,
		Schema: map[string]interface{}{"type": "array"},
	}
	param.applyPropertyOptions(opts...)
	return param
}

// mark3labsToolImpl implements the Tool interface
type mark3labsToolImpl struct {
	name        string
	description string
	handler     ToolHandler
	parameters  []ToolParameter
}

// NewTool creates a new tool with the given
// Name, description, parameters and handler
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

// addNumberPropertyOptions adds number-specific options to the property options
func addNumberPropertyOptions(
	propOpts []mcp.PropertyOption,
	schema map[string]interface{}) []mcp.PropertyOption {
	// Add minimum if present
	if min, ok := schema["minimum"].(float64); ok {
		propOpts = append(propOpts, mcp.Min(min))
	}

	// Add maximum if present
	if max, ok := schema["maximum"].(float64); ok {
		propOpts = append(propOpts, mcp.Max(max))
	}

	return propOpts
}

// addStringPropertyOptions adds string-specific options to the property options
func addStringPropertyOptions(
	propOpts []mcp.PropertyOption,
	schema map[string]interface{}) []mcp.PropertyOption {
	// Add minLength if present
	if minLength, ok := schema["minLength"].(int); ok {
		propOpts = append(propOpts, mcp.MinLength(minLength))
	}

	// Add maxLength if present
	if maxLength, ok := schema["maxLength"].(int); ok {
		propOpts = append(propOpts, mcp.MaxLength(maxLength))
	}

	// Add pattern if present
	if pattern, ok := schema["pattern"].(string); ok {
		propOpts = append(propOpts, mcp.Pattern(pattern))
	}

	return propOpts
}

// addDefaultValueOptions adds default value options based on type
func addDefaultValueOptions(
	propOpts []mcp.PropertyOption,
	defaultValue interface{}) []mcp.PropertyOption {
	switch val := defaultValue.(type) {
	case string:
		propOpts = append(propOpts, mcp.DefaultString(val))
	case float64:
		propOpts = append(propOpts, mcp.DefaultNumber(val))
	case bool:
		propOpts = append(propOpts, mcp.DefaultBool(val))
	}
	return propOpts
}

// addEnumOptions adds enum options if present
func addEnumOptions(
	propOpts []mcp.PropertyOption,
	enumValues interface{}) []mcp.PropertyOption {
	values, ok := enumValues.([]interface{})
	if !ok {
		return propOpts
	}

	// Convert values to strings for now
	strValues := make([]string, 0, len(values))
	for _, ev := range values {
		if str, ok := ev.(string); ok {
			strValues = append(strValues, str)
		}
	}

	if len(strValues) > 0 {
		propOpts = append(propOpts, mcp.Enum(strValues...))
	}

	return propOpts
}

// addObjectPropertyOptions adds object-specific options
func addObjectPropertyOptions(
	propOpts []mcp.PropertyOption,
	schema map[string]interface{}) []mcp.PropertyOption {
	// Add maxProperties if present
	if maxProps, ok := schema["maxProperties"].(int); ok {
		propOpts = append(propOpts, mcp.MaxProperties(maxProps))
	}

	// Add minProperties if present
	if minProps, ok := schema["minProperties"].(int); ok {
		propOpts = append(propOpts, mcp.MinProperties(minProps))
	}

	return propOpts
}

// addArrayPropertyOptions adds array-specific options
func addArrayPropertyOptions(
	propOpts []mcp.PropertyOption,
	schema map[string]interface{}) []mcp.PropertyOption {
	// Add minItems if present
	if minItems, ok := schema["minItems"].(int); ok {
		propOpts = append(propOpts, mcp.MinItems(minItems))
	}

	// Add maxItems if present
	if maxItems, ok := schema["maxItems"].(int); ok {
		propOpts = append(propOpts, mcp.MaxItems(maxItems))
	}

	return propOpts
}

// convertSchemaToPropertyOptions converts our schema to mcp property options
func convertSchemaToPropertyOptions(
	schema map[string]interface{}) []mcp.PropertyOption {
	var propOpts []mcp.PropertyOption

	// Add description if present
	if description, ok := schema["description"].(string); ok && description != "" {
		propOpts = append(propOpts, mcp.Description(description))
	}

	// Add required flag if present
	if required, ok := schema["required"].(bool); ok && required {
		propOpts = append(propOpts, mcp.Required())
	}

	// Skip type, description and required as they're handled separately
	for k, v := range schema {
		if k == "type" || k == "description" || k == "required" {
			continue
		}

		// Process property based on key
		switch k {
		case "minimum", "maximum":
			propOpts = addNumberPropertyOptions(propOpts, schema)
		case "minLength", "maxLength", "pattern":
			propOpts = addStringPropertyOptions(propOpts, schema)
		case "default":
			propOpts = addDefaultValueOptions(propOpts, v)
		case "enum":
			propOpts = addEnumOptions(propOpts, v)
		case "maxProperties", "minProperties":
			propOpts = addObjectPropertyOptions(propOpts, schema)
		case "minItems", "maxItems":
			propOpts = addArrayPropertyOptions(propOpts, schema)
		}
	}

	return propOpts
}

// GetHandler returns the handler for the tool
func (t *mark3labsToolImpl) GetHandler() ToolHandler {
	return t.handler
}

// toMCPServerTool converts our Tool to mcp's ServerTool
func (t *mark3labsToolImpl) toMCPServerTool() server.ServerTool {
	// Create the mcp tool with appropriate options
	var toolOpts []mcp.ToolOption

	// Add description
	toolOpts = append(toolOpts, mcp.WithDescription(t.description))

	// Add parameters with their schemas
	for _, param := range t.parameters {
		// Get property options from schema
		propOpts := convertSchemaToPropertyOptions(param.Schema)

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
