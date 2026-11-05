package discovery

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type GitRunner interface {
	Run(dir string, args ...string) (string, error)
}

type execGitRunner struct{}

func (execGitRunner) Run(dir string, args ...string) (string, error) { return Git(dir, args...) }

func NewGitPathLister(runner GitRunner) GitPathLister { return GitPathLister{runner: runner} }

type GitPathLister struct{ runner GitRunner }

func (l GitPathLister) run(dir string, args ...string) (string, error) {
	if l.runner == nil {
		return Git(dir, args...)
	}
	return l.runner.Run(dir, args...)
}

func (l GitPathLister) List(input string) ([]Evidence, string, error) {
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
	messages := map[string]string{}
	if len(ids) > 0 {
		args := append([]string{"show", "-s", "--format=%H%x00%s%n%b%x00"}, ids...)
		if output, err := l.run(root, args...); err == nil {
			parts := strings.Split(output, "\x00")
			for i := 0; i+1 < len(parts); i += 2 {
				messages[strings.TrimSpace(parts[i])] = parts[i+1]
			}
		}
	}
	commits := make([]commitSnapshot, len(ids))
	valid := make([]bool, len(ids))
	jobs := make(chan int)
	var workers sync.WaitGroup
	count := 8
	if len(ids) < count {
		count = len(ids)
	}
	for i := 0; i < count; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := range jobs {
				files, err := l.run(root, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", ids[i])
				if err == nil {
					commits[i] = commitSnapshot{id: ids[i], message: messages[ids[i]], files: files}
					valid[i] = true
				}
			}
		}()
	}
	for i := range ids {
		jobs <- i
	}
	close(jobs)
	workers.Wait()
	ordered := commits[:0]
	for i, commit := range commits {
		if valid[i] {
			ordered = append(ordered, commit)
		}
	}
	commits = ordered
	cs := collectCochangeEvidence(target, commits)
	evidence := make([]Evidence, 0, len(cs))
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

func Git(dir string, args ...string) (string, error) {
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

func collectCochangeEvidence(target string, commits []commitSnapshot) map[string]*Evidence {
	cs := map[string]*Evidence{}
	for _, commit := range commits {
		terms := DetectNoiseTerms(commit.message)
		seen := map[string]bool{}
		for _, p := range strings.Split(commit.files, "\n") {
			p = strings.TrimSpace(p)
			if p == "" || p == target || (!IsCode(p) && !IsDocument(p)) || seen[p] {
				continue
			}
			seen[p] = true
			if cs[p] == nil {
				cs[p] = &Evidence{Path: p, Commits: map[string]bool{}, Methods: []string{"git"}}
			}
			cs[p].Commits[commit.id] = true
			cs[p].Details = AppendUnique(cs[p].Details, "co-changed in commit "+commit.id)
			allTerms := append(append([]string(nil), terms...), DetectPathNoiseTerms(p)...)
			if score := NoiseScoreForTerms(terms); score > cs[p].NoiseScore {
				cs[p].NoiseScore = score
			}
			if score := NoiseScoreForTerms(DetectPathNoiseTerms(p)); score > cs[p].NoiseScore {
				cs[p].NoiseScore = score
			}
			for _, term := range allTerms {
				cs[p].NoiseTerms = AppendUnique(cs[p].NoiseTerms, term)
			}
		}
	}
	return cs
}

func DetectPathNoiseTerms(path string) []string {
	var terms []string
	for _, part := range strings.Split(strings.ToLower(filepath.ToSlash(path)), "/") {
		name := strings.TrimSuffix(part, filepath.Ext(part))
		for _, term := range []string{"benchmark", "benchmarks", "bench", "perf", "performance", "profiling", "profile"} {
			if name == term || strings.HasPrefix(name, term+"_") || strings.HasSuffix(name, "_"+term) || strings.Contains(name, "_"+term+"_") {
				terms = AppendUnique(terms, "path:"+term)
			}
		}
	}
	return terms
}

func NoiseScoreForTerms(terms []string) int {
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
