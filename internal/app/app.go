// Package app 提供应用容器，封装所有依赖和服务
package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/workerpool"
	"github.com/haierkeys/fast-note-sync-service/pkg/writequeue"
	"golang.org/x/mod/semver"

	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 客户端名称常量
const (
	// WebClientName Web 客户端名称
	WebClientName = "Web"
)

// App 应用容器，封装所有依赖和服务
type App struct {
	// 基础设施（注入的依赖）
	config *AppConfig
	logger *zap.Logger
	DB     *gorm.DB
	Dao    *dao.Dao

	// 并发控制组件
	workerPool    *workerpool.Pool
	writeQueueMgr *writequeue.Manager

	// Repository 层
	NoteRepo        domain.NoteRepository
	VaultRepo       domain.VaultRepository
	UserRepo        domain.UserRepository
	FileRepo        domain.FileRepository
	SettingRepo     domain.SettingRepository
	NoteHistoryRepo domain.NoteHistoryRepository
	ShareRepo       domain.UserShareRepository

	// Service 层
	VaultService       service.VaultService
	NoteService        service.NoteService
	UserService        service.UserService
	FileService        service.FileService
	SettingService     service.SettingService
	NoteHistoryService service.NoteHistoryService
	ConflictService    service.ConflictService
	ShareService       service.ShareService

	// 基础设施组件
	TokenManager pkgapp.TokenManager

	// 关闭控制
	shutdownCh chan struct{}
	wg         sync.WaitGroup

	// 版本检查信息
	checkVersionMu sync.RWMutex
	checkVersion   pkgapp.CheckVersionInfo
}

// NewApp 创建应用容器实例
// 初始化所有依赖并进行依赖注入
// cfg: 应用配置（必须）
// logger: zap 日志器（必须）
// db: 数据库连接（必须）
func NewApp(cfg *AppConfig, logger *zap.Logger, db *gorm.DB) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	a := &App{
		config:     cfg,
		logger:     logger,
		DB:         db,
		shutdownCh: make(chan struct{}),
	}

	// 初始化 Worker Pool
	wpConfig := cfg.GetWorkerPoolConfig()
	a.workerPool = workerpool.New(&wpConfig, logger)

	// 初始化 Write Queue Manager
	wqConfig := cfg.GetWriteQueueConfig()
	a.writeQueueMgr = writequeue.New(&wqConfig, logger)

	// 创建 DatabaseConfig 用于 DAO
	dbConfig := &dao.DatabaseConfig{
		Type:            cfg.Database.Type,
		Path:            cfg.Database.Path,
		UserName:        cfg.Database.UserName,
		Password:        cfg.Database.Password,
		Host:            cfg.Database.Host,
		Name:            cfg.Database.Name,
		TablePrefix:     cfg.Database.TablePrefix,
		AutoMigrate:     cfg.Database.AutoMigrate,
		Charset:         cfg.Database.Charset,
		ParseTime:       cfg.Database.ParseTime,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
		RunMode:         cfg.Server.RunMode,
	}

	// 初始化 DAO（使用依赖注入）
	a.Dao = dao.New(db, context.Background(),
		dao.WithConfig(dbConfig),
		dao.WithLogger(logger),
		dao.WithWriteQueueManager(a.writeQueueMgr),
	)

	// 初始化 TokenManager
	tokenConfig := pkgapp.TokenConfig{
		SecretKey:     cfg.Security.AuthTokenKey,
		Issuer:        "fast-note-sync-service",
		Expiry:        cfg.GetTokenExpiry(),
		ShareTokenKey: cfg.Security.ShareTokenKey,
		ShareExpiry:   cfg.GetShareTokenExpiry(),
	}
	a.TokenManager = pkgapp.NewTokenManager(tokenConfig)

	// 初始化 Repository 层
	a.NoteRepo = dao.NewNoteRepository(a.Dao)
	a.VaultRepo = dao.NewVaultRepository(a.Dao)
	a.UserRepo = dao.NewUserRepository(a.Dao)
	a.FileRepo = dao.NewFileRepository(a.Dao)
	a.SettingRepo = dao.NewSettingRepository(a.Dao)
	a.NoteHistoryRepo = dao.NewNoteHistoryRepository(a.Dao)
	a.ShareRepo = dao.NewUserShareRepository(a.Dao)

	// 创建 ServiceConfig（从 AppConfig 提取 Service 层需要的配置）
	svcConfig := &service.ServiceConfig{
		User: service.UserServiceConfig{
			RegisterIsEnable: cfg.User.RegisterIsEnable,
		},
		App: service.AppServiceConfig{
			SoftDeleteRetentionTime: cfg.App.SoftDeleteRetentionTime,
			HistoryKeepVersions:     cfg.App.HistoryKeepVersions,
			HistorySaveDelay:        cfg.App.HistorySaveDelay,
			ShareTokenExpiry:        cfg.Security.ShareTokenExpiry,
		},
	}

	// 初始化 Service 层（依赖注入）
	a.VaultService = service.NewVaultService(a.VaultRepo)
	a.NoteService = service.NewNoteService(a.NoteRepo, a.FileRepo, a.VaultService, svcConfig)
	a.UserService = service.NewUserService(a.UserRepo, a.TokenManager, logger, svcConfig)
	a.FileService = service.NewFileService(a.FileRepo, a.NoteRepo, a.VaultService, svcConfig)
	a.SettingService = service.NewSettingService(a.SettingRepo, a.VaultService, svcConfig)
	a.NoteHistoryService = service.NewNoteHistoryService(a.NoteHistoryRepo, a.NoteRepo, a.UserRepo, a.VaultService, logger, &svcConfig.App)
	a.ConflictService = service.NewConflictService(a.NoteRepo, a.VaultService, logger)
	a.ShareService = service.NewShareService(a.ShareRepo, a.TokenManager, a.NoteRepo, a.FileRepo, a.VaultRepo, logger, svcConfig)

	logger.Info("App container initialized successfully",
		zap.Int("workerPoolMaxWorkers", wpConfig.MaxWorkers),
		zap.Int("writeQueueCapacity", wqConfig.QueueCapacity))

	return a, nil
}

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

// Config 获取应用配置
func (a *App) Config() *AppConfig {
	return a.config
}

// Logger 获取日志器
func (a *App) Logger() *zap.Logger {
	return a.logger
}

// SubmitTask 提交任务到 Worker Pool
// 返回错误如果池已满或已关闭
func (a *App) SubmitTask(ctx context.Context, task func(context.Context) error) error {
	return a.workerPool.Submit(ctx, task)
}

// SubmitTaskAsync 异步提交任务到 Worker Pool（不等待结果）
// 返回错误如果池已满或已关闭
func (a *App) SubmitTaskAsync(ctx context.Context, task func(context.Context) error) error {
	return a.workerPool.SubmitAsync(ctx, task)
}

// Version 获取版本信息
func (a *App) Version() pkgapp.VersionInfo {
	return pkgapp.VersionInfo{
		Version:   Version,
		GitTag:    GitTag,
		BuildTime: BuildTime,
	}
}

// CheckVersion 获取版本信息
func (a *App) CheckVersion(pluginVersion string) pkgapp.CheckVersionInfo {
	a.checkVersionMu.RLock()
	defer a.checkVersionMu.RUnlock()

	cv := a.checkVersion

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

	// 如果没有更新，把版本名称设置为空
	if !cv.VersionIsNew {
		cv.VersionNewName = ""
	}
	if !cv.PluginVersionIsNew {
		cv.PluginVersionNewName = ""
	}

	// 返回给客户端的版本号不带 v 前缀
	cv.VersionNewName = strings.TrimPrefix(cv.VersionNewName, "v")
	cv.PluginVersionNewName = strings.TrimPrefix(cv.PluginVersionNewName, "v")
	// 补充链接信息
	if cv.VersionNewLink == "" && cv.VersionNewName != "" {
		cv.VersionNewLink = "https://github.com/haierkeys/fast-note-sync-service/releases/tag/" + cv.VersionNewName
	}
	if cv.PluginVersionNewLink == "" && cv.PluginVersionNewName != "" {
		cv.PluginVersionNewLink = "https://github.com/haierkeys/obsidian-fast-note-sync/releases/tag/" + cv.PluginVersionNewName
	}

	return cv
}

// SetCheckVersionInfo 设置版本检查信息
func (a *App) SetCheckVersionInfo(info pkgapp.CheckVersionInfo) {
	a.checkVersionMu.Lock()
	defer a.checkVersionMu.Unlock()
	a.checkVersion = info
}

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

// IsReturnSuccess 是否返回成功响应
func (a *App) IsReturnSuccess() bool {
	return a.config.App.IsReturnSussess
}

// GetAuthTokenKey 获取 Token 密钥
func (a *App) GetAuthTokenKey() string {
	return a.config.Security.AuthTokenKey
}

// IsProductionMode 是否为生产模式
// 根据日志配置中的 Production 字段判断
func (a *App) IsProductionMode() bool {
	return a.config.Log.Production
}

// ExecuteWrite 执行写操作（通过 Write Queue 串行化）
// uid: 用户 ID，用于确定写队列
// fn: 写操作函数
func (a *App) ExecuteWrite(ctx context.Context, uid int64, fn func() error) error {
	return a.writeQueueMgr.Execute(ctx, uid, fn)
}

// WorkerPool 获取 Worker Pool（用于高级操作）
func (a *App) WorkerPool() *workerpool.Pool {
	return a.workerPool
}

// WriteQueueManager 获取 Write Queue Manager（用于高级操作）
func (a *App) WriteQueueManager() *writequeue.Manager {
	return a.writeQueueMgr
}

// GetNoteService 获取 NoteService，支持设置客户端信息
func (a *App) GetNoteService(clientName, clientVersion string) service.NoteService {
	if clientName != "" || clientVersion != "" {
		return a.NoteService.WithClient(clientName, clientVersion)
	}
	return a.NoteService
}

// GetFileService 获取 FileService，支持设置客户端信息
func (a *App) GetFileService(clientName, clientVersion string) service.FileService {
	if clientName != "" || clientVersion != "" {
		return a.FileService.WithClient(clientName, clientVersion)
	}
	return a.FileService
}

// DefaultShutdownTimeout 默认关闭超时时间
const DefaultShutdownTimeout = 30 * time.Second

// Shutdown 优雅关闭应用容器
// 按顺序关闭：Worker Pool -> Write Queue Manager -> Database
// ctx 用于控制关闭超时，如果为 nil 则使用默认 30 秒超时
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("App container shutting down...")

	// 如果没有提供 context，使用默认超时
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), DefaultShutdownTimeout)
		defer cancel()
	}

	// 标记关闭
	select {
	case <-a.shutdownCh:
		// 已经关闭
		return nil
	default:
		close(a.shutdownCh)
	}

	var errs []error

	// 0. 关闭 ShareService（同步最后的统计数据）
	if a.ShareService != nil {
		a.logger.Info("Shutting down share service...")
		if err := a.ShareService.Shutdown(ctx); err != nil {
			a.logger.Warn("Share service shutdown error", zap.Error(err))
		}
	}

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

// IsShuttingDown 检查应用是否正在关闭
func (a *App) IsShuttingDown() bool {
	select {
	case <-a.shutdownCh:
		return true
	default:
		return false
	}
}

// ShutdownCh 返回关闭信号通道（用于监听关闭事件）
func (a *App) ShutdownCh() <-chan struct{} {
	return a.shutdownCh
}

// TrackOperation 跟踪后台操作（用于优雅关闭时等待）
// 返回一个函数，在操作完成时调用
func (a *App) TrackOperation() func() {
	a.wg.Add(1)
	return func() {
		a.wg.Done()
	}
}
