package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/contextkey"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/toolsets"
)

// Server extends mcpgo.Server
type Server struct {
	server   mcpgo.Server
	toolsets *toolsets.ToolsetGroup
}

// NewServer creates a new Server with the provided options
func NewServer(opts ...ServerOption) (*Server, error) {
	// Default configuration
	config := &serverConfig{
		version:         "1.0.0",
		enabledToolsets: []string{},
		readOnly:        false,
		serverName:      "razorpay-mcp-server",
		enableResources: true,
		enableTools:     true,
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Validate required fields
	if config.observability == nil {
		return nil, fmt.Errorf("observability is required")
	}

	// Build MCP server options with configured capabilities
	mcpOpts := []mcpgo.ServerOption{
		mcpgo.WithLogging(),
		mcpgo.WithResourceCapabilities(config.enableResources, config.enableResources),
		mcpgo.WithToolCapabilities(config.enableTools),
		mcpgo.WithHooks(mcpgo.SetupHooks(config.observability)),

		// Add Tool Middlewares
		mcpgo.WithAuthenticationMiddleware(config.client),
	}

	// Create the mcpgo server
	server := mcpgo.NewServer(
		config.serverName,
		config.version,
		mcpOpts...,
	)

	// Handle toolsets - use custom or create new
	var toolsets *toolsets.ToolsetGroup
	var err error

	toolsets = config.customToolsets
	if toolsets == nil {
		// Initialize toolsets with configuration
		toolsets, err = NewToolSets(config.observability, config.client, config.enabledToolsets, config.readOnly)
		if err != nil {
			return nil, fmt.Errorf("failed to create toolsets: %w", err)
		}
	}

	// Create the server instance
	srv := &Server{
		server:   server,
		toolsets: toolsets,
	}

	// Register tools based on configuration
	srv.RegisterTools()

	return srv, nil
}

// RegisterTools adds all available tools to the server
func (s *Server) RegisterTools() {
	s.toolsets.RegisterTools(s.server)
}

// GetMCPServer returns the underlying MCP server instance
func (s *Server) GetMCPServer() mcpgo.Server {
	return s.server
}

// getClientFromContextOrDefault returns either the provided default
// client or gets one from context.
func getClientFromContextOrDefault(
	ctx context.Context,
	defaultClient *rzpsdk.Client,
) (*rzpsdk.Client, error) {
	if defaultClient != nil {
		return defaultClient, nil
	}

	clientInterface := contextkey.ClientFromContext(ctx)
	if clientInterface == nil {
		return nil, fmt.Errorf("no client found in context")
	}

	client, ok := clientInterface.(*rzpsdk.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type in context")
	}

	return client, nil
}
