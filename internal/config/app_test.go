package config

import "testing"

// TestPipelineWindowClamped covers the S1 read-time clamp rule from the sync pipeline
// design (§7.1 S1 acceptance): negative values are treated as 0 (disabled / stop-and-wait),
// values within range pass through unchanged, and values above the ceiling are capped
// (up<=32, down<=16).
func TestPipelineWindowClamped(t *testing.T) {
	cases := []struct {
		name     string
		up       int
		down     int
		wantUp   int
		wantDown int
	}{
		{"defaults", 8, 4, 8, 4},
		{"zero disables", 0, 0, 0, 0},
		{"negative treated as zero", -1, -5, 0, 0},
		{"within range passes through", 32, 16, 32, 16},
		{"above ceiling clamped", 100, 100, 32, 16},
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
