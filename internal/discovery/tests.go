package discovery

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"itoc/internal/discovery/ctags"
)

// expandTestCandidates follows implementation candidates once. Its evidence is
// deliberately separate from independent document discovery.
// DiscoverTests collects derived clues for test paths associated with implementation paths.
func DiscoverTests(root string, implementations []string, ctagsPath string, warnings io.Writer) ([]Evidence, error) {
	return DiscoverTestsWithFollow(root, implementations, ctagsPath, warnings, false)
}

func DiscoverTestsWithFollow(root string, implementations []string, ctagsPath string, warnings io.Writer, follow bool) ([]Evidence, error) {
	return discoverTests(root, implementations, ctagsPath, warnings, execGitRunner{}, follow)
}

type testGitRunner interface {
	GitRunner
	RunInput(dir, input string, args ...string) (string, error)
}

func discoverTests(root string, implementations []string, ctagsPath string, warnings io.Writer, runner testGitRunner, follow bool) ([]Evidence, error) {
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

		args := []string{"log", "--format=%H"}
		if follow {
			args = append(args, "--follow")
		}
		args = append(args, "--", impl)
		log, e := runner.Run(root, args...)
		if e != nil {
			return nil, e
		}
		ids := strings.Fields(log)
		err := loadCommitDiffs(root, ids, diffs, runner)
		if err != nil {
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
		matcher := newIdentifierMatcher(ordered)
		// Bound the number of retained results and merge them in test order.
		for first := 0; first < len(tests); first += 32 {
			end := min(first+32, len(tests))
			batch := tests[first:end]
			matches := make([][]identifierMatch, len(batch))

			if len(matcher.states) > 1 {
				jobs := make(chan int)
				var workers sync.WaitGroup
				count := min(8, runtime.GOMAXPROCS(0), len(batch))
				for i := 0; i < count; i++ {
					workers.Add(1)
					go func() {
						defer workers.Done()
						for index := range jobs {
							matches[index] = matcher.find(contents[batch[index]])
						}
					}()
				}
				for i := range batch {
					jobs <- i
				}
				close(jobs)
				workers.Wait()
			}

			for index, p := range batch {
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

				for _, match := range matches[index] {
					add(p, fmt.Sprintf("textual identifier %q from %s at %s:%d", match.name, impl, p, match.line))
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

// loadCommitDiffs reuses results and fetches missing commits in one Git process.
func loadCommitDiffs(root string, ids []string, cache map[string]string, runner testGitRunner) error {
	var missing []string
	seen := make(map[string]bool)
	for _, id := range ids {
		if _, ok := cache[id]; !ok && !seen[id] {
			missing = append(missing, id)
			seen[id] = true
		}
	}
	if len(missing) == 0 {
		return nil
	}
	output, err := runner.RunInput(root, strings.Join(missing, "\n")+"\n", "diff-tree", "--stdin", "--root", "--name-only", "-r", "--format=%x00%H%x00")
	if err != nil {
		return err
	}
	// Merge commits may produce no output, just as with the original command.
	for _, id := range missing {
		cache[id] = ""
	}
	parts := strings.Split(output, "\x00")
	if parts[0] != "" {
		return fmt.Errorf("unexpected git diff-tree output")
	}
	for i := 1; i < len(parts); i += 2 {
		if i+1 >= len(parts) || !seen[parts[i]] || !strings.HasPrefix(parts[i+1], "\n\n") {
			return fmt.Errorf("unexpected git diff-tree output")
		}
		cache[parts[i]] = strings.TrimPrefix(parts[i+1], "\n\n")
	}
	return nil
}

type identifierMatch struct {
	line int
	name string
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
