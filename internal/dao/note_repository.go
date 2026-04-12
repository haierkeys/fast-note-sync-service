// Package dao 实现数据访问层
package dao

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

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

// noteRepository 实现 domain.NoteRepository 接口
type noteRepository struct {
	dao             *Dao
	customPrefixKey string
}

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
		// 初始化通用全文搜索表
		_ = model.CreateNoteFTSTable(g)
	}, r.GetKey(uid)+"#note_v3", r.GetKey(uid))
}

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

// EnsureFTSIndex 确保 FTS 索引存在（公开方法，可手动调用）
func (r *noteRepository) EnsureFTSIndex(ctx context.Context, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		key := r.GetKey(uid)
		ftsKey := key + "#fts_indexed"

		// 使用 onceKeys 确保每个用户只检查一次
		if _, loaded := r.dao.onceKeys.LoadOrStore(ftsKey, true); loaded {
			return nil // 已检查过
		}

		// 确保 FTS 表存在（会自动检查版本并重建）
		_ = model.CreateNoteFTSTable(db)

		// 检查 FTS 索引是否为空
		var ftsCount int64
		db.Model(&model.NoteFTSToken{}).Count(&ftsCount)
		if ftsCount > 0 {
			return nil // 已有索引
		}

		// 检查是否有笔记需要索引
		var noteCount int64
		db.Model(&model.Note{}).Where("action != ?", "delete").Count(&noteCount)
		if noteCount == 0 {
			return nil
		}

		// 同步重建索引
		var notes []model.Note
		if err := db.Where("action != ?", "delete").Find(&notes).Error; err != nil {
			return err
		}

		for _, note := range notes {
			folder := r.dao.GetNoteFolderPath(uid, note.ID)
			content, exists, err := r.dao.LoadContentFromFile(folder, "content.txt")
			if err != nil {
				return err
			}
			if !exists {
				content = ""
			}
			r.upsertFTS(db, note.ID, note.Path, content)
		}

		return nil
	})
}

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
		Size:                    note.Size,
		Ctime:                   note.Ctime,
		Mtime:                   note.Mtime,
		UpdatedTimestamp:        note.UpdatedTimestamp,
		CreatedAt:               timex.Time(note.CreatedAt),
		UpdatedAt:               timex.Time(note.UpdatedAt),
	}
}

// fillNoteContent 填充笔记内容
func (r *noteRepository) fillNoteContent(uid int64, n *domain.Note) error {
	if n == nil {
		return nil
	}
	folder := r.dao.GetNoteFolderPath(uid, n.ID)

	// 加载内容
	content, exists, err := r.dao.LoadContentFromFile(folder, "content.txt")
	if err != nil {
		return err
	}
	if exists {
		n.Content = content
	} else if n.Content != "" {
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
		// 文件不存在且没有可迁移的内容，返回错误以防止数据丢失（视为读取失败）
		return fmt.Errorf("note content file not found: %w", os.ErrNotExist)
	}

	// 加载快照
	snapshot, exists, err := r.dao.LoadContentFromFile(folder, "snapshot.txt")
	if err != nil {
		return err
	}
	if exists {
		n.ContentLastSnapshot = snapshot
	} else if n.ContentLastSnapshot != "" {
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

// GetByID 根据ID获取笔记
func (r *noteRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Note, error) {
	u := r.note(uid).Note
	m, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid)
}

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
		m.Content = ""             // 不在数据库存储内容
		m.ContentLastSnapshot = "" // 不在数据库存储快照

		createErr = u.WithContext(ctx).Create(m)
		if createErr != nil {
			return createErr
		}

		// 保存内容到文件
		folder := r.dao.GetNoteFolderPath(uid, m.ID)
		if err := r.dao.SaveContentToFile(folder, "content.txt", content); err != nil {
			return err
		}

		// 更新 FTS 索引
		r.upsertFTS(db, m.ID, m.Path, content)

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
		m.Content = "" // 不在数据库更新内容

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

		// 保存内容到文件
		folder := r.dao.GetNoteFolderPath(uid, m.ID)
		if err := r.dao.SaveContentToFile(folder, "content.txt", content); err != nil {
			return err
		}

		// 更新 FTS 索引
		r.upsertFTS(db, m.ID, m.Path, content)

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

// UpdateDelete 更新笔记为删除状态
func (r *noteRepository) UpdateDelete(ctx context.Context, note *domain.Note, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note
		m := &model.Note{
			ID:               note.ID,
			Action:           string(note.Action),
			Rename:           note.Rename,
			ClientName:       note.ClientName,
			Mtime:            note.Mtime,
			UpdatedTimestamp: timex.Now().UnixMilli(),
		}

		return u.WithContext(ctx).Where(
			u.ID.Eq(m.ID),
		).Select(
			u.ID,
			u.Action,
			u.Rename,
			u.ClientName,
			u.Mtime,
			u.UpdatedTimestamp,
		).Save(m)
	})
}

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

// Delete 物理删除笔记
func (r *noteRepository) Delete(ctx context.Context, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note
		_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).Delete()
		if err != nil {
			return err
		}

		// 删除物理文件
		folder := r.dao.GetNoteFolderPath(uid, id)
		_ = r.dao.RemoveContentFolder(folder)

		return nil
	})
}

// DeletePhysicalByTime 根据时间物理删除已标记删除的笔记
func (r *noteRepository) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.note(uid).Note

		// 先找到要删除的 ID
		list, _ := u.WithContext(ctx).Where(
			u.Action.Eq("delete"),
			u.UpdatedTimestamp.Lt(timestamp),
		).Select(u.ID).Find()

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

// DeletePhysicalByTimeAll 根据时间物理删除所有用户的已标记删除的笔记
func (r *noteRepository) DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error {
	// 获取所有用户 UID
	uids, err := r.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}

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
		// 内容搜索模式：使用 FTS5 全文搜索
		if searchMode == "content" {
			// 使用干净的 DB 连接执行 FTS 查询，避免继承 note 表上下文导致 JOIN 二义性
			ftsDB := r.dao.ResolveDB(r.GetKey(uid))

			// 确保 FTS 索引存在
			r.EnsureFTSIndex(ctx, uid)

			noteIDs, ftsErr := r.searchFTS(ftsDB, keyword, vaultID, isRecycle, sortBy, sortOrder, pageSize, app.GetPageOffset(page, pageSize))
			if ftsErr != nil {
				return nil, ftsErr
			}

			if len(noteIDs) == 0 {
				return []*domain.Note{}, nil
			}

			// 根据 FTS 返回的 ID 查询完整笔记，保持 FTS 返回的顺序
			err = q.UnderlyingDB().Where("id IN ?", noteIDs).Order(orderClause).Find(&modelList).Error
		} else {
			// 路径搜索或正则搜索：使用 LIKE
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
		note, err := r.toDomain(m, uid)
		if err != nil {
			return nil, err
		}
		res = append(res, note)
	}
	return res, nil
}

// buildOrderClause 构建排序语句
func buildOrderClause(sortBy, sortOrder string) string {
	// 默认值
	if sortBy == "" {
		sortBy = "mtime"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// 验证排序方向
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// 映射排序字段
	var field string
	switch sortBy {
	case "ctime":
		field = "ctime"
	case "path":
		field = "path"
	case "mtime":
		fallthrough
	default:
		field = "mtime"
	}

	return field + " " + sortOrder
}

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
		// 内容搜索模式：使用 FTS5 全文搜索
		if searchMode == "content" {
			// 使用干净的 DB 连接，避免二义性
			ftsDB := r.dao.ResolveDB(r.GetKey(uid))
			count, err = r.searchFTSCount(ftsDB, keyword, vaultID, isRecycle)
		} else {
			// 路径搜索或正则搜索：使用 LIKE
			var whereClause string
			var args []interface{}

			switch searchMode {
			case "regex":
				key := "%" + keyword + "%"
				whereClause = "path LIKE ?"
				args = []interface{}{key}
			default:
				key := "%" + keyword + "%"
				whereClause = "path LIKE ?"
				args = []interface{}{key}
			}
			err = q.UnderlyingDB().Where(whereClause, args...).Count(&count).Error
		}
	} else {
		count, err = q.Order(u.CreatedAt).Count()
	}

	if err != nil {
		return 0, err
	}

	return count, nil
}

// ListByUpdatedTimestamp 根据更新时间戳获取笔记列表
func (r *noteRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Note, error) {
	return r.ListByUpdatedTimestampPage(ctx, timestamp, vaultID, uid, 0, 0)
}

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

// upsertFTS 更新 FTS 索引
func (r *noteRepository) upsertFTS(db *gorm.DB, noteID int64, path, content string) {
	// 1. 更新快照表
	db.Save(&model.NoteFTS{NoteID: noteID, Path: path, Content: content})

	// 2. 更新倒排索引
	db.Where("note_id = ?", noteID).Delete(&model.NoteFTSToken{})
	tokens := util.Tokenize(path + " " + content)
	if len(tokens) == 0 {
		return
	}
	var tokenModels []model.NoteFTSToken
	for _, t := range tokens {
		tokenModels = append(tokenModels, model.NoteFTSToken{NoteID: noteID, Token: t})
	}
	db.CreateInBatches(tokenModels, 500)
}

// deleteFTS 删除 FTS 索引
func (r *noteRepository) deleteFTS(db *gorm.DB, noteID int64) {
	db.Where("note_id = ?", noteID).Delete(&model.NoteFTS{})
	db.Where("note_id = ?", noteID).Delete(&model.NoteFTSToken{})
}

// searchFTS 使用倒排索引搜索内容，返回匹配的 note_id 列表
func (r *noteRepository) searchFTS(db *gorm.DB, keyword string, vaultID int64, isRecycle bool, sortBy, sortOrder string, limit, offset int) ([]int64, error) {
	tokens := util.Tokenize(keyword)
	if len(tokens) == 0 {
		return nil, nil
	}

	var noteIDs []int64

	// 构建 action 条件
	actionCond := "note.action != 'delete'"
	if isRecycle {
		actionCond = "note.action = 'delete' AND note.rename = 0"
	}

	// 构建排序语句
	orderClause := "note." + buildOrderClause(sortBy, sortOrder)

	// 使用新的倒排索引查询
	query := db.Table("note_fts_token AS t").
		Select("t.note_id").
		Joins("INNER JOIN note ON t.note_id = note.id").
		Where("t.token IN ?", tokens).
		Where("note.vault_id = ?", vaultID).
		Where(actionCond).
		Group("t.note_id").
		Having("COUNT(DISTINCT t.token) = ?", len(tokens)).
		Order(orderClause).
		Limit(limit).
		Offset(offset)

	err := query.Scan(&noteIDs).Error
	return noteIDs, err
}

// searchFTSCount 使用倒排索引搜索计数
func (r *noteRepository) searchFTSCount(db *gorm.DB, keyword string, vaultID int64, isRecycle bool) (int64, error) {
	tokens := util.Tokenize(keyword)
	if len(tokens) == 0 {
		return 0, nil
	}

	var count int64
	actionCond := "note.action != 'delete'"
	if isRecycle {
		actionCond = "note.action = 'delete' AND note.rename = 0"
	}

	subQuery := db.Table("note_fts_token AS t").
		Select("t.note_id").
		Joins("INNER JOIN note ON t.note_id = note.id").
		Where("t.token IN ?", tokens).
		Where("note.vault_id = ?", vaultID).
		Where(actionCond).
		Group("t.note_id").
		Having("COUNT(DISTINCT t.token) = ?", len(tokens))

	err := db.Table("(?) AS sub", subQuery).Count(&count).Error
	return count, err
}

