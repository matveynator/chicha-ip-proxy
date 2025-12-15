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

// platformLimitRequests assembles the desired RLIMIT adjustments for OpenBSD.
// RLIMIT_DATA stands in for address space limits because RLIMIT_AS is unavailable on this platform.
func platformLimitRequests(logger *log.Logger) []limitRequest {
	desiredOpenFiles := uint64(100000)
	desiredProcesses := uint64(100000)

	requests := []limitRequest{
		buildInfinityRequestOpenBSD("data segment (rlimit_data)", syscall.RLIMIT_DATA),
		buildInfinityRequestOpenBSD("CPU time (rlimit_cpu)", syscall.RLIMIT_CPU),
		buildTargetRequestOpenBSD("open files (rlimit_files)", syscall.RLIMIT_NOFILE, desiredOpenFiles, logger),
	}

	if procResource, ok := processLimitResource(); ok {
		requests = append(requests, buildTargetRequestOpenBSD("process count (rlimit_proc)", procResource, desiredProcesses, logger))
	} else {
		logger.Printf("Process limit resource is unavailable on this platform; skipping rlimit_proc")
	}

	return requests
}

// buildInfinityRequestOpenBSD raises a resource to RLIM_INFINITY so the proxy is not capped prematurely.
// Using the computed infinity mirrors the unsigned fields exposed by the OpenBSD syscall package.
func buildInfinityRequestOpenBSD(label string, resource int) limitRequest {
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

// buildTargetRequestOpenBSD nudges a resource toward the requested target and keeps the hard limit unchanged when required.
// The fallback path maintains availability even if the kernel refuses to raise the maximum.
func buildTargetRequestOpenBSD(label string, resource int, target uint64, logger *log.Logger) limitRequest {
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
