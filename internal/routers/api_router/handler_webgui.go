package api_router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

// WebGUIHandler WebGUI configuration API router handler
// WebGUIHandler WebGUI 配置 API 路由处理器
// Uses App Container to inject dependencies
// 使用 App Container 注入依赖
type WebGUIHandler struct {
	*Handler
}

// NewWebGUIHandler creates WebGUIHandler instance
// NewWebGUIHandler 创建 WebGUIHandler 实例
func NewWebGUIHandler(a *app.App) *WebGUIHandler {
	return &WebGUIHandler{
		Handler: NewHandler(a),
	}
}

// webGUIConfig WebGUI configuration response structure (public interface)
// webGUIConfig WebGUI 配置响应结构（公开接口）
type webGUIConfig struct {
	FontSet          string `json:"fontSet"`          // Font set // 字体设置
	RegisterIsEnable bool   `json:"registerIsEnable"` // Registration enablement // 是否开启注册
	AdminUID         int    `json:"adminUid"`         // Admin UID // 管理员 UID
}

// webGUIAdminConfig WebGUI administrator configuration structure (admin interface)
// webGUIAdminConfig WebGUI 管理员配置结构（管理员接口）
type webGUIAdminConfig struct {
	FontSet                 string `json:"fontSet" form:"fontSet"`                                           // Font set // 字体设置
	RegisterIsEnable        bool   `json:"registerIsEnable" form:"registerIsEnable"`                         // Registration enablement // 是否开启注册
	FileChunkSize           string `json:"fileChunkSize,omitempty" form:"fileChunkSize"`                     // File chunk size // 文件分块大小
	SoftDeleteRetentionTime string `json:"softDeleteRetentionTime,omitempty" form:"softDeleteRetentionTime"` // Soft delete retention time // 软删除保留时间
	UploadSessionTimeout    string `json:"uploadSessionTimeout,omitempty" form:"uploadSessionTimeout"`       // Upload session timeout // 上传会话超时时间
	HistoryKeepVersions     int    `json:"historyKeepVersions,omitempty" form:"historyKeepVersions"`         // History versions to keep // 历史版本保留数
	HistorySaveDelay        string `json:"historySaveDelay,omitempty" form:"historySaveDelay"`               // History save delay // 历史保存延迟
	DefaultAPIFolder        string `json:"defaultApiFolder,omitempty" form:"defaultApiFolder"`               // Default API folder // 默认 API 目录
	AdminUID                int    `json:"adminUid" form:"adminUid"`                                         // Admin UID // 管理员 UID
	AuthTokenKey            string `json:"authTokenKey" form:"authTokenKey"`                                 // Auth token key // 认证 Token 密钥
	TokenExpiry             string `json:"tokenExpiry" form:"tokenExpiry"`                                   // Token expiry // Token 有效期
	ShareTokenKey           string `json:"shareTokenKey" form:"shareTokenKey"`                               // Share token key // 分享 Token 密钥
	ShareTokenExpiry        string `json:"shareTokenExpiry" form:"shareTokenExpiry"`                         // Share token expiry // 分享 Token 有效期
}

// Config retrieves WebGUI configuration (public interface)
// @Summary Get WebGUI basic config
// @Description Get non-sensitive configuration required for frontend display, such as font settings, registration status, etc.
// @Tags Config
// @Produce json
// @Success 200 {object} pkgapp.Res{data=webGUIConfig} "Success"
// @Router /api/webgui/config [get]
func (h *WebGUIHandler) Config(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	data := webGUIConfig{
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
// @Success 200 {object} pkgapp.Res{data=webGUIAdminConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config [get]
func (h *WebGUIHandler) GetConfig(c *gin.Context) {
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

	data := &webGUIAdminConfig{
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
// @Param params body webGUIAdminConfig true "Config Parameters"
// @Success 200 {object} pkgapp.Res{data=webGUIAdminConfig} "Success"
// @Failure 403 {object} pkgapp.Res "Insufficient privileges"
// @Router /api/admin/config [post]
func (h *WebGUIHandler) UpdateConfig(c *gin.Context) {
	params := &webGUIAdminConfig{}
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

	// Save configuration to file
	// 保存配置到文件
	if err := cfg.Save(); err != nil {
		logger.Error("apiRouter.WebGUI.UpdateConfig.Save err", zap.Error(err))
		response.ToResponse(code.ErrorConfigSaveFailed)
		return
	}

	response.ToResponse(code.Success.WithData(params))
}
