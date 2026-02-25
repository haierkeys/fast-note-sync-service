# WebSocket 同步协议前端对接说明 (版本 1.1)

本协议描述了最新调整后的同步流程，前端在对接 `NoteSync`, `FolderSync`, `SettingSync` 和 `FileSync` 时需遵循以下规范。

## 1. 核心变更概览

- **Request**: 所有同步请求需携带 `context` 字符串。
- **Response**: 同步结果不再合并返回，改为 **先返回统计结束消息 (End)，后发送逐条详情消息**。
- **Context**: 所有下发的同步响应都将原样透传请求中的 `context`。

## 2. 交互流程示例

以 **笔记同步 (NoteSync)** 为例：

### Step 1: 前端发起同步请求
前端需生成一个唯一的 `context`（如随机 UUID 或时间戳），用于标识本次同步任务。

**Action**: `NoteSync`
**Data**:
```json
{
  "context": "sync_task_001",
  "vault": "MyNotes",
  "lastTime": 1708800000000,
  "notes": [...]
}
```

### Step 2: 服务端返回统计消息 (End)
服务端在扫描完变更后，会立刻发送一个 End 确认消息。该消息**不再包含明细列表**，仅用于告知统计数据。

**ActionType**: `NoteSyncEnd`
**Response**:
```json
{
  "code": 200,
  "status": true,
  "message": "success",
  "vault": "MyNotes",
  "context": "sync_task_001",
  "data": {
    "lastTime": 1708900000000,
    "needUploadCount": 2,
    "needModifyCount": 1,
    "needSyncMtimeCount": 0,
    "needDeleteCount": 1
  }
}
```

### Step 3: 服务端逐条推送明细消息
随后，服务端会将具体的变更动作通过独立的 WebSocket 消息下发。

- **明细消息 1 (修改笔记)**
  **ActionType**: `NoteSyncModify`
  **Response**: `{ "context": "sync_task_001", "data": { "path": "test.md", "content": "..." }, ... }`

- **明细消息 2 (删除笔记)**
  **ActionType**: `NoteSyncDelete`
  **Response**: `{ "context": "sync_task_001", "data": { "path": "old.md" }, ... }`

## 3. 响应消息结构 (Res)

所有 WebSocket 响应均遵循以下标准结构：

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `code` | int | 业务状态码 (200 为成功) |
| `status` | bool | 成功状态 |
| `message` | string | 状态描述 |
| `data` | any | 业务数据载体 |
| `vault` | string | 保险库名称 (透传) |
| `context` | string | **任务上下文标识 (透传)** |

## 4. 前端集成建议

1. **并行处理**: 由于统计消息 (End) 提前到达，前端可以先更新同步进度 UI，随后监听后续的明细推送并动态更新本地缓存。
2. **任务匹配**: 在 WebSocket 的全局消息监听器中，建议通过响应体中的 `context` 字段来匹配本次同步请求的回调逻辑或状态。
3. **计数校验**: 前端可以通过 `SyncEnd` 消息中的 `needXXXCount` 来验证后续是否收到了足额的详情推送。

## 5. 受影响的接口 Action

| 模块 | 同步请求 Action | 统计结束消息 Type | 明细推送消息 Type |
| :--- | :--- | :--- | :--- |
| **笔记** | `NoteSync` | `NoteSyncEnd` | `NoteSyncModify`, `NoteSyncDelete`, `NoteSyncMtime`, `NoteSyncNeedPush` |
| **文件夹** | `FolderSync` | `FolderSyncEnd` | `FolderSyncModify`, `FolderSyncDelete` |
| **设置** | `SettingSync` | `SettingSyncEnd` | `SettingSyncModify`, `SettingSyncDelete`, `SettingSyncMtime`, `SettingSyncNeedUpload` |
| **文件/附件** | `FileSync` | `FileSyncEnd` | `FileSyncUpdate`, `FileSyncDelete`, `FileSyncMtime`, `FileUpload` |

---
*注：请确保前端代码能够兼容处理同一 `context` 下接连收到的多条消息。*
