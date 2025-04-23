package main

import (
	stdlog "log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/razorpay/razorpay-mcp-server/pkg/log"
)

// stdioCmd starts the mcp server in stdio transport mode
//
// TODO: implement the stdio server
var stdioCmd = &cobra.Command{
	Use:   "stdio",
	Short: "start the stdio server",
	Run: func(cmd *cobra.Command, args []string) {
		port := viper.GetString("port")
		logPath := viper.GetString("log-file")

		log, close, err := log.New(logPath)
		if err != nil {
			stdlog.Fatalf("create logger: %v", err)
		}
		defer close()

		log.Info("starting stdio server on port", "port", port)
	},
}
