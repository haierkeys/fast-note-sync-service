---
name: fns-mcp
description: Fast Note Sync Service MCP SSE Skill (Bilingual). Allows agents to access and manage notes, files, and vaults via the MCP protocol.
version: 1.0.0
---

# Fast Note Sync (FNS) MCP Skill (Bilingual/双语版)

This skill allows agents to interact with the Fast Note Sync Service using the Model Context Protocol (MCP).
本技能允许 Agent 通过 Model Context Protocol (MCP) 与 Fast Note Sync Service 交互。

## Core Capabilities / 核心能力

- **Note Management / 笔记管理**: Listing, searching, CRUD, moving, renaming, restoring, and clearing recycle bin. (列表、检索、增删改查、移动、重命名、恢复、清理回收站)
- **Note Enhancement / 笔记增强**: Frontmatter patching, appending/prepending, find & replace, and backlinks/outlinks. (Frontmatter 修补、追加/预置、查找替换、双链/外链)
- **File Management / 文件管理**: Attachment listing, metadata retrieval, reading content (Base64), renaming, and deletion. (附件列表、元数据、读取内容、重命名、删除)
- **Vault Management / 库管理**: Vault listing, creating/updating, and deletion. (库列表、创建/更新、删除)

## Configuration Guide / 配置指南

Before using this skill, the agent needs to connect to the MCP SSE interface.
在使用此技能前，Agent 需要连接到 MCP SSE 接口。

### Interface Details / 接口信息
- **SSE Endpoint**: `http://<YOUR_DOMAIN>:9000/api/mcp/sse`
- **Message Endpoint**: `http://<YOUR_DOMAIN>:9000/api/mcp/message`
- **Authentication**: Requires `Authorization: Bearer <YOUR_AUTH_TOKEN>` in headers.
- **Client Identity (Optional)**: 
    - `X-Client`: `<ClientType>` (e.g., `CherryStudio`, `OpenClaw`)
    - `X-Client-Name`: `<ClientName>` (e.g., `MyAgent`)
    - `X-Client-Version`: `<Version>`
- **Default Vault (Optional)**: `X-Default-Vault-Name: <VAULT_NAME>`

### Platform Specific Config / 平台配置示例
- [OpenClaw Configuration](configs/openclaw.json)
- [Hermes Agent Configuration](configs/hermes.yaml)
- [Cherry Studio Configuration Guide / 配置指南](configs/cherry-studio.md)

---

## Available Tools / 可用工具说明

### 1. Note Tools / 笔记工具

| Tool Name / 工具 | Description / 描述 | Arguments / 参数 |
| :--- | :--- | :--- |
| `note_list` | List notes in a vault / 列出笔记 | `vault`, `keyword` |
| `note_get` | Get note content / 获取笔记内容 | `vault`, `path` |
| `note_create_or_update` | Create/Update note / 创建或更新笔记 | `vault`, `path`, `content` |
| `note_delete` | Delete note / 删除笔记 | `vault`, `path` |
| `note_rename` | Rename note / 重命名笔记 | `vault`, `oldPath`, `newPath` |
| `note_restore` | Restore note / 恢复笔记 | `vault`, `path` |
| `note_append` | Append content / 追加内容 | `vault`, `path`, `content` |
| `note_prepend` | Prepend content / 预置内容 | `vault`, `path`, `content` |
| `note_replace` | Find & Replace / 查找替换 | `vault`, `path`, `find`, `replace`, `regex`, `all` |
| `note_patch_frontmatter` | Patch Frontmatter / 修补前置参数 | `vault`, `path`, `updates` (JSON), `remove` (JSON) |
| `note_get_backlinks` | Get backlinks / 获取反链 | `vault`, `path` |
| `note_get_outlinks` | Get outlinks / 获取出链 | `vault`, `path` |

### 2. File Tools / 文件工具

| Tool Name / 工具 | Description / 描述 | Arguments / 参数 |
| :--- | :--- | :--- |
| `file_list` | List files / 列出文件 | `vault`, `keyword` |
| `file_get_info` | Get metadata / 获取元数据 | `vault`, `path` |
| `file_read` | Read content (Base64) / 读取内容 | `vault`, `path` |
| `file_delete` | Delete file / 删除文件 | `vault`, `path` |
| `file_rename` | Rename file / 重命名文件 | `vault`, `oldPath`, `newPath` |

### 3. Vault Tools / 库工具

| Tool Name / 工具 | Description / 描述 | Arguments / 参数 |
| :--- | :--- | :--- |
| `vault_list` | List vaults / 列出库 | None |
| `vault_get` | Get vault details / 获取详情 | `id` |
| `vault_create_or_update` | Create/Update vault / 创建或更新 | `vault`, `id` |
| `vault_delete` | Delete vault / 删除库 | `id` |

---

## Best Practices / 最佳实践

1. **Vault Selection**: Explicitly specify `vault` when possible. (尽可能明确指定 `vault` 参数)
2. **Path Handling**: Paths are relative to vault root. (路径相对于库根目录)
3. **Encoding**: `file_read` returns Base64. (文件读取返回 Base64 编码)
4. **Consistency**: Be aware of path changes after rename/move. (重命名后注意路径变更)
