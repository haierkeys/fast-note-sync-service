package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/retirement"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"gorm.io/gorm"
)

// RetiredBytesError is a terminal, acknowledged upload rejection. Current may
// be a tombstone or a newer active file; the latter must never be deleted.
type RetiredBytesError struct{ Current *dto.FileDTO }

func (e *RetiredBytesError) Error() string { return "upload contains explicitly retired bytes" }

func fileWriteKey(uid, vault int64, path string) string {
	return fmt.Sprintf("file_write_%d_%d_%s", uid, vault, path)
}

func (s *fileService) fileForWrite(ctx context.Context, uid, vault int64, path, pathHash string) (*domain.File, error) {
	if s.retirement.Covers(uid, vault, path) {
		return s.fileRepo.GetByPath(ctx, path, vault, uid)
	}
	return s.fileRepo.GetByPathHash(ctx, pathHash, vault, uid)
}

func (s *fileService) guardFileWrite(uid, vault int64, params *dto.FileUpdateRequest, file *domain.File) error {
	if !s.retirement.Covers(uid, vault, params.Path) {
		return nil
	}
	bytesPath := params.SavePath
	if bytesPath == "" && file != nil {
		bytesPath = file.SavePath
	}
	blocked, _, err := s.retirement.CheckFile(uid, vault, params.Path, bytesPath)
	if err != nil {
		return err
	}
	if !blocked {
		if params.PathHash != util.EncodeHash32(params.Path) {
			return errors.New("retirement path/hash mismatch")
		}
		return nil
	}
	current := s.domainToDTO(file)
	if current == nil {
		current = &dto.FileDTO{Action: string(domain.FileActionDelete), Path: params.Path, PathHash: util.EncodeHash32(params.Path)}
	}
	return &RetiredBytesError{Current: current}
}

// WithRetirementState serializes response enqueueing with Restore/Delete/write
// and re-reads exact path state. This prevents a rejected upload's old snapshot
// from deleting content that Restore or a new upload committed in the meantime.
func WithRetirementState(ctx context.Context, svc FileService, uid int64, vaultName, path string, fallback *dto.FileDTO, deliver func(*dto.FileDTO)) error {
	s, ok := svc.(*fileService)
	if !ok {
		deliver(fallback)
		return nil
	}
	vault, err := s.vaultService.MustGetID(ctx, uid, vaultName)
	if err != nil {
		return err
	}
	unlock := s.kmu.Lock(fileWriteKey(uid, vault, path))
	defer unlock()
	file, err := s.fileRepo.GetByPath(ctx, path, vault, uid)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	current := s.domainToDTO(file)
	if current == nil {
		current = &dto.FileDTO{Action: string(domain.FileActionDelete), Path: path, PathHash: util.EncodeHash32(path)}
	}
	deliver(current)
	return nil
}

func (s *fileService) auditRestore(uid, vault int64, file *domain.File, event string) error {
	if !s.retirement.Covers(uid, vault, file.Path) {
		return nil
	}
	digest, err := retirement.SHA256File(file.SavePath)
	if err != nil {
		return err
	}
	return s.retirement.Audit(uid, vault, file.ID, file.Path, digest, event, s.clientType, s.clientName, s.clientVer)
}
