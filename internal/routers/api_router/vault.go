package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/internal/service"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/app"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/code"

	"go.uber.org/zap"
)

type Vault struct {
}

func NewVault() *Vault {
	return &Vault{}
}

// CreateOrUpdate handles both creating and updating a vault
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

// Get retrieves a vault by ID
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

// List retrieves all vaults for a user
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
