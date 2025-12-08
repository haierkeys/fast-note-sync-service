package upgrade

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
	"gorm.io/gorm"
)

// SchemaVersion 数据库版本记录表
type SchemaVersion struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Version     string    `gorm:"not null;uniqueIndex;type:varchar(64)" json:"version"`
	Description string    `gorm:"type:text" json:"description"`
	AppliedAt   time.Time `gorm:"not null" json:"applied_at"`
}

// TableName 指定表名
func (SchemaVersion) TableName() string {
	return "schema_version"
}

// Migration 定义升级接口
type Migration interface {
	Version() string
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

	// 获取已应用的数据库版本
	appliedVersions, err := m.getAppliedVersions()
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	refVer := m.getReferenceVersion()
	// 确保 reference version 有 "v" 前缀用于比较 (semver 库需要)
	if !strings.HasPrefix(refVer, "v") {
		refVer = "v" + refVer
	}

	if !semver.IsValid(refVer) {
		m.logger.Warn("reference version (from config/lastVersion) is not a valid semver, using v0.0.0", zap.String("version", refVer))
		refVer = "v0.0.0"
	}

	m.logger.Info("LastVersion", zap.String("version", refVer))

	// [NEW] 如果当前 global.Version 与 config/lastVersion 一致，则跳过后续检查
	// 这意味着在当前版本下已经运行过一次升级逻辑(无论是执行了还是跳过了)
	// 避免每次重启都进行不必要的数据库查询或日志输出
	currentGlobal := global.Version
	if !strings.HasPrefix(currentGlobal, "v") {
		currentGlobal = "v" + currentGlobal
	}
	if currentGlobal == refVer {
		m.logger.Info("skipping upgrade", zap.String("run_version", global.Version))
		return nil
	}

	// 执行所有未执行的升级
	executed := 0
	for _, migration := range m.migrations {
		migVer := migration.Version()

		// 检查是否已应用
		if appliedVersions[migVer] {
			continue
		}

		// 比较版本: 如果 migration.Version > refVer, 则跳过
		// 先标准化 format
		cmpVer := migVer
		if !strings.HasPrefix(cmpVer, "v") {
			cmpVer = "v" + cmpVer
		}

		if semver.IsValid(refVer) && semver.IsValid(cmpVer) {
			if semver.Compare(cmpVer, refVer) > 0 {
				m.logger.Info("skip migration > ref version",
					zap.String("mig", migVer),
					zap.String("ref", refVer))
				continue
			}
		}

		m.logger.Info("applying migration",
			zap.String("ver", migration.Version()),
			zap.String("desc", migration.Description()))

		// 在事务中执行升级
		if err := m.db.Transaction(func(tx *gorm.DB) error {
			// 执行升级脚本
			if err := migration.Up(tx, context.Background()); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			// 记录版本
			record := &SchemaVersion{
				Version:     migration.Version(),
				Description: migration.Description(),
				AppliedAt:   time.Now(),
			}
			if err := tx.Create(record).Error; err != nil {
				return fmt.Errorf("failed to record version: %w", err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version(), err)
		}

		m.logger.Info("migration applied successfully", zap.String("version", migration.Version()))
		executed++
	}

	if executed == 0 {
		m.logger.Info("database is already up to date")
	} else {
		m.logger.Info("upgrade completed", zap.Int("migrations_applied", executed))
	}

	// 无论是否执行了升级，最后将当前 global.Version 写入 config/lastVersion
	// 作为下一次运行的基准
	if err := m.saveReferenceVersion(global.Version); err != nil {
		m.logger.Error("save lastVersion failed", zap.Error(err))
		// 记录错误但不阻断启动
	} else {
		m.logger.Info("save lastVersion success", zap.String("ver", global.Version))
	}

	return nil
}

// getAppliedVersions 获取已应用的数据库版本
func (m *MigrationManager) getAppliedVersions() (map[string]bool, error) {
	var versions []SchemaVersion
	err := m.db.Find(&versions).Error
	if err != nil {
		return nil, err
	}

	applied := make(map[string]bool)
	for _, v := range versions {
		applied[v.Version] = true
		// Hack to support legacy integer version '1' mapping to '0.0.1' or protecting against re-run
		if v.Version == "1" {
			applied["0.0.1"] = true
		}
	}
	return applied, nil
}

// getReferenceVersion 获取参考版本号
// 从 config/lastVersion 读取，如果文件不存在或为空则返回 v0.0.0
func (m *MigrationManager) getReferenceVersion() string {
	content, err := os.ReadFile("config/lastVersion")
	if err != nil {
		if !os.IsNotExist(err) {
			m.logger.Warn("read config/lastVersion failed", zap.Error(err))
		} else {
			m.logger.Info("config/lastVersion not found, default v0.0.0")
		}
		return "v0.0.0"
	}

	ver := strings.TrimSpace(string(content))

	if ver == "" {
		m.logger.Info("config/lastVersion empty, default v0.0.0")
		return "v0.0.0"
	}
	return ver
}

// saveReferenceVersion 保存当前版本号到 config/lastVersion
func (m *MigrationManager) saveReferenceVersion(version string) error {
	return os.WriteFile("config/lastVersion", []byte(version), 0644)
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
