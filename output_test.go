package main

import (
	"strings"
	"testing"
)

func TestRenderCandidatesText(t *testing.T) {
	cs := []candidate{{path: "src/a.go", score: 4, kind: "code", commits: map[string]bool{"c": true}}}
	got, err := renderCandidates(cs, "text")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "Implementation candidates:\n 1. src/a.go (score 4;") {
		t.Fatalf("unexpected text output: %q", got)
	}
}

func TestRenderCandidatesJSON(t *testing.T) {
	got, err := renderCandidates([]candidate{{path: "src/a.go", score: 4, kind: "code"}}, "json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"path":"src/a.go"`) || !strings.HasSuffix(string(got), "\n") {
		t.Fatalf("unexpected JSON output: %q", got)
	}
}
