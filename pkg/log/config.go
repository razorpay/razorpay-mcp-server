package log

import (
	"log/slog"
)

// Logger modes
const (
	ModeStdio = "stdio"
)

// Config holds logger configuration with options pattern.
// Use NewConfig to create a new configuration with default values,
// then customize it using the With* option functions.
type Config struct {
	// mode determines the logger type (stdio or sse)
	mode string
	// Embedded configs for different logger types
	slog slogConfig
}

// slogConfig holds slog-specific configuration for stdio mode
type slogConfig struct {
	// path is the file path where logs will be written
	path string
	// logLevel sets the minimum log level to output
	logLevel slog.Leveler
}

// GetMode returns the logger mode (stdio or sse)
func (c Config) GetMode() string {
	return c.mode
}

// GetSlogConfig returns the slog logger configuration
func (c Config) GetSlogConfig() slogConfig {
	return c.slog
}

// GetLogLevel returns the log level
func (z Config) GetLogLevel() slog.Leveler {
	return z.slog.logLevel
}

// GetPath returns the log file path
func (s slogConfig) GetPath() string {
	return s.path
}

// ConfigOption represents a configuration option function
type ConfigOption func(*Config)

// WithMode sets the logger mode (stdio or sse)
func WithMode(mode string) ConfigOption {
	return func(c *Config) {
		c.mode = mode
	}
}

// WithLogPath sets the log file path
func WithLogPath(path string) ConfigOption {
	return func(c *Config) {
		c.slog.path = path
	}
}

// WithLogLevel sets the log level for the mode
func WithLogLevel(level slog.Level) ConfigOption {
	return func(c *Config) {
		c.slog.logLevel = level
	}
}

// NewConfig creates a new config with default values.
// By default, it uses stdio mode with info log level.
// Use With* options to customize the configuration.
func NewConfig(opts ...ConfigOption) *Config {
	config := &Config{
		mode: ModeStdio,
		slog: slogConfig{
			logLevel: slog.LevelInfo,
		},
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
