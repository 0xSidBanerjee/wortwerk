package compile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xSidBanerjee/wortwerk/internal/ingest/parse"
	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

func TestGolden_fixture_offline_ingest(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	mdPath := filepath.Join(root, "content", "stories", "fixture.md")
	wantPath := filepath.Join(root, "content", "compiled", "fixture.json")
	annPath := filepath.Join(root, "testdata", "ingest", "fixture-annotation.json")

	md, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := parse.Parse(string(md))
	if err != nil {
		t.Fatal(err)
	}
	annData, err := os.ReadFile(annPath)
	if err != nil {
		t.Fatal(err)
	}
	var ann Annotation
	if err := json.Unmarshal(annData, &ann); err != nil {
		t.Fatal(err)
	}
	compiled, err := Assemble(doc, ann, "en")
	if err != nil {
		t.Fatal(err)
	}
	got, err := MarshalStory(compiled)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		// Normalize through decode to pinpoint drift (byte layout may vary slightly).
		var gotStory, wantStory story.CompiledStory
		if err := json.Unmarshal(got, &gotStory); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(want, &wantStory); err != nil {
			t.Fatal(err)
		}
		gotNorm, _ := json.Marshal(gotStory)
		wantNorm, _ := json.Marshal(wantStory)
		if string(gotNorm) != string(wantNorm) {
			t.Fatalf("compiled story differs from fixture.json")
		}
		// Accept semantic match; re-marshal committed artifact when format intentionally changes.
		t.Logf("note: textual JSON layout differs (%d vs %d bytes) but content matches", len(got), len(want))
	}
}
