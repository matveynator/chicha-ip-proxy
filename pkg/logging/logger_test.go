package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupLoggerRejectsSymlinkPath(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.log")
	if err := os.WriteFile(targetPath, []byte("existing"), 0600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	linkPath := filepath.Join(dir, "proxy.log")
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}

	_, _, err := SetupLogger(linkPath)
	if err == nil {
		t.Fatal("SetupLogger accepted a symlink log path")
	}
}

func TestSetupLoggerCreatesPrivateLogFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "proxy.log")
	_, file, err := SetupLogger(logPath)
	if err != nil {
		t.Fatalf("SetupLogger returned error: %v", err)
	}
	defer file.Close()

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("os.Stat returned error: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("log file permissions = %v, want 0600", got)
	}
}
