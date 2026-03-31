// Package app provides application container, encapsulates all dependencies and services
// Package app 提供应用容器，封装所有依赖和服务
package app

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/service"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/workerpool"
	"github.com/haierkeys/fast-note-sync-service/pkg/writequeue"
	"golang.org/x/mod/semver"

	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Client name constants
// 客户端名称常量
const (
	// WebClientName Web client name
	// WebClientName Web 客户端名称
	WebClientName = "Web"
)

// App application container, encapsulates all dependencies and services
// App 应用容器，封装所有依赖和服务
type App struct {
	// Embedded sub-containers
	*Infra
	*Repositories
	*Services

	// App-level state and control
	shutdownCh       chan struct{}
	UpgradeSignal    chan string
	StartTime        time.Time
	wg               sync.WaitGroup
	checkVersionMu   sync.RWMutex
	checkVersion     pkgapp.CheckVersionInfo
	supportRecordsMu sync.RWMutex
	supportRecords   map[string][]pkgapp.SupportRecord
}

// NewApp creates application container instance
// NewApp 创建应用容器实例
// Initializes all dependencies and performs dependency injection
// 初始化所有依赖并进行依赖注入
// cfg: application configuration (required)
// cfg: 应用配置（必须）
// logger: zap logger (required)
// logger: zap 日志器（必须）
// db: database connection (required)
// db: 数据库连接（必须）
// efs: frontend files embedded file system
// efs: 前端文件嵌入文件系统
func NewApp(cfg *AppConfig, logger *zap.Logger, db *gorm.DB, efs embed.FS) (*App, error) {
	if cfg == nil || logger == nil || db == nil {
		return nil, fmt.Errorf("config, logger and db are required")
	}

	// 1. Initialize Infrastructure
	infra, err := initInfra(cfg, logger, db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize infra: %w", err)
	}

	// 2. Initialize Repositories
	repos := initRepositories(infra.Dao)

	// 3. Initialize App shell
	a := &App{
		Infra:         infra,
		Repositories:  repos,
		shutdownCh:    make(chan struct{}),
		UpgradeSignal: make(chan string, 1),
		StartTime:     time.Now(),
	}

	// 4. Initialize Services (needs app context for some reason? No, it's just wiring)
	a.Services = initServices(cfg, infra, repos, logger)

	// Load support records
	a.loadSupportRecords(efs)

	logger.Info("App container initialized successfully")
	return a, nil
}

// Close releases resources held by application container
// Close 释放应用容器持有的资源
func (a *App) Close() error {
	if a.DB != nil {
		sqlDB, err := a.DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		a.logger.Info("Database connection closed")
	}
	return nil
}

// Config gets application configuration
// Config 获取应用配置
func (a *App) Config() *AppConfig {
	return a.config
}

// Logger gets logger
// Logger 获取日志器
func (a *App) Logger() *zap.Logger {
	return a.logger
}

// SubmitTask submits task to Worker Pool
// SubmitTask 提交任务到 Worker Pool
// returns error if pool is full or closed
// 返回错误如果池已满或已关闭
func (a *App) SubmitTask(ctx context.Context, task func(context.Context) error) error {
	return a.workerPool.Submit(ctx, task)
}

// SubmitTaskAsync asynchronously submits task to Worker Pool (does not wait for results)
// SubmitTaskAsync 异步提交任务到 Worker Pool（不等待结果）
// returns error if pool is full or closed
// 返回错误如果池已满或已关闭
func (a *App) SubmitTaskAsync(ctx context.Context, task func(context.Context) error) error {
	return a.workerPool.SubmitAsync(ctx, task)
}

// Version gets version information
// Version 获取版本信息
func (a *App) Version() pkgapp.VersionInfo {
	return pkgapp.VersionInfo{
		Version:   Version,
		GitTag:    GitTag,
		BuildTime: BuildTime,
	}
}

// CheckVersion gets version information
// CheckVersion 获取版本信息
func (a *App) CheckVersion(pluginVersion string) pkgapp.CheckVersionInfo {
	a.checkVersionMu.RLock()
	defer a.checkVersionMu.RUnlock()

	cv := a.checkVersion

	// Compare plugin versions
	// 比较插件版本
	if pluginVersion != "" && cv.PluginVersionNewName != "" {
		v1 := pluginVersion
		if !strings.HasPrefix(v1, "v") {
			v1 = "v" + v1
		}
		v2 := cv.PluginVersionNewName
		if !strings.HasPrefix(v2, "v") {
			v2 = "v" + v2
		}
		cv.PluginVersionIsNew = semver.Compare(v2, v1) > 0
	}

	// Version number returned to client does not have v prefix
	// 返回给客户端的版本号不带 v 前缀
	cv.VersionNewName = strings.TrimPrefix(cv.VersionNewName, "v")
	cv.PluginVersionNewName = strings.TrimPrefix(cv.PluginVersionNewName, "v")
	// Returns the link information as-is from setting (already set by task)
	// 直接返回设置中的链接信息（已由任务设置）
	return cv
}

// SetCheckVersionInfo sets version check information
// SetCheckVersionInfo 设置版本检查信息
func (a *App) SetCheckVersionInfo(info pkgapp.CheckVersionInfo) {
	a.checkVersionMu.Lock()
	defer a.checkVersionMu.Unlock()
	a.checkVersion = info
}

// Validator gets validator
// Validator 获取验证器
func (a *App) Validator() pkgapp.ValidatorInterface {
	if binding.Validator == nil {
		return nil
	}
	if v, ok := binding.Validator.(pkgapp.ValidatorInterface); ok {
		return v
	}
	return nil
}

// IsReturnSuccess whether to return success response
// IsReturnSuccess 是否返回成功响应
func (a *App) IsReturnSuccess() bool {
	return a.config.App.IsReturnSussess
}

// GetAuthTokenKey gets Token key
// GetAuthTokenKey 获取 Token 密钥
func (a *App) GetAuthTokenKey() string {
	return a.config.Security.AuthTokenKey
}

// IsProductionMode whether it is production mode
// IsProductionMode 是否为生产模式
// Judge based on Production field in log configuration
// 根据日志配置中的 Production 字段判断
func (a *App) IsProductionMode() bool {
	return a.config.Log.Production
}

// IsPullFromGitHub returns whether current source is GitHub
// IsPullFromGitHub 返回当前拉取源是否为 GitHub
func (a *App) IsPullFromGitHub() bool {
	return a.sourceSelector.IsGitHub()
}

// ExecuteWrite executes write operation (serialized through Write Queue)
// ExecuteWrite 执行写操作（通过 Write Queue 串行化）
// uid: user ID, used to determine write queue
// uid: 用户 ID，用于确定写队列
// fn: write operation function
// fn: 写操作函数
func (a *App) ExecuteWrite(ctx context.Context, uid int64, fn func() error) error {
	return a.writeQueueMgr.Execute(ctx, strconv.FormatInt(uid, 10), fn)
}

// WorkerPool gets Worker Pool (for advanced operations)
// WorkerPool 获取 Worker Pool（用于高级操作）
func (a *App) WorkerPool() *workerpool.Pool {
	return a.workerPool
}

// WriteQueueManager gets Write Queue Manager (for advanced operations)
// WriteQueueManager 获取 Write Queue Manager（用于高级操作）
func (a *App) WriteQueueManager() *writequeue.Manager {
	return a.writeQueueMgr
}

// GetNoteService gets NoteService, supports setting client info
// GetNoteService 获取 NoteService，支持设置客户端信息
func (a *App) GetNoteService(clientName, clientVersion string) service.NoteService {
	if clientName != "" || clientVersion != "" {
		return a.NoteService.WithClient(clientName, clientVersion)
	}
	return a.NoteService
}

// GetFileService gets FileService, supports setting client info
// GetFileService 获取 FileService，支持设置客户端信息
func (a *App) GetFileService(clientName, clientVersion string) service.FileService {
	if clientName != "" || clientVersion != "" {
		return a.FileService.WithClient(clientName, clientVersion)
	}
	return a.FileService
}

// loadSupportRecords loads support records from embedded file system
// loadSupportRecords 从嵌入文件系统加载打赏记录
func (a *App) loadSupportRecords(efs embed.FS) {
	a.supportRecordsMu.Lock()
	defer a.supportRecordsMu.Unlock()
	a.supportRecords = make(map[string][]pkgapp.SupportRecord)

	docsPath := "docs"
	entries, err := fs.ReadDir(efs, docsPath)
	if err != nil {
		a.logger.Warn("Failed to read docs directory from embedded FS", zap.Error(err))
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && strings.HasPrefix(name, "Support.") && strings.HasSuffix(name, ".json") {
			// Extract language from Support.{lang}.json
			// 从 Support.{lang}.json 提取语言
			parts := strings.Split(name, ".")
			if len(parts) < 3 {
				continue
			}
			lang := strings.ToLower(parts[1])

			data, err := efs.ReadFile(docsPath + "/" + name)
			if err != nil {
				a.logger.Warn("Failed to read support record file", zap.String("file", name), zap.Error(err))
				continue
			}

			var records []pkgapp.SupportRecord
			if err := json.Unmarshal(data, &records); err != nil {
				a.logger.Warn("Failed to unmarshal support records", zap.String("file", name), zap.Error(err))
				continue
			}

			a.supportRecords[lang] = records
			a.logger.Debug("Loaded support records", zap.String("lang", lang), zap.Int("count", len(records)))
		}
	}
}

// GetSupportRecords gets all support records
// GetSupportRecords 获取所有打赏记录
func (a *App) GetSupportRecords() map[string][]pkgapp.SupportRecord {
	a.supportRecordsMu.RLock()
	defer a.supportRecordsMu.RUnlock()
	return a.supportRecords
}

// GetSupportRecordsPage gets support records with pagination and sorting
// GetSupportRecordsPage 分页并排序获取打赏记录
func (a *App) GetSupportRecordsPage(lang, sortBy, sortOrder string, page, pageSize int) ([]pkgapp.SupportRecord, int) {
	a.supportRecordsMu.RLock()
	defer a.supportRecordsMu.RUnlock()

	lang = strings.ToLower(lang)
	if lang == "" {
		lang = "en"
	}

	records, ok := a.supportRecords[lang]
	if !ok {
		records = a.supportRecords["en"]
	}

	total := len(records)
	if total == 0 {
		return []pkgapp.SupportRecord{}, 0
	}

	sortedRecords := make([]pkgapp.SupportRecord, total)
	copy(sortedRecords, records)

	if sortBy != "" {
		isDesc := strings.ToLower(sortOrder) == "desc"
		sort.SliceStable(sortedRecords, func(i, j int) bool {
			var less bool
			switch sortBy {
			case "amount":
				amountI, _ := strconv.ParseFloat(sortedRecords[i].Amount, 64)
				amountJ, _ := strconv.ParseFloat(sortedRecords[j].Amount, 64)
				less = amountI < amountJ
			case "name":
				less = sortedRecords[i].Name < sortedRecords[j].Name
			case "item":
				less = sortedRecords[i].Item < sortedRecords[j].Item
			case "time":
				fallthrough
			default:
				less = sortedRecords[i].Time < sortedRecords[j].Time
			}
			if isDesc {
				return !less
			}
			return less
		})
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return []pkgapp.SupportRecord{}, total
	}

	end := offset + pageSize
	if end > total {
		end = total
	}

	return sortedRecords[offset:end], total
}

// UpdateSupportRecords updates support records for a specific language
// UpdateSupportRecords 更新特定语言的打赏记录
func (a *App) UpdateSupportRecords(lang string, records []pkgapp.SupportRecord) {
	if lang == "" {
		return
	}
	lang = strings.ToLower(lang)
	a.supportRecordsMu.Lock()
	defer a.supportRecordsMu.Unlock()
	if a.supportRecords == nil {
		a.supportRecords = make(map[string][]pkgapp.SupportRecord)
	}
	a.supportRecords[lang] = records
	a.logger.Debug("Updated support records via background task", zap.String("lang", lang), zap.Int("count", len(records)))
}

// TriggerUpgrade triggers the upgrade process
// TriggerUpgrade 触发升级流程
func (a *App) TriggerUpgrade(newBinaryPath string) {
	a.logger.Info("Triggering upgrade", zap.String("path", newBinaryPath))
	select {
	case a.UpgradeSignal <- newBinaryPath:
	default:
		a.logger.Warn("Upgrade signal already sent")
	}
}

// DefaultShutdownTimeout default shutdown timeout duration
// DefaultShutdownTimeout 默认关闭超时时间
const DefaultShutdownTimeout = 30 * time.Second

// Shutdown gracefully shuts down application container
// Shutdown 优雅关闭应用容器
// Close in order: Worker Pool -> Write Queue Manager -> Database
// 按顺序关闭：Worker Pool -> Write Queue Manager -> Database
// ctx used to control shutdown timeout, if nil use default 30 seconds timeout
// ctx 用于控制关闭超时，如果为 nil 则使用默认 30 秒超时
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("App container shutting down...")

	// If no context provided, use default timeout
	// 如果没有提供 context，使用默认超时
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), DefaultShutdownTimeout)
		defer cancel()
	}

	// Mark shutdown
	// 标记关闭
	select {
	case <-a.shutdownCh:
		// Already shut down
		// 已经关闭
		return nil
	default:
		close(a.shutdownCh)
	}

	var errs []error

	// 0. Shutdown ShareService (sync final statistics)
	// 0. 关闭 ShareService（同步最后的统计数据）
	if a.ShareService != nil {
		a.logger.Info("Shutting down share service...")
		if err := a.ShareService.Shutdown(ctx); err != nil {
			a.logger.Warn("Share service shutdown error", zap.Error(err))
		}
	}

	// 0.1 Shutdown NgrokService
	if a.NgrokService != nil {
		a.logger.Info("Shutting down ngrok service...")
		if err := a.NgrokService.Stop(ctx); err != nil {
			a.logger.Warn("Ngrok service shutdown error", zap.Error(err))
		}
	}

	// 0.2 Shutdown CloudflareService
	if a.CloudflareService != nil {
		a.logger.Info("Shutting down cloudflare service...")
		if err := a.CloudflareService.Stop(ctx); err != nil {
			a.logger.Warn("Cloudflare service shutdown error", zap.Error(err))
		}
	}

	// 0.3 Shutdown GitSyncService (wait for all sync goroutines to finish)
	// 0.3 关闭 GitSyncService（等待所有同步 goroutine 结束）
	if a.GitSyncService != nil {
		a.logger.Info("Shutting down git sync service...")
		if err := a.GitSyncService.Shutdown(ctx); err != nil {
			a.logger.Warn("Git sync service shutdown error", zap.Error(err))
		} else {
			a.logger.Info("Git sync service shutdown completed")
		}
	}

	// 0.4 Shutdown BackupService (wait for all backup goroutines to finish)
	// 0.4 关闭 BackupService（等待所有备份 goroutine 结束）
	if a.BackupService != nil {
		a.logger.Info("Shutting down backup service...")
		if err := a.BackupService.Shutdown(ctx); err != nil {
			a.logger.Warn("Backup service shutdown error", zap.Error(err))
		} else {
			a.logger.Info("Backup service shutdown completed")
		}
	}

	// 1. Shutdown Worker Pool (stop accepting new tasks, wait for existing tasks to complete)
	// 1. 关闭 Worker Pool（停止接受新任务，等待现有任务完成）
	if a.workerPool != nil {
		a.logger.Info("Shutting down worker pool...")
		if err := a.workerPool.Shutdown(ctx); err != nil {
			a.logger.Warn("Worker pool shutdown error", zap.Error(err))
			errs = append(errs, fmt.Errorf("worker pool shutdown: %w", err))
		} else {
			a.logger.Info("Worker pool shutdown completed")
		}
	}

	// 2. Shutdown Write Queue Manager (drain all queues)
	// 2. 关闭 Write Queue Manager（排空所有队列）
	if a.writeQueueMgr != nil {
		a.logger.Info("Shutting down write queue manager...")
		if err := a.writeQueueMgr.Shutdown(ctx); err != nil {
			a.logger.Warn("write queue manager shutdown error", zap.Error(err))
			errs = append(errs, fmt.Errorf("write queue manager shutdown: %w", err))
		} else {
			a.logger.Info("write queue manager shutdown completed")
		}
	}

	// 3. Wait for all background operations to complete
	// 3. 等待所有后台操作完成
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("All background operations completed")
	case <-ctx.Done():
		a.logger.Warn("Shutdown timeout waiting for background operations")
		errs = append(errs, fmt.Errorf("background operations timeout: %w", ctx.Err()))
	}

	// 4. Close database connection
	// 4. 关闭数据库连接
	if err := a.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		a.logger.Warn("App container shutdown completed with errors",
			zap.Int("errorCount", len(errs)))
		return fmt.Errorf("shutdown completed with %d errors: %v", len(errs), errs)
	}

	a.logger.Info("App container shutdown completed successfully")
	return nil
}

// IsShuttingDown checks if application is shutting down
// IsShuttingDown 检查应用是否正在关闭
func (a *App) IsShuttingDown() bool {
	select {
	case <-a.shutdownCh:
		return true
	default:
		return false
	}
}

// ShutdownCh returns shutdown signal channel (used for listening to shutdown events)
// ShutdownCh 返回关闭信号通道（用于监听关闭事件）
func (a *App) ShutdownCh() <-chan struct{} {
	return a.shutdownCh
}

// TrackOperation tracks background operations (used to wait during graceful shutdown)
// TrackOperation 跟踪后台操作（用于优雅关闭时等待）
// returns a function to be called when operation is complete
// 返回一个函数，在操作完成时调用
func (a *App) TrackOperation() func() {
	a.wg.Add(1)
	return func() {
		a.wg.Done()
	}
}
