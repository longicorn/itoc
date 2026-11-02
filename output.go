package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

func printCandidatesTo(out *bytes.Buffer, title string, cs []candidate, showEvidence bool) {
	if len(cs) == 0 {
		return
	}
	fmt.Fprintln(out, title+":")
	for i, c := range cs {
		fmt.Fprintf(out, "%2d. %s (score %d; noise score %d; co-changed in %d commit(s); path bonus %d)", i+1, c.path, c.score, c.noiseScore, len(c.commits), c.nameBonus)
		if len(c.noiseTerms) > 0 {
			fmt.Fprintf(out, "; noise terms: %s", strings.Join(c.noiseTerms, ", "))
		}
		if showEvidence && len(c.evidence) > 0 {
			fmt.Fprintf(out, "; evidence: %s", strings.Join(c.evidence, " | "))
		}
		fmt.Fprintln(out)
	}
}

func renderCandidates(cs []candidate, format string) ([]byte, error) {
	return renderCandidatesWithEvidence(cs, format, false)
}

func renderCandidatesWithEvidence(cs []candidate, format string, showEvidence bool) ([]byte, error) {
	if format == "json" {
		return renderJSONWithEvidence(cs, showEvidence)
	}
	var out bytes.Buffer
	if len(cs) == 0 {
		fmt.Fprintln(&out, "No code candidates found.")
		return out.Bytes(), nil
	}
	var impl, tests, docs []candidate
	for _, c := range cs {
		if c.kind == "document" {
			docs = append(docs, c)
		} else if c.testLike {
			tests = append(tests, c)
		} else {
			impl = append(impl, c)
		}
	}
	printCandidatesTo(&out, "Implementation candidates", impl, showEvidence)
	printCandidatesTo(&out, "Test candidates", tests, showEvidence)
	printCandidatesTo(&out, "Document candidates", docs, showEvidence)
	return out.Bytes(), nil
}

func renderJSON(cs []candidate) ([]byte, error) {
	return renderJSONWithEvidence(cs, false)
}

func renderJSONWithEvidence(cs []candidate, showEvidence bool) ([]byte, error) {
	type row struct {
		Path             string   `json:"path"`
		Score            int      `json:"score"`
		CoChangedCommits int      `json:"co_changed_commits"`
		FilenameBonus    int      `json:"filename_bonus"`
		Kind             string   `json:"kind"`
		NoiseScore       int      `json:"noise_score"`
		NoiseTerms       []string `json:"noise_terms,omitempty"`
		Evidence         []string `json:"evidence,omitempty"`
	}
	rows := make([]row, 0, len(cs))
	for _, c := range cs {
		kind := c.kind
		if c.testLike {
			kind = "test"
		}
		evidence := []string(nil)
		if showEvidence {
			evidence = c.evidence
		}
		rows = append(rows, row{c.path, c.score, len(c.commits), c.nameBonus, kind, c.noiseScore, c.noiseTerms, evidence})
	}
	var out bytes.Buffer
	if err := json.NewEncoder(&out).Encode(rows); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
