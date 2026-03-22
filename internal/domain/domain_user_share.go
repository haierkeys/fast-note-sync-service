package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrShareCancelled = errors.New("share has been cancelled")
	ErrShareExpired   = errors.New("share has expired")
)

// UserShare 笔记分享领域模型
type UserShare struct {
	ID           int64               `json:"id"`
	UID          int64               `json:"uid"`            // 创建者 ID
	ResType      string              `json:"res_type"`       // 资源类型: note, file
	ResID        int64               `json:"res_id"`         // 资源 ID (note.id 或 file.id)
	Resources    map[string][]string `json:"res"`            // 资源授权列表 (JSON: {"note":["id1"],"file":["id2"]})
	Status       int64               `json:"status"`         // 状态: 1-有效, 2-已撤销
	ViewCount    int64               `json:"view_count"`     // 统计：访问次数
	LastViewedAt time.Time           `json:"last_viewed_at"` // 统计：最后访问时间
	ExpiresAt    time.Time           `json:"expires_at"`     // 过期时间
	ShortLink    string              `json:"short_link"`     // 短链接
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// UserShareRepository 用户分享持久化接口
type UserShareRepository interface {
	Create(ctx context.Context, uid int64, share *UserShare) error
	GetByID(ctx context.Context, uid int64, id int64) (*UserShare, error)
	GetByPath(ctx context.Context, uid int64, vaultID int64, pathHash string) (*UserShare, error)
	GetByRes(ctx context.Context, uid int64, resType string, resID int64) (*UserShare, error)
	UpdateStatus(ctx context.Context, uid int64, id int64, status int64) error
	UpdateViewStats(ctx context.Context, uid int64, id int64, viewCountIncr int64, lastViewedAt time.Time) error
	UpdateShortLink(ctx context.Context, uid int64, id int64, shortLink string) error
	ListByUID(ctx context.Context, uid int64, sortBy string, sortOrder string, offset, limit int) ([]*UserShare, error)
	CountByUID(ctx context.Context, uid int64) (int64, error)
}
