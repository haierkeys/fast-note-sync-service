package dao

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"gorm.io/gorm"
)

type noteTagRepository struct {
	dao *Dao
}

// NewNoteTagRepository creates a NoteTagRepository backed by the per-user database.
// NewNoteTagRepository 创建基于用户库的 NoteTagRepository。
func NewNoteTagRepository(dao *Dao) domain.NoteTagRepository {
	return &noteTagRepository{dao: dao}
}

func (r *noteTagRepository) db(uid int64) (*gorm.DB, error) {
	d := r.dao.ResolveDB("user_" + strconv.FormatInt(uid, 10))
	if d == nil {
		return nil, gorm.ErrInvalidDB
	}
	if err := model.AutoMigrate(d, "NoteTag"); err != nil {
		return nil, err
	}
	return d, nil
}

func (r *noteTagRepository) ReplaceForNote(ctx context.Context, noteID int64, tags []string, vaultID, uid int64, source string) error {
	d, err := r.db(uid)
	if err != nil {
		return err
	}
	d = d.WithContext(ctx)
	tags = normalizeTags(tags)
	return d.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("note_id = ? AND uid = ?", noteID, uid).
			Delete(&model.NoteTag{}).Error; err != nil {
			return err
		}
		if len(tags) == 0 {
			return nil
		}
		rows := make([]*model.NoteTag, 0, len(tags))
		for _, tag := range tags {
			rows = append(rows, &model.NoteTag{
				NoteID:  noteID,
				Tag:     tag,
				VaultID: vaultID,
				UID:     uid,
				Source:  source,
			})
		}
		return tx.CreateInBatches(rows, 200).Error
	})
}

func (r *noteTagRepository) DeleteByNoteID(ctx context.Context, noteID, uid int64) error {
	d, err := r.db(uid)
	if err != nil {
		return err
	}
	return d.WithContext(ctx).Where("note_id = ? AND uid = ?", noteID, uid).
		Delete(&model.NoteTag{}).Error
}

func (r *noteTagRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	d, err := r.db(uid)
	if err != nil {
		return err
	}
	return d.WithContext(ctx).Where("vault_id = ? AND uid = ?", vaultID, uid).
		Delete(&model.NoteTag{}).Error
}

func (r *noteTagRepository) ListByVault(ctx context.Context, vaultID, uid int64, prefix, scope string, limit int) ([]*domain.TagCount, error) {
	d, err := r.db(uid)
	if err != nil {
		return nil, err
	}
	type row struct {
		Tag   string
		Count int
	}
	var rows []row
	q := d.WithContext(ctx).Model(&model.NoteTag{}).
		Select("tag, COUNT(DISTINCT note_id) AS count").
		Where("vault_id = ? AND uid = ?", vaultID, uid)
	if p := strings.TrimSpace(prefix); p != "" {
		q = q.Where("LOWER(tag) LIKE ?", strings.ToLower(p)+"%")
	}
	// scope: 'frontmatter' keeps frontmatter-only tags, 'inline' keeps inline-only
	// ones, 'all'/empty keeps every source. A tag seen in both places is counted
	// under both scopes, so the two lists are not a partition of 'all'.
	// scope：frontmatter 只看仅在 frontmatter 出现的标签，inline 只看仅内联出现的，
	// all/空 看全部来源。两处都出现的标签会同时计入两个 scope，因此两者不是 all 的划分。
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "frontmatter":
		q = q.Where("source = ?", TagSourceFrontmatter)
	case "inline":
		q = q.Where("source = ?", TagSourceInline)
	}
	q = q.Group("tag").Order("count DESC, tag ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.TagCount, 0, len(rows))
	for _, r := range rows {
		out = append(out, &domain.TagCount{Tag: r.Tag, Count: r.Count})
	}
	return out, nil
}

func (r *noteTagRepository) GetTagsByNoteID(ctx context.Context, noteID, uid int64) ([]string, error) {
	d, err := r.db(uid)
	if err != nil {
		return nil, err
	}
	var tags []string
	if err := d.WithContext(ctx).Model(&model.NoteTag{}).
		Where("note_id = ? AND uid = ?", noteID, uid).
		Order("tag ASC").Pluck("tag", &tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *noteTagRepository) RebuildForVault(ctx context.Context, vaultID, uid int64) (int, error) {
	d, err := r.db(uid)
	if err != nil {
		return 0, err
	}
	var notes []*model.Note
	if err := d.WithContext(ctx).Select("id", "path", "content", "vault_id", "uid").
		Where("vault_id = ? AND uid = ? AND action <> ?", vaultID, uid, "delete").
		Find(&notes).Error; err != nil {
		return 0, err
	}
	if err := r.DeleteByVaultID(ctx, vaultID, uid); err != nil {
		return 0, err
	}
	written := 0
	for _, n := range notes {
		tags, source := ExtractTags(n.Content)
		if len(tags) == 0 {
			continue
		}
		if err := r.ReplaceForNote(ctx, n.ID, tags, vaultID, uid, source); err != nil {
			return written, err
		}
		written += len(tags)
	}
	return written, nil
}

// normalizeTags reuses the shared tag grammar: trim '#', drop blanks, de-dupe
// case-insensitively (original casing preserved), then sort for stable output.
// normalizeTags 复用统一的标签规则：去掉 '#'、剔除空值、忽略大小写去重（保留原始大小写），再排序以稳定输出。
func normalizeTags(tags []string) []string {
	out := dedupeTags(tags)
	sort.SliceStable(out, func(i, j int) bool { return strings.ToLower(out[i]) < strings.ToLower(out[j]) })
	return out
}
