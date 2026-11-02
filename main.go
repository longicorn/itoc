package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type stringList []string

func (s *stringList) String() string     { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

type candidateExcludeFlag []string

func (f *candidateExcludeFlag) String() string     { return "" }
func (f *candidateExcludeFlag) Set(v string) error { *f = append(*f, v); return nil }

type candidate struct {
	path       string
	score      int
	nameBonus  int
	testLike   bool
	kind       string
	noiseScore int
	noiseTerms []string
	commits    map[string]bool
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, GitPathLister{}); err != nil {
		fatal(err.Error())
	}
}

func run(args []string, stdout, stderr io.Writer, lister PathLister) error {
	config, path, err := parseCLI(args, stderr)
	if err != nil {
		return err
	}
	cs, err := discoverWithLister(path, lister)
	if err != nil {
		return err
	}
	cs = prepareCandidates(cs, config.candidateOptions)
	return writeCandidates(stdout, cs, config.format)
}

func fatal(s string) { fmt.Fprintln(os.Stderr, "itoc:", s); os.Exit(1) }

func writeCandidates(w io.Writer, cs []candidate, format string) error {
	data, err := renderCandidates(cs, format)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
