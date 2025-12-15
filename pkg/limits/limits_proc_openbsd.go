//go:build openbsd
// +build openbsd

// Package limits includes the OpenBSD RLIMIT_NPROC value so process ceilings can be raised at startup.
// Hardcoding the numeric value keeps the build pure Go without extra dependencies.
package limits

func processLimitResource() (int, bool) {
	return 7, true
}
