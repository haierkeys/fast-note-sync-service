[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

If you have any questions, please create an [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new), or join the Telegram group for help: [https://t.me/obsidian_users](https://t.me/obsidian_users)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>High-performance, low-latency note synchronization, online management, remote REST API and other service platforms</strong>
  <br>
  <em>Built with Golang + Websocket + Sqlite + React</em>
</p>

<p align="center">
  Data provision needs to be used with the client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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
    * Provides standard REST API interfaces, supporting programmatic operations (e.g., automation scripts, AI assistant integration) for CRUD operations on Obsidian notes.
    * For details, please refer to the [REST API Documentation](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/REST_API.md).
* **💻 Web Admin Panel**:
  * Built-in modern management interface for easily creating users, generating plugin configurations, managing vaults and note content.
* **🔄 Multi-device Note Sync**:
    * Supports automatic **Vault** creation.
    * Supports note management (CRUD), with changes distributed to all online devices in real-time within milliseconds.
* **🖼️ Attachment Sync Support**:
    * Perfect support for syncing non-note files such as images.
    * Supports chunked upload and download for large attachments, with configurable chunk sizes to improve synchronization efficiency.
* **⚙️ Configuration Sync**:
    * Supports synchronization of `.obsidian` configuration files.
    * Supports synchronization of `PDF` progress status.
* **📝 Note History**:
    * View the historical modification versions of each note in the Web page or the plugin side.
    * (Server v1.2+ required)
* **🗑️ Recycle Bin**:
    * Supports automatic entry into the recycle bin after note deletion.
    * Supports restoring notes from the recycle bin. (Attachment restoration will be added later)

* **🚫 Offline Sync Strategy**:
    * Supports automatic merging of note edits made offline. (Requires plugin settings)
    * Offline deletion, auto-completion or deletion synchronization after reconnection. (Requires plugin settings)

## ☕ Sponsorship and Support

- If you find this plugin useful and want it to continue being developed, please support me in the following ways:

  | Ko-fi *Non-Mainland China*                                                                                           |    | WeChat Pay *Mainland China*                                        |
  |----------------------------------------------------------------------------------------------------------------------|----|--------------------------------------------------------------------|
  | [<img src="https://ik.imagekit.io/haierkeys/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="https://ik.imagekit.io/haierkeys/wxds.png" height="150"> |

  - Supported List:
    - https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md

## ⏱️ Changelog

- ♨️ [View Changelog](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/CHANGELOG.en.md)

## 🗺️ Roadmap

We are continuously improving, here are the future development plans:

- [ ] **Sharing Feature**: Support for sharing notes.
- [ ] **MCP Support**: Add support for AI MCP related functions.
- [ ] **Directory Sync**: Support for CRUD operations on directories.
- [ ] **Git Version Control Integration**: Provides safer version backtracking for notes.
- [ ] **Cloud Storage and Backup Strategy**:
    - [ ] Custom backup strategy configuration.
    - [ ] Multi-protocol adaptation: S3 / Minio / Cloudflare R2 / Alibaba Cloud OSS / WebDAV.

> **If you have suggestions for improvement or new ideas, feel free to share them with us by submitting an issue — we will carefully evaluate and adopt appropriate suggestions.**

## 🚀 Quick Deployment

We provide multiple installation methods, we recommend using the **One-click Script** or **Docker**.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

**Main actions of the script:**

  * Automatically downloads the Release binary file adapted to the current system.
  * Installed to `/opt/fast-note` by default, and a shortcut command is created at `/usr/local/bin/fast-note`.
  * Configures and starts the Systemd service (`fast-note.service`) for auto-start at boot.
  * **Management command**: `fast-note [install|uninstall|start|stop|status|update|menu]`

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

Download the latest version for your system from [Releases](https://github.com/haierkeys/fast-note-sync-service/releases), unzip and run:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 User Guide

1.  **Access Admin Panel**:
    Open `http://{Server IP}:9000` in your browser.
2.  **Initial Setup**:
    Registration is required for the first visit. *(To disable the registration function, please set `user.register-is-enable: false` in the configuration file)*
3.  **Configure Client**:
    Log in to the management panel and click **"Copy API Configuration"**.
4.  **Connect to Obsidian**:
    Open the Obsidian plugin settings page and paste the configuration information you just copied.


## ⚙️ Configuration

The default configuration file is `config.yaml`, and the program will automatically search for it in the **root directory** or **config/** directory.

View full configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx Reverse Proxy Configuration Example

View full configuration example: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🔗 Related Resources

  * [Obsidian Fast Note Sync Plugin (Client Plugin)](https://github.com/haierkeys/obsidian-fast-note-sync)