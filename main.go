package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type stringList []string

func (s *stringList) String() string     { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

type candidateExcludeFlag []string

func (f *candidateExcludeFlag) String() string     { return "" }
func (f *candidateExcludeFlag) Set(v string) error { *f = append(*f, v); return nil }

type candidate struct {
	path       string
	score      int
	nameBonus  int
	testLike   bool
	kind       string
	noiseScore int
	noiseTerms []string
	commits    map[string]bool
}

func main() {
	format := flag.String("format", "text", "output format: text or json")
	limit := flag.Int("limit", 10, "maximum candidates per group (0 means unlimited)")
	minScore := flag.Int("min-score", 0, "minimum candidate score (0 disables score filtering)")
	var includeExt, excludeExt, includePath, excludePath stringList
	var excludeCandidate candidateExcludeFlag
	flag.Var(&excludeCandidate, "exclude-candidate", "hide a candidate group (implementation/document/test) or exact repository-relative path (repeatable)")
	flag.Var(&includeExt, "include-ext", "candidate extension allowlist (repeatable, e.g. .go)")
	flag.Var(&excludeExt, "exclude-ext", "candidate extension denylist (repeatable, e.g. .md)")
	flag.Var(&includePath, "include-path", "candidate repository-relative path substring or glob allowlist (repeatable)")
	flag.Var(&excludePath, "exclude-path", "candidate repository-relative path glob denylist (repeatable)")
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintln(out, "Usage: itoc [options] <repository-file>")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Options:")
		fmt.Fprintln(out, "  --format text|json          Output format (default: text)")
		fmt.Fprintln(out, "  --limit N                   Maximum candidates per group (default: 10; 0: unlimited)")
		fmt.Fprintln(out, "  --min-score N               Minimum score required (default: 0; disabled)")
		fmt.Fprintln(out, "  --exclude-candidate PATH    Hide a group (implementation/document/test) or exact path (repeatable)")
		fmt.Fprintln(out, "  --include-ext EXT           Include candidates with this extension (repeatable)")
		fmt.Fprintln(out, "  --exclude-ext EXT           Exclude candidates with this extension (repeatable)")
		fmt.Fprintln(out, "  --include-path PATTERN      Include paths containing this string or matching glob (repeatable)")
		fmt.Fprintln(out, "  --exclude-path GLOB         Exclude matching repository-relative paths (repeatable)")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Plain --include-path values match substrings; glob patterns match the full path.")
		fmt.Fprintln(out, "Plain --exclude-path values match substrings; glob excludes match the full path.")
		fmt.Fprintln(out, "Globs use filepath.Match syntax; '*' does not match across directory separators.")
		fmt.Fprintln(out, "Examples: --include-ext .go --exclude-path 'vendor/*'")
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	if *format != "text" && *format != "json" {
		fatal("unsupported format: " + *format)
	}
	if *limit < 0 {
		fatal("limit must be zero or greater")
	}
	if *minScore < 0 {
		fatal("min-score must be zero or greater")
	}
	out, err := discover(flag.Arg(0))
	if err != nil {
		fatal(err.Error())
	}
	out = filterCandidates(out, includeExt, excludeExt, includePath, excludePath)
	out = excludeExactCandidates(out, excludeCandidate)
	if *minScore > 0 {
		out = filterMinScore(out, *minScore)
	}
	if *limit > 0 {
		out = limitCandidates(out, *limit)
	}
	if *format == "json" {
		printJSON(out)
		return
	}
	if len(out) == 0 {
		fmt.Println("No code candidates found.")
		return
	}
	var impl, tests, docs []candidate
	for _, c := range out {
		if c.kind == "document" {
			docs = append(docs, c)
		} else if c.testLike {
			tests = append(tests, c)
		} else {
			impl = append(impl, c)
		}
	}
	printCandidates("Implementation candidates", impl)
	printCandidates("Test candidates", tests)
	printCandidates("Document candidates", docs)
}

func excludeExactCandidates(cs []candidate, patterns []string) []candidate {
	hidden := map[string]bool{}
	groups := map[string]bool{}
	for _, p := range patterns {
		switch strings.ToLower(p) {
		case "implementation", "document", "test":
			groups[strings.ToLower(p)] = true
		default:
			hidden[filepath.ToSlash(filepath.Clean(p))] = true
		}
	}
	out := make([]candidate, 0, len(cs))
	for _, c := range cs {
		kind := c.kind
		if c.testLike {
			kind = "test"
		}
		if !hidden[filepath.ToSlash(filepath.Clean(c.path))] && !groups[kind] {
			out = append(out, c)
		}
	}
	return out
}
func filterMinScore(cs []candidate, min int) []candidate {
	out := make([]candidate, 0, len(cs))
	for _, c := range cs {
		if c.score >= min {
			out = append(out, c)
		}
	}
	return out
}

func limitCandidates(cs []candidate, limit int) []candidate {
	counts := map[string]int{}
	out := make([]candidate, 0, len(cs))
	for _, c := range cs {
		group := c.kind
		if c.testLike {
			group = "test"
		}
		if counts[group] >= limit {
			continue
		}
		counts[group]++
		out = append(out, c)
	}
	return out
}

func filterCandidates(cs []candidate, includeExt, excludeExt, includePath, excludePath []string) []candidate {
	matchAny := func(patterns []string, value string) bool {
		for _, p := range patterns {
			if ok, _ := filepath.Match(p, value); ok {
				return true
			}
			if !strings.ContainsAny(p, "*?[") && strings.Contains(value, filepath.ToSlash(p)) {
				return true
			}
			// Plain path fragments match at any position; glob patterns match
			// the complete repository-relative path.
			if !strings.ContainsAny(p, "*?[") && strings.Contains(value, filepath.ToSlash(p)) {
				return true
			}
		}
		return false
	}
	out := make([]candidate, 0, len(cs))
	for _, c := range cs {
		ext := strings.ToLower(filepath.Ext(c.path))
		if len(includeExt) > 0 && !matchAny(includeExt, ext) {
			continue
		}
		if matchAny(excludeExt, ext) {
			continue
		}
		path := filepath.ToSlash(c.path)
		if len(includePath) > 0 && !matchAny(includePath, path) {
			continue
		}
		if matchAny(excludePath, path) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func printCandidates(title string, cs []candidate) {
	if len(cs) == 0 {
		return
	}
	fmt.Println(title + ":")
	for i, c := range cs {
		fmt.Printf("%2d. %s (score %d; noise score %d; co-changed in %d commit(s); path bonus %d)", i+1, c.path, c.score, c.noiseScore, len(c.commits), c.nameBonus)
		if len(c.noiseTerms) > 0 {
			fmt.Printf("; noise terms: %s", strings.Join(c.noiseTerms, ", "))
		}
		fmt.Println()
	}
}

func discover(input string) ([]candidate, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	rootOut, err := git(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("not inside a Git repository: %w", err)
	}
	root := strings.TrimSpace(rootOut)

	// Resolve the user-supplied path through Git instead of deciding repository
	// containment with filesystem string comparisons.
	pathOut, err := git(cwd, "ls-files", "--full-name", "--", input)
	if err != nil {
		return nil, err
	}
	target := ""
	for _, p := range strings.Split(strings.TrimSpace(pathOut), "\n") {
		if p != "" {
			target = p
			break
		}
	}
	if target == "" {
		// Also allow a tracked file that has been deleted from the working tree.
		logOut, logErr := git(root, "log", "--follow", "--format=%H", "--", input)
		if logErr != nil {
			return nil, logErr
		}
		if strings.TrimSpace(logOut) == "" {
			return nil, fmt.Errorf("no Git history for path: %s", input)
		}
		if filepath.IsAbs(input) {
			abs, absErr := filepath.Abs(input)
			if absErr != nil {
				return nil, absErr
			}
			target, err = filepath.Rel(root, abs)
			if err != nil {
				return nil, err
			}
		} else {
			target = filepath.ToSlash(filepath.Clean(input))
		}
	}
	target = filepath.ToSlash(target)
	log, err := git(root, "log", "--follow", "--format=%H", "--", target)
	if err != nil {
		return nil, err
	}
	ids := strings.Fields(log)
	cs := map[string]*candidate{}
	for _, id := range ids {
		message, _ := git(root, "show", "-s", "--format=%s%n%b", id)
		files, err := git(root, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", id)
		if err != nil {
			continue
		}
		seen := map[string]bool{}
		for _, p := range strings.Split(files, "\n") {
			p = strings.TrimSpace(p)
			if p == "" || p == target || (!isCode(p) && !isDocument(p)) || seen[p] {
				continue
			}
			seen[p] = true
			if cs[p] == nil {
				cs[p] = &candidate{path: p, commits: map[string]bool{}}
			}
			cs[p].score++
			cs[p].commits[id] = true
			terms := detectNoiseTerms(message)
			cs[p].noiseScore += len(terms)
			for _, term := range terms {
				cs[p].noiseTerms = appendUnique(cs[p].noiseTerms, term)
			}
		}
	}
	for _, c := range cs {
		c.nameBonus = filenameSimilarityBonus(target, c.path)
		c.score += c.nameBonus
		c.testLike = isTestPath(c.path)
		c.kind = "code"
		if isDocument(c.path) {
			c.kind = "document"
		}
	}
	out := make([]candidate, 0, len(cs))
	for _, c := range cs {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].score != out[j].score {
			return out[i].score > out[j].score
		}
		return out[i].path < out[j].path
	})
	return out, nil
}

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
func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	b, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return string(b), nil
}
func fatal(s string) { fmt.Fprintln(os.Stderr, "itoc:", s); os.Exit(1) }
func printJSON(cs []candidate) {
	type row struct {
		Path             string   `json:"path"`
		Score            int      `json:"score"`
		CoChangedCommits int      `json:"co_changed_commits"`
		FilenameBonus    int      `json:"filename_bonus"`
		Kind             string   `json:"kind"`
		NoiseScore       int      `json:"noise_score"`
		NoiseTerms       []string `json:"noise_terms,omitempty"`
	}
	rows := make([]row, 0, len(cs))
	for _, c := range cs {
		kind := c.kind
		if c.testLike {
			kind = "test"
		}
		rows = append(rows, row{c.path, c.score, len(c.commits), c.nameBonus, kind, c.noiseScore, c.noiseTerms})
	}
	if err := json.NewEncoder(os.Stdout).Encode(rows); err != nil {
		fatal(err.Error())
	}
}
