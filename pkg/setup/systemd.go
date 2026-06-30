// Package setup also contains systemd integration helpers.
// Keeping service creation alongside interactive prompts keeps first-run automation cohesive.
package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SystemdResult captures whether the operator asked to stream logs immediately.
// Returning this to main lets the caller decide how to continue running the proxy.
type SystemdResult struct {
	FollowLogs bool
}

// OfferSystemdSetup proposes creating, enabling, and starting a systemd unit.
// The function keeps user prompts sequential while delegating long-running work to goroutines where useful.
func OfferSystemdSetup(appName string, interactive *InteractiveResult, rotation time.Duration) (*SystemdResult, error) {
	reader := bufio.NewReader(os.Stdin)
	unitName := systemdUnitName(interactive.ServiceName)

	createSystemd, err := askYesDefault(reader, fmt.Sprintf("Create a systemd service '%s'?", unitName))
	if err != nil {
		return nil, err
	}
	if !createSystemd {
		return &SystemdResult{FollowLogs: false}, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %v", err)
	}

	unitContent := buildUnitFile(appName, interactive, rotation, executable)
	unitPath := filepath.Join("/etc/systemd/system", unitName)
	if err := os.WriteFile(unitPath, []byte(unitContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write systemd unit: %v", err)
	}

	if err := reloadSystemd(); err != nil {
		return nil, err
	}

	enableSystemd, err := askYesDefault(reader, "Enable the service so it starts on boot?")
	if err != nil {
		return nil, err
	}

	if enableSystemd {
		if err := runSystemctl("enable", unitName); err != nil {
			return nil, err
		}
	}

	startSystemd, err := askYesDefault(reader, "Start the service now?")
	if err != nil {
		return nil, err
	}

	if startSystemd {
		if err := runSystemctl("start", unitName); err != nil {
			return nil, err
		}
	}

	followLogs, err := askYesDefault(reader, "Follow the log file now?")
	if err != nil {
		return nil, err
	}

	return &SystemdResult{FollowLogs: followLogs}, nil
}

// StreamLogs tails the specified file and writes updates to stdout until the stop channel closes.
// Using a channel makes it easy for callers to coordinate shutdown without mutexes.
func StreamLogs(logFile string, stop <-chan struct{}) {
	file, err := os.Open(logFile)
	if err != nil {
		fmt.Printf("Failed to open log file %s: %v\n", logFile, err)
		return
	}
	defer file.Close()

	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		fmt.Printf("Failed to seek log file %s: %v\n", logFile, err)
		return
	}

	reader := bufio.NewReader(file)

	for {
		select {
		case <-stop:
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			fmt.Print(line)
		}
	}
}

// buildUnitFile composes a systemd unit with explicit log file arguments and rotation schedule.
// Embedding the rotation flag keeps the service aligned with interactive defaults.
func buildUnitFile(appName string, interactive *InteractiveResult, rotation time.Duration, executable string) string {
	args := buildArgs(interactive, rotation)

	return fmt.Sprintf(`[Unit]
Description=%s proxy service
After=network.target

[Service]
Type=simple
ExecStart=%s %s
Restart=on-failure

[Install]
WantedBy=multi-user.target
`, appName, executable, strings.Join(args, " "))
}

func systemdUnitName(serviceName string) string {
	if strings.HasSuffix(serviceName, ".service") {
		return serviceName
	}
	return serviceName + ".service"
}

// reloadSystemd triggers a daemon-reload to pick up newly written units.
// Having it as a helper keeps OfferSystemdSetup easy to read.
func reloadSystemd() error {
	return runSystemctl("daemon-reload")
}

// runSystemctl executes systemctl with the provided arguments.
// Using exec.Command avoids shell parsing while still keeping the function concise.
func runSystemctl(args ...string) error {
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s failed: %v - %s", strings.Join(args, " "), err, string(output))
	}
	return nil
}
