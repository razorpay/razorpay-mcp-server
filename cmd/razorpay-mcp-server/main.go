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

// Execute runs the root command and handles any errors
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// flags will be available for all subcommands
	rootCmd.PersistentFlags().StringP("key", "k", "", "your razorpay api key")       //nolint:lll
	rootCmd.PersistentFlags().StringP("secret", "s", "", "your razorpay api secret") //nolint:lll
	rootCmd.PersistentFlags().StringP("log-file", "l", "", "path to the log file")   //nolint:lll

	// bind flags to viper env vars
	_ = viper.BindPFlag("razorpay_key_id", rootCmd.PersistentFlags().Lookup("key"))
	_ = viper.BindPFlag("razorpay_key_secret", rootCmd.PersistentFlags().Lookup("secret")) //nolint:lll
	_ = viper.BindPFlag("razorpay_log_file", rootCmd.PersistentFlags().Lookup("log-file")) //nolint:lll

	// subcommands
	rootCmd.AddCommand(stdioCmd)
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
