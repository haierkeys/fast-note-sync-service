// Package dao implements the data access layer
// Package dao 实现数据访问层
package dao

import (
	"context"
	"strconv"

	"github.com/blevesearch/bleve/v2"
	bleveQuery "github.com/blevesearch/bleve/v2/search/query"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"go.uber.org/zap"
)

// noteFTSRepository implements domain.NoteFTSRepository interface
// noteFTSRepository 实现 domain.NoteFTSRepository 接口
type noteFTSRepository struct {
	dao             *Dao
	customPrefixKey string
}

// NewNoteFTSRepository creates domain.NoteFTSRepository instance
// NewNoteFTSRepository 创建 domain.NoteFTSRepository 实例
func NewNoteFTSRepository(dao *Dao) domain.NoteFTSRepository {
	return &noteFTSRepository{dao: dao, customPrefixKey: "user_note_history_"}
}

func (r *noteFTSRepository) GetKey(uid int64) string {
	return r.customPrefixKey + strconv.FormatInt(uid, 10)
}

// Upsert inserts or updates FTS index
// Upsert 插入或更新 FTS 索引
func (r *noteFTSRepository) Upsert(ctx context.Context, noteID int64, path, content string, uid int64) error {
	if !r.dao.BleveMgr.IsEnabled() {
		return nil // If FTS is disabled, do nothing // 若 FTS 未启用，则不进行任何操作
	}
	db := r.dao.ResolveDB("user_" + strconv.FormatInt(uid, 10))
	var note model.Note
	if err := db.Where("id = ?", noteID).First(&note).Error; err != nil {
		return err
	}

	index, err := r.dao.BleveMgr.GetIndex(uid, note.VaultID)
	if err != nil {
		return err
	}

	doc := BleveNoteDoc{
		ID:      strconv.FormatInt(noteID, 10),
		Path:    path,
		PathRaw: path,
		Content: content,
		Action:  note.Action,
		Rename:  float64(note.Rename),
		Ctime:   float64(note.Ctime),
		Mtime:   float64(note.Mtime),
	}

	return index.Index(doc.ID, doc)
}

// Delete deletes FTS index
// Delete 删除 FTS 索引
func (r *noteFTSRepository) Delete(ctx context.Context, noteID int64, uid int64) error {
	if !r.dao.BleveMgr.IsEnabled() {
		return nil // If FTS is disabled, do nothing // 若 FTS 未启用，则不进行任何操作
	}
	db := r.dao.ResolveDB("user_" + strconv.FormatInt(uid, 10))
	var note model.Note
	if err := db.Where("id = ?", noteID).First(&note).Error; err != nil {
		// Fallback: if note is already physically deleted, we don't know the vaultID.
		// Try to delete from all vaults of this user.
		// 容错：如果笔记已被物理删除，我们无法获知其 vaultID。尝试从该用户的所有仓库索引中将其删除。
		var vaults []model.Vault
		vaultDb := r.dao.ResolveDB("user_vault_" + strconv.FormatInt(uid, 10))
		if err := vaultDb.Table("vault").Find(&vaults).Error; err == nil {
			for _, v := range vaults {
				if index, err := r.dao.BleveMgr.GetIndex(uid, v.ID); err == nil {
					_ = index.Delete(strconv.FormatInt(noteID, 10))
				}
			}
		}
		return nil
	}

	index, err := r.dao.BleveMgr.GetIndex(uid, note.VaultID)
	if err != nil {
		return err
	}

	return index.Delete(strconv.FormatInt(noteID, 10))
}

// Search full-text search, returns list of matching note_id
// Search 全文搜索，返回匹配的 note_id 列表
func (r *noteFTSRepository) Search(ctx context.Context, keyword string, vaultID, uid int64, limit, offset int) ([]int64, error) {
	if !r.dao.BleveMgr.IsEnabled() {
		return nil, nil // Return empty results if FTS is disabled // 若 FTS 未启用，则返回空结果
	}
	index, err := r.dao.BleveMgr.GetIndex(uid, vaultID)
	if err != nil {
		return nil, err
	}

	pathQuery := bleve.NewMatchQuery(keyword)
	pathQuery.SetField("path")
	pathQuery.Operator = bleveQuery.MatchQueryOperatorAnd

	contentQuery := bleve.NewMatchQuery(keyword)
	contentQuery.SetField("content")
	contentQuery.Operator = bleveQuery.MatchQueryOperatorAnd

	actionQuery := bleve.NewBooleanQuery()
	actionTermQuery := bleve.NewTermQuery("delete")
	actionTermQuery.SetField("action")
	actionQuery.AddMustNot(actionTermQuery)

	query := bleve.NewConjunctionQuery(
		bleve.NewDisjunctionQuery(pathQuery, contentQuery),
		actionQuery,
	)

	req := bleve.NewSearchRequest(query)
	req.Size = limit
	req.From = offset
	req.SortBy([]string{"-mtime"})

	res, err := index.Search(req)
	if err != nil {
		return nil, err
	}

	var noteIDs []int64
	for _, hit := range res.Hits {
		id, _ := strconv.ParseInt(hit.ID, 10, 64)
		noteIDs = append(noteIDs, id)
	}

	// Log search keyword and result IDs
	// 记录搜索关键词与结果 ID 列表的日志
	r.dao.Logger().Info("FTS Search full-text search execution",
		zap.String("keyword", keyword),
		zap.Int64("uid", uid),
		zap.Int64("vaultID", vaultID),
		zap.Int64s("results", noteIDs),
		zap.Int("total_hits", int(res.Total)),
	)

	return noteIDs, nil
}

// SearchCount full-text search count
// SearchCount 全文搜索计数
func (r *noteFTSRepository) SearchCount(ctx context.Context, keyword string, vaultID, uid int64) (int64, error) {
	if !r.dao.BleveMgr.IsEnabled() {
		return 0, nil // Return 0 if FTS is disabled // 若 FTS 未启用，则返回 0
	}
	index, err := r.dao.BleveMgr.GetIndex(uid, vaultID)
	if err != nil {
		return 0, err
	}

	pathQuery := bleve.NewMatchQuery(keyword)
	pathQuery.SetField("path")
	pathQuery.Operator = bleveQuery.MatchQueryOperatorAnd

	contentQuery := bleve.NewMatchQuery(keyword)
	contentQuery.SetField("content")
	contentQuery.Operator = bleveQuery.MatchQueryOperatorAnd

	actionQuery := bleve.NewBooleanQuery()
	actionTermQuery := bleve.NewTermQuery("delete")
	actionTermQuery.SetField("action")
	actionQuery.AddMustNot(actionTermQuery)

	query := bleve.NewConjunctionQuery(
		bleve.NewDisjunctionQuery(pathQuery, contentQuery),
		actionQuery,
	)

	req := bleve.NewSearchRequest(query)
	req.Size = 0
	res, err := index.Search(req)
	if err != nil {
		return 0, err
	}

	return int64(res.Total), nil
}

// RebuildIndex rebuilds index
// RebuildIndex 重建索引
func (r *noteFTSRepository) RebuildIndex(ctx context.Context, uid int64) error {
	if !r.dao.BleveMgr.IsEnabled() {
		return nil // If FTS is disabled, do nothing // 若 FTS 未启用，则不进行任何操作
	}
	var vaults []model.Vault
	vaultDb := r.dao.ResolveDB("user_vault_" + strconv.FormatInt(uid, 10))
	if err := vaultDb.Table("vault").Where("is_deleted = 0").Find(&vaults).Error; err != nil {
		return err
	}

	for _, v := range vaults {
		_ = r.rebuildVault(ctx, uid, v.ID)
	}

	return nil
}

// rebuildVault rebuilds index for a specific vault
// rebuildVault 重建特定仓库的索引
func (r *noteFTSRepository) rebuildVault(ctx context.Context, uid, vaultID int64) error {
	if !r.dao.BleveMgr.IsEnabled() {
		return nil // If FTS is disabled, do nothing // 若 FTS 未启用，则不进行任何操作
	}
	_ = r.dao.BleveMgr.DeleteIndex(uid, vaultID)

	index, err := r.dao.BleveMgr.GetIndex(uid, vaultID)
	if err != nil {
		return err
	}

	db := r.dao.ResolveDB("user_" + strconv.FormatInt(uid, 10))
	var notes []model.Note
	if err := db.Where("vault_id = ?", vaultID).Find(&notes).Error; err != nil {
		return err
	}

	for _, note := range notes {
		folder := r.dao.GetNoteFolderPath(uid, note.ID)
		content, exists, err := r.dao.LoadContentFromFile(folder, "content.txt")
		if err != nil || !exists {
			content = ""
		}

		doc := BleveNoteDoc{
			ID:      strconv.FormatInt(note.ID, 10),
			Path:    note.Path,
			PathRaw: note.Path,
			Content: content,
			Action:  note.Action,
			Rename:  float64(note.Rename),
			Ctime:   float64(note.Ctime),
			Mtime:   float64(note.Mtime),
		}

		_ = index.Index(doc.ID, doc)
	}

	return nil
}

// DeleteByVaultID deletes all FTS records for a vault
// DeleteByVaultID 删除指定仓库的所有 FTS 记录
func (r *noteFTSRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	if !r.dao.BleveMgr.IsEnabled() {
		return nil // If FTS is disabled, do nothing // 若 FTS 未启用，则不进行任何操作
	}
	return r.dao.BleveMgr.DeleteIndex(uid, vaultID)
}

// Ensure noteFTSRepository implements domain.NoteFTSRepository interface
// 确保 noteFTSRepository 实现了 domain.NoteFTSRepository 接口
var _ domain.NoteFTSRepository = (*noteFTSRepository)(nil)
