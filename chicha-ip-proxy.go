// Package main wires CLI parsing with TCP and UDP proxy workers.
// The application favors channels and goroutines to follow Go proverbs about communication over shared memory.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/matveynator/chicha-ip-proxy/pkg/branding"
	"github.com/matveynator/chicha-ip-proxy/pkg/config"
	"github.com/matveynator/chicha-ip-proxy/pkg/limits"
	"github.com/matveynator/chicha-ip-proxy/pkg/logging"
	"github.com/matveynator/chicha-ip-proxy/pkg/proxy"
	"github.com/matveynator/chicha-ip-proxy/pkg/setup"
	"github.com/matveynator/chicha-ip-proxy/pkg/version"
)

func main() {
	localFlag := flag.String("local", "", "Local port to listen on")
	remoteFlag := flag.String("remote", "", "Remote target IP or IP:PORT")
	protoFlag := flag.String("proto", "tcp", "Protocol to proxy: tcp or udp")
	allowFlags := repeatedFlag{}
	flag.Var(&allowFlags, "allow", "Client IP or CIDR allowed to use the proxy. Repeat for multiple sources.")
	logFile := flag.String("log", "chicha-ip-proxy.log", "Path to the log file")
	rotationFrequency := flag.Duration("rotation", 24*time.Hour, "Log rotation frequency (e.g. 24h, 1h, etc.)")
	versionFlag := flag.Bool("version", false, "Print the version of the proxy and exit")

	// Legacy route flags stay registered for compatibility but are intentionally absent from help output.
	routesFlag := flag.String("routes", "", "legacy TCP routes in LOCALPORT:REMOTEIP:REMOTEPORT format")
	udpRoutesFlag := flag.String("udp-routes", "", "legacy UDP routes in LOCALPORT:REMOTEIP:REMOTEPORT format")

	flag.Usage = showFlagHelp
	flag.Parse()

	// Resolve the build version once so every subsystem prints a consistent identifier.
	appVersion := version.Resolve()

	if *versionFlag {
		fmt.Printf("chicha-ip-proxy version %s\n", appVersion)
		return
	}
	if err := validateRotationFrequency(*rotationFrequency); err != nil {
		log.Fatalf("Error: %v", err)
	}

	tcpRoutes, udpRoutes, err := parseRoutesFromFlags(*routesFlag, *udpRoutesFlag, config.SimpleRouteFlags{
		Local:  *localFlag,
		Remote: *remoteFlag,
		Proto:  *protoFlag,
	})
	if err != nil {
		log.Fatalf("Error parsing route flags: %v", err)
	}
	allowList, err := config.ParseAllowList(allowFlags.Values)
	if err != nil {
		log.Fatalf("Error parsing allowed client sources: %v", err)
	}

	actualLogFile := *logFile
	var autostartResult *setup.SystemdResult

	// Fall back to interactive setup when no routes are provided.
	if len(tcpRoutes) == 0 && len(udpRoutes) == 0 {
		interactiveResult, err := setup.RunInteractiveSetup("chicha-ip-proxy")
		if err != nil {
			if errors.Is(err, setup.ErrSetupCancelled) {
				fmt.Println("Setup cancelled.")
				return
			}
			log.Fatalf("Interactive setup failed: %v", err)
		}

		tcpRoutes = interactiveResult.TCPRoutes
		udpRoutes = interactiveResult.UDPRoutes
		allowList = interactiveResult.AllowList
		actualLogFile = interactiveResult.LogFile

		autostartResult, err = setup.OfferAutostartSetup("chicha-ip-proxy", interactiveResult, *rotationFrequency)
		if err != nil {
			log.Printf("Autostart setup encountered an issue: %v", err)
		}
	}

	if len(tcpRoutes) == 0 && len(udpRoutes) == 0 {
		log.Fatal("Error: provide -local and -remote, use legacy -routes/-udp-routes, or run without route flags for interactive setup.")
	}

	printStartupSummary(tcpRoutes, udpRoutes, allowList, actualLogFile)

	logger, file, err := logging.SetupLogger(actualLogFile)
	if err != nil {
		log.Fatalf("Error setting up logger: %v", err)
	}

	if err := limits.SetupLimits(logger); err != nil {
		logger.Printf("System limit tuning encountered an issue: %v", err)
	}

	log.Printf("Starting chicha-ip-proxy version %s", appVersion)

	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)
	logger.Printf("Using %d CPU cores", numCPUs)
	log.Printf("Using %d CPU cores", numCPUs)

	go logging.RotateLogs(actualLogFile, file, logger, *rotationFrequency, logging.DefaultMaxSizeBytes)

	for _, route := range tcpRoutes {
		listenAddr := ":" + route.LocalPort
		targetAddr := route.RemoteAddress()
		logger.Printf("Starting TCP proxy for route: local=%s remote=%s", listenAddr, targetAddr)
		go proxy.StartTCPProxy(listenAddr, targetAddr, allowList, logger)
	}

	for _, route := range udpRoutes {
		listenAddr := ":" + route.LocalPort
		targetAddr := route.RemoteAddress()
		logger.Printf("Starting UDP proxy for route: local=%s remote=%s", listenAddr, targetAddr)
		go proxy.StartUDPProxy(listenAddr, targetAddr, allowList, logger)
	}

	if autostartResult != nil && autostartResult.FollowLogs {
		stop := make(chan struct{})
		go setup.StreamLogs(actualLogFile, stop)
	}

	select {}
}

func printStartupSummary(tcpRoutes, udpRoutes []config.Route, allowList config.AllowList, logFile string) {
	fmt.Print(branding.Banner)
	for _, route := range tcpRoutes {
		fmt.Printf("tcp  :%s -> %s\n", route.LocalPort, route.RemoteAddress())
	}
	for _, route := range udpRoutes {
		fmt.Printf("udp  :%s -> %s\n", route.LocalPort, route.RemoteAddress())
	}
	fmt.Printf("allow %s\n", allowListSummary(allowList))
	fmt.Printf("log   %s\n\n", logFile)
}

func parseRoutesFromFlags(legacyTCPRoutes, legacyUDPRoutes string, simpleFlags config.SimpleRouteFlags) ([]config.Route, []config.Route, error) {
	if legacyTCPRoutes != "" || legacyUDPRoutes != "" {
		tcpRoutes, err := config.ParseRoutes(legacyTCPRoutes)
		if err != nil {
			return nil, nil, fmt.Errorf("legacy TCP routes: %v", err)
		}
		udpRoutes, err := config.ParseRoutes(legacyUDPRoutes)
		if err != nil {
			return nil, nil, fmt.Errorf("legacy UDP routes: %v", err)
		}
		return tcpRoutes, udpRoutes, nil
	}

	tcpRoutes, udpRoutes, _, err := config.ParseSimpleRoute(simpleFlags)
	return tcpRoutes, udpRoutes, err
}

func validateRotationFrequency(rotation time.Duration) error {
	if rotation <= 0 {
		return fmt.Errorf("-rotation must be positive")
	}
	return nil
}

// repeatedFlag stores every occurrence of flags such as -allow.
type repeatedFlag struct {
	Values []string
}

func (flagValue *repeatedFlag) String() string {
	return strings.Join(flagValue.Values, ",")
}

func (flagValue *repeatedFlag) Set(value string) error {
	flagValue.Values = append(flagValue.Values, value)
	return nil
}

// allowListSummary keeps CLI output explicit about whether the proxy is open or restricted.
func allowListSummary(allowList config.AllowList) string {
	values := allowList.FlagValues()
	if len(values) == 0 {
		return "all client IPs"
	}
	return strings.Join(values, ", ")
}

// showFlagHelp displays CLI usage and runnable examples for scripted runs.
func showFlagHelp() {
	fmt.Print(branding.Banner)
	fmt.Println("Usage:")
	fmt.Println("  chicha-ip-proxy -local=PORT -remote=IP|IP:PORT|[IPv6]:PORT [options]")
	fmt.Println("  chicha-ip-proxy        # setup wizard")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -local PORT")
	fmt.Println("  -remote IP|IP:PORT|[IPv6]:PORT")
	fmt.Println("  -proto tcp|udp")
	fmt.Println("  -allow IP|CIDR")
	fmt.Println("  -log PATH")
	fmt.Println("  -rotation 24h")
	fmt.Println("  -version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  chicha-ip-proxy -local=8080 -remote=203.0.113.10 -allow=198.51.100.7")
	fmt.Println("  chicha-ip-proxy -local=5353 -remote=203.0.113.20:53 -proto=udp")
	fmt.Println("  chicha-ip-proxy -local=8443 -remote=[2001:db8::10]:443")
}
