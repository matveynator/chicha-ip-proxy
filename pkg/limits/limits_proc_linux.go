//go:build linux

// Package limits embeds Linux-specific resource numbers to avoid external dependencies.
// Keeping the constant here prevents cross-platform files from importing non-standard modules.
package limits

func processLimitResource() (int, bool) {
	return 6, true
}
