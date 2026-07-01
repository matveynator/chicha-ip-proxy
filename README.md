<div align="center">

# chicha-ip-proxy
 BLAZING FAST and SECURE TCP/UDP port proxy written in GO.


<img width="100%"  alt="chicha-ip-proxy" src="https://github.com/user-attachments/assets/b23ffd1c-aafc-4534-a8a5-1492ce371def" />
<br>

![TCP](https://img.shields.io/badge/TCP-proxy-green)
![UDP](https://img.shields.io/badge/UDP-proxy-green)
![Autostart](https://img.shields.io/badge/autostart-yes-green)
![Config](https://img.shields.io/badge/config-not_required-green) <br>

[![Go Reference](https://pkg.go.dev/badge/github.com/matveynator/chicha-ip-proxy.svg)](https://pkg.go.dev/github.com/matveynator/chicha-ip-proxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/matveynator/chicha-ip-proxy)](https://goreportcard.com/report/github.com/matveynator/chicha-ip-proxy)
[![Coverage](https://codecov.io/gh/matveynator/chicha-ip-proxy/branch/master/graph/badge.svg)](https://codecov.io/gh/matveynator/chicha-ip-proxy)

[![cross-platform-release](https://github.com/matveynator/chicha-ip-proxy/actions/workflows/release.yml/badge.svg)](https://github.com/matveynator/chicha-ip-proxy/actions/workflows/release.yml)


</div>

---

## Downloads / Скачать


<details>
<summary>
  <img width="42" alt="linux" src="https://github.com/user-attachments/assets/bf3141b6-4c93-4fd6-b2d1-421b79876dcb" />
  <b><big>Linux</big></b>
  <sub>amd64 / arm64</sub>
</summary>

<br>

**amd64** · [download](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64)

```bash
sudo curl -L -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy
```

**arm64** · [download](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-arm64)

```bash
sudo curl -L -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-arm64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy
```

</details>

---

<details>
<summary>
  <img width="36" alt="mac" src="https://github.com/user-attachments/assets/946102b8-f043-494d-809a-a589e536ee9a" />
  <b><big>macOS</big></b>
  <sub>Intel / Apple Silicon</sub>
</summary>

<br>

**Intel / amd64** · [download](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-amd64)

```bash
sudo mkdir -p /usr/local/bin; sudo curl -L -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-amd64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy
```

**Apple Silicon / arm64** · [download](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-arm64)

```bash
sudo mkdir -p /usr/local/bin; sudo curl -L -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-arm64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy
```

</details>

---

<details>
<summary>
  <img width="42" alt="windows" src="https://github.com/user-attachments/assets/f6044001-95b0-4500-a4f6-1c3b08eb65fb" />
  <b><big>Windows</big></b>
  <sub>amd64 / arm64</sub>
</summary>

<br>

<sub>Run PowerShell as Administrator / Запустите PowerShell от администратора.</sub>

<br><br>

**amd64** · [download](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-windows-amd64.exe)

```powershell
$p="$env:ProgramFiles\chicha-ip-proxy\chicha-ip-proxy.exe"; New-Item -ItemType Directory -Force -Path (Split-Path $p) | Out-Null; Invoke-WebRequest -Uri "https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-windows-amd64.exe" -OutFile $p; & $p
```

**arm64** · [download](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-windows-arm64.exe)

```powershell
$p="$env:ProgramFiles\chicha-ip-proxy\chicha-ip-proxy.exe"; New-Item -ItemType Directory -Force -Path (Split-Path $p) | Out-Null; Invoke-WebRequest -Uri "https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-windows-arm64.exe" -OutFile $p; & $p
```

</details>

---

<details>
<summary>
  <img width="42" alt="freebsd" src="https://github.com/user-attachments/assets/d35baaac-d296-41b1-a281-55dc761328e9" />
  <b><big>FreeBSD</big></b>
  <sub>amd64 / arm64</sub>
</summary>

<br>

**amd64** · [download](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-amd64)

```bash
sudo fetch -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-amd64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy
```

**arm64** · [download](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-arm64)

```bash
sudo fetch -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-arm64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy
```

</details>

---

<details>
<summary>
  <img width="42" alt="openbsd" src="https://github.com/user-attachments/assets/11633d7e-5744-46da-ad2f-6e49c69e51de" />
  <b><big>OpenBSD</big></b>
  <sub>amd64 / arm64</sub>
</summary>

<br>

**amd64** · [download](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-amd64)

```bash
sudo ftp -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-amd64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy
```

**arm64** · [download](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-arm64)

```bash
sudo ftp -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-arm64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy
```

</details>

---

**All releases / Все релизы:**
[github.com/matveynator/chicha-ip-proxy/releases](https://github.com/matveynator/chicha-ip-proxy/releases/)

---

## Usage / Как пользоваться

Run without arguments / Запустите без параметров:

```bash
sudo chicha-ip-proxy
```

The wizard asks / Мастер спросит:

```text
target IP
remote port
local port
tcp / udp
allowed client IPs
```

Then it can save the proxy as an autostart service.
После этого можно сохранить прокси в автозапуск.

---

## Examples / Примеры

### TCP

```bash
sudo chicha-ip-proxy -local=8080 -remote=203.0.113.10:80 -proto=tcp
```

### UDP DNS

```bash
sudo chicha-ip-proxy -local=53 -remote=8.8.8.8:53 -proto=udp
```

### IPv6 target

```bash
sudo chicha-ip-proxy -local=8443 -remote=[2001:db8::10]:443 -proto=tcp
```

### Allow only one client IP / Разрешить только один IP

```bash
sudo chicha-ip-proxy -local=8080 -remote=203.0.113.10:80 -allow=198.51.100.7
```

---

## Flags / Флаги

```text
-local   local port / локальный порт
-remote  target IP[:PORT] or [IPv6]:PORT / куда пересылать
-proto   tcp or udp
-allow   allowed IP/CIDR
```

If `-allow` is not set, all clients are allowed.
Если `-allow` не указан, разрешены все клиенты.


# Common TCP/UDP Proxy Problems Solved by chicha-ip-proxy

These are the most common questions people ask on forums like Stack Overflow, Server Fault, Reddit, Habr, and Linux.org.ru.

* I just need a simple TCP proxy.
* I just need a simple UDP proxy.
* I need to proxy both TCP and UDP.
* Forward a TCP/UDP port to another server.
* Expose a service through a VPS with a public IP.
* Bypass NAT using a VPS.
* Proxy game servers (Minecraft, CS, Rust, etc.).
* Proxy VPN traffic (OpenVPN, WireGuard, etc.).
* Proxy SSH, RDP, VNC, FTP, SMTP, IMAP, and other non-HTTP protocols.
* Move a service to another server without changing the client configuration.
* Replace nginx stream for simple port forwarding.
* Avoid HAProxy for basic TCP/UDP proxying.
* Replace socat, xinetd, or rinetd with a single tool.
* Avoid writing iptables/nftables rules.
* Start a proxy with a single command.
* Run automatically after reboot.
* Restrict access by IP address.
* Manage dozens or hundreds of forwarded ports.
* Use a solution without configuration files.
* Use the same tool on Linux, Windows, macOS, FreeBSD, and OpenBSD.
* Replace several networking utilities with one lightweight application.
* Temporarily redirect traffic during migrations.
* Hide the real backend behind a relay server.
* Proxy raw TCP/UDP traffic instead of HTTP.

In short, most forum questions come down to one simple request:

“I just need to proxy a TCP or UDP port, but everyone recommends nginx, HAProxy, socat, xinetd, rinetd, iptables, WireGuard, or some other complex setup.”

chicha-ip-proxy is designed specifically for this use case: a lightweight, cross-platform TCP/UDP port proxy with no configuration files, simple setup, automatic startup, and optional IP-based access control.
