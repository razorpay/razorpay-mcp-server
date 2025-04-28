package razorpay

import (
	"io"
	"log/slog"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// CreateTestLogger creates a logger suitable for testing
func CreateTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// createMCPRequest creates a CallToolRequest with the given arguments
func createMCPRequest(args map[string]interface{}) mcpgo.CallToolRequest {
	return mcpgo.CallToolRequest{
		Arguments: args,
	}
}
