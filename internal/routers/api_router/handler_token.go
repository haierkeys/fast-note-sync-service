package api_router

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"go.uber.org/zap"
)

// TokenHandler token API router handler
// TokenHandler 令牌 API 路由处理器
type TokenHandler struct {
	*Handler
}

// NewTokenHandler creates TokenHandler instance
// NewTokenHandler 创建 TokenHandler 实例
func NewTokenHandler(a *app.App) *TokenHandler {
	return &TokenHandler{
		Handler: NewHandler(a),
	}
}

// List active tokens
// List 列出活跃令牌
func (h *TokenHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	uid := pkgapp.GetUID(c)

	ctx := c.Request.Context()
	tokens, err := h.App.TokenService.ListByUser(ctx, uid)
	if err != nil {
		h.logError(ctx, "TokenHandler.List", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(tokens))
}

// Create manually issues a new token
// Create 手动签发新令牌
func (h *TokenHandler) Create(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.TokenIssueRequest{}

	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	ctx := c.Request.Context()

	res, err := h.App.TokenService.Create(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "TokenHandler.Create", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(res))
}

// UpdateScope updates a token's scope
// UpdateScope 更新令牌的权限范围
func (h *TokenHandler) UpdateScope(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	
	idStr := c.Param("id")
	tokenID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid id"))
		return
	}

	params := &dto.TokenUpdateRequest{}
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	ctx := c.Request.Context()

	err = h.App.TokenService.UpdateScope(ctx, uid, tokenID, params)
	if err != nil {
		h.logError(ctx, "TokenHandler.UpdateScope", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

// Revoke revokes a token
// Revoke 注销令牌
func (h *TokenHandler) Revoke(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	
	idStr := c.Param("id")
	tokenID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid id"))
		return
	}

	uid := pkgapp.GetUID(c)
	ctx := c.Request.Context()

	err = h.App.TokenService.Revoke(ctx, uid, tokenID)
	if err != nil {
		h.logError(ctx, "TokenHandler.Revoke", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

// ListLogs lists access logs for a specific token
// ListLogs 列出特定令牌的访问日志
func (h *TokenHandler) ListLogs(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	
	idStr := c.Param("id")
	tokenID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("invalid id"))
		return
	}

	params := &dto.TokenLogListRequest{}
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	uid := pkgapp.GetUID(c)
	ctx := c.Request.Context()

	logs, totalRows, err := h.App.TokenService.ListLogs(ctx, uid, tokenID, params.Page, params.PageSize)
	if err != nil {
		h.logError(ctx, "TokenHandler.ListLogs", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponseList(code.Success, logs, int(totalRows))
}

func (h *TokenHandler) logError(ctx context.Context, method string, err error) {
	h.App.Logger().Error(method, zap.Error(err))
}
