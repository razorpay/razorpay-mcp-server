package mcpgo

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	rzpsdk "github.com/razorpay/razorpay-go"
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

// WithAuthenticationMiddleware returns a server option that adds an authentication 
// middleware to the server.
func WithAuthenticationMiddleware(client *rzpsdk.Client) ServerOption {
	return func(s OptionSetter) error {
		return s.SetOption(server.WithToolHandlerMiddleware(
			func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
				return func(
					ctx context.Context,
					request mcp.CallToolRequest,
				) (result *mcp.CallToolResult, err error) {
					// If client is provided, this is the stdio mcp server
					if client != nil {
						return next(ctx, request)
					}

					// Check if auth token is provided
					auth := AuthTokenFromContext(ctx)
					if auth == "" {
						return nil, fmt.Errorf("unauthorized: no auth token provided")
					}

					// Base64 decode the auth token
					token, err := base64.StdEncoding.DecodeString(auth)
					if err != nil {
						return nil, fmt.Errorf("unauthorized: invalid auth token")
					}

					// Split token into key:secret
					parts := strings.Split(string(token), ":")
					if len(parts) != 2 {
						return nil, fmt.Errorf("unauthorized: invalid auth token")
					}

					// Create a new client with the auth credentials
					client := rzpsdk.NewClient(parts[0], parts[1])

					// Store the client in context
					ctx = WithClient(ctx, client)

					return next(ctx, request)
				}
			}),
		)
	}
}
