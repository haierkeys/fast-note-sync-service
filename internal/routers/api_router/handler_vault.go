package api_router

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"go.uber.org/zap"
)

// VaultHandler 仓库 API 路由处理器
// 使用 App Container 注入依赖，支持统一错误处理
type VaultHandler struct {
	*Handler
}

// NewVaultHandler 创建 VaultHandler 实例
func NewVaultHandler(a *app.App) *VaultHandler {
	return &VaultHandler{
		Handler: NewHandler(a),
	}
}

// CreateOrUpdate 创建或更新仓库
// @Summary 创建或更新仓库
// @Description 根据请求参数中的 ID 判断是创建新仓库还是更新已有仓库配置
// @Tags 仓库
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Accept json
// @Produce json
// @Param params body dto.VaultPostRequest true "仓库参数"
// @Success 200 {object} pkgapp.Res{data=dto.VaultDTO} "成功"
// @Router /api/vault [post]
func (h *VaultHandler) CreateOrUpdate(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.VaultPostRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("VaultHandler.CreateOrUpdate.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("VaultHandler.CreateOrUpdate err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	var vault *dto.VaultDTO
	var err error

	if params.ID > 0 {
		// Update logic
		vault, err = h.App.VaultService.Update(ctx, uid, params.ID, params.Vault)
		if err != nil {
			h.logError(ctx, "VaultHandler.CreateOrUpdate.Update", err)
			apperrors.ErrorResponse(c, err)
			return
		}
		response.ToResponse(code.SuccessUpdate.WithData(vault))
	} else {
		// Create logic
		vault, err = h.App.VaultService.Create(ctx, uid, params.Vault)
		if err != nil {
			h.logError(ctx, "VaultHandler.CreateOrUpdate.Create", err)
			apperrors.ErrorResponse(c, err)
			return
		}
		response.ToResponse(code.SuccessCreate.WithData(vault))
	}
}

// Get 获取仓库详情
// @Summary 获取仓库详情
// @Description 根据仓库 ID 获取特定仓库的配置详情
// @Tags 仓库
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Param id query int64 true "仓库 ID"
// @Success 200 {object} pkgapp.Res{data=dto.VaultDTO} "成功"
// @Router /api/vault/get [get]
func (h *VaultHandler) Get(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.VaultGetRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("VaultHandler.Get.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)

	// 获取请求上下文
	ctx := c.Request.Context()

	vault, err := h.App.VaultService.Get(ctx, uid, params.ID)
	if err != nil {
		h.logError(ctx, "VaultHandler.Get", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(vault))
}

// List 获取仓库列表
// @Summary 获取仓库列表
// @Description 获取当前用户所有的笔记仓库清单
// @Tags 仓库
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=[]dto.VaultDTO} "成功"
// @Router /api/vault [get]
func (h *VaultHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("VaultHandler.List err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	vaults, err := h.App.VaultService.List(ctx, uid)
	if err != nil {
		h.logError(ctx, "VaultHandler.List", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(vaults))
}

// Delete 删除仓库
// @Summary 删除仓库
// @Description 永久删除指定的笔记仓库及其关联的所有笔记和附件
// @Tags 仓库
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Param params query dto.VaultGetRequest true "删除参数"
// @Success 200 {object} pkgapp.Res "成功"
// @Router /api/vault [delete]
func (h *VaultHandler) Delete(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.VaultGetRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("VaultHandler.Delete.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("VaultHandler.Delete err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	err := h.App.VaultService.Delete(ctx, uid, params.ID)
	if err != nil {
		h.logError(ctx, "VaultHandler.Delete", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.SuccessDelete)
}

// logError 记录错误日志，包含 Trace ID
func (h *VaultHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
