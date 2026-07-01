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
  <img width="50" alt="linux" src="https://github.com/user-attachments/assets/bf3141b6-4c93-4fd6-b2d1-421b79876dcb" />
  <b>Linux</b> — recommended / рекомендуется
</summary>

<br>

| Arch | Direct download |
|---|---|
| amd64 | [chicha-ip-proxy-linux-amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64) |
| arm64 | [chicha-ip-proxy-linux-arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-arm64) |

```bash
sudo curl -L -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy;
```

</details>

---

<details>
<summary>
  <img width="40" alt="mac" src="https://github.com/user-attachments/assets/946102b8-f043-494d-809a-a589e536ee9a" />
  <b>macOS</b>
</summary>

<br>

| Arch | Direct download |
|---|---|
| Intel / amd64 | [chicha-ip-proxy-darwin-amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-amd64) |
| Apple Silicon / arm64 | [chicha-ip-proxy-darwin-arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-arm64) |

```bash
sudo curl -L -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-darwin-arm64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy;
```

</details>

---

<details>
<summary>
  <img width="50" alt="windows" src="https://github.com/user-attachments/assets/f6044001-95b0-4500-a4f6-1c3b08eb65fb" />
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
  <img width="50" alt="freebsd" src="https://github.com/user-attachments/assets/d35baaac-d296-41b1-a281-55dc761328e9" />
  <b>FreeBSD</b>
</summary>

<br>

| Arch | Direct download |
|---|---|
| amd64 | [chicha-ip-proxy-freebsd-amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-amd64) |
| arm64 | [chicha-ip-proxy-freebsd-arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-arm64) |

```bash
sudo fetch -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-freebsd-amd64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy;
```

</details>

---

<details>
<summary>
  <img width="50" alt="openbsd" src="https://github.com/user-attachments/assets/11633d7e-5744-46da-ad2f-6e49c69e51de" />
  <b>OpenBSD</b>
</summary>

<br>

| Arch | Direct download |
|---|---|
| amd64 | [chicha-ip-proxy-openbsd-amd64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-amd64) |
| arm64 | [chicha-ip-proxy-openbsd-arm64](https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-arm64) |

```bash
sudo ftp -o /usr/local/bin/chicha-ip-proxy https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-openbsd-amd64; sudo chmod +x /usr/local/bin/chicha-ip-proxy; sudo chicha-ip-proxy;
```

</details>

---

**All releases / Все релизы:**  
[github.com/matveynator/chicha-ip-proxy/releases](https://github.com/matveynator/chicha-ip-proxy/releases/)

---

## Как пользоваться / Usage

Запустите без параметров / Run without arguments:

```bash
sudo chicha-ip-proxy
```

Мастер спросит / The wizard asks for:

```text
target IP
remote port
local port
tcp / udp
allowed client IPs
```

После этого можно сохранить прокси в автозапуск / Then it can save the proxy as an autostart service.


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

### Allow only from one client IP / Ограничить по 1 айпи:

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
