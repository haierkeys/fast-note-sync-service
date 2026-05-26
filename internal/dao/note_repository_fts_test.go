package dao

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupFTSTestEnv sets up a temporary workspace for GORM SQLite and Bleve
// setupFTSTestEnv 为 GORM SQLite 和 Bleve 设置临时工作区环境
func setupFTSTestEnv(t *testing.T, storeRaw bool) (*Dao, domain.NoteRepository, string, func()) {
	origWd, err := os.Getwd()
	require.NoError(t, err)

	tempDir, err := os.MkdirTemp("", "fast-note-sync-service-test-*")
	require.NoError(t, err)

	// Change to temporary directory to isolate file output
	// 切换到临时目录以隔离文件输出
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create storage directories
	// 创建存储目录
	err = os.MkdirAll(filepath.Join("storage", "database"), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join("storage", "vault_fts"), 0755)
	require.NoError(t, err)

	logger := zap.NewNop()

	bleveMgr := NewBleveManager(util.Ptr(true), util.Ptr(storeRaw), logger)

	// Initialize main GORM SQLite connection
	// 初始化主 GORM SQLite 连接
	dbPath := filepath.Join("storage", "database", "db.sqlite3")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	// Initialize GORM migration for main tables
	// 对主表初始化 GORM 迁移
	err = db.AutoMigrate(&model.Vault{})
	require.NoError(t, err)

	// Initialize Dao
	// 初始化 Dao
	dbCfg := &config.DatabaseConfig{
		Type: "sqlite",
		Path: dbPath,
	}
	daoInst := New(db, context.Background(),
		WithConfig(dbCfg),
		WithUserDatabaseConfig(dbCfg),
		WithLogger(logger),
		WithBleveManager(bleveMgr),
	)

	noteRepo := NewNoteRepository(daoInst)

	cleanup := func() {
		// Close all open Bleve indexes
		// 关闭所有打开的 Bleve 索引
		_ = bleveMgr.CloseAll()

		// Close GORM main database
		// 关闭 GORM 主数据库
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}

		// Close any user tenant database connections
		// 关闭所有用户租户数据库连接
		daoInst.mu.Lock()
		for _, entry := range daoInst.KeyDb {
			if sqlDB, err := entry.db.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
		daoInst.mu.Unlock()

		// Restore working directory and remove temp files
		// 恢复工作目录并删除临时文件
		_ = os.Chdir(origWd)
		_ = os.RemoveAll(tempDir)
	}

	return daoInst, noteRepo, tempDir, cleanup
}

// TestBleveFTSAllInOne tests indexing, search, sorting, and config change automatic rebuild
// TestBleveFTSAllInOne 测试索引、搜索、排序以及配置更改时的自动重建
func TestBleveFTSAllInOne(t *testing.T) {
	daoInst, noteRepo, _, cleanup := setupFTSTestEnv(t, true)
	defer cleanup()

	ctx := context.Background()
	uid := int64(99999)
	vaultID := int64(888)

	userDb := daoInst.ResolveDB("user_99999")
	_ = userDb.AutoMigrate(&model.Note{})
	_ = model.CreateNoteFTSTable(userDb)

	// Create test notes in DB
	// 在数据库中创建测试笔记
	notes := []model.Note{
		{
			ID:       1,
			VaultID:  vaultID,
			Path:     "A_intro.md",
			PathHash: util.EncodeHash32("A_intro.md"),
			Action:   "",
			Rename:   0,
			Ctime:    1000,
			Mtime:    5000,
		},
		{
			ID:       2,
			VaultID:  vaultID,
			Path:     "B_tutorial.md",
			PathHash: util.EncodeHash32("B_tutorial.md"),
			Action:   "",
			Rename:   0,
			Ctime:    2000,
			Mtime:    4000,
		},
		{
			ID:       3,
			VaultID:  vaultID,
			Path:     "C_advanced.md",
			PathHash: util.EncodeHash32("C_advanced.md"),
			Action:   "",
			Rename:   0,
			Ctime:    3000,
			Mtime:    3000,
		},
		{
			ID:       4,
			VaultID:  vaultID,
			Path:     "D_deleted.md",
			PathHash: util.EncodeHash32("D_deleted.md"),
			Action:   "delete",
			Rename:   0,
			Ctime:    4000,
			Mtime:    2000,
		},
	}

	for _, n := range notes {
		err := userDb.Create(&n).Error
		require.NoError(t, err)
	}

	// Write mock file contents to simulate LoadContentFromFile
	// 写入模拟文件内容以模拟 LoadContentFromFile
	noteContents := map[int64]string{
		1: "This is a quick intro guide to the fast note sync service. Highly recommended.",
		2: "Learn how to use sync features. Step by step tutorial for beginners.",
		3: "Advanced configuration and search. Bleve FTS options guide.",
		4: "This note is deleted but still has content about sync.",
	}

	for id, content := range noteContents {
		folder := daoInst.GetNoteFolderPath(uid, id)
		err := os.MkdirAll(folder, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(folder, "content.txt"), []byte(content), 0644)
		require.NoError(t, err)
	}

	// Build index
	// 建立索引
	err := noteRepo.(*noteRepository).RebuildVaultIndex(ctx, uid, vaultID)
	require.NoError(t, err)

	// 1. Test basic search (keyword "sync")
	// 1. 测试基础搜索（关键词 "sync"）
	t.Run("Basic search keyword sync", func(t *testing.T) {
		ids, err := noteRepo.(*noteRepository).searchFTS(uid, vaultID, "sync", false, "mtime", "desc", 10, 0)
		require.NoError(t, err)
		// ID 4 is deleted, so only ID 1 and 2 should match
		// ID 4 已删除，所以只有 ID 1 和 2 应该匹配
		assert.Len(t, ids, 2)
		assert.Contains(t, ids, int64(1))
		assert.Contains(t, ids, int64(2))
	})

	// 2. Test sorting options
	// 2. 测试排序选项
	t.Run("Sorting by ctime asc", func(t *testing.T) {
		ids, err := noteRepo.(*noteRepository).searchFTS(uid, vaultID, "guide", false, "ctime", "asc", 10, 0)
		require.NoError(t, err)
		// "guide" matches ID 1 (ctime 1000, "intro guide") and ID 3 (ctime 3000, "options guide")
		// "guide" 匹配 ID 1 (ctime 1000) 和 ID 3 (ctime 3000)
		assert.Len(t, ids, 2)
		assert.Equal(t, int64(1), ids[0])
		assert.Equal(t, int64(3), ids[1])
	})

	t.Run("Sorting by ctime desc", func(t *testing.T) {
		ids, err := noteRepo.(*noteRepository).searchFTS(uid, vaultID, "guide", false, "ctime", "desc", 10, 0)
		require.NoError(t, err)
		assert.Len(t, ids, 2)
		assert.Equal(t, int64(3), ids[0])
		assert.Equal(t, int64(1), ids[1])
	})

	t.Run("Sorting by mtime asc", func(t *testing.T) {
		ids, err := noteRepo.(*noteRepository).searchFTS(uid, vaultID, "guide", false, "mtime", "asc", 10, 0)
		require.NoError(t, err)
		// ID 3 (mtime 3000) < ID 1 (mtime 5000)
		assert.Len(t, ids, 2)
		assert.Equal(t, int64(3), ids[0])
		assert.Equal(t, int64(1), ids[1])
	})

	t.Run("Sorting by path asc", func(t *testing.T) {
		ids, err := noteRepo.(*noteRepository).searchFTS(uid, vaultID, "guide", false, "path", "asc", 10, 0)
		require.NoError(t, err)
		// ID 1 ("A_intro.md") < ID 3 ("C_advanced.md")
		assert.Len(t, ids, 2)
		assert.Equal(t, int64(1), ids[0])
		assert.Equal(t, int64(3), ids[1])
	})

	t.Run("Sorting by path desc", func(t *testing.T) {
		ids, err := noteRepo.(*noteRepository).searchFTS(uid, vaultID, "guide", false, "path", "desc", 10, 0)
		require.NoError(t, err)
		assert.Len(t, ids, 2)
		assert.Equal(t, int64(3), ids[0])
		assert.Equal(t, int64(1), ids[1])
	})

	// 3. Test recycle bin filter
	// 3. 测试回收站过滤
	t.Run("Recycle bin filter isRecycle=true", func(t *testing.T) {
		ids, err := noteRepo.(*noteRepository).searchFTS(uid, vaultID, "sync", true, "mtime", "desc", 10, 0)
		require.NoError(t, err)
		// Only ID 4 (action="delete", content contains "sync") should match
		// 只有 ID 4（已删除，且内容包含 "sync"）应该匹配
		assert.Len(t, ids, 1)
		assert.Equal(t, int64(4), ids[0])
	})

	// 4. Test delete FTS and Rebuild FTS
	// 4. 测试删除和重建 FTS
	t.Run("FTS deletion", func(t *testing.T) {
		// Delete ID 1
		noteRepo.(*noteRepository).deleteFTS(userDb, 1, uid)

		ids, err := noteRepo.(*noteRepository).searchFTS(uid, vaultID, "intro", false, "mtime", "desc", 10, 0)
		require.NoError(t, err)
		assert.Empty(t, ids)
	})

	// 5. Test storeRaw config change and auto-rebuild
	// 5. 测试 storeRaw 配置改变与自动重建
	t.Run("Configuration change auto-rebuild", func(t *testing.T) {
		// Change configuration to storeRaw = false
		// 修改配置为 storeRaw = false
		daoInst.BleveMgr.storeRaw = false

		// Close index to evict it from memory cache
		// 关闭索引以从内存缓存中清除它
		err := daoInst.BleveMgr.Close(uid, vaultID)
		require.NoError(t, err)

		// Reopen/Get index. This should automatically delete the old index due to metadata mismatch.
		// 重新获取索引。这应该会因元数据不匹配而自动删除旧索引。
		index, err := daoInst.BleveMgr.GetIndex(uid, vaultID)
		require.NoError(t, err)

		// Document count should be 0 because it was cleaned up
		// 由于被清空了，文档计数应该为 0
		cnt, err := index.DocCount()
		require.NoError(t, err)
		assert.Equal(t, uint64(0), cnt)

		// Search FTS should automatically trigger RebuildVaultIndex when DocCount is 0
		// 当 DocCount 为 0 时，searchFTS 应该会自动触发 RebuildVaultIndex 重建
		ids, err := noteRepo.(*noteRepository).searchFTS(uid, vaultID, "tutorial", false, "mtime", "desc", 10, 0)
		require.NoError(t, err)
		assert.Len(t, ids, 1)
		assert.Equal(t, int64(2), ids[0])

		// Verify metadata now reflects the new storeRaw value (false)
		// 验证元数据现在是否反映了新的 storeRaw 值（false）
		metaPath := filepath.Join(daoInst.BleveMgr.GetIndexPath(uid, vaultID), "meta.json")
		metaData, err := os.ReadFile(metaPath)
		require.NoError(t, err)
		assert.Contains(t, string(metaData), `"fts-bleve-store-raw":false`)
	})
}
