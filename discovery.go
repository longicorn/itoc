package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PathLister interface {
	List(input string) ([]cochangeEvidence, string, error)
}

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
		logOut, logErr := l.run(root, "log", "--follow", "--format=%H", "--", input)
		if logErr != nil {
			return "", logErr
		}
		if strings.TrimSpace(logOut) == "" {
			return "", fmt.Errorf("no Git history for path: %s", input)
		}
		if filepath.IsAbs(input) {
			abs, absErr := filepath.Abs(input)
			if absErr != nil {
				return "", absErr
			}
			target, err = filepath.Rel(root, abs)
			if err != nil {
				return "", err
			}
		} else {
			target = filepath.ToSlash(filepath.Clean(input))
		}
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
				cs[p] = &cochangeEvidence{path: p, commits: map[string]bool{}}
			}
			cs[p].commits[commit.id] = true
			cs[p].noiseScore += len(terms)
			for _, term := range terms {
				cs[p].noiseTerms = appendUnique(cs[p].noiseTerms, term)
			}
		}
	}
	return cs
}
