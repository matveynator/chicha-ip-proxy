package setup

import (
	"strings"
	"testing"
	"time"
)

func TestSystemdUnitNameAddsSuffix(t *testing.T) {
	if got := systemdUnitName("chicha-ip-proxy-tcp-8080"); got != "chicha-ip-proxy-tcp-8080.service" {
		t.Fatalf("systemdUnitName returned %q", got)
	}
}

func TestBuildLaunchdPlistIncludesProgramArguments(t *testing.T) {
	result := &InteractiveResult{
		ServiceName: "com.matveynator.chicha-ip-proxy-tcp-8080",
		LogFile:     "/Library/Logs/chicha-ip-proxy.log",
		LocalFlag:   "8080",
		RemoteFlag:  "203.0.113.20",
		ProtoFlag:   "tcp",
		AllowFlags:  []string{"198.51.100.7"},
	}

	plist := buildLaunchdPlist("chicha-ip-proxy", result, time.Hour, "/usr/local/bin/chicha-ip-proxy")
	for _, want := range []string{
		"<string>com.matveynator.chicha-ip-proxy-tcp-8080</string>",
		"<string>/usr/local/bin/chicha-ip-proxy</string>",
		"<string>-local=8080</string>",
		"<string>-remote=203.0.113.20</string>",
		"<string>-allow=198.51.100.7</string>",
	} {
		if !strings.Contains(plist, want) {
			t.Fatalf("launchd plist missing %q:\n%s", want, plist)
		}
	}
}

func TestBuildBSDRCScriptUsesPlatformSyntax(t *testing.T) {
	result := &InteractiveResult{
		ServiceName: "chicha-ip-proxy-tcp-8080",
		RoutesFlag:  "8080:203.0.113.20:8080",
	}

	freebsd := buildBSDRCScript("chicha-ip-proxy", result, time.Hour, "/usr/local/bin/chicha-ip-proxy", "freebsd")
	if !strings.Contains(freebsd, ". /etc/rc.subr") {
		t.Fatalf("FreeBSD rc script should source /etc/rc.subr:\n%s", freebsd)
	}
	if !strings.Contains(freebsd, `rcvar="chicha_ip_proxy_tcp_8080_enable"`) {
		t.Fatalf("FreeBSD rc script missing rcvar:\n%s", freebsd)
	}

	openbsd := buildBSDRCScript("chicha-ip-proxy", result, time.Hour, "/usr/local/bin/chicha-ip-proxy", "openbsd")
	if !strings.Contains(openbsd, ". /etc/rc.d/rc.subr") {
		t.Fatalf("OpenBSD rc script should source /etc/rc.d/rc.subr:\n%s", openbsd)
	}
	if !strings.Contains(openbsd, "rc_cmd $1") {
		t.Fatalf("OpenBSD rc script missing rc_cmd:\n%s", openbsd)
	}
}

func TestBuildInitScriptShellQuotesArguments(t *testing.T) {
	result := &InteractiveResult{
		ServiceName: "chicha-ip-proxy-tcp-8080",
		LocalFlag:   "8080",
		RemoteFlag:  "203.0.113.20",
		LogFile:     "/tmp/chicha log/proxy.log",
	}

	script := buildInitScript("chicha ip proxy", result, time.Hour, "/usr/local/bin/chicha ip proxy", "chicha-ip-proxy-tcp-8080")
	for _, want := range []string{
		"APP_NAME='chicha ip proxy'",
		"EXEC='/usr/local/bin/chicha ip proxy'",
		"'-log=/tmp/chicha log/proxy.log'",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("init script missing quoted value %q:\n%s", want, script)
		}
	}
}

func TestBuildUnitFileQuotesExecStart(t *testing.T) {
	result := &InteractiveResult{
		ServiceName: "chicha-ip-proxy-tcp-8080",
		LocalFlag:   "8080",
		RemoteFlag:  "203.0.113.20",
		LogFile:     "/tmp/chicha log/proxy.log",
	}

	unit := buildUnitFile("chicha-ip-proxy", result, time.Hour, "/usr/local/bin/chicha ip proxy")
	if !strings.Contains(unit, `ExecStart="/usr/local/bin/chicha ip proxy"`) {
		t.Fatalf("unit file did not quote executable path:\n%s", unit)
	}
	if !strings.Contains(unit, `"-log=/tmp/chicha log/proxy.log"`) {
		t.Fatalf("unit file did not quote log argument:\n%s", unit)
	}
}

func TestValidateAutostartNameRejectsUnsafeCharacters(t *testing.T) {
	if err := validateAutostartName("chicha-ip-proxy.tcp_8080"); err != nil {
		t.Fatalf("validateAutostartName rejected safe name: %v", err)
	}
	if err := validateAutostartName("bad;name"); err == nil {
		t.Fatal("validateAutostartName accepted unsafe name")
	}
}

func TestWindowsTaskCommandQuotesExecutableAndArgs(t *testing.T) {
	command := windowsTaskCommand(`C:\Program Files\chicha\chicha-ip-proxy.exe`, []string{
		"-local=8080",
		"-remote=203.0.113.20",
		"-allow=198.51.100.7",
	})

	want := `"C:\Program Files\chicha\chicha-ip-proxy.exe" "-local=8080" "-remote=203.0.113.20" "-allow=198.51.100.7"`
	if command != want {
		t.Fatalf("windowsTaskCommand = %q, want %q", command, want)
	}
}
