# Git-Based Relationship Discovery Product Requirements

## Purpose

Help users find candidate artifacts related to a repository path using inspectable Git history. Co-change is evidence of a possible relationship, not proof of a semantic relationship; consumers decide which candidates are relevant.

## Product direction

Use commits affecting the target to identify supported code and document paths changed alongside it. Retain the supporting commits and message-derived noise evidence for downstream analysis and classification. Exclude the target itself and count a path at most once per commit.

A historical path may be used to seek current relocation candidates. History following is optional and disabled by default.

## Current implementation

The Git-backed adapter resolves the repository root and tracked repository-relative target, reads its current-path history, or rename-following history when `--follow` is enabled, and collects commit messages and changed paths. Git command execution is injectable; the default invokes the `git` executable. Scoring and classification are handled by [analysis](../../analysis/prd.md).

## Evaluation and future signals

Git history can contain unrelated co-changes, including bulk formatting, dependency updates, and refactoring. Preserve evidence so users can inspect candidates and noise decisions. The [product PRD](../../prd.md) identifies repeated co-changes, diffs, author and chronology patterns, and work-item context as possible signals. Branch, pull-request and merge information may also be useful. These signals are not part of the current implementation or commitments to a particular approach.

## Related documents

- [Discovery](../prd.md) defines the shared adapter boundary.
- [Analysis](../../analysis/prd.md) interprets aggregated evidence.

## Relocation and history direction

Default discovery examines current-path history without following renames. Use available path and filename, identical content, and Ctags definition clues to find possible current destinations of a missing historical path, including multiple split candidates. Clues are evidence, not proof; no match does not establish deletion. An explicit `--follow` mode requests Git rename-following history. Keep each clue inspectable and report missing Ctags rather than silently asserting absence. Tree-sitter is a future option to evaluate, not a performance guarantee.
