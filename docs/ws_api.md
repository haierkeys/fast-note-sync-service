# WebSocket API 全量对接文档 (100% 完整版)

本手册为前端开发人员提供服务端 WebSocket 接口的**完全定义**。涵盖所有模块（笔记、文件夹、文件、设置）的请求、响应、推送消息及详细字段结构。

---

## 1. 协议规范

### 1.1 连接

- **Endpoint**: `GET /api/user/sync`
- **协议升级**: 标准 WebSocket (RFC 6455)。

### 1.2 消息封装格式

WebSocket 文本帧统一使用 `Action|JSON` 字符串格式。

- **示例**: `Authorization|"token_string_here"`

### 1.3 统一响应外壳 (`Res`)

服务端发回的 JSON 消息体（管道符 `|` 之后的部分）结构如下：

| JSON Key  | 类型       | 说明                                                          |
|:----------|:-----------|:--------------------------------------------------------------|
| `code`    | int        | 业务状态码 (1: 成功, 6: 无需同步, 441: 冲突, 305: 参数错误等) |
| `status`  | bool       | 逻辑成功标识                                                  |
| `message` | string     | 错误或提示信息 (i18n)                                         |
| `data`    | object/any | 具体的业务负载内容                                            |
| `details` | string     | 详细错误信息 (omitempty)                                      |
| `vault`   | string     | 所属仓库名称 (omitempty)                                      |
| `context` | string     | 上下文标识 (omitempty)                                        |

---

## 2. 基础控制消息

### 2.1 鉴权 (Authorization)

- **流向**: 客户端 -> 服务端
- **Action**: `Authorization`
- **内容 (Data)**: Token 字符串文本。
- **响应 (Data)**:
  - `version`: string (服务端版本号)
  - `gitTag`: string (Git Tag 信息)
  - `buildTime`: string (编译时间)

### 2.2 客户端信息声明 (ClientInfo)

- **流向**: 客户端 -> 服务端
- **Action**: `ClientInfo`
- **请求内容 (Data)**:

| 字段                  | 类型   | 必填 | 说明                                     |
|:----------------------|:-------|:-----|:-----------------------------------------|
| `name`                | string | 是   | 客户端名称 (设备名)                      |
| `version`             | string | 是   | 客户端当前版本                           |
| `type`                | string | 是   | 客户端类型 (如 `obsidianPlugin`)         |
| `offlineSyncStrategy` | string | 否   | 策略: `newTimeMerge` / `ignoreTimeMerge` |

- **响应 (Data)**:

| 字段                   | 类型   | 说明                 |
|:-----------------------|:-------|:---------------------|
| `versionIsNew`         | bool   | 服务端版本是否有更新 |
| `versionNewName`       | string | 服务端新版本号       |
| `versionNewLink`       | string | 服务端新版本下载链接 |
| `pluginVersionIsNew`   | bool   | 插件版本是否有更新   |
| `pluginVersionNewName` | string | 插件新版本号         |
| `pluginVersionNewLink` | string | 插件新版本下载链接   |

---

## 3. 笔记模块 (Notes)

### 3.1 核心动作对照表

| 流向   | Action             | 说明               | 数据结构 (Data)             |
|:-------|:-------------------|:-------------------|:----------------------------|
| C -> S | `NoteSync`         | 请求笔记增量同步   | `NoteSyncRequest`           |
| C -> S | `NoteModify`       | 提交修改/新建      | `NoteModifyOrCreateRequest` |
| C -> S | `NoteDelete`       | 删除笔记           | `NoteDeleteRequest`         |
| C -> S | `NoteRename`       | 重命名笔记         | `NoteRenameRequest`         |
| C -> S | `NoteCheck`        | 检查笔记更新必要性 | `NoteUpdateCheckRequest`    |
| C -> S | `NoteRePush`       | 请求重推某笔记     | `NoteGetRequest`            |
| S -> C | `NoteSyncModify`   | 推送/同步笔记详情  | `NoteSyncModifyMessage`     |
| S -> C | `NoteSyncDelete`   | 指示客户端删除     | `NoteSyncDeleteMessage`     |
| S -> C | `NoteSyncRename`   | 指示客户端重命名   | `NoteSyncRenameMessage`     |
| S -> C | `NoteSyncMtime`    | 仅同步修改时间     | `NoteSyncMtimeMessage`      |
| S -> C | `NoteSyncNeedPush` | 要求客户端上传本地 | `NoteSyncNeedPushMessage`   |
| S -> C | `NoteSyncEnd`      | 完成同步响应       | `NoteSyncEndMessage`        |

### 3.2 详细 DTO 定义

#### `NoteSyncEndMessage`

| 字段                 | 类型  | 说明                                                |
|:---------------------|:------|:----------------------------------------------------|
| `lastTime`           | int64 | **[关键]** 本次同步后的最新时间戳 (毫秒)，下传传此值 |
| `needUploadCount`    | int64 | 需要客户端上传的笔记总数                            |
| `needModifyCount`    | int64 | 服务端下发修改的笔记总数                            |
| `needSyncMtimeCount` | int64 | 仅同步时间的笔记总数                                |
| `needDeleteCount`    | int64 | 指示删除的笔记总数                                  |
| `messages`           | array | 变更消息队列，详见第 7 章节                          |

#### `NoteModifyOrCreateRequest`

| 字段          | 类型   | 说明                              |
|:--------------|:-------|:----------------------------------|
| `vault`       | string | **[必填]** 仓库名                 |
| `path`        | string | **[必填]** 笔记完整路径           |
| `pathHash`    | string | 路径哈希                          |
| `content`     | string | 笔记文本全文                      |
| `contentHash` | string | 内容哈希                          |
| `baseHash`    | string | 修改前的基础哈希 (用于冲突合并)   |
| `ctime`       | int64  | 创建时间 (秒)                     |
| `mtime`       | int64  | 修改时间 (秒)                     |
| `createOnly`  | bool   | 设置为 true 时，若笔记已存在则报错 |

---

## 4. 文件夹模块 (Folders)

### 4.1 核心动作

- `FolderSync` (C->S): `{ "vault": string, "lastTime": int64, "folders": Array<FolderSyncCheckRequest> }`
- `FolderModify` (C->S): `{ "vault": string, "path": string }`
- `FolderDelete` (C->S): `{ "vault": string, "path": string, "pathHash": string }`
- `FolderRename` (C->S): `{ "vault": string, "oldPath": string, "path": string, ... }`
- `FolderSyncEnd` (S->C): `{ "lastTime": int64, "needModifyCount": int, "needDeleteCount": int, "messages": [] }`

---

## 5. 文件模块 (Files)

### 5.1 二进制分片上传逻辑 (BC Frame)

1. 客户端发送 `FileUploadCheck` (JSON)。
2. 服务端响应 `FileUpload` (JSON)，返回 `sessionId` 和 `chunkSize`。
3. 客户端循环发送 **二进制帧**。帧前缀固定为 `BC` (ASCII 0x42 0x43)。
   - **帧格式 (Binary)**: `[36字节 SessionID][4字节 uint32 大端序 ChunkIndex][原始分片数据]`

### 5.2 文件同步动作汇总

- `FileSyncUpdate` (推送): `{ "path", "pathHash", "contentHash", "size", "ctime", "mtime", "lastTime" }`
- `FileSyncEnd` (推送): `{ "lastTime", "needUploadCount", "needModifyCount", "needSyncMtimeCount", "needDeleteCount", "messages" }`

---

## 6. 设置模块 (Settings)

- `SettingSync`: `{ "vault": string, "settings": Array, "cover": bool }`
- `SettingSyncEnd`: `{ "lastTime", "needUploadCount", "needModifyCount", "needSyncMtimeCount", "needDeleteCount", "messages" }`

---

## 7. 队列消息结构 (`WSQueuedMessage`)

在所有 `*SyncEnd` 消息的 `messages` 数组中，每个项的结构为：

```json
{
  "action": "NoteSyncModify", // 具体的推送 Action 名
  "data": { ... }             // 该 Action 对应的具体 Data 结构
}
```

---

## 8. WebSocket 专属状态码汇总

| Code  | 说明                                                           |
|:------|:---------------------------------------------------------------|
| `1`   | 成功 (Success)                                                 |
| `6`   | 服务端数据与客户端一致，无需更新 (SuccessNoUpdate)              |
| `305` | 客户端提交的 JSON 参数不符合 binding 校验 (ErrorInvalidParams) |
| `433` | 笔记保存入库失败 (ErrorNoteModifyOrCreateFailed)               |
| `441` | 发生内容冲突，且无法自动合并 (ErrorNoteConflict)                |
| `463` | 分片上传 Session 已过期或无效 (ErrorFileUploadSessionNotFound) |
| `490` | 同步逻辑冲突 (ErrorSyncConflict)                               |

---

*注：文档中所有 `int64` 时间戳除 `lastTime`(毫秒) 外均默认为 **秒**。*
