// Package dao 实现数据访问层
package dao

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
)

// noteHistoryRepository 实现 domain.NoteHistoryRepository 接口
type noteHistoryRepository struct {
	dao *Dao
}

// NewNoteHistoryRepository 创建 NoteHistoryRepository 实例
func NewNoteHistoryRepository(dao *Dao) domain.NoteHistoryRepository {
	return &noteHistoryRepository{dao: dao}
}

// modelToDomain 将数据库模型转换为领域模型
func (r *noteHistoryRepository) modelToDomain(m *NoteHistory) *domain.NoteHistory {
	if m == nil {
		return nil
	}
	return &domain.NoteHistory{
		ID:          m.ID,
		NoteID:      m.NoteID,
		VaultID:     m.VaultID,
		Path:        m.Path,
		DiffPatch:   m.DiffPatch,
		Content:     m.Content,
		ContentHash: m.ContentHash,
		ClientName:  m.ClientName,
		Version:     m.Version,
		CreatedAt:   time.Time(m.CreatedAt),
		UpdatedAt:   time.Time(m.UpdatedAt),
	}
}

// GetByID 根据ID获取历史记录
func (r *noteHistoryRepository) GetByID(ctx context.Context, id, uid int64) (*domain.NoteHistory, error) {
	history, err := r.dao.NoteHistoryGetById(id, uid)
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(history), nil
}

// GetByNoteIDAndHash 根据笔记ID和内容哈希获取历史记录
func (r *noteHistoryRepository) GetByNoteIDAndHash(ctx context.Context, noteID int64, contentHash string, uid int64) (*domain.NoteHistory, error) {
	history, err := r.dao.NoteHistoryGetByNoteIdAndHash(noteID, contentHash, uid)
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(history), nil
}

// Create 创建历史记录
func (r *noteHistoryRepository) Create(ctx context.Context, history *domain.NoteHistory, uid int64) (*domain.NoteHistory, error) {
	params := &NoteHistorySet{
		NoteID:      history.NoteID,
		VaultID:     history.VaultID,
		Path:        history.Path,
		DiffPatch:   history.DiffPatch,
		Content:     history.Content,
		ContentHash: history.ContentHash,
		ClientName:  history.ClientName,
		Version:     history.Version,
		CreatedAt:   timex.Time(history.CreatedAt),
		UpdatedAt:   timex.Time(history.UpdatedAt),
	}
	created, err := r.dao.NoteHistoryCreate(params, uid)
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(created), nil
}

// ListByNoteID 根据笔记ID获取历史记录列表
func (r *noteHistoryRepository) ListByNoteID(ctx context.Context, noteID int64, page, pageSize int, uid int64) ([]*domain.NoteHistory, int64, error) {
	histories, count, err := r.dao.NoteHistoryListByNoteId(noteID, page, pageSize, uid)
	if err != nil {
		return nil, 0, err
	}

	var results []*domain.NoteHistory
	for _, h := range histories {
		results = append(results, r.modelToDomain(h))
	}
	return results, count, nil
}

// GetLatestVersion 获取笔记的最新版本号
func (r *noteHistoryRepository) GetLatestVersion(ctx context.Context, noteID, uid int64) (int64, error) {
	return r.dao.NoteHistoryGetLatestVersion(noteID, uid)
}

// Migrate 迁移历史记录（更新 NoteID）
func (r *noteHistoryRepository) Migrate(ctx context.Context, oldNoteID, newNoteID, uid int64) error {
	return r.dao.NoteHistoryMigrate(oldNoteID, newNoteID, uid)
}

// 确保 noteHistoryRepository 实现了 domain.NoteHistoryRepository 接口
var _ domain.NoteHistoryRepository = (*noteHistoryRepository)(nil)
