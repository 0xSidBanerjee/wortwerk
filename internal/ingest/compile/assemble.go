// Package compile builds compiled stories from sources and LLM annotations.
package compile

import (
	"fmt"
	"unicode"

	"github.com/0xSidBanerjee/wortwerk/internal/ingest/parse"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/split"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/tokenize"
	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

// Annotation is LLM output for one ingest run (never includes tokens).
type Annotation struct {
	Sentences []AnnotatedSentence        `json:"sentences"`
	Lexicon   map[string]story.LexiconEntry `json:"lexicon"`
}

// AnnotatedSentence holds units and translation keyed by sentence id.
type AnnotatedSentence struct {
	ID          string       `json:"id"`
	Translation string       `json:"translation"`
	Units       []story.Unit `json:"units"`
}

// BuildSentences splits, tokenizes, and attaches LLM units.
func BuildSentences(doc parse.Document, ann Annotation) ([]story.Sentence, error) {
	annByID := map[string]AnnotatedSentence{}
	for _, s := range ann.Sentences {
		annByID[s.ID] = s
	}

	var out []story.Sentence
	sid := 0
	for p, para := range doc.Paragraphs {
		sents := split.Sentences(para)
		for _, sentText := range sents {
			id := fmt.Sprintf("s%d", sid)
			tokens := tokenize.Sentence(sentText)
			if len(tokens) == 0 {
				return nil, fmt.Errorf("sentence %q tokenized to nothing", sentText)
			}
			patchTrailingSpace(tokens)

			as, ok := annByID[id]
			if !ok {
				return nil, fmt.Errorf("missing annotation for %s", id)
			}
			out = append(out, story.Sentence{
				ID:          id,
				P:           p,
				Tokens:      tokens,
				Units:       as.Units,
				Translation: as.Translation,
			})
			sid++
		}
	}
	return out, nil
}

func patchTrailingSpace(tokens []story.Token) {
	if len(tokens) == 0 {
		return
	}
	last := &tokens[len(tokens)-1]
	if last.Text == "." || last.Text == "!" || last.Text == "?" {
		last.Space = true
	}
}

// Assemble builds a compiled story.
func Assemble(doc parse.Document, ann Annotation, glossLang string) (*story.CompiledStory, error) {
	sentences, err := BuildSentences(doc, ann)
	if err != nil {
		return nil, err
	}
	wc := 0
	for _, s := range sentences {
		for _, t := range s.Tokens {
			if isWord(t.Text) {
				wc++
			}
		}
	}
	return &story.CompiledStory{
		SchemaVersion: 1,
		ID:            doc.Meta.Slug,
		Title:         doc.Meta.Title,
		Lang:          doc.Meta.Lang,
		GlossLang:     glossLang,
		Level:         doc.Meta.Level,
		WordCount:     wc,
		Sentences:     sentences,
		Lexicon:       ann.Lexicon,
	}, nil
}

func isWord(text string) bool {
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
