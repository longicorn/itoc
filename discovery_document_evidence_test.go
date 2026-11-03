package main

import "itoc/internal/discovery"

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentEvidenceRecordsMatchTypesAndFrequency(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{"guide.md": "[alpha](alpha.go) `alpha`\n", "alpha.go": "package alpha\n", "alphabet.go": "package other\n// alpha\n", "other.go": "package other\n"}
	for p, content := range files {
		if err := os.WriteFile(filepath.Join(root, p), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	evidence, _, err := discovery.NewDocumentPathLister(root).List(filepath.Join(root, "guide.md"))
	if err != nil {
		t.Fatal(err)
	}
	byPath := map[string]string{}
	for _, e := range evidence {
		byPath[e.Path] = strings.Join(e.Details, " | ")
	}
	for _, part := range []string{"explicit path", "exact filename", "source text", "matched 2 of 3 indexed files"} {
		if !strings.Contains(byPath["alpha.go"], part) {
			t.Errorf("missing %s: %s", part, byPath["alpha.go"])
		}
	}
	if !strings.Contains(byPath["alphabet.go"], "filename substring") {
		t.Errorf("missing substring: %s", byPath["alphabet.go"])
	}
	if _, ok := byPath["other.go"]; ok {
		t.Fatal("unrelated candidate")
	}
}
