// Package dict provides a local German form index for ingestion (tooling only).
package dict

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS forms (
  form TEXT NOT NULL,
  lemma TEXT NOT NULL,
  pos TEXT,
  gloss TEXT,
  gender TEXT,
  plural TEXT,
  sep_prefix TEXT,
  sep_stem TEXT
);
CREATE INDEX IF NOT EXISTS idx_forms_form ON forms(form);
`

// Open opens or creates a SQLite dictionary database.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// wiktextractLine is a minimal Wiktextract entry subset.
type wiktextractLine struct {
	Word      string `json:"word"`
	LangCode  string `json:"lang_code"`
	POS       string `json:"pos"`
	Senses    []struct {
		Glosses []string `json:"glosses"`
		Tags    []string `json:"tags"`
	} `json:"senses"`
	Forms []struct {
		Form string   `json:"form"`
		Tags []string `json:"tags"`
	} `json:"forms"`
}

// ImportJSONL loads German entries from a Wiktextract JSONL export.
func ImportJSONL(db *sql.DB, r io.Reader) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare(`INSERT INTO forms (form, lemma, pos, gloss, gender, plural, sep_prefix, sep_stem) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	n := 0
	for sc.Scan() {
		var row wiktextractLine
		if err := json.Unmarshal(sc.Bytes(), &row); err != nil {
			continue
		}
		if row.LangCode != "de" {
			continue
		}
		gloss := ""
		if len(row.Senses) > 0 && len(row.Senses[0].Glosses) > 0 {
			gloss = row.Senses[0].Glosses[0]
		}
		gender := ""
		plural := ""
		for _, s := range row.Senses {
			for _, t := range s.Tags {
				switch t {
				case "masculine":
					gender = "m"
				case "feminine":
					gender = "f"
				case "neuter":
					gender = "n"
				}
			}
		}
		for _, f := range row.Forms {
			for _, t := range f.Tags {
				if t == "plural" {
					plural = f.Form
				}
			}
		}
		if _, err := stmt.Exec(row.Word, row.Word, row.POS, gloss, gender, plural, "", ""); err != nil {
			return n, err
		}
		n++
		for _, f := range row.Forms {
			if f.Form == "" {
				continue
			}
			if _, err := stmt.Exec(f.Form, row.Word, row.POS, gloss, gender, plural, "", ""); err != nil {
				return n, err
			}
			n++
		}
	}
	if err := sc.Err(); err != nil {
		return n, err
	}
	return n, tx.Commit()
}

// BuildFile creates dict.sqlite from a JSONL path.
func BuildFile(jsonlPath, sqlitePath string) (int, error) {
	if err := os.MkdirAll(filepath.Dir(sqlitePath), 0o755); err != nil {
		return 0, err
	}
	_ = os.Remove(sqlitePath)
	db, err := Open(sqlitePath)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	f, err := os.Open(jsonlPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n, err := ImportJSONL(db, f)
	if err != nil {
		return n, err
	}
	if n == 0 {
		return 0, fmt.Errorf("no rows imported from %s", jsonlPath)
	}
	return n, nil
}
