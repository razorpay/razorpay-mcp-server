package log

import (
	"context"
	"fmt"
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

	// Format the message
	msg := fmt.Sprintf(format, args...)

	s.logger.LogAttrs(ctx, level, msg, attrs...)
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
func (s *slogLogger) extractContextAttrs(ctx context.Context) []slog.Attr {
	// Always include all fields as attributes
	return []slog.Attr{
		//slog.String("request_id", contextkey.RequestIDFromContext(ctx)),
		//slog.String("task_id", contextkey.TaskIDFromContext(ctx)),
		//slog.String("merchant_id", contextkey.MerchantIDFromContext(ctx)),
		//slog.String("rzp_key", contextkey.RzpKeyFromContext(ctx)),
	}
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
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: config.slog.logLevel,
		})),
	}, nil
}

// NewSloggerWithFile creates a new slog logger that writes to a file
func NewSloggerWithFile(path string) (Logger, error) {
	if path == "" {
		return NewSlogger()
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return NewSlogger() // fallback to stderr logger
	}

	// Open the log file
	f, err := os.OpenFile(
		path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return NewSlogger() // fallback to stderr logger
	}

	return &slogLogger{
		logger: slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
		closer: f.Close,
	}, nil
}
