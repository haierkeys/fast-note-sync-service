// Package app 提供应用容器，封装所有依赖和服务
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App 应用容器，封装所有依赖和服务
type App struct {
	// 基础设施
	Logger *zap.Logger
	DB     *gorm.DB
	Dao    *dao.Dao

	// Repository 层
	NoteRepo        domain.NoteRepository
	VaultRepo       domain.VaultRepository
	UserRepo        domain.UserRepository
	FileRepo        domain.FileRepository
	SettingRepo     domain.SettingRepository
	NoteHistoryRepo domain.NoteHistoryRepository

	// Service 层
	VaultService       service.VaultService
	NoteService        service.NoteService
	UserService        service.UserService
	FileService        service.FileService
	SettingService     service.SettingService
	NoteHistoryService service.NoteHistoryService

	// 基础设施组件
	TokenManager app.TokenManager
}

// NewApp 创建应用容器实例
// 初始化所有依赖并进行依赖注入
// 注意：此函数依赖 global.Config 已经初始化
func NewApp(logger *zap.Logger, db *gorm.DB) (*App, error) {
	if global.Config == nil {
		return nil, fmt.Errorf("global.Config is not initialized")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	a := &App{
		Logger: logger,
		DB:     db,
	}

	// 初始化 DAO
	a.Dao = dao.New(db, context.Background())

	// 初始化 TokenManager
	tokenConfig := app.TokenConfig{
		SecretKey: global.Config.Security.AuthTokenKey,
		Issuer:    global.Name,
	}
	// 解析 Token 过期时间
	if global.Config.Security.TokenExpiry != "" {
		if expiry, err := util.ParseDuration(global.Config.Security.TokenExpiry); err == nil {
			tokenConfig.Expiry = expiry
		}
	}
	if tokenConfig.Expiry == 0 {
		tokenConfig.Expiry = 7 * 24 * time.Hour // 默认 7 天
	}
	a.TokenManager = app.NewTokenManager(tokenConfig)

	// 初始化 Repository 层
	a.NoteRepo = dao.NewNoteRepository(a.Dao)
	a.VaultRepo = dao.NewVaultRepository(a.Dao)
	a.UserRepo = dao.NewUserRepository(a.Dao)
	a.FileRepo = dao.NewFileRepository(a.Dao)
	a.SettingRepo = dao.NewSettingRepository(a.Dao)
	a.NoteHistoryRepo = dao.NewNoteHistoryRepository(a.Dao)

	// 初始化 Service 层（依赖注入）
	a.VaultService = service.NewVaultService(a.VaultRepo)
	a.NoteService = service.NewNoteService(a.NoteRepo, a.FileRepo, a.VaultService)
	a.UserService = service.NewUserService(a.UserRepo, a.TokenManager)
	a.FileService = service.NewFileService(a.FileRepo, a.VaultService)
	a.SettingService = service.NewSettingService(a.SettingRepo, a.VaultService)
	a.NoteHistoryService = service.NewNoteHistoryService(a.NoteHistoryRepo, a.NoteRepo, a.VaultService)

	logger.Info("App container initialized successfully")

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
		a.Logger.Info("Database connection closed")
	}
	return nil
}

// GetNoteService 获取 NoteService，支持设置客户端信息
func (a *App) GetNoteService(clientName, clientVersion string) service.NoteService {
	if clientName != "" || clientVersion != "" {
		return a.NoteService.WithClient(clientName, clientVersion)
	}
	return a.NoteService
}

// GetConfig 获取全局配置（便捷方法）
func (a *App) GetConfig() interface{} {
	return global.Config
}
