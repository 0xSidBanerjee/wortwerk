package validate

import (
	"testing"

	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

func TestStory_joinTokensRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		tokens []story.Token
		want   string
	}{
		{
			name: "separable_with_period",
			tokens: []story.Token{
				{ID: 0, Text: "Pinocchio", Space: true},
				{ID: 1, Text: "steht", Space: true},
				{ID: 2, Text: "früh", Space: true},
				{ID: 3, Text: "auf", Space: false},
				{ID: 4, Text: ".", Space: true},
			},
			want: "Pinocchio steht früh auf.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := joinTokens(tc.tokens)
			if got != tc.want {
				t.Fatalf("joinTokens() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestStory_duplicateTokenInUnits(t *testing.T) {
	s := &story.CompiledStory{
		SchemaVersion: 1,
		WordCount:     1,
		Sentences: []story.Sentence{
			{
				ID: "s0",
				Tokens: []story.Token{{ID: 0, Text: "Hallo", Space: false}},
				Units: []story.Unit{
					{Tokens: []int{0}, Lex: "hallo"},
					{Tokens: []int{0}, Lex: "hallo"},
				},
				Translation: "Hello.",
			},
		},
		Lexicon: map[string]story.LexiconEntry{
			"hallo": {Lemma: "hallo", Kind: "word", POS: "interjection", Meaning: "hello"},
		},
	}
	res := Story(s)
	if res.OK() {
		t.Fatal("expected errors for duplicate token assignment")
	}
}
