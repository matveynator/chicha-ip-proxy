// Package setup provides service registration workflows.
// Keeping service detection here centralizes systemd and init decisions for interactive runs.
package setup

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ServiceKind describes the init system detected on Linux.
// Explicit values keep downstream logic readable when branching for systemd or SysV init.
type ServiceKind string

const (
	serviceKindSystemd ServiceKind = "systemd"
	serviceKindSysV    ServiceKind = "sysvinit"
)

// ServiceResult captures the init system choice and whether to follow logs.
// Returning this lets main decide if it should start tailing the log output.
type ServiceResult struct {
	FollowLogs bool
	Kind       ServiceKind
}

// OfferServiceSetup chooses systemd or init script generation based on the host.
// Announcing the choice up front keeps operators informed about what will happen next.
func OfferServiceSetup(appName string, interactive *InteractiveResult, rotation time.Duration) (*ServiceResult, error) {
	kind := detectServiceKind()
	announceServicePlan(kind, interactive.ServiceName)

	switch kind {
	case serviceKindSystemd:
		result, err := offerSystemdSetup(appName, interactive, rotation)
		if err != nil {
			return nil, err
		}
		return &ServiceResult{FollowLogs: result.FollowLogs, Kind: kind}, nil
	case serviceKindSysV:
		result, err := offerInitSetup(appName, interactive, rotation)
		if err != nil {
			return nil, err
		}
		return &ServiceResult{FollowLogs: result.FollowLogs, Kind: kind}, nil
	default:
		return nil, fmt.Errorf("unsupported init system on %s", runtime.GOOS)
	}
}

// detectServiceKind inspects the host to decide whether systemd is available.
// Checking multiple signals avoids false negatives when systemctl is not in PATH.
func detectServiceKind() ServiceKind {
	if runtime.GOOS != "linux" {
		return serviceKindSysV
	}

	if _, err := exec.LookPath("systemctl"); err == nil {
		return serviceKindSystemd
	}

	if info, err := os.Stat("/run/systemd/system"); err == nil && info.IsDir() {
		return serviceKindSystemd
	}

	return serviceKindSysV
}

// announceServicePlan prints whether a new or legacy init system is in use and what will be created.
// Sharing the plan before prompts makes the setup more transparent for operators.
func announceServicePlan(kind ServiceKind, serviceName string) {
	switch kind {
	case serviceKindSystemd:
		fmt.Printf("Detected a new systemd-based Linux system. I will offer to create '%s'.\n", serviceName)
	case serviceKindSysV:
		fmt.Printf("Detected a legacy init system. I will offer to create a SysV init script for '%s'.\n", strings.TrimSuffix(serviceName, ".service"))
	default:
		fmt.Println("Detected an unknown init system. No automatic service setup will run.")
	}
}
