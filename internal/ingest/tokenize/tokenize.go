// Package tokenize turns sentence text into tokens with space flags.
package tokenize

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

// Sentence tokenizes one sentence. Token ids restart at 0 per sentence.
func Sentence(text string) []story.Token {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	var tokens []story.Token
	id := 0
	for i := 0; i < len(text); {
		for i < len(text) && text[i] == ' ' {
			i++
		}
		if i >= len(text) {
			break
		}
		r, w := utf8.DecodeRuneInString(text[i:])
		var piece string
		switch {
		case r == '„' || r == '\u201e':
			piece = "„"
			i += w
		case r == '\u201c' || r == '\u201d' || r == '"':
			piece = string(r)
			i += w
		case isWordRune(r):
			j := i
			for j < len(text) {
				r2, w2 := utf8.DecodeRuneInString(text[j:])
				if !isWordRune(r2) {
					break
				}
				j += w2
			}
			piece = text[i:j]
			i = j
		default:
			piece = string(r)
			i += w
		}
		space := false
		if i < len(text) && text[i] == ' ' {
			k := i
			for k < len(text) && text[k] == ' ' {
				k++
			}
			if k < len(text) {
				nextR, _ := utf8.DecodeRuneInString(text[k:])
				if isWordRune(nextR) || nextR == '„' || nextR == '\u201e' {
					space = true
				}
			}
		}
		tokens = append(tokens, story.Token{ID: id, Text: piece, Space: space})
		id++
	}
	return tokens
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '\u00df' ||
		strings.ContainsRune("äöüÄÖÜ", r)
}
