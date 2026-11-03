# ADR: Ctags-Based Relationship Discovery

## Status

Direction accepted; the current test-discovery implementation is experimental. Broader relationship discovery and analysis choices remain open.

## Context

The [PRD](prd.md) calls for inspectable evidence to help find related artifacts. Symbol names can connect an implementation to candidate tests even when filenames differ, but textual occurrences can be unrelated or occur in comments. Indirect relationships and unsupported syntax can be missed. A match is not a resolved reference, call, or proof of coverage; absence of a match does not prove absence of a relationship.

## Decision

Use definitions extracted by Universal Ctags as clues for textual matches, retaining the originating implementation, identifier and match location as evidence. Keep this evidence distinct from filename and Git co-change evidence. Consumers decide relevance. Distinguish textual matches from syntactic and resolved calls if those are introduced.

## Experimental implementation

Test discovery expands existing implementation candidates once, before filters and limits. For each existing regular implementation file, it runs Ctags with JSON output and uses tag names of at least three bytes. It searches classified test files line by line for names bounded by non-identifier bytes; this is textual matching, not language-aware resolution. Test classification follows path conventions, so inline tests and unrecognized conventions can be missed. The scan excludes `.git`, `vendor`, and `node_modules` directories.

Filename and Git co-change clues are collected separately. Evidence is merged by path without an independent-method agreement bonus; new test candidates start at score zero and existing scores are unchanged. Missing or non-regular implementation files are skipped. If Ctags is unavailable, fails, or produces invalid output, itoc warns and continues with filename and Git discovery. Git errors fail the operation. Expansion is enabled by default and can be disabled with `--discover-tests=false`.

## Consequences and open questions

This approach gives inspectable clues without requiring reference resolution, but ambiguous names, comments, short names, indirect calls and language-specific syntax can produce false positives or omissions. Path-based classification also limits coverage. Evaluate evidence separately and together with other methods using the [evaluation plan](../../discovery-evaluation.md), including tool versions, failures, omissions, unrelated candidates and execution time. Language coverage, code-to-code exploration, ranking, defaults and possible language-aware resolution remain open; the experiment does not commit to them.
