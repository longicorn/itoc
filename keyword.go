package main

import (
	"encoding/json"
	"fmt"
	"io"
)

// searchKeyword is a placeholder for finding starting paths from a keyword.
func searchKeyword(keyword, format string, out io.Writer) error {
	if format == "json" {
		return json.NewEncoder(out).Encode(struct {
			Status string `json:"status"`
		}{Status: "not_implemented"})
	}
	_, err := fmt.Fprintln(out, "Keyword search is not implemented yet.")
	return err
}
