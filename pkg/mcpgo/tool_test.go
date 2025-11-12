package mcpgo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestNewTool(t *testing.T) {
	t.Run("creates tool with all fields", func(t *testing.T) {
		handler := func(
			ctx context.Context, req CallToolRequest) (*ToolResult, error) {
			return NewToolResultText("success"), nil
		}
		tool := NewTool(
			"test-tool",
			"Test description",
			[]ToolParameter{WithString("param1")},
			handler,
		)
		assert.NotNil(t, tool)
		assert.NotNil(t, tool.GetHandler())
	})

	t.Run("creates tool with empty parameters", func(t *testing.T) {
		handler := func(
			ctx context.Context, req CallToolRequest) (*ToolResult, error) {
			return NewToolResultText("success"), nil
		}
		tool := NewTool("test-tool", "Test", []ToolParameter{}, handler)
		assert.NotNil(t, tool)
	})
}

func TestMark3labsToolImpl_GetHandler(t *testing.T) {
	t.Run("returns handler", func(t *testing.T) {
		handler := func(
			ctx context.Context, req CallToolRequest) (*ToolResult, error) {
			return NewToolResultText("success"), nil
		}
		tool := NewTool("test-tool", "Test", []ToolParameter{}, handler)
		assert.NotNil(t, tool.GetHandler())
	})
}

func TestMark3labsToolImpl_ToMCPServerTool(t *testing.T) {
	t.Run("converts tool with string parameter", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{WithString("param1")},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Tool)
		assert.NotNil(t, mcpTool.Handler)
	})

	t.Run("converts tool with number parameter", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{WithNumber("param1")},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Tool)
	})

	t.Run("converts tool with boolean parameter", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{WithBoolean("param1")},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Tool)
	})

	t.Run("converts tool with object parameter", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{WithObject("param1")},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Tool)
	})

	t.Run("converts tool with array parameter", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{WithArray("param1")},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Tool)
	})

	t.Run("converts tool with integer parameter", func(t *testing.T) {
		param := ToolParameter{
			Name:   "param1",
			Schema: map[string]interface{}{"type": "integer"},
		}
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{param},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Tool)
	})

	t.Run("converts tool with unknown type parameter", func(t *testing.T) {
		param := ToolParameter{
			Name:   "param1",
			Schema: map[string]interface{}{"type": "unknown"},
		}
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{param},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Tool)
	})

	t.Run("converts tool with missing type parameter", func(t *testing.T) {
		param := ToolParameter{
			Name:   "param1",
			Schema: map[string]interface{}{},
		}
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{param},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Tool)
	})

	t.Run("converts tool with non-string type", func(t *testing.T) {
		param := ToolParameter{
			Name:   "param1",
			Schema: map[string]interface{}{"type": 123},
		}
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{param},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Tool)
	})

	t.Run("handler returns error result", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultError("error occurred"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Handler)

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "test-tool",
				Arguments: map[string]interface{}{},
			},
		}
		result, err := mcpTool.Handler(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("handler returns handler error", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return nil, assert.AnError
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Handler)

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "test-tool",
				Arguments: map[string]interface{}{},
			},
		}
		result, err := mcpTool.Handler(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("handler returns error result", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultError("test error"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Handler)

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "test-tool",
				Arguments: map[string]interface{}{},
			},
		}
		result, err := mcpTool.Handler(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("handles parameter with missing type", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{
				{
					Name:   "param1",
					Schema: map[string]interface{}{}, // No type specified
				},
			},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Handler)
	})

	t.Run("handles parameter with non-string type", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{
				{
					Name: "param1",
					Schema: map[string]interface{}{
						"type": 123, // Non-string type
					},
				},
			},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Handler)
	})

	t.Run("handles unknown parameter type", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{
				{
					Name: "param1",
					Schema: map[string]interface{}{
						"type": "unknown-type",
					},
				},
			},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Handler)
	})

	t.Run("handles all parameter types", func(t *testing.T) {
		tool := NewTool(
			"test-tool",
			"Test",
			[]ToolParameter{
				{
					Name: "string_param",
					Schema: map[string]interface{}{
						"type": "string",
					},
				},
				{
					Name: "number_param",
					Schema: map[string]interface{}{
						"type": "number",
					},
				},
				{
					Name: "integer_param",
					Schema: map[string]interface{}{
						"type": "integer",
					},
				},
				{
					Name: "boolean_param",
					Schema: map[string]interface{}{
						"type": "boolean",
					},
				},
				{
					Name: "object_param",
					Schema: map[string]interface{}{
						"type": "object",
					},
				},
				{
					Name: "array_param",
					Schema: map[string]interface{}{
						"type": "array",
					},
				},
			},
			func(ctx context.Context, req CallToolRequest) (*ToolResult, error) {
				return NewToolResultText("success"), nil
			},
		)
		mcpTool := tool.toMCPServerTool()
		assert.NotNil(t, mcpTool.Handler)
	})
}

func TestPropertyOption_Min(t *testing.T) {
	t.Run("sets minimum for number", func(t *testing.T) {
		schema := map[string]interface{}{"type": "number"}
		Min(10.0)(schema)
		assert.Equal(t, 10.0, schema["minimum"])
	})

	t.Run("sets minimum for integer", func(t *testing.T) {
		schema := map[string]interface{}{"type": "integer"}
		Min(5.0)(schema)
		assert.Equal(t, 5.0, schema["minimum"])
	})

	t.Run("sets minLength for string", func(t *testing.T) {
		schema := map[string]interface{}{"type": "string"}
		Min(3.0)(schema)
		assert.Equal(t, 3, schema["minLength"])
	})

	t.Run("sets minItems for array", func(t *testing.T) {
		schema := map[string]interface{}{"type": "array"}
		Min(2.0)(schema)
		assert.Equal(t, 2, schema["minItems"])
	})

	t.Run("ignores for unknown type", func(t *testing.T) {
		schema := map[string]interface{}{"type": "boolean"}
		Min(1.0)(schema)
		assert.NotContains(t, schema, "minimum")
		assert.NotContains(t, schema, "minLength")
		assert.NotContains(t, schema, "minItems")
	})

	t.Run("ignores for missing type", func(t *testing.T) {
		schema := map[string]interface{}{}
		Min(1.0)(schema)
		assert.NotContains(t, schema, "minimum")
	})

	t.Run("ignores for non-string type", func(t *testing.T) {
		schema := map[string]interface{}{"type": 123}
		Min(1.0)(schema)
		assert.NotContains(t, schema, "minimum")
	})
}

func TestPropertyOption_Max(t *testing.T) {
	t.Run("sets maximum for number", func(t *testing.T) {
		schema := map[string]interface{}{"type": "number"}
		Max(100.0)(schema)
		assert.Equal(t, 100.0, schema["maximum"])
	})

	t.Run("sets maximum for integer", func(t *testing.T) {
		schema := map[string]interface{}{"type": "integer"}
		Max(50.0)(schema)
		assert.Equal(t, 50.0, schema["maximum"])
	})

	t.Run("sets maxLength for string", func(t *testing.T) {
		schema := map[string]interface{}{"type": "string"}
		Max(10.0)(schema)
		assert.Equal(t, 10, schema["maxLength"])
	})

	t.Run("sets maxItems for array", func(t *testing.T) {
		schema := map[string]interface{}{"type": "array"}
		Max(5.0)(schema)
		assert.Equal(t, 5, schema["maxItems"])
	})

	t.Run("ignores for unknown type", func(t *testing.T) {
		schema := map[string]interface{}{"type": "boolean"}
		Max(1.0)(schema)
		assert.NotContains(t, schema, "maximum")
	})

	t.Run("ignores for missing type", func(t *testing.T) {
		schema := map[string]interface{}{}
		Max(1.0)(schema)
		assert.NotContains(t, schema, "maximum")
	})

	t.Run("ignores for non-string type value", func(t *testing.T) {
		schema := map[string]interface{}{"type": 123}
		Max(1.0)(schema)
		assert.NotContains(t, schema, "maximum")
	})
}

func TestPropertyOption_Pattern(t *testing.T) {
	t.Run("sets pattern for string", func(t *testing.T) {
		schema := map[string]interface{}{"type": "string"}
		Pattern("^[a-z]+$")(schema)
		assert.Equal(t, "^[a-z]+$", schema["pattern"])
	})

	t.Run("ignores for non-string type", func(t *testing.T) {
		schema := map[string]interface{}{"type": "number"}
		Pattern("^[a-z]+$")(schema)
		assert.NotContains(t, schema, "pattern")
	})

	t.Run("ignores for missing type", func(t *testing.T) {
		schema := map[string]interface{}{}
		Pattern("^[a-z]+$")(schema)
		assert.NotContains(t, schema, "pattern")
	})

	t.Run("ignores for non-string type value", func(t *testing.T) {
		schema := map[string]interface{}{"type": 123}
		Pattern("^[a-z]+$")(schema)
		assert.NotContains(t, schema, "pattern")
	})
}

func TestPropertyOption_Enum(t *testing.T) {
	t.Run("sets enum values", func(t *testing.T) {
		schema := map[string]interface{}{}
		Enum("value1", "value2", "value3")(schema)
		assert.Equal(t, []interface{}{"value1", "value2", "value3"}, schema["enum"])
	})

	t.Run("sets enum with mixed types", func(t *testing.T) {
		schema := map[string]interface{}{}
		Enum("value1", 123, true)(schema)
		assert.Equal(t, []interface{}{"value1", 123, true}, schema["enum"])
	})
}

func TestPropertyOption_DefaultValue(t *testing.T) {
	t.Run("sets default string value", func(t *testing.T) {
		schema := map[string]interface{}{}
		DefaultValue("default")(schema)
		assert.Equal(t, "default", schema["default"])
	})

	t.Run("sets default number value", func(t *testing.T) {
		schema := map[string]interface{}{}
		DefaultValue(42.0)(schema)
		assert.Equal(t, 42.0, schema["default"])
	})

	t.Run("sets default boolean value", func(t *testing.T) {
		schema := map[string]interface{}{}
		DefaultValue(true)(schema)
		assert.Equal(t, true, schema["default"])
	})
}

func TestPropertyOption_MaxProperties(t *testing.T) {
	t.Run("sets maxProperties for object", func(t *testing.T) {
		schema := map[string]interface{}{"type": "object"}
		MaxProperties(5)(schema)
		assert.Equal(t, 5, schema["maxProperties"])
	})

	t.Run("ignores for non-object type", func(t *testing.T) {
		schema := map[string]interface{}{"type": "string"}
		MaxProperties(5)(schema)
		assert.NotContains(t, schema, "maxProperties")
	})

	t.Run("ignores for missing type", func(t *testing.T) {
		schema := map[string]interface{}{}
		MaxProperties(5)(schema)
		assert.NotContains(t, schema, "maxProperties")
	})
}

func TestPropertyOption_MinProperties(t *testing.T) {
	t.Run("sets minProperties for object", func(t *testing.T) {
		schema := map[string]interface{}{"type": "object"}
		MinProperties(2)(schema)
		assert.Equal(t, 2, schema["minProperties"])
	})

	t.Run("ignores for non-object type", func(t *testing.T) {
		schema := map[string]interface{}{"type": "string"}
		MinProperties(2)(schema)
		assert.NotContains(t, schema, "minProperties")
	})
}

func TestPropertyOption_Required(t *testing.T) {
	t.Run("sets required flag", func(t *testing.T) {
		schema := map[string]interface{}{}
		Required()(schema)
		assert.Equal(t, true, schema["required"])
	})
}

func TestPropertyOption_Description(t *testing.T) {
	t.Run("sets description", func(t *testing.T) {
		schema := map[string]interface{}{}
		Description("Test description")(schema)
		assert.Equal(t, "Test description", schema["description"])
	})
}

func TestToolParameter_ApplyPropertyOptions(t *testing.T) {
	t.Run("applies single option", func(t *testing.T) {
		param := ToolParameter{
			Name:   "test",
			Schema: map[string]interface{}{"type": "string"},
		}
		param.applyPropertyOptions(Description("Test desc"))
		assert.Equal(t, "Test desc", param.Schema["description"])
	})

	t.Run("applies multiple options", func(t *testing.T) {
		param := ToolParameter{
			Name:   "test",
			Schema: map[string]interface{}{"type": "string"},
		}
		param.applyPropertyOptions(
			Description("Test desc"),
			Required(),
			Min(3.0),
		)
		assert.Equal(t, "Test desc", param.Schema["description"])
		assert.Equal(t, true, param.Schema["required"])
		assert.Equal(t, 3, param.Schema["minLength"])
	})

	t.Run("applies no options", func(t *testing.T) {
		param := ToolParameter{
			Name:   "test",
			Schema: map[string]interface{}{"type": "string"},
		}
		param.applyPropertyOptions()
		assert.Equal(t, "string", param.Schema["type"])
	})
}

func TestWithString(t *testing.T) {
	t.Run("creates string parameter without options", func(t *testing.T) {
		param := WithString("test")
		assert.Equal(t, "test", param.Name)
		assert.Equal(t, "string", param.Schema["type"])
	})

	t.Run("creates string parameter with options", func(t *testing.T) {
		param := WithString("test", Description("Test"), Required(), Min(3.0))
		assert.Equal(t, "test", param.Name)
		assert.Equal(t, "string", param.Schema["type"])
		assert.Equal(t, "Test", param.Schema["description"])
		assert.Equal(t, true, param.Schema["required"])
		assert.Equal(t, 3, param.Schema["minLength"])
	})
}

func TestWithNumber(t *testing.T) {
	t.Run("creates number parameter without options", func(t *testing.T) {
		param := WithNumber("test")
		assert.Equal(t, "test", param.Name)
		assert.Equal(t, "number", param.Schema["type"])
	})

	t.Run("creates number parameter with options", func(t *testing.T) {
		param := WithNumber("test", Min(1.0), Max(100.0))
		assert.Equal(t, "test", param.Name)
		assert.Equal(t, "number", param.Schema["type"])
		assert.Equal(t, 1.0, param.Schema["minimum"])
		assert.Equal(t, 100.0, param.Schema["maximum"])
	})
}

func TestWithBoolean(t *testing.T) {
	t.Run("creates boolean parameter", func(t *testing.T) {
		param := WithBoolean("test")
		assert.Equal(t, "test", param.Name)
		assert.Equal(t, "boolean", param.Schema["type"])
	})
}

func TestWithObject(t *testing.T) {
	t.Run("creates object parameter", func(t *testing.T) {
		param := WithObject("test")
		assert.Equal(t, "test", param.Name)
		assert.Equal(t, "object", param.Schema["type"])
	})

	t.Run("creates object parameter with options", func(t *testing.T) {
		param := WithObject("test", MinProperties(1), MaxProperties(5))
		assert.Equal(t, "test", param.Name)
		assert.Equal(t, "object", param.Schema["type"])
		assert.Equal(t, 1, param.Schema["minProperties"])
		assert.Equal(t, 5, param.Schema["maxProperties"])
	})
}

func TestWithArray(t *testing.T) {
	t.Run("creates array parameter", func(t *testing.T) {
		param := WithArray("test")
		assert.Equal(t, "test", param.Name)
		assert.Equal(t, "array", param.Schema["type"])
	})

	t.Run("creates array parameter with options", func(t *testing.T) {
		param := WithArray("test", Min(1.0), Max(10.0))
		assert.Equal(t, "test", param.Name)
		assert.Equal(t, "array", param.Schema["type"])
		assert.Equal(t, 1, param.Schema["minItems"])
		assert.Equal(t, 10, param.Schema["maxItems"])
	})
}

func TestAddNumberPropertyOptions(t *testing.T) {
	t.Run("adds minimum", func(t *testing.T) {
		schema := map[string]interface{}{"minimum": 10.0}
		opts := addNumberPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})

	t.Run("adds maximum", func(t *testing.T) {
		schema := map[string]interface{}{"maximum": 100.0}
		opts := addNumberPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})

	t.Run("adds both minimum and maximum", func(t *testing.T) {
		schema := map[string]interface{}{
			"minimum": 10.0,
			"maximum": 100.0,
		}
		opts := addNumberPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})

	t.Run("handles non-float64 minimum", func(t *testing.T) {
		schema := map[string]interface{}{"minimum": "not-a-number"}
		opts := addNumberPropertyOptions(nil, schema)
		assert.Nil(t, opts)
	})
}

func TestAddStringPropertyOptions(t *testing.T) {
	t.Run("adds minLength", func(t *testing.T) {
		schema := map[string]interface{}{"minLength": 3}
		opts := addStringPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})

	t.Run("adds maxLength", func(t *testing.T) {
		schema := map[string]interface{}{"maxLength": 10}
		opts := addStringPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})

	t.Run("adds pattern", func(t *testing.T) {
		schema := map[string]interface{}{"pattern": "^[a-z]+$"}
		opts := addStringPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})

	t.Run("adds all string options", func(t *testing.T) {
		schema := map[string]interface{}{
			"minLength": 3,
			"maxLength": 10,
			"pattern":   "^[a-z]+$",
		}
		opts := addStringPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})
}

func TestAddDefaultValueOptions(t *testing.T) {
	t.Run("adds string default", func(t *testing.T) {
		opts := addDefaultValueOptions(nil, "default")
		assert.NotNil(t, opts)
	})

	t.Run("adds float64 default", func(t *testing.T) {
		opts := addDefaultValueOptions(nil, 42.0)
		assert.NotNil(t, opts)
	})

	t.Run("adds bool default", func(t *testing.T) {
		opts := addDefaultValueOptions(nil, true)
		assert.NotNil(t, opts)
	})

	t.Run("ignores unknown type", func(t *testing.T) {
		opts := addDefaultValueOptions(nil, []string{"test"})
		assert.Nil(t, opts)
	})
}

func TestAddEnumOptions(t *testing.T) {
	t.Run("adds enum with string values", func(t *testing.T) {
		enumValues := []interface{}{"value1", "value2", "value3"}
		opts := addEnumOptions(nil, enumValues)
		assert.NotNil(t, opts)
	})

	t.Run("adds enum with mixed values", func(t *testing.T) {
		enumValues := []interface{}{"value1", 123, "value2"}
		opts := addEnumOptions(nil, enumValues)
		assert.NotNil(t, opts)
	})

	t.Run("handles non-array enum", func(t *testing.T) {
		opts := addEnumOptions(nil, "not-an-array")
		assert.Nil(t, opts)
	})

	t.Run("handles empty enum array", func(t *testing.T) {
		enumValues := []interface{}{123, 456} // Non-string values
		opts := addEnumOptions(nil, enumValues)
		assert.Nil(t, opts) // Should return nil since no string values
	})
}

func TestAddObjectPropertyOptions(t *testing.T) {
	t.Run("adds maxProperties", func(t *testing.T) {
		schema := map[string]interface{}{"maxProperties": 5}
		opts := addObjectPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})

	t.Run("adds minProperties", func(t *testing.T) {
		schema := map[string]interface{}{"minProperties": 2}
		opts := addObjectPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})

	t.Run("adds both properties", func(t *testing.T) {
		schema := map[string]interface{}{
			"minProperties": 1,
			"maxProperties": 5,
		}
		opts := addObjectPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})
}

func TestAddArrayPropertyOptions(t *testing.T) {
	t.Run("adds minItems", func(t *testing.T) {
		schema := map[string]interface{}{"minItems": 1}
		opts := addArrayPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})

	t.Run("adds maxItems", func(t *testing.T) {
		schema := map[string]interface{}{"maxItems": 10}
		opts := addArrayPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})

	t.Run("adds both items", func(t *testing.T) {
		schema := map[string]interface{}{
			"minItems": 1,
			"maxItems": 10,
		}
		opts := addArrayPropertyOptions(nil, schema)
		assert.NotNil(t, opts)
	})
}

func TestConvertSchemaToPropertyOptions(t *testing.T) {
	t.Run("converts complete schema", func(t *testing.T) {
		schema := map[string]interface{}{
			"type":        "string",
			"description": "Test param",
			"required":    true,
			"minLength":   3,
			"maxLength":   10,
			"pattern":     "^[a-z]+$",
			"default":     "default",
		}
		opts := convertSchemaToPropertyOptions(schema)
		assert.NotNil(t, opts)
	})

	t.Run("converts number schema", func(t *testing.T) {
		schema := map[string]interface{}{
			"type":    "number",
			"minimum": 1.0,
			"maximum": 100.0,
			"default": 42.0,
		}
		opts := convertSchemaToPropertyOptions(schema)
		assert.NotNil(t, opts)
	})

	t.Run("converts object schema", func(t *testing.T) {
		schema := map[string]interface{}{
			"type":          "object",
			"minProperties": 1,
			"maxProperties": 5,
		}
		opts := convertSchemaToPropertyOptions(schema)
		assert.NotNil(t, opts)
	})

	t.Run("converts array schema", func(t *testing.T) {
		schema := map[string]interface{}{
			"type":     "array",
			"minItems": 1,
			"maxItems": 10,
		}
		opts := convertSchemaToPropertyOptions(schema)
		assert.NotNil(t, opts)
	})

	t.Run("converts schema with enum", func(t *testing.T) {
		schema := map[string]interface{}{
			"type": "string",
			"enum": []interface{}{"value1", "value2"},
		}
		opts := convertSchemaToPropertyOptions(schema)
		assert.NotNil(t, opts)
	})

	t.Run("handles empty description", func(t *testing.T) {
		schema := map[string]interface{}{
			"type":        "string",
			"description": "",
		}
		opts := convertSchemaToPropertyOptions(schema)
		// Empty description should not be added
		// In Go, a nil slice is valid and has length 0
		assert.Len(t, opts, 0)
	})

	t.Run("handles false required", func(t *testing.T) {
		schema := map[string]interface{}{
			"type":     "string",
			"required": false,
		}
		opts := convertSchemaToPropertyOptions(schema)
		// False required should not be added
		// In Go, a nil slice is valid and has length 0
		assert.Len(t, opts, 0)
	})
}

func TestNewToolResultJSON(t *testing.T) {
	t.Run("creates JSON result from map", func(t *testing.T) {
		data := map[string]interface{}{
			"key": "value",
			"num": 42,
		}
		result, err := NewToolResultJSON(data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.NotEmpty(t, result.Text)

		// Verify it's valid JSON
		var decoded map[string]interface{}
		err = json.Unmarshal([]byte(result.Text), &decoded)
		assert.NoError(t, err)
		assert.Equal(t, "value", decoded["key"])
	})

	t.Run("creates JSON result from struct", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		data := TestStruct{Name: "Test", Age: 30}
		result, err := NewToolResultJSON(data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("creates JSON result from array", func(t *testing.T) {
		data := []string{"item1", "item2"}
		result, err := NewToolResultJSON(data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("handles unmarshalable data", func(t *testing.T) {
		// Create a channel which cannot be marshaled to JSON
		data := make(chan int)
		result, err := NewToolResultJSON(data)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestNewToolResultText(t *testing.T) {
	t.Run("creates text result", func(t *testing.T) {
		result := NewToolResultText("test text")
		assert.NotNil(t, result)
		assert.Equal(t, "test text", result.Text)
		assert.False(t, result.IsError)
		assert.Nil(t, result.Content)
	})

	t.Run("creates empty text result", func(t *testing.T) {
		result := NewToolResultText("")
		assert.NotNil(t, result)
		assert.Equal(t, "", result.Text)
		assert.False(t, result.IsError)
	})
}

func TestNewToolResultError(t *testing.T) {
	t.Run("creates error result", func(t *testing.T) {
		result := NewToolResultError("error message")
		assert.NotNil(t, result)
		assert.Equal(t, "error message", result.Text)
		assert.True(t, result.IsError)
		assert.Nil(t, result.Content)
	})

	t.Run("creates empty error result", func(t *testing.T) {
		result := NewToolResultError("")
		assert.NotNil(t, result)
		assert.Equal(t, "", result.Text)
		assert.True(t, result.IsError)
	})
}
