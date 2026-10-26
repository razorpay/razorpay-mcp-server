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
	"github.com/razorpay/razorpay-mcp-server/pkg/region"
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

		client, profile, err := buildStdioDeps(
			viper.GetString("key"),
			viper.GetString("secret"),
			viper.GetString("razorpay_region"),
			viper.GetString("razorpay_region_config"),
		)
		if err != nil {
			stdlog.Fatalf("failed to load region profile: %v", err)
		}
		ctx = region.WithProfile(ctx, profile)

		// Get toolsets to enable from config
		enabledToolsets := viper.GetStringSlice("toolsets")

		// Get read-only mode from config
		readOnly := viper.GetBool("read_only")

		err = runStdioServer(ctx, obs, client, enabledToolsets, readOnly)
		if err != nil {
			obs.Logger.Errorf(ctx,
				"error running stdio server", "error", err)
			stdlog.Fatalf("failed to run stdio server: %v", err)
		}
	},
}

// buildStdioDeps resolves the region profile and creates an API client that
// talks to that region's Razorpay API host.
func buildStdioDeps(
	key, secret, regionCode, regionConfigPath string,
) (*rzpsdk.Client, region.Profile, error) {
	profiles, err := region.Load(regionConfigPath)
	if err != nil {
		return nil, region.Profile{}, err
	}
	profile, err := region.Get(profiles, regionCode)
	if err != nil {
		return nil, region.Profile{}, err
	}

	client := rzpsdk.NewClient(key, secret)
	client.SetUserAgent("razorpay-mcp" + version + "/stdio")
	client.Request.BaseURL = profile.APIBaseURL
	return client, profile, nil
}

// runStdioServer serves MCP over stdio for the region profile in ctx.
func runStdioServer(
	ctx context.Context,
	obs *observability.Observability,
	client *rzpsdk.Client,
	enabledToolsets []string,
	readOnly bool,
) error {
	ctx, stop := signal.NotifyContext(
		ctx,
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	profile := region.FromContext(ctx)
	srv, err := razorpay.NewRzpMcpServerForRegion(
		obs, client, profile, enabledToolsets, readOnly)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	stdioSrv, err := mcpgo.NewStdioServer(srv)
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
		"Razorpay MCP Server running on stdio (region %s)\n",
		profile.CountryCode,
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
