[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

If you have any questions, please create an [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new), or join the Telegram group for help: [https://t.me/obsidian_users](https://t.me/obsidian_users)

In mainland China, it is recommended to use the Tencent `cnb.cool` mirror: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>High-performance, low-latency note synchronization, online management, and remote REST API service platform.</strong>
  <br>
  <em>Built with Golang + Websocket + Sqlite + React</em>
</p>

<p align="center">
  Must be used with the client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
</p>

<div align="center">
  <div align="center">
    <a href="/docs/images/vault.png"><img src="/docs/images/vault.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="/docs/images/attach.png"><img src="/docs/images/attach.png" alt="fast-note-sync-service-preview" width="400" /></a>
    </div>
  <div align="center">
    <a href="/docs/images/note.png"><img src="/docs/images/note.png" alt="fast-note-sync-service-preview" width="400" /></a>
    <a href="/docs/images/setting.png"><img src="/docs/images/setting.png" alt="fast-note-sync-service-preview" width="400" /></a>
  </div>
</div>

---

## ✨ Core Features

* **🚀 REST API Support**:
    * Provides standard REST API interfaces for programmatic access (e.g., automation scripts, AI assistant integration) to perform CRUD operations on Obsidian notes.
    * For details, please refer to the [RESTful API Documentation](/docs/REST_API.md) or [OpenAPI Documentation](/docs/swagger.yaml).
* **💻 Web Admin Panel**:
  * Built-in modern management interface for easy user creation, plugin configuration generation, and repository/note content management.
* **🔄 Multi-device Note Sync**:
    * Supports automatic **Vault** creation.
    * Supports note management (Add, Delete, Modify, Search), with millisecond-level real-time distribution of changes to all online devices.
* **🖼️ Attachment Sync Support**:
    * Perfect support for syncing non-note files such as images.
    * Supports chunked upload and download for large attachments, with configurable chunk sizes to improve efficiency.
* **⚙️ Config Sync**:
    * Supports synchronization of `.obsidian` configuration files.
    * Supports `PDF` progress status synchronization.
* **📝 Note History**:
    * View historical versions of each note on the Web page or within the plugin.
    * (Requires server v1.2+)
* **🗑️ Recycle Bin**:
    * Deleted notes automatically go to the recycle bin.
    * Supports restoring notes from the recycle bin. (Attachment recovery will be added in future updates)

* **🚫 Offline Sync Strategy**:
    * Supports automatic merging of offline edits. (Requires plugin settings)
    * Offline deletion: automatically supplements or deletes sync upon reconnection. (Requires plugin settings)

* **🔗 Sharing Functionality**:
    * Create/cancel note sharing.
    * Automatically parses images, audio, video, and other attachments referenced in shared notes.
    * Provides sharing access statistics.
* **📂 Directory Sync**:
    * Supports synchronization of folder Create/Rename/Move/Delete operations.

* **🌳 Git Automation**:
    * Automatically updates and pushes to remote Git repositories when attachments or notes change.
    * Automatically releases system memory after tasks are completed.

* **☁️ Multi-Storage Backup and Unidirectional Mirror Sync**:
    * Adapts to S3/OSS/R2/WebDAV/Local and other storage protocols.
    * Supports full/incremental ZIP scheduled archive backups.
    * Supports unidirectional mirror synchronization of Vault resources to remote storage.
    * Automatically cleans up expired backups with customizable retention days.

## ☕ Sponsorship and Support

- If you find this plugin useful and want to support its continued development, please consider:

  | Ko-fi *Outside Mainland China*                                                                    |    | Wechat Scan *Inside Mainland China*            |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="/docs/images/wxds.png" height="150"> |

  - Benefactor List:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.en.md">Support.en.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.en.md">Support.en.md (cnb.cool mirror)</a>

## ⏱️ Changelog

- ♨️ [View Changelog](/docs/CHANGELOG.en.md)

## 🗺️ Roadmap

We are continuously improving. Here are the future development plans:


- [ ] **🤖 MCP Support**: Add support for AI MCP-related features.
- [ ] **Support for more database types**

> **If you have suggestions for improvement or new ideas, please share them with us by submitting an issue — we will carefully evaluate and adopt suitable suggestions.**

## 🚀 Quick Deployment

We provide multiple installation methods; **One-click script** or **Docker** is recommended.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

In China, you can use the Tencent `cnb.cool` mirror:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**Main actions of the script:**

  * Automatically downloads the Release binary for the current system.
  * Defaults to install in `/opt/fast-note` and creates a global shortcut command `fns` in `/usr/local/bin/fns`.
  * Configures and starts the Systemd (Linux) or Launchd (macOS) service for auto-start on boot.
  * **Management commands**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **Interactive menu**: Run `fns` directly to enter the interactive menu for installation/upgrade, service control, boot configuration, and switching between GitHub/CNB mirrors.

-----

### Method 2: Docker Deployment

#### Docker Run

```bash
# 1. Pull the image
docker pull haierkeys/fast-note-sync-service:latest

# 2. Start the container
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 \
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
      - "9000:9000"  # RESTful API & WebSocket port; /api/user/sync is the WebSocket interface address
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

## 📖 User Guide

1.  **Access Admin Panel**:
    Open `http://{Server_IP}:9000` in your browser.
2.  **Initial Setup**:
    Register an account on first visit. *(To disable registration, set `user.register-is-enable: false` in the configuration file)*
3.  **Configure Client**:
    Log in to the admin panel and click **"Copy API Config"**.
4.  **Connect Obsidian**:
    Open the Obsidian plugin settings page and paste the configuration information.


## ⚙️ Configuration

The default configuration file is `config.yaml`. The program automatically looks for it in the **root directory** or **config/** directory.

View the full configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx Reverse Proxy Example

View the full configuration example: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🔗 Clients & Plugins

* Obsidian Fast Note Sync Plugin
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool mirror](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* Third-party Clients
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) A command-line client for high-performance real-time synchronization based on Python and the FNS WS interface, suitable for Linux server environments without a GUI (such as OpenClaw), providing equivalent synchronization capabilities to Obsidian desktop/mobile.