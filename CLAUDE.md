# fast-note-sync-service

Go + Gin 笔记同步服务，内嵌前端静态资源（embed.FS）。

后端使用的webgui：源码在/Users/admin/Documents/Tech/project/fns/fast-note-sync-webgui

Obsidian里的插件端，源码在/Users/admin/Documents/Tech/project/fns/obsidian-fast-note-sync

如果需要探究Obsidian的API，源码在/Users/admin/Documents/Tech/project/obsidian-api

Obsidian开发者文档：/Users/admin/Documents/Tech/project/obsidian-developer-docs

短网址服务sink.cool的源码：/Users/admin/Documents/Tech/project/fns/sink

涉及在Notebook navigator里显示分享状态图标，notebook navigator的源码在/Users/admin/Documents/Tech/project/notebook-navigator

## 命令

```bash
# 本地运行（必须带 run 子命令）
go run . run

# 生产构建（推荐用 make 以注入版本信息；直接 go build 会丢失版本字段）
make build-linux-amd64   # 交叉编译 → Linux x86-64（部署目标）
make build               # 编译 native 平台（本地调试用）
# 若有新依赖下载超时，加国内代理前缀：
GOPROXY=https://goproxy.cn,direct make build-linux-amd64

# 部署（必须先 stop，否则 scp 因文件被占用报 dest open failure）
ssh -p 28256 root@u.facaix.fun "systemctl stop fast-note"
scp -P 28256 build/linux_amd64/fast-note-sync-service root@u.facaix.fun:/opt/fast-note/fast-note-sync-service
ssh -p 28256 root@u.facaix.fun "chmod +x /opt/fast-note/fast-note-sync-service && systemctl start fast-note"
```

## 前端编译 & 内嵌流程

前端通过 `embed.FS` 编译进二进制，修改前端后必须执行以下步骤：

```bash
# 1. 编译前端
cd /Users/admin/Documents/Tech/project/fns/fast-note-sync-webgui && npm run build

# 2. 同步产物到后端（share.html 不要删除）
rsync -a --delete dist/ /Users/admin/Documents/Tech/project/fns/fast-note-sync-service/frontend/

# 3. 重启后端服务
go run . run
```

## 架构

| 层     | 目录                     | 职责                                      |
| ------ | ------------------------ | ----------------------------------------- |
| 路由   | `internal/routers/`    | Gin 路由、REST Handler、WebSocket Handler |
| 领域   | `internal/domain/`     | 领域模型定义 + Repository 接口            |
| DTO    | `internal/dto/`        | API 入出参结构体                          |
| 服务   | `internal/service/`    | 核心业务逻辑                              |
| DAO    | `internal/dao/`        | 数据库 CRUD                               |
| 查询   | `internal/query/`      | GORM Gen 自动生成的类型安全查询           |
| 模型   | `internal/model/`      | ORM 数据库模型（自动生成）                |
| 中间件 | `internal/middleware/` | 认证、限流、CORS、日志等                  |
| 任务   | `internal/task/`       | 定时后台任务（备份、清理、版本检查）      |

## 核心工具包（pkg/）

| 包 | 职责 |
|----|------|
| `pkg/writequeue` | 写入队列（串行化并发写操作，防 SQLite 锁冲突）|
| `pkg/workerpool` | 协程池（限制并发 worker 数）|
| `pkg/storage`    | 文件存储抽象（本地/S3/OSS 统一接口）|
| `pkg/shortlink`  | 短链接客户端（`sink_cool.go` 调用 sink.cool API）|

## 数据库

SQLite，路径：`storage/database/db.sqlite3`

主要表：`notes`、`files`、`folders`、`vaults`、`users`、`user_shares`、`backup_configs`、`git_sync_configs`、`note_histories`、`settings`、`storages`

## API 认证

- 用户 API：Header `token`
- 分享页 API：Header `Share-Token`

## 响应格式

```go
{ code: int, status: bool, message: string, data: T }
```

## 注意事项

- `go run .` 只显示帮助，必须用 `go run . run` 才能启动服务
- 构建时优先用 `make build`（Makefile 通过 LDFLAGS 注入 GitTag/BuildTime 等版本信息到二进制）
- `frontend/share.html` 及 `share-*.js` 是分享阅读页，**不要删除**
- SQLite 时区：`glebarez/sqlite` 读取时间默认 UTC，存储时需注意时区处理
- 默认端口：9000
- 短链接客户端：`pkg/shortlink/sink_cool.go`
- 分享资源类型：`share.ResType` 可为 `"note"` 或 `"file"`，操作 `share.ResID` 前必须先 switch ResType，不能假设始终是笔记 ID（参考 `share_service.go` 中 `ListShares` 的实现模式）
- 分享状态常量：`domain.UserShareStatusActive = 1`、`domain.UserShareStatusRevoked = 2`（定义在 `domain/domain_user_share.go`），DAO 层和服务层禁止直接写字面量 `1`/`2`
- `UserShareRepository` 提供 `UpdateStatusByRes`（单条 SQL）：撤销某资源分享时优先用此方法而非 `GetByRes` + `UpdateStatus` 两步
- 笔记删除（`noteService.Delete`）会自动调用 `shareRepo.UpdateStatusByRes` 撤销关联分享，新增删除流程时需遵循此模式
- `ShareGenerate` 创建前先调 `GetByRes` 检查是否已有 active 分享，有则先 `StopShare`（幂等保证）
- 分享 Create/Cancel 成功后通过 `h.WSS.BroadcastToUser(uid, code.Success, dto.ShareSyncRefresh)` 广播通知其他设备刷新分享状态

## Git 推送

```bash
git push origin master
```

## 向上游提 PR

```bash
git fetch upstream && git rebase upstream/master
git push origin master --force
gh pr create --repo haierkeys/fast-note-sync-service --head chenxiccc:master --base master
# obsidian 插件 PR 必须提到 file 分支（--base file）

# 提错目标分支时，直接修改无需关闭重开：
gh pr edit <PR号> --repo <owner/repo> --base <正确分支>
```
