package mcpgo

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// Server defines the minimal MCP server interface needed by the application
type Server interface {
	// AddTools adds tools to the server
	AddTools(tools ...Tool)
}

// NewServer creates a new MCP server
func NewServer(name, version string, opts ...ServerOption) *Mark3labsImpl {
	// Create option setter to collect mcp options
	optSetter := &mark3labsOptionSetter{
		mcpOptions: []server.ServerOption{},
	}

	// Apply our options, which will populate the mcp options
	for _, opt := range opts {
		_ = opt(optSetter)
	}

	// Create the underlying mcp server
	mcpServer := server.NewMCPServer(
		name,
		version,
		optSetter.mcpOptions...,
	)

	return &Mark3labsImpl{
		McpServer: mcpServer,
		Name:      name,
		Version:   version,
	}
}

// Mark3labsImpl implements the Server interface using mark3labs/mcp-go
type Mark3labsImpl struct {
	McpServer *server.MCPServer
	Name      string
	Version   string
}

// mark3labsOptionSetter is used to apply options to the server
type mark3labsOptionSetter struct {
	mcpOptions []server.ServerOption
}

func (s *mark3labsOptionSetter) SetOption(option interface{}) error {
	if opt, ok := option.(server.ServerOption); ok {
		s.mcpOptions = append(s.mcpOptions, opt)
	}
	return nil
}

// AddTools adds tools to the server
func (s *Mark3labsImpl) AddTools(tools ...Tool) {
	// Convert our Tool to mcp's ServerTool
	var mcpTools []server.ServerTool
	for _, tool := range tools {
		mcpTools = append(mcpTools, tool.toMCPServerTool())
	}
	s.McpServer.AddTools(mcpTools...)
}

// OptionSetter is an interface for setting options on a configurable object
type OptionSetter interface {
	SetOption(option interface{}) error
}

// ServerOption is a function that configures a Server
type ServerOption func(OptionSetter) error

// WithLogging returns a server option that enables logging
func WithLogging() ServerOption {
	return func(s OptionSetter) error {
		return s.SetOption(server.WithLogging())
	}
}

func WithHooks(hooks *server.Hooks) ServerOption {
	return func(s OptionSetter) error {
		return s.SetOption(server.WithHooks(hooks))
	}
}

// WithResourceCapabilities returns a server option
// that enables resource capabilities
func WithResourceCapabilities(read, list bool) ServerOption {
	return func(s OptionSetter) error {
		return s.SetOption(server.WithResourceCapabilities(read, list))
	}
}

// WithToolCapabilities returns a server option that enables tool capabilities
func WithToolCapabilities(enabled bool) ServerOption {
	return func(s OptionSetter) error {
		return s.SetOption(server.WithToolCapabilities(enabled))
	}
}

// WithAuthenticationMiddleware returns a server option that adds an
// authentication middleware to the server.
func WithAuthenticationMiddleware(
	client *rzpsdk.Client,
) ServerOption {
	return func(s OptionSetter) error {
		return s.SetOption(server.WithToolHandlerMiddleware(
			func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
				return func(
					ctx context.Context,
					request mcp.CallToolRequest,
				) (result *mcp.CallToolResult, err error) {
					authenticatedCtx, err := AuthenticateRequest(ctx, client)
					if err != nil {
						return nil, err
					}
					return next(authenticatedCtx, request)
				}
			}),
		)
	}
}

// SetupHooks creates and configures the server hooks with logging
func SetupHooks(obs *observability.Observability) *server.Hooks {
	hooks := &server.Hooks{}
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod,
		message any) {
		obs.Logger.Infof(ctx, "MCP_METHOD_CALLED",
			"method", method,
			"id", id,
			"message", message)
	})

	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod,
		message any, result any) {
		logResult := result
		if method == mcp.MethodToolsList {
			if r, ok := result.(*mcp.ListToolsResult); ok {
				simplifiedTools := make([]string, 0, len(r.Tools))
				for _, tool := range r.Tools {
					simplifiedTools = append(simplifiedTools, tool.Name)
				}
				// Create new map for logging with just the tool names
				logResult = map[string]interface{}{
					"tools": simplifiedTools,
				}
			}
		}

		obs.Logger.Infof(ctx, "MCP_METHOD_SUCCEEDED",
			"method", method,
			"id", id,
			"result", logResult)
	})

	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod,
		message any, err error) {
		obs.Logger.Infof(ctx, "MCP_METHOD_FAILED",
			"method", method,
			"id", id,
			"message", message,
			"error", err)
	})

	hooks.AddBeforeCallTool(func(ctx context.Context, id any,
		message *mcp.CallToolRequest) {
		obs.Logger.Infof(ctx, "TOOL_CALL_STARTED",
			"id", id,
			"request", message)
	})

	hooks.AddAfterCallTool(func(ctx context.Context, id any,
		message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		obs.Logger.Infof(ctx, "TOOL_CALL_COMPLETED",
			"id", id,
			"request", message,
			"result", result)
	})

	return hooks
}
