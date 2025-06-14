package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-mcp-server/pkg/log"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay"
)

// sseCmd starts the mcp server in sse transport mode
var sseCmd = &cobra.Command{
	Use:   "sse",
	Short: "start the sse server",
	Run: func(cmd *cobra.Command, args []string) {
		config := log.NewConfig(
			log.WithMode(log.ModeSSE),
			log.WithLogLevel(slog.LevelInfo),
		)

		ctx, logger := log.New(context.Background(), config)

		// Create observability with SSE mode
		obs := observability.New(
			observability.WithLoggingService(logger),
		)

		// Get toolsets to enable from config
		enabledToolsets := viper.GetStringSlice("toolsets")

		// Get read-only mode from config
		readOnly := viper.GetBool("read_only")

		err := runSseServer(obs, nil, enabledToolsets, readOnly)
		if err != nil {
			obs.Logger.Errorf(ctx, "error running sse server", "error", err)
			os.Exit(1)
		}
	},
}

func runSseServer(
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
		obs,
		client,
		"1.0.0",
		enabledToolsets,
		readOnly,
	)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	srv.RegisterTools()

	sseSrv, err := mcpgo.NewSSEServer(
		srv.GetMCPServer(),
		mcpgo.NewSSEConfig(
			mcpgo.WithSSEAddress(viper.GetString("address")),
			mcpgo.WithSSEPort(viper.GetInt("port")),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create sse server: %w", err)
	}

	errC := make(chan error, 1)
	go func() {
		obs.Logger.Infof(ctx, "starting server")
		errC <- sseSrv.Start()
	}()

	obs.Logger.Infof(ctx, "Razorpay MCP Server running on sse")

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
