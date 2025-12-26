package service

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"golang.org/x/sync/singleflight"

	"github.com/gin-gonic/gin"
)

type Service struct {
	ctx        *gin.Context
	dao        *dao.Dao
	SF         *singleflight.Group
	ClientName string
}

func New(ctx *gin.Context) *Service {

	svc := Service{ctx: ctx}
	svc.dao = dao.New(global.DBEngine, ctx)
	svc.SF = &singleflight.Group{}

	// svc.dao = dao.New(otgorm.WithContext(svc.ctx, global.DBEngine))

	return &svc
}

// NewBackground 创建一个用于后台任务 / 升级脚本 的 Service 实例 (ctx 为 nil)
func NewBackground(ctx context.Context) *Service {
	svc := Service{ctx: nil}
	svc.dao = dao.New(global.DBEngine, ctx)
	svc.SF = &singleflight.Group{}
	return &svc
}

func (svc *Service) WithClientName(clientName string) *Service {
	svc.ClientName = clientName
	return svc
}

func (svc *Service) WithSF(sf *singleflight.Group) *Service {
	svc.SF = sf
	return svc
}

func (svc *Service) Ctx() *gin.Context {
	return svc.ctx
}

// ExposeAutoMigrate 暴露自动迁移接口
func (svc *Service) ExposeAutoMigrate() error {

	//先迁移
	err := svc.dao.AutoMigrate(0, "User")
	if err != nil {
		return err
	}
	uids, err := svc.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}

	for _, uid := range uids {
		err := svc.dao.AutoMigrate(uid, "")
		if err != nil {
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}

// ExposeExecuteSQL 暴露执行 SQL 接口
func (svc *Service) ExposeExecuteSQL(sql string) error {
	uids, err := svc.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}
	for _, uid := range uids {
		// 忽略单个用户的清理错误，继续清理下一个
		db := svc.dao.Query(uid)
		db.Exec(sql)
	}
	return nil
}
