package main

import (
	stdlog "log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/razorpay/razorpay-mcp-server/pkg/log"
)

// sseCmd starts the mcp server in sse transport mode
//
// TODO: implement the sse server
var sseCmd = &cobra.Command{
	Use:   "sse",
	Short: "start the sse server",
	Run: func(cmd *cobra.Command, args []string) {
		port := viper.GetString("port")
		logPath := viper.GetString("log-file")

		log, close, err := log.New(logPath)
		if err != nil {
			stdlog.Fatalf("create logger: %v", err)
		}
		defer close()

		log.Info("starting sse server on port", "port", port)
	},
}
