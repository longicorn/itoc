package main

import "itoc/internal/discovery"

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type recordingLister struct{ path *string }

func (l recordingLister) List(path string) ([]discovery.Evidence, string, error) {
	*l.path = path
	return nil, path, nil
}

func TestKeywordSearch(t *testing.T) {
	root := t.TempDir()
	cmd := exec.Command("git", "init", "-q", root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}
	for name, content := range map[string]string{"a.txt": "needle one\n", "b.txt": "needle two\n", "c.txt": "unique\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %s: %v", out, err)
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	for _, tc := range []struct{ keyword, input, want string }{{"unique", "", "c.txt"}, {"needle", "2\n", "b.txt"}, {"absent", "", ""}} {
		var diag bytes.Buffer
		got, err := searchKeyword(tc.keyword, strings.NewReader(tc.input), &diag)
		if err != nil || got != tc.want {
			t.Fatalf("search %q: got %q, %v", tc.keyword, got, err)
		}
	}
	var selected string
	var stdout, stderr bytes.Buffer
	if err := runWithInput([]string{"--discover-tests=false", "--format=json", "--search", "needle"}, strings.NewReader("2\n"), &stdout, &stderr, recordingLister{&selected}); err != nil {
		t.Fatal(err)
	}
	if selected != "b.txt" || strings.TrimSpace(stdout.String()) != "[]" {
		t.Fatalf("selected %q, output %q", selected, stdout.String())
	}
	if !strings.Contains(stderr.String(), "Select a file number") {
		t.Fatalf("missing prompt: %q", stderr.String())
	}
	if _, err := searchKeyword("needle", strings.NewReader("9\n"), &bytes.Buffer{}); err == nil {
		t.Fatal("invalid selection accepted")
	}
}
