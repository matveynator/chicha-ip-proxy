//go:build openbsd
// +build openbsd

// Package limits includes POSIX-specific limit tuning to mirror xinetd-like defaults on OpenBSD.
// OpenBSD omits RLIMIT_AS, so the code targets RLIMIT_DATA to keep memory ceilings generous.
package limits

import (
	"fmt"
	"log"
	"syscall"
)

// collectLimitRequests assembles the desired RLIMIT adjustments for OpenBSD.
// RLIMIT_DATA stands in for address space limits because RLIMIT_AS is unavailable on this platform.
func collectLimitRequests(logger *log.Logger) []limitRequest {
	desiredOpenFiles := int64(100000)
	desiredProcesses := int64(100000)

	requests := []limitRequest{
		buildInfinityRequest("data segment (rlimit_data)", syscall.RLIMIT_DATA),
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

// buildInfinityRequest raises a resource to RLIM_INFINITY so the proxy is not capped prematurely.
// Using the constant directly matches the int64 fields exposed by the OpenBSD syscall package.
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

// buildTargetRequest nudges a resource toward the requested target and keeps the hard limit unchanged when required.
// The fallback path maintains availability even if the kernel refuses to raise the maximum.
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
