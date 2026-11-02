package main

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

type cochangeEvidence struct {
	path       string
	commits    map[string]bool
	noiseScore int
	noiseTerms []string
}

type candidateOptions struct {
	includeExt, excludeExt   []string
	includePath, excludePath []string
	excludeCandidate         []string
	minScore, limit          int
}
