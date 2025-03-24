package user_repo

import "github.com/haierkeys/obsidian-better-sync-service/pkg/timex"

//
//go:generate gormgen -structs User -input . -pre pre_
type User struct {
	Uid       int64      `gorm:"column:uid;primary_key;auto_increment" json:"uid" form:"uid"`                                         //
	Email     string     `gorm:"column:email;index;default:''" json:"email" form:"email"`                                             //
	Username  string     `gorm:"column:username;default:''" json:"username" form:"username"`                                          //
	Password  string     `gorm:"column:password;default:''" json:"password" form:"password"`                                          //
	Salt      string     `gorm:"column:salt;default:''" json:"salt" form:"salt"`                                                      //
	Token     string     `gorm:"column:token;default:''" json:"token" form:"token"`                                                   //
	Avatar    string     `gorm:"column:avatar;default:''" json:"avatar" form:"avatar"`                                                //
	IsDeleted int64      `gorm:"column:is_deleted;default:0" json:"isDeleted" form:"isDeleted"`                                       //
	UpdatedAt timex.Time `gorm:"column:updated_at;type:datetime;autoUpdateTime:false;default:NULL" json:"updatedAt" form:"updatedAt"` //
	CreatedAt timex.Time `gorm:"column:created_at;type:datetime;autoUpdateTime:false;default:NULL" json:"createdAt" form:"createdAt"` //
	DeletedAt timex.Time `gorm:"column:deleted_at;type:datetime;autoUpdateTime:false;default:NULL" json:"deletedAt" form:"deletedAt"` //
}
