# Fast Note Sync Service - REST API 文档

本文档基于源码深度分析生成，提供准确的接口定义。

---

## 通用说明

### 基础 URL
```
http://{host}:9000/api
```

### 认证方式
需要认证的接口必须在请求头中携带 Token：
```
Authorization: {token}
```
Token 通过登录接口获取。

### 通用响应结构
```typescript
interface Response<T> {
  code: number;      // 状态码 (0=失败, 1+=成功)
  status: boolean;   // 操作状态
  message: string;   // 提示信息
  data: T;           // 业务数据
  details?: string[]; // 错误详情（可选）
}
```

### 分页响应结构
```typescript
interface ListResponse<T> {
  code: number;
  status: boolean;
  message: string;
  data: {
    list: T[];
    pager: {
      page: number;
      pageSize: number;
      totalRows: number;
    }
  }
}
```

### 分页参数
| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| page | number | 页码 | 1 |
| page_size | number | 每页数量 | 10 (最大100) |

---

## 错误码参考

| Code | 说明 |
|------|------|
| 0 | 失败 |
| 1-6 | 成功状态 |
| 400-446 | 业务错误 |
| 500-534 | 系统/同步错误 |

### 常见错误码
| Code | 中文 | English |
|------|------|---------|
| 405 | 用户注册已关闭 | User registration is closed |
| 407 | 用户不存在 | Username does not exist |
| 408 | 用户已经存在 | Username already exists |
| 414 | 笔记仓库不存在 | Note Vault does not exist |
| 428 | 笔记不存在 | Note does not exist |
| 445 | 此操作需要管理员权限 | This operation requires administrator privileges |
| 505 | 参数验证失败 | Invalid Params |
| 507 | 尚未登录 | Not logged in |
| 508 | 登录状态失效 | Session expired |

---

## 公开接口（无需认证）

### 1. 用户注册
```
POST /api/user/register
```

**请求参数** (JSON/Form):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | ✓ | 邮箱地址 |
| username | string | ✓ | 用户名 (3-15位，字母/数字/下划线) |
| password | string | ✓ | 密码 |
| confirmPassword | string | ✓ | 确认密码 |

**响应 Data**:
```typescript
{
  uid: number;
  email: string;
  username: string;
  token: string;
  avatar: string;
  updatedAt: string;
  createdAt: string;
}
```

---

### 2. 用户登录
```
POST /api/user/login
```

**请求参数** (Form):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| credentials | string | ✓ | 用户名或邮箱 |
| password | string | ✓ | 密码 |

**响应 Data**:
```typescript
{
  uid: number;
  email: string;
  username: string;
  token: string;      // 用于后续认证
  avatar: string;
  updatedAt: string;
  createdAt: string;
}
```

---

### 3. 获取服务端版本
```
GET /api/version
```

**响应 Data**:
```typescript
{
  version: string;    // 版本号
  gitTag: string;     // Git 标签
  buildTime: string;  // 构建时间
}
```

---

### 4. 获取 WebGUI 配置
```
GET /api/webgui/config
```

**响应 Data**:
```typescript
{
  fontSet: string;          // 字体设置 ("local" | "" | URL)
  registerIsEnable: boolean; // 是否开放注册
  adminUid: number;         // 管理员 UID (0=未设置)
}
```

---

### 5. WebSocket 连接
```
GET /api/user/sync
```
升级为 WebSocket 连接，用于实时同步。详见 WebSocket API 文档。

---

## 需认证接口

### 用户相关

#### 6. 获取用户信息
```
GET /api/user/info
```

**响应 Data**:
```typescript
{
  uid: number;
  email: string;
  username: string;
  avatar: string;
  updatedAt: string;
  createdAt: string;
}
```

---

#### 7. 修改密码
```
POST /api/user/change_password
```

**请求参数** (JSON/Form):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| oldPassword | string | ✓ | 当前密码 |
| password | string | ✓ | 新密码 |
| confirmPassword | string | ✓ | 确认新密码 |

**响应**: 成功返回 code=5, message="密码修改成功"

---

### Vault (仓库) 相关

#### 8. 获取仓库列表
```
GET /api/vault
```

**响应 Data**:
```typescript
Array<{
  id: number;
  vault: string;      // 仓库名称
  noteCount: number;  // 笔记数量
  noteSize: number;   // 笔记总大小 (bytes)
  fileCount: number;  // 附件数量
  fileSize: number;   // 附件总大小 (bytes)
  size: number;       // 总大小 (noteSize + fileSize)
}>
```

---

#### 9. 创建/更新仓库
```
POST /api/vault
```

**请求参数** (JSON/Form):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| vault | string | ✓ | 仓库名称 |
| id | number | - | 仓库ID (有值=更新，无值=创建) |

**响应 Data**:
```typescript
{
  id: number;
  vault: string;
  noteCount: number;
  noteSize: number;
  fileCount: number;
  fileSize: number;
  size: number;
}
```

---

#### 10. 删除仓库
```
DELETE /api/vault
```

**请求参数** (Query/Form):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | number | ✓ | 仓库ID (≥1) |

**响应**: 成功返回 code=4, message="删除成功"

---

### Note (笔记) 相关

#### 11. 获取笔记列表
```
GET /api/notes
```

**请求参数** (Query):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| vault | string | ✓ | 仓库名称 |
| keyword | string | - | 搜索关键词 |
| isRecycle | boolean | - | 是否查询回收站 |
| page | number | - | 页码 |
| page_size | number | - | 每页数量 |

**响应 Data** (分页):
```typescript
{
  list: Array<{
    id: number;
    action: string;           // "create" | "modify" | "delete"
    path: string;             // 笔记路径
    pathHash: string;         // 路径哈希
    version: number;          // 版本号
    ctime: number;            // 创建时间戳 (毫秒)
    mtime: number;            // 修改时间戳 (毫秒)
    updatedTimestamp: number; // 更新时间戳 (毫秒)
    updatedAt: string;        // 更新时间
    createdAt: string;        // 创建时间
  }>;
  pager: { page, pageSize, totalRows }
}
```

---

#### 12. 获取单条笔记
```
GET /api/note
```

**请求参数** (Query):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| vault | string | ✓ | 仓库名称 |
| path | string | ✓ | 笔记路径 |
| pathHash | string | - | 路径哈希 (可选，自动计算) |
| isRecycle | boolean | - | 是否查询回收站 |

**响应 Data**:
```typescript
{
  id: number;
  path: string;
  pathHash: string;
  content: string;            // 笔记内容 (Markdown)
  contentHash: string;        // 内容哈希
  fileLinks: Record<string, string>; // 嵌入文件链接映射
  version: number;
  ctime: number;
  mtime: number;
  updatedTimestamp: number;
  updatedAt: string;
  createdAt: string;
}
```

---

#### 13. 创建/更新笔记
```
POST /api/note
```

**请求参数** (JSON/Form):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| vault | string | ✓ | 仓库名称 |
| path | string | ✓ | 笔记路径 |
| content | string | - | 笔记内容 |
| pathHash | string | - | 路径哈希 (自动计算) |
| contentHash | string | - | 内容哈希 (自动计算) |
| srcPath | string | - | 原路径 (重命名时使用) |
| srcPathHash | string | - | 原路径哈希 |
| ctime | number | - | 创建时间戳 (默认当前时间) |
| mtime | number | - | 修改时间戳 (默认当前时间) |

**响应 Data**:
```typescript
{
  id: number;
  path: string;
  pathHash: string;
  content: string;
  contentHash: string;
  version: number;
  ctime: number;
  mtime: number;
  lastTime: number;  // updatedTimestamp
}
```

**副作用**: 成功后会通过 WebSocket 广播 `NoteSyncModify` 消息

---

#### 14. 删除笔记
```
DELETE /api/note
```

**请求参数** (Query/Form):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| vault | string | ✓ | 仓库名称 |
| path | string | ✓ | 笔记路径 |
| pathHash | string | - | 路径哈希 |

**响应 Data**: 返回被删除的笔记信息

**副作用**: 成功后会通过 WebSocket 广播 `NoteSyncDelete` 消息

---

#### 15. 获取笔记/附件原始内容
```
GET /api/note/file
```

**请求参数** (Query):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| vault | string | ✓ | 仓库名称 |
| path | string | ✓ | 文件路径 |
| pathHash | string | - | 路径哈希 |

**响应**: 直接返回文件内容 (非 JSON)
- Content-Type: 根据文件类型自动识别
- Cache-Control: public, max-age=31536000
- ETag: 内容哈希

---

### Note History (笔记历史) 相关

#### 16. 获取笔记历史列表
```
GET /api/note/histories
```

**请求参数** (Query):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| vault | string | ✓ | 仓库名称 |
| path | string | ✓ | 笔记路径 |
| pathHash | string | - | 路径哈希 |
| isRecycle | boolean | - | 是否查询回收站 |
| page | number | - | 页码 |
| page_size | number | - | 每页数量 |

**响应 Data** (分页):
```typescript
{
  list: Array<{
    id: number;
    noteId: number;
    vaultId: number;
    path: string;
    clientName: string;  // 修改来源客户端
    version: number;
    createdAt: string;
  }>;
  pager: { page, pageSize, totalRows }
}
```

---

#### 17. 获取历史详情
```
GET /api/note/history
```

**请求参数** (Query):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | number | ✓ | 历史记录ID |

**响应 Data**:
```typescript
{
  id: number;
  noteId: number;
  vaultId: number;
  path: string;
  diffs: Array<{      // Diff 结果
    Type: number;     // -1=删除, 0=相等, 1=插入
    Text: string;
  }>;
  content: string;    // 该版本的完整内容
  contentHash: string;
  clientName: string;
  version: number;
  createdAt: string;
}
```

---

### Admin (管理员) 相关

#### 18. 获取管理配置
```
GET /api/admin/config
```

**权限**: 需要管理员权限 (当 adminUid ≠ 0 时，仅该 UID 可访问)

**响应 Data**:
```typescript
{
  fontSet: string;
  registerIsEnable: boolean;
  fileChunkSize: string;           // 如 "512KB"
  softDeleteRetentionTime: string; // 如 "7d"
  uploadSessionTimeout: string;    // 如 "1d"
  historyKeepVersions: number;     // 如 100
  adminUid: number;
}
```

---

#### 19. 更新管理配置
```
POST /api/admin/config
```

**权限**: 需要管理员权限

**请求参数** (JSON/Form):
| 参数 | 类型 | 说明 |
|------|------|------|
| fontSet | string | 字体设置 |
| registerIsEnable | boolean | 是否开放注册 |
| fileChunkSize | string | 文件分块大小 |
| softDeleteRetentionTime | string | 软删除保留时长 |
| uploadSessionTimeout | string | 上传会话超时 |
| historyKeepVersions | number | 历史记录保留版本数 |
| adminUid | number | 管理员 UID |

**响应 Data**: 返回更新后的配置

---

## 接口路由汇总

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | /api/user/register | ✗ | 用户注册 |
| POST | /api/user/login | ✗ | 用户登录 |
| GET | /api/user/sync | ✗ | WebSocket 连接 |
| GET | /api/version | ✗ | 获取版本信息 |
| GET | /api/webgui/config | ✗ | 获取 WebGUI 配置 |
| GET | /api/user/info | ✓ | 获取用户信息 |
| POST | /api/user/change_password | ✓ | 修改密码 |
| GET | /api/vault | ✓ | 获取仓库列表 |
| POST | /api/vault | ✓ | 创建/更新仓库 |
| DELETE | /api/vault | ✓ | 删除仓库 |
| GET | /api/notes | ✓ | 获取笔记列表 |
| GET | /api/note | ✓ | 获取单条笔记 |
| POST | /api/note | ✓ | 创建/更新笔记 |
| DELETE | /api/note | ✓ | 删除笔记 |
| GET | /api/note/file | ✓ | 获取文件内容 |
| GET | /api/note/histories | ✓ | 获取历史列表 |
| GET | /api/note/history | ✓ | 获取历史详情 |
| GET | /api/admin/config | ✓ | 获取管理配置 |
| POST | /api/admin/config | ✓ | 更新管理配置 |

---

## 时间戳说明

所有时间戳字段 (`ctime`, `mtime`, `updatedTimestamp`, `lastTime`) 均为 **毫秒级 Unix 时间戳**。

---

## 哈希算法

`pathHash` 和 `contentHash` 使用 32 位哈希算法 (FNV-1a 或类似)，客户端可自动计算或由服务端生成。

---

## 全文搜索 (FTS)

服务端内置基于 SQLite FTS5 的全文搜索引擎，支持对笔记路径和内容进行高效的全文检索。

### 技术实现

- 使用 SQLite FTS5 虚拟表
- 支持 Unicode 字符分词 (`unicode61 remove_diacritics 2`)
- 自动维护索引版本，版本变更时自动重建

### FTS 表结构
```sql
CREATE VIRTUAL TABLE note_fts USING fts5(
    note_id UNINDEXED,  -- 关联的笔记 ID（不索引）
    path,               -- 笔记路径（可搜索）
    content,            -- 笔记内容（可搜索）
    tokenize='unicode61 remove_diacritics 2'
)
```

### 内部接口 (DAO 层)

FTS 功能目前通过 `NoteFTSRepository` 接口提供，包含以下方法：

| 方法 | 说明 |
|------|------|
| `Upsert(noteID, path, content, uid)` | 插入或更新 FTS 索引 |
| `Delete(noteID, uid)` | 删除 FTS 索引 |
| `Search(keyword, vaultID, uid, limit, offset)` | 全文搜索，返回匹配的 note_id 列表 |
| `SearchCount(keyword, vaultID, uid)` | 全文搜索计数 |
| `RebuildIndex(uid)` | 重建用户的全部 FTS 索引 |

### 搜索行为

- 搜索词使用双引号包裹进行精确短语匹配
- 搜索结果按相关性排序 (`ORDER BY rank`)
- 自动过滤已删除的笔记 (`action != 'delete'`)
- 支持按 vault 过滤搜索范围

### 笔记列表搜索增强

现有的 `GET /api/notes` 接口的 `keyword` 参数已支持 FTS 全文搜索：

- 当提供 `keyword` 参数时，会使用 FTS 引擎进行内容搜索
- 搜索范围包括笔记路径和笔记内容
- 返回结果按相关性排序
