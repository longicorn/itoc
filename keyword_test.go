package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestKeywordPlaceholder(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--search", "feature"}, "not implemented"},
		{[]string{"--format", "json", "--search", "feature"}, `"not_implemented"`},
	} {
		var out, stderr bytes.Buffer
		if err := run(tc.args, &out, &stderr, nil); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), tc.want) {
			t.Fatalf("output %q does not contain %q", out.String(), tc.want)
		}
	}
	for _, args := range [][]string{{"--search", ""}, {"--search", "feature", "file.go"}} {
		if _, _, err := parseCLI(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("expected error for %q", args)
		}
	}
}
