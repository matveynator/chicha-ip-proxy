//go:build darwin || freebsd
// +build darwin freebsd

// Package limits includes POSIX-specific limit tuning to mirror xinetd-like defaults on BSD-derived systems.
// Using a BSD-focused file keeps type handling compatible with the int64-based Rlimit definitions.
package limits

import (
	"fmt"
	"log"
	"syscall"
)

// collectLimitRequests assembles the desired RLIMIT adjustments for macOS and FreeBSD.
// Keeping the list together documents which resources mirror the xinetd expectations.
func collectLimitRequests(logger *log.Logger) []limitRequest {
	desiredOpenFiles := int64(100000)
	desiredProcesses := int64(100000)

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

// buildInfinityRequest raises a resource to RLIM_INFINITY so workloads are not capped unexpectedly.
// Using the constant directly avoids unsafe conversions when syscall uses int64 fields.
func buildInfinityRequest(label string, resource int) limitRequest {
	return limitRequest{
		description: fmt.Sprintf("%s -> unlimited", label),
		apply: func() error {
			current := &syscall.Rlimit{}
			if err := syscall.Getrlimit(resource, current); err != nil {
				return fmt.Errorf("failed reading %s: %w", label, err)
			}

			desired := &syscall.Rlimit{Cur: syscall.RLIM_INFINITY, Max: syscall.RLIM_INFINITY}
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

// buildTargetRequest nudges a resource toward the requested level while honoring the hard ceiling.
// When raising the hard limit is denied, the fallback keeps the process running with the best available values.
func buildTargetRequest(label string, resource int, target int64, logger *log.Logger) limitRequest {
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
