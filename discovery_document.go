package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type documentHint struct {
	value  string
	kind   string
	line   int
	strong bool
}

type documentExtractor interface {
	Extract(string) []documentHint
}

type MarkdownExtractor struct{}
type RSTExtractor struct{}

func (MarkdownExtractor) Extract(content string) []documentHint {
	return extractMarkdownHints(content)
}
func (RSTExtractor) Extract(content string) []documentHint {
	return extractRSTHints(content)
}

type DocumentPathLister struct {
	root       string
	extractors map[string]documentExtractor
}

func NewDocumentPathLister(root string) DocumentPathLister {
	return DocumentPathLister{root: root, extractors: map[string]documentExtractor{
		".md": MarkdownExtractor{}, ".mdx": MarkdownExtractor{}, ".rst": RSTExtractor{},
	}}
}

func (l DocumentPathLister) List(input string) ([]cochangeEvidence, string, error) {
	root := l.root
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, "", err
		}
	}
	abs, err := filepath.Abs(input)
	if err != nil {
		return nil, "", err
	}
	if !filepath.IsAbs(input) {
		abs = filepath.Join(root, input)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, "", err
	}
	content, err := os.ReadFile(abs)
	if err != nil {
		return nil, "", err
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return nil, "", err
	}
	target := filepath.ToSlash(rel)
	ext := l.extractors
	if ext == nil {
		ext = NewDocumentPathLister(root).extractors
	}
	extractor := ext[strings.ToLower(filepath.Ext(target))]
	if extractor == nil {
		return nil, "", fmt.Errorf("unsupported document format: %s", filepath.Ext(target))
	}
	hints := extractor.Extract(string(content))
	index, err := buildDocumentSearchIndex(root)
	if err != nil {
		return nil, "", err
	}
	byPath := map[string]*cochangeEvidence{}
	for _, hint := range hints {
		matches := documentHintMatches(root, filepath.Dir(abs), hint, index)
		for _, match := range matches {
			if match == target || !isCode(match) {
				continue
			}
			if byPath[match] == nil {
				byPath[match] = &cochangeEvidence{path: match, commits: map[string]bool{}, discoveries: []string{"document"}}
			}
			pathTerms := detectPathNoiseTerms(match)
			for _, term := range pathTerms {
				byPath[match].noiseTerms = appendUnique(byPath[match].noiseTerms, term)
			}
			if score := noiseScoreForTerms(pathTerms); score > byPath[match].noiseScore {
				byPath[match].noiseScore = score
			}
			byPath[match].evidence = appendUnique(byPath[match].evidence, fmt.Sprintf("%s %q at %s:%d", hint.kind, hint.value, target, hint.line))
		}
	}
	out := make([]cochangeEvidence, 0, len(byPath))
	for _, item := range byPath {
		out = append(out, *item)
	}
	return out, target, nil
}

type documentSearchEntry struct{ path, base, content string }

func buildDocumentSearchIndex(root string) ([]documentSearchEntry, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && (d.Name() == ".git" || d.Name() == "vendor" || d.Name() == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if isCode(path) && !isTestPath(path) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	entries := make([]documentSearchEntry, len(paths))
	workers := 8
	if len(paths) < workers {
		workers = len(paths)
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				path := paths[index]
				b, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				rel, err := filepath.Rel(root, path)
				if err != nil {
					continue
				}
				entries[index] = documentSearchEntry{path: filepath.ToSlash(rel), base: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), content: string(b)}
			}
		}()
	}
	for i := range paths {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return entries, nil
}

func documentHintMatches(root, docDir string, hint documentHint, index []documentSearchEntry) []string {
	value := strings.TrimSpace(hint.value)
	if value == "" {
		return nil
	}
	if strings.HasPrefix(value, "#") || strings.Contains(value, "://") {
		return nil
	}
	value = strings.SplitN(value, "#", 2)[0]
	value = strings.SplitN(value, "?", 2)[0]
	if value == "" {
		return nil
	}
	candidates := []string{}
	if strings.ContainsAny(value, "/\\.") {
		p := value
		if !filepath.IsAbs(p) {
			p = filepath.Join(docDir, filepath.FromSlash(p))
		}
		if rel, err := filepath.Rel(root, p); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				candidates = append(candidates, filepath.ToSlash(rel))
			}
		}
	}
	if hint.strong && usefulDocumentTerm(value) {
		for _, entry := range index {
			if entry.path == "" {
				continue
			}
			if strings.EqualFold(entry.base, value) || strings.Contains(entry.base, value) || strings.Contains(entry.content, value) {
				candidates = append(candidates, entry.path)
			}
		}
	}
	return candidates
}

func usefulDocumentTerm(value string) bool {
	if len([]rune(value)) < 3 {
		return false
	}
	lower := strings.ToLower(value)
	switch lower {
	case "itoc", "test", "tests", "document", "implementation", "code", "format", "limit", "score", "path", "file", "true", "false":
		return false
	}
	if strings.HasPrefix(value, "-") || strings.ContainsAny(value, "*/\\") {
		return false
	}
	letters := 0
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r > 127 {
			letters++
		}
	}
	return letters >= 3
}
