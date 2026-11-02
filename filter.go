package main

import (
	"path/filepath"
	"strings"
)

func excludeExactCandidates(cs []candidate, patterns []string) []candidate {
	hidden := map[string]bool{}
	groups := map[string]bool{}
	for _, p := range patterns {
		switch strings.ToLower(p) {
		case "implementation", "document", "test":
			groups[strings.ToLower(p)] = true
		default:
			hidden[filepath.ToSlash(filepath.Clean(p))] = true
		}
	}
	out := make([]candidate, 0, len(cs))
	for _, c := range cs {
		kind := c.kind
		if c.testLike {
			kind = "test"
		}
		if !hidden[filepath.ToSlash(filepath.Clean(c.path))] && !groups[kind] {
			out = append(out, c)
		}
	}
	return out
}
func filterMinScore(cs []candidate, min int) []candidate {
	out := make([]candidate, 0, len(cs))
	for _, c := range cs {
		if c.score >= min {
			out = append(out, c)
		}
	}
	return out
}

func limitCandidates(cs []candidate, limit int) []candidate {
	counts := map[string]int{}
	out := make([]candidate, 0, len(cs))
	for _, c := range cs {
		group := c.kind
		if c.testLike {
			group = "test"
		}
		if counts[group] >= limit {
			continue
		}
		counts[group]++
		out = append(out, c)
	}
	return out
}

func filterCandidates(cs []candidate, includeExt, excludeExt, includePath, excludePath []string) []candidate {
	matchAny := func(patterns []string, value string) bool {
		for _, p := range patterns {
			if ok, _ := filepath.Match(p, value); ok {
				return true
			}
			if !strings.ContainsAny(p, "*?[") && strings.Contains(value, filepath.ToSlash(p)) {
				return true
			}
			// Plain path fragments match at any position; glob patterns match
			// the complete repository-relative path.
			if !strings.ContainsAny(p, "*?[") && strings.Contains(value, filepath.ToSlash(p)) {
				return true
			}
		}
		return false
	}
	out := make([]candidate, 0, len(cs))
	for _, c := range cs {
		ext := strings.ToLower(filepath.Ext(c.path))
		if len(includeExt) > 0 && !matchAny(includeExt, ext) {
			continue
		}
		if matchAny(excludeExt, ext) {
			continue
		}
		path := filepath.ToSlash(c.path)
		if len(includePath) > 0 && !matchAny(includePath, path) {
			continue
		}
		if matchAny(excludePath, path) {
			continue
		}
		out = append(out, c)
	}
	return out
}
