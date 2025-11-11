package contextkey

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithClient(t *testing.T) {
	t.Run("adds client to context", func(t *testing.T) {
		ctx := context.Background()
		client := "test-client"

		newCtx := WithClient(ctx, client)

		assert.NotNil(t, newCtx)
		// Verify the client can be retrieved
		retrieved := ClientFromContext(newCtx)
		assert.Equal(t, client, retrieved)
	})

	t.Run("adds client to context with existing values", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "existing-key", "existing-value")
		client := map[string]interface{}{
			"key": "value",
		}

		newCtx := WithClient(ctx, client)

		assert.NotNil(t, newCtx)
		// Verify existing value is preserved
		assert.Equal(t, "existing-value", newCtx.Value("existing-key"))
		// Verify client can be retrieved
		retrieved := ClientFromContext(newCtx)
		assert.Equal(t, client, retrieved)
	})

	t.Run("adds nil client to context", func(t *testing.T) {
		ctx := context.Background()

		newCtx := WithClient(ctx, nil)

		assert.NotNil(t, newCtx)
		retrieved := ClientFromContext(newCtx)
		assert.Nil(t, retrieved)
	})

	t.Run("overwrites existing client in context", func(t *testing.T) {
		ctx := context.Background()
		client1 := "client-1"
		client2 := "client-2"

		ctx1 := WithClient(ctx, client1)
		ctx2 := WithClient(ctx1, client2)

		// Original context should still have client1
		assert.Equal(t, client1, ClientFromContext(ctx1))
		// New context should have client2
		assert.Equal(t, client2, ClientFromContext(ctx2))
	})
}

func TestClientFromContext(t *testing.T) {
	t.Run("retrieves client from context", func(t *testing.T) {
		ctx := context.Background()
		client := "test-client"

		ctx = WithClient(ctx, client)
		retrieved := ClientFromContext(ctx)

		assert.Equal(t, client, retrieved)
	})

	t.Run("returns nil when no client in context", func(t *testing.T) {
		ctx := context.Background()

		retrieved := ClientFromContext(ctx)

		assert.Nil(t, retrieved)
	})

	t.Run("retrieves complex client object", func(t *testing.T) {
		ctx := context.Background()
		client := map[string]interface{}{
			"name": "test",
			"id":   123,
		}

		ctx = WithClient(ctx, client)
		retrieved := ClientFromContext(ctx)

		assert.NotNil(t, retrieved)
		if clientMap, ok := retrieved.(map[string]interface{}); ok {
			assert.Equal(t, "test", clientMap["name"])
			assert.Equal(t, 123, clientMap["id"])
		} else {
			t.Fatal("retrieved client is not a map")
		}
	})

	t.Run("retrieves client from nested context", func(t *testing.T) {
		ctx := context.Background()
		client := "test-client"

		ctx = WithClient(ctx, client)
		ctx = context.WithValue(ctx, "other-key", "other-value")

		retrieved := ClientFromContext(ctx)
		assert.Equal(t, client, retrieved)
	})
}
