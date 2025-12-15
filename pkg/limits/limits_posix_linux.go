//go:build linux
// +build linux

// Package limits includes POSIX-specific limit tuning to mirror xinetd-like defaults on Linux.
// Using a Linux-specific file keeps type handling aligned with the platform's syscall definitions.
package limits

import (
	"fmt"
	"log"
	"syscall"
)

// collectLimitRequests assembles the desired RLIMIT adjustments for Linux.
// Grouping them in one place mirrors xinetd defaults while keeping call sites small.
func collectLimitRequests(logger *log.Logger) []limitRequest {
	desiredOpenFiles := uint64(100000)
	desiredProcesses := uint64(100000)

	requests := []limitRequest{
		buildInfinityRequest("virtual memory (rlimit_as)", syscall.RLIMIT_AS),
		buildInfinityRequest("CPU time (rlimit_cpu)", syscall.RLIMIT_CPU),
		buildTargetRequest("open files (rlimit_files)", syscall.RLIMIT_NOFILE, desiredOpenFiles, logger),
	}

	if procResource, ok := processLimitResource(); ok {
		requests = append(requests, buildTargetRequest("process count (rlimit_proc)", procResource, desiredProcesses, logger))
	} else {
		logger.Printf("Process limit resource is unavailable on this platform; skipping rlimit_proc")
	}

	return requests
}

// buildInfinityRequest raises a resource to the platform infinity constant.
// Using RLIM_INFINITY avoids unsafe conversions across architectures.
func buildInfinityRequest(label string, resource int) limitRequest {
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

// buildTargetRequest nudges a resource toward the target while respecting current maxima.
// The fallback keeps the process running even when the hard ceiling cannot be raised.
func buildTargetRequest(label string, resource int, target uint64, logger *log.Logger) limitRequest {
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
