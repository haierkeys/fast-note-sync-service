package fileurl

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPickSourceTableDriven(t *testing.T) {
	// pickSource implements the threshold strategy:
	// - exactly one reachable -> pick it
	// - neither reachable -> default GitHub (task-layer fallback retries)
	// - both reachable -> switch only when slower >= 2x faster AND slower >= 200ms
	// pickSource 实现阈值策略，见上注释。
	cases := []struct {
		name     string
		g, c     ProbeResult
		expected bool // true = GitHub
	}{
		{"only github reachable", ProbeResult{OK: true, LatencyMs: 100}, ProbeResult{OK: false}, true},
		{"only cnb reachable", ProbeResult{OK: false}, ProbeResult{OK: true, LatencyMs: 100}, false},
		{"both down -> default github", ProbeResult{OK: false}, ProbeResult{OK: false}, true},

		// both reachable, latency close -> keep GitHub (anti-jitter)
		{"both close, github slightly faster", ProbeResult{OK: true, LatencyMs: 100}, ProbeResult{OK: true, LatencyMs: 120}, true},
		{"both close, cnb slightly faster", ProbeResult{OK: true, LatencyMs: 120}, ProbeResult{OK: true, LatencyMs: 100}, true},
		{"both fast equal", ProbeResult{OK: true, LatencyMs: 50}, ProbeResult{OK: true, LatencyMs: 50}, true},

		// both reachable, big difference -> switch to faster one
		{"github much slower -> switch to cnb", ProbeResult{OK: true, LatencyMs: 600}, ProbeResult{OK: true, LatencyMs: 100}, false},
		{"cnb much slower -> keep github", ProbeResult{OK: true, LatencyMs: 100}, ProbeResult{OK: true, LatencyMs: 600}, true},

		// difference is 2x but under floor (200ms) -> no switch (anti-jitter on fast networks)
		{"2x diff but under floor", ProbeResult{OK: true, LatencyMs: 30}, ProbeResult{OK: true, LatencyMs: 60}, true},

		// difference big enough and over floor -> switch
		{"github slow over floor 2x", ProbeResult{OK: true, LatencyMs: 450}, ProbeResult{OK: true, LatencyMs: 200}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := pickSource(tc.g, tc.c)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestSourceSelectorExplicitModes(t *testing.T) {
	s := NewSourceSelector(SourceGitHub)
	assert.True(t, s.IsGitHub())
	assert.Equal(t, SourceGitHub, s.Mode())

	s2 := NewSourceSelector(SourceCNB)
	assert.False(t, s2.IsGitHub())
	assert.Equal(t, SourceCNB, s2.Mode())
}

func TestSourceSelectorSetModeInvalidatesSnapshot(t *testing.T) {
	// Probe once to populate snapshot, then SetMode should clear it.
	// 探测一次填充快照，随后 SetMode 应清空它。
	s := NewSourceSelector(SourceAuto)
	s.Probe(context.Background())
	assert.NotNil(t, s.Snapshot(), "probe should populate a snapshot")

	s.SetMode(SourceGitHub)
	assert.Nil(t, s.Snapshot(), "SetMode must invalidate the cached snapshot")
	assert.True(t, s.IsGitHub())
}

func TestSourceSelectorProbePopulatesSnapshot(t *testing.T) {
	s := NewSourceSelector(SourceAuto)
	snap := s.Probe(context.Background())
	assert.Equal(t, snap.At, s.Snapshot().At, "Probe should cache the snapshot it returns")
	// GitHub/CNB reachability is environment-dependent; just assert structural fields exist.
	// GitHub/CNB 的可达性依赖网络环境，这里只断言结构字段存在。
}

func TestSourceSelectorConcurrentAccess(t *testing.T) {
	// Run with `go test -race` to detect data races on IsGitHub/Probe/Snapshot/SetMode.
	// 使用 `go test -race` 运行，检测 IsGitHub/Probe/Snapshot/SetMode 的数据竞争。
	s := NewSourceSelector(SourceAuto)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(4)
		go func() { defer wg.Done(); _ = s.IsGitHub() }()
		go func() { defer wg.Done(); _ = s.Snapshot() }()
		go func() { defer wg.Done(); s.Probe(context.Background()) }()
		go func(j int) {
			defer wg.Done()
			if j%2 == 0 {
				s.SetMode(SourceAuto)
			} else {
				s.SetMode(SourceGitHub)
			}
		}(i)
	}
	wg.Wait()
}
