package razorpay

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/contextkey"
	"github.com/razorpay/razorpay-mcp-server/pkg/log"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// createMockObservability creates a mock observability object for testing
func createMockObservability() *observability.Observability {
	// Create a mock logger config for sse mode
	config := log.NewConfig(
		log.WithMode(log.ModeSSE),
	)

	// Create observability with a mock logger
	ctx := context.Background()
	obs, _ := observability.New(ctx, config)
	return obs
}

// TestConcurrentRequestHandling tests concurrent requests with different
// auth tokens
func TestConcurrentRequestHandling(t *testing.T) {
	var capturedClients sync.Map

	testTool := mcpgo.NewTool(
		"test_tool",
		"Test tool for concurrent testing",
		[]mcpgo.ToolParameter{
			mcpgo.WithString("test_param",
				mcpgo.Description("Test parameter"), mcpgo.Required()),
		},
		func(
			ctx context.Context,
			request mcpgo.CallToolRequest,
		) (*mcpgo.ToolResult, error) {
			client := contextkey.ClientFromContext(ctx)
			if client == nil {
				return mcpgo.NewToolResultError("no client found in context"), nil
			}

			requestID := request.Arguments["test_param"].(string)
			capturedClients.Store(requestID, client)

			return mcpgo.NewToolResultText(fmt.Sprintf("processed: %s", requestID)), nil
		},
	)

	tokens := []string{
		base64.StdEncoding.EncodeToString([]byte("key1:secret1")),
		base64.StdEncoding.EncodeToString([]byte("key2:secret2")),
		base64.StdEncoding.EncodeToString([]byte("key3:secret3")),
	}

	middleware := createAuthMiddleware(nil)

	var wg sync.WaitGroup
	var errors sync.Map
	numRequests := len(tokens)

	for i, token := range tokens {
		wg.Add(1)
		go func(requestID int, authToken string) {
			defer wg.Done()

			ctx := context.Background()
			ctx = contextkey.WithAuthToken(ctx, authToken)

			request := createMCPRequest(map[string]interface{}{
				"test_param": fmt.Sprintf("request_%d", requestID),
			})

			handler := middleware(testTool.GetHandler())
			_, err := handler(ctx, request)

			if err != nil {
				errors.Store(requestID, err.Error())
			}
		}(i, token)
	}

	wg.Wait()

	errors.Range(func(key, value interface{}) bool {
		t.Errorf("Request %v failed with error: %v", key, value)
		return true
	})

	clientCount := 0
	capturedClients.Range(func(key, value interface{}) bool {
		clientCount++
		assert.NotNil(t, value, "Client should not be nil for request %v", key)
		return true
	})
	assert.Equal(t, numRequests, clientCount,
		"Should have captured clients for all requests")
}

// createAuthMiddleware creates a middleware function for testing purposes
// that uses the shared authentication logic from mcpgo.AuthenticateRequest
func createAuthMiddleware(
	client *rzpsdk.Client,
) func(mcpgo.ToolHandler) mcpgo.ToolHandler {
	return func(next mcpgo.ToolHandler) mcpgo.ToolHandler {
		return func(
			ctx context.Context,
			request mcpgo.CallToolRequest,
		) (*mcpgo.ToolResult, error) {
			authenticatedCtx, err := mcpgo.AuthenticateRequest(ctx, client)
			if err != nil {
				return nil, err
			}
			return next(authenticatedCtx, request)
		}
	}
}
