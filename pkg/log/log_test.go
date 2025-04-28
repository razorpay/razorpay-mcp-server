package log

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultLogPath(t *testing.T) {
	path := getDefaultLogPath()

	if path == "" {
		t.Error("expected non-empty path, got empty string")
	}

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
			defer func() {
				if tt.path != "" {
					os.Remove(tt.path)
				}
				if tt.path == "" {
					os.Remove(getDefaultLogPath())
				}
			}()

			logger, cleanup, err := New(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if logger == nil {
				t.Error("expected non-nil logger")
				return
			}

			cleanup()
		})
	}
}

func TestNewWithInvalidPath(t *testing.T) {
	invalidPath := "/this/path/should/not/exist/log.txt"

	logger, cleanup, err := New(invalidPath)
	if err != nil {
		t.Errorf("New() with invalid path should not return error, got: %v", err)
	}

	if logger == nil {
		t.Error("expected non-nil logger even with invalid path")
	}

	cleanup()
}
