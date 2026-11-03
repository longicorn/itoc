# Filter Product Requirements

## Purpose

Filtering prepares discovered candidates for presentation, allowing users to focus on candidates relevant to their task.

## Requirements

- Allow users to select candidates by extension and path and exclude individual candidates or groups.
- Allow users to control the noise threshold and minimum candidate score.
- Rank candidates and apply a configurable limit per group.

Candidate scores and classifications are defined by [analysis](../analysis/prd.md), options by [CLI](../cli/prd.md), and presentation by [output](../output.md).

## Current behavior

`prepareCandidates` applies extension and path filters, candidate and group exclusions, the noise threshold, minimum score, sorting, and a per-group limit. Candidates with a noise score above 1 are hidden by default; `--max-noise-score` changes the threshold and `--show-noisy` includes all candidates. A limit of zero means unlimited.

Plain path filters match a substring of the repository-relative path. Patterns containing glob metacharacters match the full path using Go `filepath.Match` syntax; `*` does not cross `/`. Extension filters compare lower-case extensions. `--exclude-candidate` accepts a group name or an exact repository-relative candidate path.
