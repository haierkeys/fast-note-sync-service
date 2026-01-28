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

// VaultHandler vault API router handler
// VaultHandler 仓库 API 路由处理器
// Uses App Container to inject dependencies, supports unified error handling
// 使用 App Container 注入依赖，支持统一错误处理
type VaultHandler struct {
	*Handler
}

// NewVaultHandler creates VaultHandler instance
// NewVaultHandler 创建 VaultHandler 实例
func NewVaultHandler(a *app.App) *VaultHandler {
	return &VaultHandler{
		Handler: NewHandler(a),
	}
}

// CreateOrUpdate creates or updates a vault
// @Summary Create or update vault
// @Description Be used to create a new vault or update an existing vault configuration based on the ID in the request parameters
// @Tags Vault
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.VaultPostRequest true "Vault Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.VaultDTO} "Success"
// @Router /api/vault [post]
func (h *VaultHandler) CreateOrUpdate(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.VaultPostRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("VaultHandler.CreateOrUpdate.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("VaultHandler.CreateOrUpdate err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// Get request context
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

// Get retrieves vault details
// @Summary Get vault details
// @Description Get specific vault configuration details by vault ID
// @Tags Vault
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param id query int64 true "Vault ID"
// @Success 200 {object} pkgapp.Res{data=dto.VaultDTO} "Success"
// @Router /api/vault/get [get]
func (h *VaultHandler) Get(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.VaultGetRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("VaultHandler.Get.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)

	// Get request context
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

// List retrieves vault list
// @Summary Get vault list
// @Description Get all note vaults for current user
// @Tags Vault
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=[]dto.VaultDTO} "Success"
// @Router /api/vault [get]
func (h *VaultHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("VaultHandler.List err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// Get request context
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

// Delete deletes a vault
// @Summary Delete vault
// @Description Permanently delete a specific note vault and all associated notes and attachments
// @Tags Vault
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.VaultGetRequest true "Delete Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/vault [delete]
func (h *VaultHandler) Delete(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.VaultGetRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("VaultHandler.Delete.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("VaultHandler.Delete err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// Get request context
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

// logError records error log, including Trace ID
// logError 记录错误日志，包含 Trace ID
func (h *VaultHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
