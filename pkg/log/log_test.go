package log

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultLogPath(t *testing.T) {
	path := getDefaultLogPath()

	assert.NotEmpty(t, path, "expected non-empty path")
	assert.True(t, filepath.IsAbs(path),
		"expected absolute path, got: %s", path)
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "with empty path",
			path:     "",
			wantPath: getDefaultLogPath(),
			wantErr:  false,
		},
		{
			name:     "with specified path",
			path:     filepath.Join(os.TempDir(), "test-log-file.log"),
			wantPath: filepath.Join(os.TempDir(), "test-log-file.log"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if tt.path != "" {
					os.Remove(tt.path)
				}
				if tt.path == "" {
					os.Remove(getDefaultLogPath())
				}
			}()

			logger, cleanup, err := New(tt.path)
			assert.Equal(t, tt.wantErr, err != nil,
				"New() error = %v, wantErr %v", err, tt.wantErr)
			assert.NotNil(t, logger, "expected non-nil logger")

			cleanup()
		})
	}
}

func TestNewWithInvalidPath(t *testing.T) {
	invalidPath := "/this/path/should/not/exist/log.txt"

	logger, cleanup, err := New(invalidPath)
	assert.Nil(t, err, "New() with invalid path should not return error")
	assert.NotNil(t, logger, "expected non-nil logger even with invalid path")

	cleanup()
}
