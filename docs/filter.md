# /filter

## Purpose

`/filter` prepares discovered candidates for presentation by applying user selection, ranking, and per-group limits.

## Current behavior

`prepareCandidates` applies extension/path filters, candidate/group exclusions, the noise threshold, minimum score, sorting, and per-group limit. Noisy candidates are hidden by default when their score is above 1; `--max-noise-score` changes the threshold and `--show-noisy` includes all candidates. A limit of zero means unlimited.

Plain path filters match a substring of the repository-relative path. Patterns containing glob metacharacters match the full path using Go `filepath.Match` syntax; `*` does not cross `/`. Extension filters compare lower-case extensions. `--exclude-candidate` accepts a group name or an exact repository-relative candidate path.

## Related paths

- [`/cli`](cli/prd.md) defines the options.
- [`/analysis`](analysis/prd.md) defines candidate scores and classifications.
- [`/output`](output.md) renders the prepared candidates.
