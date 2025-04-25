package razorpay

import (
	"fmt"
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

	// orders tools
	s.server.AddTools(
		CreateOrder(s.log, s.client),
		FetchOrder(s.log, s.client),
	)

	// payment links tools
	s.server.AddTools(
		CreatePaymentLink(s.log, s.client),
		FetchPaymentLink(s.log, s.client),
	)
}

// GetMCPServer returns the underlying MCP server instance
func (s *Server) GetMCPServer() mcpgo.Server {
	return s.server
}

// RequiredParam gets a required parameter
func RequiredParam[T any](r mcpgo.CallToolRequest, name string) (T, error) {
	var zero T
	arg, exists := r.Arguments[name]
	if !exists {
		return zero, fmt.Errorf("missing required parameter: %s", name)
	}

	value, ok := arg.(T)
	if !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T", name, zero)
	}

	return value, nil
}

// OptionalParam gets an optional parameter
func OptionalParam[T any](r mcpgo.CallToolRequest, name string) (T, error) {
	var zero T
	arg, exists := r.Arguments[name]
	if !exists {
		return zero, nil
	}

	value, ok := arg.(T)
	if !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T", name, zero)
	}

	return value, nil
}

// RequiredInt gets a required integer parameter
func RequiredInt(r mcpgo.CallToolRequest, name string) (int, error) {
	// First get as float64 (JSON numbers are floats)
	val, err := RequiredParam[float64](r, name)
	if err != nil {
		return 0, err
	}

	return int(val), nil
}

// OptionalInt gets an optional integer parameter
func OptionalInt(r mcpgo.CallToolRequest, name string) (int, error) {
	// First get as float64 (JSON numbers are floats)
	val, err := OptionalParam[float64](r, name)
	if err != nil {
		return 0, err
	}

	return int(val), nil
}

// HandleValidationError is a helper for handling validation
// errors in tool handlers
func HandleValidationError(err error) (*mcpgo.ToolResult, error) {
	if err != nil {
		return mcpgo.NewToolResultError(err.Error()), nil
	}
	return nil, nil
}
