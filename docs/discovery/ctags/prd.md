# Ctags-Based Relationship Discovery Product Requirements

## Purpose

Help users find code and test candidates related to an implementation using inspectable symbol evidence. This supports exploration when the relationships between artifacts are not already known. Consumers decide which candidates are relevant to their task.

## Product direction

Use definitions extracted with Universal Ctags as clues for finding related artifacts. Preserve the originating implementation, identifier and matching location so users can inspect the evidence. A textual occurrence, including one in a comment, is not a resolved reference, a call, or proof of test coverage. Ambiguous names, language syntax and indirect relationships can cause false positives or omissions. No match does not prove a relationship is absent.

Distinguish textual matches from syntactic calls and statically resolved calls if those become available. Reachability alone does not establish a feature boundary. Broader code-to-code exploration, supported languages, and techniques for resolving relationships remain open questions, not commitments of the current implementation.

## Current experimental behavior

The [ADR](adr.md) records the experimental implementation and its design choices.

## Evaluation and open questions

Evaluate symbol evidence separately and together with filename and Git evidence, using reviewed examples including ambiguous names, comments, indirect calls, shared tests and missing relationships. Record tool availability and version, language support, failures, evidence, omissions, unrelated candidates and execution time. The [evaluation plan](../../discovery-evaluation.md) describes the broader comparison and initial feasibility observations. Choose coverage, ranking and defaults based on evaluation rather than assuming textual matches establish relationships.
