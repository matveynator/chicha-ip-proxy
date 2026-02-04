// Package setup contains helpers for boot-time autostart configuration.
// Keeping autostart logic here keeps the main package focused on runtime wiring.
package setup

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	fmt.Printf("Would you like to create a legacy init script for '%s'? (y/N): ", interactive.ServiceName)
	createAnswer, err := readTrimmed(reader)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(createAnswer) != "y" {
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

	fmt.Print("Enable the init script so it starts on boot? (y/N): ")
	enableAnswer, err := readTrimmed(reader)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(enableAnswer) == "y" {
		if err := enableInitScript(initName); err != nil {
			return nil, err
		}
	}

	fmt.Print("Start the service now? (y/N): ")
	startAnswer, err := readTrimmed(reader)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(startAnswer) == "y" {
		if err := runInitCommand(initName, "start"); err != nil {
			return nil, err
		}
	}

	fmt.Print("Follow the log file now? (y/N): ")
	followAnswer, err := readTrimmed(reader)
	if err != nil {
		return nil, err
	}

	return &SystemdResult{FollowLogs: strings.ToLower(followAnswer) == "y"}, nil
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

	return fmt.Sprintf(`#!/bin/sh
### BEGIN INIT INFO
# Provides:          %s
# Required-Start:    $network
# Required-Stop:     $network
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: %s proxy service
### END INIT INFO

APP_NAME="%s"
EXEC="%s"
ARGS="%s"
PIDFILE="/var/run/%s.pid"

start() {
  if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
    echo "$APP_NAME is already running"
    return 0
  fi
  echo "Starting $APP_NAME"
  nohup "$EXEC" $ARGS >/dev/null 2>&1 &
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
`, initName, appName, appName, executable, strings.Join(args, " "), initName)
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
	if interactive.RoutesFlag != "" {
		args = append(args, fmt.Sprintf("-routes=%s", interactive.RoutesFlag))
	}
	if interactive.UDPRoutesFlag != "" {
		args = append(args, fmt.Sprintf("-udp-routes=%s", interactive.UDPRoutesFlag))
	}
	args = append(args, fmt.Sprintf("-log=%s", interactive.LogFile))
	args = append(args, fmt.Sprintf("-rotation=%s", rotation.String()))
	return args
}
