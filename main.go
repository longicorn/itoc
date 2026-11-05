package main

import "itoc/internal/discovery"

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type stringList []string

func (s *stringList) String() string     { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

type candidateExcludeFlag []string

func (f *candidateExcludeFlag) String() string     { return "" }
func (f *candidateExcludeFlag) Set(v string) error { *f = append(*f, v); return nil }

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, discovery.GitPathLister{}); err != nil {
		fatal(err.Error())
	}
}

func run(args []string, stdout, stderr io.Writer, lister discovery.PathLister) error {
	return runWithInput(args, os.Stdin, stdout, stderr, lister)
}

func runWithInput(args []string, stdin io.Reader, stdout, stderr io.Writer, lister discovery.PathLister) error {
	config, path, err := parseCLI(args, stderr)
	if err != nil {
		return err
	}
	if config.search != "" {
		path, err = searchKeyword(config.search, stdin, stderr)
		if err != nil {
			return err
		}
		if path == "" {
			return nil
		}
	}
	methods, _ := parseDiscoveryMethods(config.discovery)
	var tasks []namedLister
	for _, method := range methods {
		selected := discoveryRegistry[method]()
		if method == "git" {
			if lister != nil {
				selected = lister
			}
			if gitLister, ok := selected.(discovery.GitPathLister); ok {
				ctagsPath, _ := exec.LookPath("ctags")
				selected = gitLister.WithFollow(config.follow).WithRelocation(ctagsPath, stderr)
			}
		}
		tasks = append(tasks, namedLister{name: method, lister: selected})
	}
	var cs []candidate
	if len(tasks) == 1 {
		cs, err = discoverWithLister(path, tasks[0].lister)
	} else {
		cs, err = discoverMany(path, tasks)
	}
	if err != nil {
		return err
	}
	if config.discoverTests {
		cwd, e := os.Getwd()
		if e != nil {
			return e
		}
		root, e := discovery.Git(cwd, "rev-parse", "--show-toplevel")
		if e != nil {
			return e
		}
		ctagsPath, lookupErr := exec.LookPath("ctags")
		if lookupErr != nil {
			fmt.Fprintln(stderr, "itoc: warning: ctags not found; identifier-based test discovery skipped")
		}
		cs, e = expandTestCandidatesWithFollow(strings.TrimSpace(root), cs, ctagsPath, stderr, config.follow)
		if e != nil {
			return e
		}
	}
	cs = prepareCandidates(cs, config.candidateOptions)
	return writeCandidatesWithEvidence(stdout, cs, config.format, config.verbose)
}

type namedLister struct {
	name   string
	lister discovery.PathLister
}

type discoveryResult struct {
	mode       string
	candidates []candidate
	err        error
}

func discoverMany(input string, tasks []namedLister) ([]candidate, error) {
	results := make(chan discoveryResult, len(tasks))
	var wg sync.WaitGroup
	for _, task := range tasks {
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			cs, err := discoverWithLister(input, task.lister)
			results <- discoveryResult{mode: task.name, candidates: cs, err: err}
		}()
	}
	go func() { wg.Wait(); close(results) }()

	var groups [][]candidate
	var errs []string
	for result := range results {
		if result.err != nil {
			var unsupported discovery.UnsupportedDocumentFormatError
			if result.mode == "document" && errors.As(result.err, &unsupported) {
				continue
			}
			errs = append(errs, result.mode+": "+result.err.Error())
			continue
		}
		groups = append(groups, result.candidates)
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("discovery failed (%s)", strings.Join(errs, "; "))
	}
	return mergeDiscoveryCandidates(groups...), nil
}

func mergeDiscoveryCandidates(groups ...[]candidate) []candidate {
	byPath := map[string]*candidate{}
	for _, candidates := range groups {
		for _, item := range candidates {
			if byPath[item.path] == nil {
				copy := item
				byPath[item.path] = &copy
				continue
			}
			current := byPath[item.path]
			previousCount := len(current.discoveries)
			current.commits = mergeCommitSets(current.commits, item.commits)
			for _, method := range item.discoveries {
				current.discoveries = discovery.AppendUnique(current.discoveries, method)
			}
			for _, evidence := range item.evidence {
				current.evidence = discovery.AppendUnique(current.evidence, evidence)
			}
			for _, term := range item.noiseTerms {
				current.noiseTerms = discovery.AppendUnique(current.noiseTerms, term)
			}
			if item.noiseScore > current.noiseScore {
				current.noiseScore = item.noiseScore
			}
			if item.score > current.score {
				current.score = item.score
			}
			if len(current.discoveries) > previousCount {
				current.score += len(current.discoveries) - previousCount
				current.evidence = discovery.AppendUnique(current.evidence, fmt.Sprintf("independently found by %s discovery methods (+%d agreement bonus)", strings.Join(current.discoveries, ", "), len(current.discoveries)-previousCount))
			}
		}
	}
	out := make([]candidate, 0, len(byPath))
	for _, item := range byPath {
		out = append(out, *item)
	}
	return out
}

func mergeCommitSets(a, b map[string]bool) map[string]bool {
	out := make(map[string]bool, len(a)+len(b))
	for id, present := range a {
		out[id] = present
	}
	for id, present := range b {
		out[id] = present
	}
	return out
}

func fatal(s string) { fmt.Fprintln(os.Stderr, "itoc:", s); os.Exit(1) }

func writeCandidates(w io.Writer, cs []candidate, format string) error {
	return writeCandidatesWithEvidence(w, cs, format, false)
}

func writeCandidatesWithEvidence(w io.Writer, cs []candidate, format string, showEvidence bool) error {
	data, err := renderCandidatesWithEvidence(cs, format, showEvidence)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
