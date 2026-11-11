package split

import "testing"

func TestSentences_table(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "fixture_paragraph",
			in:   "Pinocchio steht früh auf. Er geht zum Markt. Der Handschuh liegt dort.",
			want: []string{
				"Pinocchio steht früh auf.",
				"Er geht zum Markt.",
				"Der Handschuh liegt dort.",
			},
		},
		{
			name: "speech_continuation",
			in:   `„Hallo!", sagte er.`,
			want: []string{`„Hallo!", sagte er.`},
		},
		{
			name: "umlauts_and_eszett",
			in:   "Große Straße. Er läuft.",
			want: []string{"Große Straße.", "Er läuft."},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Sentences(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d sentences %v, want %v", len(got), got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("[%d] got %q want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}
