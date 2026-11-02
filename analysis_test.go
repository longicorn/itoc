package main

import "testing"

func TestFilenameSimilarityBonus(t *testing.T) {
	for _, tc := range []struct {
		target, candidate string
		want              int
	}{
		{"docs/user-guide.md", "internal/user_guide.go", 3},
		{"docs/alpha.md", "internal/beta.go", 0},
		{"docs/one-two.md", "src/one.go", 2},
	} {
		if got := filenameSimilarityBonus(tc.target, tc.candidate); got != tc.want {
			t.Errorf("bonus(%q, %q) = %d, want %d", tc.target, tc.candidate, got, tc.want)
		}
	}
}

func TestFilterAndLimitCandidates(t *testing.T) {
	cs := []candidate{{path: "src/a.go", score: 4, kind: "code"}, {path: "docs/a.md", score: 3, kind: "document"}, {path: "src/b_test.go", score: 2, kind: "code", testLike: true}}
	got := filterCandidates(cs, []string{".go"}, nil, nil, nil)
	if len(got) != 2 {
		t.Fatalf("filter got %d candidates, want 2", len(got))
	}
	got = limitCandidates(cs, 1)
	if len(got) != 3 {
		t.Fatalf("per-group limit got %d candidates, want 3", len(got))
	}
	got = filterMinScore(cs, 3)
	if len(got) != 2 {
		t.Fatalf("min score got %d candidates, want 2", len(got))
	}
}

func TestDetectNoiseTerms(t *testing.T) {
	got := detectNoiseTerms("Refactoring and formatter update")
	if len(got) != 2 || got[0] != "formatter" || got[1] != "refactoring" {
		t.Fatalf("unexpected terms: %#v", got)
	}
}
