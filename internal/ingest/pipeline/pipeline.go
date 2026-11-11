// Package pipeline orchestrates story ingestion.
package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xSidBanerjee/wortwerk/internal/dict"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/compile"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/llm"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/parse"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/validate"
	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

// Options configures one ingest run.
type Options struct {
	Root           string
	Slug           string
	GlossLang      string
	Offline        bool
	AnnotationFile string // used when Offline
	LLM            *llm.Client
	DictPath       string // data/dict.sqlite
}

// Ingest compiles one story markdown file to content/compiled/<slug>.json.
func Ingest(ctx context.Context, opts Options) error {
	if opts.GlossLang == "" {
		opts.GlossLang = "en"
	}
	mdPath := filepath.Join(opts.Root, "content", "stories", opts.Slug+".md")
	md, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}
	doc, err := parse.Parse(string(md))
	if err != nil {
		return err
	}
	if doc.Meta.Slug != opts.Slug {
		return fmt.Errorf("frontmatter slug %q != %q", doc.Meta.Slug, opts.Slug)
	}

	var ann compile.Annotation
	if opts.Offline {
		path := opts.AnnotationFile
		if path == "" {
			path = filepath.Join(opts.Root, "testdata", "ingest", opts.Slug+"-annotation.json")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("offline annotation: %w", err)
		}
		if err := json.Unmarshal(data, &ann); err != nil {
			return err
		}
	} else {
		if opts.LLM == nil {
			return fmt.Errorf("LLM client required (or use --offline)")
		}
		// Live path: build chunk payload (candidates optional for first implementation).
		payload, err := buildChunkPayload(opts.Root, doc, opts.DictPath)
		if err != nil {
			return err
		}
		userJSON, _ := json.Marshal(payload)
		ann, _, err = opts.LLM.Annotate(ctx, string(userJSON))
		if err != nil {
			return err
		}
	}

	story, err := compile.Assemble(doc, ann, opts.GlossLang)
	if err != nil {
		return err
	}
	res := validate.Story(story)
	if !res.OK() {
		for _, e := range res.Errors() {
			fmt.Fprintln(os.Stderr, e)
		}
		return fmt.Errorf("validation failed")
	}
	out, err := compile.MarshalStory(story)
	if err != nil {
		return err
	}
	outPath := filepath.Join(opts.Root, "content", "compiled", opts.Slug+".json")
	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		return err
	}
	if err := updateIndex(opts.Root, doc, story, out); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d bytes)\n", outPath, len(out))
	return nil
}

func buildChunkPayload(root string, doc parse.Document, dictPath string) (map[string]any, error) {
	_ = root
	_ = dictPath
	return map[string]any{
		"title":      doc.Meta.Title,
		"level":      doc.Meta.Level,
		"paragraphs": doc.Paragraphs,
		"prompt":     llm.PromptVersion,
	}, nil
}

func updateIndex(root string, doc parse.Document, s *story.CompiledStory, compiled []byte) error {
	_ = s
	idxPath := filepath.Join(root, "content", "compiled", "index.json")
	var idx compile.IndexFile
	data, err := os.ReadFile(idxPath)
	if err == nil {
		_ = json.Unmarshal(data, &idx)
	}
	rev := shortHash(compiled)
	minutes := (s.WordCount + 99) / 100
	if minutes < 1 {
		minutes = 1
	}
	entry := compile.IndexEntry{
		Slug:      doc.Meta.Slug,
		Title:     doc.Meta.Title,
		Level:     doc.Meta.Level,
		WordCount: s.WordCount,
		Minutes:   minutes,
		Rev:       rev,
	}
	found := false
	for i, e := range idx.Stories {
		if e.Slug == doc.Meta.Slug {
			idx.Stories[i] = entry
			found = true
			break
		}
	}
	if !found {
		idx.Stories = append(idx.Stories, entry)
	}
	out, err := compile.MarshalIndex(idx.Stories)
	if err != nil {
		return err
	}
	return os.WriteFile(idxPath, out, 0o644)
}

func shortHash(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:4])
}

// BuildDict imports JSONL into data/dict.sqlite under root.
func BuildDict(root, jsonlPath string) (int, error) {
	sqlitePath := filepath.Join(root, "data", "dict.sqlite")
	if jsonlPath == "" {
		jsonlPath = filepath.Join(root, "testdata", "dict", "sample.jsonl")
	}
	if err := os.MkdirAll(filepath.Join(root, "data"), 0o755); err != nil {
		return 0, err
	}
	return dict.BuildFile(jsonlPath, sqlitePath)
}
