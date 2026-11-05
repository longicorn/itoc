package discovery

import (
	"errors"
	"io"
	"os"
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
	id := args[len(args)-1]
	r.calls[id]++
	return "a_test.go\n", r.failures[id]
}

func TestDiscoverTestsReusesDiffsAndPreservesOrder(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.go", "b.go", "a_test.go"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("package x\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	runner := &testHistoryRunner{calls: map[string]int{}}
	got, err := discoverTests(root, []string{"a.go", "b.go"}, "", io.Discard, runner)
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

func TestLoadCommitDiffsReturnsFirstErrorInCommitOrder(t *testing.T) {
	first, second := errors.New("first"), errors.New("second")
	runner := &testHistoryRunner{calls: map[string]int{}, failures: map[string]error{"c2": first, "c1": second}}
	if err := loadCommitDiffs("", []string{"c2", "c1"}, map[string]string{}, runner); !errors.Is(err, first) {
		t.Fatalf("error = %v", err)
	}
}
