package app

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// 验证 DiffMergePaths 的检查-删除操作是原子的

func TestProperty7_DiffMergePathsAtomicOperation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// 并发访问时，每个 path 只能被处理一次
	properties.Property("each path processed exactly once under concurrent access", prop.ForAll(
		func(pathCount int) bool {
			if pathCount <= 0 {
				return true
			}

			// 生成唯一路径
			paths := make([]string, pathCount)
			for i := 0; i < pathCount; i++ {
				paths[i] = "path_" + string(rune('a'+i%26)) + string(rune('0'+i/26))
			}

			client := &WebsocketClient{
				DiffMergePaths: make(map[string]DiffMergeEntry),
			}

			// 预填充所有路径
			for _, p := range paths {
				client.DiffMergePaths[p] = DiffMergeEntry{CreatedAt: time.Now()}
			}

			// 记录每个 path 被处理的次数
			processCount := make(map[string]*int32)
			for _, p := range paths {
				var count int32 = 0
				processCount[p] = &count
			}

			// 并发尝试处理每个 path
			var wg sync.WaitGroup
			goroutines := 10

			for i := 0; i < goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for _, p := range paths {
						// 原子检查-删除操作
						client.DiffMergePathsMu.Lock()
						_, ok := client.DiffMergePaths[p]
						if ok {
							delete(client.DiffMergePaths, p)
						}
						client.DiffMergePathsMu.Unlock()

						if ok {
							atomic.AddInt32(processCount[p], 1)
						}
					}
				}()
			}

			wg.Wait()

			// 验证每个 path 只被处理一次
			for _, p := range paths {
				if *processCount[p] != 1 {
					t.Logf("Path %s processed %d times, expected 1", p, *processCount[p])
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t)
}

// 验证超时清理机制

func TestProperty8_DiffMergePathsCleanup(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// 超时条目被清理，未超时条目保留
	properties.Property("expired entries are cleaned, non-expired are kept", prop.ForAll(
		func(expiredCount, nonExpiredCount int) bool {
			client := &WebsocketClient{
				DiffMergePaths: make(map[string]DiffMergeEntry),
			}

			timeout := 100 * time.Millisecond
			now := time.Now()

			// 添加过期条目
			for i := 0; i < expiredCount; i++ {
				path := "expired_" + string(rune('a'+i))
				client.DiffMergePaths[path] = DiffMergeEntry{
					CreatedAt: now.Add(-timeout - time.Second), // 已过期
				}
			}

			// 添加未过期条目
			for i := 0; i < nonExpiredCount; i++ {
				path := "active_" + string(rune('a'+i))
				client.DiffMergePaths[path] = DiffMergeEntry{
					CreatedAt: now, // 未过期
				}
			}

			// 执行清理
			cleaned := client.CleanupExpiredDiffMergePaths(timeout)

			// 验证清理数量
			if cleaned != expiredCount {
				t.Logf("Cleaned %d, expected %d", cleaned, expiredCount)
				return false
			}

			// 验证剩余数量
			if len(client.DiffMergePaths) != nonExpiredCount {
				t.Logf("Remaining %d, expected %d", len(client.DiffMergePaths), nonExpiredCount)
				return false
			}

			return true
		},
		gen.IntRange(0, 10),
		gen.IntRange(0, 10),
	))

	properties.TestingRun(t)
}

// 单元测试: DiffMergePaths 基本操作
func TestDiffMergePaths_BasicOperations(t *testing.T) {
	client := &WebsocketClient{
		DiffMergePaths: make(map[string]DiffMergeEntry),
	}

	// 测试添加
	path := "test/note.md"
	client.DiffMergePathsMu.Lock()
	client.DiffMergePaths[path] = DiffMergeEntry{CreatedAt: time.Now()}
	client.DiffMergePathsMu.Unlock()

	// 测试检查存在
	client.DiffMergePathsMu.RLock()
	_, exists := client.DiffMergePaths[path]
	client.DiffMergePathsMu.RUnlock()

	if !exists {
		t.Error("Path should exist after adding")
	}

	// 测试原子检查-删除
	client.DiffMergePathsMu.Lock()
	_, ok := client.DiffMergePaths[path]
	if ok {
		delete(client.DiffMergePaths, path)
	}
	client.DiffMergePathsMu.Unlock()

	if !ok {
		t.Error("Path should have been found and deleted")
	}

	// 验证已删除
	client.DiffMergePathsMu.RLock()
	_, exists = client.DiffMergePaths[path]
	client.DiffMergePathsMu.RUnlock()

	if exists {
		t.Error("Path should not exist after deletion")
	}
}

// 单元测试: ClearAllDiffMergePaths
func TestClearAllDiffMergePaths(t *testing.T) {
	client := &WebsocketClient{
		DiffMergePaths: make(map[string]DiffMergeEntry),
	}

	// 添加多个条目
	paths := []string{"a.md", "b.md", "c.md"}
	for _, p := range paths {
		client.DiffMergePaths[p] = DiffMergeEntry{CreatedAt: time.Now()}
	}

	// 清理所有
	count := client.ClearAllDiffMergePaths()

	if count != len(paths) {
		t.Errorf("ClearAllDiffMergePaths() = %d, want %d", count, len(paths))
	}

	if len(client.DiffMergePaths) != 0 {
		t.Errorf("DiffMergePaths should be empty after clear, got %d", len(client.DiffMergePaths))
	}
}

// 单元测试: CleanupExpiredDiffMergePaths
func TestCleanupExpiredDiffMergePaths(t *testing.T) {
	client := &WebsocketClient{
		DiffMergePaths: make(map[string]DiffMergeEntry),
	}

	timeout := 50 * time.Millisecond

	// 添加一个过期条目
	client.DiffMergePaths["expired.md"] = DiffMergeEntry{
		CreatedAt: time.Now().Add(-100 * time.Millisecond),
	}

	// 添加一个未过期条目
	client.DiffMergePaths["active.md"] = DiffMergeEntry{
		CreatedAt: time.Now(),
	}

	// 执行清理
	cleaned := client.CleanupExpiredDiffMergePaths(timeout)

	if cleaned != 1 {
		t.Errorf("CleanupExpiredDiffMergePaths() = %d, want 1", cleaned)
	}

	if len(client.DiffMergePaths) != 1 {
		t.Errorf("Should have 1 remaining entry, got %d", len(client.DiffMergePaths))
	}

	if _, exists := client.DiffMergePaths["active.md"]; !exists {
		t.Error("active.md should still exist")
	}
}
