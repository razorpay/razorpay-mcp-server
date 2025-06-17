package log

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultLogPath(t *testing.T) {
	path := getDefaultLogPath()

	assert.NotEmpty(t, path, "expected non-empty path")
	assert.True(t, filepath.IsAbs(path),
		"expected absolute path, got: %s", path)
}

func TestNewSlogger(t *testing.T) {
	logger, err := NewSlogger()
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Test Close
	err = logger.Close()
	assert.NoError(t, err)
}

func TestNewSloggerWithFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "with empty path",
			path:    "",
			wantErr: false,
		},
		{
			name:    "with valid path",
			path:    filepath.Join(os.TempDir(), "test-log-file.log"),
			wantErr: false,
		},
		{
			name:    "with invalid path",
			path:    "/this/path/should/not/exist/log.txt",
			wantErr: false, // Should fallback to stderr
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up test file after test
			if tt.path != "" {
				defer os.Remove(tt.path)
			}

			logger, err := NewSloggerWithFile(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, logger)

			// Test logging
			ctx := context.Background()
			logger.Infof(ctx, "test message")
			logger.Debugf(ctx, "test debug")
			logger.Warningf(ctx, "test warning")
			logger.Errorf(ctx, "test error")

			// Test Close
			err = logger.Close()
			assert.NoError(t, err)

			// Verify file was created if path was specified
			if tt.path != "" && tt.path != "/this/path/should/not/exist/log.txt" {
				_, err := os.Stat(tt.path)
				assert.NoError(t, err, "log file should exist")
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "stdio mode",
			config: NewConfig(
				WithMode(ModeStdio),
				WithLogPath(""),
			),
		},
		{
			name:   "default mode",
			config: NewConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx, logger := New(ctx, tt.config)

			require.NotNil(t, newCtx)
			require.NotNil(t, logger)

			// Test logging
			logger.Infof(ctx, "test message")
			logger.Debugf(ctx, "test debug")
			logger.Warningf(ctx, "test warning")
			logger.Errorf(ctx, "test error")

			// Test Close
			err := logger.Close()
			assert.NoError(t, err)
		})
	}
}
