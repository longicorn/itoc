// Package discovery defines the boundary between evidence collection and analysis.
package discovery

// Evidence describes clues associated with a repository-relative path.
type Evidence struct {
	Path       string
	Commits    map[string]bool
	NoiseScore int
	NoiseTerms []string
	Details    []string
	Methods    []string
}

// PathLister collects evidence and resolves the repository-relative target.
type PathLister interface {
	List(input string) ([]Evidence, string, error)
}

// UnsupportedDocumentFormatError indicates that a document format cannot be extracted.
type UnsupportedDocumentFormatError struct{ Extension string }

func (e UnsupportedDocumentFormatError) Error() string {
	return "unsupported document format: " + e.Extension
}
