package log

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

// slogLogger implements Logger interface using slog
type slogLogger struct {
	logger *slog.Logger
	closer func() error
}

// logWithLevel is a helper function that handles common logging functionality
func (s *slogLogger) logWithLevel(
	ctx context.Context,
	level slog.Level,
	format string,
	args ...interface{},
) {
	// Extract context fields and add them as slog attributes
	attrs := s.extractContextAttrs(ctx)

	// Convert args to slog attributes
	attrs = append(attrs, s.convertArgsToAttrs(args...)...)

	s.logger.LogAttrs(ctx, level, format, attrs...)
}

// Infof logs an info message with context fields
func (s *slogLogger) Infof(
	ctx context.Context, format string, args ...interface{}) {
	s.logWithLevel(ctx, slog.LevelInfo, format, args...)
}

// Errorf logs an error message with context fields
func (s *slogLogger) Errorf(
	ctx context.Context, format string, args ...interface{}) {
	s.logWithLevel(ctx, slog.LevelError, format, args...)
}

// Fatalf logs a fatal message with context fields and exits
func (s *slogLogger) Fatalf(
	ctx context.Context, format string, args ...interface{}) {
	s.logWithLevel(ctx, slog.LevelError, format, args...)
	os.Exit(1)
}

// Debugf logs a debug message with context fields
func (s *slogLogger) Debugf(
	ctx context.Context, format string, args ...interface{}) {
	s.logWithLevel(ctx, slog.LevelDebug, format, args...)
}

// Warningf logs a warning message with context fields
func (s *slogLogger) Warningf(
	ctx context.Context, format string, args ...interface{}) {
	s.logWithLevel(ctx, slog.LevelWarn, format, args...)
}

// extractContextAttrs extracts fields from context and converts to slog.Attr
func (s *slogLogger) extractContextAttrs(_ context.Context) []slog.Attr {
	// Always include all fields as attributes
	return []slog.Attr{}
}

// convertArgsToAttrs converts key-value pairs to slog.Attr
func (s *slogLogger) convertArgsToAttrs(args ...interface{}) []slog.Attr {
	if len(args) == 0 {
		return nil
	}

	var attrs []slog.Attr
	for i := 0; i < len(args)-1; i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if !ok {
				continue
			}
			value := args[i+1]
			attrs = append(attrs, slog.Any(key, value))
		}
	}
	return attrs
}

// Close implements the Logger interface Close method
func (s *slogLogger) Close() error {
	if s.closer != nil {
		return s.closer()
	}
	return nil
}

// NewSlogger returns a new slog.Logger implementation of Logger interface.
// If path to log file is not provided then logger uses stderr for stdio mode
// If the log file cannot be opened, falls back to stderr
func NewSlogger() (*slogLogger, error) {
	// For stdio mode, always use stderr regardless of path
	// This ensures logs don't interfere with MCP protocol on stdout
	return &slogLogger{
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}, nil
}

func NewSloggerWithStdout(config *Config) (*slogLogger, error) {
	// For stdio mode, always use Stdout regardless of path
	// This ensures logs don't interfere with MCP protocol on stdout
	return &slogLogger{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: config.slog.logLevel,
		})),
	}, nil
}

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

// NewSloggerWithFile returns a new slog.Logger.
// If path to log file is not provided then
// logger uses a default path next to the executable
// If the log file cannot be opened, falls back to stderr
//
// TODO: add redaction of sensitive data
func NewSloggerWithFile(path string) (*slogLogger, error) {
	if path == "" {
		path = getDefaultLogPath()
	}

	// #nosec G304 - path is validated and comes from configuration
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		// Fall back to stderr if we can't open the log file
		fmt.Fprintf(
			os.Stderr,
			"Warning: Failed to open log file: %v\nFalling back to stderr\n",
			err,
		)
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		noop := func() error { return nil }
		return &slogLogger{
			logger: logger,
			closer: noop,
		}, nil
	}

	fmt.Fprintf(os.Stderr, "logs are stored in: %v\n", path)
	return &slogLogger{
		logger: slog.New(slog.NewTextHandler(file, nil)),
		closer: func() error {
			if err := file.Close(); err != nil {
				log.Printf("close log file: %v", err)
			}

			return nil
		},
	}, nil
}
