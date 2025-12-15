// Package version centralizes version resolution for the proxy binary.
// The version string is injected at build time using the sequential Git commit count.
package version

import (
	"errors"
	"os/exec"
	"strings"
)

// Number carries the build version injected via -ldflags.
// The default value stays as "dev" so local builds without Git fallback remain explicit.
var Number = "dev"

// Resolve returns the best available version string to display in CLI output and logs.
// It prefers the compile-time injected Number and falls back to the repository commit count when available.
func Resolve() string {
	if Number != "" && Number != "dev" {
		return Number
	}

	commitCount, err := commitCount()
	if err == nil && commitCount != "" {
		return commitCount
	}

	return "dev"
}

// commitCount runs a small Git command to obtain the incremental commit number.
// Using the CLI keeps dependencies minimal while still surfacing a useful build identifier.
func commitCount() (string, error) {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	cleaned := strings.TrimSpace(string(output))
	if cleaned == "" {
		return "", errors.New("empty commit count")
	}

	return cleaned, nil
}
