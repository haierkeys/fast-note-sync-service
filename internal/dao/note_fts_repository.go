// Package dao 实现数据访问层
package dao

import (
	"context"
	"strconv"

	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"gorm.io/gorm"
)

// NoteFTSRepository FTS 全文搜索仓库接口
type NoteFTSRepository interface {
	// Upsert 插入或更新 FTS 索引
	Upsert(ctx context.Context, noteID int64, path, content string, uid int64) error
	// Delete 删除 FTS 索引
	Delete(ctx context.Context, noteID int64, uid int64) error
	// Search 全文搜索，返回匹配的 note_id 列表
	Search(ctx context.Context, keyword string, vaultID, uid int64, limit, offset int) ([]int64, error)
	// SearchCount 全文搜索计数
	SearchCount(ctx context.Context, keyword string, vaultID, uid int64) (int64, error)
	// RebuildIndex 重建索引（从文件系统读取所有笔记内容）
	RebuildIndex(ctx context.Context, uid int64) error
}

// noteFTSRepository 实现 NoteFTSRepository 接口
type noteFTSRepository struct {
	dao             *Dao
	customPrefixKey string
}

// NewNoteFTSRepository 创建 NoteFTSRepository 实例
func NewNoteFTSRepository(dao *Dao) NoteFTSRepository {
	return &noteFTSRepository{dao: dao, customPrefixKey: "user_note_history_"}
}

func (r *noteFTSRepository) GetKey(uid int64) string {
	return r.customPrefixKey + strconv.FormatInt(uid, 10)
}

// ensureFTSTable 确保 FTS 相关表存在
func (r *noteFTSRepository) ensureFTSTable(uid int64) *gorm.DB {
	key := r.GetKey(uid)
	db := r.dao.ResolveDB(key)
	if db == nil {
		return nil
	}

	// 使用 onceKeys 确保只创建一次
	onceKey := key + "#note_fts_v3"
	if _, loaded := r.dao.onceKeys.LoadOrStore(onceKey, true); !loaded {
		_ = model.CreateNoteFTSTable(db)
	}

	return db
}

// Upsert 插入或更新 FTS 索引
func (r *noteFTSRepository) Upsert(ctx context.Context, noteID int64, path, content string, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		// 确保表存在
		_ = model.CreateNoteFTSTable(db)

		// 1. 更新快照表
		noteFTS := model.NoteFTS{
			NoteID:  noteID,
			Path:    path,
			Content: content,
		}
		if err := db.Save(&noteFTS).Error; err != nil {
			return err
		}

		// 2. 更新倒排索引表
		// 先删除旧索引
		if err := db.Where("note_id = ?", noteID).Delete(&model.NoteFTSToken{}).Error; err != nil {
			return err
		}

		// 分词
		tokens := util.Tokenize(path + " " + content)
		if len(tokens) == 0 {
			return nil
		}

		// 批量插入新索引
		var tokenModels []model.NoteFTSToken
		for _, t := range tokens {
			tokenModels = append(tokenModels, model.NoteFTSToken{
				NoteID: noteID,
				Token:  t,
			})
		}

		return db.CreateInBatches(tokenModels, 500).Error
	})
}

// Delete 删除 FTS 索引
func (r *noteFTSRepository) Delete(ctx context.Context, noteID int64, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		_ = db.Where("note_id = ?", noteID).Delete(&model.NoteFTS{})
		return db.Where("note_id = ?", noteID).Delete(&model.NoteFTSToken{}).Error
	})
}

// Search 全文搜索
func (r *noteFTSRepository) Search(ctx context.Context, keyword string, vaultID, uid int64, limit, offset int) ([]int64, error) {
	db := r.ensureFTSTable(uid)
	if db == nil {
		return nil, nil
	}

	tokens := util.Tokenize(keyword)
	if len(tokens) == 0 {
		return nil, nil
	}

	var noteIDs []int64

	// 构建搜索 SQL：在 NoteFTSToken 表中查找包含所有 Token 的 NoteID
	// 并关联 Note 表以过滤 VaultID 和 Action
	query := db.Table("note_fts_token AS t").
		Select("t.note_id").
		Joins("INNER JOIN note ON t.note_id = note.id").
		Where("t.token IN ?", tokens).
		Where("note.vault_id = ?", vaultID).
		Where("note.action != ?", "delete").
		Group("t.note_id").
		Having("COUNT(DISTINCT t.token) = ?", len(tokens)).
		Order("COUNT(t.id) DESC") // 简单的排名：出现频率越高排名越前

	err := query.WithContext(ctx).Limit(limit).Offset(offset).Scan(&noteIDs).Error
	if err != nil {
		return nil, err
	}

	return noteIDs, nil
}

// SearchCount 全文搜索计数
func (r *noteFTSRepository) SearchCount(ctx context.Context, keyword string, vaultID, uid int64) (int64, error) {
	db := r.ensureFTSTable(uid)
	if db == nil {
		return 0, nil
	}

	tokens := util.Tokenize(keyword)
	if len(tokens) == 0 {
		return 0, nil
	}

	var count int64

	subQuery := db.Table("note_fts_token AS t").
		Select("t.note_id").
		Joins("INNER JOIN note ON t.note_id = note.id").
		Where("t.token IN ?", tokens).
		Where("note.vault_id = ?", vaultID).
		Where("note.action != ?", "delete").
		Group("t.note_id").
		Having("COUNT(DISTINCT t.token) = ?", len(tokens))

	err := db.Table("(?) AS sub", subQuery).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// RebuildIndex 重建索引
func (r *noteFTSRepository) RebuildIndex(ctx context.Context, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		// 删除并重建表
		if err := model.DropNoteFTSTable(db); err != nil {
			return err
		}
		if err := model.CreateNoteFTSTable(db); err != nil {
			return err
		}

		// 获取所有笔记
		var notes []model.Note
		if err := db.Where("action != ?", "delete").Find(&notes).Error; err != nil {
			return err
		}

		// 重新索引
		for _, note := range notes {
			folder := r.dao.GetNoteFolderPath(uid, note.ID)
			content, exists, err := r.dao.LoadContentFromFile(folder, "content.txt")
			if err != nil {
				return err
			}
			if !exists {
				content = ""
			}

			// 手动调用 Upsert 逻辑的部分（因为已经在事务里）
			noteFTS := model.NoteFTS{NoteID: note.ID, Path: note.Path, Content: content}
			db.Save(&noteFTS)

			tokens := util.Tokenize(note.Path + " " + content)
			if len(tokens) == 0 {
				continue
			}

			var tokenModels []model.NoteFTSToken
			for _, t := range tokens {
				tokenModels = append(tokenModels, model.NoteFTSToken{NoteID: note.ID, Token: t})
			}
			db.CreateInBatches(tokenModels, 500)
		}

		return nil
	})
}

// 确保 noteFTSRepository 实现了 NoteFTSRepository 接口
var _ NoteFTSRepository = (*noteFTSRepository)(nil)
