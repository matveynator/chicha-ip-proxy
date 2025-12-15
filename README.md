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

---

## Quick Summary

* **Linux** → one command, full auto-setup
* **macOS / BSD / Windows** → run with flags manually
* **Fast, minimal, production-ready**
