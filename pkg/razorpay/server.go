package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/contextkey"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
	"github.com/razorpay/razorpay-mcp-server/pkg/toolsets"
)

// Server extends mcpgo.Server
type Server struct {
	observability *observability.Observability
	client        *rzpsdk.Client
	server        mcpgo.Server
	toolsets      *toolsets.ToolsetGroup
}

// NewServer creates a new Server
func NewServer(
	obs *observability.Observability,
	client *rzpsdk.Client,
	version string,
	enabledToolsets []string,
	readOnly bool,
) (*Server, error) {

	// Create default options
	opts := []mcpgo.ServerOption{
		mcpgo.WithLogging(),
		mcpgo.WithResourceCapabilities(true, true),
		mcpgo.WithToolCapabilities(true),
		mcpgo.WithHooks(mcpgo.SetupHooks(obs)),

		// Add Tool Middlewares
		mcpgo.WithAuthenticationMiddleware(client),
	}

	// Create the mcpgo server
	server := mcpgo.NewServer(
		"razorpay-mcp-server",
		version,
		opts...,
	)

	// Initialize toolsets
	toolsets, err := NewToolSets(obs, client, enabledToolsets, readOnly)
	if err != nil {
		return nil, err
	}

	// Create the server instance
	srv := &Server{
		observability: obs,
		client:        client,
		server:        server,
		toolsets:      toolsets,
	}

	// Register all tools
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
