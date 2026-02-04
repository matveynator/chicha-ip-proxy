# Auto-start setup (systemd and SysV init)

This document explains how the interactive setup selects a service manager and what files it writes on Linux.

## Detection

When you run the proxy without `-routes` flags, the interactive setup checks the host system:

- **Systemd available:** The setup offers to create a `*.service` unit in `/etc/systemd/system`.
- **Legacy init:** The setup offers to create a SysV init script in `/etc/init.d`.

The prompt will say which path it will take before asking for confirmation.

## Systemd unit

If you confirm systemd setup:

1. A unit file is written to `/etc/systemd/system/<service>.service`.
2. The setup can optionally run `systemctl enable <service>` to start on boot.
3. The setup can optionally run `systemctl start <service>` to start immediately.

## SysV init script

If the host does not use systemd and you confirm legacy setup:

1. An init script is written to `/etc/init.d/<service>`.
2. The setup tries `update-rc.d` (Debian/Ubuntu) or `chkconfig` (RHEL/CentOS) to register it.
3. The setup can optionally start the script right away.

The script stores its PID in `/var/run/<service>.pid` and redirects stdout/stderr to the log file chosen during interactive setup.
