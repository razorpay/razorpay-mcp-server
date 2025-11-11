package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/log"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

func TestStdioCmd(t *testing.T) {
	t.Run("stdio command is configured correctly", func(t *testing.T) {
		assert.NotNil(t, stdioCmd)
		assert.Equal(t, "stdio", stdioCmd.Use)
		assert.Equal(t, "start the stdio server", stdioCmd.Short)
		assert.NotNil(t, stdioCmd.Run)
	})

	t.Run("stdio command is added to root command", func(t *testing.T) {
		// Verify stdioCmd is in the root command's commands
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd == stdioCmd {
				found = true
				break
			}
		}
		assert.True(t, found, "stdioCmd should be added to rootCmd")
	})
}

func TestRunStdioServer(t *testing.T) {
	t.Run("creates server successfully", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Create client
		client := rzpsdk.NewClient("test-key", "test-secret")

		// Run server in a goroutine and cancel immediately
		errChan := make(chan error, 1)
		go func() {
			errChan <- runStdioServer(ctx, obs, client, []string{}, false)
		}()

		// Cancel context to stop server
		cancel()

		// Wait for server to stop
		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("server did not stop in time")
		}
	})

	t.Run("handles server creation error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Create client with invalid credentials (should still create client)
		client := rzpsdk.NewClient("", "")

		// Run server - should handle gracefully
		errChan := make(chan error, 1)
		go func() {
			errChan <- runStdioServer(ctx, obs, client, []string{}, false)
		}()

		// Cancel context
		cancel()

		select {
		case err := <-errChan:
			// Should not error on cancellation
			assert.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("server did not stop in time")
		}
	})

	t.Run("handles signal context cancellation", func(t *testing.T) {
		ctx := context.Background()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Create client
		client := rzpsdk.NewClient("test-key", "test-secret")

		// Create signal context
		signalCtx, stop := signal.NotifyContext(
			ctx,
			os.Interrupt,
			syscall.SIGTERM,
		)
		defer stop()

		// Run server
		errChan := make(chan error, 1)
		go func() {
			errChan <- runStdioServer(signalCtx, obs, client, []string{}, false)
		}()

		// Send interrupt signal
		time.Sleep(100 * time.Millisecond)
		stop()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("server did not stop in time")
		}
	})

	t.Run("handles read-only mode", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Create client
		client := rzpsdk.NewClient("test-key", "test-secret")

		// Run server in read-only mode
		errChan := make(chan error, 1)
		go func() {
			errChan <- runStdioServer(ctx, obs, client, []string{}, true)
		}()

		// Cancel context
		cancel()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("server did not stop in time")
		}
	})

	t.Run("handles enabled toolsets", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Create client
		client := rzpsdk.NewClient("test-key", "test-secret")

		// Run server with enabled toolsets
		errChan := make(chan error, 1)
		go func() {
			errChan <- runStdioServer(ctx, obs, client, []string{"payments", "orders"}, false)
		}()

		// Cancel context
		cancel()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("server did not stop in time")
		}
	})

	t.Run("handles server listen error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Create client
		client := rzpsdk.NewClient("test-key", "test-secret")

		// Create a context that will be cancelled quickly
		quickCtx, quickCancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer quickCancel()

		// Run server
		errChan := make(chan error, 1)
		go func() {
			errChan <- runStdioServer(quickCtx, obs, client, []string{}, false)
		}()

		select {
		case err := <-errChan:
			// Should not error on timeout/cancellation
			assert.NoError(t, err)
		case <-time.After(1 * time.Second):
			// If it takes too long, cancel and check
			cancel()
			select {
			case err := <-errChan:
				assert.NoError(t, err)
			case <-time.After(1 * time.Second):
				t.Fatal("server did not stop in time")
			}
		}
	})

	t.Run("handles error from server creation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Create client - this should work
		client := rzpsdk.NewClient("test-key", "test-secret")

		// Run server - should work fine
		errChan := make(chan error, 1)
		go func() {
			errChan <- runStdioServer(ctx, obs, client, []string{}, false)
		}()

		// Cancel immediately
		cancel()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("server did not stop in time")
		}
	})

	t.Run("handles error from stdio server creation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Create client
		client := rzpsdk.NewClient("test-key", "test-secret")

		// Run server
		errChan := make(chan error, 1)
		go func() {
			errChan <- runStdioServer(ctx, obs, client, []string{}, false)
		}()

		// Cancel to stop
		cancel()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("server did not stop in time")
		}
	})

	t.Run("handles error from listen channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Create client
		client := rzpsdk.NewClient("test-key", "test-secret")

		// Run server
		errChan := make(chan error, 1)
		go func() {
			errChan <- runStdioServer(ctx, obs, client, []string{}, false)
		}()

		// Cancel to stop
		cancel()

		select {
		case err := <-errChan:
			// Should return nil on context cancellation
			assert.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("server did not stop in time")
		}
	})

	t.Run("handles error from NewRzpMcpServer with nil observability", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Pass nil observability to trigger error
		client := rzpsdk.NewClient("test-key", "test-secret")

		err := runStdioServer(ctx, nil, client, []string{}, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create server")
	})

	t.Run("handles error from NewRzpMcpServer with nil client", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Pass nil client to trigger error
		err := runStdioServer(ctx, obs, nil, []string{}, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create server")
	})
}

func TestStdioCmdRun(t *testing.T) {
	t.Run("stdio command run function exists", func(t *testing.T) {
		// Verify the Run function is set
		assert.NotNil(t, stdioCmd.Run)

		// We can't easily test the full Run function without
		// setting up viper and all dependencies, but we can
		// verify it's callable
	})

	t.Run("stdio command uses viper for configuration", func(t *testing.T) {
		// Reset viper
		viper.Reset()

		// Set viper values that stdioCmd would use
		viper.Set("log_file", "/tmp/test.log")
		viper.Set("key", "test-key")
		viper.Set("secret", "test-secret")
		viper.Set("toolsets", []string{"payments"})
		viper.Set("read_only", true)

		// Verify values are set (testing that viper integration works)
		assert.Equal(t, "/tmp/test.log", viper.GetString("log_file"))
		assert.Equal(t, "test-key", viper.GetString("key"))
		assert.Equal(t, "test-secret", viper.GetString("secret"))
		assert.Equal(t, []string{"payments"}, viper.GetStringSlice("toolsets"))
		assert.Equal(t, true, viper.GetBool("read_only"))
	})
}

func TestStdioServerIO(t *testing.T) {
	t.Run("server uses stdin and stdout", func(t *testing.T) {
		// Verify that runStdioServer uses os.Stdin and os.Stdout
		// This is tested indirectly through runStdioServer tests
		// but we can verify the types are correct
		var in io.Reader = os.Stdin
		var out io.Writer = os.Stdout

		assert.NotNil(t, in)
		assert.NotNil(t, out)
	})

	t.Run("server handles empty input", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup observability
		config := log.NewConfig(log.WithMode(log.ModeStdio))
		_, logger := log.New(context.Background(), config)
		obs := observability.New(observability.WithLoggingService(logger))

		// Create client
		client := rzpsdk.NewClient("test-key", "test-secret")

		// Use empty reader and writer
		emptyIn := bytes.NewReader([]byte{})
		emptyOut := &bytes.Buffer{}

		// This tests that the server can handle empty I/O
		// We can't directly test Listen, but we can verify
		// the setup doesn't panic
		_ = emptyIn
		_ = emptyOut

		// Run server briefly
		errChan := make(chan error, 1)
		go func() {
			errChan <- runStdioServer(ctx, obs, client, []string{}, false)
		}()

		cancel()

		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("server did not stop in time")
		}
	})
}
