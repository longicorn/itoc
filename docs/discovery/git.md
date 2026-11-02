# /discovery/git

## Purpose

`/discovery/git` implements `PathLister` using the repository's Git history. It is the current production source of discovery evidence.

## Current behavior

`GitPathLister` finds the repository root, resolves the input to a tracked repository-relative path, and reads commits affecting that target with `git log --follow`. This follows a moved or renamed target from its current destination into its history. Product policy treats deleted input paths as out of scope. The current implementation still has a history-based fallback for deleted tracked paths, so it does not yet enforce that policy. For each commit it gathers the subject/body and changed paths. Git command execution can be injected through `GitRunner`; the default invokes the `git` executable.

`collectCochangeEvidence` aggregates supported code and document paths that changed in the same target commits. It deduplicates a path within each commit, excludes the target itself and unsupported file types, records the commit set, and accumulates noise terms from commit messages. Scoring and classification occur afterward in [`/analysis`](../analysis.md).

This implementation currently uses commit co-change evidence. It does not yet use branch, pull-request, merge, diff-content, author, chronology, or work-item signals described as possible inputs in the [PRD](../prd.md).

## Related paths

- [`/discovery`](../discovery.md) defines the adapter boundary.
- [`/analysis`](../analysis.md) interprets aggregated evidence.
