// Package document extracts literal discovery clues from supported document formats.
package document

// Hint describes a literal clue and its location in a document.
type Hint struct {
	Value  string
	Kind   string
	Line   int
	Strong bool
}
