package razorpay

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/razorpay/razorpay-go"
	
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

// newRzpMockClient configures a Razorpay client with a mock HTTP client for testing
// It returns the configured client and the mock server (which should be closed by the caller)
func newRzpMockClient(
	mockHttpClient func() (*http.Client, *httptest.Server),
) (*razorpay.Client, *httptest.Server) {
	mockrzpClient := razorpay.NewClient("sample_key", "sample_secret")

	var mockServer *httptest.Server
	if mockHttpClient != nil {
		var client *http.Client
		client, mockServer = mockHttpClient()

		req := mockrzpClient.Order.Request
		req.BaseURL = mockServer.URL
		req.HTTPClient = client
	}

	return mockrzpClient, mockServer
}
