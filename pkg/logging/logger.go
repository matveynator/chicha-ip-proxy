// Package logging centralizes logger creation and rotation logic.
// Keeping it here avoids coupling the proxy logic with filesystem operations.
package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

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

// RotateLogs performs periodic rotation and compression.
// Running in its own goroutine keeps the rest of the application non-blocking.
func RotateLogs(logFile string, file *os.File, logger *log.Logger, frequency time.Duration) {
	for {
		time.Sleep(frequency)

		file.Close()

		rotatedFile := logFile + "." + time.Now().Format("2006-01-02")
		if err := os.Rename(logFile, rotatedFile); err != nil {
			logger.Printf("Error rotating logs: %v", err)

			newFile, err2 := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err2 != nil {
				logger.Fatalf("Failed to reopen log file after rotation error: %v", err2)
			}

			file = newFile
			logger.SetOutput(file)
			continue
		}

		newFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Fatalf("Failed to create new log file after rotation: %v", err)
		}
		file = newFile
		logger.SetOutput(file)

		logger.Println("Log file rotated successfully, now compressing old log...")

		if err := compressFile(rotatedFile); err != nil {
			logger.Printf("Error compressing rotated file: %v", err)
		} else {
			logger.Printf("Compression successful: %s.gz", rotatedFile)
			if err := os.Remove(rotatedFile); err != nil {
				logger.Printf("Error removing uncompressed rotated file: %v", err)
			}
		}
	}
}

// compressFile compresses the provided file path with gzip.
// Keeping it private avoids accidental use outside rotation.
func compressFile(filename string) error {
	original, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file for compression: %v", err)
	}
	defer original.Close()

	gzFile, err := os.OpenFile(filename+".gz", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create gz file: %v", err)
	}
	defer gzFile.Close()

	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()

	if _, err := io.Copy(gzWriter, original); err != nil {
		return fmt.Errorf("failed to copy data for compression: %v", err)
	}

	return nil
}
