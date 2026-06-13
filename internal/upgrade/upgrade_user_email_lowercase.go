package upgrade

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// UserEmailLowercaseMigrate converts all user emails in the database to lowercase
// UserEmailLowercaseMigrate 将数据库中所有的用户邮箱转换为小写
type UserEmailLowercaseMigrate struct{}

// Version returns the migration version
// Version 返回升级版本号
func (m *UserEmailLowercaseMigrate) Version() string {
	return "3.3.2"
}

// Description returns the migration description
// Description 返回升级描述
func (m *UserEmailLowercaseMigrate) Description() string {
	return "Convert all existing user emails to lowercase"
}

// Up runs the migration
// Up 执行升级操作
func (m *UserEmailLowercaseMigrate) Up(db *gorm.DB, ctx context.Context, mc *MigrationContext) error {
	err := db.WithContext(ctx).Table("user").Where("email != LOWER(email)").Update("email", gorm.Expr("LOWER(email)")).Error
	if err != nil {
		return fmt.Errorf("failed to lowercase user emails: %w", err)
	}
	return nil
}
