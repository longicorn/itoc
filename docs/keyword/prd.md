# Keyword Search Product Requirements

## Purpose

Keyword search provides an entry point for finding repository artifacts when the user does not know a relevant path. A feature name or text displayed in a user interface may be enough to begin exploring an unfamiliar implementation or finding documentation.

## Product direction

Given an explicitly supplied keyword, itoc should find candidate repository paths and report the evidence for each match. Evidence should identify what matched and where, so a person or an AI can inspect the result. A literal match is a clue, not proof that the artifact implements or describes the intended feature. No matches do not prove that the relevant artifact is absent.

These results are starting candidates. Further exploration of relationships from a selected path belongs to [discovery](../discovery/prd.md). The user or a consuming workflow decides which candidates to investigate or accept; keyword search does not determine the scope of an assessment.

The proposed CLI forms are `itoc [options] <file-path>` and `itoc [options] --search <keyword>`. The explicit option avoids interpreting a nonexistent path as a keyword. The CLI also accepts `--search <keyword>` as an experimental entry point.

## Open questions

- Which repository files and locations should be searched, and what match evidence should be displayed?
- How should multiple starting candidates be selected and passed to discovery?
- How should matches be ranked or filtered, including common or ambiguous terms? Weights and ordering should be informed by evaluation rather than assumed.
- How should empty, invalid, or unsuccessful searches be reported?

Implementation techniques and the relationship to `keyword.go` remain design decisions. Explore them incrementally using concrete examples and evaluation.

## Experimental implementation

In a Git repository, `--search` uses literal content matches in tracked files. One matching path is selected automatically; for multiple paths, the user selects a numbered entry. The selected path is passed to discovery. Prompts go to stderr; no matches stop the workflow. Non-Git repositories and path matching remain future work.
