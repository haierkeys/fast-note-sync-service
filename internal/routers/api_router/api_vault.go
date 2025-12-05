package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"go.uber.org/zap"
)

// Vault 仓库 API 路由处理器
// 结构体名: Vault
// 说明: 处理仓库（Vault）相关的 HTTP 请求，包括增删改查。
type Vault struct {
}

// NewVault 创建 Vault 路由处理器实例
// 函数名: NewVault
// 函数使用说明: 初始化并返回一个新的 Vault 结构体实例。
// 返回值说明:
//   - *Vault: 初始化后的 Vault 实例
func NewVault() *Vault {
	return &Vault{}
}

// CreateOrUpdate 创建或更新仓库
// 函数名: CreateOrUpdate
// 函数使用说明: 处理创建或更新仓库的 HTTP 请求。根据请求参数中的 ID 判断是创建还是更新操作。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含仓库信息 (id, name, etc.)
//
// 返回值说明:
//   - JSON: 操作结果和仓库数据
func (t *Vault) CreateOrUpdate(c *gin.Context) {
	params := &service.VaultPostRequest{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.Vault.Delete.BindAndValid err: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.Vault.Delete err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}
	svc := service.New(c)
	var vault *service.Vault
	var err error

	if params.ID > 0 {
		// Update logic
		vault, err = svc.VaultUpdate(params, uid)
		if err != nil {
			global.Logger.Error("apiRouter.Vault.CreateOrUpdate svc VaultUpdate err: %v", zap.Error(err))
			response.ToResponse(code.Failed.WithDetails(err.Error()))
			return
		}
		response.ToResponse(code.SuccessUpdate.WithData(vault))
	} else {
		// Create logic
		vault, err = svc.VaultCreate(params, uid)
		if err != nil {
			global.Logger.Error("apiRouter.Vault.CreateOrUpdate svc VaultCreate err: %v", zap.Error(err))
			response.ToResponse(code.Failed.WithDetails(err.Error()))
			return
		}
		response.ToResponse(code.SuccessCreate.WithData(vault))
	}
}

// Get 获取仓库详情
// 函数名: Get
// 函数使用说明: 处理获取单个仓库详情的 HTTP 请求。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含仓库 ID
//
// 返回值说明:
//   - JSON: 包含仓库详情的响应数据
func (t *Vault) Get(c *gin.Context) {
	params := &service.VaultGetRequest{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.Vault.CreateOrUpdate.BindAndValid errs: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	svc := service.New(c)
	vault, err := svc.VaultGet(params, uid)
	if err != nil {
		global.Logger.Error("apiRouter.Vault.Get svc VaultGet err: %v", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.Success.WithData(vault))
}

// List 获取仓库列表
// 函数名: List
// 函数使用说明: 处理获取当前用户所有仓库列表的 HTTP 请求。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含用户认证信息
//
// 返回值说明:
//   - JSON: 包含仓库列表的响应数据
func (t *Vault) List(c *gin.Context) {
	response := app.NewResponse(c)
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.Vault.List err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}
	svc := service.New(c)
	vaults, err := svc.VaultList(uid)
	if err != nil {
		global.Logger.Error("apiRouter.Vault.List svc VaultList err: %v", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.Success.WithData(vaults))
}

// Delete 删除仓库
// 函数名: Delete
// 函数使用说明: 处理删除仓库的 HTTP 请求。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含待删除仓库的 ID
//
// 返回值说明:
//   - JSON: 操作结果
func (t *Vault) Delete(c *gin.Context) {
	params := &service.VaultGetRequest{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.Vault.Delete.BindAndValid err: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.Vault.Delete err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}
	svc := service.New(c)
	err := svc.VaultDelete(params, uid)
	if err != nil {
		global.Logger.Error("apiRouter.Vault.Delete svc VaultDelete err: %v", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.SuccessDelete)
}
