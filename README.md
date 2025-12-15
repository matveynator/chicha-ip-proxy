## On **Linux**, everything is automatic.

This single command will:

* download the binary
* install it system-wide
* start interactive setup
* create a **systemd service**
* enable autostart
* start forwarding immediately

### Run as root user:
```bash
curl -L https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64 > /usr/local/bin/chicha-ip-proxy; chmod +x /usr/local/bin/chicha-ip-proxy; /usr/local/bin/chicha-ip-proxy;
```

### What happens next

* Interactive wizard starts
* You choose:

  * target IP
  * TCP / UDP
  * ports
* Proxy automatically:

  * creates a **systemd service**
  * enables it on boot
  * starts it immediately
  * sets up rotating logs

➡️ **No config files. No manual systemd work.**

---

## Download for All Platforms

👉 [https://github.com/matveynator/chicha-ip-proxy/releases/](https://github.com/matveynator/chicha-ip-proxy/releases/)

All binaries are uploaded to the latest stable release, so every script or wiki snippet that downloads from `/releases/latest/download/` always targets the freshest stable build.

## Stable release automation

* A GitHub Action publishes a new release only when the commit message contains the keyword **"stable release"**.
* The release name always includes the word **stable**, keeping the `/latest` download link consistent for automation and wiki scripts.
* Cross-platform builds are produced for macOS, Linux, Windows, FreeBSD, and OpenBSD on `amd64` and `arm64`.

---

## Quick Summary

* **Linux** → one command, full auto-setup
* **macOS / BSD / Windows** → run with flags manually
* **Fast, minimal, production-ready**
