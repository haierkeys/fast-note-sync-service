[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

If you have any questions, please create an [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new) or join our Telegram group for help: [https://t.me/obsidian_users](https://t.me/obsidian_users)

For users in Mainland China, it is recommended to use the Tencent cnb.cool mirror repository: [https://cnb.cool/haierkeys/fast-note-sync-service](https://cnb.cool/haierkeys/fast-note-sync-service)


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
  <em>Built with Golang + Websocket + React</em>
</p>

<p align="center">
  Data provisioning requires integration with the client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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

## 🎯 Key Features

* **🧰 Native MCP (Model Context Protocol) Support**:
  * `FNS` can be integrated as an MCP server into compatible AI clients such as `Cherry Studio` and `Cursor`. This enables the AI to read and write your personal notes and attachments, with all changes synchronized across all clients in real-time.
* **🚀 REST API Support**:
  * Provides standard REST API interfaces to support programmatic CRUD operations on Obsidian notes (e.g., via automation scripts or AI assistant integrations).
  * For details, please refer to the [RESTful API Documentation](docs/REST_API.md) or [OpenAPI Documentation](docs/swagger.yaml).
* **💻 Web Admin Panel**:
  * Built-in modern administration interface to easily create users, generate plugin configurations, and manage vaults and note content.
  * Supports WebGUI OIDC login. See the [OIDC login runbook](docs/runbook/OIDC.en.md).
* **🔄 Multi-Device Note Synchronization**:
  * Supports automatic creation of `Vaults`.
  * Supports note management (CRUD), with millisecond-level real-time change distribution to all online devices.
* **🖼️ Attachment Sync Support**:
  * Perfectly supports synchronization of non-note files such as images.
  * Supports chunked upload and download for large attachments, with configurable chunk sizes to improve sync efficiency.
* **⚙️ Configuration Sync**:
  * Supports synchronization of `.obsidian` configuration files.
  * Supports `PDF` reading progress synchronization.
* **📝 Note History**:
  * View historical revision versions of each note via the Web UI or client plugin.
  * (Requires server v1.2+)
* **🗑️ Trash Bin**:
  * Deleted notes automatically go to the trash bin.
  * Supports restoring notes from the trash. (Attachment restoration feature will be added sequentially in the future.)

* **🚫 Offline Synchronization Strategy**:
  * Supports automatic merging of offline note edits. (Requires client plugin configuration)
  * Offline deletions are automatically resolved (re-synced or deleted) upon reconnection. (Requires client plugin configuration)

* **🔗 Sharing Feature**:
  * Easily share or unshare notes.
  * Automatically parses images, audio, video, and other attachments referenced in shared notes.
  * Provides sharing access statistics.
  * Allows password protection for shared notes.
  * Supports generating short URLs for shared notes.
* **📂 Folder Sync**:
  * Supports folder synchronization for creation, renaming, moving, and deletion.

* **🌳 Git Automation**:
  * Automatically updates and pushes changes to a remote Git repository when notes or attachments change.
  * Automatically releases system memory after tasks complete.

* **☁️ Multi-Storage Backup & One-way Mirror Sync**:
  * Adapts to S3, OSS, R2, WebDAV, local filesystem, and other storage protocols.
  * Supports scheduled full or incremental ZIP archive backups.
  * Supports one-way mirror synchronization of Vault resources to remote storage.
  * Automatically cleans up expired backups with custom retention days.

* **🗄️ Multi-Database Support**:
  * Native support for mainstream databases including SQLite, MySQL, and PostgreSQL, meeting deployment needs from individuals to teams.

## ☕ Sponsorship and Support

- If you find this plugin useful and want to support its ongoing development, please support me in the following ways:

  | Ko-fi *Outside China*                                                                            |    | WeChat Pay *China*                             |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="/docs/images/wxds.png" height="150"> |

  - Sponsor List:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.en.md">Support.en.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.en.md">Support.en.md (cnb.cool mirror repository)</a>

## ⏱️ Changelog

- ♨️ [Click to View Changelog](docs/CHANGELOG.en.md)

## 🗺️ Roadmap

- [ ] Add support for WebSocket `Protobuf` transfer format to enhance synchronization efficiency.
- [ ] Isolate and optimize the existing authorization mechanism to improve overall security.
- [ ] Add real-time note updates to the Web UI.
- [ ] Add peer-to-peer message transfer between clients (non-notes & attachments, similar to LocalSend; client-side storage is not supported, but server-side is).
- [ ] Improve various help documentations.
- [ ] Provide support for more intranet penetration (relay gateway) options.
- [ ] Quick deployment plan:
  * Deploy the FNS server by simply providing the server address (public IP) and account credentials.
- [ ] Optimize the existing offline note merging scheme and introduce conflict resolution mechanisms.

We are continuously improving, and here is our future development roadmap:

> **If you have suggestions for improvement or new ideas, feel free to share them with us by submitting an issue — we will carefully evaluate and adopt suitable ones.**

## 🚀 Quick Deployment

We provide multiple installation methods. We highly recommend using the **One-click Script** or **Docker**.

### Method 1: One-click Script (Recommended)

Automatically detects the system environment and completes the installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

For users in China, you can use the Tencent `cnb.cool` mirror source:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**Main Script Behaviors:**

  * Automatically downloads the Release binary adapted to the current system.
  * Installs to `/opt/fast-note` by default and creates a global shortcut command `fns` in `/usr/local/bin/fns`.
  * Configures and starts Systemd (Linux) or Launchd (macOS) services to enable startup on boot.
  * **Management commands**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **Interactive menu**: Run `fns` directly to enter the interactive menu, supporting install/upgrade, service control, startup configuration, and switching between GitHub and CNB mirrors.

-----

### Method 2: Docker Deployment

#### Docker Run

```bash
# 1. Pull Image
docker pull haierkeys/fast-note-sync-service:latest

# 2. Start Container
docker run -tid --name fast-note-sync-service \
    -p 9000:9000 \
    -v /data/fast-note-sync/storage/:/fast-note-sync/storage/ \
    -v /data/fast-note-sync/config/:/fast-note-sync/config/ \
    haierkeys/fast-note-sync-service:latest
```

#### Docker Compose

Create the `docker-compose.yaml` file:

```yaml
version: '3'
services:
  fast-note-sync-service:
    image: haierkeys/fast-note-sync-service:latest
    container_name: fast-note-sync-service
    restart: always
    ports:
      - "9000:9000"  # RESTful API & WebSocket port (where /api/user/sync is the WebSocket interface address)
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
    Register an account on your first visit. *(To disable registration, set `user.register-is-enable: false` in the configuration file)*
3.  **Configure Client**:
    Log in to the admin panel and click **"Copy API Config"**.
4.  **Connect to Obsidian**:
    Open the Obsidian plugin settings page and paste the configuration information you just copied.


## ⚙️ Configuration Description

The default configuration file is `config.yaml`. The application automatically searches in the **root directory** or the **config/** directory.

View the complete configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx Reverse Proxy Example

View the complete configuration example: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) Support

FNS now natively supports **MCP (Model Context Protocol)** and provides both **SSE** and **StreamableHTTP** transport protocols.

You can integrate FNS as an MCP server directly into compatible AI clients like Cherry Studio, Cursor, Claude Code, and hermes-agent. Once integrated, the AI will gain the ability to read and write your personal notes and attachments. In addition, all modifications made via MCP will be synchronized to your client devices in real-time via WebSocket.

For OAuth-protected MCP deployments with Stytch, see [docs/runbook/mcp-oauth-stytch.md](docs/runbook/mcp-oauth-stytch.md) ([简体中文](docs/runbook/mcp-oauth-stytch.zh-CN.md), [繁體中文](docs/runbook/mcp-oauth-stytch.zh-TW.md)).

### Common Request Header Parameters

The following request headers are supported regardless of the transport mode used:

- **Authorization Header**: `Authorization: Bearer <Your API Token>` (obtained from the Copy API Config option in the WebGUI)
- **Optional Header**: `X-Default-Vault-Name: <Vault Name>` (specifies the default Vault name for MCP operations if no `vault` parameter is provided during a tool call)
- **Optional Header**: `X-Client: <Client Type>` (the type of client connecting to MCP, e.g., `Cherry Studio`, `OpenClaw`)
- **Optional Header**: `X-Client-Version: <Client Version>` (the version of the client connecting, e.g., `1.1`)
- **Optional Header**: `X-Client-Name: <Client Name>` (the name of the client connecting, e.g., `Mac`)

---

### Integration Config: StreamableHTTP Mode (Recommended)

StreamableHTTP is the standard transport protocol in the MCP ecosystem. A single endpoint can handle all requests, making it more firewall-friendly. It is natively supported by newer MCP clients (such as Claude Code and hermes-agent).

- **Interface Address**: `http://<Your_Server_IP_or_Domain>:<Port>/api/mcp`
- **Request Method**: `POST` (send requests/notifications), `GET` (listen for server-side push notifications), `DELETE` (terminate sessions)

#### Example: Claude Code / hermes-agent / Cursor, etc.

*(Note: Please replace `<ServerIP>`, `<Port>`, `<Token>`, and `<VaultName>` with your actual information)*

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

### Integration Config: SSE Mode (Backward Compatible)

SSE mode is the legacy transport protocol. It is fully retained to maintain backward compatibility, suitable for MCP clients that only support SSE (such as Cherry Studio).

- **Interface Address**: `http://<Your_Server_IP_or_Domain>:<Port>/api/mcp/sse`

#### Example: Cherry Studio / Cline, etc.

*(Note: Please replace `<ServerIP>`, `<Port>`, `<Token>`, and `<VaultName>` with your actual information)*

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

## 🔗 Clients, Client Plugins & Collaboration Projects

* Obsidian Fast Note Sync Plugin
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool mirror repository](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* Third-Party Clients
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) A command-line client implemented in Python based on the FNS WebSocket synchronization protocol. It enables real-time, bi-directional sync, suitable for GUI-less Linux server environments (such as OpenClaw), providing the same sync capabilities as the Obsidian desktop and mobile clients.
  * [go-fast-note-sync](https://github.com/erichll/go-fast-note-sync) A Go CLI background synchronization daemon implemented in Go based on the FNS WebSocket sync protocol, primarily targetting headless Linux environments while also supporting macOS and Windows.
  * [Fast-note-sync-docker](https://github.com/youpingfang/obsidian-note-sync-docker) A rapid containerized deployment solution based on Docker, Python, and the FNS WebSocket sync protocol to sync vaults and configuration files to a remote server.
* Collaboration Projects
  * [Share to Save](https://github.com/chenxiccc/Obsidian-Share-to-Save) An Obsidian plugin that automatically downloads shared web page URLs as Markdown notes.
