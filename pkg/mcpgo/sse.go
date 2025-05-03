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

type sseConfig struct {
	// address is the address to bind the server to
	address string
	// port is the port to bind the server to
	port int
}

func getDefaultSseConfig() *sseConfig {
	return &sseConfig{
		address: "localhost",
		port:    8080,
	}
}

type sseConfigOpts func(*sseConfig)

func WithSseAddress(address string) sseConfigOpts {
	return func(config *sseConfig) {
		config.address = address
	}
}

func WithSsePort(port int) sseConfigOpts {
	return func(config *sseConfig) {
		config.port = port
	}
}

func NewSseConfig(opts ...sseConfigOpts) *sseConfig {
	config := getDefaultSseConfig()

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// NewSseServer creates a new sse transport server
func NewSseServer(mcpServer Server, config *sseConfig) (*mark3labsSseImpl, error) {
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
		sseConfig:    config,
	}

	return impl, nil
}

// mark3labsSseImpl implements the TransportServer
// interface for sse transport
type mark3labsSseImpl struct {
	mcpSseServer *server.SSEServer
	sseConfig    *sseConfig
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

	return s.mcpSseServer.Start(fmt.Sprintf(":%d", s.sseConfig.port))
}

// authKey is a custom context key for storing the auth token.
type authKey struct{}

// withAuthKey adds an auth key to the context.
func withAuthKey(ctx context.Context, auth string) context.Context {
	return context.WithValue(ctx, authKey{}, auth)
}

// authFromRequest extracts the auth token from the request headers.
func authFromRequest(ctx context.Context, r *http.Request) context.Context {
	auth := r.Header.Get("Authorization")

	// Split by space and take the second part (token)
	parts := strings.Split(auth, " ")
	if len(parts) > 1 {
		auth = parts[1]
	}

	return withAuthKey(ctx, auth)
}

func authFromContext(ctx context.Context) string {
	value := ctx.Value(authKey{})
	if value == nil {
		return ""
	}

	auth, ok := value.(string)
	if !ok {
		return ""
	}

	return auth
}