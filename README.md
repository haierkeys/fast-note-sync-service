[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

If you have any questions, please create a new [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new), or join the Telegram group for help: [https://t.me/obsidian_users](https://t.me/obsidian_users)

In Mainland China, it is recommended to use the Tencent `cnb.cool` mirror: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>High-performance, low-latency note synchronization, online management, remote REST API service platform</strong>
  <br>
  <em>Built with Golang + Websocket + Sqlite + React</em>
</p>

<p align="center">
  Data provision requires use with a client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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
    * Provides standard REST API interfaces, supporting programmatic operations (such as automation scripts, AI assistant integration) for CRUD operations on Obsidian notes.
    * For details, please refer to the [REST API Documentation](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/REST_API.md).
* **💻 Web Management Panel**:
  * Built-in modern management interface, easily create users, generate plugin configurations, manage repositories and note content.
* **🔄 Multi-device Note Sync**:
    * Supports automatic **Vault (Repository)** creation.
    * Supports note management (Add, Delete, Modify, Search), with millisecond-level real-time distribution of changes to all online devices.
* **🖼️ Attachment Sync Support**:
    * Perfectly supports synchronization of non-note files such as images.
    * Supports chunked upload and download for large attachments, with configurable chunk size to improve synchronization efficiency.
* **⚙️ Configuration Sync**:
    * Supports synchronization of `.obsidian` configuration files.
    * Supports `PDF` progress status synchronization.
* **📝 Note History**:
    * View the historical modification versions of each note on the Web page and plugin side.
    * (Requires server v1.2+)
* **🗑️ Trash**:
    * Supports automatic movement of deleted notes to the trash.
    * Supports restoring notes from the trash. (Attachment recovery features will be added progressively)

* **🚫 Offline Sync Strategy**:
    * Supports automatic merging of offline note edits. (Requires plugin settings)
    * Supports offline deletion, with automatic synchronization or deletion upon reconnection. (Requires plugin settings)

## ☕ Sponsorship and Support

- If you find this plugin useful and want to support its continued development, please consider the following ways:

  | Ko-fi *International*                                                                            |    | WeChat Pay *China*                             |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="/docs/images/wxds.png" height="150"> |

  - Supported List:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.en.md">Support.en.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.en.md">Support.en.md (cnb.cool mirror)</a>

## ⏱️ Changelog

- ♨️ [View Changelog](/docs/CHANGELOG.en.md)

## 🗺️ Roadmap

We are continuously improving, here are the future development plans:

- [ ] **Sharing Feature**: Support for sharing notes.
- [ ] **MCP Support**: Add support for AI MCP related features.
- [ ] **Directory Sync**: Support for CRUD operations on directories.
- [ ] **Git Version Control Integration**: Provide safer version backtracking for notes.
- [ ] **Cloud Storage and Backup Strategy**:
    - [ ] Custom backup strategy configuration.
    - [ ] Multi-protocol adaptation: S3 / Minio / Cloudflare R2 / Aliyun OSS / WebDAV.

> **If you have suggestions for improvement or new ideas, welcome to share with us by submitting an issue — we will seriously evaluate and adopt suitable suggestions.**

## 🚀 Quick Deployment

We provide multiple installation methods, with **One-click script** or **Docker** being recommended.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

In China, you can use the Tencent `cnb.cool` mirror source:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**Main behaviors of the script:**

  * Automatically downloads the Release binary file adapted to the current system.
  * Installed to `/opt/fast-note` by default, and creates a global shortcut command `fns` at `/usr/local/bin/fns`.
  * Configures and starts the Systemd (Linux) or Launchd (macOS) service, enabling auto-start on boot.
  * **Management Commands**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **Interactive Menu**: Run `fns` to open the interactive menu, which supports installing/upgrading, service control, auto-start configuration, and switching download mirror (GitHub / CNB).

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
    Register an account on the first visit. *(To disable registration, set `user.register-is-enable: false` in the configuration file)*
3.  **Configure Client**:
    Log in to the management panel and click **"Copy API Config"**.
4.  **Connect Obsidian**:
    Open the Obsidian plugin settings page and paste the configuration information you just copied.


## ⚙️ Configuration

The default configuration file is `config.yaml`, and the program will automatically search for it in the **root directory** or **config/** directory.

View full configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx Reverse Proxy Example

View full configuration example: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🔗 Related Resources

  * [Obsidian Fast Note Sync Plugin (Client Plugin)](https://github.com/haierkeys/obsidian-fast-note-sync)
  * [Obsidian Fast Note Sync Plugin (cnb.cool mirror)](https://cnb.cool/haierkeys/obsidian-fast-note-sync)