package main

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/log"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay"
)

// stdioCmd starts the mcp server in stdio transport mode
var stdioCmd = &cobra.Command{
	Use:   "stdio",
	Short: "start the stdio server",
	Run: func(cmd *cobra.Command, args []string) {
		logPath := viper.GetString("log_file")

		config := log.NewConfig(
			log.WithMode(log.ModeStdio),
			log.WithLogLevel(slog.LevelInfo),
			log.WithLogPath(logPath),
		)

		ctx, logger := log.New(context.Background(), config)

		// Create observability with SSE mode
		obs := observability.New(
			observability.WithLoggingService(logger),
		)

		key := viper.GetString("key")
		secret := viper.GetString("secret")
		client := rzpsdk.NewClient(key, secret)

		client.SetUserAgent("razorpay-mcp" + version + "/stdio")

		// Get toolsets to enable from config
		enabledToolsets := viper.GetStringSlice("toolsets")

		// Get read-only mode from config
		readOnly := viper.GetBool("read_only")

		err := runStdioServer(obs, client, enabledToolsets, readOnly)
		if err != nil {
			obs.Logger.Errorf(ctx,
				"error running stdio server", "error", err)
			stdlog.Fatalf("failed to run stdio server: %v", err)
		}
	},
}

func runStdioServer(
	obs *observability.Observability,
	client *rzpsdk.Client,
	enabledToolsets []string,
	readOnly bool,
) error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	srv, err := razorpay.NewServer(
		razorpay.WithObservability(obs),
		razorpay.WithClient(client),
		razorpay.WithVersion("1.0.0"),
		razorpay.WithEnabledToolsets(enabledToolsets),
		razorpay.WithReadOnly(readOnly),
	)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	srv.RegisterTools()

	stdioSrv, err := mcpgo.NewStdioServer(srv.GetMCPServer())
	if err != nil {
		return fmt.Errorf("failed to create stdio server: %w", err)
	}

	in, out := io.Reader(os.Stdin), io.Writer(os.Stdout)
	errC := make(chan error, 1)
	go func() {
		obs.Logger.Infof(ctx, "starting server")
		errC <- stdioSrv.Listen(ctx, in, out)
	}()

	_, _ = fmt.Fprintf(
		os.Stderr,
		"Razorpay MCP Server running on stdio\n",
	)

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		obs.Logger.Infof(ctx, "shutting down server...")
		return nil
	case err := <-errC:
		if err != nil {
			obs.Logger.Errorf(ctx, "server error", "error", err)
			return err
		}
		return nil
	}
}
