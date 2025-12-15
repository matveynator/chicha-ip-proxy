// Package limits manages OS-specific resource limits to keep the proxy responsive under heavy load.
// The code favors channels over mutexes so limit adjustments remain observable and coordinated.
package limits

import (
	"fmt"
	"log"
	"time"
)

type limitRequest struct {
	description string
	apply       func() error
}

// SetupLimits applies platform-specific limit changes in a channel-driven pipeline.
// Using goroutines ensures each adjustment can proceed without blocking unrelated work.
func SetupLimits(logger *log.Logger) error {
	requests := collectLimitRequests(logger)
	if len(requests) == 0 {
		logger.Printf("No system limit changes required on this platform")
		return nil
	}

	requestChan := make(chan limitRequest)
	resultChan := make(chan error)

	go func() {
		defer close(resultChan)
		for req := range requestChan {
			logger.Printf("Applying system limit: %s", req.description)
			resultChan <- req.apply()
		}
	}()

	go func() {
		defer close(requestChan)
		for _, req := range requests {
			requestChan <- req
		}
	}()

	for processed := 0; processed < len(requests); processed++ {
		select {
		case err := <-resultChan:
			if err != nil {
				return fmt.Errorf("system limit adjustment failed: %w", err)
			}
		case <-time.After(5 * time.Second):
			return fmt.Errorf("timeout while applying system limits")
		}
	}

	return nil
}
