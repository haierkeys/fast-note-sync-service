[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>High-performance, low-latency note sync service solution</strong>
  <br>
  <em>Built with Golang + Websocket + Sqlite + React</em>
</p>

<p align="center">
  Must be used with the client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
    <img src="https://image.diybeta.com/blog/fast-note-sync-service-2.png" alt="fast-note-sync-service-preview" width="800" />
</div>

---

## ✨ Key Features

* **💻 Web Admin Panel**: Built-in modern management interface to easily create users, generate plugin configurations, manage vaults, and note content.
* **🔄 Multi-device Note Sync**:
    * Supports automatic **Vault** creation.
    * Supports note management (Add, Delete, Edit, Search), with millisecond-level real-time distribution to all online devices.
* **🖼️ Attachment Sync Support**:
    * Perfect support for syncing images and other non-note files.
    * *(Note: Requires server v0.9+ and [Obsidian plugin v1.0+](https://github.com/haierkeys/obsidian-fast-note-sync/releases), does not support Obsidian setting files)*
* **📝 Note History**:
    * View the historical versions of each note on the Web page or within the plugin.
    * (Requires server v1.2+)
* **⚙️ Configuration Sync**:
    * Supports syncing of `.obsidian` configuration files.

## ☕ Support & Sponsorship

- If you find this plugin useful and want to support its continued development, please consider supporting me through:

  | Ko-fi *Outside China*  |  | WeChat Scan to Donate *China* |
  | --- | ---| --- |
  | [<img src="https://ik.imagekit.io/haierkeys/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="https://ik.imagekit.io/haierkeys/wxds.png" height="150"> |

## ⏱️ Changelog

- ♨️ [View Changelog](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/CHANGELOG.en.md)

## 🗺️ Roadmap

We are continuously improving, and here is our future development plan:

- [ ] **Git Version Control Integration**: Provide safer version backtracking for notes.
- [ ] **Sync Algorithm Optimization**: Integrate `google-diff-match-patch` for more efficient incremental sync.
- [ ] **Cloud Storage & Backup Strategy**:
    - [ ] Custom backup strategy configuration.
    - [ ] Multi-protocol adaptation: S3 / Minio / Cloudflare R2 / Aliyun OSS / WebDAV.

> **If you have suggestions for improvement or new ideas, please share them with us by submitting an issue — we will carefully evaluate and adopt appropriate suggestions.**

## 🚀 Quick Deployment

We provide various installation methods, with **One-click script** or **Docker** being recommended.

### Method 1: One-click script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**Main actions of the script:**

  * Automatically downloads the Release binary suited for the current system.
  * Default installation to `/opt/fast-note`, creates a shortcut at `/usr/local/bin/fast-note`.
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

Download the latest version for your system from [Releases](https://github.com/haierkeys/fast-note-sync-service/releases), unzip it, and run:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 Usage Guide

1.  **Access Admin Panel**:
    Open `http://{Server_IP}:9000` in your browser.
2.  **Initial Setup**:
    Register an account on first visit. *(To disable registration, set `user.register-is-enable: false` in the config file)*
3.  **Configure Client**:
    Log in to the admin panel and click **"Copy API Config"**.
4.  **Connect Obsidian**:
    Open the Obsidian plugin settings page and paste the configuration information you just copied.

## ⚙️ Configuration

The default configuration file is `config.yaml`, and the program will automatically search for it in the **root directory** or the **config/** directory.

View full configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)


## 🔗 Related Resources

  * [Obsidian Fast Note Sync Plugin (Client Plugin)](https://github.com/haierkeys/obsidian-fast-note-sync)