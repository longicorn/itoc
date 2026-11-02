package main

import (
	"strings"
	"testing"
)

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

type scriptedGitRunner struct{ responses map[string]string }

func (r scriptedGitRunner) Run(_ string, args ...string) (string, error) {
	return r.responses[strings.Join(args, " ")], nil
}

func TestGitPathListerCollectsCochangeEvidence(t *testing.T) {
	runner := scriptedGitRunner{responses: map[string]string{
		"rev-parse --show-toplevel":                         "/repo\n",
		"ls-files --full-name -- docs/alpha.md":             "docs/alpha.md\n",
		"log --follow --format=%H -- docs/alpha.md":         "c1\nc2\n",
		"show -s --format=%s%n%b c1":                        "refactor module\n",
		"show -s --format=%s%n%b c2":                        "feature update\n",
		"diff-tree --root --no-commit-id --name-only -r c1": "docs/alpha.md\nsrc/alpha.go\nREADME.md\n",
		"diff-tree --root --no-commit-id --name-only -r c2": "docs/alpha.md\nsrc/alpha.go\nsrc/beta_test.go\n",
	}}
	evidence, target, err := (GitPathLister{runner: runner}).List("docs/alpha.md")
	if err != nil {
		t.Fatal(err)
	}
	if target != "docs/alpha.md" || len(evidence) != 3 {
		t.Fatalf("target=%q evidence=%#v", target, evidence)
	}
	byPath := map[string]cochangeEvidence{}
	for _, e := range evidence {
		byPath[e.path] = e
	}
	alpha := byPath["src/alpha.go"]
	if len(alpha.commits) != 2 || alpha.noiseScore != 1 || len(alpha.noiseTerms) != 1 || alpha.noiseTerms[0] != "refactor" {
		t.Fatalf("alpha evidence=%#v", alpha)
	}
}

func TestCollectCochangeEvidenceDeduplicatesWithinCommit(t *testing.T) {
	commits := []commitSnapshot{
		{id: "c1", message: "format changes", files: "docs/target.md\nsrc/a.go\nsrc/a.go\nimage.png\n"},
		{id: "c2", message: "feature", files: "docs/target.md\nsrc/a.go\n"},
	}
	got := collectCochangeEvidence("docs/target.md", commits)
	if len(got) != 1 {
		t.Fatalf("evidence = %#v, want only source candidate", got)
	}
	a := got["src/a.go"]
	if a == nil || len(a.commits) != 2 || a.noiseScore != 1 || len(a.noiseTerms) != 1 || a.noiseTerms[0] != "format" {
		t.Fatalf("source evidence = %#v", a)
	}
}

func TestResolveGitTargetTrackedAndDeleted(t *testing.T) {
	runner := scriptedGitRunner{responses: map[string]string{
		"ls-files --full-name -- docs/a.md":        "docs/a.md\n",
		"ls-files --full-name -- docs/gone.md":     "",
		"log --follow --format=%H -- docs/gone.md": "commit\n",
	}}
	tracked, err := resolveGitTarget("/repo", "/repo", "docs/a.md", GitPathLister{runner: runner})
	if err != nil || tracked != "docs/a.md" {
		t.Fatalf("tracked target=%q err=%v", tracked, err)
	}
	deleted, err := resolveGitTarget("/repo", "/repo", "docs/gone.md", GitPathLister{runner: runner})
	if err != nil || deleted != "docs/gone.md" {
		t.Fatalf("deleted target=%q err=%v", deleted, err)
	}
}
