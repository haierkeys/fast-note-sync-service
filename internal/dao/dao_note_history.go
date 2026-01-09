package dao

import (
	"context"
	"strconv"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NoteHistory struct {
	ID          int64      `json:"id" form:"id"`
	NoteID      int64      `json:"noteId" form:"noteId"`
	VaultID     int64      `json:"vaultId" form:"vaultId"`
	Path        string     `json:"path" form:"path"`
	DiffPatch   string     `json:"diffPatch" form:"diffPatch"`
	Content     string     `json:"content" form:"content"`
	ContentHash string     `json:"contentHash" form:"contentHash"`
	ClientName  string     `json:"clientName" form:"clientName"`
	Version     int64      `json:"version" form:"version"`
	CreatedAt   timex.Time `json:"createdAt" form:"createdAt"`
	UpdatedAt   timex.Time `json:"updatedAt" form:"updatedAt"`
}

type NoteHistorySet struct {
	NoteID      int64      `json:"noteId" form:"noteId"`
	VaultID     int64      `json:"vaultId" form:"vaultId"`
	Path        string     `json:"path" form:"path"`
	DiffPatch   string     `json:"diffPatch" form:"diffPatch"`
	Content     string     `json:"content" form:"content"`
	ContentHash string     `json:"contentHash" form:"contentHash"`
	ClientName  string     `json:"clientName" form:"clientName"`
	Version     int64      `json:"version" form:"version"`
	CreatedAt   timex.Time `json:"createdAt" form:"createdAt"`
	UpdatedAt   timex.Time `json:"updatedAt" form:"updatedAt"`
}

func (d *Dao) noteHistory(uid int64) *query.Query {
	key := "user_note_history_" + strconv.FormatInt(uid, 10)
	return d.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "NoteHistory")
	}, key+"#noteHistory", key)
}

func (d *Dao) NoteHistoryCreate(params *NoteHistorySet, uid int64) (*NoteHistory, error) {
	var result *NoteHistory
	var createErr error

	err := d.ExecuteWrite(context.Background(), uid, func(db *gorm.DB) error {
		u := query.Use(db).NoteHistory
		m := convert.StructAssign(params, &model.NoteHistory{}).(*model.NoteHistory)

		// 暂存内容用于写文件
		diffPatch := m.DiffPatch
		content := m.Content

		// 不在数据库中保存大数据
		m.DiffPatch = ""
		m.Content = ""

		createErr = u.WithContext(d.ctx).Create(m)
		if createErr != nil {
			return createErr
		}

		// 保存到文件
		folder := d.GetNoteHistoryFolderPath(uid, m.ID)
		if err := d.SaveContentToFile(folder, "diff.patch", diffPatch); err != nil {
			return err
		}
		if err := d.SaveContentToFile(folder, "content.txt", content); err != nil {
			return err
		}

		res := convert.StructAssign(m, &NoteHistory{}).(*NoteHistory)
		res.DiffPatch = diffPatch
		res.Content = content
		result = res
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, createErr
}

func (d *Dao) NoteHistoryGetById(id int64, uid int64) (*NoteHistory, error) {
	u := d.noteHistory(uid).NoteHistory
	m, err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	res := convert.StructAssign(m, &NoteHistory{}).(*NoteHistory)
	d.fillHistoryContent(uid, res)
	return res, nil
}

func (d *Dao) NoteHistoryListByNoteId(noteId int64, page, pageSize int, uid int64) ([]*NoteHistory, int64, error) {
	u := d.noteHistory(uid).NoteHistory
	q := u.WithContext(d.ctx).Where(u.NoteID.Eq(noteId))

	count, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	modelList, err := q.Order(u.Version.Desc()).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find()
	if err != nil {
		return nil, 0, err
	}

	var list []*NoteHistory
	for _, m := range modelList {
		res := convert.StructAssign(m, &NoteHistory{}).(*NoteHistory)
		d.fillHistoryContent(uid, res)
		list = append(list, res)
	}
	return list, count, nil
}

func (d *Dao) NoteHistoryGetLatestVersion(noteId int64, uid int64) (int64, error) {
	u := d.noteHistory(uid).NoteHistory
	m, err := u.WithContext(d.ctx).Where(u.NoteID.Eq(noteId)).Order(u.Version.Desc()).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return m.Version, nil
}

func (d *Dao) NoteHistoryGetByNoteIdAndHash(noteId int64, contentHash string, uid int64) (*NoteHistory, error) {
	u := d.noteHistory(uid).NoteHistory
	m, err := u.WithContext(d.ctx).Where(u.NoteID.Eq(noteId), u.ContentHash.Eq(contentHash)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	res := convert.StructAssign(m, &NoteHistory{}).(*NoteHistory)
	d.fillHistoryContent(uid, res)
	return res, nil
}

func (d *Dao) NoteHistoryMigrate(oldNoteID, newNoteID int64, uid int64) error {
	return d.ExecuteWrite(context.Background(), uid, func(db *gorm.DB) error {
		u := query.Use(db).NoteHistory
		_, err := u.WithContext(d.ctx).Where(u.NoteID.Eq(oldNoteID)).Update(u.NoteID, newNoteID)
		return err
	})
}

// fillHistoryContent 填充历史记录内容及补丁
func (d *Dao) fillHistoryContent(uid int64, h *NoteHistory) {
	if h == nil {
		return
	}
	folder := d.GetNoteHistoryFolderPath(uid, h.ID)

	// 加载补丁
	if patch, exists, _ := d.LoadContentFromFile(folder, "diff.patch"); exists {
		h.DiffPatch = patch
	} else if h.DiffPatch != "" {
		// 懒迁移失败记录警告日志但不阻断流程
		if err := d.SaveContentToFile(folder, "diff.patch", h.DiffPatch); err != nil {
			global.Logger.Warn("lazy migration: SaveContentToFile failed for history diff patch",
				zap.Int64(logger.FieldUID, uid),
				zap.Int64("historyId", h.ID),
				zap.String(logger.FieldMethod, "Dao.fillHistoryContent"),
				zap.Error(err),
			)
		}
	}

	// 加载内容
	if content, exists, _ := d.LoadContentFromFile(folder, "content.txt"); exists {
		h.Content = content
	} else if h.Content != "" {
		// 懒迁移失败记录警告日志但不阻断流程
		if err := d.SaveContentToFile(folder, "content.txt", h.Content); err != nil {
			global.Logger.Warn("lazy migration: SaveContentToFile failed for history content",
				zap.Int64(logger.FieldUID, uid),
				zap.Int64("historyId", h.ID),
				zap.String(logger.FieldMethod, "Dao.fillHistoryContent"),
				zap.Error(err),
			)
		}
	}
}
