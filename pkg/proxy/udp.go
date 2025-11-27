// UDP support sits alongside TCP to keep both protocols available through the same interface.
// UDP is datagram based, so processing is organized around packets instead of streams.
package proxy

import (
	"log"
	"net"
	"runtime"
	"time"
)

// udpMessage represents a single datagram from a client.
// Keeping the payload in a dedicated struct makes it easy to fan out with channels.
type udpMessage struct {
	data []byte
	addr net.Addr
}

// udpSession keeps a dedicated connection to the remote for one client address.
// This avoids dialing on every packet and keeps source ports stable for servers like WireGuard.
type udpSession struct {
	clientAddr net.Addr
	remoteConn *net.UDPConn
	outbound   chan []byte
	lastActive time.Time
}

// StartUDPProxy listens for UDP datagrams and forwards them to the target endpoint.
// Work is coordinated by a session manager goroutine so there are no mutexes and no busy dialing.
func StartUDPProxy(listenAddr, targetAddr string, logger *log.Logger) {
	conn, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		logger.Fatalf("Failed to start UDP proxy on %s: %v", listenAddr, err)
	}
	defer conn.Close()

	logger.Printf("UDP proxy started on %s forwarding to %s", listenAddr, targetAddr)

	msgChan := make(chan udpMessage, runtime.NumCPU()*16)
	go manageUDPSessions(targetAddr, conn, logger, msgChan)

	buffer := make([]byte, 64*1024)
	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			logger.Printf("Error reading UDP packet on %s: %v", listenAddr, err)
			continue
		}

		payloadCopy := make([]byte, n)
		copy(payloadCopy, buffer[:n])

		msgChan <- udpMessage{data: payloadCopy, addr: addr}
	}
}

// manageUDPSessions multiplexes incoming datagrams to per-client sessions.
// A ticker retires idle sessions so resources stay bounded without manual cleanup.
func manageUDPSessions(targetAddr string, responder net.PacketConn, logger *log.Logger, msgChan <-chan udpMessage) {
	sessions := make(map[string]*udpSession)
	cleanupTicker := time.NewTicker(30 * time.Second)
	defer cleanupTicker.Stop()

	for {
		select {
		case msg := <-msgChan:
			sessionKey := msg.addr.String()
			session, ok := sessions[sessionKey]
			if !ok {
				resolvedTarget, err := net.ResolveUDPAddr("udp", targetAddr)
				if err != nil {
					logger.Printf("Failed to resolve UDP target %s: %v", targetAddr, err)
					continue
				}

				remoteConn, err := net.DialUDP("udp", nil, resolvedTarget)
				if err != nil {
					logger.Printf("Failed to dial UDP target %s: %v", targetAddr, err)
					continue
				}

				session = &udpSession{
					clientAddr: msg.addr,
					remoteConn: remoteConn,
					outbound:   make(chan []byte, 32),
					lastActive: time.Now(),
				}
				sessions[sessionKey] = session

				go forwardUDPPackets(session, logger)
				go relayUDPReplies(session, responder, logger)
			}

			session.lastActive = time.Now()

			select {
			case session.outbound <- msg.data:
			default:
				logger.Printf("Dropping UDP packet for %s due to full queue", session.clientAddr.String())
			}

		case <-cleanupTicker.C:
			for addr, session := range sessions {
				if time.Since(session.lastActive) > 60*time.Second {
					close(session.outbound)
					session.remoteConn.Close()
					delete(sessions, addr)
					logger.Printf("Closed idle UDP session for %s", addr)
				}
			}
		}
	}
}

// forwardUDPPackets pushes outbound payloads to the remote endpoint.
// Using a buffered channel keeps the hot path non-blocking when bursts happen.
func forwardUDPPackets(session *udpSession, logger *log.Logger) {
	for data := range session.outbound {
		_ = session.remoteConn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		if _, err := session.remoteConn.Write(data); err != nil {
			logger.Printf("Error sending UDP payload for %s: %v", session.clientAddr.String(), err)
			return
		}
	}
}

// relayUDPReplies reads replies from the remote server and writes them back to the originating client.
// A read deadline prevents stuck goroutines when remotes stay silent.
func relayUDPReplies(session *udpSession, responder net.PacketConn, logger *log.Logger) {
	replyBuf := make([]byte, 64*1024)
	for {
		_ = session.remoteConn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := session.remoteConn.Read(replyBuf)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// The remote can stay silent for a while, but the client may still be active.
			// Keep listening as long as the session shows recent activity so replies are not dropped.
			if time.Since(session.lastActive) < 60*time.Second {
				continue
			}
			return
		}
		if err != nil {
			logger.Printf("Error reading UDP reply for %s: %v", session.clientAddr.String(), err)
			return
		}

		if _, writeErr := responder.WriteTo(replyBuf[:n], session.clientAddr); writeErr != nil {
			logger.Printf("Error writing UDP reply to %s: %v", session.clientAddr.String(), writeErr)
			return
		}
	}
}
