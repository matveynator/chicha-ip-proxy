// Package setup hosts interactive bootstrap helpers for first-run configuration.
// Grouping prompts together keeps the main package focused on runtime wiring while still following Go's advice to communicate via channels.
package setup

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/matveynator/chicha-ip-proxy/pkg/branding"
	"github.com/matveynator/chicha-ip-proxy/pkg/config"
)

// ErrSetupCancelled reports that the operator exited before saving the generated setup.
var ErrSetupCancelled = errors.New("setup cancelled")

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
	TargetIP   string
	RemotePort string
	LocalPort  string
	Protocol   string
	AllowRaw   string
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
			if choice == "6" || choice == "" {
				return result, nil
			}
			if choice == "7" {
				return nil, ErrSetupCancelled
			}

			if err := applyReviewEdit(reader, appName, &draft, choice); err != nil {
				return nil, err
			}
		}
	}
}

func printSetupHeader() {
	fmt.Print(branding.Banner)
	fmt.Printf(colorize(cyanText, "Operating system: %s\n"), runtime.GOOS)
	fmt.Println(colorize(yellowText, "Ctrl+C exits without saving."))
	fmt.Println()
}

func askInitialSetup(reader *bufio.Reader, draft *setupDraft) error {
	if err := askTargetIP(reader, draft); err != nil {
		return err
	}
	if err := askRemotePort(reader, draft); err != nil {
		return err
	}
	if err := askLocalPort(reader, draft); err != nil {
		return err
	}
	if err := askProtocol(reader, draft); err != nil {
		return err
	}
	return askAllowedClients(reader, draft)
}

func askTargetIP(reader *bufio.Reader, draft *setupDraft) error {
	fmt.Print(colorize(greenText, promptWithDefault("1) Target IP", draft.TargetIP)))
	targetIP, err := readTrimmed(reader)
	if err != nil {
		return err
	}
	if targetIP == "" {
		targetIP = draft.TargetIP
	}
	if _, err := netip.ParseAddr(targetIP); err != nil {
		return fmt.Errorf("target IP must be a valid IP address: %v", err)
	}
	draft.TargetIP = targetIP
	return nil
}

func askRemotePort(reader *bufio.Reader, draft *setupDraft) error {
	fmt.Print(colorize(greenText, promptWithDefault("2) Remote port", draft.RemotePort)))
	port, err := readTrimmed(reader)
	if err != nil {
		return err
	}
	if port == "" {
		port = draft.RemotePort
	}
	if err := validatePort(port); err != nil {
		return err
	}
	draft.RemotePort = port
	return nil
}

func askLocalPort(reader *bufio.Reader, draft *setupDraft) error {
	defaultLocalPort := draft.LocalPort
	if defaultLocalPort == "" {
		defaultLocalPort = draft.RemotePort
	}
	fmt.Print(colorize(greenText, promptWithDefault("3) Local port", defaultLocalPort)))
	localPort, err := readTrimmed(reader)
	if err != nil {
		return err
	}
	if localPort == "" {
		localPort = defaultLocalPort
	}
	if err := validatePort(localPort); err != nil {
		return err
	}
	draft.LocalPort = localPort
	printLocalPortStatuses(localPort)
	return nil
}

func askProtocol(reader *bufio.Reader, draft *setupDraft) error {
	fmt.Print(colorize(greenText, promptWithDefault("4) Protocol", draft.Protocol)))
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
	if draft.LocalPort != "" {
		printProtocolPortStatus(draft.LocalPort, protocol)
	}
	return nil
}

func askAllowedClients(reader *bufio.Reader, draft *setupDraft) error {
	defaultAllowRaw := draft.AllowRaw
	if defaultAllowRaw == "" {
		defaultAllowRaw = "all"
	}
	fmt.Print(colorize(greenText, promptWithDefault("5) Allowed client IPs/CIDRs", defaultAllowRaw)))
	allowRaw, err := readTrimmed(reader)
	if err != nil {
		return err
	}
	if allowRaw == "" {
		allowRaw = draft.AllowRaw
	}
	if strings.EqualFold(allowRaw, "all") {
		allowRaw = ""
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
	fmt.Println(colorize(cyanText, "  2) Change remote port"))
	fmt.Println(colorize(cyanText, "  3) Change local port"))
	fmt.Println(colorize(cyanText, "  4) Change protocol"))
	fmt.Println(colorize(cyanText, "  5) Change allowed clients"))
	fmt.Println(colorize(cyanText, "  6) Save and continue"))
	fmt.Println(colorize(cyanText, "  7) Exit without saving"))
	fmt.Print(colorize(greenText, "Choose [6]: "))

	choice, err := readTrimmed(reader)
	if err != nil {
		return "", err
	}
	if choice == "" {
		return "6", nil
	}
	switch choice {
	case "1", "2", "3", "4", "5", "6", "7":
		return choice, nil
	default:
		return "", fmt.Errorf("unknown setup menu option: %s", choice)
	}
}

func applyReviewEdit(reader *bufio.Reader, appName string, draft *setupDraft, choice string) error {
	var err error
	switch choice {
	case "1":
		err = askTargetIP(reader, draft)
	case "2":
		err = askRemotePort(reader, draft)
	case "3":
		err = askLocalPort(reader, draft)
	case "4":
		err = askProtocol(reader, draft)
	case "5":
		err = askAllowedClients(reader, draft)
	default:
		return nil
	}
	if err != nil {
		return err
	}
	result, err := buildInteractiveResult(appName, *draft)
	if err != nil {
		return err
	}
	printConfigReview(result)
	return nil
}

func buildInteractiveResult(appName string, draft setupDraft) (*InteractiveResult, error) {
	allowList, err := config.ParseAllowList(splitAndClean(draft.AllowRaw))
	if err != nil {
		return nil, err
	}

	route := config.Route{LocalPort: draft.LocalPort, RemoteIP: draft.TargetIP, RemotePort: draft.RemotePort}
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
	fmt.Println(colorize(purpleText, "Connection"))
	printRoutes("TCP", result.TCPRoutes)
	printRoutes("UDP", result.UDPRoutes)
	fmt.Printf(colorize(cyanText, "     protocol %s carries the traffic\n"), strings.ToUpper(result.ProtoFlag))
	fmt.Printf(colorize(cyanText, "     clients  %s\n"), allowListText(result.AllowFlags))
	fmt.Printf(colorize(cyanText, "     log      %s\n"), result.LogFile)
	fmt.Printf(colorize(cyanText, "     service  %s\n"), result.ServiceName)
	fmt.Printf(colorize(cyanText, "     command  %s\n"), setupCommandText(result))
	fmt.Println()
}

func printRoutes(label string, routes []config.Route) {
	if len(routes) == 0 {
		return
	}

	for _, route := range routes {
		fmt.Printf(colorize(cyanText, "     :%s on this machine  ->  %s:%s\n"), route.LocalPort, route.RemoteIP, route.RemotePort)
	}
}

func promptWithDefault(label, currentValue string) string {
	if currentValue == "" {
		return label + ": "
	}
	return fmt.Sprintf("%s [%s]: ", label, currentValue)
}

func allowListText(values []string) string {
	if len(values) == 0 {
		return "all client IPs"
	}
	return strings.Join(values, ", ")
}

func setupCommandText(result *InteractiveResult) string {
	parts := []string{
		"-local=" + result.LocalFlag,
		"-remote=" + result.RemoteFlag,
		"-proto=" + result.ProtoFlag,
	}
	for _, allowFlag := range result.AllowFlags {
		parts = append(parts, "-allow="+allowFlag)
	}
	return strings.Join(parts, " ")
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
