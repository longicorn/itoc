package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitRunner interface {
	Run(dir string, args ...string) (string, error)
}

type execGitRunner struct{}

func (execGitRunner) Run(dir string, args ...string) (string, error) { return git(dir, args...) }

type GitPathLister struct{ runner GitRunner }

func (l GitPathLister) run(dir string, args ...string) (string, error) {
	if l.runner == nil {
		return git(dir, args...)
	}
	return l.runner.Run(dir, args...)
}

func (l GitPathLister) List(input string) ([]cochangeEvidence, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}
	rootOut, err := l.run(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, "", fmt.Errorf("not inside a Git repository: %w", err)
	}
	root := strings.TrimSpace(rootOut)

	target, err := resolveGitTarget(cwd, root, input, l)
	if err != nil {
		return nil, "", err
	}

	log, err := l.run(root, "log", "--follow", "--format=%H", "--", target)
	if err != nil {
		return nil, "", err
	}
	ids := strings.Fields(log)
	commits := make([]commitSnapshot, 0, len(ids))
	for _, id := range ids {
		message, _ := l.run(root, "show", "-s", "--format=%s%n%b", id)
		files, err := l.run(root, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", id)
		if err != nil {
			continue
		}
		commits = append(commits, commitSnapshot{id: id, message: message, files: files})
	}
	cs := collectCochangeEvidence(target, commits)
	evidence := make([]cochangeEvidence, 0, len(cs))
	for _, item := range cs {
		evidence = append(evidence, *item)
	}
	return evidence, target, nil
}

func resolveGitTarget(cwd, root, input string, l GitPathLister) (string, error) {
	pathOut, err := l.run(cwd, "ls-files", "--full-name", "--", input)
	if err != nil {
		return "", err
	}
	target := ""
	for _, p := range strings.Split(strings.TrimSpace(pathOut), "\n") {
		if p != "" {
			target = p
			break
		}
	}
	if target == "" {
		return "", fmt.Errorf("path is not tracked in the Git working tree: %s", input)
	}
	return filepath.ToSlash(target), nil
}

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	b, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return string(b), nil
}

type commitSnapshot struct{ id, message, files string }

func collectCochangeEvidence(target string, commits []commitSnapshot) map[string]*cochangeEvidence {
	cs := map[string]*cochangeEvidence{}
	for _, commit := range commits {
		terms := detectNoiseTerms(commit.message)
		seen := map[string]bool{}
		for _, p := range strings.Split(commit.files, "\n") {
			p = strings.TrimSpace(p)
			if p == "" || p == target || (!isCode(p) && !isDocument(p)) || seen[p] {
				continue
			}
			seen[p] = true
			if cs[p] == nil {
				cs[p] = &cochangeEvidence{Path: p, Commits: map[string]bool{}, Methods: []string{"git"}}
			}
			cs[p].Commits[commit.id] = true
			cs[p].Details = appendUnique(cs[p].Details, "co-changed in commit "+commit.id)
			allTerms := append(append([]string(nil), terms...), detectPathNoiseTerms(p)...)
			if score := noiseScoreForTerms(terms); score > cs[p].NoiseScore {
				cs[p].NoiseScore = score
			}
			if score := noiseScoreForTerms(detectPathNoiseTerms(p)); score > cs[p].NoiseScore {
				cs[p].NoiseScore = score
			}
			for _, term := range allTerms {
				cs[p].NoiseTerms = appendUnique(cs[p].NoiseTerms, term)
			}
		}
	}
	return cs
}

func detectPathNoiseTerms(path string) []string {
	var terms []string
	for _, part := range strings.Split(strings.ToLower(filepath.ToSlash(path)), "/") {
		name := strings.TrimSuffix(part, filepath.Ext(part))
		for _, term := range []string{"benchmark", "benchmarks", "bench", "perf", "performance", "profiling", "profile"} {
			if name == term || strings.HasPrefix(name, term+"_") || strings.HasSuffix(name, "_"+term) || strings.Contains(name, "_"+term+"_") {
				terms = appendUnique(terms, "path:"+term)
			}
		}
	}
	return terms
}

func noiseScoreForTerms(terms []string) int {
	score := 0
	for _, term := range terms {
		if strings.HasPrefix(term, "path:benchmark") || strings.HasPrefix(term, "path:bench") || strings.HasPrefix(term, "path:perf") || strings.HasPrefix(term, "path:performance") {
			score += 10
		} else {
			score++
		}
	}
	return score
}
