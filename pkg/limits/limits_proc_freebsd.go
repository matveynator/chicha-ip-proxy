//go:build freebsd

// Package limits defines FreeBSD's RLIMIT_NPROC numeric identifier for process cap tuning.
// Encoding it here keeps the rest of the codebase decoupled from cgo or external packages.
package limits

func processLimitResource() (int, bool) {
	return 8, true
}
