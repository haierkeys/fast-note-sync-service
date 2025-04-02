package service

import (
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/internal/dao"
	"golang.org/x/sync/singleflight"

	"github.com/gin-gonic/gin"
)

type Service struct {
	ctx *gin.Context
	dao *dao.Dao
	SF  *singleflight.Group
}

func New(ctx *gin.Context) *Service {

	svc := Service{ctx: ctx}
	svc.dao = dao.New(global.DBEngine, ctx)

	// svc.dao = dao.New(otgorm.WithContext(svc.ctx, global.DBEngine))

	return &svc
}

func (svc *Service) WithSF(sf *singleflight.Group) *Service {
	svc.SF = sf
	return svc
}

func (svc *Service) Ctx() *gin.Context {
	return svc.ctx
}
