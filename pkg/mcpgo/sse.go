package mcpgo

import (
	"context"
	"fmt"
	"net/http"
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

// NewSSEServer creates a new sse transport server
func NewSSEServer(
	mcpServer Server,
	config *SSEConfig,
) (*mark3labsSseImpl, error) {
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
	httpServer   *http.Server
	mux          *http.ServeMux
}

// Start implements the TransportServer interface
func (s *mark3labsSseImpl) Start() error {
	s.mux = http.NewServeMux()

	// Register health check endpoints
	s.mux.HandleFunc("/live", s.handleLiveness)
	s.mux.HandleFunc("/ready", s.handleReadiness)

	// Register SSE server as default handler for all other routes
	s.mux.Handle("/", s.mcpSseServer)

	// Create HTTP server with our custom mux
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.SSEConfig.port),
		Handler: s.mux,
	}

	// Start the HTTP server
	return s.httpServer.ListenAndServe()
}

// handleLiveness returns 200 OK for liveness probe
func (s *mark3labsSseImpl) handleLiveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReadiness returns 200 OK for readiness probe
func (s *mark3labsSseImpl) handleReadiness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// authFromRequest extracts the auth token from the request headers.
func authFromRequest(ctx context.Context, r *http.Request) context.Context {
	authHeader := r.Header.Get("Authorization")

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ctx
	}

	return WithAuthToken(ctx, parts[1])
}
