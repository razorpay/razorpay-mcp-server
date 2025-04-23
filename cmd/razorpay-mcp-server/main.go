package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "version"
	commit  = "commit"
	date    = "date"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "server",
	Short:   "Razorpay MCP Server",
	Version: fmt.Sprintf("%s\ncommit %s\ndate %s", version, commit, date),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// flags will be available for all subcommands
	rootCmd.PersistentFlags().StringP("port", "p", "8080", "port to listen on")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.razorpay-mcp-server.yaml)")
	rootCmd.PersistentFlags().StringP("key", "k", "", "your razorpay api key")
	rootCmd.PersistentFlags().StringP("secret", "s", "", "your razorpay api secret")
	rootCmd.PersistentFlags().StringP("mode", "m", "test", "mode to run the payments ecosystem in (test/live)")
	rootCmd.PersistentFlags().StringP("log-file", "l", "", "path to the log file")

	// bind flags to viper env vars
	_ = viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("key", rootCmd.PersistentFlags().Lookup("key"))
	_ = viper.BindPFlag("secret", rootCmd.PersistentFlags().Lookup("secret"))
	_ = viper.BindPFlag("mode", rootCmd.PersistentFlags().Lookup("mode"))
	_ = viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))

	// subcommands
	rootCmd.AddCommand(stdioCmd)
	rootCmd.AddCommand(sseCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".razorpay-mcp-server")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
