package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunEndToEndWithInjectedLister(t *testing.T) {
	var stdout, stderr bytes.Buffer
	lister := fixedPathLister{target: "docs/alpha.md", evidence: []cochangeEvidence{
		{path: "src/alpha.go", commits: map[string]bool{"c": true}},
		{path: "src/beta.go", commits: map[string]bool{"c": true}},
	}}
	err := run([]string{"--format", "json", "--include-path", "alpha", "docs/alpha.md"}, &stdout, &stderr, lister)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"path":"src/alpha.go"`) || strings.Contains(stdout.String(), "beta") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestDiscoverManySkipsUnsupportedDocumentFormat(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "doc.rb")
	if err := os.WriteFile(input, []byte("class Example; end"), 0600); err != nil {
		t.Fatal(err)
	}
	git := fixedPathLister{target: "doc.rb", evidence: []cochangeEvidence{{path: "app/example.rb", commits: map[string]bool{"c": true}}}}
	got, err := discoverMany(input, []namedLister{{"git", git}, {"document", NewDocumentPathLister(dir)}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].path != "app/example.rb" {
		t.Fatalf("got %#v", got)
	}
}

func TestDocumentDiscoveryStillRejectsUnsupportedFormatAlone(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "doc.rb")
	if err := os.WriteFile(input, []byte("class Example; end"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := discoverWithLister(input, NewDocumentPathLister(dir))
	var unsupported unsupportedDocumentFormatError
	if !errors.As(err, &unsupported) {
		t.Fatalf("error=%v", err)
	}
}

func TestRunRejectsInvalidCLI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--format", "xml", "docs/a.md"}, &stdout, &stderr, fixedPathLister{})
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("error=%v", err)
	}
}
