# ADR: Keyword Search Entry Candidates

## Status

Direction accepted; search and filtering details remain open. A CLI placeholder is available; it does not search.

## Context

A user may know a feature name or displayed text without knowing a repository path. Keyword search provides starting candidates for exploration. A match is evidence to inspect, not confirmation that a file describes or implements the intended feature; an unsuccessful search does not establish absence.

## Decision

- Accept an explicitly supplied keyword and return candidate repository paths with evidence identifying what matched and where.
- Treat these paths as starting candidates. Further relationship exploration belongs to [discovery](../discovery.md); the user or consuming workflow chooses which candidates to investigate or accept.
- Use `itoc [options] --search <keyword>` as the proposed keyword CLI form alongside `itoc [options] <file-path>`. Explicit selection avoids interpreting a nonexistent path as a keyword. The placeholder reports that search is not implemented without invoking discovery or scoring.

## Consequences and exploration

An initial approach to evaluate is literal searching of file contents, like grep, alongside matching the keyword against repository paths, like filtering a file listing. Report content and path matches as distinct evidence so their usefulness can be assessed. These are approaches to consider, not a commitment to particular commands or a complete search design. Further filtering or selection can narrow the starting candidates, but its method is not yet decided.

Literal matching is straightforward to inspect but can miss alternate terminology and produce unrelated results. Candidate selection remains separate from deciding the scope of an assessment.

## Open questions

- Which files and locations should be searched, and what evidence should be displayed?
- How should multiple candidates be narrowed, selected, and passed to discovery?
- How should common or ambiguous matches be ranked or filtered? Evaluate choices and weights using concrete examples.
- How should empty or invalid input and searches with no matches be reported?
- Which implementation techniques should be used, and how should they relate to the proposed `keyword.go`?

The existing candidate output assumes scores and Git co-change evidence. The placeholder has a separate response so it does not imply that an empty search was performed. The result model, output format, and connection to discovery remain to be designed when search is implemented.
