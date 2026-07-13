package config

import "testing"

func intPtr(v int) *int { return &v }

// TestPipelineWindowClamped covers the S1 read-time clamp rule from the sync pipeline
// design (§7.1 S1 acceptance): negative values are treated as 0 (disabled / stop-and-wait),
// values within range pass through unchanged, and values above the ceiling are capped
// (up<=32, down<=16). Fields are *int (explicit-0-vs-unset under LoadConfig's second
// defaults.Set pass, see field comments), so a nil pointer must fall back to the defaults
// defensively.
func TestPipelineWindowClamped(t *testing.T) {
	cases := []struct {
		name     string
		up       *int
		down     *int
		wantUp   int
		wantDown int
	}{
		{"defaults", intPtr(8), intPtr(4), 8, 4},
		{"explicit zero disables", intPtr(0), intPtr(0), 0, 0},
		{"negative treated as zero", intPtr(-1), intPtr(-5), 0, 0},
		{"within range passes through", intPtr(32), intPtr(16), 32, 16},
		{"above ceiling clamped", intPtr(100), intPtr(100), 32, 16},
		{"nil falls back to defaults", nil, nil, 8, 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := AppSettings{PipelineWindowUp: tc.up, PipelineWindowDown: tc.down}
			if got := a.PipelineWindowUpClamped(); got != tc.wantUp {
				t.Errorf("PipelineWindowUpClamped() = %d, want %d", got, tc.wantUp)
			}
			if got := a.PipelineWindowDownClamped(); got != tc.wantDown {
				t.Errorf("PipelineWindowDownClamped() = %d, want %d", got, tc.wantDown)
			}
		})
	}
}
