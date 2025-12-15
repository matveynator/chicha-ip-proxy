//go:build darwin
// +build darwin

// Package limits includes POSIX-specific limit tuning to mirror xinetd-like defaults on macOS.
// Using a macOS-specific file keeps type handling aligned with the unsigned Rlimit fields.
package limits

import (
	"fmt"
	"log"
	"syscall"
)

// platformLimitRequests assembles the desired RLIMIT adjustments for macOS.
// Keeping the list together documents which resources mirror the xinetd expectations.
func platformLimitRequests(logger *log.Logger) []limitRequest {
	desiredOpenFiles := uint64(100000)
	desiredProcesses := uint64(100000)

	requests := []limitRequest{
		buildInfinityRequestDarwin("virtual memory (rlimit_as)", syscall.RLIMIT_AS),
		buildInfinityRequestDarwin("CPU time (rlimit_cpu)", syscall.RLIMIT_CPU),
		buildTargetRequestDarwin("open files (rlimit_files)", syscall.RLIMIT_NOFILE, desiredOpenFiles, logger),
	}

	if procResource, ok := processLimitResource(); ok {
		requests = append(requests, buildTargetRequestDarwin("process count (rlimit_proc)", procResource, desiredProcesses, logger))
	} else {
		logger.Printf("Process limit resource is unavailable on this platform; skipping rlimit_proc")
	}

	return requests
}

// buildInfinityRequestDarwin raises a resource to the platform infinity constant.
// Using RLIM_INFINITY avoids unsafe conversions across architectures.
func buildInfinityRequestDarwin(label string, resource int) limitRequest {
	return limitRequest{
		description: fmt.Sprintf("%s -> unlimited", label),
		apply: func() error {
			current := &syscall.Rlimit{}
			if err := syscall.Getrlimit(resource, current); err != nil {
				return fmt.Errorf("failed reading %s: %w", label, err)
			}

			unlimited := ^uint64(0)
			desired := &syscall.Rlimit{Cur: unlimited, Max: unlimited}
			if current.Cur == desired.Cur && current.Max == desired.Max {
				return nil
			}

			if err := syscall.Setrlimit(resource, desired); err != nil {
				return fmt.Errorf("failed setting %s to unlimited: %w", label, err)
			}
			return nil
		},
	}
}

// buildTargetRequestDarwin nudges a resource toward the requested level while honoring the hard ceiling.
// When raising the hard limit is denied, the fallback keeps the process running with the best available values.
func buildTargetRequestDarwin(label string, resource int, target uint64, logger *log.Logger) limitRequest {
	return limitRequest{
		description: fmt.Sprintf("%s -> %d", label, target),
		apply: func() error {
			current := &syscall.Rlimit{}
			if err := syscall.Getrlimit(resource, current); err != nil {
				return fmt.Errorf("failed reading %s: %w", label, err)
			}

			desired := &syscall.Rlimit{Cur: target, Max: target}
			if current.Max > desired.Max {
				desired.Max = current.Max
			}
			if desired.Cur > desired.Max {
				desired.Cur = desired.Max
			}

			if current.Cur >= desired.Cur && current.Max >= desired.Max {
				return nil
			}

			if err := syscall.Setrlimit(resource, desired); err != nil {
				logger.Printf("Adjusting %s hit %v; trying best-effort with existing max", label, err)
				fallback := &syscall.Rlimit{Cur: desired.Cur, Max: current.Max}
				if fallback.Cur > fallback.Max {
					fallback.Cur = fallback.Max
				}
				if setErr := syscall.Setrlimit(resource, fallback); setErr != nil {
					return fmt.Errorf("failed setting %s even after fallback: %w", label, setErr)
				}
			}
			return nil
		},
	}
}
