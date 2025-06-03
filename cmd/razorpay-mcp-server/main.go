//nolint:lll
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
	rootCmd.PersistentFlags().StringP("key", "k", "", "your razorpay api key")
	rootCmd.PersistentFlags().StringP("secret", "s", "", "your razorpay api secret")
	rootCmd.PersistentFlags().StringP("log-file", "l", "", "path to the log file")
	rootCmd.PersistentFlags().StringSliceP("toolsets", "t", []string{}, "comma-separated list of toolsets to enable")
	rootCmd.PersistentFlags().Bool("read-only", false, "run server in read-only mode")
	rootCmd.PersistentFlags().StringP("address", "a", "localhost", "address to bind the sse server to")
	rootCmd.PersistentFlags().IntP("port", "p", 8080, "port to bind the sse server to")

	// bind flags to viper
	_ = viper.BindPFlag("key", rootCmd.PersistentFlags().Lookup("key"))
	_ = viper.BindPFlag("secret", rootCmd.PersistentFlags().Lookup("secret"))
	_ = viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file"))
	_ = viper.BindPFlag("toolsets", rootCmd.PersistentFlags().Lookup("toolsets"))
	_ = viper.BindPFlag("read_only", rootCmd.PersistentFlags().Lookup("read-only"))
	_ = viper.BindPFlag("address", rootCmd.PersistentFlags().Lookup("address"))
	_ = viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))

	// Set environment variable mappings
	_ = viper.BindEnv("key", "RAZORPAY_KEY_ID")        // Maps RAZORPAY_KEY_ID to key
	_ = viper.BindEnv("secret", "RAZORPAY_KEY_SECRET") // Maps RAZORPAY_KEY_SECRET to secret

	// Enable environment variable reading
	viper.AutomaticEnv()

	// subcommands
	rootCmd.AddCommand(stdioCmd)
	rootCmd.AddCommand(sseCmd)
	rootCmd.AddCommand(httpCmd)
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
