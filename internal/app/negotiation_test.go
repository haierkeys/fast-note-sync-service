package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/stretchr/testify/require"
)

// newTestApp builds a minimal *App wrapping only the config, enough to exercise config-backed
// AppContainer getters without going through the full NewApp dependency wiring (DB, workers, etc).
func newTestApp(cfg *config.AppSettings) *App {
	return &App{
		Infra: &Infra{
			config: &AppConfig{App: *cfg},
		},
	}
}

// TestApp_SyncChunkNums_And_PipelineWindows covers the S1 AppContainer getters consumed by the
// auth handshake negotiation block (design §2.3/§7.1 S1): SyncChunkNums must pass the configured
// batch sizes through as-is, and PipelineWindows must apply the read-time clamp so a live admin
// misconfiguration (e.g. pipeline-window-up: 999) can never leak an out-of-range value into the
// wire protocol.
func TestApp_SyncChunkNums_And_PipelineWindows(t *testing.T) {
	a := newTestApp(&config.AppSettings{
		SyncUpChunkNum:     100,
		SyncDownChunkNum:   200,
		PipelineWindowUp:   util.Ptr(8),
		PipelineWindowDown: util.Ptr(4),
	})

	up, down := a.SyncChunkNums()
	if up != 100 || down != 200 {
		t.Fatalf("SyncChunkNums() = (%d, %d), want (100, 200)", up, down)
	}

	pwUp, pwDown := a.PipelineWindows()
	if pwUp != 8 || pwDown != 4 {
		t.Fatalf("PipelineWindows() = (%d, %d), want (8, 4)", pwUp, pwDown)
	}
}

// TestApp_PipelineWindows_ClampsOutOfRange guards the "运行时回滚开关" story (design §8): an
// admin can set pipeline-window-up/down to any int via the config API, and the getter must
// clamp it before it ever reaches a client, regardless of what's stored in memory.
func TestApp_PipelineWindows_ClampsOutOfRange(t *testing.T) {
	a := newTestApp(&config.AppSettings{PipelineWindowUp: util.Ptr(999), PipelineWindowDown: util.Ptr(-5)})

	up, down := a.PipelineWindows()
	if up != 32 {
		t.Fatalf("PipelineWindows() up = %d, want clamped to 32", up)
	}
	if down != 0 {
		t.Fatalf("PipelineWindows() down = %d, want clamped to 0", down)
	}
}

// loadConfigApp loads the given yaml through the real LoadConfig path (defaults.Set →
// yaml.Unmarshal → second defaults.Set) and wraps it into a minimal *App, so the assertions
// below exercise exactly what a real process boot would see.
func loadConfigApp(t *testing.T, yaml string) *App {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(yaml), 0644))

	cfg, _, err := LoadConfig(configPath)
	require.NoError(t, err)

	return &App{Infra: &Infra{config: cfg}}
}

// TestLoadConfig_ExplicitZeroPipelineWindows_DisablesWindow is the direct regression lock for
// the rollback-switch bug: LoadConfig runs defaults.Set a second time after yaml.Unmarshal, and
// with plain int fields an explicit `pipeline-window-up: 0` was indistinguishable from an unset
// field and got silently overwritten back to 8/4 — meaning a container deployment configured
// purely via config file could never turn the window pipeline off. With *int fields the
// explicit 0 must survive the full real load path and reach the negotiation getter as (0,0).
// TestLoadConfig_ExplicitZeroPipelineWindows_DisablesWindow 是回滚开关 bug 的直接回归锁：
// LoadConfig 在 yaml.Unmarshal 之后二次 defaults.Set，普通 int 字段下显式
// `pipeline-window-up: 0` 与未写字段无法区分、被静默覆盖回 8/4——纯配置文件部署的容器从一开始
// 就关不掉窗口。改 *int 后，显式 0 必须穿过完整真实加载路径，在协商 getter 处得到 (0,0)。
func TestLoadConfig_ExplicitZeroPipelineWindows_DisablesWindow(t *testing.T) {
	a := loadConfigApp(t, `
app:
  pipeline-window-up: 0
  pipeline-window-down: 0
`)

	up, down := a.PipelineWindows()
	if up != 0 || down != 0 {
		t.Fatalf("PipelineWindows() after loading explicit-zero yaml = (%d, %d), want (0, 0) — explicit 0 was overwritten by the second defaults.Set pass", up, down)
	}
}

// TestLoadConfig_UnsetPipelineWindows_UsesDefaults is the counterpart: a yaml that doesn't
// mention the keys at all must come out as the defaults (8,4) via the same real load path.
func TestLoadConfig_UnsetPipelineWindows_UsesDefaults(t *testing.T) {
	a := loadConfigApp(t, `
app:
  sync-up-chunk-num: 100
`)

	up, down := a.PipelineWindows()
	if up != 8 || down != 4 {
		t.Fatalf("PipelineWindows() after loading yaml without the keys = (%d, %d), want defaults (8, 4)", up, down)
	}
}

// TestLoadConfig_PipelineWindows_SaveRoundTrip guards the admin flow that made this bug sneaky:
// runtime set-to-0 worked (pointer semantics in the handler), but cfg.Save() + restart silently
// restored 8/4. Simulate it: load, set 0 (as UpdateConfig does), Save, re-load through the real
// path, and require (0,0) to survive.
func TestLoadConfig_PipelineWindows_SaveRoundTrip(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("app:\n  sync-up-chunk-num: 100\n"), 0644))

	cfg, _, err := LoadConfig(configPath)
	require.NoError(t, err)

	// Mirror handler_admin_control.UpdateConfig's pointer assignment for a runtime rollback.
	cfg.App.PipelineWindowUp = util.Ptr(0)
	cfg.App.PipelineWindowDown = util.Ptr(0)
	require.NoError(t, cfg.Save())

	reloaded, _, err := LoadConfig(configPath)
	require.NoError(t, err)

	a := &App{Infra: &Infra{config: reloaded}}
	up, down := a.PipelineWindows()
	if up != 0 || down != 0 {
		t.Fatalf("PipelineWindows() after Save+reload of a runtime rollback = (%d, %d), want (0, 0)", up, down)
	}
}
