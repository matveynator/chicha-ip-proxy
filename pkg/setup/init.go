// Package setup also provisions legacy SysV init scripts when systemd is absent.
// Keeping this separate from systemd helpers clarifies the two distinct boot systems.
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

// offerInitSetup asks permission to generate an init script and optionally enable it.
// The workflow stays linear so operators can see each step before it runs.
func offerInitSetup(appName string, interactive *InteractiveResult, rotation time.Duration) (*SystemdResult, error) {
	reader := bufio.NewReader(os.Stdin)
	initName := initServiceName(interactive.ServiceName)

	fmt.Printf("Would you like to create a SysV init script '%s'? (y/N): ", initName)
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

	initContent := buildInitScript(appName, initName, interactive, rotation, executable)
	initPath := filepath.Join("/etc/init.d", initName)
	if err := os.WriteFile(initPath, []byte(initContent), 0755); err != nil {
		return nil, fmt.Errorf("failed to write init script: %v", err)
	}

	fmt.Print("Enable the init script so it starts on boot? (y/N): ")
	enableAnswer, err := readTrimmed(reader)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(enableAnswer) == "y" {
		if err := registerInitScript(initName); err != nil {
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

// initServiceName ensures the legacy init script name is a simple basename.
// SysV init directories expect plain script names without a systemd suffix.
func initServiceName(serviceName string) string {
	return strings.TrimSuffix(serviceName, ".service")
}

// buildInitScript renders a POSIX shell script with start/stop/status commands.
// Writing the script in one place keeps the args aligned with the systemd unit.
func buildInitScript(appName, initName string, interactive *InteractiveResult, rotation time.Duration, executable string) string {
	args := buildServiceArgs(interactive, rotation)
	escapedArgs := strings.Join(args, " ")

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
EXECUTABLE="%s"
ARGS="%s"
PIDFILE="/var/run/%s.pid"
LOGFILE="%s"

start_service() {
	if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
		echo "$APP_NAME is already running"
		return 0
	fi

	echo "Starting $APP_NAME..."
	nohup "$EXECUTABLE" $ARGS >> "$LOGFILE" 2>&1 &
	echo $! > "$PIDFILE"
}

stop_service() {
	if [ -f "$PIDFILE" ]; then
		PID="$(cat "$PIDFILE")"
		if kill -0 "$PID" 2>/dev/null; then
			echo "Stopping $APP_NAME..."
			kill "$PID"
		fi
		rm -f "$PIDFILE"
	else
		echo "$APP_NAME is not running"
	fi
}

status_service() {
	if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
		echo "$APP_NAME is running"
		exit 0
	fi
	echo "$APP_NAME is not running"
	exit 1
}

case "$1" in
	start)
		start_service
		;;
	stop)
		stop_service
		;;
	restart)
		stop_service
		start_service
		;;
	status)
		status_service
		;;
	*)
		echo "Usage: $0 {start|stop|restart|status}"
		exit 1
		;;
esac

exit 0
`, initName, appName, appName, executable, escapedArgs, initName, interactive.LogFile)
}

// registerInitScript adds the script to the default runlevels.
// Trying update-rc.d first covers Debian-style systems, with chkconfig as a fallback.
func registerInitScript(initName string) error {
	if _, err := exec.LookPath("update-rc.d"); err == nil {
		return runCommand("update-rc.d", initName, "defaults")
	}
	if _, err := exec.LookPath("chkconfig"); err == nil {
		if err := runCommand("chkconfig", "--add", initName); err != nil {
			return err
		}
		return runCommand("chkconfig", initName, "on")
	}
	return fmt.Errorf("no init registration tool found (update-rc.d or chkconfig)")
}

// runInitCommand executes the init script via service or direct invocation.
// This keeps compatibility with distributions that still ship the legacy service wrapper.
func runInitCommand(initName, action string) error {
	if _, err := exec.LookPath("service"); err == nil {
		return runCommand("service", initName, action)
	}
	return runCommand(filepath.Join("/etc/init.d", initName), action)
}

// runCommand wraps exec.Command with combined output for error reporting.
// Centralizing this avoids duplicated error formatting in service helpers.
func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %v - %s", command, strings.Join(args, " "), err, string(output))
	}
	return nil
}
