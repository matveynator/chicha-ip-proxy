// Package logging centralizes logger creation and rotation logic.
// Keeping it here avoids coupling the proxy logic with filesystem operations.
package logging

import (
	"fmt"
	"log"
	"os"
	"time"
)

// DefaultMaxSizeBytes defines the size threshold for log rotation.
// Keeping it exported lets the caller opt into consistent sizing without redefining the constant.
const DefaultMaxSizeBytes int64 = 100 * 1024 * 1024

// SetupLogger opens the target file and returns a standard logger alongside the underlying file handle.
// Returning the file lets the caller manage its lifecycle without hidden global state.
func SetupLogger(logFile string) (*log.Logger, *os.File, error) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open log file '%s': %v", logFile, err)
	}

	logger := log.New(file, "", log.LstdFlags)
	return logger, file, nil
}

// RotateLogs performs periodic rotation and keeps the logs uncompressed.
// Running in its own goroutine keeps the rest of the application non-blocking.
func RotateLogs(logFile string, file *os.File, logger *log.Logger, frequency time.Duration, maxSizeBytes int64) {
	if maxSizeBytes <= 0 {
		maxSizeBytes = DefaultMaxSizeBytes
	}

	rotationTicker := time.NewTicker(frequency)
	sizeTicker := time.NewTicker(time.Minute)
	defer rotationTicker.Stop()
	defer sizeTicker.Stop()

	currentFile := file

	for {
		select {
		case <-rotationTicker.C:
			nextFile, err := rotateOnce(logFile, currentFile, logger)
			if err == nil {
				currentFile = nextFile
			}

		case <-sizeTicker.C:
			info, err := currentFile.Stat()
			if err != nil {
				logger.Printf("Error stating log file for rotation: %v", err)
				continue
			}

			if info.Size() >= maxSizeBytes {
				nextFile, err := rotateOnce(logFile, currentFile, logger)
				if err == nil {
					currentFile = nextFile
				}
			}
		}
	}
}

// rotateOnce handles closing, renaming, and reopening the log file without compression.
// Returning the newly opened file keeps the caller in control of the active handle while
// leaving the rotated file intact for external tools that may prefer raw text.
func rotateOnce(logFile string, currentFile *os.File, logger *log.Logger) (*os.File, error) {
	if err := currentFile.Sync(); err != nil {
		logger.Printf("Error syncing log file before rotation: %v", err)
	}
	if err := currentFile.Close(); err != nil {
		logger.Printf("Error closing log file before rotation: %v", err)
	}

	rotatedFile := logFile + "." + time.Now().Format("2006-01-02")
	if err := os.Rename(logFile, rotatedFile); err != nil {
		logger.Printf("Error rotating logs: %v", err)

		reopened, reopenErr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if reopenErr != nil {
			logger.Fatalf("Failed to reopen log file after rotation error: %v", reopenErr)
			return nil, reopenErr
		}

		logger.SetOutput(reopened)
		return reopened, err
	}

	newFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Fatalf("Failed to create new log file after rotation: %v", err)
		return nil, err
	}
	logger.SetOutput(newFile)
	logger.Println("Log file rotated successfully; compression skipped to keep raw text accessible.")
	return newFile, nil
}
