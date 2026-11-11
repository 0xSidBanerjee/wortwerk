// Package split segments German paragraphs into sentences.
package split

import (
	"strings"
	"unicode"
)

var abbreviations = map[string]struct{}{
	"z.B.": {}, "Dr.": {}, "Mr.": {}, "Mrs.": {}, "Ms.": {},
}

// Sentences splits one paragraph into sentence strings (trimmed, with trailing punctuation).
func Sentences(paragraph string) []string {
	paragraph = strings.TrimSpace(paragraph)
	if paragraph == "" {
		return nil
	}
	var out []string
	runes := []rune(paragraph)
	start := 0
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r != '.' && r != '!' && r != '?' {
			continue
		}
		if isAbbreviationDot(runes, start, i) {
			continue
		}
		j := i + 1
		for j < len(runes) && unicode.IsSpace(runes[j]) {
			j++
		}
		if j < len(runes) && isLowerContinuation(runes, j) {
			continue
		}
		sent := strings.TrimSpace(string(runes[start : i+1]))
		if sent != "" {
			out = append(out, sent)
		}
		start = j
		i = j - 1
	}
	tail := strings.TrimSpace(string(runes[start:]))
	if tail != "" {
		out = append(out, tail)
	}
	return out
}

func isAbbreviationDot(runes []rune, start, dot int) bool {
	// Walk back to previous space for the token ending in dot.
	i := dot
	for i > start && runes[i] != ' ' {
		i--
	}
	token := string(runes[i : dot+1])
	_, ok := abbreviations[token]
	return ok
}

// Lowercase continuation after German closing quote: „Hallo!", sagte er.
func isLowerContinuation(runes []rune, j int) bool {
	if j >= len(runes) {
		return false
	}
	// Skip closing quote and punctuation before speech tag
	for j < len(runes) && (runes[j] == '"' || runes[j] == '\u201c' || runes[j] == '\u201d' || runes[j] == ',' || runes[j] == ' ') {
		j++
	}
	if j >= len(runes) {
		return false
	}
	return unicode.IsLower(runes[j])
}
