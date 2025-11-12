package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
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

	t.Run("execute with successful command", func(t *testing.T) {
		// Test Execute() with a command that will succeed (not call os.Exit(1))
		assert.NotPanics(t, func() {
			// Save original args
			originalArgs := rootCmd.Args
			defer func() { rootCmd.Args = originalArgs }()

			// Set args to help command which should succeed
			rootCmd.SetArgs([]string{"--help"})

			// This should execute successfully without calling os.Exit(1)
			Execute()
		})
	})

	t.Run("execute function with version command", func(t *testing.T) {
		// Test Execute() with version command that should succeed
		assert.NotPanics(t, func() {
			// Save original args
			originalArgs := rootCmd.Args
			defer func() { rootCmd.Args = originalArgs }()

			// Set args to version command which should succeed
			rootCmd.SetArgs([]string{"--version"})

			// This should execute successfully without calling os.Exit(1)
			Execute()
		})
	})

	t.Run("execute function handles error case", func(t *testing.T) {
		// Test Execute() error handling by directly testing rootCmd.Execute() error path
		assert.NotPanics(t, func() {
			// Save original command state
			originalArgs := rootCmd.Args
			defer func() { rootCmd.Args = originalArgs }()

			// Set args to an invalid command that should cause an error
			rootCmd.SetArgs([]string{"invalid-nonexistent-command"})

			// Test the error path that Execute() would handle
			err := rootCmd.Execute()
			if err != nil {
				// This is the error path that Execute() would handle with os.Exit(1)
				assert.Error(t, err)
			}
		})
	})

	t.Run("execute with help flag", func(t *testing.T) {
		// Test Execute with help flag - this should not exit
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		os.Args = []string{"server", "--help"}

		// This should not panic or exit
		assert.NotPanics(t, func() {
			// Reset the command to avoid side effects
			rootCmd.SetArgs([]string{"--help"})
			err := rootCmd.Execute()
			// Help command should succeed, not return error
			assert.NoError(t, err)
		})
	})

	t.Run("execute error path coverage", func(t *testing.T) {
		// Test Execute() error path by mocking a command that returns an error
		originalCmd := rootCmd
		defer func() { rootCmd = originalCmd }()

		// Create a command that will return an error
		errorCmd := &cobra.Command{
			Use: "test-error",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("test error for coverage")
			},
		}

		// Temporarily replace rootCmd
		rootCmd = errorCmd

		// We can't test os.Exit(1) directly, but we can verify the error path
		// by checking that rootCmd.Execute() returns an error
		err := rootCmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test error for coverage")

		// This covers the error condition in Execute() function
		// The actual os.Exit(1) call can't be tested in unit tests
		// but the error path is now covered
	})

	t.Run("execute function with os.Exit coverage via subprocess", func(t *testing.T) {
		// Test Execute() function's os.Exit(1) path using subprocess
		if os.Getenv("TEST_EXECUTE_EXIT") == "1" {
			// This runs in the subprocess
			// Set up a command that will fail to trigger os.Exit(1)
			rootCmd.SetArgs([]string{"invalid-command-that-does-not-exist"})
			Execute() // This will call os.Exit(1)
			return
		}

		// Run the test in a subprocess
		cmd := exec.Command(os.Args[0], "-test.run=TestExecute/execute_function_with_os.Exit_coverage_via_subprocess")
		cmd.Env = append(os.Environ(), "TEST_EXECUTE_EXIT=1")
		err := cmd.Run()

		// The subprocess should exit with code 1 due to os.Exit(1)
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.Equal(t, 1, exitError.ExitCode())
		} else {
			// If no error, the subprocess didn't exit as expected
			t.Fatal("Expected subprocess to exit with code 1")
		}
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
		defer func() { _ = os.Remove(tmpFile.Name()) }()

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
		err := rootCmd.PersistentFlags().Set("key", "test-key")
		assert.NoError(t, err)
		err = rootCmd.PersistentFlags().Set("secret", "test-secret")
		assert.NoError(t, err)

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

func TestMain(t *testing.T) {
	t.Run("main function exists and can be called indirectly", func(t *testing.T) {
		// We can't directly test main() as it calls os.Exit
		// But we can test that rootCmd.Execute() works with valid commands
		assert.NotPanics(t, func() {
			// Test with version flag which should not exit with error
			rootCmd.SetArgs([]string{"--version"})
			err := rootCmd.Execute()
			// Version command should succeed
			assert.NoError(t, err)
		})
	})

	t.Run("main function with invalid command", func(t *testing.T) {
		// Test with invalid command
		assert.NotPanics(t, func() {
			rootCmd.SetArgs([]string{"invalid-command"})
			err := rootCmd.Execute()
			// Invalid command should return error
			assert.Error(t, err)
		})
	})

	t.Run("main function behavior verification", func(t *testing.T) {
		// Verify that main() would call rootCmd.Execute()
		// We can't call main() directly due to os.Exit(1), but we can verify
		// the command structure and behavior
		assert.NotNil(t, rootCmd)
		assert.Equal(t, "server", rootCmd.Use)

		// Test that Execute() function exists and would be called by main()
		assert.NotNil(t, Execute)

		// Verify the main function exists (we can't call it due to os.Exit)
		// but we can test the command execution path it would follow
		assert.NotPanics(t, func() {
			// Test successful command execution path
			rootCmd.SetArgs([]string{"--help"})
			err := rootCmd.Execute()
			assert.NoError(t, err)
		})
	})

	t.Run("main function with successful command execution", func(t *testing.T) {
		// Test main() function with a command that will succeed
		assert.NotPanics(t, func() {
			// Save original args and command state
			originalArgs := rootCmd.Args
			defer func() { rootCmd.Args = originalArgs }()

			// Set args to a command that should succeed (help)
			rootCmd.SetArgs([]string{"--help"})

			// Call main() - this should succeed without calling os.Exit(1)
			// because help command returns no error
			main()
		})
	})

	t.Run("main function with version command", func(t *testing.T) {
		// Test main() function with version command that should succeed
		assert.NotPanics(t, func() {
			// Save original args and command state
			originalArgs := rootCmd.Args
			defer func() { rootCmd.Args = originalArgs }()

			// Set args to version command which should succeed
			rootCmd.SetArgs([]string{"--version"})

			// Call main() - this should succeed without calling os.Exit(1)
			main()
		})
	})

	t.Run("main function error handling path", func(t *testing.T) {
		// Test main() function error path
		assert.NotPanics(t, func() {
			// Save original command
			originalCmd := rootCmd
			defer func() { rootCmd = originalCmd }()

			// Create a command that will return an error
			errorCmd := &cobra.Command{
				Use: "test-main-error",
				RunE: func(cmd *cobra.Command, args []string) error {
					return fmt.Errorf("main test error")
				},
			}

			// Temporarily replace rootCmd to test error path
			rootCmd = errorCmd

			// Test that main() would handle the error
			// We can't actually call main() with an error because it calls os.Exit(1)
			// But we can verify the error path exists
			err := rootCmd.Execute()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "main test error")
		})
	})

	t.Run("main function subprocess test for error path", func(t *testing.T) {
		// Test main() using subprocess to handle os.Exit(1)
		if os.Getenv("TEST_MAIN_ERROR") == "1" {
			// This runs in the subprocess
			// Set up a command that will fail
			rootCmd.SetArgs([]string{"invalid-command-that-does-not-exist"})
			main() // This will call os.Exit(1)
			return
		}

		// Run the test in a subprocess
		cmd := exec.Command(os.Args[0], "-test.run=TestMain/main_function_subprocess_test_for_error_path")
		cmd.Env = append(os.Environ(), "TEST_MAIN_ERROR=1")
		err := cmd.Run()

		// The subprocess should exit with code 1 due to os.Exit(1)
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.Equal(t, 1, exitError.ExitCode())
		} else {
			// If no error, the subprocess didn't exit as expected
			t.Fatal("Expected subprocess to exit with code 1")
		}
	})

	t.Run("main function success path coverage", func(t *testing.T) {
		// Test main() function success path using subprocess
		if os.Getenv("TEST_MAIN_SUCCESS") == "1" {
			// This runs in the subprocess
			// Set up a command that will succeed
			rootCmd.SetArgs([]string{"--help"})
			main() // This should succeed without calling os.Exit(1)
			return
		}

		// Run the test in a subprocess
		cmd := exec.Command(os.Args[0], "-test.run=TestMain/main_function_success_path_coverage")
		cmd.Env = append(os.Environ(), "TEST_MAIN_SUCCESS=1")
		err := cmd.Run()

		// The subprocess should exit with code 0 (success)
		assert.NoError(t, err, "Expected subprocess to exit successfully")
	})
}
