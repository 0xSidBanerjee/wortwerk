package pipeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/0xSidBanerjee/wortwerk/internal/dict"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/candidates"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/compile"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/llm"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/parse"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/split"
	"github.com/0xSidBanerjee/wortwerk/internal/ingest/tokenize"
	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

type tokenizedSentence struct {
	ID          string                    `json:"id"`
	P           int                       `json:"p"`
	Text        string                    `json:"text"`
	Tokens      []story.Token             `json:"tokens"`
	Candidates  map[string][]candidates.Candidate `json:"candidates,omitempty"`
}

func tokenizeDocument(doc parse.Document) ([]tokenizedSentence, error) {
	var out []tokenizedSentence
	sid := 0
	for p, para := range doc.Paragraphs {
		for _, sent := range split.Sentences(para) {
			toks := tokenize.Sentence(sent)
			id := fmt.Sprintf("s%d", sid)
			out = append(out, tokenizedSentence{ID: id, P: p, Text: sent, Tokens: toks})
			sid++
		}
	}
	return out, nil
}

func attachCandidates(ctx context.Context, db *sql.DB, sents []tokenizedSentence) error {
	if db == nil {
		return nil
	}
	for i := range sents {
		forms := make([]string, 0, len(sents[i].Tokens))
		for _, t := range sents[i].Tokens {
			if isWordToken(t.Text) {
				forms = append(forms, t.Text)
			}
		}
		m, err := candidates.LookupMany(ctx, db, forms)
		if err != nil {
			return err
		}
		sents[i].Candidates = m
	}
	return nil
}

func isWordToken(text string) bool {
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func annotateLive(ctx context.Context, client *llm.Client, dictPath string, doc parse.Document) (compile.Annotation, llm.Usage, error) {
	sents, err := tokenizeDocument(doc)
	if err != nil {
		return compile.Annotation{}, llm.Usage{}, err
	}
	var db *sql.DB
	if _, err := os.Stat(dictPath); err == nil {
		var openErr error
		db, openErr = dict.Open(dictPath)
		if openErr == nil {
			defer db.Close()
			_ = attachCandidates(ctx, db, sents)
		}
	}
	const chunkSize = 10
	var merged compile.Annotation
	var usage llm.Usage
	for start := 0; start < len(sents); start += chunkSize {
		end := start + chunkSize
		if end > len(sents) {
			end = len(sents)
		}
		chunk := sents[start:end]
		payload := map[string]any{
			"prompt_version": llm.PromptVersion,
			"title":          doc.Meta.Title,
			"level":          doc.Meta.Level,
			"sentences":      chunk,
		}
		raw, _ := json.Marshal(payload)
		part, u, err := client.Annotate(ctx, string(raw))
		if err != nil {
			return compile.Annotation{}, usage, err
		}
		usage.InputTokens += u.InputTokens
		usage.OutputTokens += u.OutputTokens
		if merged.Lexicon == nil {
			merged.Lexicon = map[string]story.LexiconEntry{}
		}
		merged.Sentences = append(merged.Sentences, part.Sentences...)
		for k, v := range part.Lexicon {
			merged.Lexicon[k] = v
		}
	}
	return merged, usage, nil
}

func usageSummary(u llm.Usage) string {
	return fmt.Sprintf("LLM tokens in=%d out=%d", u.InputTokens, u.OutputTokens)
}

// autoDraftAnnotation builds a one-token-one-unit draft (ingest scaffold only).
func autoDraftAnnotation(sents []tokenizedSentence) compile.Annotation {
	ann := compile.Annotation{Lexicon: map[string]story.LexiconEntry{}}
	for _, s := range sents {
		var units []story.Unit
		for _, t := range s.Tokens {
			if !isWordToken(t.Text) {
				continue
			}
			lex := lexID(t.Text)
			units = append(units, story.Unit{Tokens: []int{t.ID}, Lex: lex})
			if _, ok := ann.Lexicon[lex]; !ok {
				ann.Lexicon[lex] = story.LexiconEntry{
					Lemma:   t.Text,
					Kind:    "word",
					POS:     "word",
					Meaning: strings.ToLower(t.Text),
				}
			}
		}
		ann.Sentences = append(ann.Sentences, compile.AnnotatedSentence{
			ID:          s.ID,
			Translation: s.Text,
			Units:       units,
		})
	}
	return ann
}

func lexID(surface string) string {
	return strings.ToLower(strings.ReplaceAll(surface, "ß", "ss"))
}
