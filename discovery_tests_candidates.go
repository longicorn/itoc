package main

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
func expandTestCandidates(root string, candidates []candidate, ctagsPath string, warnings io.Writer) ([]candidate, error) {
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
		if isCode(path) && isTestPath(path) {
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
	byPath := map[string]*candidate{}
	for i := range candidates {
		c := candidates[i]
		byPath[c.path] = &c
	}
	add := func(p, evidence string) {
		c := byPath[p]
		if c == nil {
			c = &candidate{path: p, kind: "code", testLike: true, commits: map[string]bool{}}
			terms := detectPathNoiseTerms(p)
			c.noiseTerms = terms
			c.noiseScore = noiseScoreForTerms(terms)
			byPath[p] = c
		}
		c.evidence = appendUnique(c.evidence, evidence)
	}
	for _, impl := range candidates {
		if impl.kind != "code" || impl.testLike {
			continue
		}
		abs := filepath.Join(root, filepath.FromSlash(impl.path))
		if st, e := os.Stat(abs); e != nil || !st.Mode().IsRegular() {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(impl.path), filepath.Ext(impl.path))
		names := map[string]bool{}
		if ctagsPath != "" {
			var e error
			names, e = ctags.Names(ctagsPath, abs)
			if errors.Is(e, ctags.ErrInvalidOutput) {
				fmt.Fprintf(warnings, "itoc: warning: invalid ctags output for %s; identifier-based discovery skipped\n", impl.path)
			} else if e != nil {
				fmt.Fprintf(warnings, "itoc: warning: ctags failed for %s; identifier-based discovery skipped: %v\n", impl.path, e)
			}

		}

		log, e := git(root, "log", "--format=%H", "--follow", "--", impl.path)
		if e != nil {
			return nil, e
		}
		changed := map[string][]string{}
		for _, id := range strings.Fields(log) {
			files, err := git(root, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", id)
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
			if stem == normalized && filepath.Ext(p) == filepath.Ext(impl.path) {
				add(p, fmt.Sprintf("test filename matches implementation %s", impl.path))
			}
			for _, id := range changed[p] {
				add(p, fmt.Sprintf("Git co-change with implementation %s in commit %s", impl.path, id))
			}
			lines := bytes.Split(contents[p], []byte("\n"))
			for i, line := range lines {
				for _, n := range ordered {
					if containsIdentifier(string(line), n) {
						add(p, fmt.Sprintf("textual identifier %q from %s at %s:%d", n, impl.path, p, i+1))
					}
				}
			}
		}
	}
	out := make([]candidate, 0, len(byPath))
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
func isIdentifierByte(b byte) bool { return isASCIIWord(b) || b >= 'A' && b <= 'Z' || b >= 128 }
