package service

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
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
	svc.SF = &singleflight.Group{}

	// svc.dao = dao.New(otgorm.WithContext(svc.ctx, global.DBEngine))

	return &svc
}

// NewBackground 创建一个用于后台任务的 Service 实例 (ctx 为 nil)
func NewBackground(ctx context.Context) *Service {
	svc := Service{ctx: nil}
	svc.dao = dao.New(global.DBEngine, ctx)
	svc.SF = &singleflight.Group{}
	return &svc
}

func (svc *Service) WithSF(sf *singleflight.Group) *Service {
	svc.SF = sf
	return svc
}

func (svc *Service) Ctx() *gin.Context {
	return svc.ctx
}
