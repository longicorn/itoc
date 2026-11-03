package main

import "itoc/internal/discovery"

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

type cochangeEvidence = discovery.Evidence

type candidateOptions struct {
	includeExt, excludeExt   []string
	includePath, excludePath []string
	excludeCandidate         []string
	minScore, limit          int
	maxNoiseScore            int
	showNoisy                bool
}
