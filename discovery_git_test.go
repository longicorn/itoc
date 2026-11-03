package main

import (
	"strings"
	"testing"
)

type recordingGitRunner struct{ calls int }

func (r *recordingGitRunner) Run(string, ...string) (string, error) { r.calls++; return "", nil }

func TestGitPathListerUsesInjectedRunner(t *testing.T) {
	runner := &recordingGitRunner{}
	_, _, err := (GitPathLister{runner: runner}).List("missing")
	if err == nil {
		t.Fatal("expected missing repository error")
	}
	if runner.calls != 2 {
		t.Fatalf("runner calls = %d, want 2", runner.calls)
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
		byPath[e.Path] = e
	}
	alpha := byPath["src/alpha.go"]
	if len(alpha.Commits) != 2 || alpha.NoiseScore != 1 || len(alpha.NoiseTerms) != 1 || alpha.NoiseTerms[0] != "refactor" {
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
	if a == nil || len(a.Commits) != 2 || a.NoiseScore != 1 || len(a.NoiseTerms) != 1 || a.NoiseTerms[0] != "format" {
		t.Fatalf("source evidence = %#v", a)
	}
}

func TestResolveGitTargetTrackedAndDeleted(t *testing.T) {
	runner := scriptedGitRunner{responses: map[string]string{
		"ls-files --full-name -- docs/a.md":    "docs/a.md\n",
		"ls-files --full-name -- docs/gone.md": "",
	}}
	tracked, err := resolveGitTarget("/repo", "/repo", "docs/a.md", GitPathLister{runner: runner})
	if err != nil || tracked != "docs/a.md" {
		t.Fatalf("tracked target=%q err=%v", tracked, err)
	}
	if _, err := resolveGitTarget("/repo", "/repo", "docs/gone.md", GitPathLister{runner: runner}); err == nil || !strings.Contains(err.Error(), "not tracked") {
		t.Fatalf("deleted target error=%v, want untracked path error", err)
	}
}
