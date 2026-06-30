// Package setup hosts interactive bootstrap helpers for first-run configuration.
// Grouping prompts together keeps the main package focused on runtime wiring while still following Go's advice to communicate via channels.
package setup

import (
	"bufio"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/matveynator/chicha-ip-proxy/pkg/config"
)

const (
	greenText  = "\033[32m"
	purpleText = "\033[35m"
	cyanText   = "\033[36m"
	yellowText = "\033[33m"
	resetText  = "\033[0m"
)

// InteractiveResult carries user-provided routes and derived metadata such as log and service names.
// Keeping it small makes it easy to hand over to the main package without introducing global state.
type InteractiveResult struct {
	TCPRoutes     []config.Route
	UDPRoutes     []config.Route
	AllowList     config.AllowList
	LogFile       string
	ServiceName   string
	LocalFlag     string
	RemoteFlag    string
	ProtoFlag     string
	RoutesFlag    string
	UDPRoutesFlag string
	AllowFlags    []string
}

type setupDraft struct {
	TargetIP string
	Port     string
	Protocol string
	AllowRaw string
}

// RunInteractiveSetup asks the operator for one route and source restrictions when flags are absent.
// The final review loop makes the generated service explicit before any system files are changed.
func RunInteractiveSetup(appName string) (*InteractiveResult, error) {
	reader := bufio.NewReader(os.Stdin)
	draft := setupDraft{Protocol: "tcp"}

	printSetupHeader()

	for {
		if err := askInitialSetup(reader, &draft); err != nil {
			return nil, err
		}

		for {
			result, err := buildInteractiveResult(appName, draft)
			if err != nil {
				return nil, err
			}

			printConfigReview(result)
			choice, err := askReviewChoice(reader)
			if err != nil {
				return nil, err
			}
			if choice == "5" || choice == "" {
				return result, nil
			}

			if err := applyReviewEdit(reader, &draft, choice); err != nil {
				return nil, err
			}
		}
	}
}

func printSetupHeader() {
	fmt.Println(colorize(purpleText, "Chicha IP Proxy setup"))
	fmt.Println(colorize(cyanText, "Configure one forwarding route and choose which client IPs may use it."))
	fmt.Printf(colorize(cyanText, "Operating system: %s\n"), runtime.GOOS)
	fmt.Println(colorize(greenText, "Press Enter on protocol to use TCP. Leave allowed clients empty to allow everyone."))
	fmt.Println(colorize(greenText, "Startup will tune system limits to keep the proxy fast."))
	fmt.Println()
}

func askInitialSetup(reader *bufio.Reader, draft *setupDraft) error {
	if err := askTargetIP(reader, draft); err != nil {
		return err
	}
	if err := askPort(reader, draft); err != nil {
		return err
	}
	if err := askProtocol(reader, draft); err != nil {
		return err
	}
	return askAllowedClients(reader, draft)
}

func askTargetIP(reader *bufio.Reader, draft *setupDraft) error {
	fmt.Print(colorize(greenText, "1) Target IP: "))
	targetIP, err := readTrimmed(reader)
	if err != nil {
		return err
	}
	if _, err := netip.ParseAddr(targetIP); err != nil {
		return fmt.Errorf("target IP must be a valid IP address: %v", err)
	}
	draft.TargetIP = targetIP
	return nil
}

func askPort(reader *bufio.Reader, draft *setupDraft) error {
	fmt.Print(colorize(greenText, "2) Port: "))
	port, err := readTrimmed(reader)
	if err != nil {
		return err
	}
	if err := validatePort(port); err != nil {
		return err
	}
	draft.Port = port
	printLocalPortStatuses(port)
	return nil
}

func askProtocol(reader *bufio.Reader, draft *setupDraft) error {
	fmt.Print(colorize(greenText, "3) Protocol [tcp]: "))
	protocol, err := readTrimmed(reader)
	if err != nil {
		return err
	}
	if protocol == "" {
		protocol = "tcp"
	}
	protocol = strings.ToLower(protocol)
	if protocol != "tcp" && protocol != "udp" {
		return fmt.Errorf("protocol must be tcp or udp")
	}
	draft.Protocol = protocol
	if draft.Port != "" {
		printProtocolPortStatus(draft.Port, protocol)
	}
	return nil
}

func askAllowedClients(reader *bufio.Reader, draft *setupDraft) error {
	fmt.Print(colorize(greenText, "4) Allowed client IPs/CIDRs, comma separated [all]: "))
	allowRaw, err := readTrimmed(reader)
	if err != nil {
		return err
	}
	if _, err := config.ParseAllowList(splitAndClean(allowRaw)); err != nil {
		return err
	}
	draft.AllowRaw = allowRaw
	return nil
}

func askReviewChoice(reader *bufio.Reader) (string, error) {
	fmt.Println(colorize(yellowText, "Change something before writing the service?"))
	fmt.Println(colorize(cyanText, "  1) Change target IP"))
	fmt.Println(colorize(cyanText, "  2) Change port"))
	fmt.Println(colorize(cyanText, "  3) Change protocol"))
	fmt.Println(colorize(cyanText, "  4) Change allowed clients"))
	fmt.Println(colorize(cyanText, "  5) Continue"))
	fmt.Print(colorize(greenText, "Choose [5]: "))

	choice, err := readTrimmed(reader)
	if err != nil {
		return "", err
	}
	if choice == "" {
		return "5", nil
	}
	switch choice {
	case "1", "2", "3", "4", "5":
		return choice, nil
	default:
		return "", fmt.Errorf("unknown setup menu option: %s", choice)
	}
}

func applyReviewEdit(reader *bufio.Reader, draft *setupDraft, choice string) error {
	switch choice {
	case "1":
		return askTargetIP(reader, draft)
	case "2":
		return askPort(reader, draft)
	case "3":
		return askProtocol(reader, draft)
	case "4":
		return askAllowedClients(reader, draft)
	default:
		return nil
	}
}

func buildInteractiveResult(appName string, draft setupDraft) (*InteractiveResult, error) {
	allowList, err := config.ParseAllowList(splitAndClean(draft.AllowRaw))
	if err != nil {
		return nil, err
	}

	route := config.Route{LocalPort: draft.Port, RemoteIP: draft.TargetIP, RemotePort: draft.Port}
	tcpRoutes := make([]config.Route, 0, 1)
	udpRoutes := make([]config.Route, 0, 1)
	if draft.Protocol == "udp" {
		udpRoutes = append(udpRoutes, route)
	} else {
		tcpRoutes = append(tcpRoutes, route)
	}

	identifier := strings.Join(buildIdentifier([]string{draft.Protocol}, tcpRoutes, udpRoutes), "-")
	return &InteractiveResult{
		TCPRoutes:     tcpRoutes,
		UDPRoutes:     udpRoutes,
		AllowList:     allowList,
		LogFile:       defaultLogFile(appName, identifier),
		ServiceName:   defaultAutostartName(appName, identifier),
		LocalFlag:     route.LocalPort,
		RemoteFlag:    simpleRemoteFlag(route),
		ProtoFlag:     draft.Protocol,
		RoutesFlag:    routesFlagValue(tcpRoutes),
		UDPRoutesFlag: routesFlagValue(udpRoutes),
		AllowFlags:    allowList.FlagValues(),
	}, nil
}

func printConfigReview(result *InteractiveResult) {
	fmt.Println()
	fmt.Println(colorize(purpleText, "Full configuration"))
	printRoutes("TCP", result.TCPRoutes)
	printRoutes("UDP", result.UDPRoutes)
	fmt.Printf(colorize(cyanText, "Allowed clients: %s\n"), allowListText(result.AllowFlags))
	fmt.Printf(colorize(cyanText, "CLI local: %s\n"), commandValueOrNone(result.LocalFlag))
	fmt.Printf(colorize(cyanText, "CLI remote: %s\n"), commandValueOrNone(result.RemoteFlag))
	fmt.Printf(colorize(cyanText, "CLI proto: %s\n"), commandValueOrNone(result.ProtoFlag))
	fmt.Printf(colorize(cyanText, "CLI allow flags: %s\n"), allowFlagsText(result.AllowFlags))
	fmt.Printf(colorize(cyanText, "Log file: %s\n"), result.LogFile)
	fmt.Printf(colorize(cyanText, "Autostart name: %s\n"), result.ServiceName)
	fmt.Println()
}

func printRoutes(label string, routes []config.Route) {
	if len(routes) == 0 {
		fmt.Printf(colorize(cyanText, "%s route: none\n"), label)
		return
	}

	for _, route := range routes {
		fmt.Printf(colorize(cyanText, "%s route: local %s -> %s:%s\n"), label, route.LocalPort, route.RemoteIP, route.RemotePort)
	}
}

func allowListText(values []string) string {
	if len(values) == 0 {
		return "all client IPs"
	}
	return strings.Join(values, ", ")
}

func allowFlagsText(values []string) string {
	if len(values) == 0 {
		return "none"
	}

	flags := make([]string, 0, len(values))
	for _, value := range values {
		flags = append(flags, "-allow="+value)
	}
	return strings.Join(flags, " ")
}

func commandValueOrNone(value string) string {
	if value == "" {
		return "none"
	}
	return value
}

func validatePort(port string) error {
	return config.ValidatePort(port)
}

type localPortStatus struct {
	Protocol  string
	Available bool
	Err       error
}

func printLocalPortStatuses(port string) {
	fmt.Printf(colorize(yellowText, "Local port %s status now:\n"), port)
	for _, status := range checkLocalPortStatuses(port) {
		fmt.Println(formatLocalPortStatus(status))
	}
}

func printProtocolPortStatus(port, protocol string) {
	status := checkLocalPortStatus(port, protocol)
	fmt.Println(formatLocalPortStatus(status))
}

func checkLocalPortStatuses(port string) []localPortStatus {
	return []localPortStatus{
		checkLocalPortStatus(port, "tcp"),
		checkLocalPortStatus(port, "udp"),
	}
}

func checkLocalPortStatus(port, protocol string) localPortStatus {
	switch protocol {
	case "tcp":
		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			return localPortStatus{Protocol: "tcp", Available: false, Err: err}
		}
		listener.Close()
		return localPortStatus{Protocol: "tcp", Available: true}
	case "udp":
		conn, err := net.ListenPacket("udp", ":"+port)
		if err != nil {
			return localPortStatus{Protocol: "udp", Available: false, Err: err}
		}
		conn.Close()
		return localPortStatus{Protocol: "udp", Available: true}
	default:
		return localPortStatus{Protocol: protocol, Available: false, Err: fmt.Errorf("unknown protocol")}
	}
}

func formatLocalPortStatus(status localPortStatus) string {
	protocol := strings.ToUpper(status.Protocol)
	if status.Available {
		return colorize(greenText, fmt.Sprintf("  %s: free", protocol))
	}
	return colorize(yellowText, fmt.Sprintf("  %s: busy or unavailable (%v)", protocol, status.Err))
}

func simpleRemoteFlag(route config.Route) string {
	if route.RemotePort == route.LocalPort {
		return route.RemoteIP
	}
	return route.RemoteIP + ":" + route.RemotePort
}

func defaultLogFile(appName, identifier string) string {
	fileName := fmt.Sprintf("%s-%s.log", appName, identifier)

	switch runtime.GOOS {
	case "windows":
		programData := os.Getenv("ProgramData")
		if programData == "" {
			programData = `C:\ProgramData`
		}
		return filepath.Join(programData, appName, fileName)
	case "darwin":
		return filepath.Join("/Library/Logs", fileName)
	case "freebsd", "openbsd", "linux":
		return filepath.Join("/var/log", fileName)
	default:
		return fileName
	}
}

func defaultAutostartName(appName, identifier string) string {
	name := fmt.Sprintf("%s-%s", appName, identifier)
	if runtime.GOOS == "darwin" {
		return "com.matveynator." + name
	}
	return name
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
