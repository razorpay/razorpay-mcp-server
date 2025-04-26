package mcpgo

import (
	"context"
	"io"
)

// TransportServer defines a server that can listen for MCP connections
type TransportServer interface {
	// Listen listens for connections
	Listen(ctx context.Context, in io.Reader, out io.Writer) error
}
