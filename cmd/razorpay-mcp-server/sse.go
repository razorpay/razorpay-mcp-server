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

	rzpsdk "github.com/razorpay/razorpay-go/v2"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay"
)

// sseCmd starts the mcp server in sse transport mode
var sseCmd = &cobra.Command{
	Use:   "sse",
	Short: "start the sse server",
	Run: func(cmd *cobra.Command, args []string) {
		// Create stdout logger
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))

		// Get toolsets to enable from config
		enabledToolsets := viper.GetStringSlice("toolsets")

		// Get read-only mode from config
		readOnly := viper.GetBool("read_only")

		err := runSseServer(logger, nil, enabledToolsets, readOnly)
		if err != nil {
			logger.Error("error running sse server", "error", err)
			os.Exit(1)
		}
	},
}

func runSseServer(
	log *slog.Logger,
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
		log,
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
		log.Info("starting server")
		errC <- sseSrv.Start()
	}()

	log.Info("Razorpay MCP Server running on sse\n")

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		log.Info("shutting down server...")
		return nil
	case err := <-errC:
		if err != nil {
			log.Error("server error", "error", err)
			return err
		}
		return nil
	}
}
