package main

import "itoc/internal/discovery"

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunEndToEndWithInjectedLister(t *testing.T) {
	var stdout, stderr bytes.Buffer
	lister := fixedPathLister{target: "docs/alpha.md", evidence: []discovery.Evidence{
		{Path: "src/alpha.go", Commits: map[string]bool{"c": true}},
		{Path: "src/beta.go", Commits: map[string]bool{"c": true}},
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
	git := fixedPathLister{target: "doc.rb", evidence: []discovery.Evidence{{Path: "app/example.rb", Commits: map[string]bool{"c": true}}}}
	got, err := discoverMany(input, []namedLister{{"git", git}, {"document", discovery.NewDocumentPathLister(dir)}})
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
	_, err := discoverWithLister(input, discovery.NewDocumentPathLister(dir))
	var unsupported discovery.UnsupportedDocumentFormatError
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

func TestRunWithoutCtagsContinuesTestDiscovery(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{"widget.go": "package x\n", "widget_test.go": "package x\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "test@example.com"}, {"config", "user.name", "Test"}, {"add", "."}, {"commit", "-qm", "initial"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %v: %s", err, out)
		}
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	if err := os.Symlink(gitPath, filepath.Join(bin, "git")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Error(err)
		}
	})
	var stdout, stderr bytes.Buffer
	lister := fixedPathLister{target: "guide.md", evidence: []discovery.Evidence{{Path: "widget.go"}}}
	if err := run([]string{"--format", "json", "guide.md"}, &stdout, &stderr, lister); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "ctags not found") {
		t.Fatalf("missing warning: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), `"path":"widget_test.go"`) {
		t.Fatalf("missing test candidate: %s", stdout.String())
	}
}
