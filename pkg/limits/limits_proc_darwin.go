//go:build darwin
// +build darwin

// Package limits records the Darwin RLIMIT_NPROC identifier without pulling extra headers.
// Explicit constants keep cross-compilation predictable for release automation.
package limits

func processLimitResource() (int, bool) {
	return 7, true
}
