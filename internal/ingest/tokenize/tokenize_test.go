package tokenize

import (
	"strings"
	"testing"

	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

func join(tokens []story.Token) string {
	var b strings.Builder
	for _, tok := range tokens {
		b.WriteString(tok.Text)
		if tok.Space {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

func TestSentence_roundTrip(t *testing.T) {
	tests := []string{
		"Pinocchio steht früh auf.",
		"Er geht zum Markt.",
		"Der Handschuh liegt dort.",
		"„Hallo!“, sagte er.",
	}
	for _, sent := range tests {
		tok := Sentence(sent)
		if join(tok) != sent {
			t.Errorf("round-trip %q got %q", sent, join(tok))
		}
	}
}

func TestSentence_fixture_s0(t *testing.T) {
	want := []story.Token{
		{ID: 0, Text: "Pinocchio", Space: true},
		{ID: 1, Text: "steht", Space: true},
		{ID: 2, Text: "früh", Space: true},
		{ID: 3, Text: "auf", Space: false},
		{ID: 4, Text: ".", Space: false},
	}
	got := Sentence("Pinocchio steht früh auf.")
	if len(got) != len(want) {
		t.Fatalf("len %d want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] got %+v want %+v", i, got[i], want[i])
		}
	}
}
