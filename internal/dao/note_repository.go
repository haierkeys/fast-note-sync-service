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
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// noteRepository 实现 domain.NoteRepository 接口
type noteRepository struct {
	dao *Dao
}

// NewNoteRepository 创建 NoteRepository 实例
func NewNoteRepository(dao *Dao) domain.NoteRepository {
	return &noteRepository{dao: dao}
}

// note 获取笔记查询对象
func (r *noteRepository) note(uid int64) *query.Query {
	key := "user_" + strconv.FormatInt(uid, 10)
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "Note")
	}, key+"#note", key)
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
		_ = r.dao.SaveContentToFile(folder, "content.txt", n.Content)
	}

	// 加载快照
	if snapshot, exists, _ := r.dao.LoadContentFromFile(folder, "snapshot.txt"); exists {
		n.ContentLastSnapshot = snapshot
	} else if n.ContentLastSnapshot != "" {
		_ = r.dao.SaveContentToFile(folder, "snapshot.txt", n.ContentLastSnapshot)
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
	u := r.note(uid).Note
	m := r.toModel(note)

	m.UpdatedTimestamp = timex.Now().UnixMilli()
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()

	content := m.Content
	m.Content = ""             // 不在数据库存储内容
	m.ContentLastSnapshot = "" // 不在数据库存储快照

	err := u.WithContext(ctx).Create(m)
	if err != nil {
		return nil, err
	}

	// 保存内容到文件
	folder := r.dao.GetNoteFolderPath(uid, m.ID)
	_ = r.dao.SaveContentToFile(folder, "content.txt", content)

	result := r.toDomain(m, uid)
	result.Content = content
	return result, nil
}

// Update 更新笔记
func (r *noteRepository) Update(ctx context.Context, note *domain.Note, uid int64) (*domain.Note, error) {
	u := r.note(uid).Note
	m := r.toModel(note)

	m.UpdatedTimestamp = timex.Now().UnixMilli()
	m.UpdatedAt = timex.Now()

	content := m.Content
	m.Content = "" // 不在数据库更新内容

	err := u.WithContext(ctx).Where(
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
		u.UpdatedAt,
		u.UpdatedTimestamp,
	).Save(m)

	if err != nil {
		return nil, err
	}

	// 保存内容到文件
	folder := r.dao.GetNoteFolderPath(uid, m.ID)
	_ = r.dao.SaveContentToFile(folder, "content.txt", content)

	result := r.toDomain(m, uid)
	result.Content = content
	return result, nil
}

// UpdateDelete 更新笔记为删除状态
func (r *noteRepository) UpdateDelete(ctx context.Context, note *domain.Note, uid int64) error {
	u := r.note(uid).Note
	m := &model.Note{
		ID:               note.ID,
		Action:           string(note.Action),
		Rename:           note.Rename,
		ClientName:       note.ClientName,
		UpdatedTimestamp: timex.Now().UnixMilli(),
	}

	return u.WithContext(ctx).Where(
		u.ID.Eq(m.ID),
	).Select(
		u.ID,
		u.Action,
		u.Rename,
		u.ClientName,
		u.UpdatedTimestamp,
	).Save(m)
}

// UpdateMtime 更新笔记修改时间
func (r *noteRepository) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	u := r.note(uid).Note

	_, err := u.WithContext(ctx).Where(
		u.ID.Eq(id),
	).UpdateSimple(
		u.Mtime.Value(mtime),
		u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

// UpdateSnapshot 更新笔记快照
func (r *noteRepository) UpdateSnapshot(ctx context.Context, snapshot, snapshotHash string, version, id, uid int64) error {
	u := r.note(uid).Note

	// 保存快照到文件
	folder := r.dao.GetNoteFolderPath(uid, id)
	_ = r.dao.SaveContentToFile(folder, "snapshot.txt", snapshot)

	_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).UpdateSimple(
		u.ContentLastSnapshot.Value(""),
		u.ContentLastSnapshotHash.Value(snapshotHash),
		u.Version.Value(version),
	)
	return err
}

// Delete 物理删除笔记
func (r *noteRepository) Delete(ctx context.Context, id, uid int64) error {
	u := r.note(uid).Note
	_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}

	// 删除物理文件
	folder := r.dao.GetNoteFolderPath(uid, id)
	_ = r.dao.RemoveContentFolder(folder)

	return nil
}

// DeletePhysicalByTime 根据时间物理删除已标记删除的笔记
func (r *noteRepository) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
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
}

// List 分页获取笔记列表
func (r *noteRepository) List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool) ([]*domain.Note, error) {
	u := r.note(uid).Note
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
	)

	if isRecycle {
		q = q.Where(u.Action.Eq("delete"), u.Rename.Eq(0))
	} else {
		q = q.Where(u.Action.Neq("delete"))
	}

	var modelList []*model.Note
	var err error

	if keyword != "" {
		key := "%" + keyword + "%"

		if isRecycle {
			err = q.UnderlyingDB().Where("path LIKE ? ", key).
				Order("updated_timestamp DESC").
				Limit(pageSize).
				Offset(app.GetPageOffset(page, pageSize)).
				Find(&modelList).Error
		} else {
			err = q.UnderlyingDB().Where("path LIKE ? ", key).
				Order("path DESC, created_at DESC").
				Limit(pageSize).
				Offset(app.GetPageOffset(page, pageSize)).
				Find(&modelList).Error
		}
	} else {
		if isRecycle {
			modelList, err = q.Order(
				u.UpdatedTimestamp.Desc(),
			).Limit(pageSize).
				Offset(app.GetPageOffset(page, pageSize)).
				Find()
		} else {
			modelList, err = q.Order(
				u.Path.Desc(),
				u.CreatedAt.Desc(),
			).Limit(pageSize).
				Offset(app.GetPageOffset(page, pageSize)).
				Find()
		}
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

// ListCount 获取笔记数量
func (r *noteRepository) ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool) (int64, error) {
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
		key := "%" + keyword + "%"
		err = q.UnderlyingDB().Where("path LIKE ? OR content LIKE ?", key, key).Count(&count).Error
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

// 确保 noteRepository 实现了 domain.NoteRepository 接口
var _ domain.NoteRepository = (*noteRepository)(nil)
