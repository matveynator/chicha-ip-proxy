// Package config contains configuration structures and parsing helpers for the proxy.
// Comments explain the reasoning to keep the package easy to navigate.
package config

import (
	"fmt"
	"strings"
)

// Route describes a single forwarding rule.
// Keeping it small keeps the configuration payload easy to pass across channels.
type Route struct {
	LocalPort  string // LocalPort is the port that should be opened locally.
	RemoteIP   string // RemoteIP is the target host for forwarded traffic.
	RemotePort string // RemotePort is the port on the target host.
}

// ParseRoutes splits a flag string in the form LOCALPORT:REMOTEIP:REMOTEPORT into Route values.
// Returning a slice keeps the main package free from parsing details while following Go's preference for simple data flows.
func ParseRoutes(routesFlag string) ([]Route, error) {
	if routesFlag == "" {
		return nil, nil
	}

	parts := strings.Split(routesFlag, ",")
	routes := make([]Route, 0, len(parts))

	for _, part := range parts {
		segments := strings.Split(part, ":")
		if len(segments) != 3 {
			return nil, fmt.Errorf("invalid route format: '%s' (expected LOCALPORT:REMOTEIP:REMOTEPORT)", part)
		}

		routes = append(routes, Route{
			LocalPort:  segments[0],
			RemoteIP:   segments[1],
			RemotePort: segments[2],
		})
	}

	return routes, nil
}
