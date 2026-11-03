# CLI Product Requirements

## Purpose

The CLI provides a command-line entry point for discovering relationships between repository artifacts. It brings discovery, candidate preparation, and rendering together in one workflow, so users can inspect candidates and the evidence behind them.

## Requirements

- Accept a repository file path as the starting point for discovery.
- Let users select available discovery methods and control candidate filtering, noise thresholds, and result limits.
- Present candidates in text or JSON, with detailed evidence available when requested. Candidates are clues, not proof of a relationship.
- Report invalid inputs and workflow errors to the user with an unsuccessful exit status.

The CLI coordinates the workflow; [discovery](../discovery/prd.md) obtains and analyzes candidates, [filtering](../filter/prd.md) prepares them, and [output](../output.md) renders them. Supported options and examples are documented in the [README](../../README.md).

## Current behavior

The executable accepts one repository file path. It validates the output format, non-negative score threshold, and non-negative per-group limit. Include and exclude flags may be repeated.

`--discovery` accepts a method name, comma-separated names, or `all`; the default is `git`. Selected methods run concurrently, and each additional method finding a candidate adds a provisional one-point agreement bonus. Document discovery supports Markdown/MDX and reStructuredText through separate extractors.

Noise rules hide candidates above the default threshold; `--show-noisy` and `--max-noise-score` adjust this. `--verbose` exposes detailed evidence. Errors are printed to stderr and cause an unsuccessful exit.

The experimental `--search` entry point can find a starting path from a keyword; its behavior and open questions are described in the [keyword search PRD](../keyword/prd.md).
