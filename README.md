[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)


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
  To be used with client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

---

## ✨ Key Features

* **💻 Web Management Panel**: Built-in modern management interface for easy user creation, plugin configuration generation, repository, and note management.
* **🔄 Multi-terminal Note Sync**:
    * Support for automatic **Vault** creation.
    * Support for note management (Create, Read, Update, Delete) with millisecond-level real-time distribution to all online devices.
* **🖼️ Attachment Sync Support**:
    * Perfectly supports synchronization of non-note files such as images.
    * *(Note: Requires server v0.9+ and [Obsidian plugin v1.0+](https://github.com/haierkeys/obsidian-fast-note-sync/releases), does not support Obsidian configuration files)*
* **📝 Note History**:
    * View historical modification versions of each note on the Web page and plugin side.
    * (Requires server v1.2+)
* **⚙️ Config Sync**:
    * Supports synchronization of `.obsidian` configuration files.

## ⏱️ Changelog

- ♨️ [Visit to view the changelog](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/CHANGELOG.en.md)

## 🗺️ Roadmap

We are continuously improving, and here are the future development plans:

- [ ] **Git Version Control Integration**: Provides more secure version rollback for notes.
- [ ] **Sync Algorithm Optimization**: Integrate `google-diff-match-patch` for more efficient incremental synchronization.
- [ ] **Cloud Storage and Backup Strategy**:
    - [ ] Custom backup strategy configuration.
    - [ ] Multi-protocol adaptation: S3 / Minio / Cloudflare R2 / Alibaba Cloud OSS / WebDAV.

> **If you have suggestions for improvement or new ideas, feel free to share them by submitting an issue—we will carefully evaluate and adopt suitable suggestions.**

## 🚀 Quick Deployment

We provide multiple installation methods, recommending the **One-click script** or **Docker**.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**Main actions of the script:**

  * Automatically downloads the Release binary compatible with the current system.
  * Default installation to `/opt/fast-note`, and creates a shortcut command at `/usr/local/bin/fast-note`.
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
    Open `http://{Server IP}:9000` in your browser.
2.  **Initial Setup**:
    Register an account on your first visit. *(If you want to disable registration, set `user.register-is-enable: false` in the configuration file)*
3.  **Configure Client**:
    Log in to the management panel and click **"Copy API Configuration"**.
4.  **Connect Obsidian**:
    Open the Obsidian plugin settings page and paste the configuration information you just copied.

## ⚙️ Configuration Description

The default configuration file is `config.yaml`, and the program will automatically search in the **root directory** or **config/** directory.

View complete configuration example: [config/config.yaml](https://www.google.com/search?q=config/config.yaml)

## 📅 Version History

To view the complete version iteration records, please visit the [Releases page](https://github.com/haierkeys/fast-note-sync-service/releases).

## ☕ Support & Sponsorship

This project is completely open-source and free. If you find it helpful, feel free to **Star** the project or buy the author a cup of coffee. This will motivate me to continue maintaining it. Thank you!

[<img src="https://cdn.ko-fi.com/cdn/kofi3.png?v=3" alt="BuyMeACoffee" width="100">](https://ko-fi.com/haierkeys)

## 🔗 Related Resources

  * [Obsidian Fast Note Sync Plugin (Client Plugin)](https://github.com/haierkeys/obsidian-fast-note-sync)