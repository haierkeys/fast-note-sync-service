package api_router

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
	"go.uber.org/zap"
)

// AdminControlHandler Admin control configuration API router handler
// AdminControlHandler 管理控制配置 API 路由处理器
// Uses App Container to inject dependencies
// 使用 App Container 注入依赖
type AdminControlHandler struct {
	*Handler
	wss *pkgapp.WebsocketServer
}

// NewAdminControlHandler creates AdminControlHandler instance
// NewAdminControlHandler 创建 AdminControlHandler 实例
func NewAdminControlHandler(a *app.App, wss *pkgapp.WebsocketServer) *AdminControlHandler {
	return &AdminControlHandler{
		Handler: NewHandler(a),
		wss:     wss,
	}
}

// Config retrieves WebGUI configuration (public interface)
// @Summary Get WebGUI basic config
// @Description Get non-sensitive configuration required for frontend display, such as font settings, registration status, etc.
// @Tags Config
// @Produce json
// @Success 200 {object} pkgapp.Res{data=dto.AdminWebGUIConfig} "Success"
// @Router /api/webgui/config [get]
func (h *AdminControlHandler) Config(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	data := dto.AdminWebGUIConfig{
		FontSet:          cfg.WebGUI.FontSet,
		RegisterIsEnable: cfg.User.RegisterIsEnable,
		AdminUID:         cfg.User.AdminUID,
	}
	response.ToResponse(code.Success.WithData(data))
}

// GetConfig retrieves admin configuration (requires admin privileges)
// @Summary Get full admin config
// @Description Get full system configuration information, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=dto.AdminConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config [get]
func (h *AdminControlHandler) GetConfig(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.WebGUI.GetConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Deny access if AdminUID is configured and current user is not an admin
	// 当配置了管理员 UID 且当前用户不是管理员时，拒绝访问
	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	data := &dto.AdminConfig{
		FontSet:                 cfg.WebGUI.FontSet,
		RegisterIsEnable:        cfg.User.RegisterIsEnable,
		FileChunkSize:           cfg.App.FileChunkSize,
		SoftDeleteRetentionTime: cfg.App.SoftDeleteRetentionTime,
		UploadSessionTimeout:    cfg.App.UploadSessionTimeout,
		HistoryKeepVersions:     cfg.App.HistoryKeepVersions,
		HistorySaveDelay:        cfg.App.HistorySaveDelay,
		// DefaultAPIFolder:        cfg.App.DefaultAPIFolder,
		AdminUID:         cfg.User.AdminUID,
		AuthTokenKey:     cfg.Security.AuthTokenKey,
		TokenExpiry:      cfg.Security.TokenExpiry,
		ShareTokenKey:    cfg.Security.ShareTokenKey,
		ShareTokenExpiry: cfg.Security.ShareTokenExpiry,
		PullSource:       cfg.App.PullSource,
	}

	response.ToResponse(code.Success.WithData(data))
}

// UpdateConfig updates admin configuration (requires admin privileges)
// @Summary Update admin config
// @Description Modify full system configuration information, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.AdminConfig true "Config Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.AdminConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config [post]
func (h *AdminControlHandler) UpdateConfig(c *gin.Context) {
	params := &dto.AdminConfig{}
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		logger.Error("apiRouter.WebGUI.UpdateConfig.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.WebGUI.UpdateConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Deny access if AdminUID is configured and current user is not an admin
	// 当配置了管理员 UID 且当前用户不是管理员时，拒绝访问
	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	// Validate historyKeepVersions cannot be less than 100
	// 验证 historyKeepVersions 不能小于 100
	if params.HistoryKeepVersions > 0 && params.HistoryKeepVersions < 100 {
		logger.Warn("apiRouter.WebGUI.UpdateConfig invalid historyKeepVersions",
			zap.Int("value", params.HistoryKeepVersions))
		response.ToResponse(code.ErrorInvalidParams.WithDetails("historyKeepVersions must be at least 100"))
		return
	}

	// Validate historySaveDelay cannot be less than 10 seconds
	// 验证 historySaveDelay 不能小于 10 秒
	if params.HistorySaveDelay != "" {
		delay, err := util.ParseDuration(params.HistorySaveDelay)
		if err != nil {
			logger.Warn("apiRouter.WebGUI.UpdateConfig invalid historySaveDelay format",
				zap.String("value", params.HistorySaveDelay))
			response.ToResponse(code.ErrorInvalidParams.WithDetails("historySaveDelay format invalid, e.g. 10s, 1m"))
			return
		}
		if delay < 10*time.Second {
			logger.Warn("apiRouter.WebGUI.UpdateConfig historySaveDelay too small",
				zap.String("value", params.HistorySaveDelay))
			response.ToResponse(code.ErrorInvalidParams.WithDetails("historySaveDelay must be at least 10s"))
			return
		}
	}

	// Update configuration
	// 更新配置
	cfg.WebGUI.FontSet = params.FontSet
	cfg.User.RegisterIsEnable = params.RegisterIsEnable
	cfg.App.FileChunkSize = params.FileChunkSize
	cfg.App.SoftDeleteRetentionTime = params.SoftDeleteRetentionTime
	cfg.App.UploadSessionTimeout = params.UploadSessionTimeout
	cfg.App.HistoryKeepVersions = params.HistoryKeepVersions
	cfg.App.HistorySaveDelay = params.HistorySaveDelay
	//cfg.App.DefaultAPIFolder = params.DefaultAPIFolder
	cfg.User.AdminUID = params.AdminUID
	cfg.Security.AuthTokenKey = params.AuthTokenKey
	cfg.Security.TokenExpiry = params.TokenExpiry
	cfg.Security.ShareTokenKey = params.ShareTokenKey
	cfg.Security.ShareTokenExpiry = params.ShareTokenExpiry
	cfg.App.PullSource = params.PullSource

	// Save configuration to file
	// 保存配置到文件
	if err := cfg.Save(); err != nil {
		logger.Error("apiRouter.WebGUI.UpdateConfig.Save err", zap.Error(err))
		response.ToResponse(code.ErrorConfigSaveFailed)
		return
	}

	response.ToResponse(code.Success.WithData(params))
}

// GetUserDatabaseConfig retrieves user database configuration (requires admin privileges)
// @Summary Get user database config
// @Description Get user database configuration information, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=dto.AdminUserDatabaseConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config/user_database [get]
func (h *AdminControlHandler) GetUserDatabaseConfig(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.AdminControl.GetUserDatabaseConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Deny access if AdminUID is configured and current user is not an admin
	// 当配置了管理员 UID 且当前用户不是管理员时，拒绝访问
	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	dbCfg := cfg.UserDatabase
	data := &dto.AdminUserDatabaseConfig{
		Type:                dbCfg.Type,
		Path:                dbCfg.Path,
		UserName:            dbCfg.UserName,
		Password:            dbCfg.Password,
		Host:                dbCfg.Host,
		Port:                dbCfg.Port,
		Name:                dbCfg.Name,
		SSLMode:             dbCfg.SSLMode,
		Schema:              dbCfg.Schema,
		MaxIdleConns:        dbCfg.MaxIdleConns,
		MaxOpenConns:        dbCfg.MaxOpenConns,
		ConnMaxLifetime:     dbCfg.ConnMaxLifetime,
		ConnMaxIdleTime:     dbCfg.ConnMaxIdleTime,
		MaxWriteConcurrency: dbCfg.MaxWriteConcurrency,
		Charset:             dbCfg.Charset,
		ParseTime:           dbCfg.ParseTime,
	}
	response.ToResponse(code.Success.WithData(data))
}

// UpdateUserDatabaseConfig updates user database configuration (requires admin privileges)
// @Summary Update user database config
// @Description Modify user database configuration information, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.AdminUserDatabaseConfig true "Config Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.AdminUserDatabaseConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config/user_database [post]
func (h *AdminControlHandler) UpdateUserDatabaseConfig(c *gin.Context) {
	params := &dto.AdminUserDatabaseConfig{}
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		logger.Error("apiRouter.AdminControl.UpdateUserDatabaseConfig.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.AdminControl.UpdateUserDatabaseConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Deny access if AdminUID is configured and current user is not an admin
	// 当配置了管理员 UID 且当前用户不是管理员时，拒绝访问
	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	// Update configuration
	// 更新配置
	cfg.UserDatabase.Type = params.Type
	cfg.UserDatabase.Path = params.Path
	cfg.UserDatabase.UserName = params.UserName
	cfg.UserDatabase.Password = params.Password
	cfg.UserDatabase.Host = params.Host
	cfg.UserDatabase.Port = params.Port
	cfg.UserDatabase.Name = params.Name
	cfg.UserDatabase.SSLMode = params.SSLMode
	cfg.UserDatabase.Schema = params.Schema
	cfg.UserDatabase.MaxIdleConns = params.MaxIdleConns
	cfg.UserDatabase.MaxOpenConns = params.MaxOpenConns
	cfg.UserDatabase.ConnMaxLifetime = params.ConnMaxLifetime
	cfg.UserDatabase.ConnMaxIdleTime = params.ConnMaxIdleTime
	cfg.UserDatabase.MaxWriteConcurrency = params.MaxWriteConcurrency
	cfg.UserDatabase.Charset = params.Charset
	cfg.UserDatabase.ParseTime = params.ParseTime

	// MySQL specific hardcoded defaults
	// MySQL 的硬编码默认逻辑
	if params.Type == "mysql" {
		cfg.UserDatabase.Charset = "utf8mb4"
		cfg.UserDatabase.ParseTime = true
	}

	if params.Type == "sqlite" {
		enableQueue := true
		cfg.UserDatabase.EnableWriteQueue = &enableQueue
	} else if params.Type == "mysql" || params.Type == "postgres" {
		enableQueue := false
		cfg.UserDatabase.EnableWriteQueue = &enableQueue
	}

	// Save configuration to file
	// 保存配置到文件
	if err := cfg.Save(); err != nil {
		logger.Error("apiRouter.AdminControl.UpdateUserDatabaseConfig.Save err", zap.Error(err))
		response.ToResponse(code.ErrorConfigSaveFailed)
		return
	}

	response.ToResponse(code.Success.WithData(params))
}

// ValidateUserDatabaseConfig tests user database connection (requires admin privileges)
// @Summary Test user database connection
// ValidateUserDatabaseConfig tests user database connection (requires admin privileges)
// @Summary Test user database connection
// @Description Test if the provided database configuration can connect successfully, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.AdminUserDatabaseConfig true "Config Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Failure 400 {object} pkgapp.Res "Connection failed"
// @Router /api/admin/config/user_database/test [post]
func (h *AdminControlHandler) ValidateUserDatabaseConfig(c *gin.Context) {
	params := &dto.AdminUserDatabaseConfig{}
	response := pkgapp.NewResponse(c)
	logger := h.App.Logger()

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		logger.Error("apiRouter.AdminControl.ValidateUserDatabaseConfig.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	cfg := h.App.Config()
	// Deny access if AdminUID is configured and current user is not an admin
	// 当配置了管理员 UID 且当前用户不是管理员时，拒绝访问
	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	// Map DTO to DatabaseConfig
	// 将 DTO 映射到 DatabaseConfig
	enableQueue := false
	if params.Type == "sqlite" {
		enableQueue = true
	}
	dbCfg := config.DatabaseConfig{
		Type:                params.Type,
		Path:                params.Path,
		UserName:            params.UserName,
		Password:            params.Password,
		Host:                params.Host,
		Port:                params.Port,
		Name:                params.Name,
		SSLMode:             params.SSLMode,
		Schema:              params.Schema,
		AutoMigrate:         true,
		MaxIdleConns:        params.MaxIdleConns,
		MaxOpenConns:        params.MaxOpenConns,
		ConnMaxLifetime:     params.ConnMaxLifetime,
		ConnMaxIdleTime:     params.ConnMaxIdleTime,
		EnableWriteQueue:    &enableQueue,
		MaxWriteConcurrency: params.MaxWriteConcurrency,
		Charset:             params.Charset,
		ParseTime:           params.ParseTime,
	}

	// Apply hardcoded default rules for MySQL during validation
	// 在测试连接时也应用 MySQL 的硬编码默认规则
	if params.Type == "mysql" {
		dbCfg.Charset = "utf8mb4"
		dbCfg.ParseTime = true
	}

	// Use dao.NewEngine to test connection
	// 使用 dao.NewEngine 测试连接
	db, err := dao.NewEngine(dbCfg, h.App.Logger())
	if err != nil {
		logger.Warn("Database connection test failed", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails("Connection failed: " + err.Error()))
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		response.ToResponse(code.Failed.WithDetails("Failed to get DB instance: " + err.Error()))
		return
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		logger.Warn("Database ping failed", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails("Ping failed: " + err.Error()))
		return
	}

	// For MySQL, verify CREATE DATABASE privilege
	// 针对 MySQL，验证是否具有创建数据库的权限
	if params.Type == "mysql" {
		tempDBName := fmt.Sprintf("fn_perm_check_%d", time.Now().Unix())
		if err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", tempDBName)).Error; err != nil {
			logger.Warn("MySQL CREATE DATABASE permission test failed", zap.Error(err))
			response.ToResponse(code.Failed.WithDetails("Missing CREATE DATABASE permission: " + err.Error()))
			return
		}
		// Clean up immediately after successful verification
		// 验证成功后立即清理
		_ = db.Exec(fmt.Sprintf("DROP DATABASE %s", tempDBName))
	}

	response.ToResponse(code.Success.WithDetails("Database connection and permission verification successful"))
}

// GetNgrokConfig retrieves Ngrok tunnel configuration (requires admin privileges)
// @Summary Get Ngrok config
// @Description Get Ngrok tunnel configuration, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=dto.AdminNgrokConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config/ngrok [get]
func (h *AdminControlHandler) GetNgrokConfig(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.AdminControl.GetNgrokConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	data := &dto.AdminNgrokConfig{
		Enabled:   cfg.Ngrok.Enabled,
		AuthToken: cfg.Ngrok.AuthToken,
		Domain:    cfg.Ngrok.Domain,
	}

	response.ToResponse(code.Success.WithData(data))
}

// UpdateNgrokConfig updates Ngrok tunnel configuration (requires admin privileges)
// @Summary Update Ngrok config
// @Description Modify Ngrok tunnel configuration, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.AdminNgrokConfig true "Config Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.AdminNgrokConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config/ngrok [post]
func (h *AdminControlHandler) UpdateNgrokConfig(c *gin.Context) {
	params := &dto.AdminNgrokConfig{}
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		logger.Error("apiRouter.AdminControl.UpdateNgrokConfig.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.AdminControl.UpdateNgrokConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	cfg.Ngrok.Enabled = params.Enabled
	cfg.Ngrok.AuthToken = params.AuthToken
	cfg.Ngrok.Domain = params.Domain

	if err := cfg.Save(); err != nil {
		logger.Error("apiRouter.AdminControl.UpdateNgrokConfig.Save err", zap.Error(err))
		response.ToResponse(code.ErrorConfigSaveFailed)
		return
	}

	response.ToResponse(code.Success.WithData(params))
}

// GetCloudflareConfig retrieves Cloudflare tunnel configuration (requires admin privileges)
// @Summary Get Cloudflare config
// @Description Get Cloudflare tunnel configuration, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=dto.AdminCloudflareConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config/cloudflare [get]
func (h *AdminControlHandler) GetCloudflareConfig(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.AdminControl.GetCloudflareConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	data := &dto.AdminCloudflareConfig{
		Enabled:    cfg.Cloudflare.Enabled,
		Token:      cfg.Cloudflare.Token,
		LogEnabled: cfg.Cloudflare.LogEnabled,
	}

	response.ToResponse(code.Success.WithData(data))
}

// UpdateCloudflareConfig updates Cloudflare tunnel configuration (requires admin privileges)
// @Summary Update Cloudflare config
// @Description Modify Cloudflare tunnel configuration, requires admin privileges
// @Tags Config
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.AdminCloudflareConfig true "Config Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.AdminCloudflareConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config/cloudflare [post]
func (h *AdminControlHandler) UpdateCloudflareConfig(c *gin.Context) {
	params := &dto.AdminCloudflareConfig{}
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		logger.Error("apiRouter.AdminControl.UpdateCloudflareConfig.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.AdminControl.UpdateCloudflareConfig err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	cfg.Cloudflare.Enabled = params.Enabled
	cfg.Cloudflare.Token = params.Token
	cfg.Cloudflare.LogEnabled = params.LogEnabled

	if err := cfg.Save(); err != nil {
		logger.Error("apiRouter.AdminControl.UpdateCloudflareConfig.Save err", zap.Error(err))
		response.ToResponse(code.ErrorConfigSaveFailed)
		return
	}

	response.ToResponse(code.Success.WithData(params))
}

// GetSystemInfo retrieves system monitoring information (requires admin privileges)
// @Summary Get system stats
// @Description Get server runtime, CPU, memory, host and process info, requires admin privileges
// @Tags System
// @Produce json
// @Security UserAuthToken
// @Success 200 {object} pkgapp.Res{data=dto.AdminSystemInfo} "Success"
// @Router /api/admin/system/info [get]
func (h *AdminControlHandler) GetSystemInfo(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		logger.Error("apiRouter.WebGUI.GetSystemInfo err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	// Go Runtime
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// CPU
	cpuInfoList, _ := cpu.Info()
	cpuModel := ""
	if len(cpuInfoList) > 0 {
		cpuModel = cpuInfoList[0].ModelName
	}
	physCores, _ := cpu.Counts(false)
	logicCores, _ := cpu.Counts(true)
	cpuPercents, _ := cpu.Percent(time.Second, true)
	loadStat, _ := load.Avg()

	// Memory
	vMem, _ := mem.VirtualMemory()
	swapMem, _ := mem.SwapMemory()

	// Host
	hInfo, _ := host.Info()

	// Process
	p, _ := process.NewProcess(int32(os.Getpid()))
	pName, _ := p.Name()
	pPPid, _ := p.Ppid()
	pCPU, _ := p.CPUPercent()
	pMem, _ := p.MemoryPercent()

	data := dto.AdminSystemInfo{
		StartTime: h.App.StartTime,
		Uptime:    time.Since(h.App.StartTime).Seconds(),
		RuntimeStatus: dto.AdminRuntimeInfo{
			NumGoroutine: runtime.NumGoroutine(),
			MemAlloc:     m.Alloc,
			MemTotal:     m.TotalAlloc,
			MemSys:       m.Sys,
			HeapSys:      m.HeapSys,
			HeapIdle:     m.HeapIdle,
			HeapInuse:    m.HeapInuse,
			HeapReleased: m.HeapReleased,
			StackSys:     m.StackSys,
			MSpanSys:     m.MSpanSys,
			MCacheSys:    m.MCacheSys,
			BuckHashSys:  m.BuckHashSys,
			GCSys:        m.GCSys,
			OtherSys:     m.OtherSys,
			NextGC:       m.NextGC,
			NumGC:        m.NumGC,
		},
		CPU: dto.AdminCPUInfo{
			ModelName:     cpuModel,
			PhysicalCores: physCores,
			LogicalCores:  logicCores,
			Percent:       cpuPercents,
			LoadAvg: &dto.AdminLoadInfo{
				Load1:  loadStat.Load1,
				Load5:  loadStat.Load5,
				Load15: loadStat.Load15,
			},
		},
		Memory: dto.AdminMemoryInfo{
			Total:           vMem.Total,
			Available:       vMem.Available,
			Used:            vMem.Used,
			UsedPercent:     vMem.UsedPercent,
			SwapTotal:       swapMem.Total,
			SwapUsed:        swapMem.Used,
			SwapUsedPercent: swapMem.UsedPercent,
		},
		Host: dto.AdminHostInfo{
			Hostname:      hInfo.Hostname,
			OS:            hInfo.OS,
			OSPretty:      util.GetOSPrettyName(),
			Platform:      hInfo.Platform,
			Arch:          hInfo.KernelArch,
			KernelVersion: hInfo.KernelVersion,
			Uptime:        hInfo.Uptime,
			CurrentTime:   time.Now(),
			TimeZone:      time.Now().Location().String(),
			TimeZoneOffset: func() int {
				_, offset := time.Now().Zone()
				return offset
			}(),
		},
		Process: dto.AdminProcessInfo{
			PID:           p.Pid,
			PPID:          pPPid,
			Name:          pName,
			CPUPercent:    pCPU,
			MemoryPercent: pMem,
		},
	}

	response.ToResponse(code.Success.WithData(data))
}

// Upgrade triggers server automatic upgrade
// @Summary Trigger server upgrade
// @Description Download latest version and restart server
// @Tags System
// @Produce json
// @Security UserAuthToken
// @Param version query string true "Version to upgrade (e.g. 2.0.10 or latest)"
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/admin/upgrade [get]
func (h *AdminControlHandler) Upgrade(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	uid := pkgapp.GetUID(c)

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	var upgradeReq dto.UpgradeRequest
	if ok, validErrs := pkgapp.BindAndValid(c, &upgradeReq); !ok {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(validErrs.Errors()...))
		return
	}

	checkInfo := h.App.CheckVersion("")
	version := ""

	if upgradeReq.Version == "latest" {
		if !checkInfo.VersionIsNew {
			response.ToResponse(code.Success.WithDetails("Current version is already up to date"))
			return
		}
		version = checkInfo.VersionNewName
	} else {
		version = upgradeReq.Version
	}

	versionRaw := strings.TrimPrefix(version, "v")

	// Determine download URL
	// 确定下载地址
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Example: fast-note-sync-service-2.0.10-linux-amd64.tar.gz
	fileName := fmt.Sprintf("fast-note-sync-service-%s-%s-%s.tar.gz", versionRaw, goos, goarch)
	downloadURL := ""
	if checkInfo.GithubAvailable {
		// GitHub releases/download/[tag]/[filename]
		// Based on user feedback: URL should NOT have 'v' in the tag part if the tag itself doesn't have it
		downloadURL = fmt.Sprintf("https://github.com/haierkeys/fast-note-sync-service/releases/download/%s/%s", versionRaw, fileName)
	} else {
		// CNB download URL format
		downloadURL = fmt.Sprintf("https://cnb.cool/haierkeys/fast-note-sync-service/-/releases/download/%s/%s", versionRaw, fileName)
	}

	h.App.Logger().Info("Starting upgrade download", zap.String("url", downloadURL), zap.String("version", versionRaw))

	// Prepare temp directory
	// 使用 storage/temp/upgrade 作为临时目录
	tempDir := filepath.Join("storage", "temp", "upgrade")
	_ = os.RemoveAll(tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		response.ToResponse(code.Failed.WithDetails("Failed to create temp directory: " + err.Error()))
		return
	}

	// Download
	tarPath := filepath.Join(tempDir, fileName)
	if err := h.downloadFile(downloadURL, tarPath); err != nil {
		response.ToResponse(code.Failed.WithDetails("Download failed: " + err.Error()))
		return
	}

	// Extract
	binaryName := "fast-note-sync-service"
	if goos == "windows" {
		binaryName += ".exe"
	}
	extractedBinaryPath := filepath.Join(tempDir, binaryName)

	if err := h.extractBinary(tarPath, tempDir, binaryName); err != nil {
		response.ToResponse(code.Failed.WithDetails("Extract failed: " + err.Error()))
		return
	}

	// Trigger upgrade in App
	h.App.TriggerUpgrade(extractedBinaryPath)

	response.ToResponse(code.Success.WithDetails("Upgrade triggered, server is restarting..."))
}

// Restart triggers server automatic restart
// @Summary Trigger server restart
// @Description Gracefully restart the server
// @Tags System
// @Produce json
// @Security UserAuthToken
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/admin/restart [get]
func (h *AdminControlHandler) Restart(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	uid := pkgapp.GetUID(c)

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	currentBinary, err := os.Executable()
	if err != nil {
		response.ToResponse(code.Failed.WithDetails("Failed to get current executable path: " + err.Error()))
		return
	}

	h.App.TriggerUpgrade(currentBinary)

	response.ToResponse(code.Success.WithDetails("Restart triggered, server is restarting..."))
}

// GC triggers manual garbage collection and releases memory to OS (requires admin privileges)
// GC 手动触发垃圾回收并释放内存给操作系统（需要管理员权限）
// @Summary Trigger manual GC
// @Description Manually run Go runtime GC and release memory to OS, requires admin privileges
// @Tags System
// @Produce json
// @Security UserAuthToken
// @Success 200 {object} pkgapp.Res "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/gc [get]
func (h *AdminControlHandler) GC(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	logger := h.App.Logger()

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)

	startTime := time.Now()
	// Trigger GC // 触发 GC
	runtime.GC()
	// Release memory to OS // 释放内存给操作系统
	debug.FreeOSMemory()
	duration := time.Since(startTime)

	runtime.ReadMemStats(&mAfter)

	memReleased := int64(mBefore.Alloc) - int64(mAfter.Alloc)
	logger.Info("Manual GC completed",
		zap.Duration("duration", duration),
		zap.Uint64("allocBefore", mBefore.Alloc),
		zap.Uint64("allocAfter", mAfter.Alloc),
		zap.Int64("released", memReleased),
	)

	data := gin.H{
		"duration":    duration.String(),
		"allocBefore": mBefore.Alloc,
		"allocAfter":  mAfter.Alloc,
		"released":    memReleased,
	}

	response.ToResponse(code.Success.WithData(data).WithDetails("Manual GC completed successfully"))
}

// GetWSClients retrieves all currently connected WebSocket clients (requires admin privileges)
// @Summary Get connected WebSocket clients
// @Description Get a list of all current WebSocket connections, requires admin privileges
// @Tags System
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=[]pkgapp.WSClientInfo} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/ws_clients [get]
func (h *AdminControlHandler) GetWSClients(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	uid := pkgapp.GetUID(c)

	if uid == 0 {
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	clients := h.wss.GetClients()
	response.ToResponse(code.Success.WithData(clients))
}

func (h *AdminControlHandler) downloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (h *AdminControlHandler) extractBinary(tarPath string, destDir string, binaryName string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Check if it's the binary we're looking for
		// Often files in tar.gz are in a subdirectory or have different names
		// In alpha-release.yml: tar -czvf ... . (contents of build/platform dir)
		if filepath.Base(header.Name) == binaryName {
			target := filepath.Join(destDir, binaryName)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
			return nil
		}
	}

	return fmt.Errorf("binary %s not found in archive", binaryName)
}

// CloudflaredTunnelDownload triggers cloudflared binary download (requires admin privileges)
// @Summary Download cloudflared binary
// @Description Trigger the download of cloudflared binary for the current platform
// @Tags System
// @Security UserAuthToken
// @Produce json
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/admin/cloudflared_tunnel_download [get]
func (h *AdminControlHandler) CloudflaredTunnelDownload(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	uid := pkgapp.GetUID(c)

	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	h.App.Logger().Info("Starting manual cloudflared binary download via API")

	path, err := h.App.CloudflareService.DownloadBinary()
	if err != nil {
		h.App.Logger().Error("Manual cloudflared download failed", zap.Error(err))
		// 返回详细的错误提示（包含 DownloadBinary 中构造的建议）
		response.ToResponse(code.ErrorCloudflaredDownloadFailed.WithDetails(err.Error()))
		return
	}

	response.ToResponse(code.Success.WithData(gin.H{"path": path}).WithDetails("Cloudflared binary is ready"))
}
