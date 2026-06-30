package main

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/matveynator/chicha-ip-proxy/pkg/config"
)

func TestParseRoutesFromFlagsUsesSimpleFlags(t *testing.T) {
	tcpRoutes, udpRoutes, err := parseRoutesFromFlags("", "", config.SimpleRouteFlags{
		Local:  "8080",
		Remote: "203.0.113.10",
		Proto:  "tcp",
	})
	if err != nil {
		t.Fatalf("parseRoutesFromFlags returned error: %v", err)
	}
	if len(udpRoutes) != 0 {
		t.Fatalf("UDP route count = %d, want 0", len(udpRoutes))
	}
	if len(tcpRoutes) != 1 {
		t.Fatalf("TCP route count = %d, want 1", len(tcpRoutes))
	}
}

func TestParseRoutesFromFlagsKeepsLegacyPrecedence(t *testing.T) {
	tcpRoutes, udpRoutes, err := parseRoutesFromFlags("9000:203.0.113.9:90", "", config.SimpleRouteFlags{
		Local:  "8080",
		Remote: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("parseRoutesFromFlags returned error: %v", err)
	}
	if len(udpRoutes) != 0 {
		t.Fatalf("UDP route count = %d, want 0", len(udpRoutes))
	}
	if len(tcpRoutes) != 1 {
		t.Fatalf("TCP route count = %d, want 1", len(tcpRoutes))
	}
	if tcpRoutes[0].LocalPort != "9000" {
		t.Fatalf("legacy route was not preferred: %#v", tcpRoutes[0])
	}
}

func TestValidateRotationFrequencyRejectsNonPositive(t *testing.T) {
	if err := validateRotationFrequency(time.Hour); err != nil {
		t.Fatalf("validateRotationFrequency rejected positive duration: %v", err)
	}
	if err := validateRotationFrequency(0); err == nil {
		t.Fatal("validateRotationFrequency accepted zero duration")
	}
	if err := validateRotationFrequency(-time.Second); err == nil {
		t.Fatal("validateRotationFrequency accepted negative duration")
	}
}

func TestShowFlagHelpHidesLegacyRouteFlags(t *testing.T) {
	helpOutput := captureStdout(t, showFlagHelp)
	for _, want := range []string{"-local", "-remote", "-proto", "-allow"} {
		if !strings.Contains(helpOutput, want) {
			t.Fatalf("help output missing %q:\n%s", want, helpOutput)
		}
	}
	for _, hidden := range []string{"-routes", "-udp-routes"} {
		if strings.Contains(helpOutput, hidden) {
			t.Fatalf("help output should hide %q:\n%s", hidden, helpOutput)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}

	os.Stdout = writer
	fn()
	writer.Close()
	os.Stdout = originalStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll returned error: %v", err)
	}
	reader.Close()
	return string(output)
}
