package log

import (
	"fmt"
	"log"
	"log/slog"
	"os"
)

// New returns a new slog.Logger.
//
// If path is empty, it will log to stdout.
// Otherwise, it will log to the file at the given path.
//
// TODO: add redaction of sensitive data
func New(path string) (*slog.Logger, func(), error) {
	close := func() {}
	if path == "" {
		return slog.New(slog.NewTextHandler(os.Stdout, nil)), close, nil
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}
	close = func() {
		if err := file.Close(); err != nil {
			log.Printf("close log file: %v", err)
		}
	}

	log := slog.New(slog.NewTextHandler(file, nil))

	return log, close, nil
}
