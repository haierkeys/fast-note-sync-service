package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameUserOIDCIdentity = "user_oidc_identity"

// UserOIDCIdentity stores an external OIDC subject binding.
type UserOIDCIdentity struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id" form:"id"`
	UID       int64      `gorm:"column:uid;index:idx_user_oidc_identity_uid,priority:1;not null" json:"uid" form:"uid"`
	Issuer    string     `gorm:"column:issuer;type:varchar(512);uniqueIndex:idx_user_oidc_identity_issuer_subject,priority:1;not null" json:"issuer" form:"issuer"`
	Subject   string     `gorm:"column:subject;type:varchar(512);uniqueIndex:idx_user_oidc_identity_issuer_subject,priority:2;not null" json:"subject" form:"subject"`
	Email     string     `gorm:"column:email;type:varchar(255);default:''" json:"email" form:"email"`
	Username  string     `gorm:"column:username;type:varchar(255);default:''" json:"username" form:"username"`
	CreatedAt timex.Time `gorm:"column:created_at;default:NULL;autoCreateTime:false" json:"createdAt" form:"createdAt"`
	UpdatedAt timex.Time `gorm:"column:updated_at;default:NULL;autoUpdateTime:false" json:"updatedAt" form:"updatedAt"`
}

func (*UserOIDCIdentity) TableName() string {
	return TableNameUserOIDCIdentity
}
