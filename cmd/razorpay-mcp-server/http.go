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

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay"
)

// httpCmd starts the mcp server in http transport mode
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "start the http server for direct JSON-RPC calls",
	Run: func(cmd *cobra.Command, args []string) {
		// Create stdout logger
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))

		// Get toolsets to enable from config
		enabledToolsets := viper.GetStringSlice("toolsets")

		// Get read-only mode from config
		readOnly := viper.GetBool("read_only")

		err := runHTTPServer(logger, nil, enabledToolsets, readOnly)
		if err != nil {
			logger.Error("error running http server", "error", err)
			os.Exit(1)
		}
	},
}

func runHTTPServer(
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

	httpSrv, err := razorpay.NewHTTPServer(
		srv,
		razorpay.NewHTTPConfig(
			razorpay.WithHTTPAddress(viper.GetString("address")),
			razorpay.WithHTTPPort(viper.GetInt("port")),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create http server: %w", err)
	}

	errC := make(chan error, 1)
	go func() {
		log.Info("starting http server")
		errC <- httpSrv.Start()
	}()

	log.Info("Razorpay MCP Server running on http\n")

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		log.Info("shutting down server...")
		return httpSrv.Shutdown(ctx)
	case err := <-errC:
		if err != nil {
			log.Error("server error", "error", err)
			return err
		}
		return nil
	}
}
