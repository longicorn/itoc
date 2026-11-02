package main

import (
	"bytes"
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

func TestRunRejectsInvalidCLI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--format", "xml", "docs/a.md"}, &stdout, &stderr, fixedPathLister{})
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("error=%v", err)
	}
}
