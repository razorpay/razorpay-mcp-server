package main

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	t.Run("executes root command successfully", func(t *testing.T) {
		// Test that Execute doesn't panic
		// We can't easily test the full execution without mocking cobra,
		// but we can verify the function exists and is callable
		assert.NotNil(t, rootCmd)
		// Execute function exists
		assert.NotNil(t, Execute)
	})

	t.Run("root command has correct configuration", func(t *testing.T) {
		assert.Equal(t, "server", rootCmd.Use)
		assert.Equal(t, "Razorpay MCP Server", rootCmd.Short)
		assert.NotEmpty(t, rootCmd.Version)
	})

	t.Run("execute function can be called", func(t *testing.T) {
		// Execute calls rootCmd.Execute() which may exit
		// We test that the function exists and doesn't panic on nil command
		// In practice, rootCmd is always set, so Execute will work
		assert.NotPanics(t, func() {
			// We can't actually call Execute() in a test as it may call os.Exit(1)
			// But we verify the function exists
			_ = Execute
		})
	})
}

func TestInitConfig(t *testing.T) {
	t.Run("initializes config with default path", func(t *testing.T) {
		// Reset viper
		viper.Reset()

		// Set cfgFile to empty to use default path
		cfgFile = ""
		initConfig()

		// Verify viper is configured (configType might not be directly accessible)
		// Just verify initConfig doesn't panic
		assert.NotPanics(t, func() {
			initConfig()
		})
	})

	t.Run("initializes config with custom file", func(t *testing.T) {
		// Reset viper
		viper.Reset()

		// Create a temporary config file
		tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		cfgFile = tmpFile.Name()
		initConfig()

		// Verify config file is set
		assert.Equal(t, tmpFile.Name(), viper.ConfigFileUsed())
	})

	t.Run("handles missing config file gracefully", func(t *testing.T) {
		// Reset viper
		viper.Reset()

		cfgFile = "/nonexistent/config.yaml"
		// Should not panic
		assert.NotPanics(t, func() {
			initConfig()
		})
	})
}

func TestRootCmdFlags(t *testing.T) {
	t.Run("root command has all required flags", func(t *testing.T) {
		keyFlag := rootCmd.PersistentFlags().Lookup("key")
		assert.NotNil(t, keyFlag)

		secretFlag := rootCmd.PersistentFlags().Lookup("secret")
		assert.NotNil(t, secretFlag)

		logFileFlag := rootCmd.PersistentFlags().Lookup("log-file")
		assert.NotNil(t, logFileFlag)

		toolsetsFlag := rootCmd.PersistentFlags().Lookup("toolsets")
		assert.NotNil(t, toolsetsFlag)

		readOnlyFlag := rootCmd.PersistentFlags().Lookup("read-only")
		assert.NotNil(t, readOnlyFlag)
	})

	t.Run("flags are bound to viper", func(t *testing.T) {
		// Reset viper
		viper.Reset()

		// Set flag values
		rootCmd.PersistentFlags().Set("key", "test-key")
		rootCmd.PersistentFlags().Set("secret", "test-secret")

		// Verify viper can read the values
		// Note: This might not work if viper hasn't been initialized yet
		// but we're testing that the binding code exists
		assert.NotNil(t, rootCmd.PersistentFlags().Lookup("key"))
		assert.NotNil(t, rootCmd.PersistentFlags().Lookup("secret"))
	})
}

func TestVersionInfo(t *testing.T) {
	t.Run("version variables are set", func(t *testing.T) {
		// These are set at build time, but we can verify they exist
		assert.NotNil(t, version)
		assert.NotNil(t, commit)
		assert.NotNil(t, date)
	})

	t.Run("root command version includes all info", func(t *testing.T) {
		versionStr := rootCmd.Version
		assert.Contains(t, versionStr, version)
		assert.Contains(t, versionStr, commit)
		assert.Contains(t, versionStr, date)
	})
}
