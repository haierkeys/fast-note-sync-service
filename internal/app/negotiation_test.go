package app

import (
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/config"
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
		PipelineWindowUp:   8,
		PipelineWindowDown: 4,
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
	a := newTestApp(&config.AppSettings{PipelineWindowUp: 999, PipelineWindowDown: -5})

	up, down := a.PipelineWindows()
	if up != 32 {
		t.Fatalf("PipelineWindows() up = %d, want clamped to 32", up)
	}
	if down != 0 {
		t.Fatalf("PipelineWindows() down = %d, want clamped to 0", down)
	}
}
