//go:build linux || darwin || freebsd || openbsd
// +build linux darwin freebsd openbsd

// Package limits wires POSIX-specific limit tuning behind a single entry point.
// Using a shared wrapper keeps only one collectLimitRequests symbol in the build
// even when platform helpers vary by operating system.
package limits

import "log"

// collectLimitRequests delegates to the platform-specific implementation.
// Keeping this wrapper separate avoids duplicate symbol definitions when
// multiple OS-specific files exist in the package tree.
func collectLimitRequests(logger *log.Logger) []limitRequest {
	return platformLimitRequests(logger)
}
