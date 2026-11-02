package main

import (
	"fmt"
	"strings"
)

type PathLister interface {
	List(input string) ([]cochangeEvidence, string, error)
}

type discoveryMode string

const (
	discoveryGit      discoveryMode = "git"
	discoveryDocument discoveryMode = "document"
)

var discoveryRegistry = map[string]func() PathLister{
	"git":      func() PathLister { return GitPathLister{} },
	"document": func() PathLister { return NewDocumentPathLister("") },
}

var discoveryMethodOrder = []string{"git", "document"}

func parseDiscoveryMethods(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "all" {
		methods := make([]string, 0, len(discoveryRegistry))
		for _, name := range discoveryMethodOrder {
			if _, ok := discoveryRegistry[name]; ok {
				methods = append(methods, name)
			}
		}
		return methods, nil
	}
	if value == "" {
		return nil, fmt.Errorf("discovery method cannot be empty")
	}
	seen := map[string]bool{}
	var methods []string
	for _, part := range strings.Split(value, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			return nil, fmt.Errorf("invalid discovery list %q: empty method", value)
		}
		if _, ok := discoveryRegistry[name]; !ok {
			return nil, fmt.Errorf("unsupported discovery method: %s", name)
		}
		if !seen[name] {
			methods = append(methods, name)
			seen[name] = true
		}
	}
	return methods, nil
}

func discover(input string) ([]candidate, error) {
	return discoverWithLister(input, GitPathLister{})
}

func discoverWithLister(input string, lister PathLister) ([]candidate, error) {
	evidence, target, err := lister.List(input)
	if err != nil {
		return nil, err
	}
	return analyzeCandidates(target, evidence), nil
}
