package razorpay

import (
	"log/slog"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"

	rzpsdk "github.com/razorpay/razorpay-go"
)

// Server extends mcpgo.Server
type Server struct {
	log    *slog.Logger
	client *rzpsdk.Client
	server mcpgo.Server
}

// NewServer creates a new Server
func NewServer(
	log *slog.Logger,
	client *rzpsdk.Client,
	version string,
) *Server {
	// Create default options
	opts := []mcpgo.ServerOption{
		mcpgo.WithLogging(),
		mcpgo.WithResourceCapabilities(true, true),
		mcpgo.WithToolCapabilities(true),
	}

	// Create the mcpgo server
	server := mcpgo.NewServer(
		"razorpay-mcp-server",
		version,
		opts...,
	)

	return &Server{
		log:    log,
		client: client,
		server: server,
	}
}

// RegisterTools adds all available tools to the server
func (s *Server) RegisterTools() {
	// payments tools
	s.server.AddTools(FetchPayment(s.log, s.client))
}

// GetMCPServer returns the underlying MCP server instance
func (s *Server) GetMCPServer() mcpgo.Server {
	return s.server
}
