package dao

import (
	"github.com/haierkeys/obsidian-better-sync-service/internal/model"
	"github.com/haierkeys/obsidian-better-sync-service/internal/query"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/convert"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type User struct {
	UID       int64      `gorm:"column:uid;primaryKey" json:"uid" type:"uid" form:"uid"`
	Email     string     `gorm:"column:email" json:"email" type:"email" form:"email"`
	Username  string     `gorm:"column:username" json:"username" type:"username" form:"username"`
	Password  string     `gorm:"column:password" json:"password" type:"password" form:"password"`
	Salt      string     `gorm:"column:salt" json:"salt" type:"salt" form:"salt"`
	Token     string     `gorm:"column:token" json:"token" type:"token" form:"token"`
	Avatar    string     `gorm:"column:avatar" json:"avatar" type:"avatar" form:"avatar"`
	IsDeleted int64      `gorm:"column:is_deleted" json:"isDeleted" type:"isDeleted" form:"isDeleted"`
	UpdatedAt timex.Time `gorm:"column:updated_at;type:datetime;autoUpdateTime:false" json:"updatedAt" type:"updatedAt" form:"updatedAt"`
	CreatedAt timex.Time `gorm:"column:created_at;type:datetime;autoUpdateTime:false" json:"createdAt" type:"createdAt" form:"createdAt"`
	DeletedAt timex.Time `gorm:"column:deleted_at;type:datetime;autoUpdateTime:false" json:"deletedAt" type:"deletedAt" form:"deletedAt"`
}

func (d *Dao) user() *query.Query {
	return d.Use(
		func(g *gorm.DB) {
			model.AutoMigrate(g, "User")
		},
	)
}

// GetUserByUID 根据用户ID获取用户信息
func (d *Dao) GetUserByUID(uid int64) (*User, error) {
	u := d.user().User
	m, err := u.WithContext(d.ctx).Where(u.UID.Eq(uid), u.IsDeleted.Eq(0)).First()
	// 如果发生错误，返回 nil 和错误
	if err != nil {
		return nil, err
	}
	// 将查询结果转换为 User 结构体，并返回
	return convert.StructAssign(m, &User{}).(*User), nil
}

// GetUserByEmail 根据电子邮件获取用户信息
func (d *Dao) GetUserByEmail(email string) (*User, error) {
	u := d.user().User
	m, err := u.WithContext(d.ctx).Where(u.Email.Eq(email), u.IsDeleted.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &User{}).(*User), nil
}

// GetUserByUsername 根据用户名获取用户信息

func (d *Dao) GetUserByUsername(username string) (*User, error) {
	u := d.user().User
	m, err := u.WithContext(d.ctx).Where(u.Username.Eq(username), u.IsDeleted.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &User{}).(*User), nil
}

// CreateUser 创建用户
func (d *Dao) CreateUser(dao *User) (*User, error) { // 修改函数名为 CreateUser
	m := convert.StructAssign(dao, &model.User{}).(*model.User)
	u := d.user().User
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &User{}).(*User), nil
}

// UpdateUser 更新用户密码
func (d *Dao) UserUpdatePassword(password string, uid int64) error {
	u := d.user().User

	_, err := u.WithContext(d.ctx).Where(
		u.UID.Eq(uid),
	).UpdateSimple(
		u.Password.Value(password),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}
