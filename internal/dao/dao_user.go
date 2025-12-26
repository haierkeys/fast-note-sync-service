package dao

import (
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
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

// user 获取用户查询对象
// 函数名: user
// 函数使用说明: 获取用户表的查询对象,内部方法。
// 返回值说明:
//   - *query.Query: 查询对象
func (d *Dao) user() *query.Query {
	return d.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "User")
	}, "user#user", "user")
}

// GetUserByUID 根据用户ID获取用户信息
// 函数名: GetUserByUID
// 函数使用说明: 根据用户ID查询未删除的用户信息。
// 参数说明:
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *User: 用户数据
//   - error: 出错时返回错误
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
// 函数名: GetUserByEmail
// 函数使用说明: 根据电子邮件地址查询未删除的用户信息。
// 参数说明:
//   - email string: 电子邮件地址
//
// 返回值说明:
//   - *User: 用户数据
//   - error: 出错时返回错误
func (d *Dao) GetUserByEmail(email string) (*User, error) {
	u := d.user().User
	m, err := u.WithContext(d.ctx).Where(u.Email.Eq(email), u.IsDeleted.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &User{}).(*User), nil
}

// GetUserByUsername 根据用户名获取用户信息
// 函数名: GetUserByUsername
// 函数使用说明: 根据用户名查询未删除的用户信息。
// 参数说明:
//   - username string: 用户名
//
// 返回值说明:
//   - *User: 用户数据
//   - error: 出错时返回错误
func (d *Dao) GetUserByUsername(username string) (*User, error) {
	u := d.user().User
	m, err := u.WithContext(d.ctx).Where(u.Username.Eq(username), u.IsDeleted.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &User{}).(*User), nil
}

// CreateUser 创建用户
// 函数名: CreateUser
// 函数使用说明: 在数据库中创建新的用户记录。
// 参数说明:
//   - dao *User: 用户数据
//
// 返回值说明:
//   - *User: 创建后的用户数据
//   - error: 出错时返回错误
func (d *Dao) CreateUser(dao *User) (*User, error) {
	m := convert.StructAssign(dao, &model.User{}).(*model.User)
	u := d.user().User
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &User{}).(*User), nil
}

// UserUpdatePassword 更新用户密码
// 函数名: UserUpdatePassword
// 函数使用说明: 根据用户ID更新用户密码,同时更新更新时间。
// 参数说明:
//   - password string: 新密码
//   - uid int64: 用户ID
//
// 返回值说明:
//   - error: 出错时返回错误
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

// GetAllUserUIDs 获取所有用户的UID
// 函数名: GetAllUserUIDs
// 函数使用说明: 查询所有未删除用户的UID列表。
// 返回值说明:
//   - []int64: 用户UID列表
//   - error: 出错时返回错误
func (d *Dao) GetAllUserUIDs() ([]int64, error) {
	var uids []int64
	u := d.user().User
	// 查询所有未删除的用户UID
	err := u.WithContext(d.ctx).Select(u.UID).Where(u.IsDeleted.Eq(0)).Scan(&uids)
	if err != nil {
		return nil, err
	}
	return uids, nil
}
