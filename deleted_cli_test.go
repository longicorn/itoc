package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeletedInputRejectsDocumentDiscovery(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.md")
	var out, stderr bytes.Buffer
	err := run([]string{"--allow-deleted", "--discovery=git,document", missing}, &out, &stderr, nil)
	if err == nil || !strings.Contains(err.Error(), "document discovery cannot use a deleted input") {
		t.Fatalf("error = %v", err)
	}
}
