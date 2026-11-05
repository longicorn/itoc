package discovery

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"itoc/internal/discovery/ctags"
)

// expandTestCandidates follows implementation candidates once. Its evidence is
// deliberately separate from independent document discovery.
// DiscoverTests collects derived clues for test paths associated with implementation paths.
func DiscoverTests(root string, implementations []string, ctagsPath string, warnings io.Writer) ([]Evidence, error) {
	return discoverTests(root, implementations, ctagsPath, warnings, execGitRunner{})
}

func discoverTests(root string, implementations []string, ctagsPath string, warnings io.Writer, runner GitRunner) ([]Evidence, error) {
	var tests []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && (d.Name() == ".git" || d.Name() == "vendor" || d.Name() == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if IsCode(path) && IsTestPath(path) {
			p, e := filepath.Rel(root, path)
			if e != nil {
				return e
			}
			tests = append(tests, filepath.ToSlash(p))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	contents := map[string][]byte{}
	for _, p := range tests {
		b, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if e != nil {
			return nil, e
		}
		contents[p] = b
	}
	byPath := map[string]*Evidence{}
	add := func(p, detail string) {
		item := byPath[p]
		if item == nil {
			item = &Evidence{Path: p}
			byPath[p] = item
		}
		item.Details = AppendUnique(item.Details, detail)
	}
	diffs := make(map[string]string)
	for _, impl := range implementations {
		abs := filepath.Join(root, filepath.FromSlash(impl))
		if st, e := os.Stat(abs); e != nil || !st.Mode().IsRegular() {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(impl), filepath.Ext(impl))
		names := map[string]bool{}
		if ctagsPath != "" {
			var e error
			names, e = ctags.Names(ctagsPath, abs)
			if errors.Is(e, ctags.ErrInvalidOutput) {
				fmt.Fprintf(warnings, "itoc: warning: invalid ctags output for %s; identifier-based discovery skipped\n", impl)
			} else if e != nil {
				fmt.Fprintf(warnings, "itoc: warning: ctags failed for %s; identifier-based discovery skipped: %v\n", impl, e)
			}

		}

		log, e := runner.Run(root, "log", "--format=%H", "--follow", "--", impl)
		if e != nil {
			return nil, e
		}
		ids := strings.Fields(log)
		if err := loadCommitDiffs(root, ids, diffs, runner); err != nil {
			return nil, err
		}
		changed := map[string][]string{}
		for _, id := range ids {
			for _, p := range strings.Split(diffs[id], "\n") {
				changed[p] = append(changed[p], id)
			}
		}

		ordered := make([]string, 0, len(names))
		for n := range names {
			ordered = append(ordered, n)
		}
		sort.Strings(ordered)
		for _, p := range tests {
			testStem := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
			normalized := strings.TrimSuffix(strings.TrimPrefix(testStem, "test_"), "_test")
			normalized = strings.TrimSuffix(normalized, "_spec")
			normalized = strings.TrimSuffix(normalized, ".test")
			normalized = strings.TrimSuffix(normalized, ".spec")
			if stem == normalized && filepath.Ext(p) == filepath.Ext(impl) {
				add(p, fmt.Sprintf("test filename matches implementation %s", impl))
			}
			for _, id := range changed[p] {
				add(p, fmt.Sprintf("Git co-change with implementation %s in commit %s", impl, id))
			}
			lines := bytes.Split(contents[p], []byte("\n"))
			for i, line := range lines {
				for _, n := range ordered {
					if containsIdentifier(string(line), n) {
						add(p, fmt.Sprintf("textual identifier %q from %s at %s:%d", n, impl, p, i+1))
					}
				}
			}
		}
	}
	out := make([]Evidence, 0, len(byPath))
	for _, c := range byPath {
		out = append(out, *c)
	}
	return out, nil
}

// loadCommitDiffs reuses successful results across implementations and reports
// errors in commit order, independent of worker completion order.
func loadCommitDiffs(root string, ids []string, cache map[string]string, runner GitRunner) error {
	var missing []string
	seen := make(map[string]bool)
	for _, id := range ids {
		if _, ok := cache[id]; !ok && !seen[id] {
			missing = append(missing, id)
			seen[id] = true
		}
	}
	type result struct {
		files string
		err   error
	}
	results := make([]result, len(missing))
	jobs := make(chan int)
	var workers sync.WaitGroup
	count := min(8, len(missing))
	for i := 0; i < count; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := range jobs {
				results[i].files, results[i].err = runner.Run(root, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", missing[i])
			}
		}()
	}
	for i := range missing {
		jobs <- i
	}
	close(jobs)
	workers.Wait()
	for i, id := range missing {
		if results[i].err != nil {
			return results[i].err
		}
		cache[id] = results[i].files
	}
	return nil
}

func containsIdentifier(line, name string) bool {
	for offset := 0; offset < len(line); {
		i := strings.Index(line[offset:], name)
		if i < 0 {
			return false
		}
		i += offset
		end := i + len(name)
		if (i == 0 || !isIdentifierByte(line[i-1])) && (end == len(line) || !isIdentifierByte(line[end])) {
			return true
		}
		offset = end
	}
	return false
}
func isIdentifierByte(b byte) bool { return IsASCIIWord(b) || b >= 'A' && b <= 'Z' || b >= 128 }
