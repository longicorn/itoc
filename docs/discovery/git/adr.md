# ADR: Git-Based Relationship Discovery

## Status

Accepted.

## Context

The [PRD](prd.md) calls for inspectable Git evidence to find potentially related artifacts. A co-change is a clue, not proof of a semantic relationship. Bulk changes and unrelated work in a commit can produce misleading candidates; consumers decide relevance.

## Decision

Discover candidates from commits affecting a tracked repository path. A deleted input path is rejected by default; `--allow-deleted` enables Git history discovery for paths found in history. For a moved or renamed target, start from its current destination and follow its history. Retain supporting commit IDs and noise evidence for analysis rather than treating co-change as confirmation.

Resolve the repository root and input to a repository-relative tracked path, then use `git log --follow` to obtain its commit IDs. For each commit, collect its subject and body and the paths reported by `git diff-tree --root --no-commit-id --name-only -r`. Include supported code and document paths, excluding the target itself, and count each path at most once per commit. Aggregate evidence by path, retaining commit IDs, co-change evidence, and noise terms derived from messages and paths. Noise scoring and candidate interpretation are described in [analysis](../../analysis/prd.md).

Git execution is injectable; by default the adapter invokes the `git` executable. Failure to resolve the repository, resolve the target, or obtain the log fails discovery. A message retrieval failure leaves that commit without message evidence; a diff-tree failure skips that commit. These behaviors can result in incomplete evidence.

## Consequences and limitations

History and supported path types bound the results. Co-changes can be unrelated, and relevant files may never have changed in the same commit. `--follow` traces the target history, but the exclusion compares changed paths with the current target path; an earlier name can therefore appear as a candidate. The adapter does not collect diffs, authors, chronology, branch, pull-request or work-item context as evidence.

Target resolution checks the working tree by default. With `--allow-deleted`, it can resolve a missing path from Git history. Scoring and classification are separate from discovery.
