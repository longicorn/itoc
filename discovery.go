package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func discover(input string) ([]candidate, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	rootOut, err := git(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("not inside a Git repository: %w", err)
	}
	root := strings.TrimSpace(rootOut)

	// Resolve the user-supplied path through Git instead of deciding repository
	// containment with filesystem string comparisons.
	pathOut, err := git(cwd, "ls-files", "--full-name", "--", input)
	if err != nil {
		return nil, err
	}
	target := ""
	for _, p := range strings.Split(strings.TrimSpace(pathOut), "\n") {
		if p != "" {
			target = p
			break
		}
	}
	if target == "" {
		// Also allow a tracked file that has been deleted from the working tree.
		logOut, logErr := git(root, "log", "--follow", "--format=%H", "--", input)
		if logErr != nil {
			return nil, logErr
		}
		if strings.TrimSpace(logOut) == "" {
			return nil, fmt.Errorf("no Git history for path: %s", input)
		}
		if filepath.IsAbs(input) {
			abs, absErr := filepath.Abs(input)
			if absErr != nil {
				return nil, absErr
			}
			target, err = filepath.Rel(root, abs)
			if err != nil {
				return nil, err
			}
		} else {
			target = filepath.ToSlash(filepath.Clean(input))
		}
	}
	target = filepath.ToSlash(target)
	log, err := git(root, "log", "--follow", "--format=%H", "--", target)
	if err != nil {
		return nil, err
	}
	ids := strings.Fields(log)
	cs := map[string]*candidate{}
	for _, id := range ids {
		message, _ := git(root, "show", "-s", "--format=%s%n%b", id)
		files, err := git(root, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", id)
		if err != nil {
			continue
		}
		seen := map[string]bool{}
		for _, p := range strings.Split(files, "\n") {
			p = strings.TrimSpace(p)
			if p == "" || p == target || (!isCode(p) && !isDocument(p)) || seen[p] {
				continue
			}
			seen[p] = true
			if cs[p] == nil {
				cs[p] = &candidate{path: p, commits: map[string]bool{}}
			}
			cs[p].score++
			cs[p].commits[id] = true
			terms := detectNoiseTerms(message)
			cs[p].noiseScore += len(terms)
			for _, term := range terms {
				cs[p].noiseTerms = appendUnique(cs[p].noiseTerms, term)
			}
		}
	}
	for _, c := range cs {
		c.nameBonus = filenameSimilarityBonus(target, c.path)
		c.score += c.nameBonus
		c.testLike = isTestPath(c.path)
		c.kind = "code"
		if isDocument(c.path) {
			c.kind = "document"
		}
	}
	out := make([]candidate, 0, len(cs))
	for _, c := range cs {
		out = append(out, *c)
	}

	return out, nil
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
