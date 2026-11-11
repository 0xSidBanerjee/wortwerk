// Package parse reads story markdown sources.
package parse

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Meta is YAML frontmatter for a story file.
type Meta struct {
	Title  string
	Slug   string
	Level  string
	Lang   string
	Source string
}

// Document is a parsed story source ready for sentence splitting.
type Document struct {
	Meta       Meta
	Paragraphs []string
}

// Parse reads frontmatter and plain paragraph bodies.
func Parse(src string) (Document, error) {
	src = strings.TrimPrefix(src, "\ufeff")
	src = strings.TrimSpace(src)
	if !strings.HasPrefix(src, "---") {
		return Document{}, fmt.Errorf("missing frontmatter opener ---")
	}
	rest := src[3:]
	rest = strings.TrimLeft(rest, "\r\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return Document{}, fmt.Errorf("missing frontmatter closer ---")
	}
	fm := rest[:end]
	body := strings.TrimSpace(rest[end+4:])

	meta, err := parseFrontmatter(fm)
	if err != nil {
		return Document{}, err
	}
	if meta.Slug == "" || meta.Title == "" || meta.Lang == "" {
		return Document{}, fmt.Errorf("frontmatter requires title, slug, lang")
	}
	body = norm.NFC.String(body)
	paras, err := splitParagraphs(body)
	if err != nil {
		return Document{}, err
	}
	return Document{Meta: meta, Paragraphs: paras}, nil
}

func parseFrontmatter(fm string) (Meta, error) {
	var m Meta
	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		switch k {
		case "title":
			m.Title = v
		case "slug":
			m.Slug = v
		case "level":
			m.Level = v
		case "lang":
			m.Lang = v
		case "source":
			m.Source = v
		}
	}
	return m, nil
}

func splitParagraphs(body string) ([]string, error) {
	if body == "" {
		return nil, fmt.Errorf("story body is empty")
	}
	raw := strings.Split(body, "\n\n")
	var out []string
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if err := lintParagraph(p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no paragraphs found")
	}
	return out, nil
}

func lintParagraph(p string) error {
	for _, line := range strings.Split(p, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			return fmt.Errorf("headings not allowed in body")
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			return fmt.Errorf("lists not allowed in body")
		}
	}
	if norm.NFC.String(p) != p {
		return fmt.Errorf("paragraph must be Unicode NFC")
	}
	for _, abbr := range []string{"z.B.", "d.h.", "u.a.", "usw."} {
		if strings.Contains(p, abbr) {
			return fmt.Errorf("abbreviation %q not allowed in MVP stories", abbr)
		}
	}
	return nil
}

// WordCount counts word-like tokens in all paragraphs (for lint).
func WordCount(paragraphs []string) int {
	n := 0
	for _, p := range paragraphs {
		for _, f := range strings.Fields(p) {
			if hasLetterOrDigit(f) {
				n++
			}
		}
	}
	return n
}

func hasLetterOrDigit(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
