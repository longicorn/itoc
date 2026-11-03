# Git-Based Relationship Discovery Product Requirements

## Purpose

Help users find candidate artifacts related to a repository path using inspectable Git history. Co-change is evidence of a possible relationship, not proof of a semantic relationship; consumers decide which candidates are relevant.

## Product direction

Use commits affecting the target to identify supported code and document paths changed alongside it. Retain the supporting commits and message-derived noise evidence for downstream analysis and classification. Exclude the target itself and count a path at most once per commit.

A deleted input path is out of scope. When a target is moved or renamed, discovery should start from its current destination and follow its history. The current implementation still resolves some deleted tracked paths from history, so it does not yet enforce the deleted-path policy.

## Current implementation

The Git-backed adapter resolves the repository root and tracked repository-relative target, follows its history with `git log --follow`, and collects commit messages and changed paths. Git command execution is injectable; the default invokes the `git` executable. Scoring and classification are handled by [analysis](../../analysis/prd.md).

## Evaluation and future signals

Git history can contain unrelated co-changes, including bulk formatting, dependency updates, and refactoring. Preserve evidence so users can inspect candidates and noise decisions. The [product PRD](../../prd.md) identifies repeated co-changes, diffs, author and chronology patterns, and work-item context as possible signals. Branch, pull-request and merge information may also be useful. These signals are not part of the current implementation or commitments to a particular approach.

## Related documents

- [Discovery](../../discovery.md) defines the shared adapter boundary.
- [Analysis](../../analysis/prd.md) interprets aggregated evidence.
