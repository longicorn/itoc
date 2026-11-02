package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseCLIDefaults(t *testing.T) {
	var out bytes.Buffer
	got, path, err := parseCLI([]string{"docs/adr.md"}, &out)
	if err != nil {
		t.Fatal(err)
	}
	if path != "docs/adr.md" || got.format != "text" || got.limit != 10 || got.minScore != 0 {
		t.Fatalf("config=%#v path=%q", got, path)
	}
}

func TestParseCLIOptionsAndValidation(t *testing.T) {
	var out bytes.Buffer
	got, _, err := parseCLI([]string{"--format", "json", "--limit", "2", "--min-score", "3", "--include-ext", ".go", "--include-ext", ".rs", "src/doc.md"}, &out)
	if err != nil {
		t.Fatal(err)
	}
	if got.format != "json" || got.limit != 2 || got.minScore != 3 || len(got.includeExt) != 2 {
		t.Fatalf("config=%#v", got)
	}
	_, _, err = parseCLI([]string{"--limit", "-1", "x.md"}, &out)
	if err == nil || !strings.Contains(err.Error(), "limit must be zero") {
		t.Fatalf("validation error=%v", err)
	}
}
