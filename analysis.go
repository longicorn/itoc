package main

import (
	"os"
	"path/filepath"
	"strings"
)

func detectNoiseTerms(message string) []string {
	keywords := []string{"lint", "linter", "format", "formatter", "refactor", "refactoring"}
	lower := strings.ToLower(message)
	var found []string
	for _, word := range keywords {
		if containsWord(lower, word) {
			found = append(found, word)
		}
	}
	return found
}
func containsWord(s, word string) bool {
	for offset := 0; offset < len(s); {
		i := strings.Index(s[offset:], word)
		if i < 0 {
			return false
		}
		i += offset
		end := i + len(word)
		if (i == 0 || !isASCIIWord(s[i-1])) && (end == len(s) || !isASCIIWord(s[end])) {
			return true
		}
		offset = end
	}
	return false
}
func isASCIIWord(b byte) bool { return b == '_' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' }
func appendUnique(items []string, value string) []string {
	for _, v := range items {
		if v == value {
			return items
		}
	}
	return append(items, value)
}

func isTestPath(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	for _, part := range strings.Split(strings.ToLower(filepath.ToSlash(path)), "/") {
		if part == "test" || part == "tests" || part == "__tests__" || part == "spec" || part == "specs" {
			return true
		}
	}
	ext := strings.ToLower(filepath.Ext(base))
	switch ext {
	case ".go":
		return strings.HasSuffix(stem, "_test")
	case ".py":
		return strings.HasPrefix(stem, "test_") || strings.HasSuffix(stem, "_test")
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
		return strings.HasSuffix(stem, ".test") || strings.HasSuffix(stem, ".spec")
	case ".java", ".kt":
		return strings.HasSuffix(stem, "test") || strings.HasSuffix(stem, "tests")
	case ".rb":
		return strings.HasSuffix(stem, "_test") || strings.HasSuffix(stem, "_spec")
	case ".rs":
		return strings.HasSuffix(stem, "_test")
	}
	return false
}

func absOrDot(p string) string {
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return filepath.Dir(p)
	}
	return p
}
func isCode(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".go", ".py", ".js", ".jsx", ".ts", ".tsx", ".java", ".kt", ".rs", ".c", ".h", ".cc", ".cpp", ".hpp", ".cs", ".rb", ".php", ".swift", ".scala", ".sh":
		return true
	}
	return false
}
func isDocument(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".md", ".mdx", ".rst", ".adoc", ".txt", ".org":
		return true
	}
	return false
}
func filenameSimilarityBonus(target, candidate string) int {
	a, b := weightedPathTokens(target), weightedPathTokens(candidate)
	sharedWeight := 0
	for token, aw := range a {
		if bw, ok := b[token]; ok {
			if aw < bw {
				sharedWeight += aw
			} else {
				sharedWeight += bw
			}
		}
	}
	if sharedWeight == 0 {
		return 0
	}
	if sharedWeight >= 4 {
		return 3
	}
	if sharedWeight >= 2 {
		return 2
	}
	return 1
}

func weightedPathTokens(path string) map[string]int {
	parts := strings.Split(filepath.ToSlash(path), "/")
	out := map[string]int{}
	for i, part := range parts {
		weight := 1
		if i == len(parts)-1 {
			weight = 2
		}
		for token := range filenameTokens(part) {
			if weight > out[token] {
				out[token] = weight
			}
		}
	}
	return out
}
func filenameTokens(name string) map[string]bool {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	var parts []string
	start := 0
	for i, r := range name {
		if r == '_' || r == '-' || r == '.' || r == ' ' {
			if i > start {
				parts = append(parts, name[start:i])
			}
			start = i + 1
			continue
		}
		if i > start && r >= 'A' && r <= 'Z' {
			prev := rune(name[i-1])
			if prev >= 'a' && prev <= 'z' || prev >= '0' && prev <= '9' {
				parts = append(parts, name[start:i])
				start = i
			}
		}
	}
	if start < len(name) {
		parts = append(parts, name[start:])
	}
	out := map[string]bool{}
	for _, p := range parts {
		p = strings.ToLower(p)
		if len(p) > 1 {
			out[p] = true
		}
	}
	return out
}

func analyzeCandidates(target string, evidence []cochangeEvidence) []candidate {
	out := make([]candidate, 0, len(evidence))
	for _, e := range evidence {
		c := candidate{path: e.Path, score: len(e.Commits), nameBonus: filenameSimilarityBonus(target, e.Path), testLike: isTestPath(e.Path), noiseScore: e.NoiseScore, noiseTerms: e.NoiseTerms, commits: e.Commits, evidence: e.Details, discoveries: e.Methods, kind: "code"}
		c.score += c.nameBonus
		if isDocument(c.path) {
			c.kind = "document"
		}
		out = append(out, c)
	}
	return out
}
