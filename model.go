package main

type candidate struct {
	path        string
	score       int
	nameBonus   int
	testLike    bool
	kind        string
	noiseScore  int
	noiseTerms  []string
	commits     map[string]bool
	evidence    []string
	discoveries []string
}

type cochangeEvidence struct {
	path        string
	commits     map[string]bool
	noiseScore  int
	noiseTerms  []string
	evidence    []string
	discoveries []string
}

type candidateOptions struct {
	includeExt, excludeExt   []string
	includePath, excludePath []string
	excludeCandidate         []string
	minScore, limit          int
	maxNoiseScore            int
	showNoisy                bool
}
