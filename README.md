[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

If you have any questions, please create a new [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new), or join our Telegram group for help: [https://t.me/obsidian_users](https://t.me/obsidian_users)

For users in Mainland China, the Tencent `cnb.cool` mirror repository is highly recommended: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


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
  <em>Built with Golang + WebSocket + React</em>
</p>

<p align="center">
  Client data provision requires: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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
  * `FNS` can act as an MCP server connected to compatible AI clients like `Cherry Studio` and `Cursor`, enabling AI to read and write private notes and attachments, with all changes synchronized in real time across endpoints.
* **🚀 REST API Support**:
  * Provides standard REST API endpoints, supporting programmatic operations (e.g., automation scripts, AI assistant integrations) to perform CRUD actions on Obsidian notes.
  * Refer to the [RESTful API Docs](/docs/REST_API.md) or [OpenAPI Docs](/docs/swagger.yaml) for details.
* **💻 Web Administration Panel**:
  * Built-in modern dashboard to easily create users, generate plugin configurations, manage repositories, and browse note contents.
* **🔄 Multi-device Note Sync**:
  * Supports automatic **Vault** creation.
  * Supports note management (create, read, update, delete) with millisecond-level real-time updates distributed to all active devices.
* **🖼️ Attachment Sync**:
  * Perfectly supports synchronization of non-note files such as images.
  * Supports chunked upload/download for large attachments, with configurable chunk sizes to boost synchronization efficiency.
* **⚙️ Configuration Sync**:
  * Supports syncing `.obsidian` configurations.
  * Supports syncing `PDF` progress status.
* **📝 Note History**:
  * View the historical modified versions of each note directly from the Web page or the plugin client (requires server v1.2+).
* **🗑️ Trash Bin**:
  * Deleted notes automatically enter the trash bin.
  * Supports restoring notes from the trash bin (attachment restoration will be introduced in subsequent updates).
* **🚫 Offline Sync Strategy**:
  * Supports automatic conflict-resolution merging for offline note edits (requires plugin configuration).
  * Supports automatic cleanup or restoration of offline deletions upon reconnecting (requires plugin configuration).
* **🔗 Sharing Capabilities**:
  * Create/cancel note sharing.
  * Automatically parses attachments referenced in shared notes, such as images, audio, and video.
  * Provides access statistics for shared notes.
  * Set passwords for accessing shared notes.
  * Generate short links for shared notes.
* **📂 Directory Sync**:
  * Supports synchronization of folder creation, renaming, moving, and deletion.
* **🌳 Automated Git Integration**:
  * Automatically updates and pushes changes to a remote Git repository when attachments and notes are modified.
  * Automatically frees up system memory upon task completion.
* **☁️ Multi-storage Backup & One-way Mirror Sync**:
  * Compatible with multiple storage protocols including S3, Alibaba Cloud OSS, Cloudflare R2, WebDAV, and local filesystems.
  * Supports scheduled full/incremental ZIP archive backups.
  * Supports one-way mirror synchronization of Vault resources to remote storage.
  * Automatically prunes expired backups with customizable retention periods.
* **🗄️ Multi-database Support**:
  * Native support for SQLite, MySQL, and PostgreSQL to satisfy different deployment scales from personal to team environments.

## ☕ Sponsorship & Support

- If you find this plugin helpful and wish to support its continued development, please consider sponsoring me through:

  | Ko-fi *Non-Mainland China*                                                                       |    | WeChat Pay *Mainland China*                    |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="/docs/images/wxds.png" height="150"> |

  - List of Supporters:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.en.md">Support.en.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.en.md">Support.en.md (cnb.cool Mirror)</a>

## ⏱️ Changelog

- ♨️ [View Changelog](/docs/CHANGELOG.en.md)

## 🗺️ Roadmap

- [ ] Add support for WebSocket `Protobuf` transmission to enhance sync efficiency.
- [ ] Isolate and optimize authorization mechanisms to improve overall security.
- [ ] Implement real-time WebGui note updates.
- [ ] Add peer-to-peer message transmission between clients (for messages other than notes/attachments, similar to LocalSend; saves to the server, not saved locally).
- [ ] Improve all documentation and help guides.
- [ ] Provide more intranet penetration (relay gateway) integrations.
- [ ] Fast deployment setups: deploy FNS servers easily just by providing the server address, username, and password.
- [ ] Optimize the existing offline note merging strategy and introduce a conflict-resolution system.

We are continuously improving, and the above represents our future roadmap plans:

> **If you have suggestions for improvement or new ideas, feel free to share them by opening an issue—we carefully evaluate and adopt matching ideas.**

## 🚀 Quick Deployment

We offer multiple installation methods. Using the **One-click Script** or **Docker** is highly recommended.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment, performs the installation, and registers the service.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

Chinese users can use the Tencent `cnb.cool` mirror:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```

**Main actions of the script:**

  * Automatically downloads the release binary adapted to the current operating system.
  * Installs to `/opt/fast-note` by default and creates a global shortcut command `fns` under `/usr/local/bin/fns`.
  * Configures and starts the Systemd (Linux) or Launchd (macOS) service for auto-start on boot.
  * **Management Commands**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **Interactive Menu**: Run `fns` directly to enter the interactive menu, supporting installs/upgrades, service control, boot configuration, and toggling between GitHub and CNB mirror sources.

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
      - "9000:9000"  # RESTful API & WebSocket port, where /api/user/sync is the WebSocket endpoint
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

Download the latest release for your operating system from [Releases](https://github.com/haierkeys/fast-note-sync-service/releases), extract it, and run:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 Usage Guide

1.  **Access Admin Panel**:
    Open `http://{Server_IP}:9000` in your browser.
2.  **Initial Settings**:
    Register an account upon your first visit. *(To disable public registration, set `user.register-is-enable: false` in the config file)*
3.  **Configure Client**:
    Log in to the admin panel and click **"Copy API Configuration"**.
4.  **Connect Obsidian**:
    Open Obsidian, go to the plugin settings page, and paste the copied configuration.

## ⚙️ Configuration

The default configuration file is `config.yaml`. The application automatically searches for it in the **root** or **config/** directories.

View full configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx Reverse Proxy Example

View full configuration example: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) Support

FNS now natively supports **MCP (Model Context Protocol)**, providing both **SSE** and **StreamableHTTP** transport protocols.

You can connect FNS as an MCP server directly into compatible AI clients such as Cherry Studio, Cursor, Claude Code, and hermes-agent. Once connected, AI gains the ability to read and write your private notes and attachments. Simultaneously, all edits generated by the MCP integration will sync in real time via WebSockets to all your device terminals.

### Common Header Parameters

Regardless of the transport mode used, the following headers are supported:

- **Authorization Header**: `Authorization: Bearer <Your API Token>` (retrieved from the WebGUI's copy API configuration)
- **Optional Header**: `X-Default-Vault-Name: <Vault Name>` (specifies the default vault for MCP operations. If tool calls do not specify a vault parameter, this value is used)
- **Optional Header**: `X-Client: <Client Type>` (the type of client connecting to MCP, e.g., Cherry Studio / OpenClaw)
- **Optional Header**: `X-Client-Version: <Client Version>` (the client version, e.g., 1.1)
- **Optional Header**: `X-Client-Name: <Client Name>` (the client device name, e.g., Mac)

---

### Integration: StreamableHTTP Mode (Recommended)

StreamableHTTP is the standard transport protocol for the MCP ecosystem. A single endpoint handles all requests, making it fire-wall friendly and natively supported by newer MCP clients (like Claude Code and hermes-agent).

- **Endpoint URL**: `http://<Your_Server_IP_or_Domain>:<Port>/api/mcp`
- **Request Methods**: `POST` (send requests/notifications), `GET` (listen to server-side push notifications), `DELETE` (terminate sessions)

#### Example: Claude Code / hermes-agent / Cursor, etc.

*(Note: Please replace `<ServerIP>`, `<Port>`, `<Token>`, and `<VaultName>` with your actual details)*

```json
{
  "mcpServers": {
    "fns": {
      "url": "http://<ServerIP>:<Port>/api/mcp",
      "type": "http",
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

---

### Integration: SSE Mode (Backward Compatibility)

SSE mode is the legacy transport protocol, fully retained for backward compatibility. It is suitable for MCP clients that only support SSE (such as Cherry Studio).

- **Endpoint URL**: `http://<Your_Server_IP_or_Domain>:<Port>/api/mcp/sse`

#### Example: Cherry Studio / Cline, etc.

*(Note: Please replace `<ServerIP>`, `<Port>`, `<Token>`, and `<VaultName>` with your actual details)*

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

## 🔗 Clients & Client Plugins

* Obsidian Fast Note Sync Plugin
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool Mirror](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* Third-Party Clients
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) A command-line client implementing two-way real-time synchronization based on Python and the FNS WebSocket protocol. Ideal for headless Linux servers (such as OpenClaw), providing sync capabilities equivalent to Obsidian desktop/mobile clients.
  * [go-fast-note-sync](https://github.com/erichll/go-fast-note-sync) A Go CLI background synchronization daemon based on Go and the FNS WebSocket protocol, mainly targeting headless Linux environments, while also supporting macOS and Windows.
  * [Fast-note-sync-docker](https://github.com/youpingfang/obsidian-note-sync-docker) A quick containerized deployment solution based on Docker, Python, and the FNS WebSocket protocol, implementing note vaults and configuration file sync to remote servers.