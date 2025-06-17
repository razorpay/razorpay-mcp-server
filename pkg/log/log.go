package log

import (
	"context"
	"fmt"
	"os"
)

// Logger is an interface for logging, it is used internally
// at present but has scope for external implementations
//
//nolint:interfacebloat
type Logger interface {
	Infof(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Fatalf(ctx context.Context, format string, args ...interface{})
	Debugf(ctx context.Context, format string, args ...interface{})
	Warningf(ctx context.Context, format string, args ...interface{})
	Close() error
}

// New creates a new logger based on the provided configuration.
// It returns an enhanced context and a logger implementation.
// For stdio mode, it creates a file-based slog logger.
// For sse mode, it creates a stdout-based slog logger.
func New(ctx context.Context, config *Config) (context.Context, Logger) {
	var (
		logger Logger
		err    error
	)

	switch config.GetMode() {
	case ModeStdio:
		// For stdio mode, use slog logger that writes to file
		logger, err = NewSloggerWithFile(config.GetSlogConfig().GetPath())
		if err != nil {
			fmt.Printf("failed to initialize logger\n")
			os.Exit(1)
		}
	}

	return ctx, logger
}
