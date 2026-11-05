package main

import (
	"io"
	"itoc/internal/discovery"
)

// expandTestCandidates merges derived test clues without changing existing scores.
func expandTestCandidates(root string, candidates []candidate, ctagsPath string, warnings io.Writer) ([]candidate, error) {
	return expandTestCandidatesWithFollow(root, candidates, ctagsPath, warnings, false)
}

func expandTestCandidatesWithFollow(root string, candidates []candidate, ctagsPath string, warnings io.Writer, follow bool) ([]candidate, error) {
	implementations := make([]string, 0, len(candidates))
	byPath := make(map[string]*candidate, len(candidates))
	for _, item := range candidates {
		c := item
		byPath[c.path] = &c
		if c.kind == "code" && !c.testLike {
			implementations = append(implementations, c.path)
		}
	}
	evidence, err := discovery.DiscoverTestsWithFollow(root, implementations, ctagsPath, warnings, follow)
	if err != nil {
		return nil, err
	}
	for _, item := range evidence {
		c := byPath[item.Path]
		if c == nil {
			terms := discovery.DetectPathNoiseTerms(item.Path)
			c = &candidate{path: item.Path, kind: "code", testLike: true, commits: map[string]bool{}, noiseTerms: terms, noiseScore: discovery.NoiseScoreForTerms(terms)}
			byPath[item.Path] = c
		}
		for _, detail := range item.Details {
			c.evidence = discovery.AppendUnique(c.evidence, detail)
		}
	}
	out := make([]candidate, 0, len(byPath))
	for _, c := range byPath {
		out = append(out, *c)
	}
	return out, nil
}
