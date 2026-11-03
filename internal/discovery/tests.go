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

	"itoc/internal/discovery/ctags"
)

// expandTestCandidates follows implementation candidates once. Its evidence is
// deliberately separate from independent document discovery.
// DiscoverTests collects derived clues for test paths associated with implementation paths.
func DiscoverTests(root string, implementations []string, ctagsPath string, warnings io.Writer) ([]Evidence, error) {
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

		log, e := Git(root, "log", "--format=%H", "--follow", "--", impl)
		if e != nil {
			return nil, e
		}
		changed := map[string][]string{}
		for _, id := range strings.Fields(log) {
			files, err := Git(root, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", id)
			if err != nil {
				return nil, err
			}
			for _, p := range strings.Split(files, "\n") {
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
