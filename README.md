[中文文档](/docs/README.zh-CN.md) / [English Document](/README.md)

<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>High-performance, low-latency note synchronization service solution</strong>
  <br>
  <em>Built with Golang + Websocket + Sqlite + React</em>
</p>

<p align="center">
  Requires client-side plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

---

## ✨ Core Features

* **💻 Web Management Panel**: Built-in modern management interface to easily create users, generate plugin configurations, manage repositories, and note content.
* **🔄 Multi-terminal Note Sync**:
    * Supports automatic **Vault** creation.
    * Supports note management (Add, Delete, Modify, Search), with millisecond-level real-time distribution of changes to all online devices.
* **🖼️ Attachment Sync Support**:
    * Perfect support for syncing non-note files such as images.
    * *(Note: Requires server v0.9+ and [Obsidian plugin v1.0+](https://github.com/haierkeys/obsidian-fast-note-sync/releases); Obsidian setting files are not supported)*.
* **⚙️ Configuration Sync**:
    * Supports synchronization of `.obsidian` configuration files.

## 🗺️ Roadmap

We are continuously improving. Here is the future development plan:

- [ ] **Git Version Control Integration**: Provides safer version tracking for notes.
- [ ] **Sync Algorithm Optimization**: Integrate `google-diff-match-patch` for more efficient incremental synchronization.
- [ ] **Cloud Storage and Backup Strategy**:
    - [ ] Custom backup strategy configuration.
    - [ ] Multi-protocol adaptation: S3 / Minio / Cloudflare R2 / Aliyun OSS / WebDAV.

> **If you have improvement suggestions or new ideas, feel free to share them by submitting an issue—we will carefully evaluate and adopt appropriate suggestions.**

## 🚀 Rapid Deployment

We provide multiple installation methods; using the **one-click script** or **Docker** is recommended.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**Main actions of the script:**

  * Automatically downloads the Release binary suited for the current system.
  * Default installation to `/opt/fast-note`, with a shortcut command created at `/usr/local/bin/fast-note`.
  * Configures and starts the Systemd service (`fast-note.service`) for auto-start on boot.
  * **Management commands**: `fast-note [install|uninstall|start|stop|status|update|menu]`

-----

### Method 2: Docker Deployment

#### Docker Run

```bash
# 1. Pull the image
docker pull haierkeys/fast-note-sync-service:latest

# 2. Start the container
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
      - ./storage:/fast-note-sync/storage  # Data storage
      - ./config:/fast-note-sync/config    # Configuration files
```

Start the service:

```bash
docker compose up -d
```

-----

### Method 3: Manual Binary Installation

Download the latest version for your system from [Releases](https://github.com/haierkeys/fast-note-sync-service/releases), extract it, and run:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 Usage Guide

1.  **Access Management Panel**:
    Open `http://{Server_IP}:9000` in your browser.
2.  **Initial Setup**:
    Registration is required for the first visit. *(To disable registration, set `user.register-is-enable: false` in the configuration file)*.
3.  **Configure Client**:
    Log in to the management panel and click **"Copy API Config"**.
4.  **Connect Obsidian**:
    Open the Obsidian plugin settings page and paste the copied configuration information.

## ⚙️ Configuration Description

The default configuration file is `config.yaml`. The program automatically looks for it in the **root directory** or the **config/** directory.

View full configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 📅 Change Log

To view the full version iteration record, please visit the [Releases page](https://github.com/haierkeys/fast-note-sync-service/releases).

## ☕ Sponsorship and Support

This project is completely open source and free. If you find it helpful, feel free to **Star** the project or buy the author a coffee. This will motivate me to continue maintaining it. Thank you!

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 🔗 Related Resources

  * [Obsidian Fast Note Sync Plugin (Client Plugin)](https://github.com/haierkeys/obsidian-fast-note-sync)