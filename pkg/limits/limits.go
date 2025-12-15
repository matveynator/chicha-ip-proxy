// Package limits manages OS-specific resource limits to keep the proxy responsive under heavy load.
// The code favors channels over mutexes so limit adjustments remain observable and coordinated.
package limits

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type limitRequest struct {
	description string
	apply       func() error
}

// limitResult keeps the outcome of a single resource adjustment so callers can
// report both successes and failures to the console and file logs.
type limitResult struct {
	description string
	err         error
}

// SetupLimits applies platform-specific limit changes in a channel-driven pipeline.
// Using goroutines ensures each adjustment can proceed without blocking unrelated work.
func SetupLimits(logger *log.Logger) error {
	requests := collectLimitRequests(logger)
	if len(requests) == 0 {
		logger.Printf("No system limit changes required on this platform")
		log.Printf("No system limit changes required on this platform")
		return nil
	}

	requestChan := make(chan limitRequest)
	resultChan := make(chan limitResult, len(requests))

	go func() {
		defer close(resultChan)
		for req := range requestChan {
			logger.Printf("Applying system limit: %s", req.description)
			log.Printf("Applying system limit: %s", req.description)
			resultChan <- limitResult{description: req.description, err: req.apply()}
		}
	}()

	go func() {
		defer close(requestChan)
		for _, req := range requests {
			requestChan <- req
		}
	}()

	successful := make([]string, 0, len(requests))
	failures := make([]string, 0)
	var firstErr error

	for processed := 0; processed < len(requests); processed++ {
		select {
		case res := <-resultChan:
			if res.err != nil {
				entry := fmt.Sprintf("%s failed: %v", res.description, res.err)
				failures = append(failures, entry)
				if firstErr == nil {
					firstErr = fmt.Errorf("system limit adjustment failed: %w", res.err)
				}
			} else {
				successful = append(successful, res.description)
			}
		case <-time.After(5 * time.Second):
			timeoutMsg := "system limit adjustment timed out"
			failures = append(failures, timeoutMsg)
			if firstErr == nil {
				firstErr = fmt.Errorf(timeoutMsg)
			}
		}
	}

	if len(failures) == 0 {
		summary := fmt.Sprintf("System limits applied successfully: %s", strings.Join(successful, "; "))
		logger.Printf("%s", summary)
		log.Printf("%s", summary)
		return nil
	}

	logger.Printf("System limits encountered issues: %s", strings.Join(failures, "; "))
	log.Printf("System limits encountered issues: %s", strings.Join(failures, "; "))
	return firstErr
}
