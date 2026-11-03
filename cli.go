package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

type cliConfig struct {
	search        string
	format        string
	discovery     string
	verbose       bool
	discoverTests bool
	candidateOptions
}

func parseCLI(args []string, output io.Writer) (cliConfig, string, error) {
	var config cliConfig
	fs := flag.NewFlagSet("itoc", flag.ContinueOnError)
	fs.SetOutput(output)
	fs.StringVar(&config.search, "search", "", "keyword search placeholder")
	fs.StringVar(&config.format, "format", "text", "output format: text or json")
	fs.StringVar(&config.discovery, "discovery", "git", "comma-separated discovery methods or all")
	fs.BoolVar(&config.discoverTests, "discover-tests", true, "find test candidates from implementation candidates")
	fs.BoolVar(&config.verbose, "verbose", false, "show evidence details")
	fs.IntVar(&config.maxNoiseScore, "max-noise-score", 1, "hide candidates above this noise score")
	fs.BoolVar(&config.showNoisy, "show-noisy", false, "include candidates hidden by the noise threshold")
	fs.IntVar(&config.limit, "limit", 10, "maximum candidates per group (0 means unlimited)")
	fs.IntVar(&config.minScore, "min-score", 0, "minimum candidate score (0 disables score filtering)")
	fs.Var((*stringList)(&config.includeExt), "include-ext", "candidate extension allowlist (repeatable, e.g. .go)")
	fs.Var((*stringList)(&config.excludeExt), "exclude-ext", "candidate extension denylist (repeatable, e.g. .md)")
	fs.Var((*stringList)(&config.includePath), "include-path", "candidate repository-relative path substring or glob allowlist (repeatable)")
	fs.Var((*stringList)(&config.excludePath), "exclude-path", "candidate repository-relative path glob denylist (repeatable)")
	fs.Var((*candidateExcludeFlag)(&config.excludeCandidate), "exclude-candidate", "hide a candidate group or exact repository-relative path (repeatable)")
	fs.Usage = func() { printUsage(output) }
	if err := fs.Parse(args); err != nil {
		return config, "", err
	}
	searchSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "search" {
			searchSet = true
		}
	})
	if searchSet && config.search == "" {
		return config, "", errors.New("--search requires a non-empty keyword")
	}
	if searchSet && fs.NArg() != 0 {
		return config, "", errors.New("--search cannot be combined with a repository file")
	}
	if !searchSet && fs.NArg() != 1 {
		return config, "", errors.New("expected exactly one repository file")
	}
	if config.format != "text" && config.format != "json" {
		return config, "", fmt.Errorf("unsupported format: %s", config.format)
	}
	if _, err := parseDiscoveryMethods(config.discovery); err != nil {
		return config, "", err
	}
	if config.limit < 0 {
		return config, "", errors.New("limit must be zero or greater")
	}
	if config.minScore < 0 {
		return config, "", errors.New("min-score must be zero or greater")
	}
	if config.maxNoiseScore < 0 {
		return config, "", errors.New("max-noise-score must be zero or greater")
	}
	return config, fs.Arg(0), nil
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage: itoc [options] <repository-file>")
	fmt.Fprintln(out, "       itoc [options] --search <keyword>")
	fmt.Fprintln(out, "\nOptions:")
	fmt.Fprintln(out, "  --search KEYWORD            Keyword search placeholder")
	fmt.Fprintln(out, "  --format text|json          Output format (default: text)")
	fmt.Fprintln(out, "  --discovery METHODS         Comma-separated methods or all (default: git)")
	fmt.Fprintln(out, "  --discover-tests=false      Disable expansion of implementation candidates to tests")
	fmt.Fprintln(out, "  --verbose                   Show evidence details")
	fmt.Fprintln(out, "  --max-noise-score N         Hide candidates above noise score (default: 1)")
	fmt.Fprintln(out, "  --show-noisy                Include candidates hidden by noise rules")
	fmt.Fprintln(out, "  --limit N                   Maximum candidates per group (default: 10; 0: unlimited)")
	fmt.Fprintln(out, "  --min-score N               Minimum candidate score (default: 0; disabled)")
	fmt.Fprintln(out, "  --exclude-candidate PATH    Hide a group (implementation/document/test) or exact repository-relative path (repeatable)")
	fmt.Fprintln(out, "  --include-ext EXT           Include candidates with this extension (repeatable)")
	fmt.Fprintln(out, "  --exclude-ext EXT           Exclude candidates with this extension (repeatable)")
	fmt.Fprintln(out, "  --include-path PATTERN      Include paths containing this string or matching glob (repeatable)")
	fmt.Fprintln(out, "  --exclude-path GLOB         Exclude matching repository-relative paths (repeatable)")
	fmt.Fprintln(out, "\nPlain path values match substrings; glob patterns match the full path. Globs use filepath.Match syntax; '*' does not cross directories.")
	fmt.Fprintln(out, "Examples: --include-ext .go --exclude-path 'vendor/*'")
}
