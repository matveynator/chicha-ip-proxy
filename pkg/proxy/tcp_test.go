package proxy

import (
	"errors"
	"io"
	"log"
	"net"
	"net/netip"
	"syscall"
	"testing"
	"time"
)

func TestRejectTCPConnectionWithResetDoesNotCloseGracefully(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen returned error: %v", err)
	}
	defer listener.Close()

	accepted := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			accepted <- err
			return
		}
		rejectTCPConnectionWithReset(conn, log.New(io.Discard, "", 0))
		accepted <- nil
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial returned error: %v", err)
	}
	defer clientConn.Close()

	if err := <-accepted; err != nil {
		t.Fatalf("listener.Accept returned error: %v", err)
	}

	if err := clientConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline returned error: %v", err)
	}

	_, err = clientConn.Read(make([]byte, 1))
	if err == nil {
		t.Fatal("Read succeeded after rejected TCP connection")
	}
	if errors.Is(err, io.EOF) {
		t.Fatalf("Read returned graceful EOF, want TCP reset error")
	}
	if !errors.Is(err, syscall.ECONNRESET) {
		t.Fatalf("Read returned %v, want connection reset", err)
	}
}

func TestRemoteAddrIPAcceptsIPv6SocketAddress(t *testing.T) {
	addr := &net.TCPAddr{IP: net.ParseIP("2001:db8::7"), Port: 51820}

	got, ok := remoteAddrIP(addr)
	if !ok {
		t.Fatal("remoteAddrIP rejected IPv6 socket address")
	}
	if got != netip.MustParseAddr("2001:db8::7") {
		t.Fatalf("remoteAddrIP = %s", got)
	}
}
