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
	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay"
)

// stdioCmd starts the mcp server in stdio transport mode
//
// TODO: implement the stdio server
var stdioCmd = &cobra.Command{
	Use:   "stdio",
	Short: "start the stdio server",
	Run: func(cmd *cobra.Command, args []string) {
		logPath := viper.GetString("razorpay_log_file")
		log, close, err := log.New(logPath)
		if err != nil {
			stdlog.Fatalf("create logger: %v", err)
		}
		defer close()

		key := viper.GetString("razorpay_key_id")
		secret := viper.GetString("razorpay_key_secret")
		client := rzpsdk.NewClient(key, secret)

		err = runStdioServer(log, client)
		if err != nil {
			log.Error("error running stdio server", "error", err)
			stdlog.Fatalf("failed to run stdio server: %v", err)
		}
	},
}

func runStdioServer(log *slog.Logger, client *rzpsdk.Client) error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	srv := razorpay.NewServer(log, client, version)
	srv.RegisterTools()

	stdioSrv, err := mcpgo.NewStdioServer(srv.GetMCPServer())
	if err != nil {
		return fmt.Errorf("failed to create stdio server: %w", err)
	}

	in, out := io.Reader(os.Stdin), io.Writer(os.Stdout)
	errC := make(chan error, 1)
	go func() {
		log.Info("starting server")
		errC <- stdioSrv.Listen(ctx, in, out)
	}()

	_, _ = fmt.Fprintf(
		os.Stderr,
		"Razorpay MCP Server running on stdio\n",
	)

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
