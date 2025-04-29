package mcpgo

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		version    string
		options    []ServerOption
		expectName string
		expectVer  string
	}{
		{
			name:       "Create server with no options",
			serverName: "test-server",
			version:    "1.0.0",
			options:    nil,
			expectName: "test-server",
			expectVer:  "1.0.0",
		},
		{
			name:       "Create server with all options",
			serverName: "test-server-all",
			version:    "1.0.4",
			options: []ServerOption{
				WithLogging(),
				WithResourceCapabilities(true, true),
				WithToolCapabilities(true),
			},
			expectName: "test-server-all",
			expectVer:  "1.0.4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := NewServer(tc.serverName, tc.version, tc.options...)

			assert.Equal(t, tc.expectName, srv.name)
			assert.Equal(t, tc.expectVer, srv.version)
			assert.NotNil(t, srv.mcpServer)
		})
	}
}

func TestToMCPServerTool(t *testing.T) {
	simpleHandler := func(
		ctx context.Context,
		req CallToolRequest,
	) (*ToolResult, error) {
		return NewToolResultText("test-result"), nil
	}

	tests := []struct {
		name       string
		tool       Tool
		expectName string
	}{
		{
			name: "Simple tool with no parameters",
			tool: NewTool(
				"simple-tool",
				"A simple test tool",
				[]ToolParameter{},
				simpleHandler,
			),
			expectName: "simple-tool",
		},
		{
			name: "Tool with string parameter",
			tool: NewTool(
				"string-param-tool",
				"A tool with a string parameter",
				[]ToolParameter{
					WithString("str_param", Description("A string parameter"), Required()),
				},
				simpleHandler,
			),
			expectName: "string-param-tool",
		},
		{
			name: "Tool with mixed parameters",
			tool: NewTool(
				"mixed-param-tool",
				"A tool with mixed parameters",
				[]ToolParameter{
					WithString("str_param", Description("A string parameter"), Required()),
					WithNumber(
						"num_param",
						Description("A number parameter"),
						Min(0),
						Max(100),
					),
					WithBoolean("bool_param", Description("A boolean parameter")),
					WithObject("obj_param", Description("An object parameter")),
					WithArray("arr_param", Description("An array parameter")),
				},
				simpleHandler,
			),
			expectName: "mixed-param-tool",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mcpTool := tc.tool.toMCPServerTool()

			assert.Equal(t, tc.expectName, mcpTool.Tool.Name)
			assert.NotNil(t, mcpTool.Handler)
		})
	}
}

func TestToMCPServerToolHandler(t *testing.T) {
	t.Run("SuccessHandler", func(t *testing.T) {
		testHandler := func(
			ctx context.Context,
			req CallToolRequest,
		) (*ToolResult, error) {
			assert.Equal(t, "test-tool", req.Name,
				"Expected tool name to be 'test-tool'")

			textParam, exists := req.Arguments["text"]
			assert.True(t, exists, "Expected 'text' parameter to exist")
			if !exists {
				return NewToolResultError("text parameter missing"), nil
			}

			textValue, ok := textParam.(string)
			assert.True(t, ok, "Expected text parameter to be string")
			if !ok {
				return NewToolResultError("text parameter not a string"), nil
			}

			assert.Equal(t, "hello world", textValue,
				"Expected text value to be 'hello world'")

			return NewToolResultText("success: " + textValue), nil
		}

		tool := NewTool(
			"test-tool",
			"Test tool",
			[]ToolParameter{
				WithString("text", Description("Text parameter"), Required()),
			},
			testHandler,
		)

		mcpTool := tool.toMCPServerTool()

		mcpReq := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name: "test-tool",
				Arguments: map[string]interface{}{
					"text": "hello world",
				},
				Meta: nil,
			},
		}

		result, err := mcpTool.Handler(context.Background(), mcpReq)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)

		textContent, ok := result.Content[0].(mcp.TextContent)
		assert.True(t, ok, "Expected TextContent")
		assert.Equal(t, "success: hello world", textContent.Text)
	})

	t.Run("ErrorHandler", func(t *testing.T) {
		testHandler := func(
			ctx context.Context,
			req CallToolRequest,
		) (*ToolResult, error) {
			return NewToolResultError("intentional error"), nil
		}

		tool := NewTool(
			"error-tool",
			"A tool that always returns an error",
			[]ToolParameter{},
			testHandler,
		)

		mcpTool := tool.toMCPServerTool()

		mcpReq := mcp.CallToolRequest{
			Request: mcp.Request{},
			Params: struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments,omitempty"`
				Meta      *struct {
					ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
				} `json:"_meta,omitempty"`
			}{
				Name:      "error-tool",
				Arguments: map[string]interface{}{},
				Meta:      nil,
			},
		}

		result, err := mcpTool.Handler(context.Background(), mcpReq)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)

		errorContent, ok := result.Content[0].(mcp.TextContent)
		assert.True(t, ok, "Expected TextContent")
		assert.Equal(t, "intentional error", errorContent.Text)
	})
}

func TestSchemaAdditionsForToolCreation(t *testing.T) {
	simpleHandler := func(
		ctx context.Context,
		req CallToolRequest,
	) (*ToolResult, error) {
		return NewToolResultText("test-result"), nil
	}

	comprehensiveTool := NewTool(
		"comprehensive-tool",
		"A comprehensive test tool with all schema additions",
		[]ToolParameter{
			WithString("str_param",
				Description("A string parameter"),
				Required(),
				Min(5),
				Max(100),
				Pattern("^[a-z]+$"),
				Enum("option1", "option2", "option3"),
			),
			WithNumber("num_param",
				Description("A number parameter"),
				Min(0),
				Max(1000),
				DefaultValue(50),
			),
			WithBoolean("bool_param",
				Description("A boolean parameter"),
				DefaultValue(true),
			),
			WithObject("obj_param",
				Description("An object parameter"),
				MinProperties(1),
				MaxProperties(10),
			),
			WithArray("arr_param",
				Description("An array parameter"),
				Min(1),
				Max(5),
			),
		},
		simpleHandler,
	)

	mcpTool := comprehensiveTool.toMCPServerTool()

	assert.Equal(t, "comprehensive-tool", mcpTool.Tool.Name)

	for _, param := range comprehensiveTool.parameters {
		switch param.Name {
		case "str_param":
			assert.Equal(t, "string", param.Schema["type"])
			assert.Equal(t, true, param.Schema["required"])
			assert.Equal(t, 5, param.Schema["minLength"])
		case "num_param":
			assert.Equal(t, "number", param.Schema["type"])
			assert.Equal(t, float64(0), param.Schema["minimum"])

			t.Logf(
				"Default value type: %T, value: %v",
				param.Schema["default"],
				param.Schema["default"],
			)

			switch v := param.Schema["default"].(type) {
			case int:
				assert.Equal(t, 50, v)
			case float64:
				assert.Equal(t, 50.0, v)
			default:
				t.Errorf(
					"Expected default value to be int or float64, got %T",
					param.Schema["default"],
				)
			}
		}
	}
}

func TestOptionSetterBehavior(t *testing.T) {
	setter := &mark3labsOptionSetter{
		mcpOptions: []server.ServerOption{},
	}

	validOption := server.WithLogging()
	err := setter.SetOption(validOption)
	assert.NoError(t, err)
	assert.Len(t, setter.mcpOptions, 1, "Expected 1 option to be added")

	invalidOption := "not an option"
	err = setter.SetOption(invalidOption)
	assert.NoError(t, err)
	assert.Len(t, setter.mcpOptions, 1, "Expected still 1 option")
}
