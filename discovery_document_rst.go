package main

import (
	"regexp"
	"strings"
)

var (
	rstLinkRE    = regexp.MustCompile("`[^`<>]+\\s*<([^>]+)>`_")
	rstRoleRE    = regexp.MustCompile("(?i):(?:file|doc|ref|func|class|mod|meth):`([^`]+)`")
	rstLiteralRE = regexp.MustCompile("``([^`]+)``")
)

func extractRSTHints(content string) []documentHint {
	var out []documentHint
	for i, line := range strings.Split(content, "\n") {
		lineNo := i + 1
		for _, m := range rstLinkRE.FindAllStringSubmatch(line, -1) {
			out = append(out, documentHint{m[1], "reStructuredText link", lineNo, false})
		}
		for _, m := range rstRoleRE.FindAllStringSubmatch(line, -1) {
			out = append(out, documentHint{m[1], "reStructuredText role", lineNo, true})
		}
		for _, m := range rstLiteralRE.FindAllStringSubmatch(line, -1) {
			out = append(out, documentHint{m[1], "reStructuredText literal", lineNo, true})
		}
	}
	return out
}
