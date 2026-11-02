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

func TestSortBeforePerGroupLimit(t *testing.T) {
	cs := []candidate{
		{path: "src/z.go", score: 1, kind: "code"},
		{path: "src/a.go", score: 5, kind: "code"},
		{path: "src/test_test.go", score: 9, kind: "code", testLike: true},
		{path: "src/b.go", score: 4, kind: "code"},
	}
	got := limitCandidates(sortCandidates(cs), 1)
	if len(got) != 2 || got[0].path != "src/test_test.go" || got[1].path != "src/a.go" {
		t.Fatalf("sort then limit = %#v; want highest candidate in each group", got)
	}
}

func TestPrepareCandidatesPipeline(t *testing.T) {
	cs := []candidate{
		{path: "src/low.go", score: 2, kind: "code"},
		{path: "src/high.go", score: 7, kind: "code"},
		{path: "docs/ref.md", score: 9, kind: "document"},
		{path: "src/high_test.go", score: 8, kind: "code", testLike: true},
	}
	got := prepareCandidates(cs, candidateOptions{
		includeExt: []string{".go"}, excludeCandidate: []string{"src/low.go"}, minScore: 5, limit: 1,
	})
	if len(got) != 2 || got[0].path != "src/high_test.go" || got[1].path != "src/high.go" {
		t.Fatalf("prepared candidates = %#v", got)
	}
}

func TestAnalyzeCandidatesFromEvidence(t *testing.T) {
	cs := analyzeCandidates("docs/user-guide.md", []cochangeEvidence{{
		path: "internal/user_guide_test.go", commits: map[string]bool{"a": true, "b": true}, noiseScore: 1,
		noiseTerms: []string{"refactor"},
	}})
	if len(cs) != 1 || cs[0].score != 5 || cs[0].nameBonus != 3 || !cs[0].testLike || cs[0].noiseScore != 1 || len(cs[0].commits) != 2 {
		t.Fatalf("analyzed candidate = %#v", cs)
	}
}

type fixedPathLister struct {
	evidence []cochangeEvidence
	target   string
}

func (f fixedPathLister) List(string) ([]cochangeEvidence, string, error) {
	return f.evidence, f.target, nil
}

func TestPathListerInterfaceFeedsAnalysis(t *testing.T) {
	lister := fixedPathLister{target: "docs/alpha.md", evidence: []cochangeEvidence{{path: "src/alpha.go", commits: map[string]bool{"c": true}}}}
	evidence, target, err := lister.List("ignored")
	if err != nil {
		t.Fatal(err)
	}
	got := analyzeCandidates(target, evidence)
	if len(got) != 1 || got[0].score != 3 {
		t.Fatalf("got %#v", got)
	}
}

func TestDiscoverWithListerRunsFullDiscoveryFlow(t *testing.T) {
	lister := fixedPathLister{target: "docs/alpha.md", evidence: []cochangeEvidence{
		{path: "src/alpha.go", commits: map[string]bool{"c": true}},
	}}
	got, err := discoverWithLister("ignored", lister)
	if err != nil {
		t.Fatal(err)
	}
	got = prepareCandidates(got, candidateOptions{limit: 1})
	if len(got) != 1 || got[0].path != "src/alpha.go" || got[0].score != 3 {
		t.Fatalf("discovery flow returned %#v", got)
	}
}

type recordingGitRunner struct{ calls int }

func (r *recordingGitRunner) Run(string, ...string) (string, error) { r.calls++; return "", nil }

func TestGitPathListerUsesInjectedRunner(t *testing.T) {
	runner := &recordingGitRunner{}
	_, _, err := (GitPathLister{runner: runner}).List("missing")
	if err == nil {
		t.Fatal("expected missing repository error")
	}
	if runner.calls != 3 {
		t.Fatalf("runner calls = %d, want 3", runner.calls)
	}
}
