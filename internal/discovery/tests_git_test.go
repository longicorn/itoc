package discovery

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type testHistoryRunner struct {
	mu       sync.Mutex
	calls    map[string]int
	failures map[string]error
}

func (r *testHistoryRunner) Run(_ string, args ...string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if args[0] == "log" {
		if args[len(args)-1] == "a.go" {
			return "c2\nc1\n", nil
		}
		return "c1\nc3\n", nil
	}
	return "", nil
}

func (r *testHistoryRunner) RunInput(_ string, input string, _ ...string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out strings.Builder
	for _, id := range strings.Fields(input) {
		r.calls[id]++
		if err := r.failures[id]; err != nil {
			return "", err
		}
		out.WriteString("\x00" + id + "\x00\n\na_test.go\n")
	}
	return out.String(), nil
}

func TestDiscoverTestsReusesDiffsAndPreservesOrder(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.go", "b.go", "a_test.go"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("package x\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	runner := &testHistoryRunner{calls: map[string]int{}}
	got, err := discoverTests(root, []string{"a.go", "b.go"}, "", io.Discard, runner, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(runner.calls, map[string]int{"c1": 1, "c2": 1, "c3": 1}) {
		t.Fatalf("calls = %v", runner.calls)
	}
	var details []string
	for _, item := range got {
		if item.Path == "a_test.go" {
			for _, d := range item.Details {
				if strings.Contains(d, "Git co-change") {
					details = append(details, d)
				}
			}
		}
	}
	want := []string{"Git co-change with implementation a.go in commit c2", "Git co-change with implementation a.go in commit c1", "Git co-change with implementation b.go in commit c1", "Git co-change with implementation b.go in commit c3"}
	if !reflect.DeepEqual(details, want) {
		t.Fatalf("details = %v, want %v", details, want)
	}
}

func TestLoadCommitDiffsReturnsBatchError(t *testing.T) {
	first, second := errors.New("first"), errors.New("second")
	runner := &testHistoryRunner{calls: map[string]int{}, failures: map[string]error{"c2": first, "c1": second}}
	if err := loadCommitDiffs("", []string{"c2", "c1"}, map[string]string{}, runner); !errors.Is(err, first) {
		t.Fatalf("error = %v", err)
	}
}

func TestBatchDiffMatchesIndividualGitCommands(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("config", "user.name", "Test")
	git("config", "user.email", "test@example.com")
	write := func(name, contents string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("a_test.go", "one")
	git("add", ".")
	git("commit", "-qm", "root")
	base := git("rev-parse", "HEAD")
	git("checkout", "-qb", "feature")
	write("feature.go", "feature")
	git("add", ".")
	git("commit", "-qm", "feature")
	feature := git("rev-parse", "HEAD")
	git("checkout", "-q", "--detach", base)
	write("main.go", "main")
	git("add", ".")
	git("commit", "-qm", "main")
	git("merge", "--no-ff", "-qm", "merge", "feature")
	merge := git("rev-parse", "HEAD")
	git("commit", "--allow-empty", "-qm", "empty")
	empty := git("rev-parse", "HEAD")
	ids := []string{base, feature, merge, empty}
	cache := map[string]string{}
	if err := loadCommitDiffs(root, ids, cache, execGitRunner{}); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		want, err := Git(root, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", id)
		if err != nil {
			t.Fatal(err)
		}
		if cache[id] != want {
			t.Errorf("commit %s: got %q, want %q", id, cache[id], want)
		}
	}
}
