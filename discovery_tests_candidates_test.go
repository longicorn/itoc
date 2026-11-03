package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandTestCandidates(t *testing.T) {
	ctagsPath, err := exec.LookPath("ctags")
	if err != nil {
		t.Skip("ctags unavailable")
	}
	root := t.TempDir()
	for p, body := range map[string]string{"widget.go": "package x\nfunc UniqueWidget() {}\n", "widget_test.go": "package x\nfunc TestWidget() { UniqueWidget() }\n", "tests/other_test.go": "package x\n// UniqueWidget is mentioned\n", "tests/unrelated_test.go": "package x\n// UniqueWidgetExtra\n"} {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "test@example.com"}, {"config", "user.name", "Test"}, {"add", "."}, {"commit", "-qm", "initial"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %v: %s", err, b)
		}
	}
	got, err := expandTestCandidates(root, []candidate{{path: "widget.go", kind: "code"}, {path: "widget_test.go", kind: "code", testLike: true, evidence: []string{"existing"}}}, ctagsPath, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]candidate{}
	for _, c := range got {
		found[c.path] = c
	}
	c := found["widget_test.go"]
	evidence := strings.Join(c.evidence, " | ")
	if !strings.Contains(evidence, "existing") || !strings.Contains(evidence, "filename matches") || !strings.Contains(evidence, "textual identifier") || !strings.Contains(evidence, "Git co-change") {
		t.Fatalf("missing evidence: %s", evidence)
	}
	if !strings.Contains(strings.Join(found["tests/other_test.go"].evidence, " "), "textual identifier") {
		t.Fatal("missing textual match")
	}
	if strings.Contains(strings.Join(found["tests/unrelated_test.go"].evidence, " "), "textual identifier") {
		t.Fatal("matched identifier substring")
	}
	if c.score != 0 {
		t.Fatalf("unexpected score %d", c.score)
	}
}

func TestExpandTestsWithoutCtags(t *testing.T) {
	root := t.TempDir()
	for p, content := range map[string]string{"widget.go": "package x", "widget_test.go": "package x"} {
		if err := os.WriteFile(filepath.Join(root, p), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "test@example.com"}, {"config", "user.name", "Test"}, {"add", "."}, {"commit", "-qm", "initial"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %v: %s", err, b)
		}
	}
	for _, ctags := range []string{"", filepath.Join(root, "nonexistent-ctags")} {
		var warnings strings.Builder
		got, err := expandTestCandidates(root, []candidate{{path: "widget.go", kind: "code"}, {path: "widget_test.go", kind: "code", testLike: true, score: 7, evidence: []string{"existing"}}}, ctags, &warnings)
		if err != nil {
			t.Fatal(err)
		}
		var evidence string
		for _, c := range got {
			if c.path == "widget_test.go" {
				evidence = strings.Join(c.evidence, " ")
				if c.score != 7 || !strings.Contains(evidence, "existing") {
					t.Fatalf("existing candidate changed: %#v", c)
				}
			}
		}
		if !strings.Contains(evidence, "filename matches") || !strings.Contains(evidence, "Git co-change") {
			t.Fatalf("ctags=%q evidence=%q", ctags, evidence)
		}
		if ctags != "" && !strings.Contains(warnings.String(), "warning") {
			t.Fatal("missing warning")
		}
	}
}
