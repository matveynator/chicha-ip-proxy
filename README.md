# Chicha IP Proxy

<img width="100%"  alt="chicha-ip-proxy" src="https://github.com/user-attachments/assets/0dc3b863-583f-41ea-b6a4-cfcf773a249c" />

Small TCP/UDP port proxy with an automatic setup wizard.

Run it without arguments, answer a few questions, and it can install itself into the right autostart system for your operating system.

```bash
chicha-ip-proxy
```

No config files are required.

## What It Does

Chicha IP Proxy opens a local port and forwards traffic to another IP address.

Common uses:

* expose a service from another server through this machine
* forward one public port to one private IP
* create a simple TCP or UDP relay
* restrict who can connect by source IP
* install a persistent proxy service without writing system files by hand

## Automatic Setup

Start the binary with no route flags:

```bash
chicha-ip-proxy
```

The wizard asks for:

1. target IP
2. remote port
3. local port, default is the same as remote port
4. protocol, default `tcp`
5. allowed client IPs or CIDR ranges, optional

When you enter the local port, the wizard immediately checks whether that port is free or already busy for TCP and UDP.

Then it shows a simple connection diagram. You can change any field, save and continue, press `Ctrl+C`, or choose exit without saving.

By default, the wizard can:

* create an autostart entry
* enable it on boot
* start it immediately
* follow the log file

Supported autostart systems:

* Linux: `systemd` or SysV init
* macOS: `launchd`
* FreeBSD: `rc.d`
* OpenBSD: `rc.d` / `rcctl`
* Windows: Task Scheduler

## Quick Examples

Forward local TCP port `8080` to `203.0.113.10:8080`:

```bash
chicha-ip-proxy -local=8080 -remote=203.0.113.10
```

Forward local TCP port `8080` to remote port `80`:

```bash
chicha-ip-proxy -local=8080 -remote=203.0.113.10:80
```

Forward UDP DNS traffic:

```bash
chicha-ip-proxy -local=5353 -remote=203.0.113.20:53 -proto=udp
```

Allow only one client IP:

```bash
chicha-ip-proxy -local=8080 -remote=203.0.113.10 -allow=198.51.100.7
```

Allow a whole network:

```bash
chicha-ip-proxy -local=8080 -remote=203.0.113.10 -allow=10.0.0.0/24
```

Allow several sources:

```bash
chicha-ip-proxy -local=8080 -remote=203.0.113.10 -allow=198.51.100.7 -allow=10.0.0.0/24
```

If no `-allow` flag is provided, all client IPs are allowed.

## Install On Linux

Run as root:

```bash
curl -L https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64 > /usr/local/bin/chicha-ip-proxy
chmod +x /usr/local/bin/chicha-ip-proxy
chicha-ip-proxy
```

The wizard will guide the rest.

## Flags

```text
-local PORT
    Local port to listen on.

-remote IP[:PORT]
    Remote target IP and optional port.
    If the port is omitted, the local port is used.

-proto tcp|udp
    Protocol to proxy. Default is tcp.

-allow IP|CIDR
    Source IP or network allowed to connect.
    Repeat the flag to allow multiple sources.

-log PATH
    Log file path.

-rotation DURATION
    Log rotation interval, for example 24h or 1h.
```

## Notes

Denied TCP clients receive a reset. Denied UDP packets are dropped before a proxy session is created.

Legacy `-routes` and `-udp-routes` flags still work for existing multi-route scripts, but new setups should use `-local`, `-remote`, and `-proto`.

## Downloads

Releases are available here:

[https://github.com/matveynator/chicha-ip-proxy/releases/](https://github.com/matveynator/chicha-ip-proxy/releases/)
