package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func searchKeyword(keyword string, in io.Reader, diagnostic io.Writer) (string, error) {
	rootBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("keyword search requires a Git repository: %w", err)
	}
	root := strings.TrimSpace(string(rootBytes))
	cmd := exec.Command("git", "ls-files", "-z", "--cached")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("keyword search failed: %w", err)
	}
	var paths []string
	seen := map[string]bool{}
	for _, path := range strings.Split(string(output), "\x00") {
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("keyword search %s: %w", path, err)
		}
		if bytes.Contains(data, []byte(keyword)) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		_, err = fmt.Fprintln(diagnostic, "No matching files found.")
		return "", err
	}

	index := 0
	if len(paths) > 1 {
		for i, path := range paths {
			if _, err := fmt.Fprintf(diagnostic, "%d. %s\n", i+1, path); err != nil {
				return "", err
			}
		}
		if _, err := fmt.Fprint(diagnostic, "Select a file number: "); err != nil {
			return "", err
		}
		line, err := bufio.NewReader(in).ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		choice, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || choice < 1 || choice > len(paths) {
			return "", fmt.Errorf("invalid file selection")
		}
		index = choice - 1
	}
	// Existing discovery resolves input paths relative to the working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Rel(cwd, filepath.Join(root, filepath.FromSlash(paths[index])))
}
