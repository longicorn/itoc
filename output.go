package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func printCandidates(title string, cs []candidate) {
	if len(cs) == 0 {
		return
	}
	fmt.Println(title + ":")
	for i, c := range cs {
		fmt.Printf("%2d. %s (score %d; noise score %d; co-changed in %d commit(s); path bonus %d)", i+1, c.path, c.score, c.noiseScore, len(c.commits), c.nameBonus)
		if len(c.noiseTerms) > 0 {
			fmt.Printf("; noise terms: %s", strings.Join(c.noiseTerms, ", "))
		}
		fmt.Println()
	}
}

func printJSON(cs []candidate) {
	type row struct {
		Path             string   `json:"path"`
		Score            int      `json:"score"`
		CoChangedCommits int      `json:"co_changed_commits"`
		FilenameBonus    int      `json:"filename_bonus"`
		Kind             string   `json:"kind"`
		NoiseScore       int      `json:"noise_score"`
		NoiseTerms       []string `json:"noise_terms,omitempty"`
	}
	rows := make([]row, 0, len(cs))
	for _, c := range cs {
		kind := c.kind
		if c.testLike {
			kind = "test"
		}
		rows = append(rows, row{c.path, c.score, len(c.commits), c.nameBonus, kind, c.noiseScore, c.noiseTerms})
	}
	if err := json.NewEncoder(os.Stdout).Encode(rows); err != nil {
		fatal(err.Error())
	}
}
