package mcpgo

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	rzpsdk "github.com/razorpay/razorpay-go/v2"
)

// Server defines the minimal MCP server interface needed by the application
type Server interface {
	// AddTools adds tools to the server
	AddTools(tools ...Tool)
}

// NewServer creates a new MCP server
func NewServer(name, version string, opts ...ServerOption) *mark3labsImpl {
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

	return &mark3labsImpl{
		mcpServer: mcpServer,
		name:      name,
		version:   version,
	}
}

// mark3labsImpl implements the Server interface using mark3labs/mcp-go
type mark3labsImpl struct {
	mcpServer *server.MCPServer
	name      string
	version   string
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
func (s *mark3labsImpl) AddTools(tools ...Tool) {
	// Convert our Tool to mcp's ServerTool
	var mcpTools []server.ServerTool
	for _, tool := range tools {
		mcpTools = append(mcpTools, tool.toMCPServerTool())
	}
	s.mcpServer.AddTools(mcpTools...)
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
	log *slog.Logger,
) ServerOption {
	return func(s OptionSetter) error {
		return s.SetOption(server.WithToolHandlerMiddleware(
			func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
				return func(
					ctx context.Context,
					request mcp.CallToolRequest,
				) (result *mcp.CallToolResult, err error) {
					log.Info("MCP_CALL_REQUEST", "request", request)

					authenticatedCtx, err := AuthenticateRequest(ctx, client)
					if err != nil {
						return nil, err
					}

					resp, err := next(authenticatedCtx, request)
					if err != nil {
						log.Info("MCP_ERROR_RESPONSE", "response", resp)

						return nil, err
					}

					log.Info("MCP_CALL_RESPONSE", "response", resp)
					return resp, nil
				}
			}),
		)
	}
}
