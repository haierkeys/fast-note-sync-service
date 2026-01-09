// Package service 实现业务逻辑层
package service

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
)

// DBUtils 数据库工具服务，提供数据库迁移和 SQL 执行功能
// 用于后台任务和升级脚本
type DBUtils struct {
	dao *dao.Dao
}

// NewDBUtils 创建 DBUtils 实例
func NewDBUtils(ctx context.Context) *DBUtils {
	return &DBUtils{
		dao: dao.New(global.DBEngine, ctx),
	}
}

// ExposeAutoMigrate 暴露自动迁移接口
func (u *DBUtils) ExposeAutoMigrate() error {
	// 先迁移
	err := u.dao.AutoMigrate(0, "User")
	if err != nil {
		return err
	}
	uids, err := u.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}

	for _, uid := range uids {
		err = u.dao.AutoMigrate(uid, "")
		if err != nil {
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}

// ExecuteSQL 执行 SQL 接口
func (u *DBUtils) ExecuteSQL(sql string) error {
	db := u.dao.UseKey()
	if db != nil {
		db.Exec(sql)
	}
	return nil
}

// UserExecuteSQL 用户执行 SQL 接口
func (u *DBUtils) UserExecuteSQL(sql string) error {
	uids, err := u.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}
	for _, uid := range uids {
		// 忽略单个用户的清理错误，继续清理下一个
		db := u.dao.UserDB(uid)
		if db != nil {
			db.Exec(sql)
		}
	}
	return nil
}

// GetAllUserUIDs 获取所有用户的 UID
func (u *DBUtils) GetAllUserUIDs() ([]int64, error) {
	return u.dao.GetAllUserUIDs()
}
