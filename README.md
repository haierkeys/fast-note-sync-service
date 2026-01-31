[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

If you have any questions, please create an [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new), or join our Telegram group for help: [https://t.me/obsidian_users](https://t.me/obsidian_users)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>High-performance, low-latency note synchronization, online management, and remote REST API service platform.</strong>
  <br>
  <em>Built with Golang + Websocket + Sqlite + React</em>
</p>

<p align="center">
  Data provision requires use with a client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
  <div align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/vault.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/vault.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/attach.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/attach.png" alt="fast-note-sync-service-preview" width="400" /></a>
    </div>
  <div align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/note.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/note.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/images/setting.png"><img src="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/refs/heads/master/docs/images/setting.png" alt="fast-note-sync-service-preview" width="400" /></a>
  </div>
</div>

---

## ✨ Core Features

* **🚀 REST API Support**:
    * Provides standard REST API interfaces, supporting programmatic operations (e.g., automation scripts, AI assistant integration) for creating, reading, updating, and deleting Obsidian notes.
    * For details, please refer to the [REST API Documentation](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/REST_API.md).
* **💻 Web Management Panel**:
  * Built-in modern management interface for easily creating users, generating plugin configurations, and managing vaults and note content.
* **🔄 Multi-device Note Sync**:
    * Supports automatic creation of **Vaults**.
    * Supports note management (Add, Delete, Edit, Query), with millisecond-level real-time distribution to all online devices.
* **🖼️ Attachment Sync Support**:
    * Perfect support for syncing non-note files such as images.
    * Supports chunked upload and download for large attachments, with configurable chunk sizes to improve synchronization efficiency.
* **⚙️ Configuration Sync**:
    * Supports synchronization of `.obsidian` configuration files.
    * Supports synchronization of `PDF` progress status.
* **📝 Note History**:
    * View the historical modification versions of each note via the Web page or the plugin client.
    * (Requires server v1.2+)
* **🗑️ Recycle Bin**:
    * Automatically moves deleted notes to the recycle bin.
    * Supports restoring notes from the recycle bin. (Attachment restoration will be added in future updates)

* **🚫 Offline Sync Strategy**:
    * Supports automatic merging of offline note edits. (Requires plugin settings)
    * Handles offline deletions, with automatic synchronization or deletion upon reconnection. (Requires plugin settings)

## ☕ Sponsorship and Support

- If you find this plugin useful and want to support its continued development, please consider the following options:

  | Ko-fi *Outside China*                                                                                                |    | WeChat Pay *China Only*                                            |
  |----------------------------------------------------------------------------------------------------------------------|----|--------------------------------------------------------------------|
  | [<img src="https://ik.imagekit.io/haierkeys/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="https://ik.imagekit.io/haierkeys/wxds.png" height="150"> |

  - Supported List:
    - https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md

## ⏱️ Changelog

- ♨️ [View Changelog](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/CHANGELOG.en.md)

## 🗺️ Roadmap

We are continuously improving; here are our future development plans:

- [ ] **Sharing Feature**: Support for sharing notes.
- [ ] **MCP Support**: Add support for AI MCP-related features.
- [ ] **Directory Sync**: Support for adding, deleting, modifying, and querying directories.
- [ ] **Git Version Control Integration**: Provide a more secure version history for notes.
- [ ] **Cloud Storage and Backup Strategy**:
    - [ ] Configuration for custom backup strategies.
    - [ ] Multi-protocol adaptation: S3 / Minio / Cloudflare R2 / Alibaba Cloud OSS / WebDAV.

> **If you have suggestions for improvement or new ideas, feel free to share them with us by submitting an issue—we will carefully evaluate and adopt suitable suggestions.**

## 🚀 Quick Deployment

We offer multiple installation methods; **One-click script** or **Docker** is recommended.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**Main Script Actions:**

  * Automatically downloads the Release binary suited for the current system.
  * Default installation to `/opt/fast-note`, with a shortcut command created at `/usr/local/bin/fast-note`.
  * Configures and starts the Systemd service (`fast-note.service`), enabling auto-start on boot.
  * **Management Commands**: `fast-note [install|uninstall|start|stop|status|update|menu]`

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
    Register an account on your first visit. *(To disable registration, set `user.register-is-enable: false` in the configuration file.)*
3.  **Configure Client**:
    Log in to the management panel and click **"Copy API Config"**.
4.  **Connect Obsidian**:
    Open the Obsidian plugin settings page and paste the configuration information you just copied.


## ⚙️ Configuration Notes

The default configuration file is `config.yaml`. The program will automatically search for it in the **root directory** or the **config/** directory.

View the full configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx Reverse Proxy Example

View the full configuration example: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🔗 Related Resources

  * [Obsidian Fast Note Sync Plugin (Client Plugin)](https://github.com/haierkeys/obsidian-fast-note-sync)