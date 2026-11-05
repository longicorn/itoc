package discovery

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs/alpha.md"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	runner := scriptedGitRunner{responses: map[string]string{
		"rev-parse --show-toplevel":                         root + "\n",
		"ls-files --full-name -- docs/alpha.md":             "docs/alpha.md\n",
		"log --format=%H -- docs/alpha.md":                  "c1\nc2\n",
		"show -s --format=%H%x00%s%n%b%x00 c1 c2":           "c1\x00refactor module\n\x00\nc2\x00feature update\n\x00\n",
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
	byPath := map[string]Evidence{}
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

func TestRelocationFindsSplitBySymbolsAndContent(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.name", "Test")
	git("config", "user.email", "test@example.com")
	old := "package x\nfunc Parse() {}\nfunc Validate() {}\n"
	if err := os.WriteFile(filepath.Join(root, "old.go"), []byte(old), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-qm", "initial")
	if err := os.Remove(filepath.Join(root, "old.go")); err != nil {
		t.Fatal(err)
	}
	for p, body := range map[string]string{"parser.go": "package x\nfunc Parse() {}\n", "validator.go": "package x\nfunc Validate() {}\n", "copy.go": old} {
		if err := os.WriteFile(filepath.Join(root, p), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	git("add", "-A")
	git("commit", "-qm", "split")
	ctags := filepath.Join(t.TempDir(), "ctags")
	script := "#!/bin/sh\nfor x do path=$x; done\nif grep -q Parse \"$path\"; then echo '{\"_type\":\"tag\",\"name\":\"Parse\"}'; fi\nif grep -q Validate \"$path\"; then echo '{\"_type\":\"tag\",\"name\":\"Validate\"}'; fi\n"
	if err := os.WriteFile(ctags, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	got, target, err := (GitPathLister{}).WithRelocation(ctags, io.Discard).List(filepath.Join(root, "old.go"))
	if err != nil {
		t.Fatal(err)
	}
	if target != "old.go" {
		t.Fatalf("target %q", target)
	}
	found := map[string]string{}
	for _, e := range got {
		found[e.Path] = strings.Join(e.Details, " ")
	}
	for _, name := range []string{"parser.go", "validator.go", "copy.go"} {
		if found[name] == "" {
			t.Errorf("missing %s: %v", name, found)
		}
	}
	if !strings.Contains(found["copy.go"], "identical content") || !strings.Contains(found["parser.go"], "Parse") {
		t.Fatalf("clues: %v", found)
	}
}

func TestGitPathListerFollowIsOptional(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.name", "Test")
	git("config", "user.email", "test@example.com")
	for name, body := range map[string]string{"old.go": "package x\n", "related.go": "package x\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	git("add", ".")
	git("commit", "-qm", "initial")
	git("mv", "old.go", "new.go")
	git("commit", "-qm", "rename")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	for _, tc := range []struct{ follow, want bool }{{false, false}, {true, true}} {
		evidence, _, err := (GitPathLister{}).WithFollow(tc.follow).List("new.go")
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range evidence {
			if e.Path == "related.go" {
				found = true
			}
		}
		if found != tc.want {
			t.Errorf("follow=%v found related=%v, want %v", tc.follow, found, tc.want)
		}
	}
}
