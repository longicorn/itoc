package main

import (
	"flag"
	"fmt"
	"io"
	"os"
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
	out = prepareCandidates(out, candidateOptions{
		includeExt: includeExt, excludeExt: excludeExt,
		includePath: includePath, excludePath: excludePath,
		excludeCandidate: excludeCandidate, minScore: *minScore, limit: *limit,
	})
	if err := writeCandidates(os.Stdout, out, *format); err != nil {
		fatal(err.Error())
	}
}

func fatal(s string) { fmt.Fprintln(os.Stderr, "itoc:", s); os.Exit(1) }

func writeCandidates(w io.Writer, cs []candidate, format string) error {
	data, err := renderCandidates(cs, format)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
