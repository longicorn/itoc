package main

import (
	"fmt"
	"strings"

	"itoc/internal/discovery"
)

type discoveryMethod struct {
	name   string
	create func() discovery.PathLister
}

var discoveryMethods = []discoveryMethod{
	{"git", func() discovery.PathLister { return discovery.GitPathLister{} }},
	{"document", func() discovery.PathLister { return discovery.NewDocumentPathLister("") }},
}

var discoveryRegistry = func() map[string]func() discovery.PathLister {
	registry := make(map[string]func() discovery.PathLister, len(discoveryMethods))
	for _, method := range discoveryMethods {
		registry[method.name] = method.create
	}
	return registry
}()

func parseDiscoveryMethods(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "all" {
		methods := make([]string, 0, len(discoveryRegistry))
		for _, method := range discoveryMethods {
			methods = append(methods, method.name)
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

func discoverWithLister(input string, lister discovery.PathLister) ([]candidate, error) {
	evidence, target, err := lister.List(input)
	if err != nil {
		return nil, err
	}
	return analyzeCandidates(target, evidence), nil
}
