# ADR: Evidence-Based Artifact Relationship Discovery

## Status

Accepted direction; implementation details remain open.

## Context

`itoc` aims to make relationships between intent-bearing artifacts and implementation artifacts easier to discover and reuse. Development tools may rely on AI-driven exploration to resolve code from semantic paths. AI can help interpret ambiguous evidence, but AI alone makes discovery difficult to reproduce, explain, and reuse consistently.

The initial use case is Forward development: use committed documents and repository history to find candidate implementation artifacts. Locating related artifacts is distinct from understanding their intent. This ADR focuses on candidate discovery and does not define a complete semantic interpretation system.

Git history provides observable evidence, including co-changes, later changes to an established document, repeated change patterns, diffs, changed paths, authorship, chronology, and available work-item context. Branch, pull request, and merge context may also help: changes grouped in the same branch or PR can be related, and a bug fix may address the underlying cause or be included as a small companion fix. Such grouping can increase the likelihood of a relationship, but cannot establish it by itself.

History also contains noise, such as bulk formatting, dependency updates, and rename-only refactoring. Commit messages may be vague, and a commit or PR may combine unrelated work. Candidate results therefore need to retain their evidence and uncertainty.

## Current discovery direction

Discovery starts from a repository path or keywords, including a feature name or displayed text. Nagi resolves its semantic path to a document path before passing it to `itoc`. Results are candidates with evidence; consumers decide which to use for assessment or another purpose. A textual match, co-change, or reachable call does not by itself establish an implementation or test relationship.

Discovery can locate entry candidates and follow relationships to other artifacts. Static call relationships are one possible source of evidence, but reachability alone does not define a feature boundary. Distinguish textual matches, syntactic calls, and statically resolved calls, and expose unresolved relationships and analysis limitations. Initially consider static analysis using readily available external tools or libraries; tool selection and language coverage remain open. Runtime observation is a future possibility, not a current commitment.

The desired candidates include documents that define or explain the subject, implementation carrying its behavior or constraints, and tests verifying it. The inclusion of higher-level documents and shared code depends on the purpose and remains open. Ranking purpose, scores, weights, and the balance between missed and irrelevant candidates should be decided through concrete examples and evaluation. Refine implementation incrementally as experience accumulates.

## Earlier Git-focused decisions

The following records the initial direction and CLI decisions. The broader discovery direction above is not limited to committed documents or Git history; specific proposed Git signals and grouping approaches still require implementation and evaluation.

## Decision

- Implement `itoc` in Go. Follow common Go project layout conventions rather than introducing a bespoke directory structure.
- Establish the initial Go project as a minimal executable module: use the repository root for the command entry point, declare the module in `go.mod`, and avoid additional packages or dependencies until an implementation need arises.
- Start with Forward discovery of implementation candidates for committed documents, using Git history as a primary reproducible evidence source.
- Accept a repository file path as the initial command-line input and use that file's Git history to identify and rank likely related implementation code files. The output is a set of evidence-backed candidates, not a claim of confirmed semantic relationships.
- Treat a branch/PR change group containing the target document's first commit as the initial search unit. If later branch/PR groups also change that document, process those groups as subsequent search units. Within each group, prioritize co-changes, follow-up changes, and diffs; grouping controls search priority and evidence context, not whether a semantic relationship is confirmed.
- Combine signals where available: co-changes and repeated co-changes, subsequent changes, diffs and paths, authorship and chronology, work-item context, and branch/PR/merge grouping.
- Treat each signal as evidence that may increase or decrease a candidate's likelihood, not as proof of semantic relationship. In particular, same-commit, same-branch, same-PR, or same-merge membership does not confirm a relationship.
- Present the supporting evidence and remaining uncertainty so a person can assess the candidate.
- Use token similarity across repository-relative paths as a modest ranking signal in addition to Git history, and expose its contribution separately. The repository root is excluded; matching path components can connect documents and source files with different extensions or basenames (for example, `users/prd.md` and `users/users.go`). Path similarity alone does not establish a relationship.
- Detect likely test files using common language-specific filename prefixes/suffixes and directory conventions. Evaluate filename markers on the stem after removing the extension (for example, `_spec.rb` and `_test.go`). Keep these candidates but classify and list them separately after implementation candidates; the heuristic is best-effort and must not imply certainty.
- Allow code files as input as well as documents. Include related document files among the candidates and present them in a separate group.
- Provide repeatable include/exclude options for candidate file extensions and repository-relative paths to let users reduce noisy results. Define and document the glob matching behavior.
- For path filters, treat a plain string as a substring of the repository-relative path; patterns containing glob metacharacters match the full path.
- Treat plain `--include-path` and `--exclude-path` values as substrings of repository-relative paths; patterns containing glob metacharacters match the full path.
- Limit results to the top 10 candidates per group by default, with a command-line override and an explicit unlimited setting.
- Allow a minimum score threshold in addition to the per-group result limit. Apply the score threshold first, then the limit.
- Allow users to hide candidate groups (`implementation`, `document`, `test`) or individual candidates by exact repository-relative path, repeatably, to remove known noise.
- Track a separate noise score from commit-message terms such as `lint`, `linter`, `format`, `formatter`, `refactor`, and `refactoring. Do not combine it with the candidate score or use it to rank/filter candidates yet; expose it as evidence for later evaluation.
- Limit results to the top 10 candidates per group by default, with a command-line override and an explicit unlimited setting. This keeps the initial output reviewable while allowing broader inspection.
- Keep AI as optional assistance for interpretation, candidate expansion, or summarization; it does not replace reproducible evidence.
- Defer Jev-based noise classification from the initial scope. Keep AI/LLM assistance optional and selectable by the user for interpretive support; it does not replace reproducible evidence.
- Write results to stdout by default, with options to choose an output format and an output file.
- Evaluate first on the team's own product, then expand to public OSS once the tool is working. A person makes the final assessment of candidate relevance.
- Keep artifact location and deeper understanding of code intent as distinct concerns. The initial goal is primarily to locate candidates; insufficient evidence must remain visible as uncertainty.

## Consequences

Repeated and diverse evidence can make candidate ranking more useful, including when commit messages are uninformative. Branch, PR, and merge metadata can contribute context for related bug fixes and companion changes. However, broad changes and incidental fixes can also create misleading associations, so the tool must not turn grouping metadata into a definitive link.

Forward history may later support Reverse Sync, where code is the starting point. This is a possible future benefit, not part of the initial implementation commitment.

## Open Questions

- How should candidate quality, precision/recall, review time, and search effort be evaluated?
- What additional value do vocabulary, structure, and tests provide over history alone?
- How should noise be identified, and what measurable benefit does Jev add over Git-only rules?
- Should confirmed relationships be persisted and reused, and how should freshness and rename, split, or merge changes be handled?
- What repository hosting metadata is available for branch, PR, and merge evidence, and how should it be used when unavailable?
