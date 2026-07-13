// Package dao implements the data access layer
// Package dao 实现数据访问层
package dao

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/blevesearch/bleve/v2"
	bleveQuery "github.com/blevesearch/bleve/v2/search/query"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// noteRepository implements domain.NoteRepository interface
// noteRepository 实现 domain.NoteRepository 接口
type noteRepository struct {
	dao             *Dao
	customPrefixKey string
}

// NewNoteRepository creates NoteRepository instance
// NewNoteRepository 创建 NoteRepository 实例
func NewNoteRepository(dao *Dao) domain.NoteRepository {
	return &noteRepository{dao: dao, customPrefixKey: "user_"}
}

func (r *noteRepository) GetKey(uid int64) string {
	return r.customPrefixKey + strconv.FormatInt(uid, 10)
}

func init() {
	RegisterModel(ModelConfig{
		Name: "Note",
		RepoFactory: func(d *Dao) daoDBCustomKey {
			return NewNoteRepository(d).(daoDBCustomKey)
		},
	})
}

// note 获取笔记查询对象
func (r *noteRepository) note(uid int64) *query.Query {
	return r.dao.QueryWithOnceInit(func(g *gorm.DB) {
		model.AutoMigrate(g, "Note")
		// Initialize universal full-text search table
		// 初始化通用全文搜索表
		_ = model.CreateNoteFTSTable(g)
	}, r.GetKey(uid)+"#note_v3", r.GetKey(uid))
}

// ListByIDs retrieves note list by ID list
// ListByIDs 根据ID列表获取笔记列表
func (r *noteRepository) ListByIDs(ctx context.Context, ids []int64, uid int64) ([]*domain.Note, error) {
	if len(ids) == 0 {
		return []*domain.Note{}, nil
	}
	u := r.note(uid).Note
	ms, err := u.WithContext(ctx).Where(u.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.Note
	for _, m := range ms {
		note, err := r.toDomain(m, uid)
		if err != nil {
			return nil, err
		}
		res = append(res, note)
	}
	return res, nil
}

// EnsureFTSIndex ensures FTS index exists (public method, can be called manually)
// EnsureFTSIndex 确保 FTS 索引存在（公开方法，可手动调用）
func (r *noteRepository) EnsureFTSIndex(ctx context.Context, uid int64) error {
	var vaults []model.Vault
	vaultDb := r.dao.ResolveDB("user_vault_" + strconv.FormatInt(uid, 10))
	if err := vaultDb.Table("vault").Where("is_deleted = 0").Find(&vaults).Error; err != nil {
		return err
	}

	for _, v := range vaults {
		path := r.dao.BleveMgr.GetIndexPath(uid, v.ID)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			_ = r.RebuildVaultIndex(ctx, uid, v.ID)
		}
	}
	return nil
}

// toDomain converts DAO Note to domain model
// toDomain 将 DAO Note 转换为领域模型
func (r *noteRepository) toDomain(m *model.Note, uid int64) (*domain.Note, error) {
	if m == nil {
		return nil, nil
	}
	note := &domain.Note{
		ID:                      m.ID,
		VaultID:                 m.VaultID,
		Action:                  domain.NoteAction(m.Action),
		Rename:                  m.Rename,
		FID:                     m.FID,
		Path:                    m.Path,
		PathHash:                m.PathHash,
		Content:                 m.Content,
		ContentHash:             m.ContentHash,
		ContentLastSnapshot:     m.ContentLastSnapshot,
		ContentLastSnapshotHash: m.ContentLastSnapshotHash,
		Version:                 m.Version,
		ClientName:              m.ClientName,
		ClientType:              m.ClientType,
		ClientVersion:           m.ClientVersion,
		Size:                    m.Size,
		Ctime:                   m.Ctime,
		Mtime:                   m.Mtime,
		UpdatedTimestamp:        m.UpdatedTimestamp,
		CreatedAt:               time.Time(m.CreatedAt),
		UpdatedAt:               time.Time(m.UpdatedAt),
	}
	if err := r.fillNoteContent(uid, note); err != nil {
		return nil, err
	}
	return note, nil
}

// toModel converts domain model to database model
// toModel 将领域模型转换为数据库模型
func (r *noteRepository) toModel(note *domain.Note) *model.Note {
	if note == nil {
		return nil
	}
	return &model.Note{
		ID:                      note.ID,
		VaultID:                 note.VaultID,
		Action:                  string(note.Action),
		Rename:                  note.Rename,
		FID:                     note.FID,
		Path:                    note.Path,
		PathHash:                note.PathHash,
		Content:                 note.Content,
		ContentHash:             note.ContentHash,
		ContentLastSnapshot:     note.ContentLastSnapshot,
		ContentLastSnapshotHash: note.ContentLastSnapshotHash,
		Version:                 note.Version,
		ClientName:              note.ClientName,
		ClientType:              note.ClientType,
		ClientVersion:           note.ClientVersion,
		Size:                    note.Size,
		Ctime:                   note.Ctime,
		Mtime:                   note.Mtime,
		UpdatedTimestamp:        note.UpdatedTimestamp,
		CreatedAt:               timex.Time(note.CreatedAt),
		UpdatedAt:               timex.Time(note.UpdatedAt),
	}
}

// fillNoteContent fills note content
// fillNoteContent 填充笔记内容
func (r *noteRepository) fillNoteContent(uid int64, n *domain.Note) error {
	if n == nil {
		return nil
	}
	folder := r.dao.GetNoteFolderPath(uid, n.ID)

	// Load content
	// 加载内容
	content, exists, err := r.dao.LoadContentFromFile(folder, "content.txt")
	if err != nil {
		return err
	}
	if exists {
		n.Content = content
	} else if n.Content != "" {
		// Lazy migration failed, log warning but do not block flow
		// 懒迁移失败记录警告日志但不阻断流程
		if err := r.dao.SaveContentToFile(folder, "content.txt", n.Content); err != nil {
			r.dao.Logger().Warn("lazy migration: SaveContentToFile failed for note content",
				zap.Int64(logger.FieldUID, uid),
				zap.Int64("noteId", n.ID),
				zap.String(logger.FieldMethod, "noteRepository.fillNoteContent"),
				zap.Error(err),
			)
		}
	} else {
		// File does not exist and no content to migrate, return error to prevent data loss (treated as read failure)
		// 文件不存在且没有可迁移的内容，返回错误以防止数据丢失（视为读取失败）
		return fmt.Errorf("note content file not found: %w", os.ErrNotExist)
	}

	// Load snapshot
	// 加载快照
	snapshot, exists, err := r.dao.LoadContentFromFile(folder, "snapshot.txt")
	if err != nil {
		return err
	}
	if exists {
		n.ContentLastSnapshot = snapshot
	} else if n.ContentLastSnapshot != "" {
		// Lazy migration failed, log warning but do not block flow
		// 懒迁移失败记录警告日志但不阻断流程
		if err := r.dao.SaveContentToFile(folder, "snapshot.txt", n.ContentLastSnapshot); err != nil {
			r.dao.Logger().Warn("lazy migration: SaveContentToFile failed for note snapshot",
				zap.Int64(logger.FieldUID, uid),
				zap.Int64("noteId", n.ID),
				zap.String(logger.FieldMethod, "noteRepository.fillNoteContent"),
				zap.Error(err),
			)
		}
	}

	return nil
}

// toDomainMeta converts DAO Note to domain model without loading content from disk
// toDomainMeta 将 DAO Note 转换为领域模型，但不从磁盘加载正文/快照（仅用于哈希比对等元数据场景）
func (r *noteRepository) toDomainMeta(m *model.Note) *domain.Note {
	if m == nil {
		return nil
	}
	return &domain.Note{
		ID:                      m.ID,
		VaultID:                 m.VaultID,
		Action:                  domain.NoteAction(m.Action),
		Rename:                  m.Rename,
		FID:                     m.FID,
		Path:                    m.Path,
		PathHash:                m.PathHash,
		ContentHash:             m.ContentHash,
		ContentLastSnapshotHash: m.ContentLastSnapshotHash,
		Version:                 m.Version,
		ClientName:              m.ClientName,
		ClientType:              m.ClientType,
		ClientVersion:           m.ClientVersion,
		Size:                    m.Size,
		Ctime:                   m.Ctime,
		Mtime:                   m.Mtime,
		UpdatedTimestamp:        m.UpdatedTimestamp,
		CreatedAt:               time.Time(m.CreatedAt),
		UpdatedAt:               time.Time(m.UpdatedAt),
	}
}

// GetByID retrieves note by ID
// GetByID 根据ID获取笔记
func (r *noteRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Note, error) {
	u := r.note(uid).Note
	m, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid)
}

// GetByPathHash retrieves note by path hash (excluding deleted)
// GetByPathHash 根据路径哈希获取笔记（排除已删除）
func (r *noteRepository) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Note, error) {
	u := r.note(uid).Note
	m, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(pathHash),
		u.Action.Neq("delete"),
	).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid)
}

// GetByPathHashIncludeRecycle retrieves note by path hash (optionally including recycle bin)
// GetByPathHashIncludeRecycle 根据路径哈希获取笔记（可选包含回收站）
func (r *noteRepository) GetByPathHashIncludeRecycle(ctx context.Context, pathHash string, vaultID, uid int64, isRecycle bool) (*domain.Note, error) {
	u := r.note(uid).Note
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(pathHash),
	)

	if isRecycle {
		q = q.Where(u.Action.Eq("delete"), u.Rename.Eq(0))
	} else {
		q = q.Where(u.Action.Neq("delete"))
	}

	m, err := q.First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid)
}

// GetAllByPathHash retrieves note by path hash (including all statuses)
// GetAllByPathHash 根据路径哈希获取笔记（包含所有状态）
func (r *noteRepository) GetAllByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Note, error) {
	u := r.note(uid).Note
	m, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(pathHash),
	).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid)
}

// ListByPathHash retrieves note list by path hash (handling duplicate records)
// ListByPathHash 根据路径哈希获取笔记列表（处理重复记录）
func (r *noteRepository) ListByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*domain.Note, error) {
	u := r.note(uid).Note
	ms, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(pathHash),
	).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.Note
	for _, m := range ms {
		note, err := r.toDomain(m, uid)
		if err != nil {
			return nil, err
		}
		res = append(res, note)
	}
	return res, nil
}

// GetByPath retrieves note by path
// GetByPath 根据路径获取笔记
func (r *noteRepository) GetByPath(ctx context.Context, path string, vaultID, uid int64) (*domain.Note, error) {
	u := r.note(uid).Note
	m, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Path.Eq(path),
	).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid)
}

// Create creates a note
// Create 创建笔记
func (r *noteRepository) Create(ctx context.Context, note *domain.Note, uid int64) (*domain.Note, error) {
	var result *domain.Note
	var createErr error

	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note
		m := r.toModel(note)

		m.UpdatedTimestamp = timex.Now().UnixMilli()
		m.CreatedAt = timex.Now()
		m.UpdatedAt = timex.Now()

		content := m.Content
		m.Content = ""             // Do not store content in database // 不在数据库存储内容
		m.ContentLastSnapshot = "" // Do not store snapshot in database // 不在数据库存储快照

		createErr = u.WithContext(ctx).Create(m)
		if createErr != nil {
			return createErr
		}

		// Save content to file
		// 保存内容到文件
		folder := r.dao.GetNoteFolderPath(uid, m.ID)
		if err := r.dao.SaveContentToFile(folder, "content.txt", content); err != nil {
			return err
		}

		// 更新 FTS 索引
		r.upsertFTS(m, content, uid)

		noteRes, err := r.toDomain(m, uid)
		if err != nil {
			return err
		}
		result = noteRes

		result.Content = content
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, createErr
}

// Update updates a note
// Update 更新笔记
func (r *noteRepository) Update(ctx context.Context, note *domain.Note, uid int64) (*domain.Note, error) {
	var result *domain.Note
	var updateErr error

	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note
		m := r.toModel(note)

		m.UpdatedTimestamp = timex.Now().UnixMilli()
		m.UpdatedAt = timex.Now()

		content := m.Content
		m.Content = "" // Do not update content in database // 不在数据库更新内容

		updateErr = u.WithContext(ctx).Where(
			u.ID.Eq(m.ID),
		).Select(
			u.ID,
			u.VaultID,
			u.Action,
			u.Rename,
			u.Path,
			u.PathHash,
			u.Content,
			u.ContentHash,
			u.ClientName,
			u.ClientType,
			u.ClientVersion,
			u.Size,
			u.Ctime,
			u.Mtime,
			u.Version,
			u.UpdatedAt,
			u.UpdatedTimestamp,
			u.FID,
		).Save(m)

		if updateErr != nil {
			return updateErr
		}

		// Save content to file
		// 保存内容到文件
		folder := r.dao.GetNoteFolderPath(uid, m.ID)
		if err := r.dao.SaveContentToFile(folder, "content.txt", content); err != nil {
			return err
		}

		// 更新 FTS 索引
		r.upsertFTS(m, content, uid)

		noteRes, err := r.toDomain(m, uid)
		if err != nil {
			return err
		}
		result = noteRes

		result.Content = content
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, updateErr
}

// UpdateDelete updates note to deleted status
// UpdateDelete 更新笔记为删除状态
func (r *noteRepository) UpdateDelete(ctx context.Context, note *domain.Note, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note
		m := &model.Note{
			ID:               note.ID,
			Action:           string(note.Action),
			Rename:           note.Rename,
			ClientName:       note.ClientName,
			ClientType:       note.ClientType,
			ClientVersion:    note.ClientVersion,
			Mtime:            note.Mtime,
			UpdatedTimestamp: timex.Now().UnixMilli(),
		}

		err := u.WithContext(ctx).Where(
			u.ID.Eq(m.ID),
		).Select(
			u.ID,
			u.Action,
			u.Rename,
			u.ClientName,
			u.ClientType,
			u.ClientVersion,
			u.Mtime,
			u.UpdatedTimestamp,
		).Save(m)
		if err == nil {
			// 把实际写入的 UpdatedTimestamp 回写到调用方的 note 上，
			// 使调用方无需重新查库即可拿到写入后的准确值
			// Write the actually-persisted UpdatedTimestamp back onto the caller's note,
			// so the caller doesn't need a re-query to get the post-write value
			note.UpdatedTimestamp = m.UpdatedTimestamp
		}
		return err
	})
}

// UpdateMtime updates note modification time
// UpdateMtime 更新笔记修改时间
func (r *noteRepository) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note

		_, err := u.WithContext(ctx).Where(
			u.ID.Eq(id),
		).UpdateSimple(
			u.Mtime.Value(mtime),
			u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
			u.UpdatedAt.Value(timex.Now()),
		)
		return err
	})
}

// UpdateActionMtime updates note modification time
// UpdateActionMtime 更新笔记修改时间
func (r *noteRepository) UpdateActionMtime(ctx context.Context, action domain.NoteAction, mtime int64, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note

		_, err := u.WithContext(ctx).Where(
			u.ID.Eq(id),
		).UpdateSimple(
			u.Action.Value(string(action)),
			u.Mtime.Value(mtime),
			u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
			u.UpdatedAt.Value(timex.Now()),
		)
		return err
	})
}

// UpdateSnapshot updates note snapshot
// UpdateSnapshot 更新笔记快照
func (r *noteRepository) UpdateSnapshot(ctx context.Context, snapshot, snapshotHash string, version, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note

		// 保存快照到文件
		folder := r.dao.GetNoteFolderPath(uid, id)
		if err := r.dao.SaveContentToFile(folder, "snapshot.txt", snapshot); err != nil {
			return err
		}

		_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).UpdateSimple(
			u.ContentLastSnapshot.Value(""),
			u.ContentLastSnapshotHash.Value(snapshotHash),
			u.Version.Value(version),
		)
		return err
	})
}

// Delete physically deletes a note
// Delete 物理删除笔记
func (r *noteRepository) Delete(ctx context.Context, id, vaultID, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note

		// 在物理删除之前清理全文检索索引
		r.deleteFTS(id, vaultID, uid)

		_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).Delete()
		if err != nil {
			return err
		}

		// Delete physical files
		// 删除物理文件
		folder := r.dao.GetNoteFolderPath(uid, id)
		_ = r.dao.RemoveContentFolder(folder)

		return nil
	})
}

// DeletePhysicalByTime physically deletes notes marked as deleted by time
// DeletePhysicalByTime 根据时间物理删除已标记删除的笔记
func (r *noteRepository) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note

		// 先找到要删除的 ID
		list, _ := u.WithContext(ctx).Where(
			u.Action.Eq("delete"),
			u.UpdatedTimestamp.Lt(timestamp),
		).Select(u.ID, u.VaultID).Find()

		for _, m := range list {
			r.deleteFTS(m.ID, m.VaultID, uid)
		}

		_, err := u.WithContext(ctx).Where(
			u.Action.Eq("delete"),
			u.UpdatedTimestamp.Lt(timestamp),
		).Delete()

		if err == nil {
			for _, m := range list {
				folder := r.dao.GetNoteFolderPath(uid, m.ID)
				_ = r.dao.RemoveContentFolder(folder)
			}
		}
		return err
	})
}

// DeletePhysicalByTimeAll physically deletes notes marked as deleted for all users by time
// DeletePhysicalByTimeAll 根据时间物理删除所有用户的已标记删除的笔记
func (r *noteRepository) DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error {
	// Get all user UIDs
	// 获取所有用户 UID
	uids, err := r.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}

	// Execute cleanup user by user
	// 逐用户执行清理
	for i, uid := range uids {
		// 增加错峰延迟，避免瞬间触发大量写事务
		if i > 0 {
			time.Sleep(500 * time.Millisecond)
		}
		if err := r.DeletePhysicalByTime(ctx, timestamp, uid); err != nil {
			// 记录错误但继续处理其他用户
			continue
		}
	}
	return nil
}

// List retrieves note list by page
// List 分页获取笔记列表
func (r *noteRepository) List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool, sortBy string, sortOrder string, paths []string) ([]*domain.Note, error) {
	u := r.note(uid).Note
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
	)

	if isRecycle {
		q = q.Where(u.Action.Eq("delete"), u.Rename.Eq(0))
	} else {
		q = q.Where(u.Action.Neq("delete"))
	}

	// 构建排序语句
	orderClause := buildOrderClause(sortBy, sortOrder)

	var modelList []*model.Note
	var err error

	if len(paths) > 0 {
		// 精确路径列表查询（分享筛选模式），忽略 keyword
		err = q.UnderlyingDB().Where("path IN ?", paths).
			Order(orderClause).
			Limit(pageSize).
			Offset(app.GetPageOffset(page, pageSize)).
			Find(&modelList).Error
	} else if keyword != "" {
		// 内容搜索模式：使用 Bleve 全文搜索
		if searchMode == "content" && r.dao.BleveMgr.IsEnabled() {
			// 确保 FTS 索引存在
			_ = r.EnsureFTSIndex(ctx, uid)

			noteIDs, ftsErr := r.searchFTS(uid, vaultID, keyword, isRecycle, sortBy, sortOrder, pageSize, app.GetPageOffset(page, pageSize))
			if ftsErr != nil {
				return nil, ftsErr
			}

			if len(noteIDs) == 0 {
				return []*domain.Note{}, nil
			}

			// 根据 FTS 返回的 ID 查询完整笔记，保持 FTS 返回的顺序
			err = q.UnderlyingDB().Where("id IN ?", noteIDs).Order(orderClause).Find(&modelList).Error
		} else {
			// 路径搜索：使用 LIKE
			key := "%" + keyword + "%"
			err = q.UnderlyingDB().Where("path LIKE ?", key).
				Order(orderClause).
				Limit(pageSize).
				Offset(app.GetPageOffset(page, pageSize)).
				Find(&modelList).Error
		}
	} else {
		err = q.UnderlyingDB().
			Order(orderClause).
			Limit(pageSize).
			Offset(app.GetPageOffset(page, pageSize)).
			Find(&modelList).Error
	}

	if err != nil {
		return nil, err
	}

	var list []*domain.Note
	for _, m := range modelList {
		note, err := r.toDomain(m, uid)
		if err != nil {
			return nil, err
		}
		list = append(list, note)

	}
	return list, nil
}

func (r *noteRepository) ListByPathPrefix(ctx context.Context, pathPrefix string, vaultID, uid int64) ([]*domain.Note, error) {
	u := r.note(uid).Note
	// Use LIKE 'prefix/%'
	// 使用 LIKE 'prefix/%'
	pattern := pathPrefix + "/%"
	ms, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Path.Like(pattern),
		u.Action.Neq("delete"),
	).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.Note
	for _, m := range ms {
		if !isPathWithinPrefix(m.Path, pathPrefix) {
			continue
		}
		note, err := r.toDomain(m, uid)
		if err != nil {
			return nil, err
		}
		res = append(res, note)
	}
	return res, nil
}

// getSortField maps sort fields
// getSortField 映射排序字段
func getSortField(sortBy string) string {
	switch sortBy {
	case "ctime":
		return "ctime"
	case "path":
		return "path"
	default:
		return "mtime"
	}
}

// buildOrderClause builds order clause
// buildOrderClause 构建排序语句
func buildOrderClause(sortBy, sortOrder string) string {
	// 默认值
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// 验证排序方向
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return getSortField(sortBy) + " " + sortOrder
}

// ListCount retrieves note count
// ListCount 获取笔记数量
func (r *noteRepository) ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool, paths []string) (int64, error) {
	u := r.note(uid).Note
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
	)

	if isRecycle {
		q = q.Where(u.Action.Eq("delete"), u.Rename.Eq(0))
	} else {
		q = q.Where(u.Action.Neq("delete"))
	}

	var count int64
	var err error

	if len(paths) > 0 {
		// 精确路径列表计数（分享筛选模式）
		err = q.UnderlyingDB().Where("path IN ?", paths).Count(&count).Error
	} else if keyword != "" {
		// 内容搜索模式：使用 Bleve 全文搜索
		if searchMode == "content" && r.dao.BleveMgr.IsEnabled() {
			count, err = r.searchFTSCount(uid, vaultID, keyword, isRecycle)
		} else {
			// 路径搜索：使用 LIKE
			key := "%" + keyword + "%"
			err = q.UnderlyingDB().Where("path LIKE ?", key).Count(&count).Error
		}
	} else {
		count, err = q.Order(u.CreatedAt).Count()
	}

	if err != nil {
		return 0, err
	}

	return count, nil
}

// ListByUpdatedTimestamp retrieves note list by updated timestamp
// ListByUpdatedTimestamp 根据更新时间戳获取笔记列表
func (r *noteRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Note, error) {
	return r.ListByUpdatedTimestampPage(ctx, timestamp, vaultID, uid, 0, 0)
}

// ListByUpdatedTimestampPage retrieves note list by updated timestamp by page
// ListByUpdatedTimestampPage 根据更新时间戳分页获取笔记列表
func (r *noteRepository) ListByUpdatedTimestampPage(ctx context.Context, timestamp, vaultID, uid int64, offset, limit int) ([]*domain.Note, error) {
	u := r.note(uid).Note
	query := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.UpdatedTimestamp.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc())

	var mList []*model.Note
	var err error
	if limit > 0 {
		mList, _, err = query.FindByPage(offset, limit)
	} else {
		mList, err = query.Find()
	}

	if err != nil {
		return nil, err
	}

	var list []*domain.Note
	for _, m := range mList {
		note, err := r.toDomain(m, uid)
		if err != nil {
			return nil, err
		}
		list = append(list, note)

	}
	return list, nil
}

// ListByPathHashesMeta retrieves note metadata (no content) for a batch of path hashes in a
// single query, including all statuses (e.g. soft-deleted). Used for batch existence
// pre-checks (e.g. before deleting a batch of client-reported notes) to avoid N+1
// per-item lookups.
// ListByPathHashesMeta 单次查询批量获取一组路径哈希对应的笔记元数据（不读正文），包含所有状态
// （含已软删除）。用于批量存在性预检查（例如批量处理客户端删除上报前），避免逐条查询的 N+1。
func (r *noteRepository) ListByPathHashesMeta(ctx context.Context, pathHashes []string, vaultID, uid int64) (map[string]*domain.Note, error) {
	result := make(map[string]*domain.Note, len(pathHashes))
	if len(pathHashes) == 0 {
		return result, nil
	}

	u := r.note(uid).Note
	ms, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.In(pathHashes...),
	).Find()
	if err != nil {
		return nil, err
	}

	for _, m := range ms {
		n := r.toDomainMeta(m)
		// 同一 pathHash 可能存在历史遗留的重复记录，保留 UpdatedTimestamp 最新的一条
		// A pathHash may have legacy duplicate rows; keep the one with the latest UpdatedTimestamp
		if existing, ok := result[n.PathHash]; !ok || n.UpdatedTimestamp > existing.UpdatedTimestamp {
			result[n.PathHash] = n
		}
	}
	return result, nil
}

// ListByUpdatedTimestampMeta retrieves note metadata list by updated timestamp, skipping the
// content/snapshot file reads (content.txt/snapshot.txt). Used by the sync-download diff path,
// which only needs ContentHash/Mtime for comparison and reads full content on demand for the
// small subset of notes that actually need to be sent.
// ListByUpdatedTimestampMeta 根据更新时间戳获取笔记元数据列表，跳过正文/快照文件读取
// （content.txt/snapshot.txt）。用于同步下发的差量比对路径——该路径只需要 ContentHash/Mtime
// 做比对，真正需要下发的少数笔记再按需读取正文。
func (r *noteRepository) ListByUpdatedTimestampMeta(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Note, error) {
	return r.ListByUpdatedTimestampPageMeta(ctx, timestamp, vaultID, uid, 0, 0)
}

// ListByUpdatedTimestampPageMeta is the paged variant of ListByUpdatedTimestampMeta.
// ListByUpdatedTimestampPageMeta 是 ListByUpdatedTimestampMeta 的分页变体。
func (r *noteRepository) ListByUpdatedTimestampPageMeta(ctx context.Context, timestamp, vaultID, uid int64, offset, limit int) ([]*domain.Note, error) {
	u := r.note(uid).Note
	query := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.UpdatedTimestamp.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc())

	var mList []*model.Note
	var err error
	if limit > 0 {
		mList, _, err = query.FindByPage(offset, limit)
	} else {
		mList, err = query.Find()
	}

	if err != nil {
		return nil, err
	}

	list := make([]*domain.Note, 0, len(mList))
	for _, m := range mList {
		list = append(list, r.toDomainMeta(m))
	}
	return list, nil
}

// ListContentUnchanged retrieves note list with unchanged content
// ListContentUnchanged 获取内容未变更的笔记列表
func (r *noteRepository) ListContentUnchanged(ctx context.Context, uid int64) ([]*domain.Note, error) {
	u := r.note(uid).Note
	var mList []*model.Note

	err := u.WithContext(ctx).UnderlyingDB().Where(
		"action != ?", "delete",
	).Where("content_hash != content_last_snapshot_hash").
		Find(&mList).Error

	if err != nil {
		return nil, err
	}

	var list []*domain.Note
	for _, m := range mList {
		note, err := r.toDomain(m, uid)
		if err != nil {
			return nil, err
		}
		list = append(list, note)

	}
	return list, nil
}

// CountSizeSum 获取笔记数量和大小总和
func (r *noteRepository) CountSizeSum(ctx context.Context, vaultID, uid int64) (*domain.CountSizeResult, error) {
	u := r.note(uid).Note

	result := &struct {
		Size  int64
		Count int64
	}{}

	err := u.WithContext(ctx).Select(u.Size.Sum().As("size"), u.Size.Count().As("count")).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
		u.Rename.Eq(0),
	).Scan(result)

	if err != nil {
		return nil, err
	}

	return &domain.CountSizeResult{
		Count: result.Count,
		Size:  result.Size,
	}, nil
}

// ListByFID 根据文件夹ID获取笔记列表
func (r *noteRepository) ListByFID(ctx context.Context, fid, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.Note, error) {
	u := r.note(uid).Note
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.FID.Eq(fid),
		u.Action.Neq("delete"),
	)

	// 构建排序语句
	orderClause := buildOrderClause(sortBy, sortOrder)

	var modelList []*model.Note
	err := q.UnderlyingDB().
		Order(orderClause).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find(&modelList).Error

	if err != nil {
		return nil, err
	}

	var list []*domain.Note
	for _, m := range modelList {
		note, err := r.toDomain(m, uid)
		if err != nil {
			return nil, err
		}
		list = append(list, note)

	}
	return list, nil
}

// ListByFIDCount 根据文件夹ID获取笔记数量
func (r *noteRepository) ListByFIDCount(ctx context.Context, fid, vaultID, uid int64) (int64, error) {
	u := r.note(uid).Note
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.FID.Eq(fid),
		u.Action.Neq("delete"),
	)

	return q.Count()
}

func (r *noteRepository) ListByFIDs(ctx context.Context, fids []int64, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.Note, error) {
	u := r.note(uid).Note
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.FID.In(fids...),
		u.Action.Neq("delete"),
	)

	orderClause := buildOrderClause(sortBy, sortOrder)

	var modelList []*model.Note
	err := q.UnderlyingDB().
		Order(orderClause).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find(&modelList).Error

	if err != nil {
		return nil, err
	}

	var list []*domain.Note
	for _, m := range modelList {
		note, err := r.toDomain(m, uid)
		if err != nil {
			return nil, err
		}
		list = append(list, note)

	}
	return list, nil
}

func (r *noteRepository) ListByFIDsCount(ctx context.Context, fids []int64, vaultID, uid int64) (int64, error) {
	u := r.note(uid).Note
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.FID.In(fids...),
		u.Action.Neq("delete"),
	)

	return q.Count()
}

// CountByFIDs 按文件夹 ID 分组统计笔记数量，一次查询取回所有传入 fid 的计数
// （用于替代对每个文件夹单独调用 ListByFIDCount 造成的 N+1）
// CountByFIDs groups by folder ID and returns note counts for all given fids in a single
// query (replaces calling ListByFIDCount once per folder, which is N+1).
func (r *noteRepository) CountByFIDs(ctx context.Context, fids []int64, vaultID, uid int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(fids))
	if len(fids) == 0 {
		return result, nil
	}

	u := r.note(uid).Note
	// 显式 column tag：GORM 的默认命名转换会把 "FID" 猜成 "f_id" 而不是实际列名 "fid"，
	// 不加 tag 会导致 Scan 后 FID 全部读成 0（已由单测捕获）。
	// Explicit column tags: GORM's default naming convention guesses "FID" as "f_id"
	// instead of the actual "fid" column, silently scanning FID back as 0 without the
	// tag (caught by a unit test).
	var rows []struct {
		FID   int64 `gorm:"column:fid"`
		Count int64 `gorm:"column:count"`
	}
	err := u.WithContext(ctx).Select(u.FID, u.FID.Count().As("count")).Where(
		u.VaultID.Eq(vaultID),
		u.FID.In(fids...),
		u.Action.Neq("delete"),
	).Group(u.FID).Scan(&rows)

	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.FID] = row.Count
	}
	return result, nil
}

// RecycleClear 清理回收站
func (r *noteRepository) RecycleClear(ctx context.Context, path, pathHash string, vaultID, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note
		q := u.WithContext(ctx).Where(u.VaultID.Eq(vaultID), u.Action.Eq(string(domain.NoteActionDelete)), u.Rename.Eq(0))
		if pathHash != "" {
			q = q.Where(u.PathHash.Eq(pathHash))
		}
		_, err := q.UpdateSimple(
			u.Rename.Value(2),
			u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
			u.UpdatedAt.Value(timex.Now()),
		)
		return err
	})
}

// UpdateFID 仅更新笔记的文件夹关联 ID，不更新 updated_timestamp
// 用于 SyncResourceFID 内部整理，避免污染增量同步时间戳
// Only updates the folder ID (FID) without touching updated_timestamp
// Used by SyncResourceFID to avoid polluting incremental sync timestamps
func (r *noteRepository) UpdateFID(ctx context.Context, id, fid, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note
		_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).UpdateSimple(u.FID.Value(fid))
		return err
	})
}

// 确保 noteRepository 实现了 domain.NoteRepository 接口
var _ domain.NoteRepository = (*noteRepository)(nil)

// upsertFTS asynchronously queues an update of the Bleve FTS index
// upsertFTS 异步投递一次 Bleve FTS 索引更新
// m 为调用方刚写入/更新的笔记记录，所需字段均已具备，无需重新查库；实际的索引写入由
// BleveManager 后台 worker 攒批异步执行，本方法只投递不等待，不阻塞写队列关键路径
// m is the note record just written/updated by the caller; all needed fields are already
// present, no need to re-query. The actual index write is batched and executed
// asynchronously by BleveManager's background worker; this method only enqueues and
// does not block the write-queue's critical path.
func (r *noteRepository) upsertFTS(m *model.Note, content string, uid int64) {
	doc := BleveNoteDoc{
		ID:      strconv.FormatInt(m.ID, 10),
		Path:    m.Path,
		PathRaw: m.Path,
		Content: content,
		Action:  m.Action,
		Rename:  float64(m.Rename),
		Ctime:   float64(m.Ctime),
		Mtime:   float64(m.Mtime),
	}

	r.dao.BleveMgr.EnqueueUpsert(uid, m.VaultID, doc)
}

// deleteFTS asynchronously queues a note delete from the Bleve FTS index
// deleteFTS 异步投递一次笔记从 Bleve FTS 索引中删除
// vaultID 由调用方直接传入（写路径在删除前已知该笔记所属仓库），无需为此再做一次冗余 SELECT
// vaultID is passed in directly by the caller (the write path already knows the note's
// vault before deleting it), avoiding a redundant SELECT just to look it up here.
func (r *noteRepository) deleteFTS(noteID, vaultID, uid int64) {
	r.dao.BleveMgr.EnqueueDelete(uid, vaultID, strconv.FormatInt(noteID, 10))
}

// searchFTS uses Bleve to search note IDs, returning matching ID slice
// searchFTS 使用 Bleve 搜索内容，返回匹配的 note_id 列表
func (r *noteRepository) searchFTS(uid, vaultID int64, keyword string, isRecycle bool, sortBy, sortOrder string, limit, offset int) ([]int64, error) {
	index, err := r.dao.BleveMgr.GetIndex(uid, vaultID)
	if err != nil {
		return nil, err
	}

	// Rebuild if empty
	// 如果索引为空则自动重建
	var docCount uint64
	if res, err := index.DocCount(); err == nil {
		docCount = res
	}
	if docCount == 0 {
		_ = r.RebuildVaultIndex(context.Background(), uid, vaultID)
		if idxNew, err := r.dao.BleveMgr.GetIndex(uid, vaultID); err == nil {
			index = idxNew
		}
	}

	// Construct boolean query
	// 构造布尔组合查询
	var actionQuery bleveQuery.Query
	if isRecycle {
		actionTermQuery := bleve.NewTermQuery("delete")
		actionTermQuery.SetField("action")

		renameRangeQuery := bleve.NewNumericRangeQuery(util.Ptr(-0.5), util.Ptr(0.5))
		renameRangeQuery.SetField("rename")

		actionQuery = bleve.NewConjunctionQuery(
			actionTermQuery,
			renameRangeQuery,
		)
	} else {
		actionTermQuery := bleve.NewTermQuery("delete")
		actionTermQuery.SetField("action")

		boolQuery := bleve.NewBooleanQuery()
		boolQuery.AddMustNot(actionTermQuery)
		actionQuery = boolQuery
	}

	pathQuery := bleve.NewMatchQuery(keyword)
	pathQuery.SetField("path")
	pathQuery.Operator = bleveQuery.MatchQueryOperatorAnd

	contentQuery := bleve.NewMatchQuery(keyword)
	contentQuery.SetField("content")
	contentQuery.Operator = bleveQuery.MatchQueryOperatorAnd

	query := bleve.NewConjunctionQuery(
		bleve.NewDisjunctionQuery(
			pathQuery,
			contentQuery,
		),
		actionQuery,
	)

	req := bleve.NewSearchRequest(query)
	req.Size = limit
	req.From = offset

	// Sort mapping
	// 排序映射
	if sortOrder == "" {
		sortOrder = "desc"
	}
	isDesc := sortOrder == "desc"

	sortField := getSortField(sortBy)
	if sortField == "path" {
		sortField = "path_raw"
	}

	sortFieldPrefix := ""
	if isDesc {
		sortFieldPrefix = "-"
	}

	req.SortBy([]string{sortFieldPrefix + sortField})

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
	r.dao.Logger().Info("FTS search execution",
		zap.String("keyword", keyword),
		zap.Int64("uid", uid),
		zap.Int64("vaultID", vaultID),
		zap.Int64s("results", noteIDs),
		zap.Int("total_hits", int(res.Total)),
	)

	return noteIDs, nil
}

// searchFTSCount returns search matches count
// searchFTSCount 返回全文搜索匹配计数
func (r *noteRepository) searchFTSCount(uid, vaultID int64, keyword string, isRecycle bool) (int64, error) {
	index, err := r.dao.BleveMgr.GetIndex(uid, vaultID)
	if err != nil {
		return 0, err
	}

	// Construct query
	var actionQuery bleveQuery.Query
	if isRecycle {
		actionTermQuery := bleve.NewTermQuery("delete")
		actionTermQuery.SetField("action")

		renameRangeQuery := bleve.NewNumericRangeQuery(util.Ptr(-0.5), util.Ptr(0.5))
		renameRangeQuery.SetField("rename")

		actionQuery = bleve.NewConjunctionQuery(
			actionTermQuery,
			renameRangeQuery,
		)
	} else {
		actionTermQuery := bleve.NewTermQuery("delete")
		actionTermQuery.SetField("action")

		boolQuery := bleve.NewBooleanQuery()
		boolQuery.AddMustNot(actionTermQuery)
		actionQuery = boolQuery
	}

	pathQuery := bleve.NewMatchQuery(keyword)
	pathQuery.SetField("path")
	pathQuery.Operator = bleveQuery.MatchQueryOperatorAnd

	contentQuery := bleve.NewMatchQuery(keyword)
	contentQuery.SetField("content")
	contentQuery.Operator = bleveQuery.MatchQueryOperatorAnd

	query := bleve.NewConjunctionQuery(
		bleve.NewDisjunctionQuery(
			pathQuery,
			contentQuery,
		),
		actionQuery,
	)

	req := bleve.NewSearchRequest(query)
	req.Size = 0 // Count only // 仅计数
	res, err := index.Search(req)
	if err != nil {
		return 0, err
	}

	return int64(res.Total), nil
}

// RebuildVaultIndex rebuilds index from database and file contents for a specific vault
// RebuildVaultIndex 从数据库和物理文件内容重建指定仓库的索引
func (r *noteRepository) RebuildVaultIndex(ctx context.Context, uid, vaultID int64) error {
	_ = r.dao.BleveMgr.DeleteIndex(uid, vaultID)

	index, err := r.dao.BleveMgr.GetIndex(uid, vaultID)
	if err != nil {
		return err
	}

	db := r.dao.ResolveDB(r.GetKey(uid))
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

// DeleteByVaultID physically deletes all notes in a vault
// DeleteByVaultID 物理删除仓库下的所有笔记
func (r *noteRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note

		// 查找该仓库下的所有笔记 ID
		notes, err := u.WithContext(ctx).Where(u.VaultID.Eq(vaultID)).Select(u.ID).Find()
		if err != nil {
			return err
		}

		if len(notes) == 0 {
			return nil
		}

		var ids []int64
		for _, n := range notes {
			ids = append(ids, n.ID)
		}

		// 从数据库删除
		_, err = u.WithContext(ctx).Where(u.VaultID.Eq(vaultID)).Delete()
		if err != nil {
			return err
		}

		// 删除物理文件夹
		for _, id := range ids {
			folder := r.dao.GetNoteFolderPath(uid, id)
			_ = r.dao.RemoveContentFolder(folder)
		}

		return nil
	})
}

// Ensure noteRepository implements domain.NoteRepository interface
// 确保 noteRepository 实现了 domain.NoteRepository 接口
var _ domain.NoteRepository = (*noteRepository)(nil)
