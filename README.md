**Chicha IP Proxy** is a lightweight, high-performance **L2 IP proxy** written in **Go**.
It forwards **TCP and UDP ports** with minimal overhead and no configuration files.

Designed as a simple and fast alternative to `xinetd`.

---

## Key Features

* 🚀 **Very fast** — Go, multi-core, low latency
* 🧩 **Zero config files** — CLI only
* 🔁 **TCP & UDP support**
* 🛠 **Interactive auto-setup (Linux only)**
* 📦 **systemd integration**
* 📝 **Built-in log rotation**
* 🖥 **Single static binary**

---

## Platform Behavior

### Linux (Auto Setup)

* Run `chicha-ip-proxy` **without flags**
* Interactive wizard starts automatically
* Asks for:

  * target IP
  * protocol (TCP / UDP)
  * local ports
* Automatically:

  * creates a **systemd service**
  * enables **autostart on boot**
  * starts the service
  * configures **rotating logs**

➡️ **No manual setup required**

---

### Other Platforms (macOS, BSD, Windows)

* No interactive wizard
* Run **with flags only**
* Services / autostart must be configured **manually**
* Same performance and features (except systemd)

---

## Typical Use Cases

* Mirror websites (80 / 443)
* Forward game or VPN ports (UDP)
* Replace `xinetd` or `iptables` forwarding
* Expose internal services quickly

---

## Example Usage

### TCP forwarding

```bash
chicha-ip-proxy \
  -routes "80:203.0.113.10:80,443:203.0.113.10:443"
```

### UDP forwarding (e.g. WireGuard)

```bash
chicha-ip-proxy \
  -udp-routes "51820:203.0.113.20:51820"
```

### Mixed TCP + UDP

```bash
chicha-ip-proxy \
  -routes "8080:203.0.113.10:80" \
  -udp-routes "5353:203.0.113.20:53"
```

---

## Logs

* Plain text
* Automatically rotated
* Default rotation: **every 24h**
* Easy to grep and monitor

Example:

```
/var/log/chicha-ip-proxy-tcp-443-80.log
```

---

## Download (All Platforms)

👉 **Releases:**
[https://github.com/matveynator/chicha-ip-proxy/releases/](https://github.com/matveynator/chicha-ip-proxy/releases/)


