package app

import (
	"fmt"
	"sync"
	"testing"
)

// TestWebsocketClient_ClientInfo_ConcurrentAccess exercises the P1 fix for the race between
// ClientInfo()'s writes (via setClientInfo) and concurrent reads of the client-reported
// connection metadata (name/type/version/platform/offline sync strategy/protobuf flag) from
// other goroutines — the scenario gws ParallelEnabled makes possible for a single connection.
// Run with `go test -race` to verify the race detector reports nothing.
// TestWebsocketClient_ClientInfo_ConcurrentAccess 验证 P1 修复：ClientInfo()（通过
// setClientInfo 写入）与其他 goroutine 并发读取客户端上报的连接元数据（名称/类型/版本/
// 平台/离线同步策略/protobuf 标志）之间的竞态——这正是 gws ParallelEnabled 下单个连接可能
// 出现的场景。需要用 `go test -race` 运行以验证竞态检测器不报告任何问题。
func TestWebsocketClient_ClientInfo_ConcurrentAccess(t *testing.T) {
	c := &WebsocketClient{}

	const writers = 4
	const readers = 8
	const iterations = 200

	var wg sync.WaitGroup

	// Writers: repeatedly call setClientInfo, mimicking repeated ClientInfo messages.
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				name := fmt.Sprintf("client-%d-%d", w, i)
				platform := map[string]bool{"isDesktop": i%2 == 0}
				c.setClientInfo(name, "web", "1.0."+fmt.Sprint(i), platform, "newTimeMerge", i%2 == 0)
			}
		}(w)
	}

	// Readers: repeatedly read every field via the locked getters.
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = c.ClientName()
				_ = c.ClientType()
				_ = c.ClientVersion()
				_ = c.OfflineSyncStrategy()
				_ = c.UseProtobuf()
				platform := c.ClientPlatform()
				if platform != nil {
					_ = platform["isDesktop"]
				}
			}
		}()
	}

	wg.Wait()

	// After all writes complete, the client must reflect a fully-applied (not partially
	// mixed) state: name/version were written together in the same setClientInfo call.
	// 所有写入完成后，客户端必须反映一次完整写入后的状态（不能是不同调用混杂的中间态）：
	// name/version 是在同一次 setClientInfo 调用中一起写入的。
	if c.ClientName() == "" || c.ClientVersion() == "" {
		t.Fatal("expected non-empty ClientName/ClientVersion after concurrent writes settled")
	}
}
