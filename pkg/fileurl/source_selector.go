package fileurl

import (
	"context"
	"net/http"
	"sync"
	"time"
)

const (
	SourceAuto   = "auto"
	SourceGitHub = "github"
	SourceCNB    = "cnb"

	// GitHubProbeURL / CNBProbeURL are the real release endpoints used for probing.
	// 探测目标使用真实的 release 接口，而不是根域名 —— 根域名可达不代表 release 接口可达
	// （GitHub 对未鉴权的 /repos/.../releases 有 60 次/小时/IP 限流，命中限流根域名仍 200）。
	GitHubProbeURL = "https://api.github.com/repos/haierkeys/fast-note-sync-service/releases"
	CNBProbeURL    = "https://api.cnb.cool/haierkeys/fast-note-sync-service/-/releases"

	// defaultProbeTimeout is the per-source probe timeout.
	// 单源探测超时；超过即判定该源不可用，直接切换到另一源。
	defaultProbeTimeout = 3 * time.Second
	// defaultCacheTTL is how long a probe snapshot is reused before re-probing.
	// 探测快照缓存时长；避免后台任务每 10 分钟轮询时反复打探测请求。
	defaultCacheTTL = 5 * time.Minute

	// latencySwitchFactor: when both sources are reachable, the slower one must
	// be at least latencySwitchFactor × faster one AND exceed latencySwitchFloor
	// before we switch. Prevents jitter between two equally-fast sources.
	// 延迟差倍数阈值：双可达时，慢方 ≥ 2× 快方 且 慢方 ≥ 200ms 才切换，避免在两个
	// 延迟接近的源之间反复抖动。
	latencySwitchFactor = 2.0
	latencySwitchFloor  = int64(200) // ms
)

// ProbeResult is the reachability + latency result of a single source probe.
// ProbeResult 单源探测结果（可达性 + 延迟）。
type ProbeResult struct {
	OK        bool  // true if the source responded with a non-5xx status
	LatencyMs int64 // round-trip time in milliseconds (0 if request failed before any response)
}

// ProbeSnapshot is a point-in-time view of both sources produced by one parallel probe.
// ProbeSnapshot 一次并行探测两源的快照，供前端测速面板展示与选源决策复用。
type ProbeSnapshot struct {
	GitHub     ProbeResult
	CNB        ProbeResult
	UseGitHub  bool      // derived recommended source based on the threshold strategy
	At         time.Time // when this snapshot was taken
}

// SourceSelector handles the logic of selecting the data source (GitHub or CNB).
// SourceSelector 处理选择数据源（GitHub 或 CNB）的逻辑。
//
// Concurrency: IsGitHub / Probe / Snapshot / SetMode are all goroutine-safe.
// 并发安全：所有公开方法均可被多个 goroutine 并发调用（后台任务、HTTP handler、
// WebSocket 广播都会调用 IsGitHub）。
type SourceSelector struct {
	mu       sync.Mutex
	mode     string
	snapshot *ProbeSnapshot
	cacheTTL time.Duration
}

// NewSourceSelector creates a new SourceSelector.
// NewSourceSelector 创建一个新的 SourceSelector。
func NewSourceSelector(mode string) *SourceSelector {
	return &SourceSelector{
		mode:     mode,
		cacheTTL: defaultCacheTTL,
	}
}

// SetMode updates the selection mode (auto | github | cnb) at runtime.
// SetMode 运行时更新选源模式（auto | github | cnb），例如配置热重载后。
func (s *SourceSelector) SetMode(mode string) {
	s.mu.Lock()
	s.mode = mode
	s.snapshot = nil // invalidate cached snapshot so the new mode takes effect immediately
	s.mu.Unlock()
}

// IsGitHub returns whether the current source should be GitHub.
// IsGitHub 返回当前推荐源是否为 GitHub。
//
// For explicit github/cnb modes it returns a fixed answer. For auto mode it
// reuses a cached snapshot within cacheTTL, otherwise probes synchronously.
// 显式 github/cnb 模式返回固定结果；auto 模式在缓存有效期内复用快照，否则同步探测。
func (s *SourceSelector) IsGitHub() bool {
	s.mu.Lock()
	mode := s.mode
	s.mu.Unlock()
	switch mode {
	case SourceGitHub:
		return true
	case SourceCNB:
		return false
	default:
		return s.recentSnapshot(context.Background()).UseGitHub
	}
}

// Probe forces a fresh parallel probe of both sources (bypasses cache) and
// caches the result. Intended for the webgui "test latency" button.
// Probe 立即重新并行探测两源（绕过缓存）并缓存结果，供前端「测试延迟」按钮调用。
func (s *SourceSelector) Probe(ctx context.Context) ProbeSnapshot {
	snap := probeBoth(ctx)
	s.mu.Lock()
	s.snapshot = &snap
	s.mu.Unlock()
	return snap
}

// Snapshot returns the cached probe snapshot if still fresh, otherwise nil.
// Snapshot 返回仍在有效期内的缓存快照；过期或从未探测过则返回 nil（前端展示用）。
func (s *SourceSelector) Snapshot() *ProbeSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot == nil || time.Since(s.snapshot.At) > s.cacheTTL {
		return nil
	}
	snap := *s.snapshot
	return &snap
}

// Mode returns the current selection mode.
// Mode 返回当前选源模式。
func (s *SourceSelector) Mode() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mode
}

// recentSnapshot returns a fresh-enough snapshot, probing synchronously when the
// cache is stale. Probe uses a bounded timeout derived from ctx (or the default).
// recentSnapshot 返回足够新的快照；缓存过期时同步探测一次。
func (s *SourceSelector) recentSnapshot(ctx context.Context) ProbeSnapshot {
	s.mu.Lock()
	snap := s.snapshot
	mode := s.mode
	s.mu.Unlock()
	if snap != nil && time.Since(snap.At) <= s.cacheTTL {
		return *snap
	}
	// Explicit modes don't need probing for selection, but still probe so that
	// the frontend latency panel works regardless of mode.
	// 显式模式下选源是固定的，但仍探测一次以便前端测速面板能展示延迟。
	fresh := probeBoth(ctx)
	if mode != SourceGitHub && mode != SourceCNB {
		// auto: snapshot drives selection, cache it
		// auto 模式：快照驱动选源，需要缓存
		s.mu.Lock()
		s.snapshot = &fresh
		s.mu.Unlock()
	}
	return fresh
}

// probeBoth probes GitHub and CNB in parallel and derives the recommended source.
// probeBoth 并行探测 GitHub 与 CNB，并基于阈值策略推导推荐源。
func probeBoth(ctx context.Context) ProbeSnapshot {
	ctx, cancel := context.WithTimeout(ctx, defaultProbeTimeout+500*time.Millisecond)
	defer cancel()

	var g, c ProbeResult
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		g = probeOne(ctx, GitHubProbeURL, "")
	}()
	go func() {
		defer wg.Done()
		c = probeOne(ctx, CNBProbeURL, "")
	}()
	wg.Wait()

	return ProbeSnapshot{
		GitHub:    g,
		CNB:       c,
		UseGitHub: pickSource(g, c),
		At:        time.Now(),
	}
}

// pickSource implements the threshold strategy:
//  1. if exactly one source is reachable, pick it;
//  2. if neither is reachable, default to GitHub and let the task-layer fallback retry handle it;
//  3. if both are reachable, switch only when the slower one is at least
//     latencySwitchFactor × the faster one AND exceeds latencySwitchFloor (anti-jitter).
//
// pickSource 实现阈值策略：
//  1. 仅一源可达 → 选它；
//  2. 两源都不可达 → 默认 GitHub，交由 task 层 fallback 重试兜底；
//  3. 双可达 → 仅当慢方 ≥ 2× 快方 且 ≥ 200ms 才切，避免抖动。
func pickSource(g, c ProbeResult) bool {
	switch {
	case g.OK && !c.OK:
		return true
	case !g.OK && c.OK:
		return false
	case !g.OK && !c.OK:
		return true
	default:
		fast, slow := g.LatencyMs, c.LatencyMs
		slowIsGitHub := false
		if c.LatencyMs < g.LatencyMs {
			fast, slow = c.LatencyMs, g.LatencyMs
			slowIsGitHub = true
		}
		if slow >= latencySwitchFloor && float64(slow) >= latencySwitchFactor*float64(fast) {
			// switch to the faster one
			// 切到更快的那一方
			return !slowIsGitHub
		}
		return true // close enough, keep GitHub as default
	}
}

// probeOne issues a HEAD request against url with a bounded timeout and reports
// reachability + latency. A 2xx/3xx response counts as reachable (CNB may 302).
// probeOne 对 url 发起带超时的 HEAD 请求，返回可达性与延迟；2xx/3xx 视为可达（CNB 可能 302）。
func probeOne(ctx context.Context, url, token string) ProbeResult {
	cli := &http.Client{Timeout: defaultProbeTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return ProbeResult{OK: false}
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	start := time.Now()
	resp, err := cli.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return ProbeResult{OK: false, LatencyMs: latency}
	}
	defer resp.Body.Close()
	ok := resp.StatusCode < http.StatusInternalServerError
	return ProbeResult{OK: ok, LatencyMs: latency}
}
