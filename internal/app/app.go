// Package app provides application container, encapsulates all dependencies and services
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
	// Infrastructure (injected dependencies)
	// 基础设施（注入的依赖）
	config *AppConfig
	logger *zap.Logger
	DB     *gorm.DB
	Dao    *dao.Dao

	// Concurrency control components
	// 并发控制组件
	workerPool    *workerpool.Pool
	writeQueueMgr *writequeue.Manager

	// Repository layer
	// Repository 层
	NoteRepo        domain.NoteRepository
	VaultRepo       domain.VaultRepository
	UserRepo        domain.UserRepository
	FileRepo        domain.FileRepository
	SettingRepo     domain.SettingRepository
	NoteHistoryRepo domain.NoteHistoryRepository
	NoteLinkRepo    domain.NoteLinkRepository
	ShareRepo       domain.UserShareRepository
	FolderRepo      domain.FolderRepository
	StorageRepo     domain.StorageRepository

	// Service layer
	// Service 层
	VaultService       service.VaultService
	NoteService        service.NoteService
	UserService        service.UserService
	FileService        service.FileService
	SettingService     service.SettingService
	NoteHistoryService service.NoteHistoryService
	ConflictService    service.ConflictService
	ShareService       service.ShareService
	NoteLinkService    service.NoteLinkService
	FolderService      service.FolderService
	StorageService     service.StorageService

	// Infrastructure components
	// 基础设施组件
	TokenManager pkgapp.TokenManager

	// Shutdown control
	// 关闭控制
	shutdownCh chan struct{}
	wg         sync.WaitGroup

	// Version check info
	// 版本检查信息
	checkVersionMu sync.RWMutex
	checkVersion   pkgapp.CheckVersionInfo

	// Startup time (for uptime calculation)
	// 启动时间（用于计算 uptime）
	StartTime time.Time
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
		StartTime:  time.Now(),
	}

	// Initialize Worker Pool
	// 初始化 Worker Pool
	wpConfig := cfg.GetWorkerPoolConfig()
	a.workerPool = workerpool.New(&wpConfig, logger)

	// Initialize Write Queue Manager
	// 初始化 Write Queue Manager
	wqConfig := cfg.GetWriteQueueConfig()
	a.writeQueueMgr = writequeue.New(&wqConfig, logger)

	// Create DatabaseConfig for DAO
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

	// Initialize DAO (using dependency injection)
	// 初始化 DAO（使用依赖注入）
	a.Dao = dao.New(db, context.Background(),
		dao.WithConfig(dbConfig),
		dao.WithLogger(logger),
		dao.WithWriteQueueManager(a.writeQueueMgr),
	)

	// Initialize TokenManager
	// 初始化 TokenManager
	tokenConfig := pkgapp.TokenConfig{
		SecretKey:     cfg.Security.AuthTokenKey,
		Issuer:        "fast-note-sync-service",
		Expiry:        cfg.GetTokenExpiry(),
		ShareTokenKey: cfg.Security.ShareTokenKey,
		ShareExpiry:   cfg.GetShareTokenExpiry(),
	}
	a.TokenManager = pkgapp.NewTokenManager(tokenConfig)

	// Initialize Repository layer
	// 初始化 Repository 层
	a.NoteRepo = dao.NewNoteRepository(a.Dao)
	a.VaultRepo = dao.NewVaultRepository(a.Dao)
	a.UserRepo = dao.NewUserRepository(a.Dao)
	a.FileRepo = dao.NewFileRepository(a.Dao)
	a.SettingRepo = dao.NewSettingRepository(a.Dao)
	a.NoteHistoryRepo = dao.NewNoteHistoryRepository(a.Dao)
	a.NoteLinkRepo = dao.NewNoteLinkRepository(a.Dao)
	a.ShareRepo = dao.NewUserShareRepository(a.Dao)
	a.FolderRepo = dao.NewFolderRepository(a.Dao)
	a.StorageRepo = dao.NewStorageRepository(a.Dao)

	// Create ServiceConfig (extract config needed by Service layer from AppConfig)
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

	// Initialize Service layer (dependency injection)
	a.VaultService = service.NewVaultService(a.VaultRepo)
	a.FolderService = service.NewFolderService(a.FolderRepo, a.NoteRepo, a.FileRepo, a.VaultService)
	a.NoteService = service.NewNoteService(a.NoteRepo, a.NoteLinkRepo, a.FileRepo, a.VaultService, a.FolderService, svcConfig)
	a.UserService = service.NewUserService(a.UserRepo, a.TokenManager, logger, svcConfig)
	a.FileService = service.NewFileService(a.FileRepo, a.NoteRepo, a.VaultService, a.FolderService, svcConfig)
	a.SettingService = service.NewSettingService(a.SettingRepo, a.VaultService, svcConfig)
	a.NoteHistoryService = service.NewNoteHistoryService(a.NoteHistoryRepo, a.NoteRepo, a.UserRepo, a.VaultService, a.FolderService, a.NoteService, logger, &svcConfig.App)
	a.ConflictService = service.NewConflictService(a.NoteRepo, a.VaultService, logger)
	a.ShareService = service.NewShareService(a.ShareRepo, a.TokenManager, a.NoteRepo, a.FileRepo, a.VaultRepo, logger, svcConfig)
	a.NoteLinkService = service.NewNoteLinkService(a.NoteLinkRepo, a.NoteRepo, a.VaultService)
	a.StorageService = service.NewStorageService(a.StorageRepo)

	logger.Info("App container initialized successfully",
		zap.Int("workerPoolMaxWorkers", wpConfig.MaxWorkers),
		zap.Int("writeQueueCapacity", wqConfig.QueueCapacity))

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

	// If there is no update, set version name to empty
	// 如果没有更新，把版本名称设置为空
	if !cv.VersionIsNew {
		cv.VersionNewName = ""
	}
	if !cv.PluginVersionIsNew {
		cv.PluginVersionNewName = ""
	}

	// Version number returned to client does not have v prefix
	// 返回给客户端的版本号不带 v 前缀
	cv.VersionNewName = strings.TrimPrefix(cv.VersionNewName, "v")
	cv.PluginVersionNewName = strings.TrimPrefix(cv.PluginVersionNewName, "v")
	// Supplying link information
	// 补充链接信息
	if cv.VersionNewLink == "" && cv.VersionNewName != "" {
		cv.VersionNewLink = "https://github.com/haierkeys/fast-note-sync-service/releases/tag/" + cv.VersionNewName
	}
	if cv.PluginVersionNewLink == "" && cv.PluginVersionNewName != "" {
		cv.PluginVersionNewLink = "https://github.com/haierkeys/obsidian-fast-note-sync/releases/tag/" + cv.PluginVersionNewName
	}

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

// ExecuteWrite executes write operation (serialized through Write Queue)
// ExecuteWrite 执行写操作（通过 Write Queue 串行化）
// uid: user ID, used to determine write queue
// uid: 用户 ID，用于确定写队列
// fn: write operation function
// fn: 写操作函数
func (a *App) ExecuteWrite(ctx context.Context, uid int64, fn func() error) error {
	return a.writeQueueMgr.Execute(ctx, uid, fn)
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
