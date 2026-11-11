// Package validate checks compiled stories without I/O beyond the caller.
package validate

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

const maxMeaningLen = 60
const longSentenceTokens = 14

// Issue is one validation finding.
type Issue struct {
	Level   string // "error" or "warn"
	Where   string
	Message string
}

// Result holds errors and warnings.
type Result struct {
	Issues []Issue
}

// OK reports whether there are no errors (warnings are OK).
func (r Result) OK() bool {
	for _, i := range r.Issues {
		if i.Level == "error" {
			return false
		}
	}
	return true
}

// Errors returns error-level messages only.
func (r Result) Errors() []string {
	var out []string
	for _, i := range r.Issues {
		if i.Level == "error" {
			out = append(out, fmt.Sprintf("%s: %s", i.Where, i.Message))
		}
	}
	return out
}

// Story validates a compiled story per project plan §6.1 (dictionary checks deferred).
func Story(s *story.CompiledStory) Result {
	var r Result

	if s == nil {
		r.Issues = append(r.Issues, Issue{Level: "error", Where: "story", Message: "nil story"})
		return r
	}
	if s.SchemaVersion != 1 {
		r.Issues = append(r.Issues, Issue{
			Level: "error", Where: "story", Message: fmt.Sprintf("unsupported schemaVersion %d", s.SchemaVersion),
		})
	}

	lexUsed := map[string]struct{}{}
	wordTokens := 0

	for _, sent := range s.Sentences {
		where := "sentence " + sent.ID
		surface := joinTokens(sent.Tokens)
		if surface == "" {
			r.Issues = append(r.Issues, Issue{Level: "error", Where: where, Message: "empty sentence surface"})
		}
		if surface != "" && !isNFC(surface) {
			r.Issues = append(r.Issues, Issue{Level: "warn", Where: where, Message: "text is not Unicode NFC"})
		}
		if strings.TrimSpace(sent.Translation) == "" {
			r.Issues = append(r.Issues, Issue{Level: "error", Where: where, Message: "empty sentence translation"})
		}

		tokenByID := map[int]story.Token{}
		for _, t := range sent.Tokens {
			tokenByID[t.ID] = t
		}

		assigned := map[int]int{}
		for ui, u := range sent.Units {
			uWhere := fmt.Sprintf("%s unit[%d]", where, ui)
			if len(u.Tokens) == 0 {
				r.Issues = append(r.Issues, Issue{Level: "error", Where: uWhere, Message: "unit has no tokens"})
				continue
			}
			if !sortedUnique(u.Tokens) {
				r.Issues = append(r.Issues, Issue{Level: "error", Where: uWhere, Message: "unit token ids must be unique and sorted"})
			}
			for _, tid := range u.Tokens {
				if _, ok := tokenByID[tid]; !ok {
					r.Issues = append(r.Issues, Issue{Level: "error", Where: uWhere, Message: fmt.Sprintf("invalid token id %d", tid)})
					continue
				}
				if prev, ok := assigned[tid]; ok {
					r.Issues = append(r.Issues, Issue{
						Level: "error", Where: uWhere,
						Message: fmt.Sprintf("token %d already in unit[%d]", tid, prev),
					})
				}
				assigned[tid] = ui
			}
			if u.Lex == "" {
				r.Issues = append(r.Issues, Issue{Level: "error", Where: uWhere, Message: "missing lex"})
				continue
			}
			lexUsed[u.Lex] = struct{}{}
			entry, ok := s.Lexicon[u.Lex]
			if !ok {
				r.Issues = append(r.Issues, Issue{Level: "error", Where: uWhere, Message: "unknown lex id " + u.Lex})
				continue
			}
			if strings.TrimSpace(entry.Meaning) == "" {
				r.Issues = append(r.Issues, Issue{Level: "error", Where: uWhere, Message: "empty meaning for lex " + u.Lex})
			} else if len([]rune(entry.Meaning)) > maxMeaningLen {
				r.Issues = append(r.Issues, Issue{
					Level: "error", Where: uWhere,
					Message: fmt.Sprintf("meaning too long (%d > %d)", len([]rune(entry.Meaning)), maxMeaningLen),
				})
			}
		}

		var wordsInSentence int
		for _, t := range sent.Tokens {
			if isWordToken(t.Text) {
				wordsInSentence++
				wordTokens++
				if _, ok := assigned[t.ID]; !ok {
					r.Issues = append(r.Issues, Issue{
						Level: "error", Where: where,
						Message: fmt.Sprintf("word token %d %q not in any unit", t.ID, t.Text),
					})
				}
			}
		}
		if wordsInSentence > longSentenceTokens {
			r.Issues = append(r.Issues, Issue{
				Level: "warn", Where: where,
				Message: fmt.Sprintf("sentence has %d word tokens (> %d)", wordsInSentence, longSentenceTokens),
			})
		}
	}

	for lexID, entry := range s.Lexicon {
		lWhere := "lexicon " + lexID
		if strings.TrimSpace(entry.Meaning) == "" {
			r.Issues = append(r.Issues, Issue{Level: "error", Where: lWhere, Message: "empty meaning"})
		}
		if entry.POS == "noun" && entry.Kind != "name" {
			if entry.Gender == "" {
				r.Issues = append(r.Issues, Issue{Level: "error", Where: lWhere, Message: "noun missing gender"})
			}
			if entry.Article == "" {
				r.Issues = append(r.Issues, Issue{Level: "error", Where: lWhere, Message: "noun missing article"})
			}
			if entry.Plural == nil {
				r.Issues = append(r.Issues, Issue{Level: "error", Where: lWhere, Message: "noun missing plural (use null if none)"})
			}
		}
		if entry.Kind == "separable" {
			if entry.Separable == nil {
				r.Issues = append(r.Issues, Issue{Level: "error", Where: lWhere, Message: "separable verb missing separable metadata"})
			} else {
				combined := entry.Separable.Prefix + entry.Separable.Stem
				if combined != entry.Lemma && !strings.EqualFold(combined, entry.Lemma) {
					r.Issues = append(r.Issues, Issue{
						Level: "error", Where: lWhere,
						Message: fmt.Sprintf("separable prefix+stem %q != lemma %q", combined, entry.Lemma),
					})
				}
			}
		}
		if entry.Kind == "compound" && len(entry.Parts) > 0 {
			if !compoundPartsMatchLemma(entry.Parts, entry.Lemma) {
				r.Issues = append(r.Issues, Issue{
					Level: "warn", Where: lWhere,
					Message: "compound parts do not concatenate to lemma",
				})
			}
		}
		if _, used := lexUsed[lexID]; !used {
			r.Issues = append(r.Issues, Issue{Level: "error", Where: lWhere, Message: "orphan lexicon entry"})
		}
	}

	if s.WordCount != wordTokens {
		r.Issues = append(r.Issues, Issue{
			Level: "error", Where: "story",
			Message: fmt.Sprintf("wordCount %d does not match counted word tokens %d", s.WordCount, wordTokens),
		})
	}

	return r
}

func joinTokens(tokens []story.Token) string {
	var b strings.Builder
	for _, t := range tokens {
		b.WriteString(t.Text)
		if t.Space {
			b.WriteByte(' ')
		}
	}
	return strings.TrimRight(b.String(), " ")
}

func isWordToken(text string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func sortedUnique(ids []int) bool {
	if len(ids) == 0 {
		return false
	}
	prev := ids[0]
	for i := 1; i < len(ids); i++ {
		if ids[i] <= prev {
			return false
		}
		prev = ids[i]
	}
	return true
}

func compoundPartsMatchLemma(parts []string, lemma string) bool {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	return strings.EqualFold(b.String(), lemma)
}

func isNFC(s string) bool {
	return norm.NFC.String(s) == s
}
