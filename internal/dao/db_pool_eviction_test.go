package dao

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestGetOrCreateDB_DoesNotEvictRecentlyUsedConnection verifies the P0-adjacent fix:
// when the cached-connection cap is exceeded, GetOrCreateDB must not evict a connection
// that was checked out more recently than dbConnMinIdleBeforeEvict — even though it may
// still be the "oldest lastUsed" entry among a small cache. This protects a long-running
// query from having its connection closed out from under it.
// TestGetOrCreateDB_DoesNotEvictRecentlyUsedConnection 验证防误关修复：
// 缓存连接数超过上限时，GetOrCreateDB 不能淘汰一个刚被取出、空闲时间小于
// dbConnMinIdleBeforeEvict 的连接——即便它在小缓存里恰好是"最久未用"的一个。
// 用于保护一个仍在跑的慢查询，不被从底下关掉连接。
func TestGetOrCreateDB_DoesNotEvictRecentlyUsedConnection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fast-note-sync-service-dbpool-test-*")
	require.NoError(t, err)
	origWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		_ = os.Chdir(origWd)
		_ = os.RemoveAll(tempDir)
	}()
	require.NoError(t, os.MkdirAll(filepath.Join("storage", "database"), 0755))

	dbPath := filepath.Join("storage", "database", "db.sqlite3")
	mainDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)
	defer func() {
		if sqlDB, err := mainDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	dbCfg := &config.DatabaseConfig{
		Type:             "sqlite",
		Path:             dbPath,
		EnableWriteQueue: util.Ptr(false),
	}

	// Cap of 2 cached connections, and a min-idle-before-evict of 200ms so the test can
	// deterministically control which entries are "idle long enough" without a real clock wait.
	// 缓存上限为 2，最短空闲淘汰阈值 200ms，测试用短暂 sleep 就能确定性地控制
	// 哪些连接"已经空闲够久"。
	daoInst := New(mainDB, context.Background(),
		WithConfig(dbCfg),
		WithUserDatabaseConfig(dbCfg),
		WithLogger(zap.NewNop()),
		WithMaxCachedDBConns(2),
		WithDBConnMinIdleBeforeEvict(200*time.Millisecond),
	)

	// Fill the cache to the cap with two connections, then let them age past the
	// min-idle threshold so they become eligible for eviction.
	// 先建满 2 个连接把缓存填到上限，然后让它们空闲超过阈值，使其可被淘汰。
	require.NotNil(t, daoInst.GetOrCreateDB("tenant-a"))
	require.NotNil(t, daoInst.GetOrCreateDB("tenant-b"))
	time.Sleep(250 * time.Millisecond)

	// "Check out" tenant-a again right before the cap-triggering call, simulating a
	// long-running query still holding onto it (lastUsed refreshed at checkout time).
	// 在触发淘汰的调用之前，重新"取出" tenant-a，模拟一个慢查询仍在用它
	// （lastUsed 会在取出时刷新）。
	require.NotNil(t, daoInst.GetOrCreateDB("tenant-a"))

	// Adding a third distinct key pushes the cache over the cap of 2. tenant-b is the
	// only one idle long enough (250ms > 200ms threshold); tenant-a was just refreshed
	// and must survive.
	// 新增第三个 key 会把缓存推过上限 2。tenant-b 是唯一空闲时间超过阈值（250ms >
	// 200ms）的连接；tenant-a 刚被刷新，必须存活下来。
	require.NotNil(t, daoInst.GetOrCreateDB("tenant-c"))

	daoInst.mu.RLock()
	_, aStillCached := daoInst.KeyDb["tenant-a"]
	_, bStillCached := daoInst.KeyDb["tenant-b"]
	_, cStillCached := daoInst.KeyDb["tenant-c"]
	daoInst.mu.RUnlock()

	require.True(t, aStillCached, "tenant-a was checked out recently and must not be evicted")
	require.False(t, bStillCached, "tenant-b was idle past the threshold and should have been evicted")
	require.True(t, cStillCached, "tenant-c was just created and must be present")
}

// TestGetOrCreateDB_SkipsEvictionWhenNothingIdleLongEnough verifies that when every
// cached connection is "too fresh" to evict, GetOrCreateDB lets the cache temporarily
// exceed its cap rather than closing a connection that might still be in use.
// TestGetOrCreateDB_SkipsEvictionWhenNothingIdleLongEnough 验证当所有缓存连接
// 都"太新"不该被淘汰时，GetOrCreateDB 宁可让缓存暂时超过上限，也不关掉
// 可能仍在使用的连接。
func TestGetOrCreateDB_SkipsEvictionWhenNothingIdleLongEnough(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fast-note-sync-service-dbpool-test-*")
	require.NoError(t, err)
	origWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		_ = os.Chdir(origWd)
		_ = os.RemoveAll(tempDir)
	}()
	require.NoError(t, os.MkdirAll(filepath.Join("storage", "database"), 0755))

	dbPath := filepath.Join("storage", "database", "db.sqlite3")
	mainDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)
	defer func() {
		if sqlDB, err := mainDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	dbCfg := &config.DatabaseConfig{
		Type:             "sqlite",
		Path:             dbPath,
		EnableWriteQueue: util.Ptr(false),
	}

	daoInst := New(mainDB, context.Background(),
		WithConfig(dbCfg),
		WithUserDatabaseConfig(dbCfg),
		WithLogger(zap.NewNop()),
		WithMaxCachedDBConns(1),
		WithDBConnMinIdleBeforeEvict(time.Hour), // effectively nothing qualifies within this test
	)

	require.NotNil(t, daoInst.GetOrCreateDB("tenant-a"))
	require.NotNil(t, daoInst.GetOrCreateDB("tenant-b"))

	daoInst.mu.RLock()
	count := len(daoInst.KeyDb)
	_, aStillCached := daoInst.KeyDb["tenant-a"]
	daoInst.mu.RUnlock()

	require.Equal(t, 2, count, "cache should be allowed to exceed the cap when nothing is idle long enough to evict")
	require.True(t, aStillCached, "tenant-a must not have been evicted")
}
