package log

import (
	"fmt"
	"log"
	"log/slog"
	"os"
)

const defaultLogFilePath = "./logs"

// New returns a new slog.Logger.
// If path to log file is not provided then
// logger uses the [defaultLogFilePath]
//
// TODO: add redaction of sensitive data
func New(path string) (*slog.Logger, func(), error) {
	if path == "" {
		path = defaultLogFilePath
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}
	close := func() {
		if err := file.Close(); err != nil {
			log.Printf("close log file: %v", err)
		}
	}

	log := slog.New(slog.NewTextHandler(file, nil))

	return log, close, nil
}
