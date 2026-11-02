# /cli

## Purpose

`/cli` is the command-line entry point for running discovery, candidate preparation, and rendering as one workflow.

## Current behavior

The executable accepts exactly one repository file path. `parseCLI` parses options and validates the output format, non-negative score threshold, and non-negative per-group limit. Repeatable include and exclude flags are collected in `cliConfig`.

`run` coordinates the workflow: parse arguments, discover candidates through an injected `PathLister`, prepare them, and write the selected text or JSON representation. `--discovery` selects `git` (default), `document`, or `both`. Both runs the first two concurrently and merges candidates without changing scores when they agree. Document mode dispatches Markdown/MDX and reStructuredText to separate extractors. Noise rules hide candidates above the default score threshold; `--show-noisy` or `--max-noise-score` adjusts this. `--verbose` reveals detailed evidence strings. Errors are returned to `main`, which reports them on stderr and exits unsuccessfully.

The CLI's supported options and examples are listed in the [README](../README.md); this page describes responsibility boundaries rather than duplicating the full reference.

## Related paths

- [`/discovery`](discovery.md) obtains and analyzes candidates.
- [`/filter`](filter.md) prepares the candidate set.
- [`/output`](output.md) renders results.
