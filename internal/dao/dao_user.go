package dao

import (
	"github.com/gookit/goutil/dump"
	"github.com/haierkeys/obsidian-better-sync-service/internal/model"
	"github.com/haierkeys/obsidian-better-sync-service/internal/query"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/convert"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/timex"
)

type User struct {
	Uid       int64      `gorm:"column:uid;AUTO_INCREMENT" json:"uid" form:"uid"`                       //
	Email     string     `gorm:"column:email;default:''" json:"email" form:"email"`                     //
	Username  string     `gorm:"column:username;default:''" json:"username" form:"username"`            //
	Password  string     `gorm:"column:password;default:''" json:"password" form:"password"`            //
	Salt      string     `gorm:"column:salt;default:''" json:"salt" form:"salt"`                        //
	Token     string     `gorm:"column:token;default:''" json:"token" form:"token"`                     //
	Avatar    string     `gorm:"column:avatar;default:''" json:"avatar" form:"avatar"`                  //
	IsDeleted int64      `gorm:"column:is_deleted;default:0" json:"isDeleted" form:"isDeleted"`         //
	UpdatedAt timex.Time `gorm:"column:updated_at;time;default:NULL" json:"updatedAt" form:"updatedAt"` //
	CreatedAt timex.Time `gorm:"column:created_at;time;default:NULL" json:"createdAt" form:"createdAt"` //
	DeletedAt timex.Time `gorm:"column:deleted_at;time;default:NULL" json:"deletedAt" form:"deletedAt"` //
}

// GetUserByUID 根据用户ID获取用户信息
func (d *Dao) GetUserByUID(uid int64) (*User, error) {
	u := query.Use(d.Db).User
	//x, _ := d.Use2("User").Where(d.User.UID.Eq(10)).First()

	x, _ := u.WithContext(d.ctx).Where(u.UID.Eq(10)).First()

	dump.P(x)

	// 使用 user_repo 构建查询，查找 UID 等于给定 uid 的用户，并且未被删除
	m, err := user_repo.NewQueryBuilder(d.Db).
		WhereUid(model.Eq, uid).
		WhereIsDeleted(model.Eq, 0).
		First()
	// 如果发生错误，返回 nil 和错误
	if err != nil {
		return nil, err
	}
	// 将查询结果转换为 User 结构体，并返回
	return convert.StructAssign(m, &User{}).(*User), nil
}

// GetUserByEmail 根据电子邮件获取用户信息
func (d *Dao) GetUserByEmail(email string) (*User, error) {

	m, err := user_repo.NewQueryBuilder(d.Db).
		WhereEmail(model.Eq, email).
		WhereIsDeleted(model.Eq, 0).
		First()

	if err != nil {
		return nil, err
	}

	return convert.StructAssign(m, &User{}).(*User), nil

}

func (d *Dao) GetUserByUsername(username string) (*User, error) {

	m, err := user_repo.NewQueryBuilder(d.Db).
		WhereUsername(model.Eq, username).
		WhereIsDeleted(model.Eq, 0).
		First()

	if err != nil {
		return nil, err
	}

	return convert.StructAssign(m, &User{}).(*User), nil

}

// CreateMember 创建用户
func (d *Dao) CreateMember(dao *User) (int64, error) { // 修改参数类型为 User

	m := convert.StructAssign(dao, &user_repo.User{}).(*user_repo.User)

	id, err := m.Create(d.Db)

	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateUser 创建用户
func (d *Dao) CreateUser(dao *User) (int64, error) { // 修改函数名为 CreateUser

	m := convert.StructAssign(dao, user_repo.NewModel()).(*user_repo.User)

	id, err := m.Create(d.Db)

	if err != nil {
		return 0, err
	}
	return id, nil
}
