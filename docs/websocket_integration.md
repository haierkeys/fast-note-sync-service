# WebSocket 同步协议更新说明 (2026-03-05)

本文档详细说明了近期 WebSocket 协议的变更，主要涉及笔记、附件和配置同步消息中 `lastTime` 字段的补充。

## 1. 核心变更：引入 `lastTime`

为了增强前端增量同步的可靠性，我们在多个资源变更消息中补充了 `lastTime` 字段。

- **字段名**: `lastTime` (对应后端 `UpdatedTimestamp`)
- **数据类型**: `int64` (毫秒级时间戳)
- **物理意义**: 该资源记录在数据库中的最后更新时间。前端在进行增量同步请求时，应记录此值，并作为下次同步请求的起点。

---

## 2. 字段更新详情

### 2.1 笔记消息 (Note)

| 消息 Action (Action) | 新增字段 | 说明 |
| :--- | :--- | :--- |
| `NoteSyncRename` | `lastTime` | 笔记重命名后的同步消息 |
| `NoteSyncMtime` | `lastTime` | 笔记修改时间变更（无需下载内容时） |
| `NoteSyncDelete` | `lastTime` | 笔记已删除的同步消息 |
| `NoteSyncModify` | `lastTime` | (固有) 笔记创建或更新消息 |

### 2.2 附件/文件消息 (File)

| 消息 Action (Action) | 新增字段 | 说明 |
| :--- | :--- | :--- |
| `FileSyncRename` | `lastTime` | 附件重命名后的同步消息 |
| `FileSyncMtime` | `lastTime` | 附件修改时间变更 |
| `FileSyncDelete` | `lastTime` | 附件已删除的同步消息 |
| `FileSyncUpdate` | `lastTime` | (固有) 附件创建或更新消息 |

### 2.3 配置消息 (Setting)

| 消息 Action (Action) | 新增字段 | 说明 |
| :--- | :--- | :--- |
| `SettingSyncMtime` | `lastTime` | 配置修改时间变更 |
| `SettingSyncDelete` | `lastTime`, `pathHash`, `ctime`, `mtime` | 配置已删除同步（结构大幅增强以保持一致性） |
| `SettingSyncModify` | `lastTime` | (固有) 配置创建或更新消息 |

---

## 3. 详细结构定义 (JSON 示例)

### 配置删除消息示例 (结构增强)
```json
{
    "action": "SettingSyncDelete",
    "data": {
        "path": "User/Theme",
        "pathHash": "shash789",
        "ctime": 1700000000,
        "mtime": 1700000000,
        "lastTime": 1700000001
    }
}
```

### 笔记重命名消息示例
```json
{
    "action": "NoteSyncRename",
    "data": {
        "path": "NewName.md",
        "pathHash": "nfhash123",
        "oldPath": "OldName.md",
        "oldPathHash": "ofhash456",
        "lastTime": 1700001000,
        ...
    }
}
```

---

## 4. 前端对接建议

1. **状态更新**: 当收到上述任何带有 `lastTime` 字段的消息时，前端应更新本地缓存中对应资源的 `lastTime` 属性。
2. **同步基准**: 在调用 `NoteSync`, `FileSync`, `SettingSync` 时，参数中的 `lastTime` 应取本地所有资源（包含已逻辑删除的资源）中最大的那个 `lastTime` 值，以确保不遗漏任何服务端变更。
3. **删除处理**: `SettingSyncDelete` 现在的字段更丰富（补全了 `pathHash` 等），前端可以更统一地处理各类资源的删除逻辑。
