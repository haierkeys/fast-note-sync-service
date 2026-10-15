package websocket_router

import (
	"context"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

func reconcileRetiredUpload(ctx context.Context, svc service.FileService, uid int64, session *FileUploadBinaryChunkSession, rejected *service.RetiredBytesError, send func(*code.Code, ...string), remove func()) error {
	err := service.WithRetirementState(ctx, svc, uid, session.Vault, session.Path, rejected.Current, func(current *dto.FileDTO) {
		finishRetiredUpload(session, current, send, remove)
	})
	if err != nil {
		finishRetiredUpload(session, nil, send, remove)
	}
	return err
}

// Claim completion exactly once while preventing a late chunk from creating a
// new file handle and changing the bytes between SHA verification and commit.
func claimFileUploadCompletion(session *FileUploadBinaryChunkSession, index uint32, size int64) (uploadedBytes, uploadedChunks, totalChunks int64, complete bool) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if _, duplicate := session.uploadedChunks[index]; !duplicate && !session.isFinalizing && !session.isCompleted {
		session.uploadedChunks[index] = struct{}{}
		session.UploadedChunks++
		session.UploadedBytes += size
		complete = session.UploadedBytes >= session.Size || session.UploadedChunks >= session.TotalChunks
		if complete {
			session.isFinalizing = true
		}
	}
	return session.UploadedBytes, session.UploadedChunks, session.TotalChunks, complete
}

// This is terminal success for transport, followed by authoritative content
// reconciliation. It never broadcasts the rejected upload to other clients.
func finishRetiredUpload(session *FileUploadBinaryChunkSession, current *dto.FileDTO, send func(*code.Code, ...string), remove func()) {
	var timestamp int64
	if current != nil {
		timestamp = current.UpdatedTimestamp
	}
	send(code.Success.WithData(dto.FileUploadAckMessage{LastTime: timestamp, Path: session.Path, PathHash: session.PathHash}).WithVault(session.Vault).WithContext(session.Context), string(FileUploadAck))
	session.mu.Lock()
	session.isCompleted = true
	session.mu.Unlock()
	remove()
	session.Cleanup()
	if current == nil {
		return
	}
	if current.Action == string(domain.FileActionDelete) {
		send(code.Success.WithData(dto.FileSyncDeleteMessage{
			Path: current.Path, PathHash: current.PathHash, Ctime: current.Ctime,
			Mtime: current.Mtime, Size: current.Size, UpdatedTimestamp: current.UpdatedTimestamp,
		}).WithVault(session.Vault), string(FileSyncDelete))
		return
	}
	// An explicit Restore or different new bytes may already be active.
	// Reconcile only the sender to that current state, preserving its content.
	send(code.Success.WithData(dto.FileSyncModifyMessage{
		Path: current.Path, PathHash: current.PathHash, ContentHash: current.ContentHash,
		Size: current.Size, Ctime: current.Ctime, Mtime: current.Mtime, UpdatedTimestamp: current.UpdatedTimestamp,
	}).WithVault(session.Vault), string(FileSyncUpdate))
}
