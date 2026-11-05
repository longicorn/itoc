package discovery

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
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

func TestDiscoverTestsIdentifierMatchesKeepLineAndBoundaries(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{"a.go": "package x\n", "a_test.go": "Alpha ExtraAlpha\nBeta Alpha\nAlphaSuffix\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctags := filepath.Join(t.TempDir(), "ctags")
	script := "#!/bin/sh\nprintf '%s\\n' '{\"_type\":\"tag\",\"name\":\"Alpha\"}' '{\"_type\":\"tag\",\"name\":\"Beta\"}' '{\"_type\":\"tag\",\"name\":\"Absent\"}'\n"
	if err := os.WriteFile(ctags, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	runner := &testHistoryRunner{calls: map[string]int{}}
	got, err := discoverTests(root, []string{"a.go"}, ctags, io.Discard, runner, false)
	if err != nil {
		t.Fatal(err)
	}
	var details []string
	for _, e := range got {
		if e.Path == "a_test.go" {
			for _, d := range e.Details {
				if strings.HasPrefix(d, "textual identifier") {
					details = append(details, d)
				}
			}
		}
	}
	want := []string{`textual identifier "Alpha" from a.go at a_test.go:1`, `textual identifier "Alpha" from a.go at a_test.go:2`, `textual identifier "Beta" from a.go at a_test.go:2`}
	if !reflect.DeepEqual(details, want) {
		t.Fatalf("got %q, want %q", details, want)
	}
}

func TestFindIdentifierMatchesMatchesLineScan(t *testing.T) {
	names := []string{"Alpha", "Alpha-Beta", "Beta", "ééé", "foo.bar"}
	sort.Strings(names)
	for _, content := range []string{
		"Alpha Alpha Beta\n_Beta Alpha-Beta\nfoo.bar! ééé\n", "xAlpha AlphaX\nBeta\n\n", "", "Alpha-BetaAlpha-Beta\n", "foo.bar fooXbar éééx\n",
	} {
		var want []identifierMatch
		for i, line := range strings.Split(content, "\n") {
			for _, name := range names {
				if containsIdentifier(line, name) {
					want = append(want, identifierMatch{line: i + 1, name: name})
				}
			}
		}
		got := newIdentifierMatcher(names).find([]byte(content))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("content %q: got %#v, want %#v", content, got, want)
		}
	}
}

func TestParallelTestScanPreservesEvidence(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package x\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 40; i++ {
		path := filepath.Join(root, fmt.Sprintf("case_%02d_test.go", i))
		if err := os.WriteFile(path, []byte("Alpha Beta\nAlpha\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctags := filepath.Join(t.TempDir(), "ctags")
	if err := os.WriteFile(ctags, []byte("#!/bin/sh\necho '{\"_type\":\"tag\",\"name\":\"Alpha\"}'\necho '{\"_type\":\"tag\",\"name\":\"Beta\"}'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	original := runtime.GOMAXPROCS(1)
	t.Cleanup(func() { runtime.GOMAXPROCS(original) })
	run := func() map[string][]string {
		t.Helper()
		runner := &testHistoryRunner{calls: map[string]int{}}
		evidence, err := discoverTests(root, []string{"a.go"}, ctags, io.Discard, runner, false)
		if err != nil {
			t.Fatal(err)
		}
		got := map[string][]string{}
		for _, item := range evidence {
			got[item.Path] = item.Details
		}
		return got
	}
	sequential := run()
	runtime.GOMAXPROCS(4)
	parallel := run()
	if !reflect.DeepEqual(sequential, parallel) {
		t.Fatal("parallel scan changed evidence or its order")
	}
	if len(parallel) != 40 {
		t.Fatalf("got %d test files, want 40", len(parallel))
	}
}
