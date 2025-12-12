# WebSocket Note API Documentation

本文档整理了 `fast-note-sync-service` 中涉及笔记的 WebSocket 消息接口，供前端对接使用。

## 核心通信协议 (Protocol)

服务端与客户端的通信采用 **自定义文本协议**，格式如下：

`Action|JSON_Payload`

- **Action**: 消息路由键/类型 (String)
- **|**: 分隔符 (Pipe)
- **JSON_Payload**: 实际数据的 JSON 字符串

**示例**:
```text
NoteModify|{"vault":"Default","path":"notes/test.md","content":"# Hello",...}
```

### JSON Payload 响应结构
服务端返回的 JSON 部分通常遵循以下统一结构：
```typescript
interface Response<T> {
  code: number;   // 状态码 (0 为成功)
  status: string; // 状态描述 (e.g. "success")
  msg: string;    // 提示信息
  data: T;        // 具体的业务数据
}
```

---

## 服务端推送消息 (Server Push Messages)

服务端会推送以下 Format 的消息：`Action|Response_JSON`

### 1. `NoteSyncModify`
**触发场景**: 笔记被创建或修改后，通知客户端更新笔记内容。
**Response Data Structure** (`Response.data`)：

```typescript
interface NoteSyncModifyData {
  path: string;         // 笔记路径
  pathHash: string;     // 路径哈希值
  content: string;      // 笔记完整内容
  contentHash: string;  // 内容哈希值
  ctime: number;        // 创建时间戳 (毫秒)
  mtime: number;        // 修改时间戳 (毫秒)
  lastTime: number;     // 记录更新时间戳 (毫秒)
}
```

**完整消息示例**:
```text
NoteSyncModify|{"code":0,"status":"success","msg":"","data":{"path":"notes/test.md","pathHash":"abc123","content":"# Hello World","contentHash":"def456","ctime":1702345678000,"mtime":1702345678000,"lastTime":1702345678000}}
```

### 2. `NoteSyncMtime`
**触发场景**: 笔记内容一致但修改时间不一致，仅需更新元数据。
**Response Data Structure**:

```typescript
interface NoteSyncMtimeData {
  path: string;   // 笔记路径
  ctime: number;  // 创建时间戳 (毫秒)
  mtime: number;  // 修改时间戳 (毫秒)
}
```

### 3. `NoteSyncNeedPush`
**触发场景**: 同步时发现客户端笔记需要上传（客户端笔记比服务端新，或服务端没有该笔记）。
**Response Data Structure**:

```typescript
interface NoteSyncNeedPushData {
  path: string; // 笔记路径
}
```

### 4. `NoteSyncDelete`
**触发场景**: 通知客户端删除笔记。
**Response Data Structure**:

```typescript
interface NoteSyncDeleteData {
  path: string; // 要删除的笔记路径
}
```

### 5. `NoteSyncEnd`
**触发场景**: 笔记同步检查结束。
**Response Data Structure**:

```typescript
interface NoteSyncEndData {
  vault: string;    // 仓库名称
  lastTime: number; // 服务端最新时间戳 (毫秒，用于下次增量同步)
}
```

---

## 客户端请求 (Client Requests)

客户端发送消息格式必须为 `Action|JSON`。

### 1. 修改或创建笔记 (`NoteModify`)
**用途**: 创建新笔记或修改现有笔记。
**Format**: `NoteModify|{...}`

**JSON Data**:
```typescript
{
  vault: string;        // 仓库名称
  path: string;         // 笔记路径
  pathHash: string;     // 路径哈希值 (必填)
  content: string;      // 笔记完整内容
  contentHash: string;  // 内容哈希值 (必填)
  ctime: number;        // 创建时间戳 (毫秒，必填)
  mtime: number;        // 修改时间戳 (毫秒，必填)
}
```

**可能的响应**:
- 成功: 无特定 Action 响应，但会广播 `NoteSyncModify` 给所有客户端
- `NoteSyncMtime`: 仅需更新修改时间（内容一致但时间不同）
- 无响应: 内容和时间都一致，无需更新

**示例**:
```text
NoteModify|{"vault":"Default","path":"notes/test.md","pathHash":"abc123","content":"# Hello World","contentHash":"def456","ctime":1702345678000,"mtime":1702345678000}
```

### 2. 检查笔记修改 (`NoteModifyCheck`)
**用途**: 仅检查笔记是否需要修改，不实际修改内容。用于优化上传流程。
**Format**: `NoteModifyCheck|{...}`

**JSON Data**:
```typescript
{
  vault: string;        // 仓库名称
  path: string;         // 笔记路径
  pathHash: string;     // 路径哈希值
  contentHash: string;  // 内容哈希值
  ctime: number;        // 创建时间戳 (毫秒)
  mtime: number;        // 修改时间戳 (毫秒)
}
```

**可能的响应**:
- `NoteSyncNeedPush`: 需要上传笔记内容
- `NoteSyncMtime`: 仅需更新修改时间
- 无响应: 无需更新

### 3. 删除笔记 (`NoteDelete`)
**用途**: 删除指定笔记。
**Format**: `NoteDelete|{...}`

**JSON Data**:
```typescript
{
  vault: string;   // 仓库名称
  path: string;    // 要删除的笔记路径
  pathHash: string; // 路径哈希值
}
```

**响应**: 成功后会广播 `NoteSyncDelete` 给所有客户端。

**示例**:
```text
NoteDelete|{"vault":"Default","path":"notes/test.md","pathHash":"abc123"}
```

### 4. 批量笔记同步检查 (`NoteSync`)
**用途**: 批量检查笔记更新，用于增量同步。
**Format**: `NoteSync|{...}`

**JSON Data**:
```typescript
{
  vault: string;      // 仓库名称
  lastTime: number;   // 上次同步的时间戳 (毫秒，首次同步传 0)
  notes: Array<{      // 客户端当前笔记列表
    path: string;         // 笔记路径
    pathHash: string;     // 路径哈希值
    contentHash: string;  // 内容哈希值
    mtime: number;        // 修改时间戳 (毫秒)
  }>;
}
```

**同步逻辑**:
服务端会比较客户端笔记列表与服务端笔记，并返回以下消息组合：

- `NoteSyncModify`: 客户端需要更新的笔记（包含完整内容）
- `NoteSyncNeedPush`: 客户端需要上传的笔记
- `NoteSyncMtime`: 仅需更新修改时间的笔记
- `NoteSyncDelete`: 客户端需要删除的笔记
- `NoteSyncEnd`: 同步结束 (必定最后发送)

**示例**:
```text
NoteSync|{"vault":"Default","lastTime":1702345678000,"notes":[{"path":"notes/test.md","pathHash":"abc123","contentHash":"def456","mtime":1702345678000}]}
```

---

## 典型使用场景

### 场景 1: 用户创建或修改笔记
```
客户端: NoteModify|{"vault":"Default","path":"notes/test.md","content":"# Hello",...}
服务端: (响应成功)
服务端: (广播给所有客户端) NoteSyncModify|{"path":"notes/test.md","content":"# Hello",...}
```

### 场景 2: 用户删除笔记
```
客户端: NoteDelete|{"vault":"Default","path":"notes/test.md",...}
服务端: (响应成功)
服务端: (广播给所有客户端) NoteSyncDelete|{"path":"notes/test.md"}
```

### 场景 3: 批量同步笔记
```
客户端: NoteSync|{"vault":"Default","lastTime":1702345678000,"notes":[...]}
服务端: NoteSyncModify|{...}      (需要更新的笔记，包含完整内容)
服务端: NoteSyncNeedPush|{...}    (需要上传的笔记)
服务端: NoteSyncMtime|{...}       (仅需更新时间的笔记)
服务端: NoteSyncDelete|{...}      (需要删除的笔记)
服务端: NoteSyncEnd|{"lastTime":1702345999000}  (同步结束)
```

### 场景 4: 优化上传流程（先检查再上传）
```
客户端: NoteModifyCheck|{"vault":"Default","path":"notes/test.md",...}
服务端: NoteSyncNeedPush|{"path":"notes/test.md"}
客户端: NoteModify|{"vault":"Default","path":"notes/test.md","content":"...",...}
服务端: (广播) NoteSyncModify|{...}
```

---

## 同步逻辑说明

### NoteSync 详细流程

当客户端发送 `NoteSync` 请求时，服务端会执行以下比对逻辑：

1. **服务端已删除的笔记**
   - 如果客户端有该笔记 → 发送 `NoteSyncDelete`

2. **服务端存在的笔记**
   - **客户端也有该笔记**:
     - 内容和时间都一致 → 跳过
     - 内容不一致:
       - 服务端 mtime 更新 → 发送 `NoteSyncModify` (包含完整内容)
       - 客户端 mtime 更新 → 发送 `NoteSyncNeedPush` (要求客户端上传)
     - 内容一致但时间不一致 → 发送 `NoteSyncMtime`
   - **客户端没有该笔记** → 发送 `NoteSyncModify` (包含完整内容)

3. **客户端独有的笔记**
   - 服务端没有 → 发送 `NoteSyncNeedPush` (要求客户端上传)

---

## 注意事项

1. **时间戳单位**: 所有时间戳字段 (`ctime`, `mtime`, `lastTime`) 均为**毫秒**。

2. **哈希值**: `pathHash` 和 `contentHash` 用于快速比对，建议使用 MD5 或 SHA256。

3. **内容传输**: 笔记内容直接在 JSON 中传输，无需分块上传（与文件不同）。

4. **广播机制**: 笔记修改、删除操作会广播给同一用户的所有在线客户端，实现多端同步。

5. **增量同步**: 使用 `lastTime` 实现增量同步，只传输自上次同步后发生变化的笔记。

6. **冲突处理**: 当内容不一致时，通过比较 `mtime` 决定以哪一端为准：
   - 服务端 mtime 更新 → 客户端下载
   - 客户端 mtime 更新 → 客户端上传

7. **优化建议**:
   - 对于大量笔记，建议先使用 `NoteModifyCheck` 检查，只上传需要更新的笔记
   - 使用 `NoteSync` 进行批量同步，减少网络请求次数
