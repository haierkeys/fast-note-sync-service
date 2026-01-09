# Architecture Documentation

## Overview

Fast Note Sync Service 采用 Clean Architecture 设计，实现了清晰的分层架构和依赖注入模式。

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Target Architecture                      │
├─────────────────────────────────────────────────────────────┤
│  Transport Layer (routers/)                                  │
│  ├── HTTP Handlers (Gin)                                     │
│  ├── WebSocket Handlers                                      │
│  ├── 请求解析、响应格式化                                    │
│  ├── Trace ID 注入                                           │
│  └── 统一错误转换                                            │
├─────────────────────────────────────────────────────────────┤
│  Application Layer (service/)                                │
│  ├── 业务逻辑编排                                            │
│  ├── 事务管理                                                │
│  └── 依赖接口注入                                            │
├─────────────────────────────────────────────────────────────┤
│  Domain Layer (domain/)                                      │
│  ├── 领域模型（Note, Vault, User, File）                     │
│  ├── 领域服务                                                │
│  └── Repository 接口定义                                     │
├─────────────────────────────────────────────────────────────┤
│  Infrastructure Layer                                        │
│  ├── Repository 实现 (dao/)                                  │
│  ├── 外部服务适配器                                          │
│  └── 配置、日志、追踪                                        │
├─────────────────────────────────────────────────────────────┤
│  Shared Kernel (pkg/)                                        │
│  ├── code/ - 不可变错误码                                    │
│  ├── app/ - Token 管理（可配置过期时间）                     │
│  ├── errors/ - 统一错误处理                                  │
│  └── tracer/ - 请求追踪                                      │
└─────────────────────────────────────────────────────────────┘
```

## Dependency Flow

```
Transport → Application → Domain ← Infrastructure
                ↓
          Shared Kernel
```

## Key Components

### 1. Application Container (`internal/app/app.go`)

应用容器封装了所有依赖和服务的初始化：

- **Config**: 配置管理
- **Logger**: 日志服务
- **DB**: 数据库连接
- **Services**: 业务服务（UserService, NoteService, VaultService, FileService）
- **TokenManager**: JWT Token 管理

### 2. Domain Layer (`internal/domain/`)

领域层定义了核心业务模型和 Repository 接口：

- **Note**: 笔记实体
- **Vault**: 仓库实体
- **User**: 用户实体
- **File**: 文件实体
- **Repository Interfaces**: 数据访问抽象接口

### 3. Service Layer (`internal/service/`)

服务层实现业务逻辑：

- **VaultService**: Vault 管理，使用 Singleflight 合并并发请求
- **NoteService**: 笔记 CRUD 操作
- **UserService**: 用户认证和管理
- **FileService**: 文件上传和管理

### 4. Infrastructure Layer (`internal/dao/`)

基础设施层实现 Repository 接口：

- **NoteRepository**: 笔记数据访问
- **VaultRepository**: Vault 数据访问
- **UserRepository**: 用户数据访问
- **FileRepository**: 文件数据访问

### 5. Transport Layer (`internal/routers/`)

传输层处理 HTTP 和 WebSocket 请求：

- **api_router/**: REST API 处理器
- **websocket_router/**: WebSocket 消息处理器
- **middleware/**: 中间件（Trace、Auth、Lang）

## Design Principles

### 1. Immutable Code Objects

错误码对象采用不可变设计，所有 `With*` 方法返回新实例：

```go
// 正确用法
newCode := code.ErrorInvalidParams.WithDetails("field is required")

// 原对象不会被修改
```

### 2. Dependency Injection

所有服务通过构造函数注入依赖：

```go
func NewNoteService(noteRepo domain.NoteRepository, vaultSvc VaultService) NoteService {
    return &noteService{
        noteRepo:     noteRepo,
        vaultService: vaultSvc,
    }
}
```

### 3. Request Tracing

每个请求都有唯一的 Trace ID，用于链路追踪：

- 请求头 `X-Trace-ID` 传递
- 响应头包含 Trace ID
- 日志记录包含 Trace ID

### 4. Unified Error Handling

统一的错误响应格式：

```json
{
    "code": 10001,
    "message": "Invalid parameters",
    "details": ["field is required"],
    "traceId": "1704067200-abc123",
    "timestamp": "2024-01-01T00:00:00Z"
}
```

## Configuration

### Token Configuration

```yaml
security:
  auth-token-key: "your-secret-key"
  token-expiry: "7d"  # 支持: 7d, 24h, 30m
```

### Database Connection Pool

```yaml
database:
  max-open-conns: 100      # 最大打开连接数
  max-idle-conns: 10       # 最大空闲连接数
  conn-max-lifetime: "30m" # 连接最大生命周期
  conn-max-idle-time: "10m" # 空闲连接最大生命周期
```

### Request Tracing

```yaml
tracer:
  enabled: true
  header: "X-Trace-ID"
```

## Security Notes

⚠️ **重要**: 首次部署时请修改默认密钥！

系统会检测以下默认密钥并发出警告：
- `6666`
- `fast-note-sync-Auth-Token`
- 空字符串

生成安全密钥：
```bash
openssl rand -base64 32
```
