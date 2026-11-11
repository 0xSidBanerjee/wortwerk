package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xSidBanerjee/wortwerk/internal/ingest/parse"
)

// WriteDraftAnnotation writes a scaffold annotation JSON for manual or LLM refinement.
func WriteDraftAnnotation(root, slug string) error {
	mdPath := filepath.Join(root, "content", "stories", slug+".md")
	md, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}
	doc, err := parse.Parse(string(md))
	if err != nil {
		return err
	}
	sents, err := tokenizeDocument(doc)
	if err != nil {
		return err
	}
	ann := autoDraftAnnotation(sents)
	outPath := filepath.Join(root, "testdata", "ingest", slug+"-annotation.json")
	b, err := json.MarshalIndent(ann, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, append(b, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s (review and refine before ingest --offline)\n", outPath)
	return nil
}
