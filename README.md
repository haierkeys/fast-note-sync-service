[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

For issues, please open a new [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new), or join our Telegram group for help: [https://t.me/obsidian_users](https://t.me/obsidian_users)

For users in mainland China, we recommend using the Tencent `cnb.cool` mirror: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)



<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>High-performance, low-latency note sync, online management, and remote REST API service platform</strong>
  <br>
  <em>Built with Golang + WebSocket + SQLite + React</em>
</p>

<p align="center">
  Data access requires the client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

## 🎯 Core Features

* **🧰 MCP (Model Context Protocol) Native Support**:
  * `FNS` can serve as an MCP server, connecting to compatible AI clients such as `Cherry Studio` and `Cursor`, enabling AI to read and write private notes and attachments, with all changes synced to all devices in real time via WebSocket.
* **🚀 REST API Support**:
  * Provides standard REST API interfaces, supporting programmatic access (e.g., automation scripts, AI assistant integration) for CRUD operations on Obsidian notes.
  * See the [RESTful API Documentation](/docs/REST_API.md) or [OpenAPI Documentation](/docs/swagger.yaml).
* **💻 Web Management Panel**:
  * Built-in modern management interface to easily create users, generate plugin configurations, manage vaults, and note content.
* **🔄 Multi-device Note Sync**:
  * Supports automatic **Vault** creation.
  * Supports note management (add, delete, update, query), with changes distributed in real-time to all online devices within milliseconds.
* **🖼️ Attachment Sync Support**:
  * Full support for syncing non-note files such as images.
  * Supports chunked upload/download for large attachments, with configurable chunk sizes for improved sync efficiency.
* **⚙️ Config Sync**:
  * Supports synchronization of `.obsidian` configuration files.
  * Supports `PDF` reading progress synchronization.
* **📝 Note History**:
  * View the historical revision versions of each note on the web page and the plugin side.
  * (Requires server v1.1+)
* **🗑️ Recycle Bin**:
  * Supports automatic movement of deleted notes to the recycle bin.
  * Supports restoring notes from the recycle bin. (Attachment recovery will be added in subsequent updates)

* **🚫 Offline Sync Strategy**:
  * Supports automatic merging of notes edited offline. (Requires plugin-side settings)
  * Offline deletions are automatically supplemented or synced upon reconnection. (Requires plugin-side settings)

* **🔗 Sharing Feature**:
  * Create/cancel note sharing.
  * Automatically resolves referenced images, audio, video, and other attachments in shared notes.
  * Provides sharing access statistics.
  * Supports setting an access password for shared notes.
  * Supports generating short links for shared notes.
* **📂 Directory Sync**:
  * Supports folder create/rename/move/delete synchronization.

* **🌳 Git Automation**:
  * Automatically updates and pushes to a remote Git repository when attachments and notes change.
  * Automatically releases system memory after tasks complete.

* **☁️ Multi-storage Backup & Unidirectional Mirror Sync**:
  * Compatible with multiple storage protocols: S3/OSS/R2/WebDAV/Local.
  * Supports full/incremental ZIP scheduled archive backups.
  * Supports unidirectional mirror sync of Vault resources to remote storage.
  * Automatic cleanup of expired backups with configurable retention days.

* **🗄️ Multi-database Support**:
  * Native support for SQLite, MySQL, PostgreSQL, and other mainstream databases to meet different deployment needs from individuals to teams.

## ☕ Sponsorship & Support

- If you find this project useful and want to support its continued development, please support me via:

  | Ko-fi *(Outside China)*                                                                          |    | WeChat Pay *(China)*                           |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="/docs/images/wxds.png" height="150"> |

  - Supporter list:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md (cnb.cool mirror)</a>

## ⏱️ Changelog

- ♨️ [View Changelog](/docs/CHANGELOG.md)

## 🗺️ Roadmap

- [ ] Add **Mock** testing covering all levels.
- [ ] Support WebSocket **Protobuf** transmission format to enhance synchronization efficiency.
- [ ] Backend support for querying various operational logs, such as sync logs and operation logs.
- [ ] Isolate and optimize the existing authorization mechanism to improve overall security.
- [ ] Enable real-time note updates in the Web GUI.
- [ ] Add client-to-client point-to-point messaging (non-note/attachment, similar to LocalSend; no client-side storage, server-side storage optional).
- [ ] Improve various help documents.
- [ ] Support more intranet penetration (relay gateway) solutions.
- [ ] Quick deployment plan
  * Deploy FNS server by only providing the server address (public), username, and password.
- [ ] Optimize the existing offline note merging plan and add a conflict resolution mechanism.

We are continuously improving. Here are our future development plans:

> **If you have suggestions for improvements or new ideas, feel free to share them by opening an issue — we will carefully evaluate and adopt suitable proposals.**

## 🚀 Quick Deployment

We provide multiple installation methods. **One-click script** or **Docker** is recommended.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

Users in China can use the Tencent `cnb.cool` mirror:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**Main script behavior:**

  * Automatically downloads the Release binary for the current system.
  * Installed to `/opt/fast-note` by default, with a global shortcut command `fns` created at `/usr/local/bin/fns`.
  * Configures and starts a Systemd (Linux) or Launchd (macOS) service for automatic startup on boot.
  * **Management commands**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **Interactive menu**: Running `fns` directly opens an interactive menu supporting install/upgrade, service control, boot startup configuration, and switching between GitHub/CNB mirrors.

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
      - "9000:9000"  # RESTful API & WebSocket port; /api/user/sync is the WebSocket endpoint
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

Download the latest version for your system from [Releases](https://github.com/haierkeys/fast-note-sync-service/releases), extract and run:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 Usage Guide

1.  **Access the Management Panel**:
    Open `http://{ServerIP}:9000` in your browser.
2.  **Initial Setup**:
    Register an account on first access. *(To disable registration, set `user.register-is-enable: false` in the configuration file)*
3.  **Configure the Client**:
    Log in to the management panel and click **"Copy API Configuration"**.
4.  **Connect Obsidian**:
    Open the Obsidian plugin settings page and paste the copied configuration.


## ⚙️ Configuration

The default configuration file is `config.yaml`. The program will automatically look for it in the **root directory** or the **config/** directory.

View the full configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx Reverse Proxy Configuration Example

View the full configuration example: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) Support

FNS natively supports **MCP (Model Context Protocol)**.

`FNS` can serve as an MCP server, connecting to compatible AI clients such as `Cherry Studio` and `Cursor`, enabling AI to read and write private notes and attachments, with all changes synced to all devices in real time via WebSocket.

### Connection Configuration (SSE Mode)

FNS provides the MCP interface via the **SSE protocol**, with the following general parameters:
- **Endpoint URL**: `http://<your-server-ip-or-domain>:<port>/api/mcp/sse`
- **Auth Header**: `Authorization: Bearer <your-api-token>` (obtained from the "Copy API Configuration" in the WebGUI)
- **Optional Header**: `X-Default-Vault-Name: <VaultName>` (used to specify the default vault for MCP operations if the `vault` parameter is not provided in the tool call)


#### Example: Cherry Studio / Cursor / Cline, etc.

Please refer to the following configuration in your MCP client:
*(Note: Replace `<ServerIP>`, `<Port>`, `<Token>`, and `<VaultName>` with your actual information)*

```json
{
  "mcpServers": {
    "fns": {
      "url": "http://<ServerIP>:<Port>/api/mcp/sse",
      "type": "sse",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer <Token>",
        "X-Default-Vault-Name": "<VaultName>"
      }
    }
  }
}
```

## 🔗 Clients & Client Plugins

* Obsidian Fast Note Sync Plugin
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool mirror](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* Third-party Clients
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) — A bidirectional real-time sync command-line client based on Python and the FNS WS interface, suitable for headless Linux server environments (such as OpenClaw), achieving sync capabilities equivalent to Obsidian desktop/mobile.