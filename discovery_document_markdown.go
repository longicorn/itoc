package main

import (
	"regexp"
	"strings"
)

var (
	mdLinkRE     = regexp.MustCompile(`!?\[[^\]]*\]\(([^)]+)\)`)
	mdRefRE      = regexp.MustCompile(`(?m)^\s*\[[^\]]+\]:\s*(\S+)`)
	mdCodeRE     = regexp.MustCompile("`+([^`\n]+)`+")
	mdEmphasisRE = regexp.MustCompile(`(?:\*\*|__)([^*_\n]+)(?:\*\*|__)|(?:\*|_)([^*_\n]+)(?:\*|_)`)
)

func extractMarkdownHints(content string) []documentHint {
	var out []documentHint
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lineNo := i + 1
		for _, m := range mdLinkRE.FindAllStringSubmatch(line, -1) {
			out = append(out, documentHint{m[1], "markdown link", lineNo, false})
		}
		for _, m := range mdRefRE.FindAllStringSubmatch(line, -1) {
			out = append(out, documentHint{m[1], "markdown reference link", lineNo, false})
		}
		for _, m := range mdCodeRE.FindAllStringSubmatch(line, -1) {
			out = append(out, documentHint{m[1], "markdown inline code", lineNo, true})
		}
		for _, m := range mdEmphasisRE.FindAllStringSubmatch(line, -1) {
			v := m[1]
			if v == "" {
				v = m[2]
			}
			if v != "" {
				out = append(out, documentHint{v, "markdown emphasis", lineNo, true})
			}
		}
	}
	return out
}
