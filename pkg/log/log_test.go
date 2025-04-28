package log

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultLogPath(t *testing.T) {
	// Call the function
	path := getDefaultLogPath()

	// Verify the path is not empty
	if path == "" {
		t.Error("expected non-empty path, got empty string")
	}

	// Check that the path is absolute
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got: %s", path)
	}
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
			// Clean up any test log files
			defer func() {
				if tt.path != "" {
					os.Remove(tt.path)
				}
				if tt.path == "" {
					os.Remove(getDefaultLogPath())
				}
			}()

			logger, cleanup, err := New(tt.path)

			// Check error return
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Ensure logger is not nil
			if logger == nil {
				t.Error("expected non-nil logger")
				return
			}

			// Clean up log file
			cleanup()
		})
	}
}

func TestNewWithInvalidPath(t *testing.T) {
	// Use a path that should not be writable
	invalidPath := "/this/path/should/not/exist/log.txt"

	logger, cleanup, err := New(invalidPath)

	// Should not error, as it falls back to stderr
	if err != nil {
		t.Errorf("New() with invalid path should not return error, got: %v", err)
	}

	// Logger should still be created (stderr fallback)
	if logger == nil {
		t.Error("expected non-nil logger even with invalid path")
	}

	// Cleanup should be a noop, but we call it anyway
	cleanup()
}
