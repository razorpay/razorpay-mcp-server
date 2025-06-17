package mcpgo

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/mark3labs/mcp-go/server"
)

// ErrInvalidServerImplementation indicates that the server
// implementation is not compatible
var ErrInvalidServerImplementation = errors.New(
	"invalid server implementation",
)

// NewStdioServer creates a new stdio transport server
func NewStdioServer(mcpServer Server) (*mark3labsStdioImpl, error) {
	sImpl, ok := mcpServer.(*Mark3labsImpl)
	if !ok {
		return nil, fmt.Errorf("%w: expected *Mark3labsImpl, got %T",
			ErrInvalidServerImplementation, mcpServer)
	}

	return &mark3labsStdioImpl{
		mcpStdioServer: server.NewStdioServer(sImpl.McpServer),
	}, nil
}

// mark3labsStdioImpl implements the TransportServer
// interface for stdio transport
type mark3labsStdioImpl struct {
	mcpStdioServer *server.StdioServer
}

// Listen implements the TransportServer interface
func (s *mark3labsStdioImpl) Listen(
	ctx context.Context, in io.Reader, out io.Writer) error {
	return s.mcpStdioServer.Listen(ctx, in, out)
}
