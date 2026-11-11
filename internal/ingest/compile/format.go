package compile

import (
	"encoding/json"
	"sort"

	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

// MarshalStory writes deterministic JSON for git diffs (sorted lexicon keys, trailing newline).
func MarshalStory(s *story.CompiledStory) ([]byte, error) {
	lex := make(map[string]story.LexiconEntry, len(s.Lexicon))
	keys := make([]string, 0, len(s.Lexicon))
	for k := range s.Lexicon {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		lex[k] = s.Lexicon[k]
	}
	c := *s
	c.Lexicon = lex
	b, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// IndexEntry is one row in content/compiled/index.json.
type IndexEntry struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Level     string `json:"level"`
	WordCount int    `json:"wordCount"`
	Minutes   int    `json:"minutes"`
	Rev       string `json:"rev"`
}

// IndexFile model.
type IndexFile struct {
	Stories []IndexEntry `json:"stories"`
}

// MarshalIndex writes index.json with sorted slugs.
func MarshalIndex(entries []IndexEntry) ([]byte, error) {
	sort.Slice(entries, func(i, j int) bool { return entries[i].Slug < entries[j].Slug })
	file := IndexFile{Stories: entries}
	b, err := json.MarshalIndent(&file, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
