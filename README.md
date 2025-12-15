<img src="https://github.com/matveynator/chicha-ip-proxy/blob/master/chicha-ip-proxy.png?raw=true" width="100%" align="right" />

# Chicha IP Proxy

**Chicha IP Proxy** is a lightweight, high-performance **Layer 2 (L2) IP proxy** written in **Go**.
It forwards TCP and UDP ports with minimal overhead and zero configuration files.

Designed as a fast and simple alternative to tools like `xinetd`.

---

## What You Get

* 🚀 **Very fast** — Go, multi-core, low latency
* 🧩 **Zero config files** — CLI only
* 🔁 **TCP & UDP support**
* 🛠 **Interactive auto-configuration**
* 📦 **systemd integration**
* 📝 **Built-in log rotation**
* 🖥 **Single static binary**

---

## Typical Use Cases

* Mirror a website (80 / 443)
* Forward game or VPN ports (UDP)
* Replace `xinetd` / `iptables` port forwarding
* Quickly expose internal services

---

## Download

Download and setup the latest Linux binary:

```bash
curl -L https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64 -o chicha-ip-proxy;
install -m 0755 chicha-ip-proxy /usr/local/bin/chicha-ip-proxy;
/usr/local/bin/chicha-ip-proxy;
```

That’s it. No dependencies.

---

## Quick Setup (Recommended)

Just run the binary **without flags**.

```bash
chicha-ip-proxy
```

### What happens next?

An interactive wizard starts automatically and asks a few simple questions:

* target IP address
* protocol (TCP / UDP)
* local ports
* systemd service creation
* enable on boot
* start now
* follow logs

### Example Session

```
No routes provided via flags. Starting interactive configuration...

Enter target IP address to proxy to: 203.0.113.44
Enter protocols (tcp, udp): tcp
Enter local TCP ports: 80,443

Planned log file:
  /var/log/chicha-ip-proxy-tcp-443-80.log

Planned systemd service:
  chicha-ip-proxy-tcp-443-80.service

Create systemd service? (y/N): y
Enable on boot? (y/N): y
Start now? (y/N): y
Follow logs now? (y/N): y
```

---

## Result

After the wizard finishes:

✅ Ports are forwarded
✅ systemd service is running
✅ Logs are written and rotated daily
✅ Service starts automatically on reboot

You don’t need to touch anything else.

---

## Run Without Wizard (Flags Only)

### TCP Port Forwarding

```bash
chicha-ip-proxy \
  -routes "80:198.51.100.5:80,443:198.51.100.5:443" \
  -log /var/log/chicha-ip-proxy.log
```

### UDP Port Forwarding (example: WireGuard)

```bash
chicha-ip-proxy \
  -udp-routes "51820:198.51.100.20:51820" \
  -log /var/log/chicha-ip-proxy.log
```

---

## Logs

* Logs are **plain text**
* Rotated automatically
* Default rotation: **every 24 hours**
* No compression (easy to grep)

Example:

```
/var/log/chicha-ip-proxy-tcp-443-80.log
```

---

## Command Reference

```bash
chicha-ip-proxy --help
```

```
-log string
    Path to log file (default: chicha-ip-proxy.log)

-rotation duration
    Log rotation interval (default: 24h)

-routes string
    TCP routes:
    LOCALPORT:REMOTEIP:REMOTEPORT[, ...]

-udp-routes string
    UDP routes:
    LOCALPORT:REMOTEIP:REMOTEPORT[, ...]

-version
    Print version and exit
```

---

## Performance

Chicha IP Proxy is designed to be as close as possible to direct traffic.

### Benchmark Summary

| Method          | Requests/sec | Response Time |
| --------------- | ------------ | ------------- |
| Direct requests | ~1045        | ~0.03s        |
| chicha-ip-proxy | ~1073        | ~0.03s        |
| xinetd          | ~1020        | ~0.04s        |



