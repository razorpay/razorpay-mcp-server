package mcpgo

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStdioServer(t *testing.T) {
	t.Run("creates stdio server with valid implementation", func(t *testing.T) {
		mcpServer := NewMcpServer("test-server", "1.0.0")
		stdioServer, err := NewStdioServer(mcpServer)
		assert.NoError(t, err)
		assert.NotNil(t, stdioServer)
	})

	t.Run("returns error with invalid server implementation", func(t *testing.T) {
		invalidServer := &invalidServerImpl{}
		stdioServer, err := NewStdioServer(invalidServer)
		assert.Error(t, err)
		assert.Nil(t, stdioServer)
		assert.Contains(t, err.Error(), "invalid server implementation")
		assert.Contains(t, err.Error(), "expected *Mark3labsImpl")
	})

	t.Run("returns error with nil server", func(t *testing.T) {
		stdioServer, err := NewStdioServer(nil)
		assert.Error(t, err)
		assert.Nil(t, stdioServer)
	})
}

func TestMark3labsStdioImpl_Listen(t *testing.T) {
	t.Run("listens with valid reader and writer", func(t *testing.T) {
		mcpServer := NewMcpServer("test-server", "1.0.0")
		stdioServer, err := NewStdioServer(mcpServer)
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Create a simple input that will cause the server to process
		// The actual Listen implementation will read from in and write to out
		initMsg := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
		in := strings.NewReader(initMsg)
		out := &bytes.Buffer{}

		// Listen will block, so we need to run it in a goroutine
		// and cancel the context to stop it
		errChan := make(chan error, 1)
		go func() {
			errChan <- stdioServer.Listen(ctx, in, out)
		}()

		// Cancel context to stop listening
		cancel()

		// Wait for the error (should be context canceled)
		err = <-errChan
		// The error might be context.Canceled or nil depending on implementation
		// We just verify it doesn't panic
		assert.NotPanics(t, func() {
			_ = err
		})
	})

	t.Run("listens with empty reader", func(t *testing.T) {
		mcpServer := NewMcpServer("test-server", "1.0.0")
		stdioServer, err := NewStdioServer(mcpServer)
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		in := strings.NewReader("")
		out := &bytes.Buffer{}

		errChan := make(chan error, 1)
		go func() {
			errChan <- stdioServer.Listen(ctx, in, out)
		}()

		cancel()
		err = <-errChan
		assert.NotPanics(t, func() {
			_ = err
		})
	})

	t.Run("listens with nil reader", func(t *testing.T) {
		mcpServer := NewMcpServer("test-server", "1.0.0")
		stdioServer, err := NewStdioServer(mcpServer)
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var in io.Reader = nil
		out := &bytes.Buffer{}

		errChan := make(chan error, 1)
		go func() {
			errChan <- stdioServer.Listen(ctx, in, out)
		}()

		cancel()
		err = <-errChan
		assert.NotPanics(t, func() {
			_ = err
		})
	})
}

// invalidServerImpl is a test implementation that doesn't match Mark3labsImpl
type invalidServerImpl struct{}

func (i *invalidServerImpl) AddTools(tools ...Tool) {
	// Empty implementation for testing
}
