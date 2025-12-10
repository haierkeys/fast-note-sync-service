[中文文档](readme-zh.md) / [English Document](README.md)

<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>


<p align="center">
  <strong>High-Performance, Low-Latency Note Synchronization Service Solution</strong>
  <br>
  <em>Built with Golang + WebSocket + SQLite + React</em>
</p>


<p align="center">
Requires the client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

-----

## ✨ Core Features

  * **💻 Web Admin Panel**: Built-in modern management interface to easily create users, generate plugin configurations, and manage vaults and note content.
  * **🔄 Multi-Device Real-Time Sync**:
      * Supports automatic **Vault** creation.
      * Supports note management (Create, Read, Update, Delete), with changes distributed to all online devices in milliseconds.
  * **🖼️ Attachment Sync Support**:
      * Perfectly supports syncing non-note files such as images.
      * *(Note: Requires Server v0.9+ and Obsidian Plugin v1.0+. Does not support Obsidian settings files)*

## 🗺️ Roadmap

We are continuously improving. Here is the future development plan:

  - [ ] **Git Version Control Integration**: Safer version rollback for notes.
  - [ ] **Sync Algorithm Optimization**: Integrate `google-diff-match-patch` for more efficient incremental syncing.
  - [ ] **Cloud Storage & Backup Strategy**:
      - [ ] Custom backup policy configuration.
      - [ ] Multi-protocol adaptation: S3 / Minio / Cloudflare R2 / Aliyun OSS / WebDAV.

## 🚀 Quick Deployment

We provide multiple installation methods. **One-click Script** or **Docker** is recommended.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/quest_install.sh)
```

**Script Main Actions:**

  * Automatically downloads the Release binary adapted to the current system.
  * Installs to `/opt/fast-note` by default and creates a shortcut command at `/usr/local/bin/fast-note`.
  * Configures and starts the Systemd service (`fast-note.service`) for auto-start on boot.
  * **Management Command**: `fast-note [install|uninstall|start|stop|status|update|menu]`

-----

### Method 2: Docker Deployment

#### Docker Run

```bash
# 1. Pull image
docker pull haierkeys/fast-note-sync-service:latest

# 2. Start container
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 -p 9001:9001 \
    -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
    -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
    haierkeys/fast-note-sync-service:latest
```

#### Docker Compose

Create a `docker-compose.yaml` file:

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # API Port
      - "9001:9001"  # WebSocket Port
    volumes:
      - ./storage:/fast-note-sync/storage    # Data Storage
      - ./config:/fast-note-sync/config      # Configuration File
```

Start the service:

```bash
docker compose up -d
```

-----

### Method 3: Manual Binary Installation

Download the latest version for your system from [Releases](https://github.com/haierkeys/fast-note-sync-service/releases), unzip, and run:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 User Guide

1.  **Access Admin Panel**:
    Open `http://{Server-IP}:9000` in your browser.
2.  **Initial Setup**:
    Registration is required for the first visit. *(To disable registration, set `user.register-is-enable: false` in the config file)*.
3.  **Configure Client**:
    Log in to the admin panel and click **"Copy API Config"**.
4.  **Connect Obsidian**:
    Open the Obsidian plugin settings page and paste the configuration information you just copied.

## ⚙️ Configuration

The default configuration file is `config.yaml`. The program will automatically look for it in the **root directory** or the **config/** directory.

View full configuration example: [config/config.yaml](https://www.google.com/search?q=https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 📅 Changelog

For full version history, please visit the [Releases Page](https://github.com/haierkeys/fast-note-sync-service/releases).

## ☕ Sponsor & Support

This project is completely open-source and free. If you find it helpful, please **Star** this project or buy the author a coffee. This will motivate me to continue maintenance. Thank you\!

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 🔗 Related Resources

  * [Obsidian Fast Note Sync Plugin (Client Plugin)](https://github.com/haierkeys/obsidian-fast-note-sync)