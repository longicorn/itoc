// Package ctags collects identifier clues from Universal Ctags JSON output.
package ctags

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
)

// ErrInvalidOutput indicates that Ctags returned malformed or oversized output.
var ErrInvalidOutput = errors.New("invalid ctags output")

// Names returns distinct tag names of at least three bytes.
func Names(binary, path string) (map[string]bool, error) {
	raw, err := exec.Command(binary, "--output-format=json", "--fields=+n", "-o", "-", path).Output()
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		var tag struct {
			Type string `json:"_type"`
			Name string `json:"name"`
		}
		if json.Unmarshal(scanner.Bytes(), &tag) != nil {
			return nil, ErrInvalidOutput
		}
		if tag.Type == "tag" && len(tag.Name) >= 3 {
			names[tag.Name] = true
		}
	}
	if scanner.Err() != nil {
		return nil, ErrInvalidOutput
	}
	return names, nil
}
