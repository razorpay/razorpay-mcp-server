package mcpgo

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"

	"github.com/razorpay/razorpay-mcp-server/pkg/contextkey"
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
		server.WithBaseURL(config.address),
		server.WithSSEContextFunc(authFromRequest),
		server.WithKeepAlive(true),
		server.WithKeepAliveInterval(10*time.Second),
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
		Addr:              fmt.Sprintf(":%d", s.SSEConfig.port),
		Handler:           s.mux,
		ReadHeaderTimeout: 1 * time.Second,
	}

	// Start the HTTP server
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the SSE server
func (s *mark3labsSseImpl) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// handleLiveness returns 200 OK for liveness probe
func (s *mark3labsSseImpl) handleLiveness(
	w http.ResponseWriter,
	_ *http.Request,
) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}
}

// handleReadiness returns 200 OK for readiness probe
func (s *mark3labsSseImpl) handleReadiness(
	w http.ResponseWriter,
	_ *http.Request,
) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}
}

// authFromRequest extracts the auth token from the request headers.
func authFromRequest(ctx context.Context, r *http.Request) context.Context {
	// Extract task ID from X-Task-Id header
	if taskID := r.Header.Get("X-Task-Id"); taskID != "" {
		ctx = contextkey.WithTaskID(ctx, taskID)
	} else {
		ctx = contextkey.WithTaskID(ctx, uuid.New().String())
	}

	ctx = contextkey.WithRequestID(ctx, uuid.New().String())

	authHeader := r.Header.Get("Authorization")

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ctx
	}

	ctx = setRzpKeyInContext(ctx, parts[1])
	return WithAuthToken(ctx, parts[1])
}

func setRzpKeyInContext(
	ctx context.Context,
	basicAuthToken string,
) context.Context {
	token, err := base64.StdEncoding.DecodeString(basicAuthToken)
	if err != nil {
		return ctx
	}

	parts := strings.Split(string(token), ":")
	if len(parts) != 2 {
		return ctx
	}

	if parts[0] != "" {
		ctx = contextkey.WithRzpKey(ctx, parts[0])
	}

	return ctx
}
