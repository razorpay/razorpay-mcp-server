package razorpay

import (
	"context"
	"fmt"
	"log/slog"

	rzpsdk "github.com/razorpay/razorpay-go/v2"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/toolsets"
)

// Server extends mcpgo.Server
type Server struct {
	log      *slog.Logger
	client   *rzpsdk.Client
	server   mcpgo.Server
	toolsets *toolsets.ToolsetGroup
}

// NewServer creates a new Server
func NewServer(
	log *slog.Logger,
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
	toolsets, err := NewToolSets(log, client, enabledToolsets, readOnly)
	if err != nil {
		return nil, err
	}

	// Create the server instance
	srv := &Server{
		log:      log,
		client:   client,
		server:   server,
		toolsets: toolsets,
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

// GetAllTools returns all registered tools
func (s *Server) GetAllTools() []mcpgo.Tool {
	var allTools []mcpgo.Tool

	// Iterate through all toolsets and collect their tools
	for _, toolset := range s.toolsets.Toolsets {
		if toolset.Enabled {
			allTools = append(allTools, toolset.ReadTools()...)
			if !s.toolsets.ReadOnly() {
				allTools = append(allTools, toolset.WriteTools()...)
			}
		}
	}

	return allTools
}

// CallTool calls a specific tool by name with the provided arguments
func (s *Server) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (interface{}, error) {
	// Find the tool by name
	tools := s.GetAllTools()
	var targetTool mcpgo.Tool

	for _, tool := range tools {
		if tool.GetName() == name {
			targetTool = tool
			break
		}
	}

	if targetTool == nil {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	// Create a call tool request
	request := mcpgo.CallToolRequest{
		Name:      name,
		Arguments: arguments,
	}

	// Call the tool
	result, err := targetTool.Call(ctx, request)
	if err != nil {
		return nil, err
	}

	return result, nil
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

	clientInterface := mcpgo.ClientFromContext(ctx)
	if clientInterface == nil {
		return nil, fmt.Errorf("no client found in context")
	}

	client, ok := clientInterface.(*rzpsdk.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type in context")
	}

	return client, nil
}
