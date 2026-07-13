package websocket_router

import (
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"go.uber.org/zap"
)

// withFakeSendSyncPage swaps sendSyncPageFunc for the duration of the test with a fake that
// records the SentPage of every call instead of writing real frames (this codebase's own
// convention: existing WebsocketClient tests, e.g. pkg/app/websocket_client_test.go, never
// exercise the real conn write path either — see the seam's doc comment on sendSyncPageFunc).
// The fake mirrors the real sendSyncPage's isLast rule: the page at totalPages()-1 is always
// the last one, regardless of whether the final page is a partial page.
func withFakeSendSyncPage(t *testing.T, sentLog *[]int) {
	t.Helper()
	orig := sendSyncPageFunc
	sendSyncPageFunc = func(c *pkgapp.WebsocketClient, entry *syncDownloadEntry) bool {
		*sentLog = append(*sentLog, entry.SentPage)
		return entry.SentPage == entry.totalPages()-1
	}
	t.Cleanup(func() { sendSyncPageFunc = orig })
}

// newTestDownloadEntry builds a syncDownloadEntry with numPages pages of 1 message each (content
// is irrelevant to pump/handlePageAck, which never look inside MessageQueue — only
// sendSyncPageFunc's real implementation does, and it's faked out in these tests).
func newTestDownloadEntry(context string, numPages int, window int) *syncDownloadEntry {
	return &syncDownloadEntry{
		Context:      context,
		TypeName:     "note",
		Vault:        "test-vault",
		MessageQueue: make([]dto.WSQueuedMessage, numPages), // PageSize=1 => 1 page each
		PageSize:     1,
		Window:       window,
	}
}

// TestTotalPages covers the totalPages() helper's edge cases: exact multiples, remainders, and
// the PageSize<=0 defensive case.
func TestTotalPages(t *testing.T) {
	cases := []struct {
		name      string
		queueLen  int
		pageSize  int
		wantPages int
	}{
		{"exact multiple", 10, 5, 2},
		{"remainder rounds up", 11, 5, 3},
		{"single message", 1, 5, 1},
		{"empty queue", 0, 5, 0},
		{"pageSize zero defensive", 10, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &syncDownloadEntry{MessageQueue: make([]dto.WSQueuedMessage, tc.queueLen), PageSize: tc.pageSize}
			if got := e.totalPages(); got != tc.wantPages {
				t.Fatalf("totalPages() = %d, want %d", got, tc.wantPages)
			}
		})
	}
}

// TestPump_WindowZero_StopAndWaitEquivalence covers design §4.2's central equivalence claim:
// with Window=0, pump sends exactly one page per call and never runs ahead of AckedPage — this
// is required to be message-for-message identical to pre-3.6.0 stop-and-wait behavior.
func TestPump_WindowZero_StopAndWaitEquivalence(t *testing.T) {
	var sentLog []int
	withFakeSendSyncPage(t, &sentLog)

	entry := newTestDownloadEntry("ctx-w0", 5, 0)

	pump(nil, entry) // first pull equivalent: AckedPage already 0
	if got := sentLog; len(got) != 1 || got[0] != 0 {
		t.Fatalf("first pump() with Window=0: sent %v, want [0]", got)
	}
	if entry.SentPage != 1 {
		t.Fatalf("SentPage after first pump = %d, want 1", entry.SentPage)
	}

	// Advancing AckedPage one page at a time must release exactly one more page per call.
	for wantPage := 1; wantPage < 5; wantPage++ {
		entry.AckedPage = wantPage
		pump(nil, entry)
		if len(sentLog) != wantPage+1 {
			t.Fatalf("after acking page %d: sent %v, want %d pages total", wantPage-1, sentLog, wantPage+1)
		}
		if got := sentLog[len(sentLog)-1]; got != wantPage {
			t.Fatalf("after acking page %d: last sent page = %d, want %d", wantPage-1, got, wantPage)
		}
	}
}

// TestPump_Window4_BurstsAndCatchesUp exercises the W_down=4 timeline from design §4.1: first
// pull sends pages 0..3, then a single ack(0) releases exactly page 4.
func TestPump_Window4_BurstsAndCatchesUp(t *testing.T) {
	var sentLog []int
	withFakeSendSyncPage(t, &sentLog)

	entry := newTestDownloadEntry("ctx-w4", 10, 4)

	pump(nil, entry) // first pull
	want := []int{0, 1, 2, 3}
	if !intSliceEqual(sentLog, want) {
		t.Fatalf("first pump() with Window=4: sent %v, want %v", sentLog, want)
	}

	entry.AckedPage = 1 // ack(0) => AckedPage=1
	pump(nil, entry)
	want = append(want, 4)
	if !intSliceEqual(sentLog, want) {
		t.Fatalf("after ack(0): sent %v, want %v", sentLog, want)
	}
}

// TestPump_OutOfOrderAck_AdvancesToHighestWatermark covers design §4.3's "highest ack watermark
// wins" rule and the S4 acceptance example: an out-of-order ack(2) arriving first must advance
// AckedPage to 3 (not 3 separate small advances) and release pages up through 6.
func TestPump_OutOfOrderAck_AdvancesToHighestWatermark(t *testing.T) {
	var sentLog []int
	withFakeSendSyncPage(t, &sentLog)

	entry := newTestDownloadEntry("ctx-ooo", 10, 4)
	pump(nil, entry) // sends 0..3

	entry.AckedPage = 3 // ack(2) => AckedPage = 2+1 = 3
	pump(nil, entry)

	want := []int{0, 1, 2, 3, 4, 5, 6}
	if !intSliceEqual(sentLog, want) {
		t.Fatalf("after out-of-order ack(2): sent %v, want %v", sentLog, want)
	}
	if entry.SentPage != 7 {
		t.Fatalf("SentPage = %d, want 7", entry.SentPage)
	}
}

// TestPump_IsLast_DeletesEntry_AndStops verifies the isLast page terminates the pump loop and
// destroys the cache entry (design §4.1/§4.5: "isLast 页发完即删 cache").
func TestPump_IsLast_DeletesEntry_AndStops(t *testing.T) {
	var sentLog []int
	withFakeSendSyncPage(t, &sentLog)

	ctx := "ctx-islast-" + t.Name()
	entry := newTestDownloadEntry(ctx, 3, 10) // window bigger than totalPages: would try to send all 3 in one go
	syncDownloadStore(ctx, "note", entry)

	pump(nil, entry)

	want := []int{0, 1, 2}
	if !intSliceEqual(sentLog, want) {
		t.Fatalf("sent %v, want %v (pump must stop right after the isLast page)", sentLog, want)
	}
	if _, ok := syncDownloadGet(ctx, "note"); ok {
		t.Fatal("expected entry to be deleted from cache after isLast page was sent")
	}
}

// TestHandlePageAck_TableDriven covers every branch of the §4.2 PageAck branch table on a
// pre-seeded entry state, asserting both the resulting AckedPage/SentPage and which pages (if
// any) get (re)sent.
func TestHandlePageAck_TableDriven(t *testing.T) {
	cases := []struct {
		name          string
		numPages      int
		window        int
		seedAckedPage int
		seedSentPage  int
		pageIndex     int
		wantSent      []int
		wantAckedPage int
		wantSentPage  int
	}{
		{
			// -1 only ever arrives once, right after the entry is freshly constructed
			// (SentPage==0 from the zero value) — design §4.1's "首拉". pump() only resets
			// AckedPage on -1, it does NOT reset SentPage, so seeding a nonzero SentPage here
			// would model a scenario the real protocol never produces.
			name:     "first pull (-1) on a fresh entry sends one window's worth",
			numPages: 10, window: 4,
			seedAckedPage: 0, seedSentPage: 0,
			pageIndex:     -1,
			wantSent:      []int{0, 1, 2, 3},
			wantAckedPage: 0,
			wantSentPage:  4,
		},
		{
			name:     "expired duplicate ack (n < AckedPage-1) is ignored",
			numPages: 10, window: 4,
			seedAckedPage: 5, seedSentPage: 8,
			pageIndex:     2, // AckedPage-1 = 4, 2 < 4
			wantSent:      nil,
			wantAckedPage: 5,
			wantSentPage:  8,
		},
		{
			name:     "retransmitted previous ack (n == AckedPage-1) rewinds and resends the window",
			numPages: 10, window: 4,
			seedAckedPage: 3, seedSentPage: 7,
			pageIndex:     2, // AckedPage-1 == 2
			wantSent:      []int{3, 4, 5, 6},
			wantAckedPage: 3,
			wantSentPage:  7,
		},
		{
			name:     "normal advance (AckedPage-1 < n < SentPage) moves watermark and pumps",
			numPages: 10, window: 4,
			seedAckedPage: 0, seedSentPage: 4,
			pageIndex:     0,
			wantSent:      []int{4},
			wantAckedPage: 1,
			wantSentPage:  5,
		},
		{
			name:     "out-of-order advance jumps watermark straight to n+1",
			numPages: 10, window: 4,
			seedAckedPage: 0, seedSentPage: 4,
			pageIndex:     2,
			wantSent:      []int{4, 5, 6},
			wantAckedPage: 3,
			wantSentPage:  7,
		},
		{
			name:     "ack for a page never sent (n >= SentPage) is ignored",
			numPages: 10, window: 4,
			seedAckedPage: 0, seedSentPage: 4,
			pageIndex:     4, // == SentPage
			wantSent:      nil,
			wantAckedPage: 0,
			wantSentPage:  4,
		},
		{
			name:     "Window=0 stop-and-wait: normal ack sends exactly the next page",
			numPages: 5, window: 0,
			seedAckedPage: 1, seedSentPage: 2,
			pageIndex:     1,
			wantSent:      []int{2},
			wantAckedPage: 2,
			wantSentPage:  3,
		},
		{
			name:     "Window=0 stop-and-wait: mismatch fallback resends current page (64be9cbc)",
			numPages: 5, window: 0,
			seedAckedPage: 2, seedSentPage: 3,
			pageIndex:     1, // AckedPage-1 == 1
			wantSent:      []int{2},
			wantAckedPage: 2,
			wantSentPage:  3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sentLog []int
			withFakeSendSyncPage(t, &sentLog)

			ctx := "ctx-" + t.Name()
			entry := newTestDownloadEntry(ctx, tc.numPages, tc.window)
			entry.AckedPage = tc.seedAckedPage
			entry.SentPage = tc.seedSentPage
			syncDownloadStore(ctx, "note", entry)
			t.Cleanup(func() { syncDownloadDelete(ctx, "note") })

			handlePageAck(nil, entry, tc.pageIndex, "note", zap.NewNop(), "test-trace")

			if !intSliceEqual(sentLog, tc.wantSent) {
				t.Errorf("sent pages = %v, want %v", sentLog, tc.wantSent)
			}
			if entry.AckedPage != tc.wantAckedPage {
				t.Errorf("AckedPage = %d, want %d", entry.AckedPage, tc.wantAckedPage)
			}
			if entry.SentPage != tc.wantSentPage {
				t.Errorf("SentPage = %d, want %d", entry.SentPage, tc.wantSentPage)
			}
		})
	}
}

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
