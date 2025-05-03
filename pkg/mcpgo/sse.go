package mcpgo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/mark3labs/mcp-go/server"
)

type SSEConfig struct {
	// address is the address to bind the server to
	address string
	// port is the port to bind the server to
	port int
}

// getDefaultSSEConfig returns a default configuration for the SSE server
func getDefaultSSEConfig() *SSEConfig {
	return &SSEConfig{
		address: "localhost",
		port:    8080,
	}
}

// SSEConfigOpts defines a function type for applying configuration options
type SSEConfigOpts func(*SSEConfig)

// WithSSEAddress returns an option to set the server address
func WithSSEAddress(address string) SSEConfigOpts {
	return func(config *SSEConfig) {
		config.address = address
	}
}

// WithSSEPort returns an option to set the server port
func WithSSEPort(port int) SSEConfigOpts {
	return func(config *SSEConfig) {
		config.port = port
	}
}

// NewSSEConfig creates a new SSE server configuration with the provided options
func NewSSEConfig(opts ...SSEConfigOpts) *SSEConfig {
	config := getDefaultSSEConfig()

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// NewSseServer creates a new sse transport server
func NewSseServer(mcpServer Server, config *SSEConfig) (*mark3labsSseImpl, error) {
	sImpl, ok := mcpServer.(*mark3labsImpl)
	if !ok {
		return nil, fmt.Errorf("%w: expected *mark3labsImpl, got %T",
			ErrInvalidServerImplementation, mcpServer)
	}

	// Create a new SSE server with the base options
	sseServer := server.NewSSEServer(
		sImpl.mcpServer,
		server.WithBaseURL(fmt.Sprintf("http://%s:%d", config.address, config.port)),
		server.WithSSEContextFunc(authFromRequest),
	)

	// Wrap the server with a recovery handler
	impl := &mark3labsSseImpl{
		mcpSseServer: sseServer,
		SSEConfig:    config,
	}

	return impl, nil
}

// mark3labsSseImpl implements the TransportServer
// interface for sse transport
type mark3labsSseImpl struct {
	mcpSseServer *server.SSEServer
	SSEConfig    *SSEConfig
}

// Start implements the TransportServer interface
func (s *mark3labsSseImpl) Start() error {
	// Add panic recovery to the start method
	defer func() {
		if r := recover(); r != nil {
			// Log the panic, but don't crash the server
			fmt.Fprintf(os.Stderr, "Panic recovered in SSE server: %v\n", r)
			debug.PrintStack()
		}
	}()

	return s.mcpSseServer.Start(fmt.Sprintf(":%d", s.SSEConfig.port))
}

// authFromRequest extracts the auth token from the request headers.
func authFromRequest(ctx context.Context, r *http.Request) context.Context {
	auth := r.Header.Get("Authorization")

	// Split by space and take the second part (token)
	parts := strings.Split(auth, " ")
	if len(parts) > 1 {
		auth = parts[1]
	}

	return WithAuthToken(ctx, auth)
}
