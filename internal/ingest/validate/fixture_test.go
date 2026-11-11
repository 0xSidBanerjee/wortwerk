package validate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

func TestFixtureCompiledJSON(t *testing.T) {
	path := filepath.Join("..", "..", "..", "content", "compiled", "fixture.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var compiled story.CompiledStory
	if err := json.Unmarshal(data, &compiled); err != nil {
		t.Fatalf("decode: %v", err)
	}

	res := Story(&compiled)
	for _, w := range res.Issues {
		if w.Level == "warn" {
			t.Logf("warn: %s: %s", w.Where, w.Message)
		}
	}
	if !res.OK() {
		for _, e := range res.Errors() {
			t.Error(e)
		}
		t.Fatalf("%d errors", len(res.Errors()))
	}
}
