package main

import "itoc/internal/discovery"

import (
	"testing"
)

type fixedPathLister struct {
	evidence []discovery.Evidence
	target   string
}

func (f fixedPathLister) List(string) ([]discovery.Evidence, string, error) {
	return f.evidence, f.target, nil
}

func TestPathListerInterfaceFeedsAnalysis(t *testing.T) {
	lister := fixedPathLister{target: "docs/alpha.md", evidence: []discovery.Evidence{{Path: "src/alpha.go", Commits: map[string]bool{"c": true}}}}
	evidence, target, err := lister.List("ignored")
	if err != nil {
		t.Fatal(err)
	}
	got := analyzeCandidates(target, evidence)
	if len(got) != 1 || got[0].score != 3 {
		t.Fatalf("got %#v", got)
	}
}
func TestDiscoverWithListerRunsFullDiscoveryFlow(t *testing.T) {
	lister := fixedPathLister{target: "docs/alpha.md", evidence: []discovery.Evidence{
		{Path: "src/alpha.go", Commits: map[string]bool{"c": true}},
	}}
	got, err := discoverWithLister("ignored", lister)
	if err != nil {
		t.Fatal(err)
	}
	got = prepareCandidates(got, candidateOptions{limit: 1})
	if len(got) != 1 || got[0].path != "src/alpha.go" || got[0].score != 3 {
		t.Fatalf("discovery flow returned %#v", got)
	}
}
