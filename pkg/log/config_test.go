package log

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogLevel(t *testing.T) {
	t.Run("returns log level from config", func(t *testing.T) {
		config := NewConfig(WithLogLevel(slog.LevelDebug))
		level := config.GetLogLevel()
		assert.Equal(t, slog.LevelDebug, level)
	})

	t.Run("returns default log level", func(t *testing.T) {
		config := NewConfig()
		level := config.GetLogLevel()
		assert.Equal(t, slog.LevelInfo, level)
	})

	t.Run("returns custom log level", func(t *testing.T) {
		config := NewConfig(WithLogLevel(slog.LevelWarn))
		level := config.GetLogLevel()
		assert.Equal(t, slog.LevelWarn, level)
	})
}

func TestWithLogLevel(t *testing.T) {
	t.Run("sets log level in config", func(t *testing.T) {
		config := NewConfig(WithLogLevel(slog.LevelDebug))
		assert.Equal(t, slog.LevelDebug, config.GetLogLevel())
	})

	t.Run("sets error log level", func(t *testing.T) {
		config := NewConfig(WithLogLevel(slog.LevelError))
		assert.Equal(t, slog.LevelError, config.GetLogLevel())
	})

	t.Run("overwrites previous log level", func(t *testing.T) {
		config := NewConfig(
			WithLogLevel(slog.LevelDebug),
			WithLogLevel(slog.LevelWarn),
		)
		assert.Equal(t, slog.LevelWarn, config.GetLogLevel())
	})
}
