// Package proxy hosts both TCP and UDP forwarding logic.
// Splitting TCP and UDP into separate files keeps protocol details isolated.
package proxy

import (
	"io"
	"log"
	"net"
	"net/netip"
	"runtime"

	"github.com/matveynator/chicha-ip-proxy/pkg/config"
)

// StartTCPProxy listens on the provided address and forwards connections to the target.
// Using a channel for accepted connections keeps synchronization explicit without mutexes.
func StartTCPProxy(listenAddr, targetAddr string, allowList config.AllowList, logger *log.Logger) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Fatalf("Failed to start proxy on %s: %v", listenAddr, err)
	}
	defer listener.Close()

	logger.Printf("TCP proxy started on %s forwarding to %s", listenAddr, targetAddr)

	connChan := make(chan net.Conn)

	for i := 0; i < runtime.NumCPU(); i++ {
		go handleTCPConnections(connChan, targetAddr, logger)
	}

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			logger.Printf("Error accepting TCP connection on %s: %v", listenAddr, err)
			continue
		}

		clientIP, ok := remoteAddrIP(clientConn.RemoteAddr())
		if !ok || !allowList.Allows(clientIP) {
			logger.Printf("Rejected TCP connection from %s on %s: source IP is not allowed", clientConn.RemoteAddr().String(), listenAddr)
			rejectTCPConnectionWithReset(clientConn, logger)
			continue
		}

		connChan <- clientConn
	}
}

// remoteAddrIP extracts the host IP from network addresses before allowlist checks.
func remoteAddrIP(addr net.Addr) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return netip.Addr{}, false
	}

	parsed, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return parsed, true
}

// rejectTCPConnectionWithReset sends TCP RST when the platform exposes a TCP connection.
// Resetting denied clients makes allowlist failures immediate and avoids a graceful half-open flow.
func rejectTCPConnectionWithReset(conn net.Conn, logger *log.Logger) {
	tcpConn, ok := conn.(*net.TCPConn)
	if ok {
		if err := tcpConn.SetLinger(0); err != nil {
			logger.Printf("Failed to set TCP reset close for %s: %v", conn.RemoteAddr().String(), err)
		}
	}
	if err := conn.Close(); err != nil {
		logger.Printf("Failed to close rejected TCP connection from %s: %v", conn.RemoteAddr().String(), err)
	}
}

// handleTCPConnections establishes bidirectional copy pipelines for every TCP client.
// Each direction gets its own goroutine so that slow receivers do not block senders.
func handleTCPConnections(connChan <-chan net.Conn, targetAddr string, logger *log.Logger) {
	for {
		select {
		case clientConn, ok := <-connChan:
			if !ok {
				return
			}

			go func(conn net.Conn) {
				defer conn.Close()

				clientAddr := conn.RemoteAddr().String()
				logger.Printf("New TCP connection: %s -> %s", clientAddr, targetAddr)

				serverConn, err := net.Dial("tcp", targetAddr)
				if err != nil {
					logger.Printf("Failed to connect to TCP server %s: %v", targetAddr, err)
					return
				}
				defer serverConn.Close()

				done := make(chan struct{}, 2)

				go func() {
					_, err := io.Copy(serverConn, conn)
					if err != nil && err != io.EOF {
						logger.Printf("Error copying from TCP client %s to server %s: %v", clientAddr, targetAddr, err)
					}
					done <- struct{}{}
				}()

				go func() {
					_, err := io.Copy(conn, serverConn)
					if err != nil && err != io.EOF {
						logger.Printf("Error copying from TCP server %s to client %s: %v", targetAddr, clientAddr, err)
					}
					done <- struct{}{}
				}()

				<-done
				<-done

				logger.Printf("TCP connection closed: %s -> %s", clientAddr, targetAddr)
			}(clientConn)
		}
	}
}
