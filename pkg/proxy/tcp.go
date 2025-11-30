// Package proxy hosts both TCP and UDP forwarding logic.
// Splitting TCP and UDP into separate files keeps protocol details isolated.
package proxy

import (
	"io"
	"log"
	"net"
	"runtime"
)

// StartTCPProxy listens on the provided address and forwards connections to the target.
// Using a channel for accepted connections keeps synchronization explicit without mutexes.
func StartTCPProxy(listenAddr, targetAddr string, logger *log.Logger) {
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

		connChan <- clientConn
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
