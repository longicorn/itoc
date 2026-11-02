package main

import (
	"flag"
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
	config, path, err := parseCLI(os.Args[1:], flag.CommandLine.Output())
	if err != nil {
		fatal(err.Error())
	}
	if path == "" {
		flag.Usage()
		os.Exit(2)
	}
	cs, err := discover(path)
	if err != nil {
		fatal(err.Error())
	}
	cs = prepareCandidates(cs, config.candidateOptions)
	if err := writeCandidates(os.Stdout, cs, config.format); err != nil {
		fatal(err.Error())
	}
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
