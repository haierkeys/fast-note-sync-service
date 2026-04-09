package service

import (
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
)

func TestExtractSharedNoteFileRefs(t *testing.T) {
	content := `
![[assets/photo.png|240]]
![inline](../images/demo.jpg "title")
<img src="./img/html.png" alt="demo">
`

	refs := extractSharedNoteFileRefs(content)
	expected := map[string]struct{}{
		"assets/photo.png":   {},
		"../images/demo.jpg": {},
		"./img/html.png":     {},
	}

	if len(refs) != len(expected) {
		t.Fatalf("unexpected refs count: got %d want %d", len(refs), len(expected))
	}

	for _, ref := range refs {
		if _, ok := expected[ref]; !ok {
			t.Fatalf("unexpected ref: %s", ref)
		}
	}
}

func TestBuildSharePathCandidates(t *testing.T) {
	candidates := buildSharePathCandidates("notes/daily/today.md", "../images/demo.png")
	expected := []string{"notes/images/demo.png"}

	if len(candidates) != len(expected) {
		t.Fatalf("unexpected candidates count: got %d want %d", len(candidates), len(expected))
	}

	for i := range expected {
		if candidates[i] != expected[i] {
			t.Fatalf("candidate[%d]: got %s want %s", i, candidates[i], expected[i])
		}
	}
}

func TestRewriteMarkdownImageLinks(t *testing.T) {
	content := `![demo](./images/demo.png "title")`
	fileRefs := map[string]*domain.File{
		"./images/demo.png": {ID: 42},
	}

	rewritten := rewriteMarkdownImageLinks(content, fileRefs, "share-token", "pwd")
	expected := `![demo](/api/share/file?id=42&share_token=share-token&password=pwd "title")`

	if rewritten != expected {
		t.Fatalf("unexpected rewritten content: got %s want %s", rewritten, expected)
	}
}
