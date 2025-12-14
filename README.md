<img src="https://github.com/matveynator/chicha-ip-proxy/blob/master/chicha-ip-proxy.png?raw=true" width="100%" align="right"></img>


# Chicha IP Proxy is the fastest and simplest solution for port forwarding.

Chicha IP Proxy is a lightweight **Layer 2 (L2) proxy** written in **Go**, designed for efficient port forwarding. It’s faster and simpler than traditional tools like `xinetd`, making it ideal for handling high traffic with minimal setup.

---

### **Why Choose Chicha IP Proxy?**
- **Blazing Fast**: Written in Go, it fully utilizes all CPU cores for optimal performance.
- **Simple to Use**: One command to forward traffic—no complex configurations.
- **Log Rotation**: Rotates logs daily while keeping them uncompressed for easy inspection.
- **Cross-Platform**: Compatible with all major operating systems and architectures.
- **Efficient**: Low resource usage even under heavy load.

---

### **Quick Setup / Краткий запуск**

#### **English — minimal auto-setup**
The built-in wizard does the heavy lifting (systemd unit, enable, start, log tail). You only download the binary and run it. Replace `203.0.113.44` with your origin.
```bash
curl -L https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64 -o chicha-ip-proxy
sudo install -m 0755 chicha-ip-proxy /usr/local/bin/chicha-ip-proxy
sudo /usr/local/bin/chicha-ip-proxy
```
Short transcript so you know what to expect:
```
root@mirror:~# /usr/local/bin/chicha-ip-proxy
No routes provided via flags. Starting interactive configuration...
Enter target IP address to proxy to: 203.0.113.44
Enter protocols (comma separated, supported: tcp, udp): tcp
Enter local TCP ports (comma separated): 80, 443
Planned log file: /var/log/chicha-ip-proxy-tcp-443-80.log
Planned systemd service name: chicha-ip-proxy-tcp-443-80.service
Would you like to create a systemd service 'chicha-ip-proxy-tcp-443-80.service'? (y/N): y
Enable the service so it starts on boot? (y/N): y
Start the service now? (y/N): y
Follow the log file now? (y/N): y
========== CHICHA IP PROXY ==========
TCP Routes:
  LocalPort=443 -> RemoteIP=203.0.113.44 RemotePort=443
  LocalPort=80 -> RemoteIP=203.0.113.44 RemotePort=80
Log file: /var/log/chicha-ip-proxy-tcp-443-80.log
Log rotation frequency: 24h0m0s
======================================
```

#### **English — quick flags instead of wizard**
- Mirror a site over TCP:
  ```bash
  sudo chicha-ip-proxy -routes "80:198.51.100.5:80,443:198.51.100.5:443" -log /var/log/chicha-ip-proxy.log
  ```
- Forward WireGuard over UDP:
  ```bash
  sudo chicha-ip-proxy -udp-routes "51820:198.51.100.20:51820" -log /var/log/chicha-ip-proxy.log
  ```

Daily log rotation keeps files uncompressed for fast inspection.

---

#### **Русский — минимальная автонастройка**
Мастер сам создаёт unit, включает, запускает и показывает лог. Вам нужно только скачать бинарник и запустить его. Замените `203.0.113.44` на ваш источник.
```bash
curl -L https://github.com/matveynator/chicha-ip-proxy/releases/latest/download/chicha-ip-proxy-linux-amd64 -o chicha-ip-proxy
sudo install -m 0755 chicha-ip-proxy /usr/local/bin/chicha-ip-proxy
sudo /usr/local/bin/chicha-ip-proxy
```
Короткий диалог мастера:
```
root@mirror:~# /usr/local/bin/chicha-ip-proxy
No routes provided via flags. Starting interactive configuration...
Enter target IP address to proxy to: 203.0.113.44
Enter protocols (comma separated, supported: tcp, udp): tcp
Enter local TCP ports (comma separated): 80, 443
Planned log file: /var/log/chicha-ip-proxy-tcp-443-80.log
Planned systemd service name: chicha-ip-proxy-tcp-443-80.service
Would you like to create a systemd service 'chicha-ip-proxy-tcp-443-80.service'? (y/N): y
Enable the service so it starts on boot? (y/N): y
Start the service now? (y/N): y
Follow the log file now? (y/N): y
========== CHICHA IP PROXY ==========
TCP Routes:
  LocalPort=443 -> RemoteIP=203.0.113.44 RemotePort=443
  LocalPort=80 -> RemoteIP=203.0.113.44 RemotePort=80
Log file: /var/log/chicha-ip-proxy-tcp-443-80.log
Log rotation frequency: 24h0m0s
======================================
```

#### **Русский — запуск с флагами вместо мастера**
- Зеркало сайта по TCP:
  ```bash
  sudo chicha-ip-proxy -routes "80:198.51.100.5:80,443:198.51.100.5:443" -log /var/log/chicha-ip-proxy.log
  ```
- Проброс WireGuard по UDP:
  ```bash
  sudo chicha-ip-proxy -udp-routes "51820:198.51.100.20:51820" -log /var/log/chicha-ip-proxy.log
  ```

Логи ротируются ежедневно без сжатия и остаются удобными для просмотра и поиска.

---

### **Command reference / Справка по командам**

```
chicha-ip-proxy-linux --help
Usage of chicha-ip-proxy-linux:
  -log string
    Path to the log file (default "chicha-ip-proxy.log")
  -rotation duration
    Log rotation frequency (e.g. 24h, 1h, etc.) (default 24h0m0s)
  -routes string
    Comma-separated list of TCP routes in the format LOCALPORT:REMOTEIP:REMOTEPORT
  -udp-routes string
    Comma-separated list of UDP routes in the format LOCALPORT:REMOTEIP:REMOTEPORT
  -version
    Print the version of the proxy and exit
```

```
chicha-ip-proxy-linux --help
Подсказка по флагам:
  -log string
    Путь к файлу логов (по умолчанию "chicha-ip-proxy.log")
  -rotation duration
    Как часто ротировать логи (например 24h, 1h и т.д.) (по умолчанию 24h0m0s)
  -routes string
    TCP-маршруты через запятую в виде LOCALPORT:REMOTEIP:REMOTEPORT
  -udp-routes string
    UDP-маршруты через запятую в виде LOCALPORT:REMOTEIP:REMOTEPORT
  -version
    Показать версию и выйти
```

### **Why Chicha IP Proxy?**
- **Go-Powered Performance**: Written in Go, ensuring speed and reliability.
- **Multi-Port Support**: Easily forward traffic for one or multiple ports.
- **No Complexity**: Simple commands, no bloated configs.
- **Ready for Production**: Log rotation with readable archives and systemd integration make it production-ready.

---
### **Benchmarks**

#### **chicha-ip-proxy:**

```
siege http://localhost:8081 -t 15s -c 100
** SIEGE 4.0.4
** Preparing 100 concurrent users for battle.
The server is now under siege...
Lifting the server siege...
Transactions:		       15728 hits
Availability:		      100.00 %
Elapsed time:		       14.65 secs
Data transferred:	      124.07 MB
Response time:		        0.03 secs
Transaction rate:	     1073.58 trans/sec
Throughput:		        8.47 MB/sec
Concurrency:		       33.37
Successful transactions:       11824
Failed transactions:	           0
Longest transaction:	        0.25
Shortest transaction:	        0.00
```

#### **xinetd:**
```
siege http://localhost:8082 -t 15s -c 100
** SIEGE 4.0.4
** Preparing 100 concurrent users for battle.
The server is now under siege...
Lifting the server siege...
Transactions:		       14863 hits
Availability:		      100.00 %
Elapsed time:		       14.57 secs
Data transferred:	      117.20 MB
Response time:		        0.04 secs
Transaction rate:	     1020.11 trans/sec
Throughput:		        8.04 MB/sec
Concurrency:		       36.99
Successful transactions:       11178
Failed transactions:	           0
Longest transaction:	        0.55
Shortest transaction:	        0.00
 
```

#### **direct requests:**
```
siege http://files.zabiyaka.net:80 -t 15s -c 100
** SIEGE 4.0.4
** Preparing 100 concurrent users for battle.
The server is now under siege...
Lifting the server siege...
Transactions:		       14778 hits
Availability:		      100.00 %
Elapsed time:		       14.14 secs
Data transferred:	      116.52 MB
Response time:		        0.03 secs
Transaction rate:	     1045.12 trans/sec
Throughput:		        8.24 MB/sec
Concurrency:		       35.07
Successful transactions:       11112
Failed transactions:	           0
Longest transaction:	        0.19
Shortest transaction:	        0.00
```

---


Chicha IP Proxy is the **fastest and simplest solution** for port forwarding. Whether forwarding one port or dozens, it's the ideal tool for sysadmins looking for performance and ease of use!
