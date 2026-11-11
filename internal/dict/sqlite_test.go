package dict

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0xSidBanerjee/wortwerk/internal/ingest/candidates"
)

func TestBuildFile_and_lookup(t *testing.T) {
	root := filepath.Join("..", "..")
	jsonl := filepath.Join(root, "testdata", "dict", "sample.jsonl")
	dbPath := filepath.Join(t.TempDir(), "dict.sqlite")
	n, err := BuildFile(jsonl, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if n < 5 {
		t.Fatalf("imported %d rows", n)
	}
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cands, err := candidates.Lookup(t.Context(), db, "geht", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) == 0 {
		t.Fatal("expected candidates for geht")
	}
	_ = os.Getenv
}
