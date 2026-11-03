package discovery

import (
	"path/filepath"
	"strings"
)

func DetectNoiseTerms(message string) []string {
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
		if (i == 0 || !IsASCIIWord(s[i-1])) && (end == len(s) || !IsASCIIWord(s[end])) {
			return true
		}
		offset = end
	}
	return false
}
func IsASCIIWord(b byte) bool { return b == '_' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' }
func AppendUnique(items []string, value string) []string {
	for _, v := range items {
		if v == value {
			return items
		}
	}
	return append(items, value)
}

func IsTestPath(path string) bool {
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

func IsCode(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".go", ".py", ".js", ".jsx", ".ts", ".tsx", ".java", ".kt", ".rs", ".c", ".h", ".cc", ".cpp", ".hpp", ".cs", ".rb", ".php", ".swift", ".scala", ".sh":
		return true
	}
	return false
}
func IsDocument(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".md", ".mdx", ".rst", ".adoc", ".txt", ".org":
		return true
	}
	return false
}
