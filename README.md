# chicha-ip-proxy

TCP / UDP proxy

<img width="45%" alt="chicha-ip-proxy" src="https://github.com/user-attachments/assets/0dc3b863-583f-41ea-b6a4-cfcf773a249c" />

![TCP](https://img.shields.io/badge/TCP-proxy-blue)
![UDP](https://img.shields.io/badge/UDP-proxy-green)
![Autostart](https://img.shields.io/badge/autostart-yes-orange)
![Config](https://img.shields.io/badge/config-not_required-lightgrey)
<br>[![cross-platform-release](https://github.com/matveynator/chicha-ip-proxy/actions/workflows/release.yml/badge.svg)](https://github.com/matveynator/chicha-ip-proxy/actions/workflows/release.yml)

> **RU:** маленький TCP/UDP прокси портов.  
> **EN:** small TCP/UDP port proxy.

---

## Downloads / Скачать

Откройте нужную платформу и скачайте подходящую сборку.  
Open your platform and download the right build.

<p>
  <a href="https://github.com/matveynator/chicha-ip-proxy/releases/latest">
    <img alt="Latest release" src="https://img.shields.io/badge/latest-release-blue">
  </a>
  <a href="https://github.com/matveynator/chicha-ip-proxy/releases/">
    <img alt="All releases" src="https://img.shields.io/badge/all-releases-lightgrey">
  </a>
</p>

---

<details open>
<summary>
  <img src="https://sitebrush.com/04b158d78c93b65c714bb6256da221a4.png" width="22" alt="Linux">
  <b>Linux</b> — recommended / рекомендуется
</summary>

<br>

| Arch | Direct download |
|---|---|
| amd64 | [chicha-ip-proxy-linux-amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64) |
| arm64 | [chicha-ip-proxy-linux-arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-arm64) |

```bash
sudo curl -L -o /usr/local/bin/chicha-ip-proxy \
  https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64

sudo chmod +x /usr/local/bin/chicha-ip-proxy
sudo chicha-ip-proxy
```

</details>

---

<details>
<summary>
  <img src="https://sitebrush.com/fbad588e1b8c94b6b80708bc9917706e.png" width="22" alt="macOS">
  <b>macOS</b>
</summary>

<br>

| Arch | Direct download |
|---|---|
| Intel / amd64 | [chicha-ip-proxy-darwin-amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-amd64) |
| Apple Silicon / arm64 | [chicha-ip-proxy-darwin-arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-arm64) |

```bash
curl -L -o chicha-ip-proxy \
  https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-arm64

chmod +x chicha-ip-proxy
sudo ./chicha-ip-proxy
```

</details>

---

<details>
<summary>
  <img src="https://sitebrush.com/66aab89d1af641ee0ae190f6b3ea4e09.png" width="22" alt="Windows">
  <b>Windows</b>
</summary>

<br>

| Arch | Direct download |
|---|---|
| amd64 | [chicha-ip-proxy-windows-amd64.exe](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-windows-amd64.exe) |
| arm64 | [chicha-ip-proxy-windows-arm64.exe](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-windows-arm64.exe) |

Run PowerShell as Administrator:

```powershell
$Dir = "$env:ProgramFiles\chicha-ip-proxy"
$Exe = "$Dir\chicha-ip-proxy.exe"

New-Item -ItemType Directory -Force -Path $Dir | Out-Null

Invoke-WebRequest `
  -Uri "https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-windows-amd64.exe" `
  -OutFile $Exe

& $Exe
```

</details>

---

<details>
<summary>
  <img src="https://sitebrush.com/c1ce8baa90a2ffd348069e69fa4fda93.png" width="22" alt="FreeBSD">
  <b>FreeBSD</b>
</summary>

<br>

| Arch | Direct download |
|---|---|
| amd64 | [chicha-ip-proxy-freebsd-amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-amd64) |
| arm64 | [chicha-ip-proxy-freebsd-arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-arm64) |

```bash
sudo fetch -o /usr/local/bin/chicha-ip-proxy \
  https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-amd64

sudo chmod +x /usr/local/bin/chicha-ip-proxy
sudo chicha-ip-proxy
```

</details>

---

<details>
<summary>
  <img src="https://sitebrush.com/e3124d65b5feeb6af8ec8f882b167a35.png" width="22" alt="OpenBSD">
  <b>OpenBSD</b>
</summary>

<br>

| Arch | Direct download |
|---|---|
| amd64 | [chicha-ip-proxy-openbsd-amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-amd64) |
| arm64 | [chicha-ip-proxy-openbsd-arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-arm64) |

```bash
sudo ftp -o /usr/local/bin/chicha-ip-proxy \
  https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-amd64

sudo chmod +x /usr/local/bin/chicha-ip-proxy
sudo chicha-ip-proxy
```

</details>

---

**All releases / Все релизы:**  
[github.com/matveynator/chicha-ip-proxy/releases](https://github.com/matveynator/chicha-ip-proxy/releases/)

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
