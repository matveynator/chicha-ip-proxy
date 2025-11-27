// Package main wires CLI parsing with TCP and UDP proxy workers.
// The application favors channels and goroutines to follow Go proverbs about communication over shared memory.
package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/matveynator/chicha-tcp-proxy/pkg/config"
	"github.com/matveynator/chicha-tcp-proxy/pkg/logging"
	"github.com/matveynator/chicha-tcp-proxy/pkg/proxy"
)

// version holds the current version of the proxy application.
// A programmer might increment this as they update the application.
var version = "dev"

func main() {
	routesFlag := flag.String("routes", "", "Comma-separated list of TCP routes in the format LOCALPORT:REMOTEIP:REMOTEPORT")
	udpRoutesFlag := flag.String("udp-routes", "", "Comma-separated list of UDP routes in the format LOCALPORT:REMOTEIP:REMOTEPORT")
	logFile := flag.String("log", "chicha-tcp-proxy.log", "Path to the log file")
	rotationFrequency := flag.Duration("rotation", 24*time.Hour, "Log rotation frequency (e.g. 24h, 1h, etc.)")
	versionFlag := flag.Bool("version", false, "Print the version of the proxy and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("chicha-tcp-proxy version %s\n", version)
		return
	}

	tcpRoutes, err := config.ParseRoutes(*routesFlag)
	if err != nil {
		log.Fatalf("Error parsing TCP routes: %v", err)
	}
	udpRoutes, err := config.ParseRoutes(*udpRoutesFlag)
	if err != nil {
		log.Fatalf("Error parsing UDP routes: %v", err)
	}

	if len(tcpRoutes) == 0 && len(udpRoutes) == 0 {
		log.Fatal("Error: At least one of -routes or -udp-routes must be provided.")
	}

	fmt.Println("========== CHICHA TCP PROXY ==========")
	fmt.Println("TCP Routes:")
	for _, route := range tcpRoutes {
		fmt.Printf("  LocalPort=%s -> RemoteIP=%s RemotePort=%s\n", route.LocalPort, route.RemoteIP, route.RemotePort)
	}
	fmt.Println("UDP Routes:")
	for _, route := range udpRoutes {
		fmt.Printf("  LocalPort=%s -> RemoteIP=%s RemotePort=%s\n", route.LocalPort, route.RemoteIP, route.RemotePort)
	}
	fmt.Printf("Log file: %s\n", *logFile)
	fmt.Printf("Log rotation frequency: %v\n", *rotationFrequency)
	fmt.Println("======================================")

	logger, file, err := logging.SetupLogger(*logFile)
	if err != nil {
		log.Fatalf("Error setting up logger: %v", err)
	}

	log.Printf("Starting chicha-tcp-proxy version %s", version)

	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)
	logger.Printf("Using %d CPU cores", numCPUs)
	log.Printf("Using %d CPU cores", numCPUs)

	go logging.RotateLogs(*logFile, file, logger, *rotationFrequency)

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

	select {}
}
