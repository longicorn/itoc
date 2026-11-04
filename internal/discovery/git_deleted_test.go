package discovery

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDeletedTargetHistory(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = root
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	git("init")
	for _, name := range []string{"target.md", "related.go"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	git("add", ".")
	git("commit", "-m", "initial")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(cwd) })
	if err := os.Remove("target.md"); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"unstaged", "staged", "committed"} {
		if phase == "staged" {
			git("add", "-u")
		}
		if phase == "committed" {
			git("commit", "-m", "delete")
		}
		if _, _, err := (GitPathLister{}).List("target.md"); err == nil {
			t.Fatalf("%s: expected default rejection", phase)
		}
		evidence, _, err := (GitPathLister{AllowDeleted: true}).List("target.md")
		if err != nil {
			t.Fatalf("%s: %v", phase, err)
		}
		if len(evidence) != 1 || evidence[0].Path != "related.go" {
			t.Fatalf("%s: evidence %#v", phase, evidence)
		}
	}
	if _, _, err := (GitPathLister{AllowDeleted: true}).List("unknown.md"); err == nil {
		t.Fatal("expected unknown path rejection")
	}
}
