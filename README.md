[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

If you have any questions, please create an [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new), or join the Telegram group for help: [https://t.me/obsidian_users](https://t.me/obsidian_users)

For Mainland China users, it is recommended to use the Tencent `cnb.cool` mirror: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>High-performance, low-latency note synchronization, online management, and remote REST API service platform</strong>
  <br>
  <em>Built with Golang + Websocket + Sqlite + React</em>
</p>

<p align="center">
  Data synchronization requires the client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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
    * Provides standard REST API interfaces, supporting programmatic CRUD operations on Obsidian notes (e.g., automation scripts, AI assistant integration).
    * For details, please refer to the [REST API Documentation](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/REST_API.md).
* **💻 Web Management Panel**:
  * Built-in modern management interface for easily creating users, generating plugin configurations, and managing repositories and note content.
* **🔄 Multi-device Note Sync**:
    * Supports automatic creation of **Vaults**.
    * Supports note management (CRUD), with millisecond-level real-time distribution of changes to all online devices.
* **🖼️ Attachment Sync Support**:
    * Perfect support for syncing images and other non-note files.
    * Supports chunked upload and download for large attachments, with configurable chunk sizes to improve synchronization efficiency.
* **⚙️ Configuration Sync**:
    * Supports synchronization of `.obsidian` configuration files.
    * Supports synchronization of `PDF` progress status.
* **📝 Note history**:
    * You can view the historical modification versions of each note on the web page and the plugin side.
    * (Requires server v1.2+)
* **🗑️ Recycle Bin**:
    * Supports automatic transfer to the recycle bin after note deletion.
    * Supports recovering notes from the recycle bin. (Attachment recovery features will be added progressively)

* **🚫 Offline Sync Strategy**:
    * Supports automatic merging of offline note edits (requires plugin-side configuration).
    * Offline deletion, automatic completion or deletion synchronization after reconnection (requires plugin-side configuration).

## ☕ Sponsorship and Support

- If you find this plugin useful and want to support its continued development, please support me in the following ways:

  | Ko-fi *Non-China Region*                                                                         |    | WeChat Pay *China Region*                      |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="/docs/images/wxds.png" height="150"> |

  - Supported list:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.en.md">Support.en.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.en.md">Support.en.md (cnb.cool mirror)</a>

## ⏱️ Changelog

- ♨️ [View Changelog](/docs/CHANGELOG.en.md)

## 🗺️ Roadmap

We are continuously improving, here are our future development plans:

- [ ] **Sharing Feature**: Support note sharing.
- [ ] **MCP Support**: Add AI MCP related functionality.
- [ ] **Directory Sync**: Support directory CRUD operations.
- [ ] **Git Version Control Integration**: Provide more secure version tracking for notes.
- [ ] **Cloud Storage and Backup Strategy**:
    - [ ] Customizable backup strategy configuration.
    - [ ] Multi-protocol adaptation: S3 / Minio / Cloudflare R2 / Alibaba Cloud OSS / WebDAV.

> **If you have improvement suggestions or new ideas, feel free to share them with us by submitting an issue - we will carefully evaluate and adopt appropriate suggestions.**

## 🚀 Quick Deployment

We provide various installation methods, with **one-click script** or **Docker** being recommended.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

In China, you can use the Tencent `cnb.cool` mirror source:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**Main script actions:**

  * Automatically downloads the Release binary adapted to the current system.
  * Default installation to `/opt/fast-note`, and creates a global shortcut command `fns` in `/usr/local/bin/fns`.
  * Configures and starts the Systemd (Linux) or Launchd (macOS) service for auto-start on boot.
  * **Management commands**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **Interactive menu**: Run `fns` directly to enter the interactive menu, supporting installation/upgrade, service control, auto-start configuration, and switching between GitHub / CNB mirrors.

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
      - "9000:9000"  # RESTful API & WebSocket port, where /api/user/sync is the WebSocket interface address
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
    Register an account on the first visit. *(To disable registration, set `user.register-is-enable: false` in the configuration file)*
3.  **Configure Client**:
    Log in to the management panel and click **"Copy API Config"**.
4.  **Connect Obsidian**:
    Open the Obsidian plugin settings page and paste the configuration information you just copied.


## ⚙️ Configuration Instructions

The default configuration file is `config.yaml`, which the program will automatically search for in the **root directory** or **config/** directory.

View full configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx Reverse Proxy Configuration Example

View full configuration example: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🔗 Related Resources

  * [Obsidian Fast Note Sync Plugin (Client Plugin)](https://github.com/haierkeys/obsidian-fast-note-sync)
  * [Obsidian Fast Note Sync Plugin (cnb.cool mirror)](https://cnb.cool/haierkeys/obsidian-fast-note-sync)