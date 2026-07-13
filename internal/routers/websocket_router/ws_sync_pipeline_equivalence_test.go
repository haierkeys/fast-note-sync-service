package websocket_router

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

// This file covers S5's core acceptance claim (design §7.1 S5 / §8 risk table): windowed
// pipelining and stop-and-wait must be equivalent at the "what gets collected / what gets
// delivered" level — only the RTT shape changes, not the data. Scale is "数百条量级" per the
// work order (a full 10k-note mock is out of proportion to server-side unit tests; the
// equivalence property being asserted here does not get stronger with more data, only slower).

// ---- Upload side: batch collection must be arrival-order independent ----

// TestUploadBatchCollection_ItemSetEquivalent_OrderIndependent exercises the real production
// primitives added/touched in S3 (syncBatchGetOrCreate, markBatchReceived) under two different
// batch arrival orders — strictly sequential (models stop-and-wait / W_up=0) and shuffled
// (models W_up=8 window pipelining, where network reordering lets later batches arrive before
// earlier ones). doNoteSync/doFileSync/etc. only ever consume the final collected entry.Items,
// so if the collected item SET is identical regardless of arrival order, the doSync input is
// guaranteed identical between the two upload modes — this is what the design's "doSync 输入
// 集合...逐元素相等" acceptance line is actually gated on.
func TestUploadBatchCollection_ItemSetEquivalent_OrderIndependent(t *testing.T) {
	const totalBatches = 130 // 数百批量级
	const itemsPerBatch = 7

	buildBatches := func() [][]string {
		batches := make([][]string, totalBatches)
		for i := 0; i < totalBatches; i++ {
			items := make([]string, itemsPerBatch)
			for j := range items {
				items[j] = fmt.Sprintf("batch-%d-item-%d", i, j)
			}
			batches[i] = items
		}
		return batches
	}

	collect := func(name string, order []int) map[string]struct{} {
		ctx := "ctx-upload-" + t.Name() + "-" + name
		t.Cleanup(func() { syncBatchDelete(ctx, "note") })

		batches := buildBatches()
		var entry *syncBatchEntry
		for _, idx := range order {
			e, _ := syncBatchGetOrCreate(ctx, "note", totalBatches)
			entry = e
			e.mu.Lock()
			if !e.markBatchReceived(idx) {
				for _, it := range batches[idx] {
					e.Items = append(e.Items, it)
				}
				e.ReceivedCount++
			}
			e.mu.Unlock()
		}

		if entry.ReceivedCount != totalBatches {
			t.Fatalf("[%s] ReceivedCount = %d, want %d", name, entry.ReceivedCount, totalBatches)
		}

		got := make(map[string]struct{}, len(entry.Items))
		for _, it := range entry.Items {
			got[it.(string)] = struct{}{}
		}
		return got
	}

	sequential := make([]int, totalBatches)
	for i := range sequential {
		sequential[i] = i
	}

	// Models W_up=8 out-of-order arrival: shuffle deterministically for a reproducible test.
	shuffled := append([]int(nil), sequential...)
	rand.New(rand.NewSource(7)).Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	setSequential := collect("sequential", sequential)
	setShuffled := collect("shuffled", shuffled)

	wantCount := totalBatches * itemsPerBatch
	if len(setSequential) != wantCount {
		t.Fatalf("sequential arrival: collected %d items, want %d", len(setSequential), wantCount)
	}
	if !reflect.DeepEqual(setSequential, setShuffled) {
		t.Fatal("collected item set differs between sequential (stop-and-wait) and shuffled (windowed) batch arrival order")
	}
}

// ---- Download side: page delivery set must be delivery-order independent ----

// simulateDownloadDelivery drives the real pump/handlePageAck state machine (via the
// sendSyncPageFunc seam from S4) through a full download round for the given window size,
// simulating a well-behaved client: pages complete in a randomized order (as real concurrent
// writes to disk would), and the client acks the highest page for which all pages up to and
// including it are locally complete — exactly the contract C3 imposes on the real client
// (design §4.3 "发 ack(n) 前必须确认页 ≤n 均已... 完成"). Returns the full sequence of page
// indices the server sent (each element is a page number, may legitimately repeat only if a
// rewind/retransmit branch fired — this simulation never triggers rewind, so no repeats are
// expected on the happy path this test covers).
func simulateDownloadDelivery(t *testing.T, numPages, window int, seed int64) []int {
	t.Helper()

	var sentLog []int
	withFakeSendSyncPage(t, &sentLog)

	ctx := fmt.Sprintf("ctx-download-%s-w%d-seed%d", t.Name(), window, seed)
	entry := newTestDownloadEntry(ctx, numPages, window)
	syncDownloadStore(ctx, "note", entry)
	log := zap.NewNop()

	handlePageAck(nil, entry, -1, "note", log, "t") // first pull

	completed := make(map[int]bool, numPages)
	rng := rand.New(rand.NewSource(seed))

	for i := 0; i < numPages*4; i++ { // generous iteration cap, real loop converges in <=numPages
		e, ok := syncDownloadGet(ctx, "note")
		if !ok {
			return sentLog // isLast delivered, entry destroyed: round finished
		}

		e.mu.Lock()
		sent, acked := e.SentPage, e.AckedPage
		e.mu.Unlock()

		var inFlight []int
		for p := acked; p < sent; p++ {
			if !completed[p] {
				inFlight = append(inFlight, p)
			}
		}
		if len(inFlight) == 0 {
			// Nothing new to complete yet; pump has already sent everything currently in the
			// window. Since we always ack immediately after a completion advances the
			// watermark, this only happens if the round is already done (handled by the !ok
			// branch above) — treat as a stall guard.
			t.Fatalf("simulation stalled: sent=%d acked=%d completed=%d entry still present", sent, acked, len(completed))
		}

		p := inFlight[rng.Intn(len(inFlight))]
		completed[p] = true

		highest := acked - 1
		for completed[highest+1] {
			highest++
		}
		if highest >= acked {
			handlePageAck(nil, entry, highest, "note", log, "t")
		}
	}

	t.Fatalf("simulation did not converge within iteration cap for window=%d", window)
	return nil
}

// TestDownloadDelivery_SetEquivalence_WindowVsStopAndWait covers the design's central download
// equivalence claim (§4.2 "Window=0 时...与 3.5.x 前逐页行为逐消息等价" + §7.1 S5 "下发明细集合
// 一致"): regardless of window size or the order in which the client's per-page completions (and
// therefore acks) happen, every page must be delivered exactly once, and the delivered page set
// must be identical between stop-and-wait (Window=0) and windowed (Window=4) delivery.
func TestDownloadDelivery_SetEquivalence_WindowVsStopAndWait(t *testing.T) {
	const numPages = 733 // 数百页量级

	stopAndWait := simulateDownloadDelivery(t, numPages, 0, 1)
	windowed := simulateDownloadDelivery(t, numPages, 4, 2)

	wantSet := make(map[int]struct{}, numPages)
	for i := 0; i < numPages; i++ {
		wantSet[i] = struct{}{}
	}

	toSet := func(log []int) map[int]struct{} {
		s := make(map[int]struct{}, len(log))
		for _, p := range log {
			s[p] = struct{}{}
		}
		return s
	}

	if len(stopAndWait) != numPages {
		t.Fatalf("stop-and-wait: sent %d page messages, want exactly %d (no dupes, no gaps)", len(stopAndWait), numPages)
	}
	if len(windowed) != numPages {
		t.Fatalf("windowed: sent %d page messages, want exactly %d (no dupes, no gaps)", len(windowed), numPages)
	}
	if !reflect.DeepEqual(toSet(stopAndWait), wantSet) {
		t.Fatal("stop-and-wait delivered page set != full page range")
	}
	if !reflect.DeepEqual(toSet(windowed), wantSet) {
		t.Fatal("windowed delivered page set != full page range")
	}
}

// TestDownloadDelivery_MultipleWindowSizes_AllConverge is a lighter-weight sweep across window
// sizes (including the clamp ceiling from S1's PipelineWindowDownClamped, 16) confirming the
// state machine always fully drains regardless of W_down, per §7.1 S1's "上限钳制 ≤32/≤16"
// bound.
func TestDownloadDelivery_MultipleWindowSizes_AllConverge(t *testing.T) {
	const numPages = 257
	wantSet := make(map[int]struct{}, numPages)
	for i := 0; i < numPages; i++ {
		wantSet[i] = struct{}{}
	}

	for _, window := range []int{0, 1, 2, 4, 8, 16} {
		window := window
		t.Run(fmt.Sprintf("window=%d", window), func(t *testing.T) {
			got := simulateDownloadDelivery(t, numPages, window, int64(window)+100)
			if len(got) != numPages {
				t.Fatalf("sent %d page messages, want %d", len(got), numPages)
			}
			gotSet := make(map[int]struct{}, len(got))
			for _, p := range got {
				gotSet[p] = struct{}{}
			}
			if !reflect.DeepEqual(gotSet, wantSet) {
				t.Fatal("delivered page set != full page range")
			}
		})
	}
}
