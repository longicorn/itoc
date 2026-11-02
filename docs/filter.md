# /filter

## Purpose

`/filter` prepares discovered candidates for presentation by applying user selection, ranking, and per-group limits.

## Current behavior

`prepareCandidates` applies stages in this order: extension/path include and exclude filters; exact candidate or group exclusions; minimum score; descending score sort with path as a stable tie-break; and the per-group limit. A limit of zero means unlimited. Group membership is implementation, document, or test, with test-like candidates separated from other code candidates.

Plain path filters match a substring of the repository-relative path. Patterns containing glob metacharacters match the full path using Go `filepath.Match` syntax; `*` does not cross `/`. Extension filters compare lower-case extensions. `--exclude-candidate` accepts a group name or an exact repository-relative candidate path.

## Related paths

- [`/cli`](cli.md) defines the options.
- [`/analysis`](analysis.md) defines candidate scores and classifications.
- [`/output`](output.md) renders the prepared candidates.
