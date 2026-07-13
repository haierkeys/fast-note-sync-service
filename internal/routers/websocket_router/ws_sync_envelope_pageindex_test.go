package websocket_router

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	v1 "github.com/haierkeys/fast-note-sync-service/internal/proto/v1"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"google.golang.org/protobuf/proto"
)

// These tests pin the 1-based envelope pageIndex wire semantics: internal page 0 must go on the
// wire as 1, because WSResponse.pageIndex is a non-optional proto3 int32 — a 0-based page 0
// would be elided by pb encoding and decode as 0, indistinguishable from non-paginated messages
// (which also decode as 0), breaking the client's "pageIndex === undefined → legacy path"
// routing for page 0. Wire semantics: 0/absent = non-paginated, n>0 = internal page n-1.
// Internal SentPage/AckedPage and the client→server PageAck.pageIndex stay 0-based.

// TestEnvelopePageIndex_IsOneBased pins the mapping helper itself: internal page 0 → wire 1.
func TestEnvelopePageIndex_IsOneBased(t *testing.T) {
	for _, tc := range []struct {
		internalPage int
		wantWire     int
	}{
		{0, 1},
		{1, 2},
		{732, 733},
	} {
		c := envelopePageIndex(code.Success, tc.internalPage)
		if !c.HavePageIndex() {
			t.Fatalf("internal page %d: HavePageIndex() = false, want true", tc.internalPage)
		}
		if got := c.PageIndex(); got != tc.wantWire {
			t.Errorf("internal page %d: wire pageIndex = %d, want %d", tc.internalPage, got, tc.wantWire)
		}
	}
}

// buildEnvelopeRes mirrors WebsocketClient.ToResponse's Code→Res mapping for the fields the
// envelope tests care about (Data/Vault/Context/PageIndex). ToResponse itself needs a live
// gws.Conn to call, so the mapping is reproduced here; it's the same four HaveX() guards.
func buildEnvelopeRes(codeObj *code.Code) *pkgapp.Res {
	content := &pkgapp.Res{
		Code:   codeObj.Code(),
		Status: codeObj.Status(),
		Data:   codeObj.Data(),
	}
	if codeObj.HaveVault() {
		content.Vault = codeObj.Vault()
	}
	if codeObj.HaveContext() {
		content.Context = codeObj.Context()
	}
	if codeObj.HavePageIndex() {
		content.PageIndex = codeObj.PageIndex()
	}
	return content
}

// TestEnvelopePageIndex_Page0_JSONWireValue asserts the JSON encoding of an internal page 0
// detail envelope carries "pageIndex":1 explicitly.
func TestEnvelopePageIndex_Page0_JSONWireValue(t *testing.T) {
	codeObj := envelopePageIndex(code.Success.WithData(dto.SyncPageMessage{
		PageIndex:  0, // payload-level page index stays 0-based (unchanged legacy field)
		PageSize:   200,
		TotalCount: 200,
		IsLast:     false,
	}), 0).WithVault("v").WithContext("ctx")

	raw, err := json.Marshal(buildEnvelopeRes(codeObj))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"pageIndex":1`) {
		t.Fatalf("JSON wire for internal page 0 = %s, want it to contain \"pageIndex\":1", raw)
	}
}

// TestEnvelopePageIndex_Page0_ProtobufWireValue asserts the pb encoding of an internal page 0
// detail envelope decodes with PageIndex==1 — and that a non-paginated message decodes with
// PageIndex==0, i.e. the two are distinguishable on the pv>=2 pb main path.
func TestEnvelopePageIndex_Page0_ProtobufWireValue(t *testing.T) {
	decodePageIndex := func(t *testing.T, res *pkgapp.Res) int32 {
		t.Helper()
		raw, err := EnSendDTOToProtobuf(NoteSyncPage, res)
		if err != nil {
			t.Fatal(err)
		}
		if len(raw) < 2 || raw[0] != 'p' || raw[1] != 'b' {
			t.Fatalf("pb frame missing 'pb' prefix: % x", raw[:2])
		}
		var env v1.WSMessage
		if err := proto.Unmarshal(raw[2:], &env); err != nil {
			t.Fatal(err)
		}
		var wsResp v1.WSResponse
		if err := proto.Unmarshal(env.Data, &wsResp); err != nil {
			t.Fatal(err)
		}
		return wsResp.GetPageIndex()
	}

	pageMsg := dto.SyncPageMessage{PageIndex: 0, PageSize: 200, TotalCount: 200, IsLast: false}

	// Internal page 0, sent through the envelope mapping → wire must decode as 1.
	page0 := buildEnvelopeRes(envelopePageIndex(code.Success.WithData(pageMsg), 0).WithVault("v").WithContext("ctx"))
	if got := decodePageIndex(t, page0); got != 1 {
		t.Fatalf("pb wire for internal page 0 decoded PageIndex = %d, want 1", got)
	}

	// Non-paginated message (no WithPageIndex) → wire must decode as 0, proving page 0 and
	// non-paginated frames are distinguishable under pb zero-value elision.
	nonPaginated := buildEnvelopeRes(code.Success.WithData(pageMsg).WithVault("v").WithContext("ctx"))
	if got := decodePageIndex(t, nonPaginated); got != 0 {
		t.Fatalf("pb wire for non-paginated message decoded PageIndex = %d, want 0", got)
	}
}
