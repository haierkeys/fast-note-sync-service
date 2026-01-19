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

// WebGUIHandler WebGUI 配置 API 路由处理器
// 使用 App Container 注入依赖
type WebGUIHandler struct {
	*Handler
}

// NewWebGUIHandler 创建 WebGUIHandler 实例
func NewWebGUIHandler(a *app.App) *WebGUIHandler {
	return &WebGUIHandler{
		Handler: NewHandler(a),
	}
}

// webGUIConfig WebGUI 配置响应结构（公开接口）
type webGUIConfig struct {
	FontSet          string `json:"fontSet"`
	RegisterIsEnable bool   `json:"registerIsEnable"`
	AdminUID         int    `json:"adminUid"`
}

// webGUIAdminConfig WebGUI 管理员配置结构（管理员接口）
type webGUIAdminConfig struct {
	FontSet                 string `json:"fontSet" form:"fontSet"`
	RegisterIsEnable        bool   `json:"registerIsEnable" form:"registerIsEnable"`
	FileChunkSize           string `json:"fileChunkSize,omitempty" form:"fileChunkSize"`
	SoftDeleteRetentionTime string `json:"softDeleteRetentionTime,omitempty" form:"softDeleteRetentionTime"`
	UploadSessionTimeout    string `json:"uploadSessionTimeout,omitempty" form:"uploadSessionTimeout"`
	HistoryKeepVersions     int    `json:"historyKeepVersions,omitempty" form:"historyKeepVersions"`
	HistorySaveDelay        string `json:"historySaveDelay,omitempty" form:"historySaveDelay"`
	AdminUID                int    `json:"adminUid" form:"adminUid"`
	AuthTokenKey            string `json:"authTokenKey" form:"authTokenKey"`
	TokenExpiry             string `json:"tokenExpiry" form:"tokenExpiry"`
	ShareTokenKey           string `json:"shareTokenKey" form:"shareTokenKey"`
	ShareTokenExpiry        string `json:"shareTokenExpiry" form:"shareTokenExpiry"`
}

// Config 获取 WebGUI 配置（公开接口，无需认证）
// @Summary 获取 WebGUI 基础配置
// @Description 获取前端展示所需的非敏感配置，如字体设置、是否开放注册等
// @Tags 配置
// @Produce json
// @Success 200 {object} pkgapp.Res{data=webGUIConfig} "成功"
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

// GetConfig 获取管理员配置（需要管理员权限）
// @Summary 获取管理员全量配置
// @Description 获取系统底层的全量配置信息，需要管理员权限
// @Tags 配置
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=webGUIAdminConfig} "成功"
// @Failure 403 {object} pkgapp.Res "权限不足"
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
		AdminUID:                cfg.User.AdminUID,
		AuthTokenKey:            cfg.Security.AuthTokenKey,
		TokenExpiry:             cfg.Security.TokenExpiry,
		ShareTokenKey:           cfg.Security.ShareTokenKey,
		ShareTokenExpiry:        cfg.Security.ShareTokenExpiry,
	}

	response.ToResponse(code.Success.WithData(data))
}

// UpdateConfig 更新管理员配置（需要管理员权限）
// @Summary 更新管理员配置
// @Description 修改系统底层的全量配置信息，需要管理员权限
// @Tags 配置
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Accept json
// @Produce json
// @Param params body webGUIAdminConfig true "配置参数"
// @Success 200 {object} pkgapp.Res{data=webGUIAdminConfig} "成功"
// @Failure 403 {object} pkgapp.Res "权限不足"
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

	// 当配置了管理员 UID 且当前用户不是管理员时，拒绝访问
	if cfg.User.AdminUID != 0 && uid != int64(cfg.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	// 验证 historyKeepVersions 不能小于 100
	if params.HistoryKeepVersions > 0 && params.HistoryKeepVersions < 100 {
		logger.Warn("apiRouter.WebGUI.UpdateConfig invalid historyKeepVersions",
			zap.Int("value", params.HistoryKeepVersions))
		response.ToResponse(code.ErrorInvalidParams.WithDetails("historyKeepVersions must be at least 100"))
		return
	}

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

	// 更新配置
	cfg.WebGUI.FontSet = params.FontSet
	cfg.User.RegisterIsEnable = params.RegisterIsEnable
	cfg.App.FileChunkSize = params.FileChunkSize
	cfg.App.SoftDeleteRetentionTime = params.SoftDeleteRetentionTime
	cfg.App.UploadSessionTimeout = params.UploadSessionTimeout
	cfg.App.HistoryKeepVersions = params.HistoryKeepVersions
	cfg.App.HistorySaveDelay = params.HistorySaveDelay
	cfg.User.AdminUID = params.AdminUID
	cfg.Security.AuthTokenKey = params.AuthTokenKey
	cfg.Security.TokenExpiry = params.TokenExpiry
	cfg.Security.ShareTokenKey = params.ShareTokenKey
	cfg.Security.ShareTokenExpiry = params.ShareTokenExpiry

	// 保存配置到文件
	if err := cfg.Save(); err != nil {
		logger.Error("apiRouter.WebGUI.UpdateConfig.Save err", zap.Error(err))
		response.ToResponse(code.ErrorConfigSaveFailed)
		return
	}

	response.ToResponse(code.Success.WithData(params))
}
