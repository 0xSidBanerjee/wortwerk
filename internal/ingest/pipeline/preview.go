// Package pipeline — preview helpers for story authoring.
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xSidBanerjee/wortwerk/internal/ingest/parse"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/split"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/tokenize"
	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

// PreviewTokens prints tokenized sentences for a story slug (authoring aid).
func PreviewTokens(root, slug string) error {
	mdPath := filepath.Join(root, "content", "stories", slug+".md")
	md, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}
	doc, err := parse.Parse(string(md))
	if err != nil {
		return err
	}
	type row struct {
		ID     string        `json:"id"`
		P      int           `json:"p"`
		Text   string        `json:"text"`
		Tokens []story.Token `json:"tokens"`
	}
	var rows []row
	sid := 0
	for p, para := range doc.Paragraphs {
		for _, sent := range split.Sentences(para) {
			toks := tokenize.Sentence(sent)
			id := fmt.Sprintf("s%d", sid)
			rows = append(rows, row{ID: id, P: p, Text: sent, Tokens: toks})
			sid++
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}
