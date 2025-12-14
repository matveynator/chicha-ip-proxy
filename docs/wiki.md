# Chicha IP Proxy Quick Guide / Краткое руководство

This wiki keeps the setup short and respectful: copy, run, and get a TCP mirror or WireGuard UDP relay running in minutes.

## English
### One-line install from GitHub releases with systemd
Use this when you want automatic download plus a ready-to-run service.
```bash
#!/usr/bin/env bash
set -euo pipefail
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  armv7*|armv6*) ARCH=arm ;;
  *) echo "Unsupported arch: $ARCH" >&2; exit 1 ;;
esac
BIN="chicha-ip-proxy-${OS}-${ARCH}"
BASE="https://github.com/matveynator/chicha-ip-proxy/releases/latest/download"
sudo curl -L "$BASE/$BIN" -o /usr/local/bin/chicha-ip-proxy
sudo chmod +x /usr/local/bin/chicha-ip-proxy
sudo tee /etc/systemd/system/chicha-ip-proxy.service <<'UNIT' >/dev/null
[Unit]
Description=Chicha IP Proxy
After=network.target

[Service]
ExecStart=/usr/local/bin/chicha-ip-proxy \
  -routes "80:203.0.113.10:80,443:203.0.113.10:443" \
  -udp-routes "51820:203.0.113.10:51820" \
  -log /var/log/chicha-ip-proxy.log
Restart=on-failure

[Install]
WantedBy=multi-user.target
UNIT
sudo systemctl daemon-reload
sudo systemctl enable --now chicha-ip-proxy
```
Replace `203.0.113.10` with your origin. Add comma-separated rules to `-routes` (TCP) or `-udp-routes` (UDP) as needed.

### Quick run without systemd
- Mirror a site over TCP: `sudo chicha-ip-proxy -routes "80:198.51.100.5:80,443:198.51.100.5:443" -log /var/log/chicha-ip-proxy.log`
- Forward WireGuard over UDP: `sudo chicha-ip-proxy -udp-routes "51820:198.51.100.20:51820" -log /var/log/chicha-ip-proxy.log`

Logs rotate daily without compression, so you can read them directly.

---

## Русский
### Установка из GitHub релизов и запуск через systemd одной командой
Копируйте и запускайте, чтобы скачать бинарник и сразу поднять сервис.
```bash
#!/usr/bin/env bash
set -euo pipefail
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  armv7*|armv6*) ARCH=arm ;;
  *) echo "Неподдерживаемая архитектура: $ARCH" >&2; exit 1 ;;
esac
BIN="chicha-ip-proxy-${OS}-${ARCH}"
BASE="https://github.com/matveynator/chicha-ip-proxy/releases/latest/download"
sudo curl -L "$BASE/$BIN" -o /usr/local/bin/chicha-ip-proxy
sudo chmod +x /usr/local/bin/chicha-ip-proxy
sudo tee /etc/systemd/system/chicha-ip-proxy.service <<'UNIT' >/dev/null
[Unit]
Description=Chicha IP Proxy
After=network.target

[Service]
ExecStart=/usr/local/bin/chicha-ip-proxy \
  -routes "80:203.0.113.10:80,443:203.0.113.10:443" \
  -udp-routes "51820:203.0.113.10:51820" \
  -log /var/log/chicha-ip-proxy.log
Restart=on-failure

[Install]
WantedBy=multi-user.target
UNIT
sudo systemctl daemon-reload
sudo systemctl enable --now chicha-ip-proxy
```
Замените `203.0.113.10` на свой сервер. Дополнительно добавляйте порты через запятую в `-routes` (TCP) и `-udp-routes` (UDP).

### Быстрый запуск без systemd
- Зеркало сайта по TCP: `sudo chicha-ip-proxy -routes "80:198.51.100.5:80,443:198.51.100.5:443" -log /var/log/chicha-ip-proxy.log`
- WireGuard по UDP: `sudo chicha-ip-proxy -udp-routes "51820:198.51.100.20:51820" -log /var/log/chicha-ip-proxy.log`

Логи ротируются ежедневно без сжатия, остаются в понятном виде для просмотра и поиска.
