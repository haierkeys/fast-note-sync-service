package dao

import (
	"context"
	"strconv"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Setting struct {
	ID               int64      `json:"id" form:"id"`                             // ID
	VaultID          int64      `json:"vaultId" form:"vaultId"`                   // 保险库ID
	Action           string     `json:"action" form:"action"`                     // 操作
	Path             string     `json:"path" form:"path"`                         // 路径
	PathHash         string     `json:"pathHash" form:"pathHash"`                 // 路径哈希
	Content          string     `json:"content" form:"content"`                   // 内容
	ContentHash      string     `json:"contentHash" form:"contentHash"`           // 内容哈希
	Size             int64      `json:"size" form:"size"`                         // 大小
	Ctime            int64      `json:"ctime" form:"ctime"`                       // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`                       // 修改时间戳
	UpdatedTimestamp int64      `json:"updatedTimestamp" form:"updatedTimestamp"` // 更新时间戳
	CreatedAt        timex.Time `json:"createdAt" form:"createdAt"`               // 创建时间
	UpdatedAt        timex.Time `json:"updatedAt" form:"updatedAt"`               // 更新时间
}

type SettingSet struct {
	VaultID     int64  `json:"vaultId" form:"vaultId"`         // 保险库ID
	Action      string `json:"action" form:"action"`           // 操作
	Path        string `json:"path" form:"path"`               // 路径
	PathHash    string `json:"pathHash" form:"pathHash"`       // 路径哈希
	Content     string `json:"content" form:"content"`         // 内容
	ContentHash string `json:"contentHash" form:"contentHash"` // 内容哈希
	Size        int64  `json:"size" form:"size"`               // 大小
	Ctime       int64  `json:"ctime" form:"ctime"`             // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"`             // 修改时间戳
}

// setting 获取配置查询对象
func (d *Dao) setting(uid int64) *query.Query {
	key := "user_setting_" + strconv.FormatInt(uid, 10)
	return d.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "Setting")
	}, key+"#setting", key)
}

// SettingGetByPathHash 根据路径哈希获取配置
func (d *Dao) SettingGetByPathHash(hash string, vaultID int64, uid int64) (*Setting, error) {
	u := d.setting(uid).Setting
	m, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(hash),
	).First()
	if err != nil {
		return nil, err
	}
	res := convert.StructAssign(m, &Setting{}).(*Setting)
	d.fillSettingContent(uid, res)
	return res, nil
}

// SettingCreate 创建配置记录
func (d *Dao) SettingCreate(params *SettingSet, uid int64) (*Setting, error) {
	var result *Setting
	var createErr error

	err := d.ExecuteWrite(context.Background(), uid, func(db *gorm.DB) error {
		u := query.Use(db).Setting
		m := convert.StructAssign(params, &model.Setting{}).(*model.Setting)

		m.UpdatedTimestamp = timex.Now().UnixMilli()
		m.CreatedAt = timex.Now()
		m.UpdatedAt = timex.Now()

		content := m.Content
		m.Content = "" // 不在数据库存储内容

		createErr = u.WithContext(d.ctx).Create(m)
		if createErr != nil {
			return createErr
		}

		// 保存到文件存储
		folderPath := d.GetSettingFolderPath(uid, m.ID)
		if err := d.SaveContentToFile(folderPath, "content.txt", params.Content); err != nil {
			return err
		}

		res := convert.StructAssign(m, &Setting{}).(*Setting)
		res.Content = content
		result = res
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, createErr
}

// SettingUpdate 更新配置记录
func (d *Dao) SettingUpdate(params *SettingSet, id int64, uid int64) (*Setting, error) {
	var result *Setting
	var updateErr error

	err := d.ExecuteWrite(context.Background(), uid, func(db *gorm.DB) error {
		u := query.Use(db).Setting
		m := convert.StructAssign(params, &model.Setting{}).(*model.Setting)
		m.UpdatedTimestamp = timex.Now().UnixMilli()
		m.UpdatedAt = timex.Now()
		m.ID = id

		content := m.Content
		m.Content = "" // 不在数据库更新内容

		updateErr = u.WithContext(d.ctx).Where(
			u.ID.Eq(id),
		).Save(m)

		if updateErr != nil {
			return updateErr
		}

		// 保存到文件存储
		folderPath := d.GetSettingFolderPath(uid, id)
		if err := d.SaveContentToFile(folderPath, "content.txt", params.Content); err != nil {
			return err
		}

		res := convert.StructAssign(m, &Setting{}).(*Setting)
		res.Content = content
		result = res
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, updateErr
}

// SettingUpdateMtime 更新配置的修改时间
func (d *Dao) SettingUpdateMtime(mtime int64, id int64, uid int64) error {
	return d.ExecuteWrite(context.Background(), uid, func(db *gorm.DB) error {
		u := query.Use(db).Setting

		_, err := u.WithContext(d.ctx).Where(
			u.ID.Eq(id),
		).UpdateSimple(
			u.Mtime.Value(mtime),
			u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
			u.UpdatedAt.Value(timex.Now()),
		)

		return err
	})
}

// SettingListByUpdatedTimestamp 根据更新时间戳获取配置列表
func (d *Dao) SettingListByUpdatedTimestamp(timestamp int64, vaultID int64, uid int64) ([]*Setting, error) {
	u := d.setting(uid).Setting
	mList, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.UpdatedTimestamp.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*Setting
	for _, m := range mList {
		res := convert.StructAssign(m, &Setting{}).(*Setting)
		d.fillSettingContent(uid, res)
		list = append(list, res)
	}
	return list, nil
}

// SettingDeletePhysicalByTime 根据时间物理删除配置记录
func (d *Dao) SettingDeletePhysicalByTime(timestamp int64, uid int64) error {
	return d.ExecuteWrite(context.Background(), uid, func(db *gorm.DB) error {
		u := query.Use(db).Setting

		// 查找待物理删除的记录，清理文件
		mList, err := u.WithContext(d.ctx).Where(
			u.Action.Eq("delete"),
			u.UpdatedTimestamp.Lt(timestamp),
		).Find()

		if err == nil {
			for _, m := range mList {
				folder := d.GetSettingFolderPath(uid, m.ID)
				_ = d.RemoveContentFolder(folder)
			}
		}

		_, err = u.WithContext(d.ctx).Where(
			u.Action.Eq("delete"),
			u.UpdatedTimestamp.Lt(timestamp),
		).Delete()

		return err
	})
}

// fillSettingContent 填充配置内容
func (d *Dao) fillSettingContent(uid int64, s *Setting) {
	if s == nil {
		return
	}
	folder := d.GetSettingFolderPath(uid, s.ID)

	// 加载内容
	if content, exists, _ := d.LoadContentFromFile(folder, "content.txt"); exists {
		s.Content = content
	} else if s.Content != "" {
		// 懒迁移失败记录警告日志但不阻断流程
		if err := d.SaveContentToFile(folder, "content.txt", s.Content); err != nil {
			global.Logger.Warn("lazy migration: SaveContentToFile failed for setting content",
				zap.Int64(logger.FieldUID, uid),
				zap.Int64("settingId", s.ID),
				zap.String(logger.FieldMethod, "Dao.fillSettingContent"),
				zap.Error(err),
			)
		}
	}
}
