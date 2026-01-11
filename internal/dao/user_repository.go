// Package dao 实现数据访问层
package dao

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// userRepository 实现 domain.UserRepository 接口
type userRepository struct {
	dao *Dao
}

// NewUserRepository 创建 UserRepository 实例
func NewUserRepository(dao *Dao) domain.UserRepository {
	return &userRepository{dao: dao}
}

func (r *userRepository) GetKey(uid int64) string {
	return ""
}

// user 获取用户查询对象
func (r *userRepository) user() *query.Query {
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "User")
	}, "user#user")
}

// toDomain 将数据库模型转换为领域模型
func (r *userRepository) toDomain(m *model.User) *domain.User {
	if m == nil {
		return nil
	}
	return &domain.User{
		UID:       m.UID,
		Email:     m.Email,
		Username:  m.Username,
		Password:  m.Password,
		Salt:      m.Salt,
		Token:     m.Token,
		Avatar:    m.Avatar,
		IsDeleted: m.IsDeleted == 1,
		CreatedAt: time.Time(m.CreatedAt),
		UpdatedAt: time.Time(m.UpdatedAt),
		DeletedAt: time.Time(m.DeletedAt),
	}
}

// toModel 将领域模型转换为数据库模型
func (r *userRepository) toModel(user *domain.User) *model.User {
	if user == nil {
		return nil
	}
	isDeleted := int64(0)
	if user.IsDeleted {
		isDeleted = 1
	}
	return &model.User{
		UID:       user.UID,
		Email:     user.Email,
		Username:  user.Username,
		Password:  user.Password,
		Salt:      user.Salt,
		Token:     user.Token,
		Avatar:    user.Avatar,
		IsDeleted: isDeleted,
		CreatedAt: timex.Time(user.CreatedAt),
		UpdatedAt: timex.Time(user.UpdatedAt),
		DeletedAt: timex.Time(user.DeletedAt),
	}
}

// GetByUID 根据UID获取用户
func (r *userRepository) GetByUID(ctx context.Context, uid int64) (*domain.User, error) {
	u := r.user().User
	m, err := u.WithContext(ctx).Where(u.UID.Eq(uid), u.IsDeleted.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := r.user().User
	m, err := u.WithContext(ctx).Where(u.Email.Eq(email), u.IsDeleted.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	u := r.user().User
	m, err := u.WithContext(ctx).Where(u.Username.Eq(username), u.IsDeleted.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	u := r.user().User
	m := r.toModel(user)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()

	err := u.WithContext(ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

// UpdatePassword 更新用户密码
func (r *userRepository) UpdatePassword(ctx context.Context, password string, uid int64) error {
	u := r.user().User

	_, err := u.WithContext(ctx).Where(
		u.UID.Eq(uid),
	).UpdateSimple(
		u.Password.Value(password),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

// GetAllUIDs 获取所有用户UID
func (r *userRepository) GetAllUIDs(ctx context.Context) ([]int64, error) {
	var uids []int64
	u := r.user().User
	err := u.WithContext(ctx).Select(u.UID).Where(u.IsDeleted.Eq(0)).Scan(&uids)
	if err != nil {
		return nil, err
	}
	return uids, nil
}

// 确保 userRepository 实现了 domain.UserRepository 接口
var _ domain.UserRepository = (*userRepository)(nil)
