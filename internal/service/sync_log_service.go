// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"go.uber.org/zap"
)

// SyncLogService defines the sync log business service interface
// SyncLogService 定义同步日志业务服务接口
type SyncLogService interface {
	// Log asynchronously records a sync log entry, does not block the caller
	// Log 异步记录一条同步日志，不阻塞调用方
	Log(
		uid int64,
		vaultID int64,
		logType domain.SyncLogType,
		action domain.SyncLogAction,
		changedFields string, // e.g. "content,mtime" / "mtime" / "path" / "" // 如 "content,mtime" / "mtime" / "path" / ""
		path string,
		pathHash string,
		clientType string,
		clientName string,
		clientVersion string,
		size int64,
	)

	// List retrieves sync logs with pagination
	// List 分页查询同步日志
	List(ctx context.Context, uid int64, vaultID int64, logType, action string, page, pageSize int) ([]*dto.SyncLogDTO, int64, error)

	// CleanupByTime removes sync logs older than the given cutoff time for all users
	// CleanupByTime 清理所有用户在指定截止时间之前的同步日志
	CleanupByTime(ctx context.Context, cutoffTime int64) error
}

// syncLogService implements SyncLogService
// syncLogService 实现 SyncLogService 接口
type syncLogService struct {
	repo domain.SyncLogRepository // Sync log repository // 同步日志仓储
}

// NewSyncLogService creates a new SyncLogService instance
// NewSyncLogService 创建 SyncLogService 实例
func NewSyncLogService(repo domain.SyncLogRepository) SyncLogService {
	return &syncLogService{repo: repo}
}

// Log asynchronously records a sync log entry
// Log 异步记录一条同步日志
func (s *syncLogService) Log(
	uid int64,
	vaultID int64,
	logType domain.SyncLogType,
	action domain.SyncLogAction,
	changedFields string,
	path string,
	pathHash string,
	clientType string,
	clientName string,
	clientVersion string,
	size int64,
) {
	go func() {
		// Use Background context for asynchronous logging
		// 使用 Background context 进行异步日志记录
		ctx := context.Background()
		entry := &domain.SyncLog{
			UID:           uid,
			VaultID:       vaultID,
			Type:          logType,
			Action:        action,
			ChangedFields: changedFields,
			Path:          path,
			PathHash:      pathHash,
			Size:           size,
			ClientType:    clientType,
			ClientName:    clientName,
			ClientVersion: clientVersion,
			Status:        1, // success // 成功
			CreatedAt:     time.Now(),
		}
		if err := s.repo.Create(ctx, entry, uid); err != nil {
			zap.L().Warn("SyncLogService.Log: failed to create sync log",
				zap.Int64("uid", uid),
				zap.Int64("vaultID", vaultID),
				zap.String("type", string(logType)),
				zap.String("action", string(action)),
				zap.String("path", path),
				zap.Error(err),
			)
		}
	}()
}

// List retrieves sync logs with optional filters and pagination
// List 按条件分页查询同步日志
func (s *syncLogService) List(ctx context.Context, uid int64, vaultID int64, logType, action string, page, pageSize int) ([]*dto.SyncLogDTO, int64, error) {
	logs, total, err := s.repo.List(ctx, uid, logType, action, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]*dto.SyncLogDTO, 0, len(logs))
	for _, l := range logs {
		if vaultID > 0 && l.VaultID != vaultID {
			continue
		}
		result = append(result, s.domainToDTO(l))
	}
	return result, total, nil
}

// CleanupByTime removes sync logs older than the given cutoff time for all users
// CleanupByTime 清理所有用户在指定截止时间之前的同步日志
func (s *syncLogService) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	return s.repo.CleanupByTimeAll(ctx, cutoffTime)
}

// domainToDTO converts domain SyncLog to DTO
// domainToDTO 将领域模型转换为 DTO
func (s *syncLogService) domainToDTO(l *domain.SyncLog) *dto.SyncLogDTO {
	return &dto.SyncLogDTO{
		ID:            l.ID,
		VaultID:       l.VaultID,
		Type:          string(l.Type),
		Action:        string(l.Action),
		ChangedFields: l.ChangedFields,
		Path:          l.Path,
		PathHash:      l.PathHash,
		Size:          l.Size,
		ClientName:    l.ClientName,
		ClientType:    l.ClientType,
		ClientVersion: l.ClientVersion,
		Status:        l.Status,
		Message:       l.Message,
		CreatedAt:     l.CreatedAt,
	}
}

// Ensure syncLogService implements SyncLogService
// 确保 syncLogService 实现了 SyncLogService 接口
var _ SyncLogService = (*syncLogService)(nil)
