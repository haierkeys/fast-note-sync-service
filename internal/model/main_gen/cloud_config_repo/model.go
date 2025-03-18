package cloud_config_repo

import "github.com/haierkeys/obsidian-better-sync-service/pkg/timex"

//
//go:generate gormgen -structs CloudConfig -input . -pre pre_
type CloudConfig struct {
	Id              int64      `gorm:"column:id;primary_key;auto_increment" json:"id" form:"id"`                                            //
	Uid             int64      `gorm:"column:uid;index;default:0" json:"uid" form:"uid"`                                                    //
	Type            string     `gorm:"column:type;default:''" json:"type" form:"type"`                                                      //
	Endpoint        string     `gorm:"column:endpoint;default:''" json:"endpoint" form:"endpoint"`                                          //
	Region          string     `gorm:"column:region;default:''" json:"region" form:"region"`                                                //
	AccountId       string     `gorm:"column:account_id;default:''" json:"accountId" form:"accountId"`                                      //
	BucketName      string     `gorm:"column:bucket_name;default:''" json:"bucketName" form:"bucketName"`                                   //
	AccessKeyId     string     `gorm:"column:access_key_id;default:''" json:"accessKeyId" form:"accessKeyId"`                               //
	AccessKeySecret string     `gorm:"column:access_key_secret;default:''" json:"accessKeySecret" form:"accessKeySecret"`                   //
	CustomPath      string     `gorm:"column:custom_path;default:''" json:"customPath" form:"customPath"`                                   //
	AccessUrlPrefix string     `gorm:"column:access_url_prefix;default:''" json:"accessUrlPrefix" form:"accessUrlPrefix"`                   //
	IsEnabled       int64      `gorm:"column:is_enabled;default:1" json:"isEnabled" form:"isEnabled"`                                       //
	IsDeleted       int64      `gorm:"column:is_deleted;default:0" json:"isDeleted" form:"isDeleted"`                                       //
	UpdatedAt       timex.Time `gorm:"column:updated_at;type:datetime;autoUpdateTime:false;default:NULL" json:"updatedAt" form:"updatedAt"` //
	CreatedAt       timex.Time `gorm:"column:created_at;type:datetime;autoUpdateTime:false;default:NULL" json:"createdAt" form:"createdAt"` //
	DeletedAt       timex.Time `gorm:"column:deleted_at;type:datetime;autoUpdateTime:false;default:NULL" json:"deletedAt" form:"deletedAt"` //
}
