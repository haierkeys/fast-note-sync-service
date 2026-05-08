[简体中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-CN.md) / [English](https://github.com/haierkeys/fast-note-sync-service/blob/master/README.md) / [日本語](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ja.md) / [한국어](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.ko.md) / [繁體中文](https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/README.zh-TW.md)

If you have any issues, please open a new [issue](https://github.com/haierkeys/fast-note-sync-service/issues/new), or join the Telegram group for help: [https://t.me/obsidian_users](https://t.me/obsidian_users)

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
  <em>Built with Golang + WebSocket + React</em>
</p>

<p align="center">
  Data sync requires the client plugin: <a href="https://github.com/haierkeys/obsidian-fast-note-sync">Obsidian Fast Note Sync Plugin</a>
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
  * `FNS` can act as an MCP server and integrate with compatible AI clients such as `Cherry Studio`, `Cursor`, etc., enabling AI to read and write private notes and attachments, with all changes synced in real time across all devices.
* **🚀 REST API Support**:
  * Provides standard REST API endpoints, supporting programmatic access (e.g., automation scripts, AI assistant integrations) for CRUD operations on Obsidian notes.
  * See [RESTful API Documentation](/docs/REST_API.md) or [OpenAPI Documentation](/docs/swagger.yaml) for details.
* **💻 Web Management Panel**:
  * Built-in modern management interface for easily creating users, generating plugin configurations, and managing vaults and note content.
* **🔄 Multi-device Note Sync**:
  * Supports automatic **Vault** creation.
  * Supports note management (create, delete, update, query) with millisecond-level real-time distribution to all online devices.
* **🖼️ Attachment Sync Support**:
  * Full support for syncing non-note files such as images.
  * Supports chunked upload/download for large attachments with configurable chunk sizes for improved sync efficiency.
* **⚙️ Config Sync**:
  * Supports syncing `.obsidian` configuration files.
  * Supports syncing `PDF` reading progress state.
* **📝 Note History**:
  * View historical revision versions of each note on the web panel and plugin side.
  * (Requires server v1.2+)
* **🗑️ Recycle Bin**:
  * Automatically moves deleted notes to the recycle bin.
  * Supports restoring notes from the recycle bin. (Attachment restore will be added in future updates)

* **🚫 Offline Sync Strategy**:
  * Supports automatic merging of offline note edits. (Requires plugin-side settings)
  * Offline deletions are automatically reconciled upon reconnection. (Requires plugin-side settings)

* **🔗 Sharing Feature**:
  * Create/revoke note sharing links.
  * Automatically resolves referenced images, audio, video, and other attachments in shared notes.
  * Provides sharing access statistics.
  * Supports setting access passwords for shared notes.
  * Supports generating short links for shared notes.
* **📂 Directory Sync**:
  * Supports create/rename/move/delete sync for folders.

* **🌳 Git Automation**:
  * Automatically commits and pushes changes to a remote Git repository when attachments or notes are modified.
  * Automatically releases system memory after tasks complete.

* **☁️ Multi-Storage Backup & One-way Mirror Sync**:
  * Supports multiple storage protocols: S3, OSS, R2, WebDAV, local, and more.
  * Supports scheduled full/incremental ZIP archive backups.
  * Supports one-way mirror sync of Vault resources to remote storage.
  * Automatically cleans up expired backups with configurable retention days.

* **🗄️ Multi-Database Support**:
  * Natively supports SQLite, MySQL, PostgreSQL, and other mainstream databases, catering to both personal and team deployment needs.

## ☕ Sponsorship & Support

- If you find this project useful and want to support its continued development, please consider supporting me through:

  | Ko-fi *(outside mainland China)*                                                                               |    | WeChat Donation *(mainland China)*                        |
  |--------------------------------------------------------------------------------------------------|----|------------------------------------------------|
  | [<img src="/docs/images/kofi.png" alt="BuyMeACoffee" height="150">](https://ko-fi.com/haierkeys) | or | <img src="/docs/images/wxds.png" height="150"> |

  - Supporter list:
    - <a href="https://github.com/haierkeys/fast-note-sync-service/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md</a>
    - <a href="https://cnb.cool/haierkeys/fast-note-sync-service/-/blob/master/docs/Support.zh-CN.md">Support.zh-CN.md (cnb.cool mirror)</a>

## ⏱️ Changelog

- ♨️ [View Changelog](/docs/CHANGELOG.md)

## 🗺️ Roadmap


- [ ] Add WebSocket `Protobuf` transport format support to enhance sync efficiency.
- [ ] Isolate and optimize the existing authorization mechanism to improve overall security.
- [ ] Add real-time note updates in WebGui.
- [ ] Add client-to-client peer-to-peer messaging (non-note & attachment, similar to localsend, without client-side save, server-side save supported).
- [ ] Improve various help documentation.
- [ ] Support more intranet penetration (relay gateway) methods.
- [ ] Quick deployment plan:
  * Only requires a public server address, username, and password to complete FNS server deployment.
- [ ] Optimize the current offline note merging solution and add conflict resolution mechanisms.

We are continuously improving. Here are our future development plans:

> **If you have suggestions for improvement or new ideas, feel free to share them by submitting an issue — we will carefully evaluate and adopt suitable suggestions.**

## 🚀 Quick Deployment

We provide multiple installation methods. **One-click script** or **Docker** is recommended.

### Method 1: One-Click Script (Recommended)

Automatically detects the system environment and completes installation and service registration.

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh)
```

For users in mainland China, you can use the Tencent `cnb.cool` mirror:
```bash
bash <(curl -fsSL https://cnb.cool/haierkeys/fast-note-sync-service/-/git/raw/master/scripts/quest_install.sh) --cnb
```


**Script main behaviors:**

  * Automatically downloads the Release binary for the current system.
  * Installs to `/opt/fast-note` by default and creates a global shortcut command `fns` at `/usr/local/bin/fns`.
  * Configures and starts a Systemd (Linux) or Launchd (macOS) service for auto-start on boot.
  * **Management commands**: `fns [install|uninstall|start|stop|status|update|menu]`
  * **Interactive menu**: Run `fns` directly to enter the interactive menu, which supports install/upgrade, service control, auto-start configuration, and switching between GitHub / CNB mirrors.

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
      - ./config:/fast-note-sync/config    # Config files
```

Start the service:

```bash
docker compose up -d
```

-----

### Method 3: Manual Binary Installation

Download the latest release for your OS from [Releases](https://github.com/haierkeys/fast-note-sync-service/releases), extract, and run:

```bash
./fast-note-sync-service run -c config/config.yaml
```

## 📖 Usage Guide

1.  **Access the management panel**:
    Open `http://{ServerIP}:9000` in your browser.
2.  **Initial setup**:
    Register an account on first access. *(To disable registration, set `user.register-is-enable: false` in the config file)*
3.  **Configure the client**:
    Log in to the management panel and click **"Copy API Config"**.
4.  **Connect Obsidian**:
    Open the Obsidian plugin settings page and paste the copied config.


## ⚙️ Configuration

The default config file is `config.yaml`, which the program will automatically look for in the **root directory** or the **config/** directory.

View the full configuration example: [config/config.yaml](https://github.com/haierkeys/fast-note-sync-service/blob/master/config/config.yaml)

## 🌐 Nginx Reverse Proxy Configuration Example

View the full configuration example: [https-nginx-example.conf](https://github.com/haierkeys/fast-note-sync-service/blob/master/scripts/https-nginx-example.conf)

## 🧰 MCP (Model Context Protocol) Support

FNS now natively supports **MCP (Model Context Protocol)** with both **SSE** and **StreamableHTTP** transport modes.

You can connect FNS as an MCP server directly to compatible AI clients such as Cherry Studio, Cursor, Claude Code, hermes-agent, etc. Once connected, the AI can read and write private notes and attachments. All MCP-generated changes are synced in real time to all your devices via WebSocket.

### Common Request Headers

The following headers are supported regardless of transport mode:

- **Auth Header**: `Authorization: Bearer <your API Token>` (obtained from the Copy API Config in WebGUI)
- **Optional Header**: `X-Default-Vault-Name: <vault name>` (specifies the default vault for MCP operations; used if the `vault` parameter is not specified in a tool call)
- **Optional Header**: `X-Client: <client type>` (client type connecting via MCP, e.g.: Cherry Studio / OpenClaw)
- **Optional Header**: `X-Client-Version: <client version>` (client version connecting via MCP, e.g.: 1.1)
- **Optional Header**: `X-Client-Name: <client name>` (client name connecting via MCP, e.g.: Mac)

---

### Integration: StreamableHTTP Mode (Recommended)

StreamableHTTP is the standard transport protocol in the MCP ecosystem. It uses a single endpoint for all requests, is more firewall-friendly, and is natively supported by newer MCP clients (e.g., Claude Code, hermes-agent).

- **Endpoint**: `http://<your server IP or domain>:<port>/api/mcp`
- **Methods**: `POST` (send request/notification), `GET` (listen for server-sent events), `DELETE` (terminate session)

#### Example: Claude Code / hermes-agent / Cursor, etc.

*(Note: Replace `<ServerIP>`, `<Port>`, `<Token>`, and `<VaultName>` with your actual values)*

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

### Integration: SSE Mode (Backward Compatible)

SSE mode is the legacy transport protocol, fully retained for backward compatibility. Suitable for MCP clients that only support SSE (e.g., Cherry Studio).

- **Endpoint**: `http://<your server IP or domain>:<port>/api/mcp/sse`

#### Example: Cherry Studio / Cline, etc.

*(Note: Replace `<ServerIP>`, `<Port>`, `<Token>`, and `<VaultName>` with your actual values)*

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
  * [Obsidian Fast Note Sync Plugin](https://github.com/haierkeys/obsidian-fast-note-sync) / [cnb.cool mirror](https://cnb.cool/haierkeys/obsidian-fast-note-sync)
* Third-party Clients
  * [FastNodeSync-CLI](https://github.com/Go1c/FastNodeSync-CLI) — A Python-based command-line client implementing bidirectional real-time sync via the FNS WebSocket API. Designed for headless Linux server environments (e.g., OpenClaw), delivering sync capabilities equivalent to Obsidian desktop/mobile clients.