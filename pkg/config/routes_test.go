package config

import (
	"net/netip"
	"testing"
)

func TestParseAllowListAcceptsIPsAndCIDRs(t *testing.T) {
	allowList, err := ParseAllowList([]string{"203.0.113.10", "10.0.0.0/24", " 2001:db8::1 "})
	if err != nil {
		t.Fatalf("ParseAllowList returned error: %v", err)
	}

	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{name: "exact IPv4", ip: "203.0.113.10", want: true},
		{name: "inside CIDR", ip: "10.0.0.25", want: true},
		{name: "outside CIDR", ip: "10.0.1.25", want: false},
		{name: "exact IPv6", ip: "2001:db8::1", want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addr := netip.MustParseAddr(test.ip)
			if got := allowList.Allows(addr); got != test.want {
				t.Fatalf("Allows(%s) = %v, want %v", test.ip, got, test.want)
			}
		})
	}
}

func TestParseAllowListEmptyAllowsEveryone(t *testing.T) {
	allowList, err := ParseAllowList(nil)
	if err != nil {
		t.Fatalf("ParseAllowList returned error: %v", err)
	}

	if !allowList.Allows(netip.MustParseAddr("198.51.100.77")) {
		t.Fatal("empty allowlist should allow every client IP")
	}
}

func TestParseAllowListRejectsInvalidEntries(t *testing.T) {
	_, err := ParseAllowList([]string{"not-an-ip"})
	if err == nil {
		t.Fatal("ParseAllowList accepted an invalid IP")
	}
}

func TestParseSimpleRouteCreatesDefaultTCPRoute(t *testing.T) {
	tcpRoutes, udpRoutes, used, err := ParseSimpleRoute(SimpleRouteFlags{
		Local:  "8080",
		Remote: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("ParseSimpleRoute returned error: %v", err)
	}
	if !used {
		t.Fatal("ParseSimpleRoute did not report simple flags as used")
	}
	if len(udpRoutes) != 0 {
		t.Fatalf("UDP route count = %d, want 0", len(udpRoutes))
	}
	if len(tcpRoutes) != 1 {
		t.Fatalf("TCP route count = %d, want 1", len(tcpRoutes))
	}
	route := tcpRoutes[0]
	if route.LocalPort != "8080" || route.RemoteIP != "203.0.113.10" || route.RemotePort != "8080" {
		t.Fatalf("route = %#v", route)
	}
}

func TestParseSimpleRouteCreatesUDPRouteWithRemotePort(t *testing.T) {
	tcpRoutes, udpRoutes, used, err := ParseSimpleRoute(SimpleRouteFlags{
		Local:  "5353",
		Remote: "203.0.113.20:53",
		Proto:  "udp",
	})
	if err != nil {
		t.Fatalf("ParseSimpleRoute returned error: %v", err)
	}
	if !used {
		t.Fatal("ParseSimpleRoute did not report simple flags as used")
	}
	if len(tcpRoutes) != 0 {
		t.Fatalf("TCP route count = %d, want 0", len(tcpRoutes))
	}
	if len(udpRoutes) != 1 {
		t.Fatalf("UDP route count = %d, want 1", len(udpRoutes))
	}
	route := udpRoutes[0]
	if route.LocalPort != "5353" || route.RemoteIP != "203.0.113.20" || route.RemotePort != "53" {
		t.Fatalf("route = %#v", route)
	}
}

func TestParseSimpleRouteRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name  string
		flags SimpleRouteFlags
	}{
		{name: "missing local", flags: SimpleRouteFlags{Remote: "203.0.113.10"}},
		{name: "missing remote", flags: SimpleRouteFlags{Local: "8080"}},
		{name: "invalid protocol", flags: SimpleRouteFlags{Local: "8080", Remote: "203.0.113.10", Proto: "icmp"}},
		{name: "invalid local port", flags: SimpleRouteFlags{Local: "0", Remote: "203.0.113.10"}},
		{name: "invalid remote IP", flags: SimpleRouteFlags{Local: "8080", Remote: "example.com"}},
		{name: "invalid remote port", flags: SimpleRouteFlags{Local: "8080", Remote: "203.0.113.10:70000"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, _, err := ParseSimpleRoute(test.flags)
			if err == nil {
				t.Fatal("ParseSimpleRoute accepted invalid input")
			}
		})
	}
}

func TestParseSimpleRouteIgnoresOnlyProto(t *testing.T) {
	_, _, used, err := ParseSimpleRoute(SimpleRouteFlags{Proto: "udp"})
	if err != nil {
		t.Fatalf("ParseSimpleRoute returned error: %v", err)
	}
	if used {
		t.Fatal("ParseSimpleRoute should not treat -proto alone as a route")
	}
}
