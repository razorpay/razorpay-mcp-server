package razorpay

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/contextkey"
)

func TestNewRzpMcpServer(t *testing.T) {
	t.Run("creates server successfully", func(t *testing.T) {
		obs := CreateTestObservability()
		client := rzpsdk.NewClient("test-key", "test-secret")

		server, err := NewRzpMcpServer(obs, client, []string{}, false)
		assert.NoError(t, err)
		assert.NotNil(t, server)
	})

	t.Run("returns error with nil observability", func(t *testing.T) {
		client := rzpsdk.NewClient("test-key", "test-secret")

		server, err := NewRzpMcpServer(nil, client, []string{}, false)
		assert.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "observability is required")
	})

	t.Run("returns error with nil client", func(t *testing.T) {
		obs := CreateTestObservability()

		server, err := NewRzpMcpServer(obs, nil, []string{}, false)
		assert.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "razorpay client is required")
	})

	t.Run("creates server with enabled toolsets", func(t *testing.T) {
		obs := CreateTestObservability()
		client := rzpsdk.NewClient("test-key", "test-secret")

		server, err := NewRzpMcpServer(
			obs, client, []string{"payments", "orders"}, false)
		assert.NoError(t, err)
		assert.NotNil(t, server)
	})

	t.Run("creates server in read-only mode", func(t *testing.T) {
		obs := CreateTestObservability()
		client := rzpsdk.NewClient("test-key", "test-secret")

		server, err := NewRzpMcpServer(obs, client, []string{}, true)
		assert.NoError(t, err)
		assert.NotNil(t, server)
	})

	t.Run("creates server with custom mcp options", func(t *testing.T) {
		obs := CreateTestObservability()
		client := rzpsdk.NewClient("test-key", "test-secret")

		server, err := NewRzpMcpServer(obs, client, []string{}, false)
		assert.NoError(t, err)
		assert.NotNil(t, server)
	})
}

func TestGetClientFromContextOrDefault(t *testing.T) {
	t.Run("returns default client when provided", func(t *testing.T) {
		ctx := context.Background()
		client := rzpsdk.NewClient("test-key", "test-secret")

		result, err := getClientFromContextOrDefault(ctx, client)
		assert.NoError(t, err)
		assert.Equal(t, client, result)
	})

	t.Run("returns client from context", func(t *testing.T) {
		ctx := context.Background()
		client := rzpsdk.NewClient("test-key", "test-secret")
		ctx = contextkey.WithClient(ctx, client)

		result, err := getClientFromContextOrDefault(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, client, result)
	})

	t.Run("returns error when no client in context and no default",
		func(t *testing.T) {
			ctx := context.Background()

			result, err := getClientFromContextOrDefault(ctx, nil)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "no client found in context")
		})

	t.Run("returns error when client in context has wrong type",
		func(t *testing.T) {
			ctx := context.Background()
			ctx = contextkey.WithClient(ctx, "not-a-client")

			result, err := getClientFromContextOrDefault(ctx, nil)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "invalid client type in context")
		})

	t.Run("prefers default client over context client", func(t *testing.T) {
		ctx := context.Background()
		defaultClient := rzpsdk.NewClient("default-key", "default-secret")
		contextClient := rzpsdk.NewClient("context-key", "context-secret")
		ctx = contextkey.WithClient(ctx, contextClient)

		result, err := getClientFromContextOrDefault(ctx, defaultClient)
		assert.NoError(t, err)
		assert.Equal(t, defaultClient, result)
		assert.NotEqual(t, contextClient, result)
	})
}
