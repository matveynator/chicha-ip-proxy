// Package setup hosts interactive bootstrap helpers for first-run configuration.
// Grouping prompts together keeps the main package focused on runtime wiring while still following Go's advice to communicate via channels.
package setup

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/matveynator/chicha-ip-proxy/pkg/config"
)

const (
	greenText  = "\033[32m"
	purpleText = "\033[35m"
	resetText  = "\033[0m"
)

// InteractiveResult carries user-provided routes and derived metadata such as log and service names.
// Keeping it small makes it easy to hand over to the main package without introducing global state.
type InteractiveResult struct {
	TCPRoutes     []config.Route
	UDPRoutes     []config.Route
	LogFile       string
	ServiceName   string
	RoutesFlag    string
	UDPRoutesFlag string
}

// RunInteractiveSetup asks the operator for target IP, protocols, and ports when flags are absent.
// Returning the computed routes keeps follow-up automation straightforward.
func RunInteractiveSetup(appName string) (*InteractiveResult, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(colorize(purpleText, "Interactive setup (Linux only)"))
	fmt.Println(colorize(greenText, "We will ask for the destination IP and ports. Press Enter to confirm your choice."))
	fmt.Println(colorize(greenText, "Note: startup will tune system limits to keep the proxy fast."))

	fmt.Print(colorize(greenText, "Target IP (where traffic should go): "))
	targetIP, err := readTrimmed(reader)
	if err != nil {
		return nil, err
	}
	if targetIP == "" {
		return nil, fmt.Errorf("target IP cannot be empty")
	}

	fmt.Print(colorize(greenText, "Protocols to handle (tcp, udp). Use commas, e.g., tcp,udp: "))
	protocolsRaw, err := readTrimmed(reader)
	if err != nil {
		return nil, err
	}
	protocols := splitAndClean(protocolsRaw)
	if len(protocols) == 0 {
		return nil, fmt.Errorf("at least one protocol is required")
	}

	tcpRoutes := make([]config.Route, 0)
	udpRoutes := make([]config.Route, 0)

	for _, proto := range protocols {
		switch strings.ToLower(proto) {
		case "tcp":
			fmt.Print(colorize(purpleText, "Local TCP ports (comma separated, e.g., 8080,8443): "))
			portsRaw, err := readTrimmed(reader)
			if err != nil {
				return nil, err
			}
			ports := splitAndClean(portsRaw)
			for _, port := range ports {
				remotePort, err := askRemotePort(reader, port)
				if err != nil {
					return nil, err
				}
				tcpRoutes = append(tcpRoutes, config.Route{LocalPort: port, RemoteIP: targetIP, RemotePort: remotePort})
			}
		case "udp":
			fmt.Print(colorize(purpleText, "Local UDP ports (comma separated, e.g., 5353,6000): "))
			portsRaw, err := readTrimmed(reader)
			if err != nil {
				return nil, err
			}
			ports := splitAndClean(portsRaw)
			for _, port := range ports {
				remotePort, err := askRemotePort(reader, port)
				if err != nil {
					return nil, err
				}
				udpRoutes = append(udpRoutes, config.Route{LocalPort: port, RemoteIP: targetIP, RemotePort: remotePort})
			}
		default:
			return nil, fmt.Errorf("unsupported protocol: %s", proto)
		}
	}

	identifierParts := buildIdentifier(protocols, tcpRoutes, udpRoutes)
	identifier := strings.Join(identifierParts, "-")

	result := &InteractiveResult{
		TCPRoutes:     tcpRoutes,
		UDPRoutes:     udpRoutes,
		LogFile:       fmt.Sprintf("/var/log/%s-%s.log", appName, identifier),
		ServiceName:   fmt.Sprintf("%s-%s.service", appName, identifier),
		RoutesFlag:    routesFlagValue(tcpRoutes),
		UDPRoutesFlag: routesFlagValue(udpRoutes),
	}

	fmt.Println(colorize(purpleText, "Planned paths:"))
	fmt.Printf(colorize(greenText, "  Log file: %s\n"), result.LogFile)
	fmt.Printf(colorize(greenText, "  Systemd service name: %s\n"), result.ServiceName)
	return result, nil
}

// readTrimmed reads a line from stdin and trims whitespace.
// Keeping the helper small makes the interactive loop easier to follow.
func readTrimmed(reader *bufio.Reader) (string, error) {
	raw, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed reading input: %v", err)
	}
	return strings.TrimSpace(raw), nil
}

// splitAndClean turns a comma separated string into a slice without blanks.
// Sorting guarantees stable identifiers for log and service names.
func splitAndClean(raw string) []string {
	pieces := strings.Split(raw, ",")
	cleaned := make([]string, 0, len(pieces))
	for _, part := range pieces {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	sort.Strings(cleaned)
	return cleaned
}

// askRemotePort asks the operator for a remote port, defaulting to the provided local port when empty.
// Returning the chosen port keeps the call sites simple.
func askRemotePort(reader *bufio.Reader, localPort string) (string, error) {
	fmt.Printf(colorize(greenText, "Remote port for local %s (press Enter to reuse local port): "), localPort)
	remotePort, err := readTrimmed(reader)
	if err != nil {
		return "", err
	}
	if remotePort == "" {
		return localPort, nil
	}
	return remotePort, nil
}

// buildIdentifier produces a consistent suffix with protocols and ports to use in filenames and service names.
// Keeping all components in a single slice simplifies composing multiple strings later.
func buildIdentifier(protocols []string, tcpRoutes, udpRoutes []config.Route) []string {
	parts := make([]string, 0)
	lowerProtocols := make([]string, len(protocols))
	for i, proto := range protocols {
		lowerProtocols[i] = strings.ToLower(proto)
	}
	sort.Strings(lowerProtocols)

	for _, proto := range lowerProtocols {
		parts = append(parts, proto)
		switch proto {
		case "tcp":
			ports := collectPorts(tcpRoutes)
			parts = append(parts, ports...)
		case "udp":
			ports := collectPorts(udpRoutes)
			parts = append(parts, ports...)
		}
	}
	return parts
}

// collectPorts extracts sorted local ports from routes to maintain stable identifiers.
func collectPorts(routes []config.Route) []string {
	ports := make([]string, 0, len(routes))
	for _, route := range routes {
		ports = append(ports, route.LocalPort)
	}
	sort.Strings(ports)
	return ports
}

// routesFlagValue converts route slices into the CLI flag syntax used by the main process and the systemd unit.
// Keeping the formatter here avoids duplication between the interactive layer and main.
func routesFlagValue(routes []config.Route) string {
	if len(routes) == 0 {
		return ""
	}

	values := make([]string, 0, len(routes))
	for _, route := range routes {
		values = append(values, fmt.Sprintf("%s:%s:%s", route.LocalPort, route.RemoteIP, route.RemotePort))
	}
	return strings.Join(values, ",")
}

// colorize wraps a message with ANSI color codes so prompts stay readable without extra dependencies.
// Using simple color helpers keeps interactive output friendly without complicating logic.
func colorize(color, message string) string {
	return color + message + resetText
}
