package setup

import (
	"bufio"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBuildInteractiveResultCreatesTCPRouteAndAllowFlags(t *testing.T) {
	result, err := buildInteractiveResult("chicha-ip-proxy", setupDraft{
		TargetIP:   "203.0.113.20",
		RemotePort: "8080",
		LocalPort:  "8080",
		Protocol:   "tcp",
		AllowRaw:   "198.51.100.7, 10.0.0.0/24",
	})
	if err != nil {
		t.Fatalf("buildInteractiveResult returned error: %v", err)
	}

	if len(result.TCPRoutes) != 1 {
		t.Fatalf("TCP route count = %d, want 1", len(result.TCPRoutes))
	}
	if len(result.UDPRoutes) != 0 {
		t.Fatalf("UDP route count = %d, want 0", len(result.UDPRoutes))
	}
	if result.RoutesFlag != "8080:203.0.113.20:8080" {
		t.Fatalf("RoutesFlag = %q", result.RoutesFlag)
	}
	if result.LocalFlag != "8080" || result.RemoteFlag != "203.0.113.20" || result.ProtoFlag != "tcp" {
		t.Fatalf("simple flags = local %q remote %q proto %q", result.LocalFlag, result.RemoteFlag, result.ProtoFlag)
	}
	if !reflect.DeepEqual(result.AllowFlags, []string{"10.0.0.0/24", "198.51.100.7"}) {
		t.Fatalf("AllowFlags = %#v", result.AllowFlags)
	}
}

func TestBuildInteractiveResultCreatesUDPRoute(t *testing.T) {
	result, err := buildInteractiveResult("chicha-ip-proxy", setupDraft{
		TargetIP:   "203.0.113.53",
		RemotePort: "5353",
		LocalPort:  "5353",
		Protocol:   "udp",
	})
	if err != nil {
		t.Fatalf("buildInteractiveResult returned error: %v", err)
	}

	if len(result.TCPRoutes) != 0 {
		t.Fatalf("TCP route count = %d, want 0", len(result.TCPRoutes))
	}
	if result.UDPRoutesFlag != "5353:203.0.113.53:5353" {
		t.Fatalf("UDPRoutesFlag = %q", result.UDPRoutesFlag)
	}
	if result.LocalFlag != "5353" || result.RemoteFlag != "203.0.113.53" || result.ProtoFlag != "udp" {
		t.Fatalf("simple flags = local %q remote %q proto %q", result.LocalFlag, result.RemoteFlag, result.ProtoFlag)
	}
}

func TestBuildArgsIncludesAllowFlags(t *testing.T) {
	result, err := buildInteractiveResult("chicha-ip-proxy", setupDraft{
		TargetIP:   "203.0.113.20",
		RemotePort: "8080",
		LocalPort:  "8080",
		Protocol:   "tcp",
		AllowRaw:   "198.51.100.7",
	})
	if err != nil {
		t.Fatalf("buildInteractiveResult returned error: %v", err)
	}

	args := buildArgs(result, time.Hour)
	want := []string{
		"-local=8080",
		"-remote=203.0.113.20",
		"-allow=198.51.100.7",
		"-log=" + defaultLogFile("chicha-ip-proxy", "tcp-8080"),
		"-rotation=1h0m0s",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("buildArgs = %#v, want %#v", args, want)
	}
}

func TestBuildInteractiveResultAllowsDifferentLocalPort(t *testing.T) {
	result, err := buildInteractiveResult("chicha-ip-proxy", setupDraft{
		TargetIP:   "203.0.113.20",
		RemotePort: "80",
		LocalPort:  "8080",
		Protocol:   "tcp",
	})
	if err != nil {
		t.Fatalf("buildInteractiveResult returned error: %v", err)
	}

	if len(result.TCPRoutes) != 1 {
		t.Fatalf("TCP route count = %d, want 1", len(result.TCPRoutes))
	}
	route := result.TCPRoutes[0]
	if route.LocalPort != "8080" || route.RemotePort != "80" {
		t.Fatalf("route ports = local %q remote %q", route.LocalPort, route.RemotePort)
	}
	if result.RemoteFlag != "203.0.113.20:80" {
		t.Fatalf("RemoteFlag = %q", result.RemoteFlag)
	}
}

func TestBuildInteractiveResultFormatsIPv6RemoteFlags(t *testing.T) {
	result, err := buildInteractiveResult("chicha-ip-proxy", setupDraft{
		TargetIP:   "2001:db8::20",
		RemotePort: "443",
		LocalPort:  "8443",
		Protocol:   "tcp",
		AllowRaw:   "2001:db8::7",
	})
	if err != nil {
		t.Fatalf("buildInteractiveResult returned error: %v", err)
	}

	if result.RemoteFlag != "[2001:db8::20]:443" {
		t.Fatalf("RemoteFlag = %q", result.RemoteFlag)
	}
	if result.RoutesFlag != "8443:[2001:db8::20]:443" {
		t.Fatalf("RoutesFlag = %q", result.RoutesFlag)
	}
	if !reflect.DeepEqual(result.AllowFlags, []string{"2001:db8::7"}) {
		t.Fatalf("AllowFlags = %#v", result.AllowFlags)
	}
}

func TestPromptWithDefaultShowsCurrentValue(t *testing.T) {
	prompt := promptWithDefault("3) Local port", "8080")
	if prompt != "3) Local port [8080]: " {
		t.Fatalf("promptWithDefault = %q", prompt)
	}
}

func TestAskReviewChoiceAcceptsExitWithoutSaving(t *testing.T) {
	choice, err := askReviewChoice(bufio.NewReader(strings.NewReader("7\n")))
	if err != nil {
		t.Fatalf("askReviewChoice returned error: %v", err)
	}
	if choice != "7" {
		t.Fatalf("choice = %q, want 7", choice)
	}
}

func TestAskReviewChoiceDefaultsToSaveAndContinue(t *testing.T) {
	choice, err := askReviewChoice(bufio.NewReader(strings.NewReader("\n")))
	if err != nil {
		t.Fatalf("askReviewChoice returned error: %v", err)
	}
	if choice != "6" {
		t.Fatalf("choice = %q, want 6", choice)
	}
}

func TestSetupCommandTextShowsDifferentLocalAndRemotePorts(t *testing.T) {
	result, err := buildInteractiveResult("chicha-ip-proxy", setupDraft{
		TargetIP:   "203.0.113.20",
		RemotePort: "80",
		LocalPort:  "8080",
		Protocol:   "tcp",
		AllowRaw:   "198.51.100.7",
	})
	if err != nil {
		t.Fatalf("buildInteractiveResult returned error: %v", err)
	}

	commandText := setupCommandText(result)
	for _, want := range []string{"-local=8080", "-remote=203.0.113.20:80", "-proto=tcp", "-allow=198.51.100.7"} {
		if !strings.Contains(commandText, want) {
			t.Fatalf("setupCommandText missing %q: %s", want, commandText)
		}
	}
}

func TestFormatLocalPortStatus(t *testing.T) {
	freeStatus := formatLocalPortStatus(localPortStatus{Protocol: "tcp", Available: true})
	if !strings.Contains(freeStatus, "TCP: free") {
		t.Fatalf("free status text = %q", freeStatus)
	}

	busyStatus := formatLocalPortStatus(localPortStatus{Protocol: "udp", Err: errors.New("address already in use")})
	if !strings.Contains(busyStatus, "UDP: busy or unavailable") {
		t.Fatalf("busy status text = %q", busyStatus)
	}
}
