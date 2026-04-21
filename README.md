[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

For any questions, please create a new [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new), or join the Telegram chat group for help: [https://t.me/obsidian_users](https://t.me/obsidian_users)

For mainland China, the Tencent `cnb.cool` mirror repository is recommended: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


<h1 align="center">Fast Note Sync Service</h1>

<p align="center">
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/release/haierkeys/fast-note-sync-service?style=flat-square" alt="release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/releases"><img src="https://img.shields.io/github/v/tag/haierkeys/fast-note-sync-service?label=release-alpha&style=flat-square" alt="alpha-release"></a>
    <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/LICENSE"><img src="https://img.shields.io/github/license/haierkeys/fast-note-sync-service?style=flat-square" alt="license"></a>
    <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square" alt="Go">
</p>

<p align="center">
  <strong>High-performance, low-latency note syncing, online management, and remote REST API service platform</strong>
  <br>
  <em>Built with Golang + Websocket + React</em>
</p>

<p align="center">
  Data provision requires the use of the client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

* **🧰 Native MCP (Model Context Protocol) Support**:
  * `FNS` can act as an MCP server connecting to `Cherry Studio`, `Cursor`, and other compatible AI clients. This grants AI the ability to read and write your private notes and attachments, with all changes syncing to all clients in real time.
* **🚀 REST API Support**:
  * Provides standard REST API interfaces, supporting automated program access (e.g., automation scripts, AI assistant integration) for CRUD operations on Obsidian notes.
  * For more details, please refer to the [RESTful API Documentation](/docs/REST_API.md) or [OpenAPI Documentation](/docs/swagger.yaml).
* **💻 Web Management Panel**:
  * Built-in modern management interface to easily create users, generate plugin configurations, and manage vaults and note contents.
* **🔄 Multi-device Note Syncing**:
  * Supports automatic **Vault** creation.
  * Supports note management (Add, Delete, Modify, Read) with millisecond-level real-time distribution of changes to all online devices.
* **🖼️ Attachment Syncing Support**:
  * Perfect support for syncing non-note files like images.
  * Supports chunked upload and download of large attachments, with configurable chunk sizes, improving syncing efficiency.
* **⚙️ Configuration Syncing**:
  * Supports syncing `.obsidian` configuration files.
  * Supports syncing `PDF` progress states.
* **📝 Note History**:
  * Ability to view historical modification versions of each note via the Web page and plugin client.
  * (Requires Server v1.2+)
* **🗑️ Recycle Bin**:
  * Supports automatic transfer of notes to the recycle bin upon deletion.
  * Supports recovering notes from the recycle bin. (Attachment recovery features will be added progressively).

* **🚫 Offline Sync Strategy**:
  * Supports automatic merging of offline note edits. (Requires setup on the plugin client).
  * Offline deletion automatically synchronizes with server padding or deletions after reconnection. (Requires setup on the plugin client).

* **🔗 Share Feature**:
  * Ability to Create/Cancel note sharing.
  * Automatically parses attachments such as images, audio, and video referenced in shared notes.
  * Provides sharing access statistics.
  * Ability to set a password for shared notes.
  * Ability to generate short links for shared notes.
* **📂 Directory Syncing**:
  * Supports Create/Rename/Move/Delete syncing for folders.

* **🌳 Git Automation**:
  * Automatically updates and pushes to the remote Git repository when attachments or notes undergo changes.
  * Automatically releases system memory after the task strictly finishes.

* **☁️ Multi-Storage Backup & One-way Mirror Syncing**:
  * Adapts to S3/OSS/R2/WebDAV/Local and other storage protocols.
  * Supports full/incremental ZIP scheduled archive backups.
  * Supports one-way mirror syncing of Vault resources to remote storage.
  * Automatically cleans up expired backups, with support for custom retention days.

* **🗄️ Multi-Database Support**:
  * Natively supports mainstream databases such as SQLite, MySQL, PostgreSQL, meeting deployment needs ranging from individuals to teams.

## ☕ Sponsorship & Support

- If you find this plugin useful and want its development to continue, please support me via the following channels:

  | Ko-fi *Non-China Region*                                                                               |    | WeChat QR Donation *China Region*                        |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | OR | <img src="/docs/images/wxds.png" height="150"> |

  - Supported List:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.en.md">Support.en.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.en.md">Support.en.md (cnb.cool Mirror)</a>

## ⏱️ Changelog

- ♨️ [View Changelog](/docs/CHANGELOG.en.md)

## 🗺️ Roadmap

- [ ] Add **Mock** testing covering all levels.
- [ ] Add WebSocket `Protobuf` transmission format support, enhancing synchronization efficiency.
- [ ] The backend to include queries for various operational logs such as sync logs and operation logs.
- [ ] Isolate and optimize the current authorization mechanism to elevate overall security.
- [ ] Add WebGui note real-time update capability.
- [ ] Add client Peer-to-Peer message transmission (non-note & attachments, similar to localsend; not saved closely on the client, saves to the server).
- [ ] Enhance various help documents.
- [ ] Support more Intranet Penetration (Relay gateway).
- [ ] Quick deployment plan:
  * Deploy FNS Server securely with just the server's public IP address and account credentials.
- [ ] Optimize the current offline note merging scheme and introduce conflict-handling mechanisms.

We are continually improving. Here are our future development plans:

> **If you have improvement suggestions or new ideas, please submit an issue to share them with us. We will sincerely evaluate and adopt appropriate suggestions.**

## 🚀 Quick Deployment

We offer multiple installation methods. We recommend utilizing the **One-click Script** or **Docker**.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment, completes the installation, and registers the service.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

Users in China can utilize the Tencent `cnb.cool` mirror source:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```

**Main Script Actions:**

  * Automatically downloads the optimal Release binary file for your system.
  * Default installation path is `/opt/fast-note`, creating a global quick command abstractly named `fns` in `/usr/local/bin/fns`.
  * Configures and launches Systemd (Linux) or Launchd (macOS) services to realize auto-start on boot.
  * **Management Commands**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **Interactive Menu**: Run `fns` directly to enter an interactive menu enabling installation/upgrade, service control, auto-start configuration, and switching between GitHub / CNB mirrors.

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
      - "9000:9000"  # RESTful API & WebSocket ports where /api/user/sync is the WebSocket interface address
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

Download the latest version corresponding to your system from [Releases](https://github.com/haierkeys/fast-note-sync-service/releases), unzip, and run:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 Usage Guide

1.  **Access the Management Panel**:
    Open `http://{Server IP}:9000` via your browser.
2.  **Initial Setup**:
    Register an account on your first visit. *(To disable registration, configure `user.register-is-enable: false` in the settings configuration file)*
3.  **Configure Client**:
    Log into the Management Panel, and click **"Copy API Configuration"**.
4.  **Connect to Obsidian**:
    Navigate to the Obsidian plugin configuration page, and paste the previously copied configuration details.


## ⚙️ Configuration Instructions

The default configuration file is `config.yaml`. The application will search for it automatically in the **Root Directory** or the **config/** directory.

View a complete configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx Reverse Proxy Configuration Example

View a complete configuration example: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) Support

FNS natively supports **MCP (Model Context Protocol)**.

You can directly incorporate FNS as an MCP server with Cherry Studio, Cursor, and similar compatible AI clients. Once configured, AI attains the capacity to interpret and write within your own private notes and attachments. Furthermore, WebSocket synchronizes all MCP-directed alterations in real time across the entirety of your devices.

### Access Configuration (SSE Mode)

FNS furnishes an MCP interface primarily through the **SSE protocol**. General parameter requirements are as follows:
- **Interface Address**: `http://<Your Server IP or Domain>:<Port>/api/mcp/sse`
- **Authentication Header**: `Authorization: Bearer <Your API Token>` (Obtained from “Copy API Configuration” in the WebGUI).
- **Optional Header**: `X-Default-Vault-Name: <Vault Name>` (Identifies the default vault for MCP operations; utilized if the `vault` parameter is unspecified during a tool call)
- **Optional Header**: `X-Client: <Client Type>` (Relates to the client type connecting the MCP, e.g., Cherry Studio / OpenClaw)
- **Optional Header**: `X-Client-Version: <Client Version>` (Pertains to the client's actual version interfacing with the MCP, e.g., 1.1)
- **Optional Header**: `X-Client-Name: <Client Name>` (Designates the client's name linking with the MCP, e.g., Mac)



#### Example: Cherry Studio / Cursor / Cline, etc.

Kindly incorporate the below JSON inside your MCP client configuration parameters:
*(Note: Please swap `<ServerIP>`, `<Port>`, `<Token>`, and `<VaultName>` with your proper contextual details)*

```json
{
  "mcpServers": {
    "fns": {
      "url": "http://<ServerIP>:<Port>/api/mcp/sse",
      "type": "sse",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer <Token>",
        "X-Default-Vault-Name": "<VaultName>",
        "X-Client": "<Client>",
        "X-Client-Version": "<ClientVersion>",
        "X-Client-Name": "<ClientName>"
      }
    }
  }
}
```

## 🔗 Client & Client Plugins

* Obsidian Fast Note Sync Plugin
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool Mirror](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* Third-Party Clients
  * [FastNodeSync-CLI ](https://github.com/Go1c/FastNodeSync-CLI) Python and FNS WS API grounded bidirectional real-time synchronization CLI client customized for headless Linux server infrastructures (like OpenClaw), providing analogous synchronization faculties equivalent to Obsidian's desktop/mobile clients.