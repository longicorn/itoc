package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func searchKeyword(keyword string, in io.Reader, diagnostic io.Writer) (string, error) {
	rootBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("keyword search requires a Git repository: %w", err)
	}
	root := strings.TrimSpace(string(rootBytes))
	cmd := exec.Command("git", "grep", "-l", "-z", "-F", "-e", keyword, "--")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			_, err = fmt.Fprintln(diagnostic, "No matching files found.")
			return "", err
		}
		return "", fmt.Errorf("keyword search failed: %w", err)
	}
	paths := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
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
