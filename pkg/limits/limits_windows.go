//go:build windows
// +build windows

// Package limits documents the lack of configurable RLIMIT-style knobs on Windows.
// The stub still runs through the channel pipeline so the caller sees consistent behavior.
package limits

import "log"

func collectLimitRequests(logger *log.Logger) []limitRequest {
	logger.Printf("Windows relies on dynamic kernel limits; no explicit RLIMIT tuning applied")
	return nil
}
