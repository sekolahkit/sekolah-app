# Deployment

## Overview

Aplikasi bisa di-deploy dengan beberapa cara:
1. Binary langsung di server
2. Docker
3. Reverse proxy dengan Nginx/Caddy

---

## Binary Langsung

### 1. Download Binary

```bash
wget https://github.com/Sekolahkit/sekolah-app/releases/latest/download/sekolah-app-linux-amd64
chmod +x sekolah-app-linux-amd64
```

### 2. Buat Service (systemd)

Buat file `/etc/systemd/system/sekolah-app.service`:

```ini
[Unit]
Description=SekolahApp
After=network.target

[Service]
Type=simple
User=sekolah
Group=sekolah
WorkingDirectory=/opt/sekolah-app
ExecStart=/opt/sekolah-app/sekolah-app
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 3. Enable & Start

```bash
sudo systemctl daemon-reload
sudo systemctl enable sekolah-app
sudo systemctl start sekolah-app
sudo systemctl status sekolah-app
```

### 4. Cek Log

```bash
sudo journalctl -u sekolah-app -f
```

---

## Docker

### docker-compose.yml

```yaml
version: '3.8'

services:
  sekolah-app:
    image: sekolahkit/sekolah-app:latest
    container_name: sekolah-app
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./uploads:/app/uploads
      - ./backups:/app/backups
      - ./config.yaml:/app/config.yaml
      - ./.env:/app/.env
    environment:
      - TZ=Asia/Jakarta
    restart: unless-stopped
```

### Jalankan

```bash
docker-compose up -d
docker-compose logs -f
```

### Update

```bash
docker-compose pull
docker-compose up -d
```

---

## Reverse Proxy

### Nginx

Install Nginx:

```bash
sudo apt update
sudo apt install nginx
```

Buat config `/etc/nginx/sites-available/sekolah-app`:

```nginx
server {
    listen 80;
    server_name sekolah.example.com;

    # Redirect HTTP ke HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name sekolah.example.com;

    # SSL
    ssl_certificate /etc/letsencrypt/live/sekolah.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sekolah.example.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Upload limit
    client_max_body_size 10M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Static files (jika tidak embed)
    location /uploads/ {
        alias /opt/sekolah-app/uploads/;
        expires 30d;
    }
}
```

Enable site:

```bash
sudo ln -s /etc/nginx/sites-available/sekolah-app /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### SSL dengan Let's Encrypt

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d sekolah.example.com
```

### Caddy (Alternatif)

Caddyfile:

```
sekolah.example.com {
    reverse_proxy localhost:8080
    encode gzip
    header {
        X-Frame-Options "SAMEORIGIN"
        X-Content-Type-Options "nosniff"
    }
    client_max_body_size 10M
}
```

---

## Server Requirements

### Minimum
- CPU: 1 core
- RAM: 512 MB
- Storage: 10 GB
- OS: Ubuntu 20.04+ / Debian 10+

### Recommended
- CPU: 2 cores
- RAM: 1 GB
- Storage: 50 GB (tergantung jumlah siswa dan upload)
- OS: Ubuntu 22.04 LTS

### Contoh Provider
- DigitalOcean: $6/bulan (1GB RAM)
- Vultr: $6/bulan (1GB RAM)
- Linode: $5/bulan (1GB RAM)
- VPS Lokal: Tergantung provider

---

## Backup Strategy

### Auto Backup

Aplikasi sudah built-in auto backup. Konfigurasi di `config.yaml`:

```yaml
backup:
  enabled: true
  schedule: "0 2 * * *"    # Jam 2 pagi
  retention: 7             # Simpan 7 hari
  path: "./backups"
```

### Manual Backup

```bash
# Backup database
cp /opt/sekolah-app/data/sekolah.db /backup/sekolah-$(date +%Y%m%d).db

# Backup uploads
tar -czf /backup/uploads-$(date +%Y%m%d).tar.gz /opt/sekolah-app/uploads/
```

### Restore

```bash
# Restore database
cp /backup/sekolah-20240115.db /opt/sekolah-app/data/sekolah.db

# Restore uploads
tar -xzf /backup/uploads-20240115.tar.gz -C /
```

---

## Monitoring

### Cek Status

```bash
# Systemd
sudo systemctl status sekolah-app

# Docker
docker-compose ps
docker-compose logs --tail=100
```

### Cek Resource

```bash
# CPU & Memory
htop

# Disk
df -h

# Database size
ls -lh /opt/sekolah-app/data/sekolah.db
```

### Log

```bash
# Systemd
sudo journalctl -u sekolah-app -f

# Docker
docker-compose logs -f

# File log
tail -f /opt/sekolah-app/logs/app.log
```

---

## Troubleshooting

### Port Conflict

```bash
# Cek port 8080
sudo lsof -i :8080

# Ganti port di config.yaml
app:
  port: 9090
```

### Database Locked

```bash
# Cek proses yang akses database
fuser /opt/sekolah-app/data/sekolah.db

# Restart aplikasi
sudo systemctl restart sekolah-app
```

### WhatsApp Disconnect

```bash
# Cek log WhatsApp
grep -i whatsapp /opt/sekolah-app/logs/app.log

# Restart aplikasi untuk reconnect
sudo systemctl restart sekolah-app
```

### Disk Penuh

```bash
# Cek ukuran folder
du -sh /opt/sekolah-app/*

# Hapus backup lama
find /opt/sekolah-app/backups -mtime +7 -delete

# Hapus log lama
find /opt/sekolah-app/logs -mtime +30 -delete
```
