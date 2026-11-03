package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type unsupportedDocumentFormatError struct{ extension string }

func (e unsupportedDocumentFormatError) Error() string {
	return "unsupported document format: " + e.extension
}

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
		return nil, "", unsupportedDocumentFormatError{extension: filepath.Ext(target)}
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
			path := match.path
			if path == target || !isCode(path) {
				continue
			}
			if byPath[path] == nil {
				byPath[path] = &cochangeEvidence{path: path, commits: map[string]bool{}, discoveries: []string{"document"}}
			}
			pathTerms := detectPathNoiseTerms(path)
			for _, term := range pathTerms {
				byPath[path].noiseTerms = appendUnique(byPath[path].noiseTerms, term)
			}
			if score := noiseScoreForTerms(pathTerms); score > byPath[path].noiseScore {
				byPath[path].noiseScore = score
			}
			byPath[path].evidence = appendUnique(byPath[path].evidence, documentMatchEvidence(hint, target, match))
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

func documentMatchEvidence(hint documentHint, target string, match documentMatch) string {
	prefix := fmt.Sprintf("%s %q at %s:%d; %s", hint.kind, hint.value, target, hint.line, match.kind)
	if match.kind == "explicit path" {
		return prefix
	}
	return fmt.Sprintf("%s; matched %d of %d indexed files", prefix, match.frequency, match.indexSize)
}

type documentMatch struct {
	path, kind           string
	frequency, indexSize int
}

func documentHintMatches(root, docDir string, hint documentHint, index []documentSearchEntry) []documentMatch {
	value := strings.TrimSpace(hint.value)
	if value == "" || strings.HasPrefix(value, "#") || strings.Contains(value, "://") {
		return nil
	}
	value = strings.SplitN(strings.SplitN(value, "#", 2)[0], "?", 2)[0]
	if value == "" {
		return nil
	}
	var matches []documentMatch
	if strings.ContainsAny(value, "/\\.") {
		p := value
		if !filepath.IsAbs(p) {
			p = filepath.Join(docDir, filepath.FromSlash(p))
		}
		if rel, err := filepath.Rel(root, p); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				matches = append(matches, documentMatch{path: filepath.ToSlash(rel), kind: "explicit path", frequency: 1, indexSize: len(index)})
			}
		}
	}
	if hint.strong && usefulDocumentTerm(value) {
		var terms []documentMatch
		indexedCount := 0
		for _, entry := range index {
			if entry.path != "" {
				indexedCount++
			}
		}
		matchedFiles := map[string]bool{}
		for _, entry := range index {
			if entry.path == "" {
				continue
			}
			if strings.EqualFold(entry.base, value) || strings.EqualFold(filepath.Base(entry.path), value) {
				terms = append(terms, documentMatch{path: entry.path, kind: "exact filename"})
				matchedFiles[entry.path] = true
			} else if strings.Contains(entry.base, value) {
				terms = append(terms, documentMatch{path: entry.path, kind: "filename substring"})
				matchedFiles[entry.path] = true
			}
			if strings.Contains(entry.content, value) {
				terms = append(terms, documentMatch{path: entry.path, kind: "source text"})
				matchedFiles[entry.path] = true
			}
		}
		for _, m := range terms {
			m.frequency = len(matchedFiles)
			m.indexSize = indexedCount
			matches = append(matches, m)
		}
	}
	return matches
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
