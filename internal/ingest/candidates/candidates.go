// Package candidates looks up dictionary senses for surface forms.
package candidates

import (
	"context"
	"database/sql"
	"fmt"
)

// Candidate is one sense option for the LLM.
type Candidate struct {
	Form       string `json:"form"`
	Lemma      string `json:"lemma"`
	POS        string `json:"pos"`
	Gloss      string `json:"gloss"`
	Gender     string `json:"gender,omitempty"`
	Plural     string `json:"plural,omitempty"`
	SepPrefix  string `json:"separable_prefix,omitempty"`
	SepStem    string `json:"separable_stem,omitempty"`
}

// Lookup returns up to limit dictionary rows for a form.
func Lookup(ctx context.Context, db *sql.DB, form string, limit int) ([]Candidate, error) {
	if limit <= 0 {
		limit = 3
	}
	rows, err := db.QueryContext(ctx,
		`SELECT form, lemma, pos, gloss, gender, plural, sep_prefix, sep_stem FROM forms WHERE form = ? LIMIT ?`,
		form, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Candidate
	for rows.Next() {
		var c Candidate
		var gender, plural, pre, stem sql.NullString
		if err := rows.Scan(&c.Form, &c.Lemma, &c.POS, &c.Gloss, &gender, &plural, &pre, &stem); err != nil {
			return nil, err
		}
		if gender.Valid {
			c.Gender = gender.String
		}
		if plural.Valid {
			c.Plural = plural.String
		}
		if pre.Valid {
			c.SepPrefix = pre.String
		}
		if stem.Valid {
			c.SepStem = stem.String
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// LookupMany returns candidates keyed by form for each distinct word token text in order.
func LookupMany(ctx context.Context, db *sql.DB, forms []string) (map[string][]Candidate, error) {
	out := make(map[string][]Candidate)
	for _, f := range forms {
		if f == "" {
			continue
		}
		if _, ok := out[f]; ok {
			continue
		}
		cands, err := Lookup(ctx, db, f, 3)
		if err != nil {
			return nil, fmt.Errorf("lookup %q: %w", f, err)
		}
		out[f] = cands
	}
	return out, nil
}
