package log

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

// getDefaultLogPath returns an absolute path for the logs directory
func getDefaultLogPath() string {
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to temp directory if we can't determine executable path
		return filepath.Join(os.TempDir(), "razorpay-mcp-server-logs")
	}

	execDir := filepath.Dir(execPath)

	return filepath.Join(execDir, "logs")
}

// New returns a new slog.Logger.
// If path to log file is not provided then
// logger uses a default path next to the executable
// If the log file cannot be opened, falls back to stderr
//
// TODO: add redaction of sensitive data
func New(path string) (*slog.Logger, func(), error) {
	if path == "" {
		path = getDefaultLogPath()
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// Fall back to stderr if we can't open the log file
		fmt.Fprintf(
			os.Stderr,
			"Warning: Failed to open log file: %v\nFalling back to stderr\n",
			err,
		)
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		noop := func() {}
		return logger, noop, nil
	}

	close := func() {
		if err := file.Close(); err != nil {
			log.Printf("close log file: %v", err)
		}
	}

	fmt.Fprintf(os.Stderr, "logs are stored in: %v\n", path)
	logger := slog.New(slog.NewTextHandler(file, nil))

	return logger, close, nil
}
