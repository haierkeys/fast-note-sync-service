// Package dao 实现数据访问层
package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
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

// note 获取笔记查询对象
func (r *noteRepository) note(uid int64) *query.Query {
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "Note")
		// 初始化 FTS5 全文搜索表并检查是否需要重建索引
		_ = model.CreateNoteFTSTable(g)
	}, r.GetKey(uid)+"#note", r.GetKey(uid))
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
		res = append(res, r.toDomain(m, uid))
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

		// 检查 FTS 表是否为空
		var ftsCount int64
		db.Raw("SELECT COUNT(*) FROM note_fts").Scan(&ftsCount)
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
			content, _, _ := r.dao.LoadContentFromFile(folder, "content.txt")
			r.upsertFTS(db, note.ID, note.Path, content)
		}

		return nil
	})
}

// toDomain 将 DAO Note 转换为领域模型
func (r *noteRepository) toDomain(m *model.Note, uid int64) *domain.Note {
	if m == nil {
		return nil
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
	r.fillNoteContent(uid, note)
	return note
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
func (r *noteRepository) fillNoteContent(uid int64, n *domain.Note) {
	if n == nil {
		return
	}
	folder := r.dao.GetNoteFolderPath(uid, n.ID)

	// 加载内容
	if content, exists, _ := r.dao.LoadContentFromFile(folder, "content.txt"); exists {
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
	}

	// 加载快照
	if snapshot, exists, _ := r.dao.LoadContentFromFile(folder, "snapshot.txt"); exists {
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
}

// GetByID 根据ID获取笔记
func (r *noteRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Note, error) {
	u := r.note(uid).Note
	m, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid), nil
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
	return r.toDomain(m, uid), nil
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
	return r.toDomain(m, uid), nil
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
	return r.toDomain(m, uid), nil
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
	return r.toDomain(m, uid), nil
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

		result = r.toDomain(m, uid)
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

		result = r.toDomain(m, uid)
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
func (r *noteRepository) List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool, sortBy string, sortOrder string) ([]*domain.Note, error) {
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

	if keyword != "" {
		// 内容搜索模式：使用 FTS5 全文搜索
		if searchMode == "content" {
			db := q.UnderlyingDB()

			// 确保 FTS 索引存在
			r.EnsureFTSIndex(ctx, uid)

			noteIDs, ftsErr := r.searchFTS(db, keyword, vaultID, isRecycle, sortBy, sortOrder, pageSize, app.GetPageOffset(page, pageSize))
			if ftsErr != nil {
				return nil, ftsErr
			}

			if len(noteIDs) == 0 {
				return []*domain.Note{}, nil
			}

			// 根据 FTS 返回的 ID 查询完整笔记，保持 FTS 返回的顺序
			err = db.Where("id IN ?", noteIDs).Order(orderClause).Find(&modelList).Error
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
		list = append(list, r.toDomain(m, uid))
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
		res = append(res, r.toDomain(m, uid))
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
func (r *noteRepository) ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool) (int64, error) {
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

	if keyword != "" {
		// 内容搜索模式：使用 FTS5 全文搜索
		if searchMode == "content" {
			db := q.UnderlyingDB()
			count, err = r.searchFTSCount(db, keyword, vaultID, isRecycle)
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
	u := r.note(uid).Note
	mList, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.UpdatedTimestamp.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*domain.Note
	for _, m := range mList {
		list = append(list, r.toDomain(m, uid))
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
		list = append(list, r.toDomain(m, uid))
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
		list = append(list, r.toDomain(m, uid))
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
		list = append(list, r.toDomain(m, uid))
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

// 确保 noteRepository 实现了 domain.NoteRepository 接口
var _ domain.NoteRepository = (*noteRepository)(nil)

// upsertFTS 更新 FTS 索引
func (r *noteRepository) upsertFTS(db *gorm.DB, noteID int64, path, content string) {
	// 先删除旧记录
	_ = db.Exec("DELETE FROM note_fts WHERE note_id = ?", noteID).Error
	// 插入新记录
	_ = db.Exec("INSERT INTO note_fts (note_id, path, content) VALUES (?, ?, ?)", noteID, path, content).Error
}

// deleteFTS 删除 FTS 索引
func (r *noteRepository) deleteFTS(db *gorm.DB, noteID int64) {
	_ = db.Exec("DELETE FROM note_fts WHERE note_id = ?", noteID).Error
}

// searchFTS 使用 FTS5 MATCH 搜索内容，返回匹配的 note_id 列表
func (r *noteRepository) searchFTS(db *gorm.DB, keyword string, vaultID int64, isRecycle bool, sortBy, sortOrder string, limit, offset int) ([]int64, error) {
	var noteIDs []int64

	// 构建 action 条件
	actionCond := "n.action != 'delete'"
	if isRecycle {
		actionCond = "n.action = 'delete' AND n.rename = 0"
	}

	// 构建 FTS5 搜索词
	searchTerm := buildFTSSearchTerm(keyword)

	// 构建排序语句
	orderClause := "n." + buildOrderClause(sortBy, sortOrder)

	// 使用 FTS5 MATCH 查询
	sql := `
		SELECT f.note_id
		FROM note_fts f
		INNER JOIN note n ON f.note_id = n.id
		WHERE note_fts MATCH ?
		AND n.vault_id = ?
		AND ` + actionCond + `
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`

	err := db.Raw(sql, searchTerm, vaultID, limit, offset).Scan(&noteIDs).Error

	// 如果 FTS 查询失败或无结果，降级到 LIKE 查询
	if err != nil || len(noteIDs) == 0 {
		return r.searchFTSFallback(db, keyword, vaultID, isRecycle, sortBy, sortOrder, limit, offset)
	}

	return noteIDs, nil
}

// searchFTSFallback LIKE 降级查询
func (r *noteRepository) searchFTSFallback(db *gorm.DB, keyword string, vaultID int64, isRecycle bool, sortBy, sortOrder string, limit, offset int) ([]int64, error) {
	var noteIDs []int64

	actionCond := "n.action != 'delete'"
	if isRecycle {
		actionCond = "n.action = 'delete' AND n.rename = 0"
	}

	// 构建排序语句
	orderClause := "n." + buildOrderClause(sortBy, sortOrder)

	sql := `
		SELECT f.note_id
		FROM note_fts f
		INNER JOIN note n ON f.note_id = n.id
		WHERE (f.content LIKE ? OR f.path LIKE ?)
		AND n.vault_id = ?
		AND ` + actionCond + `
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`

	likeKeyword := "%" + keyword + "%"
	err := db.Raw(sql, likeKeyword, likeKeyword, vaultID, limit, offset).Scan(&noteIDs).Error
	return noteIDs, err
}

// searchFTSCount 使用 FTS5 MATCH 搜索计数
func (r *noteRepository) searchFTSCount(db *gorm.DB, keyword string, vaultID int64, isRecycle bool) (int64, error) {
	var count int64

	actionCond := "n.action != 'delete'"
	if isRecycle {
		actionCond = "n.action = 'delete' AND n.rename = 0"
	}

	searchTerm := buildFTSSearchTerm(keyword)

	sql := `
		SELECT COUNT(*)
		FROM note_fts f
		INNER JOIN note n ON f.note_id = n.id
		WHERE note_fts MATCH ?
		AND n.vault_id = ?
		AND ` + actionCond

	err := db.Raw(sql, searchTerm, vaultID).Scan(&count).Error

	// 如果 FTS 查询失败，降级到 LIKE 查询
	if err != nil || count == 0 {
		return r.searchFTSCountFallback(db, keyword, vaultID, isRecycle)
	}

	return count, nil
}

// searchFTSCountFallback LIKE 降级计数
func (r *noteRepository) searchFTSCountFallback(db *gorm.DB, keyword string, vaultID int64, isRecycle bool) (int64, error) {
	var count int64

	actionCond := "n.action != 'delete'"
	if isRecycle {
		actionCond = "n.action = 'delete' AND n.rename = 0"
	}

	sql := `
		SELECT COUNT(*)
		FROM note_fts f
		INNER JOIN note n ON f.note_id = n.id
		WHERE (f.content LIKE ? OR f.path LIKE ?)
		AND n.vault_id = ?
		AND ` + actionCond

	likeKeyword := "%" + keyword + "%"
	err := db.Raw(sql, likeKeyword, likeKeyword, vaultID).Scan(&count).Error
	return count, err
}

// buildFTSSearchTerm 构建 FTS5 搜索词
// 对于中文，将每个字符用空格分隔，使用 AND 连接
// 对于英文，直接使用原词加通配符
func buildFTSSearchTerm(keyword string) string {
	if keyword == "" {
		return ""
	}

	// 转义 FTS5 特殊字符
	keyword = escapeFTSSpecialChars(keyword)

	// 检查是否包含中文字符
	hasChinese := false
	for _, r := range keyword {
		if r >= 0x4E00 && r <= 0x9FFF {
			hasChinese = true
			break
		}
	}

	if hasChinese {
		// 中文：将每个字符作为独立 token，用 AND 连接
		var terms []string
		for _, r := range keyword {
			if r != ' ' {
				terms = append(terms, string(r))
			}
		}
		if len(terms) == 0 {
			return ""
		}
		// 使用 content: 前缀限定搜索内容字段
		result := "content:" + terms[0]
		for i := 1; i < len(terms); i++ {
			result += " AND content:" + terms[i]
		}
		return result
	}

	// 英文：使用前缀匹配
	return "content:" + keyword + "*"
}

// escapeFTSSpecialChars 转义 FTS5 特殊字符
func escapeFTSSpecialChars(s string) string {
	// FTS5 特殊字符：" * - + ( ) : ^ 需要转义或避免
	result := ""
	for _, r := range s {
		switch r {
		case '"', '*', '-', '+', '(', ')', ':', '^':
			// 跳过特殊字符
			continue
		default:
			result += string(r)
		}
	}
	return result
}
