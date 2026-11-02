package main

import (
	"fmt"
	"io"
	"os"
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
	if err := run(os.Args[1:], os.Stdout, os.Stderr, GitPathLister{}); err != nil {
		fatal(err.Error())
	}
}

func run(args []string, stdout, stderr io.Writer, lister PathLister) error {
	config, path, err := parseCLI(args, stderr)
	if err != nil {
		return err
	}
	var cs []candidate
	switch config.discovery {
	case "document":
		cs, err = discoverWithLister(path, NewDocumentPathLister(""))
	case "both":
		gitLister := lister
		if gitLister == nil {
			gitLister = GitPathLister{}
		}
		cs, err = discoverBoth(path, gitLister, NewDocumentPathLister(""))
	default:
		if lister == nil {
			lister = GitPathLister{}
		}
		cs, err = discoverWithLister(path, lister)
	}
	if err != nil {
		return err
	}
	cs = prepareCandidates(cs, config.candidateOptions)
	return writeCandidatesWithEvidence(stdout, cs, config.format, config.verbose)
}

type discoveryResult struct {
	mode       string
	candidates []candidate
	err        error
}

func discoverBoth(input string, gitLister, documentLister PathLister) ([]candidate, error) {
	results := make(chan discoveryResult, 2)
	var wg sync.WaitGroup
	for _, task := range []struct {
		mode   string
		lister PathLister
	}{{"git", gitLister}, {"document", documentLister}} {
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			cs, err := discoverWithLister(input, task.lister)
			results <- discoveryResult{mode: task.mode, candidates: cs, err: err}
		}()
	}
	go func() { wg.Wait(); close(results) }()

	var groups [][]candidate
	var errs []string
	for result := range results {
		if result.err != nil {
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
			current.commits = mergeCommitSets(current.commits, item.commits)
			for _, evidence := range item.evidence {
				current.evidence = appendUnique(current.evidence, evidence)
			}
			for _, term := range item.noiseTerms {
				current.noiseTerms = appendUnique(current.noiseTerms, term)
			}
			if item.noiseScore > current.noiseScore {
				current.noiseScore = item.noiseScore
			}
			if item.score > current.score {
				current.score = item.score
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
