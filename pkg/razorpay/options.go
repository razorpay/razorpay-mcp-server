package razorpay

import (
	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// ServerOption represents a configuration option for the server
type ServerOption func(*serverConfig)

// serverConfig holds the configuration for creating a new server
type serverConfig struct {
	observability   *observability.Observability
	client          *rzpsdk.Client
	version         string
	enabledToolsets []string
	readOnly        bool
}

// WithObservability sets the observability instance
func WithObservability(obs *observability.Observability) ServerOption {
	return func(c *serverConfig) {
		c.observability = obs
	}
}

// WithClient sets the Razorpay client
func WithClient(client *rzpsdk.Client) ServerOption {
	return func(c *serverConfig) {
		c.client = client
	}
}

// WithVersion sets the server version
func WithVersion(version string) ServerOption {
	return func(c *serverConfig) {
		c.version = version
	}
}

// WithEnabledToolsets sets the list of enabled toolsets
func WithEnabledToolsets(toolsets []string) ServerOption {
	return func(c *serverConfig) {
		c.enabledToolsets = toolsets
	}
}

// WithReadOnly sets whether the server operates in read-only mode
func WithReadOnly(readOnly bool) ServerOption {
	return func(c *serverConfig) {
		c.readOnly = readOnly
	}
}
