// Package proxy hosts both TCP and UDP forwarding logic.
// Splitting TCP and UDP into separate files keeps protocol details isolated.
package proxy

import (
	"log"
	"net"
	"net/netip"
	"runtime"
	"time"

	"github.com/matveynator/chicha-ip-proxy/pkg/config"
)

const (
	defaultMaxTCPConnectionsPerRoute = 1024
	tcpDialTimeout                   = 10 * time.Second
	tcpIdleTimeout                   = 5 * time.Minute
	tcpWriteTimeout                  = 30 * time.Second
)

type tcpConnJob struct {
	conn    net.Conn
	release <-chan struct{}
}

// StartTCPProxy listens on the provided address and forwards connections to the target.
// Using a channel for accepted connections keeps synchronization explicit without mutexes.
func StartTCPProxy(listenAddr, targetAddr string, allowList config.AllowList, logger *log.Logger) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Fatalf("Failed to start proxy on %s: %v", listenAddr, err)
	}
	defer listener.Close()

	logger.Printf("TCP proxy started on %s forwarding to %s", listenAddr, targetAddr)

	connChan := make(chan tcpConnJob)
	activeConnections := make(chan struct{}, defaultMaxTCPConnectionsPerRoute)

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

		select {
		case activeConnections <- struct{}{}:
		default:
			logger.Printf("Rejected TCP connection from %s on %s: connection limit reached", clientConn.RemoteAddr().String(), listenAddr)
			rejectTCPConnectionWithReset(clientConn, logger)
			continue
		}

		connChan <- tcpConnJob{conn: clientConn, release: activeConnections}
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
	resetTCPConnection(conn, logger)
}

// resetTCPConnection makes proxy-side failures visible to clients as immediate TCP failures.
// This keeps denied clients and unreachable upstream targets from looking like silent hangs.
func resetTCPConnection(conn net.Conn, logger *log.Logger) {
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
func handleTCPConnections(connChan <-chan tcpConnJob, targetAddr string, logger *log.Logger) {
	for {
		select {
		case job, ok := <-connChan:
			if !ok {
				return
			}

			go handleTCPConnection(job, targetAddr, logger)
		}
	}
}

func handleTCPConnection(job tcpConnJob, targetAddr string, logger *log.Logger) {
	conn := job.conn
	defer func() {
		<-job.release
	}()
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	logger.Printf("New TCP connection: %s -> %s", clientAddr, targetAddr)

	dialer := net.Dialer{Timeout: tcpDialTimeout}
	serverConn, err := dialer.Dial("tcp", targetAddr)
	if err != nil {
		logger.Printf("Failed to connect to TCP server %s: %v", targetAddr, err)
		resetTCPConnection(conn, logger)
		return
	}
	defer serverConn.Close()

	done := make(chan struct{}, 2)
	go copyTCPStream(serverConn, conn, "client", clientAddr, targetAddr, logger, done)
	go copyTCPStream(conn, serverConn, "server", clientAddr, targetAddr, logger, done)

	<-done
	conn.Close()
	serverConn.Close()
	<-done

	logger.Printf("TCP connection closed: %s -> %s", clientAddr, targetAddr)
}

func copyTCPStream(dst net.Conn, src net.Conn, direction, clientAddr, targetAddr string, logger *log.Logger, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	buffer := make([]byte, 32*1024)
	for {
		_ = src.SetReadDeadline(time.Now().Add(tcpIdleTimeout))
		n, readErr := src.Read(buffer)
		if n > 0 {
			_ = dst.SetWriteDeadline(time.Now().Add(tcpWriteTimeout))
			if writeErr := writeFull(dst, buffer[:n]); writeErr != nil {
				logger.Printf("Error writing TCP %s stream for %s -> %s: %v", direction, clientAddr, targetAddr, writeErr)
				return
			}
		}
		if readErr != nil {
			if netErr, ok := readErr.(net.Error); ok && netErr.Timeout() {
				logger.Printf("Closing idle TCP %s stream for %s -> %s", direction, clientAddr, targetAddr)
			}
			return
		}
	}
}

func writeFull(conn net.Conn, payload []byte) error {
	for len(payload) > 0 {
		n, err := conn.Write(payload)
		if err != nil {
			return err
		}
		payload = payload[n:]
	}
	return nil
}
