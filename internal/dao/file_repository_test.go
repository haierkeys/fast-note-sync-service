package dao

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupFileRepoTestEnv sets up a temporary workspace for a *Dao backed by SQLite, with the
// write queue disabled so ExecuteWrite runs synchronously without needing a writequeue.Manager.
// setupFileRepoTestEnv 为 *Dao（SQLite 后端）搭建临时工作区，关闭写队列以便 ExecuteWrite
// 同步执行，无需注入 writequeue.Manager。
func setupFileRepoTestEnv(t *testing.T) (domain.FileRepository, func()) {
	tempDir, err := os.MkdirTemp("", "fast-note-sync-service-filerepo-test-*")
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))

	require.NoError(t, os.MkdirAll(filepath.Join("storage", "database"), 0755))

	logger := zap.NewNop()

	dbPath := filepath.Join("storage", "database", "db.sqlite3")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	dbCfg := &config.DatabaseConfig{
		Type:             "sqlite",
		Path:             dbPath,
		EnableWriteQueue: util.Ptr(false),
	}
	daoInst := New(db, context.Background(),
		WithConfig(dbCfg),
		WithUserDatabaseConfig(dbCfg),
		WithLogger(logger),
	)

	fileRepo := NewFileRepository(daoInst)

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		_ = os.Chdir(origWd)
		_ = os.RemoveAll(tempDir)
	}

	return fileRepo, cleanup
}

// TestFileRepository_Create_RenameFailure_PropagatesErrorAndCleansUp verifies the P0 fix:
// when the temp-file rename fails, Create must return the error (not silently succeed) and
// must not leave an orphaned DB row that claims the file exists.
// TestFileRepository_Create_RenameFailure_PropagatesErrorAndCleansUp 验证 P0 修复：
// 临时文件 rename 失败时，Create 必须返回错误（不能静默成功），且不能残留一条声称文件
// 存在的孤儿数据库记录。
func TestFileRepository_Create_RenameFailure_PropagatesErrorAndCleansUp(t *testing.T) {
	fileRepo, cleanup := setupFileRepoTestEnv(t)
	defer cleanup()

	ctx := context.Background()
	const uid = int64(1)

	f := &domain.File{
		VaultID:     1,
		Path:        "a.md",
		PathHash:    "hash-a",
		ContentHash: "content-hash-a",
		// Non-existent temp path makes os.Rename fail deterministically.
		// 不存在的临时路径，使 os.Rename 必然失败。
		SavePath: filepath.Join(os.TempDir(), "does-not-exist-fast-note-sync-test", "tmp.dat"),
		Size:     10,
		Mtime:    1000,
		Ctime:    1000,
		Action:   domain.FileActionCreate,
	}

	created, err := fileRepo.Create(ctx, f, uid)
	assert.Error(t, err, "Create must return an error when the rename fails")
	assert.Nil(t, created)

	// The row must not remain in the DB claiming the file exists.
	// 数据库中不能残留一条声称文件存在的记录。
	got, getErr := fileRepo.GetByPathHash(ctx, "hash-a", 1, uid)
	assert.Error(t, getErr, "expected record-not-found, orphaned row should have been cleaned up")
	assert.Nil(t, got)
}

// TestFileRepository_Update_RenameFailure_AbortsBeforeDBWrite verifies the P0 fix: when the
// temp-file rename fails during Update, the DB row must be left untouched (pointing at the
// previous, still-valid content) instead of being updated to claim new content landed.
// TestFileRepository_Update_RenameFailure_AbortsBeforeDBWrite 验证 P0 修复：Update 过程中
// rename 失败时，数据库行必须保持不变（仍指向此前有效的内容），而不是被更新为声称新内容已落地。
func TestFileRepository_Update_RenameFailure_AbortsBeforeDBWrite(t *testing.T) {
	fileRepo, cleanup := setupFileRepoTestEnv(t)
	defer cleanup()

	ctx := context.Background()
	const uid = int64(1)

	// First create a valid file (no SavePath, so no rename happens; a pure metadata row).
	// 先创建一条不带 SavePath 的合法记录（不触发 rename，作为基线元数据行）。
	f := &domain.File{
		VaultID:     1,
		Path:        "b.md",
		PathHash:    "hash-b",
		ContentHash: "content-hash-b-v1",
		Size:        10,
		Mtime:       1000,
		Ctime:       1000,
		Action:      domain.FileActionCreate,
	}
	created, err := fileRepo.Create(ctx, f, uid)
	require.NoError(t, err)
	require.NotNil(t, created)

	// Now attempt an update that provides a new (non-existent) temp file, forcing rename to fail.
	// 现在尝试更新并提供一个不存在的新临时文件，强制 rename 失败。
	created.ContentHash = "content-hash-b-v2"
	created.SavePath = filepath.Join(os.TempDir(), "does-not-exist-fast-note-sync-test", "tmp2.dat")

	updated, err := fileRepo.Update(ctx, created, uid)
	assert.Error(t, err, "Update must return an error when the rename fails")
	assert.Nil(t, updated)

	// The DB row must still reflect the old content hash, not the failed update.
	// 数据库行必须仍反映旧的 content hash，而不是失败的更新。
	got, getErr := fileRepo.GetByPathHash(ctx, "hash-b", 1, uid)
	require.NoError(t, getErr)
	require.NotNil(t, got)
	assert.Equal(t, "content-hash-b-v1", got.ContentHash, "DB row must not have been overwritten when rename failed")
}
