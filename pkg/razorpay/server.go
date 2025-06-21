package razorpay

import (
	"context"
	"fmt"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/contextkey"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

func NewRzpMcpServer(
	obs *observability.Observability,
	client *rzpsdk.Client,
	enabledToolsets []string,
	readOnly bool,
	mcpOpts ...mcpgo.ServerOption,
) (mcpgo.Server, error) {
	// Validate required parameters
	if obs == nil {
		return nil, fmt.Errorf("observability is required")
	}
	if client == nil {
		return nil, fmt.Errorf("razorpay client is required")
	}

	// Set up default MCP options with Razorpay-specific hooks
	defaultOpts := []mcpgo.ServerOption{
		mcpgo.WithLogging(),
		mcpgo.WithResourceCapabilities(true, true),
		mcpgo.WithToolCapabilities(true),
		mcpgo.WithHooks(mcpgo.SetupHooks(obs)),
	}
	// Merge with user-provided options
	allOpts := append(defaultOpts, mcpOpts...)

	// Create server
	server := mcpgo.NewMcpServer("razorpay-mcp-server", "1.0.0", allOpts...)

	// Register Razorpay tools
	toolsets, err := NewToolSets(obs, client, enabledToolsets, readOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to create toolsets: %w", err)
	}
	toolsets.RegisterTools(server)

	return server, nil
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
