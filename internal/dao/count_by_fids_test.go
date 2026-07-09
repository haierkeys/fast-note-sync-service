package dao

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupCountByFIDsTestEnv is a minimal variant of setupFileRepoTestEnv that also exposes the
// raw *Dao, so the test can raw-insert rows and exercise CountByFIDs (a GORM Group()+Select()+
// Scan() query) against a real SQLite database instead of only compiling against mocks.
// setupCountByFIDsTestEnv 是 setupFileRepoTestEnv 的精简变体，额外暴露 *Dao 本身，
// 以便测试直接写入原始行，针对真实 SQLite 数据库验证 CountByFIDs（一个 GORM
// Group()+Select()+Scan() 查询），而不是只在 mock 层面编译通过。
func setupCountByFIDsTestEnv(t *testing.T) (*Dao, func()) {
	tempDir, err := os.MkdirTemp("", "fast-note-sync-service-countfids-test-*")
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))

	require.NoError(t, os.MkdirAll(filepath.Join("storage", "database"), 0755))

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
		WithLogger(zap.NewNop()),
	)

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		_ = os.Chdir(origWd)
		_ = os.RemoveAll(tempDir)
	}

	return daoInst, cleanup
}

// TestNoteRepository_CountByFIDs_GroupsCorrectly verifies the GetTree N+1 fix: CountByFIDs
// must return per-folder note counts in one grouped query, excluding soft-deleted notes and
// notes outside the requested fid set / vault.
// TestNoteRepository_CountByFIDs_GroupsCorrectly 验证 GetTree N+1 修复：CountByFIDs
// 必须用一次分组查询返回按文件夹 ID 分组的笔记计数，且排除软删除笔记、非目标 fid/vault 的笔记。
func TestNoteRepository_CountByFIDs_GroupsCorrectly(t *testing.T) {
	daoInst, cleanup := setupCountByFIDsTestEnv(t)
	defer cleanup()

	ctx := context.Background()
	const uid = int64(1)
	const vaultID = int64(1)

	noteRepo := NewNoteRepository(daoInst).(*noteRepository)

	// Trigger schema auto-migration via the repository's own query accessor.
	// 通过仓库自身的查询访问器触发建表。
	_, err := noteRepo.CountByFIDs(ctx, []int64{1}, vaultID, uid)
	require.NoError(t, err)

	db := daoInst.ResolveDB(noteRepo.GetKey(uid))
	rows := []*model.Note{
		{VaultID: vaultID, FID: 10, Action: "modify", Path: "a.md", PathHash: "ha"},
		{VaultID: vaultID, FID: 10, Action: "create", Path: "b.md", PathHash: "hb"},
		{VaultID: vaultID, FID: 20, Action: "modify", Path: "c.md", PathHash: "hc"},
		// Soft-deleted note in folder 10 must not be counted.
		// 文件夹 10 下的软删除笔记不应被计数。
		{VaultID: vaultID, FID: 10, Action: "delete", Path: "d.md", PathHash: "hd"},
		// Different vault must not leak into the count.
		// 不同 vault 的记录不应混入计数。
		{VaultID: 999, FID: 10, Action: "modify", Path: "e.md", PathHash: "he"},
	}
	for _, r := range rows {
		require.NoError(t, db.Create(r).Error)
	}

	counts, err := noteRepo.CountByFIDs(ctx, []int64{10, 20, 30}, vaultID, uid)
	require.NoError(t, err)

	require.Equal(t, int64(2), counts[10], "folder 10 should count 2 non-deleted notes")
	require.Equal(t, int64(1), counts[20], "folder 20 should count 1 note")
	_, has30 := counts[30]
	require.False(t, has30, "folder 30 has no notes and should be absent from the result map")
}

// TestFileRepository_CountByFIDs_GroupsCorrectly mirrors the note test for files.
// TestFileRepository_CountByFIDs_GroupsCorrectly 是文件版本的对应测试。
func TestFileRepository_CountByFIDs_GroupsCorrectly(t *testing.T) {
	daoInst, cleanup := setupCountByFIDsTestEnv(t)
	defer cleanup()

	ctx := context.Background()
	const uid = int64(1)
	const vaultID = int64(1)

	fileRepo := NewFileRepository(daoInst).(*fileRepository)

	_, err := fileRepo.CountByFIDs(ctx, []int64{1}, vaultID, uid)
	require.NoError(t, err)

	db := daoInst.ResolveDB(fileRepo.GetKey(uid))
	rows := []*model.File{
		{VaultID: vaultID, FID: 10, Action: "modify", Path: "a.dat", PathHash: "ha"},
		{VaultID: vaultID, FID: 20, Action: "modify", Path: "b.dat", PathHash: "hb"},
		{VaultID: vaultID, FID: 20, Action: "modify", Path: "c.dat", PathHash: "hc"},
		{VaultID: vaultID, FID: 20, Action: "delete", Path: "d.dat", PathHash: "hd"},
	}
	for _, r := range rows {
		require.NoError(t, db.Create(r).Error)
	}

	counts, err := fileRepo.CountByFIDs(ctx, []int64{10, 20}, vaultID, uid)
	require.NoError(t, err)

	require.Equal(t, int64(1), counts[10])
	require.Equal(t, int64(2), counts[20])
}
