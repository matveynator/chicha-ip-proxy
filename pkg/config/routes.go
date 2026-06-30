// Package config contains configuration structures and parsing helpers for the proxy.
// Comments explain the reasoning to keep the package easy to navigate.
package config

import (
	"fmt"
	"net"
	"net/netip"
	"sort"
	"strconv"
	"strings"
)

// Route describes a single forwarding rule.
// Keeping it small keeps the configuration payload easy to pass across channels.
type Route struct {
	LocalPort  string // LocalPort is the port that should be opened locally.
	RemoteIP   string // RemoteIP is the target host for forwarded traffic.
	RemotePort string // RemotePort is the port on the target host.
}

// SimpleRouteFlags carries the short public CLI form for one forwarding rule.
type SimpleRouteFlags struct {
	Local  string
	Remote string
	Proto  string
}

// AllowList contains normalized client source prefixes allowed to use proxy routes.
// Empty lists are intentionally open so existing CLI commands keep their behavior.
type AllowList struct {
	Prefixes []netip.Prefix
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

// ParseSimpleRoute converts -local, -remote, and -proto into TCP or UDP route slices.
// The short form intentionally covers the common one-port setup while legacy route flags keep multi-route compatibility.
func ParseSimpleRoute(flags SimpleRouteFlags) ([]Route, []Route, bool, error) {
	hasSimpleFlag := flags.Local != "" || flags.Remote != ""
	if !hasSimpleFlag {
		return nil, nil, false, nil
	}

	if flags.Local == "" {
		return nil, nil, true, fmt.Errorf("-local is required when using -remote or -proto")
	}
	if flags.Remote == "" {
		return nil, nil, true, fmt.Errorf("-remote is required when using -local or -proto")
	}

	if err := ValidatePort(flags.Local); err != nil {
		return nil, nil, true, fmt.Errorf("invalid -local port: %v", err)
	}

	protocol := strings.ToLower(strings.TrimSpace(flags.Proto))
	if protocol == "" {
		protocol = "tcp"
	}
	if protocol != "tcp" && protocol != "udp" {
		return nil, nil, true, fmt.Errorf("-proto must be tcp or udp")
	}

	remoteIP, remotePort, err := parseRemoteTarget(flags.Remote, flags.Local)
	if err != nil {
		return nil, nil, true, err
	}

	route := Route{LocalPort: flags.Local, RemoteIP: remoteIP, RemotePort: remotePort}
	if protocol == "udp" {
		return nil, []Route{route}, true, nil
	}
	return []Route{route}, nil, true, nil
}

// ValidatePort rejects invalid TCP/UDP port strings before listeners are started.
func ValidatePort(port string) error {
	number, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("port must be a number: %v", err)
	}
	if number < 1 || number > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

// ParseAllowList accepts exact IP addresses and CIDR ranges from repeated -allow flags.
// Normalizing here lets proxy workers make fast yes/no decisions without parsing per packet.
func ParseAllowList(values []string) (AllowList, error) {
	prefixes := make([]netip.Prefix, 0, len(values))

	for _, raw := range values {
		candidates := strings.Split(raw, ",")
		for _, candidate := range candidates {
			trimmed := strings.TrimSpace(candidate)
			if trimmed == "" {
				continue
			}

			prefix, err := parseAllowPrefix(trimmed)
			if err != nil {
				return AllowList{}, err
			}
			prefixes = append(prefixes, prefix)
		}
	}

	sort.Slice(prefixes, func(i, j int) bool {
		return prefixes[i].String() < prefixes[j].String()
	})

	return AllowList{Prefixes: prefixes}, nil
}

// Allows reports whether a client IP can use the proxy.
// An empty allowlist means unrestricted access for backward compatibility.
func (allowList AllowList) Allows(addr netip.Addr) bool {
	if len(allowList.Prefixes) == 0 {
		return true
	}

	for _, prefix := range allowList.Prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

// FlagValues renders the allowlist back into repeated CLI flag values.
func (allowList AllowList) FlagValues() []string {
	values := make([]string, 0, len(allowList.Prefixes))
	for _, prefix := range allowList.Prefixes {
		if prefix.Bits() == prefix.Addr().BitLen() {
			values = append(values, prefix.Addr().String())
			continue
		}
		values = append(values, prefix.String())
	}
	return values
}

// parseAllowPrefix keeps exact IPs and CIDR ranges under one normalized representation.
func parseAllowPrefix(raw string) (netip.Prefix, error) {
	if strings.Contains(raw, "/") {
		prefix, err := netip.ParsePrefix(raw)
		if err != nil {
			return netip.Prefix{}, fmt.Errorf("invalid allow CIDR '%s': %v", raw, err)
		}
		return prefix.Masked(), nil
	}

	addr, err := netip.ParseAddr(raw)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("invalid allow IP '%s': %v", raw, err)
	}
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}

func parseRemoteTarget(remote, defaultPort string) (string, string, error) {
	trimmed := strings.TrimSpace(remote)
	if trimmed == "" {
		return "", "", fmt.Errorf("-remote cannot be empty")
	}

	if host, port, err := net.SplitHostPort(trimmed); err == nil {
		if err := validateRemoteIP(host); err != nil {
			return "", "", err
		}
		if err := ValidatePort(port); err != nil {
			return "", "", fmt.Errorf("invalid -remote port: %v", err)
		}
		return strings.Trim(host, "[]"), port, nil
	}

	if strings.Count(trimmed, ":") == 1 {
		host, port, ok := strings.Cut(trimmed, ":")
		if ok {
			if err := validateRemoteIP(host); err != nil {
				return "", "", err
			}
			if err := ValidatePort(port); err != nil {
				return "", "", fmt.Errorf("invalid -remote port: %v", err)
			}
			return host, port, nil
		}
	}

	host := strings.Trim(trimmed, "[]")
	if err := validateRemoteIP(host); err != nil {
		return "", "", err
	}
	return host, defaultPort, nil
}

func validateRemoteIP(remoteIP string) error {
	if _, err := netip.ParseAddr(remoteIP); err != nil {
		return fmt.Errorf("invalid -remote IP '%s': %v", remoteIP, err)
	}
	return nil
}
