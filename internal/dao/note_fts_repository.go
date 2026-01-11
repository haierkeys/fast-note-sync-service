// Package dao 实现数据访问层
package dao

import (
	"context"
	"strconv"

	"github.com/haierkeys/fast-note-sync-service/internal/model"
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

// ensureFTSTable 确保 FTS 表存在
func (r *noteFTSRepository) ensureFTSTable(uid int64) *gorm.DB {
	key := r.GetKey(uid)
	db := r.dao.UseKey(key)
	if db == nil {
		return nil
	}

	// 使用 onceKeys 确保只创建一次
	onceKey := key + "#note_fts"
	if _, loaded := r.dao.onceKeys.LoadOrStore(onceKey, true); !loaded {
		_ = model.CreateNoteFTSTable(db)
	}

	return db
}

// Upsert 插入或更新 FTS 索引
func (r *noteFTSRepository) Upsert(ctx context.Context, noteID int64, path, content string, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		// 确保 FTS 表存在
		_ = model.CreateNoteFTSTable(db)

		// 先删除旧记录
		if err := db.Exec("DELETE FROM note_fts WHERE note_id = ?", noteID).Error; err != nil {
			return err
		}

		// 插入新记录
		return db.Exec("INSERT INTO note_fts (note_id, path, content) VALUES (?, ?, ?)",
			noteID, path, content).Error
	})
}

// Delete 删除 FTS 索引
func (r *noteFTSRepository) Delete(ctx context.Context, noteID int64, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		return db.Exec("DELETE FROM note_fts WHERE note_id = ?", noteID).Error
	})
}

// Search 全文搜索，返回匹配的 note_id 列表
func (r *noteFTSRepository) Search(ctx context.Context, keyword string, vaultID, uid int64, limit, offset int) ([]int64, error) {
	db := r.ensureFTSTable(uid)
	if db == nil {
		return nil, nil
	}

	var noteIDs []int64

	// FTS5 MATCH 查询
	// 使用子查询关联 note 表过滤 vault_id 和 action
	sql := `
		SELECT f.note_id
		FROM note_fts f
		INNER JOIN note n ON f.note_id = n.id
		WHERE note_fts MATCH ?
		AND n.vault_id = ?
		AND n.action != 'delete'
		ORDER BY rank
		LIMIT ? OFFSET ?
	`

	// FTS5 查询语法：使用 * 进行前缀匹配
	searchTerm := escapeForFTS(keyword)

	err := db.WithContext(ctx).Raw(sql, searchTerm, vaultID, limit, offset).Scan(&noteIDs).Error
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

	var count int64

	sql := `
		SELECT COUNT(*)
		FROM note_fts f
		INNER JOIN note n ON f.note_id = n.id
		WHERE note_fts MATCH ?
		AND n.vault_id = ?
		AND n.action != 'delete'
	`

	searchTerm := escapeForFTS(keyword)

	err := db.WithContext(ctx).Raw(sql, searchTerm, vaultID).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// RebuildIndex 重建索引
func (r *noteFTSRepository) RebuildIndex(ctx context.Context, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		// 删除并重建 FTS 表
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

		// 批量插入 FTS 索引
		for _, note := range notes {
			// 从文件系统读取内容
			folder := r.dao.GetNoteFolderPath(uid, note.ID)
			content, _, _ := r.dao.LoadContentFromFile(folder, "content.txt")

			if err := db.Exec("INSERT INTO note_fts (note_id, path, content) VALUES (?, ?, ?)",
				note.ID, note.Path, content).Error; err != nil {
				// 记录错误但继续处理
				continue
			}
		}

		return nil
	})
}

// escapeForFTS 转义 FTS5 特殊字符
func escapeForFTS(s string) string {
	// FTS5 特殊字符：" * - + ( ) : ^
	// 对于简单搜索，我们使用双引号包裹整个搜索词
	// 这样可以进行精确短语匹配
	return `"` + s + `"`
}

// 确保 noteFTSRepository 实现了 NoteFTSRepository 接口
var _ NoteFTSRepository = (*noteFTSRepository)(nil)
