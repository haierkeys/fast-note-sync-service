package service

import (
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/internal/dao"

	"github.com/gin-gonic/gin"
)

type Service struct {
	ctx *gin.Context
	dao *dao.Dao
}

func New(ctx *gin.Context) Service {

	svc := Service{ctx: ctx}
	svc.dao = dao.New(global.DBEngine, ctx)

	// svc.dao = dao.New(otgorm.WithContext(svc.ctx, global.DBEngine))

	return svc
}

func (svc *Service) Ctx() *gin.Context {
	return svc.ctx
}
