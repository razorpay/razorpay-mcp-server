package razorpay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/log"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
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

// CreateTestObservability creates an observability stack suitable for testing
func CreateTestObservability() *observability.Observability {
	// Create a logger that discards output
	_, logger := log.New(context.Background(), log.NewConfig(
		log.WithMode(log.ModeStdio)),
	)
	return &observability.Observability{
		Logger: logger,
	}
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
) (*rzpsdk.Client, *httptest.Server) {
	rzpMockClient := rzpsdk.NewClient("sample_key", "sample_secret")

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
	toolCreator func(*observability.Observability, *rzpsdk.Client) mcpgo.Tool,
	objectType string,
) {
	mockRzpClient, mockServer := newMockRzpClient(tc.MockHttpClient)
	if mockServer != nil {
		defer mockServer.Close()
	}

	obs := CreateTestObservability()
	tool := toolCreator(obs, mockRzpClient)

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
