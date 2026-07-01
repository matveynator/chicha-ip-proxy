# chicha-ip-proxy

TCP / UDP proxy

<img width="45%" alt="chicha-ip-proxy" src="https://github.com/user-attachments/assets/0dc3b863-583f-41ea-b6a4-cfcf773a249c" />

![TCP](https://img.shields.io/badge/TCP-proxy-blue)
![UDP](https://img.shields.io/badge/UDP-proxy-green)
![Autostart](https://img.shields.io/badge/autostart-yes-orange)
![Config](https://img.shields.io/badge/config-not_required-lightgrey)
[![cross-platform-release](https://github.com/matveynator/chicha-ip-proxy/actions/workflows/release.yml/badge.svg)](https://github.com/matveynator/chicha-ip-proxy/actions/workflows/release.yml)

> **RU:** маленький TCP/UDP прокси портов.  
> **EN:** small TCP/UDP port proxy.

---

## Downloads / Скачать

### Linux

| Arch | Download |
|---|---|
| amd64 | [Linux amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64) |
| arm64 | [Linux arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-arm64) |

### macOS

| Arch | Download |
|---|---|
| Intel / amd64 | [macOS Intel](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-amd64) |
| Apple Silicon / arm64 | [macOS Apple Silicon](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-arm64) |

### Windows

| Arch | Download |
|---|---|
| amd64 | [Windows amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-windows-amd64.exe) |
| arm64 | [Windows arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-windows-arm64.exe) |

### FreeBSD

| Arch | Download |
|---|---|
| amd64 | [FreeBSD amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-amd64) |
| arm64 | [FreeBSD arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-arm64) |

### OpenBSD

| Arch | Download |
|---|---|
| amd64 | [OpenBSD amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-amd64) |
| arm64 | [OpenBSD arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-arm64) |

**All releases / Все релизы:**  
[github.com/matveynator/chicha-ip-proxy/releases](https://github.com/matveynator/chicha-ip-proxy/releases/)

---

## Быстрый старт / Quick start

### Linux install

```bash
sudo curl -L https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64 -o /usr/local/bin/chicha-ip-proxy
sudo chmod +x /usr/local/bin/chicha-ip-proxy
sudo chicha-ip-proxy
```

### macOS / FreeBSD / OpenBSD

```bash
chmod +x chicha-ip-proxy-*
sudo ./chicha-ip-proxy-*
```

### Windows

Download `.exe` and run it as Administrator.

---

## Как пользоваться / Usage

Запустите без параметров:

```bash
sudo chicha-ip-proxy
```

Мастер спросит:

```text
target IP
remote port
local port
tcp / udp
allowed client IPs
```

После этого можно сохранить прокси в автозапуск.

Run without arguments:

```bash
sudo chicha-ip-proxy
```

The wizard asks for:

```text
target IP
remote port
local port
tcp / udp
allowed client IPs
```

Then it can save the proxy as an autostart service.

---

## Examples / Примеры

### TCP

```bash
sudo chicha-ip-proxy -local=8080 -remote=203.0.113.10:80 -proto=tcp
```

### UDP DNS

```bash
sudo chicha-ip-proxy -local=54 -remote=8.8.8.8:53 -proto=udp
```

### IPv6 target

```bash
sudo chicha-ip-proxy -local=8443 -remote=[2001:db8::10]:443 -proto=tcp
```

### Allow only one client IP

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
