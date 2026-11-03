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
func TestDetectNoiseTerms(t *testing.T) {
	got := detectNoiseTerms("Refactoring and formatter update")
	if len(got) != 2 || got[0] != "formatter" || got[1] != "refactoring" {
		t.Fatalf("unexpected terms: %#v", got)
	}
}
func TestAnalyzeCandidatesFromEvidence(t *testing.T) {
	cs := analyzeCandidates("docs/user-guide.md", []cochangeEvidence{{
		Path: "internal/user_guide_test.go", Commits: map[string]bool{"a": true, "b": true}, NoiseScore: 1,
		NoiseTerms: []string{"refactor"},
	}})
	if len(cs) != 1 || cs[0].score != 5 || cs[0].nameBonus != 3 || !cs[0].testLike || cs[0].noiseScore != 1 || len(cs[0].commits) != 2 {
		t.Fatalf("analyzed candidate = %#v", cs)
	}
}
