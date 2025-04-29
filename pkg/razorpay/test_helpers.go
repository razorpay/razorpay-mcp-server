package razorpay

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"

	"github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// RazorpayToolTestCase defines a common structure for Razorpay tool tests
type RazorpayToolTestCase struct {
	Name           string
	Request        map[string]interface{}
	MockHttpClient func() (*http.Client, *httptest.Server)
	ExpectError    bool
	ExpectedResult map[string]interface{}
	ExpectedErrMsg string
}

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

// newMockRzpClient configures a Razorpay client with a mock
// HTTP client for testing. It returns the configured client
// and the mock server (which should be closed by the caller)
func newMockRzpClient(
	mockHttpClient func() (*http.Client, *httptest.Server),
) (*razorpay.Client, *httptest.Server) {
	rzpMockClient := razorpay.NewClient("sample_key", "sample_secret")

	var mockServer *httptest.Server
	if mockHttpClient != nil {
		var client *http.Client
		client, mockServer = mockHttpClient()

		// This Request object is shared by reference across all
		// API resources in the client
		req := rzpMockClient.Order.Request
		req.BaseURL = mockServer.URL
		req.HTTPClient = client
	}

	return rzpMockClient, mockServer
}

// runToolTest executes a common test pattern for Razorpay tools
func runToolTest(
	t *testing.T,
	tc RazorpayToolTestCase,
	toolCreator func(*slog.Logger, *razorpay.Client) mcpgo.Tool,
	objectType string,
) {
	mockRzpClient, mockServer := newMockRzpClient(tc.MockHttpClient)
	if mockServer != nil {
		defer mockServer.Close()
	}

	log := CreateTestLogger()
	tool := toolCreator(log, mockRzpClient)

	request := createMCPRequest(tc.Request)
	result, err := tool.GetHandler()(context.Background(), request)

	assert.NoError(t, err)

	if tc.ExpectError {
		assert.NotNil(t, result)
		assert.Contains(t, result.Text, tc.ExpectedErrMsg)
		return
	}

	assert.NotNil(t, result)

	var returnedObj map[string]interface{}
	err = json.Unmarshal([]byte(result.Text), &returnedObj)
	assert.NoError(t, err)

	if diff := deep.Equal(tc.ExpectedResult, returnedObj); diff != nil {
		t.Errorf("%s mismatch: %s", objectType, diff)
	}
}
