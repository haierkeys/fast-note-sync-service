// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/dao"

	"gorm.io/gorm"
)

// DBUtils database utility service, providing database migration and SQL execution functions
// Used for background tasks and upgrade scripts
// DBUtils 数据库工具服务，提供数据库迁移和 SQL 执行功能
// 用于后台任务和升级脚本
type DBUtils struct {
	dao *dao.Dao
}

// NewDBUtils creates DBUtils instance
// db: Database connection (required)
// ctx: Context
// opts: Dao configuration options
// NewDBUtils 创建 DBUtils 实例
// db: 数据库连接（必须）
// ctx: 上下文
// opts: Dao 配置选项
func NewDBUtils(db *gorm.DB, ctx context.Context, opts ...dao.DaoOption) *DBUtils {
	return &DBUtils{
		dao: dao.New(db, ctx, opts...),
	}
}

// ExposeAutoMigrate exposes automatic migration interface
// ExposeAutoMigrate 暴露自动迁移接口
func (u *DBUtils) ExposeAutoMigrate() error {
	// Migrate user table first
	// 先迁移 User 表
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

// ExecuteSQL executes SQL interface
// ExecuteSQL 执行 SQL 接口
func (u *DBUtils) ExecuteSQL(sql string) error {
	db := u.dao.UseKey()
	if db != nil {
		db.Exec(sql)
	}
	return nil
}

// UserExecuteSQL user executes SQL interface
// UserExecuteSQL 用户执行 SQL 接口
func (u *DBUtils) UserExecuteSQL(sql string) error {
	uids, err := u.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}
	for _, uid := range uids {
		// Ignore cleanup errors for individual users, continue with the next
		// 忽略单个用户的清理错误，继续清理下一个
		db := u.dao.UserDB(uid)
		if db != nil {
			db.Exec(sql)
		}
	}
	return nil
}

// GetAllUserUIDs retrieves all user UIDs
// GetAllUserUIDs 获取所有用户的 UID
func (u *DBUtils) GetAllUserUIDs() ([]int64, error) {
	return u.dao.GetAllUserUIDs()
}
