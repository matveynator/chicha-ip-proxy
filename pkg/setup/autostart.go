// Package setup contains helpers for boot-time autostart configuration.
// Keeping autostart logic here keeps the main package focused on runtime wiring.
package setup

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// linuxInfo keeps distribution details so we can surface them while deciding on init systems.
// Capturing this data makes the operator aware of what we detected without guessing.
type linuxInfo struct {
	ID        string
	VersionID string
}

// ----- Autostart entrypoints -----

// OfferAutostartSetup selects the appropriate init system and guides the operator through setup.
// The function keeps user prompts sequential while delegating long-running work to helpers.
func OfferAutostartSetup(appName string, interactive *InteractiveResult, rotation time.Duration) (*SystemdResult, error) {
	reader := bufio.NewReader(os.Stdin)
	if err := validateAutostartName(interactive.ServiceName); err != nil {
		return nil, err
	}

	switch runtime.GOOS {
	case "linux":
		return offerLinuxAutostartSetup(appName, interactive, rotation, reader)
	case "darwin":
		return OfferLaunchdSetup(appName, interactive, rotation, reader)
	case "freebsd":
		return OfferBSDRCSetup(appName, interactive, rotation, reader, "freebsd")
	case "openbsd":
		return OfferBSDRCSetup(appName, interactive, rotation, reader, "openbsd")
	case "windows":
		return OfferWindowsTaskSetup(appName, interactive, rotation, reader)
	default:
		fmt.Printf("No supported autostart integration for %s; skipping autostart configuration.\n", runtime.GOOS)
		return &SystemdResult{FollowLogs: false}, nil
	}
}

func offerLinuxAutostartSetup(appName string, interactive *InteractiveResult, rotation time.Duration, reader *bufio.Reader) (*SystemdResult, error) {
	info := readLinuxInfo()
	if info.ID != "" || info.VersionID != "" {
		fmt.Printf("Detected Linux distribution: %s %s\n", info.ID, info.VersionID)
	}

	systemdAvailable := isSystemdAvailable()
	initAvailable := isInitAvailable()

	if systemdAvailable {
		fmt.Println("Systemd detected, offering systemd autostart setup.")
		return OfferSystemdSetup(appName, interactive, rotation)
	}

	if initAvailable {
		fmt.Println("Systemd not found, using legacy init script setup.")
		return OfferInitSetup(appName, interactive, rotation, reader)
	}

	fmt.Println("No supported init system detected; skipping autostart configuration.")
	return &SystemdResult{FollowLogs: false}, nil
}

// ----- Legacy init workflow -----

// OfferInitSetup creates a SysV-style init script and optionally enables and starts it.
// Using a shared reader keeps the input flow consistent with systemd setup.
func OfferInitSetup(appName string, interactive *InteractiveResult, rotation time.Duration, reader *bufio.Reader) (*SystemdResult, error) {
	createInit, err := askYesDefault(reader, fmt.Sprintf("Create a legacy init script for '%s'?", interactive.ServiceName))
	if err != nil {
		return nil, err
	}
	if !createInit {
		return &SystemdResult{FollowLogs: false}, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %v", err)
	}

	initName := initServiceName(interactive.ServiceName)
	scriptContent := buildInitScript(appName, interactive, rotation, executable, initName)
	scriptPath := filepath.Join("/etc/init.d", initName)
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return nil, fmt.Errorf("failed to write init script: %v", err)
	}

	enableInit, err := askYesDefault(reader, "Enable the init script so it starts on boot?")
	if err != nil {
		return nil, err
	}

	if enableInit {
		if err := enableInitScript(initName); err != nil {
			return nil, err
		}
	}

	startInit, err := askYesDefault(reader, "Start the service now?")
	if err != nil {
		return nil, err
	}

	if startInit {
		if err := runInitCommand(initName, "start"); err != nil {
			return nil, err
		}
	}

	followLogs, err := askYesDefault(reader, "Follow the log file now?")
	if err != nil {
		return nil, err
	}

	return &SystemdResult{FollowLogs: followLogs}, nil
}

// ----- macOS launchd workflow -----

// OfferLaunchdSetup creates a LaunchDaemon plist and optionally bootstraps it.
func OfferLaunchdSetup(appName string, interactive *InteractiveResult, rotation time.Duration, reader *bufio.Reader) (*SystemdResult, error) {
	createLaunchd, err := askYesDefault(reader, fmt.Sprintf("Create a macOS launchd daemon '%s'?", interactive.ServiceName))
	if err != nil {
		return nil, err
	}
	if !createLaunchd {
		return &SystemdResult{FollowLogs: false}, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %v", err)
	}

	plistPath := filepath.Join("/Library/LaunchDaemons", interactive.ServiceName+".plist")
	plistContent := buildLaunchdPlist(appName, interactive, rotation, executable)
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write launchd plist: %v", err)
	}

	enableLaunchd, err := askYesDefault(reader, "Enable the launchd daemon so it starts on boot?")
	if err != nil {
		return nil, err
	}
	if enableLaunchd {
		if err := runCommand("launchctl", "bootstrap", "system", plistPath); err != nil {
			return nil, err
		}
	}

	startLaunchd, err := askYesDefault(reader, "Start the daemon now?")
	if err != nil {
		return nil, err
	}
	if startLaunchd {
		if err := runCommand("launchctl", "kickstart", "-k", "system/"+interactive.ServiceName); err != nil {
			return nil, err
		}
	}

	followLogs, err := askYesDefault(reader, "Follow the log file now?")
	if err != nil {
		return nil, err
	}
	return &SystemdResult{FollowLogs: followLogs}, nil
}

// ----- BSD rc.d workflow -----

// OfferBSDRCSetup creates an rc.d script for FreeBSD or OpenBSD and optionally enables it.
func OfferBSDRCSetup(appName string, interactive *InteractiveResult, rotation time.Duration, reader *bufio.Reader, osName string) (*SystemdResult, error) {
	createRC, err := askYesDefault(reader, fmt.Sprintf("Create a %s rc.d service '%s'?", osName, interactive.ServiceName))
	if err != nil {
		return nil, err
	}
	if !createRC {
		return &SystemdResult{FollowLogs: false}, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %v", err)
	}

	scriptPath := bsdRCPath(interactive.ServiceName, osName)
	scriptContent := buildBSDRCScript(appName, interactive, rotation, executable, osName)
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return nil, fmt.Errorf("failed to write rc.d script: %v", err)
	}

	enableRC, err := askYesDefault(reader, "Enable the rc.d service so it starts on boot?")
	if err != nil {
		return nil, err
	}
	if enableRC {
		if err := enableBSDRC(interactive.ServiceName, osName); err != nil {
			return nil, err
		}
	}

	startRC, err := askYesDefault(reader, "Start the service now?")
	if err != nil {
		return nil, err
	}
	if startRC {
		if err := startBSDRC(interactive.ServiceName, osName); err != nil {
			return nil, err
		}
	}

	followLogs, err := askYesDefault(reader, "Follow the log file now?")
	if err != nil {
		return nil, err
	}
	return &SystemdResult{FollowLogs: followLogs}, nil
}

// ----- Windows Task Scheduler workflow -----

// OfferWindowsTaskSetup uses Task Scheduler because console binaries are not native Windows services.
func OfferWindowsTaskSetup(appName string, interactive *InteractiveResult, rotation time.Duration, reader *bufio.Reader) (*SystemdResult, error) {
	createTask, err := askYesDefault(reader, fmt.Sprintf("Create a Windows startup task '%s'?", interactive.ServiceName))
	if err != nil {
		return nil, err
	}
	if !createTask {
		return &SystemdResult{FollowLogs: false}, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %v", err)
	}

	command := windowsTaskCommand(executable, buildArgs(interactive, rotation))
	if err := runCommand("schtasks", "/Create", "/F", "/TN", interactive.ServiceName, "/SC", "ONSTART", "/RU", "SYSTEM", "/RL", "HIGHEST", "/TR", command); err != nil {
		return nil, err
	}

	startTask, err := askYesDefault(reader, "Start the task now?")
	if err != nil {
		return nil, err
	}
	if startTask {
		if err := runCommand("schtasks", "/Run", "/TN", interactive.ServiceName); err != nil {
			return nil, err
		}
	}

	followLogs, err := askYesDefault(reader, "Follow the log file now?")
	if err != nil {
		return nil, err
	}
	return &SystemdResult{FollowLogs: followLogs}, nil
}

// ----- Detection helpers -----

// readLinuxInfo gathers ID and VERSION_ID from os-release for visibility.
// Reading /etc/os-release is the most portable way to detect Linux distribution details.
func readLinuxInfo() linuxInfo {
	paths := []string{"/etc/os-release", "/usr/lib/os-release"}
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		defer file.Close()

		info := linuxInfo{}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "ID=") {
				info.ID = strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
			}
			if strings.HasPrefix(line, "VERSION_ID=") {
				info.VersionID = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
			}
		}
		return info
	}

	return linuxInfo{}
}

// ----- Init system probes -----

// isSystemdAvailable checks systemd presence via runtime paths and binaries.
// Looking for both a runtime directory and systemctl makes detection more reliable.
func isSystemdAvailable() bool {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	if _, err := exec.LookPath("systemctl"); err == nil {
		return true
	}
	return false
}

// isInitAvailable checks for a legacy init system using common paths.
// We keep the detection conservative to avoid writing scripts on unsupported systems.
func isInitAvailable() bool {
	if _, err := os.Stat("/sbin/init"); err == nil {
		if _, err := os.Stat("/etc/init.d"); err == nil {
			return true
		}
	}
	return false
}

// ----- Script builders -----

// initServiceName removes a systemd suffix for init script naming.
// Stripping the suffix keeps SysV script names aligned with common conventions.
func initServiceName(serviceName string) string {
	return strings.TrimSuffix(serviceName, ".service")
}

// buildInitScript renders a SysV-style init script with start/stop commands.
// Using a pidfile keeps lifecycle management simple without extra dependencies.
func buildInitScript(appName string, interactive *InteractiveResult, rotation time.Duration, executable, initName string) string {
	args := buildArgs(interactive, rotation)
	commandArgs := shellJoin(args)

	return fmt.Sprintf(`#!/bin/sh
### BEGIN INIT INFO
# Provides:          %s
# Required-Start:    $network
# Required-Stop:     $network
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: %s proxy service
### END INIT INFO

APP_NAME=%s
EXEC=%s
PIDFILE=%s

start() {
  if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
    echo "$APP_NAME is already running"
    return 0
  fi
  echo "Starting $APP_NAME"
  nohup "$EXEC" %s >/dev/null 2>&1 &
  echo $! > "$PIDFILE"
}

stop() {
  if [ ! -f "$PIDFILE" ]; then
    echo "$APP_NAME is not running"
    return 0
  fi
  PID="$(cat "$PIDFILE")"
  if kill -0 "$PID" 2>/dev/null; then
    echo "Stopping $APP_NAME"
    kill "$PID"
  fi
  rm -f "$PIDFILE"
}

case "$1" in
  start)
    start
    ;;
  stop)
    stop
    ;;
  restart)
    stop
    start
    ;;
  status)
    if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
      echo "$APP_NAME is running"
    else
      echo "$APP_NAME is stopped"
    fi
    ;;
  *)
    echo "Usage: $0 {start|stop|restart|status}"
    exit 1
    ;;
esac

exit 0
`, initName, appName, shellQuote(appName), shellQuote(executable), shellQuote(filepath.Join("/var/run", initName+".pid")), commandArgs)
}

// buildLaunchdPlist renders a LaunchDaemon with explicit arguments instead of shell parsing.
func buildLaunchdPlist(appName string, interactive *InteractiveResult, rotation time.Duration, executable string) string {
	args := buildArgs(interactive, rotation)
	arguments := make([]string, 0, len(args)+1)
	arguments = append(arguments, executable)
	arguments = append(arguments, args...)

	lines := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		lines = append(lines, fmt.Sprintf("    <string>%s</string>", xmlEscape(argument)))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
%s
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>%s</string>
  <key>StandardErrorPath</key>
  <string>%s</string>
</dict>
</plist>
`, xmlEscape(interactive.ServiceName), strings.Join(lines, "\n"), xmlEscape(interactive.LogFile), xmlEscape(interactive.LogFile))
}

// buildBSDRCScript renders a small rc.d script for systems that use rc.subr.
func buildBSDRCScript(appName string, interactive *InteractiveResult, rotation time.Duration, executable, osName string) string {
	name := shellIdentifier(interactive.ServiceName)
	args := buildArgs(interactive, rotation)

	if osName == "openbsd" {
		return fmt.Sprintf(`#!/bin/ksh

daemon=%s
daemon_flags=%s

. /etc/rc.d/rc.subr

rc_cmd $1
`, shellQuote(executable), shellQuote(strings.Join(args, " ")))
	}

	return fmt.Sprintf(`#!/bin/sh

# PROVIDE: %s
# REQUIRE: NETWORKING
# KEYWORD: shutdown

. /etc/rc.subr

name="%s"
rcvar="%s_enable"
command=%s
command_args=%s
pidfile=%s

load_rc_config "$name"
: ${%s_enable:="NO"}

run_rc_command "$1"
`, name, name, name, shellQuote(executable), shellQuote(strings.Join(args, " ")), shellQuote(filepath.Join("/var/run", name+".pid")), name)
}

func bsdRCPath(serviceName, osName string) string {
	name := shellIdentifier(serviceName)
	if osName == "freebsd" {
		return filepath.Join("/usr/local/etc/rc.d", name)
	}
	return filepath.Join("/etc/rc.d", name)
}

func enableBSDRC(serviceName, osName string) error {
	name := shellIdentifier(serviceName)
	if osName == "freebsd" {
		return runCommand("sysrc", name+"_enable=YES")
	}
	return runCommand("rcctl", "enable", name)
}

func startBSDRC(serviceName, osName string) error {
	name := shellIdentifier(serviceName)
	if osName == "freebsd" {
		return runCommand("service", name, "start")
	}
	return runCommand("rcctl", "start", name)
}

func windowsTaskCommand(executable string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, windowsCommandQuote(executable))
	for _, arg := range args {
		parts = append(parts, windowsCommandQuote(arg))
	}
	return strings.Join(parts, " ")
}

func windowsCommandQuote(value string) string {
	escaped := strings.ReplaceAll(value, `"`, `\"`)
	return `"` + escaped + `"`
}

func shellIdentifier(value string) string {
	replacer := strings.NewReplacer("-", "_", ".", "_")
	return replacer.Replace(value)
}

func validateAutostartName(value string) error {
	if value == "" {
		return fmt.Errorf("autostart name cannot be empty")
	}
	for _, char := range value {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= 'A' && char <= 'Z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '-' || char == '_' || char == '.' {
			continue
		}
		return fmt.Errorf("autostart name %q contains unsafe character %q", value, char)
	}
	return nil
}

func shellJoin(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, shellQuote(value))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func systemdJoin(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, systemdQuote(value))
	}
	return strings.Join(quoted, " ")
}

func systemdQuote(value string) string {
	if value == "" {
		return `""`
	}
	if strings.IndexFunc(value, func(char rune) bool {
		return !(char == '-' || char == '_' || char == '.' || char == '/' || char == ':' || char == '=' || char == ',' || char == '+' || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9'))
	}) == -1 {
		return value
	}
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

func xmlEscape(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(value)
}

// ----- Init system integration -----

// enableInitScript hooks the script into the default runlevels using available tools.
// Supporting both update-rc.d and chkconfig keeps compatibility across distributions.
func enableInitScript(initName string) error {
	if _, err := exec.LookPath("update-rc.d"); err == nil {
		return runCommand("update-rc.d", initName, "defaults")
	}
	if _, err := exec.LookPath("chkconfig"); err == nil {
		if err := runCommand("chkconfig", "--add", initName); err != nil {
			return err
		}
		return runCommand("chkconfig", initName, "on")
	}
	return fmt.Errorf("no init enablement tool found (update-rc.d or chkconfig)")
}

// runInitCommand executes the init script with the provided action.
// Using exec.Command avoids shell interpretation while keeping output available.
func runInitCommand(initName, action string) error {
	return runCommand(filepath.Join("/etc/init.d", initName), action)
}

// ----- Command execution -----

// runCommand executes a command and surfaces combined output on failure.
// Returning detailed errors makes it easier for operators to diagnose issues.
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %v - %s", name, strings.Join(args, " "), err, string(output))
	}
	return nil
}

// ----- Shared argument builder -----

// buildArgs renders CLI flags for systemd or init scripts.
// Having a single formatter ensures consistent startup arguments.
func buildArgs(interactive *InteractiveResult, rotation time.Duration) []string {
	args := make([]string, 0)
	if interactive.LocalFlag != "" && interactive.RemoteFlag != "" {
		args = append(args, fmt.Sprintf("-local=%s", interactive.LocalFlag))
		args = append(args, fmt.Sprintf("-remote=%s", interactive.RemoteFlag))
		if interactive.ProtoFlag != "" && interactive.ProtoFlag != "tcp" {
			args = append(args, fmt.Sprintf("-proto=%s", interactive.ProtoFlag))
		}
	} else {
		if interactive.RoutesFlag != "" {
			args = append(args, fmt.Sprintf("-routes=%s", interactive.RoutesFlag))
		}
		if interactive.UDPRoutesFlag != "" {
			args = append(args, fmt.Sprintf("-udp-routes=%s", interactive.UDPRoutesFlag))
		}
	}
	for _, allowValue := range interactive.AllowFlags {
		args = append(args, fmt.Sprintf("-allow=%s", allowValue))
	}
	args = append(args, fmt.Sprintf("-log=%s", interactive.LogFile))
	args = append(args, fmt.Sprintf("-rotation=%s", rotation.String()))
	return args
}

// askYesDefault keeps destructive-looking setup prompts explicit while matching the installer's happy path.
func askYesDefault(reader *bufio.Reader, prompt string) (bool, error) {
	fmt.Print(colorize(greenText, prompt+" (Y/n): "))
	answer, err := readTrimmed(reader)
	if err != nil {
		return false, err
	}
	return strings.ToLower(answer) != "n", nil
}
