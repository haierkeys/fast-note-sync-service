package upgrade

import (
	"context"
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SchemaVersion 数据库版本记录表
type SchemaVersion struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Version     int       `gorm:"not null;uniqueIndex" json:"version"`
	Description string    `gorm:"type:text" json:"description"`
	AppliedAt   time.Time `gorm:"not null" json:"applied_at"`
}

// TableName 指定表名
func (SchemaVersion) TableName() string {
	return "schema_version"
}

// Migration 定义升级接口
type Migration interface {
	Version() int
	Description() string
	Up(db *gorm.DB, ctx context.Context) error
}

// MigrationManager 升级管理器
type MigrationManager struct {
	db         *gorm.DB
	logger     *zap.Logger
	migrations []Migration
}

// NewMigrationManager 创建升级管理器
func NewMigrationManager(db *gorm.DB, logger *zap.Logger) *MigrationManager {
	return &MigrationManager{
		db:     db,
		logger: logger,
		migrations: []Migration{
			// 在这里注册所有的升级脚本
			&VaultMigrate{},
		},
	}
}

// Run 执行升级
func (m *MigrationManager) Run() error {
	// 确保 schema_version 表存在
	if err := m.db.AutoMigrate(&SchemaVersion{}); err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}

	// 获取当前数据库版本
	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	m.logger.Info("current database version", zap.Int("version", currentVersion))

	// 执行所有未执行的升级
	executed := 0
	for _, migration := range m.migrations {
		if migration.Version() <= currentVersion {
			//continue
		}

		m.logger.Info("applying migration",
			zap.Int("version", migration.Version()),
			zap.String("description", migration.Description()))

		// 在事务中执行升级
		if err := m.db.Transaction(func(tx *gorm.DB) error {
			// 执行升级脚本
			if err := migration.Up(tx, context.Background()); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			// 记录版本
			// record := &SchemaVersion{
			// 	Version:     migration.Version(),
			// 	Description: migration.Description(),
			// 	AppliedAt:   time.Now(),
			// }
			// if err := tx.Create(record).Error; err != nil {
			// 	return fmt.Errorf("failed to record version: %w", err)
			// }

			return nil
		}); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version(), err)
		}

		m.logger.Info("migration applied successfully", zap.Int("version", migration.Version()))
		executed++
	}

	if executed == 0 {
		m.logger.Info("database is already up to date")
	} else {
		m.logger.Info("upgrade completed", zap.Int("migrations_applied", executed))
	}

	return nil
}

// getCurrentVersion 获取当前数据库版本
func (m *MigrationManager) getCurrentVersion() (int, error) {
	var version SchemaVersion
	err := m.db.Order("version DESC").First(&version).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return version.Version, nil
}

// Execute 执行升级(便捷方法)
func Execute() error {
	if global.DBEngine == nil {
		return fmt.Errorf("database not initialized")
	}

	if global.Logger == nil {
		return fmt.Errorf("logger not initialized")
	}

	manager := NewMigrationManager(global.DBEngine, global.Logger)
	return manager.Run()
}
