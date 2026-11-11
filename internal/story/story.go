// Package story defines the compiled story JSON format (schema v1).
package story

// CompiledStory is the on-disk format served to the web app.
type CompiledStory struct {
	SchemaVersion int                      `json:"schemaVersion"`
	ID            string                   `json:"id"`
	Title         string                   `json:"title"`
	Lang          string                   `json:"lang"`
	GlossLang     string                   `json:"glossLang"`
	Level         string                   `json:"level"`
	WordCount     int                      `json:"wordCount"`
	Sentences     []Sentence               `json:"sentences"`
	Lexicon       map[string]LexiconEntry  `json:"lexicon"`
}

// Sentence is one sentence with tokens, units, and an English gloss.
type Sentence struct {
	ID          string `json:"id"`
	P           int    `json:"p"`
	Tokens      []Token `json:"tokens"`
	Units       []Unit  `json:"units"`
	Translation string `json:"translation"`
}

// Token is a text span; Space means whitespace follows this token when rendering.
type Token struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	Space bool   `json:"space"`
}

// Unit groups token ids that share one tap target (word, phrase, separable verb, etc.).
type Unit struct {
	Tokens []int  `json:"tokens"`
	Lex    string `json:"lex"`
	Form   string `json:"form,omitempty"`
	Gloss  string `json:"gloss,omitempty"`
}

// LexiconEntry is metadata for one lex id used in this story.
type LexiconEntry struct {
	Lemma     string     `json:"lemma"`
	Kind      string     `json:"kind"`
	POS       string     `json:"pos"`
	Meaning   string     `json:"meaning"`
	Gender    string     `json:"gender,omitempty"`
	Article   string     `json:"article,omitempty"`
	Plural    *string    `json:"plural,omitempty"`
	Separable *Separable `json:"separable,omitempty"`
	Parts     []string   `json:"parts,omitempty"`
	Audio     string     `json:"audio,omitempty"`
}

// Separable holds prefix and stem for separable verbs.
type Separable struct {
	Prefix string `json:"prefix"`
	Stem   string `json:"stem"`
}
