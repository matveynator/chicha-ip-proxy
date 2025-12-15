// Package main wires CLI parsing with TCP and UDP proxy workers.
// The application favors channels and goroutines to follow Go proverbs about communication over shared memory.
package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/matveynator/chicha-ip-proxy/pkg/config"
	"github.com/matveynator/chicha-ip-proxy/pkg/limits"
	"github.com/matveynator/chicha-ip-proxy/pkg/logging"
	"github.com/matveynator/chicha-ip-proxy/pkg/proxy"
	"github.com/matveynator/chicha-ip-proxy/pkg/setup"
)

// version holds the current version of the proxy application.
// A programmer might increment this as they update the application.
var version = "dev"

func main() {
	routesFlag := flag.String("routes", "", "Comma-separated list of TCP routes in the format LOCALPORT:REMOTEIP:REMOTEPORT")
	udpRoutesFlag := flag.String("udp-routes", "", "Comma-separated list of UDP routes in the format LOCALPORT:REMOTEIP:REMOTEPORT")
	logFile := flag.String("log", "chicha-ip-proxy.log", "Path to the log file")
	rotationFrequency := flag.Duration("rotation", 24*time.Hour, "Log rotation frequency (e.g. 24h, 1h, etc.)")
	versionFlag := flag.Bool("version", false, "Print the version of the proxy and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("chicha-ip-proxy version %s\n", version)
		return
	}

	// Parse routes passed through flags so scripted runs stay fast.
	tcpRoutes, err := config.ParseRoutes(*routesFlag)
	if err != nil {
		log.Fatalf("Error parsing TCP routes: %v", err)
	}
	udpRoutes, err := config.ParseRoutes(*udpRoutesFlag)
	if err != nil {
		log.Fatalf("Error parsing UDP routes: %v", err)
	}

	actualLogFile := *logFile
	var systemdResult *setup.SystemdResult

	// Fall back to interactive setup when no routes are provided.
	if len(tcpRoutes) == 0 && len(udpRoutes) == 0 {
		if runtime.GOOS != "linux" {
			showNonLinuxHelp()
			return
		}

		interactiveResult, err := setup.RunInteractiveSetup("chicha-ip-proxy")
		if err != nil {
			log.Fatalf("Interactive setup failed: %v", err)
		}

		tcpRoutes = interactiveResult.TCPRoutes
		udpRoutes = interactiveResult.UDPRoutes
		actualLogFile = interactiveResult.LogFile
		*routesFlag = interactiveResult.RoutesFlag
		*udpRoutesFlag = interactiveResult.UDPRoutesFlag

		systemdResult, err = setup.OfferSystemdSetup("chicha-ip-proxy", interactiveResult, *rotationFrequency)
		if err != nil {
			log.Printf("Systemd setup encountered an issue: %v", err)
		}
	}

	if len(tcpRoutes) == 0 && len(udpRoutes) == 0 {
		log.Fatal("Error: At least one of -routes or -udp-routes must be provided.")
	}

	// Print a concise summary before the workers launch to make deployments traceable.
	fmt.Println("========== CHICHA IP PROXY ==========")
	fmt.Println("TCP Routes:")
	for _, route := range tcpRoutes {
		fmt.Printf("  LocalPort=%s -> RemoteIP=%s RemotePort=%s\n", route.LocalPort, route.RemoteIP, route.RemotePort)
	}
	fmt.Println("UDP Routes:")
	for _, route := range udpRoutes {
		fmt.Printf("  LocalPort=%s -> RemoteIP=%s RemotePort=%s\n", route.LocalPort, route.RemoteIP, route.RemotePort)
	}
	fmt.Printf("Log file: %s\n", actualLogFile)
	fmt.Printf("Log rotation frequency: %v\n", *rotationFrequency)
	fmt.Println("Speed-up notice: system limits will be tuned on startup to keep the proxy responsive.")
	fmt.Println("======================================")

	logger, file, err := logging.SetupLogger(actualLogFile)
	if err != nil {
		log.Fatalf("Error setting up logger: %v", err)
	}

	if err := limits.SetupLimits(logger); err != nil {
		logger.Printf("System limit tuning encountered an issue: %v", err)
	}

	log.Printf("Starting chicha-ip-proxy version %s", version)

	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)
	logger.Printf("Using %d CPU cores", numCPUs)
	log.Printf("Using %d CPU cores", numCPUs)

	go logging.RotateLogs(actualLogFile, file, logger, *rotationFrequency, logging.DefaultMaxSizeBytes)

	for _, route := range tcpRoutes {
		listenAddr := ":" + route.LocalPort
		targetAddr := route.RemoteIP + ":" + route.RemotePort
		logger.Printf("Starting TCP proxy for route: local=%s remote=%s", listenAddr, targetAddr)
		go proxy.StartTCPProxy(listenAddr, targetAddr, logger)
	}

	for _, route := range udpRoutes {
		listenAddr := ":" + route.LocalPort
		targetAddr := route.RemoteIP + ":" + route.RemotePort
		logger.Printf("Starting UDP proxy for route: local=%s remote=%s", listenAddr, targetAddr)
		go proxy.StartUDPProxy(listenAddr, targetAddr, logger)
	}

	if systemdResult != nil && systemdResult.FollowLogs {
		stop := make(chan struct{})
		go setup.StreamLogs(actualLogFile, stop)
	}

	select {}
}

// showNonLinuxHelp displays CLI usage and runnable examples when interactive setup is unavailable.
// Keeping the helper small ensures the main path remains readable while offering guidance for other platforms.
func showNonLinuxHelp() {
	fmt.Println("Interactive setup works only on Linux. Please start the proxy with flags on this system.")
	fmt.Println()
	fmt.Println("Usage:")
	flag.CommandLine.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  TCP only : ./chicha-ip-proxy -routes=8080:203.0.113.10:80 -log=proxy.log")
	fmt.Println("  UDP only : ./chicha-ip-proxy -udp-routes=5353:203.0.113.20:53 -rotation=12h")
	fmt.Println("  Mixed run: ./chicha-ip-proxy -routes=8080:203.0.113.10:80 -udp-routes=5353:203.0.113.20:53")
	fmt.Println()
	fmt.Println("Tip: combine multiple ports with commas, for example -routes=" + strings.Join([]string{"8080:203.0.113.10:80", "8443:203.0.113.11:443"}, ","))
}
