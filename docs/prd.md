# itoc Product Requirements Document

## Overview

`itoc` is a dedicated tool for discovering and maintaining relationships between intent-bearing artifacts and implementation artifacts. Its name combines `ito` (意図, intent, evoking the Japanese word for thread) with code: “ito to code” and “ito + code.” The tool addresses a practical traceability problem across software development.

## Problem

Development tools may resolve code from semantic paths using AI-driven exploration. Relying on non-deterministic AI reasoning alone makes it difficult to reproduce, explain, and consistently reuse how relevant code was found. AI assistance remains valuable, but by itself it leaves too much uncertainty and repeated exploration.

Developers often have to reconstruct which implementation, tests, and related artifacts correspond to a document or feature. This work is an implicit, recurring cost, and there has not been a broadly effective way to preserve and recover these relationships.

Reverse workflows make the need more immediate. They must reason from existing code toward intent, follow relationships from one code artifact to related code, and use those relationships to help extract semantic information. Reliable relationship discovery can support both Forward development and Reverse work.

## Product Direction

`itoc` will make artifact relationships easier to discover by combining reproducible evidence with AI assistance. Deterministic methods should identify and rank candidates using evidence such as Git history, changed paths, diffs, and repository structure. Results should explain their supporting evidence and expose uncertainty. AI may help interpret ambiguous evidence, expand candidate searches, or summarize relationships, but should supplement rather than replace reproducible methods.

An association is a candidate unless the evidence establishes more. For example, two artifacts appearing in the same commit is useful evidence, but does not by itself prove that they are semantically related.

### Discovery inputs and results

Discovery starts from either a repository path or keywords, such as a feature name or text displayed in a user interface. A known document or code path can lead to related artifacts; keywords can provide an entry point when no path is known. In Nagi, a semantic path can first be resolved to a document path and used as an input to `itoc`.

The desired results are candidates for documents that define or explain the subject, implementation that carries its behavior or constraints, and tests that verify it. Higher-level documents and shared code may provide context; whether they belong in a particular task's scope depends on their role. Results retain the evidence used to discover them, and consumers decide which candidates to accept for their purpose, such as assessment.

Discovery may first locate entry candidates and then follow relationships to find further artifacts. Static call relationships are one possible way to explore implementation, but reachability alone does not determine the boundary of a feature. Distinguish text matches, syntactic calls, and statically resolved calls, and expose limitations and unresolved relationships. Initially, consider static analysis using readily available external tools such as Ctags or libraries such as tree-sitter; specific tools and language coverage remain open. Runtime observations are a longer-term possibility, not a current requirement.

## Initial Goal

An initial concrete workflow is Forward development: when a document is committed, use its content and Git history to find implementation code related to that document. Present candidate code artifacts with the historical and other available evidence that supports each relationship.

The initial exploration includes Git-backed discovery. Git evidence may include not only document and code changes in the same commit, but also later commits that modify an established document, repeated co-changes, diffs, author and chronology patterns, and work-item context when available. These signals can help even when commit messages are uninformative. Repeated evidence may strengthen a candidate, while no single signal should be treated as proof.

Discovery should distinguish locating artifacts from understanding their intent. The initial goal is primarily to locate candidate implementation artifacts. Deeper interpretation of what the code implements and which intent it serves is a broader need, especially for Reverse Sync, and must retain uncertainty when repository evidence is insufficient.

The initial implementation supports both Git-backed co-change discovery and deterministic document-reference discovery. Document discovery starts with Markdown/MDX and reStructuredText, extracting explicit local links, inline code/literals, and emphasized terms, then matching literal paths, file names, or source text. It does not attempt semantic paraphrase. These methods remain independently selectable and extensible by format-specific extractors, with evidence retained for each result. Confirmed relationships may eventually be retained and reused across Forward and Reverse work; persistence, freshness, and handling splits or merges remain open design questions. For Git-backed discovery, a deleted input path is out of scope; when a target is moved or renamed, discovery should continue from its current destination while following its history.

## Noise and Evidence Quality

Evaluate ways to discover test candidates from implementation candidates using separate evidence from symbol names, file naming and Git co-changes. A textual occurrence of a symbol is not a resolved reference or proof that a test covers the implementation. Also evaluate ranking document clues by match type and frequency while preserving their evidence. Common terms should not be excluded solely because of frequency, and weights should be chosen through evaluation rather than assumed. The [evaluation plan](discovery-evaluation.md) describes the proposed experiments; test discovery has an initial implementation, while document-clue ranking remains a direction for investigation. Future work may improve language-aware discovery accuracy.

Git history contains changes that can obscure semantic relationships, including bulk linter or formatter application, package updates, and refactoring that only renames or moves files. Distinguishing these from changes that express or implement intent is a key product challenge.

Potential signals include commit messages and comments, change shape, file moves, and other repository evidence. Jev may be explored as a way to classify whether a change is noise. The classification approach is not yet decided; the tool should preserve the evidence behind its decisions so users can review them.

## Future Direction

In Reverse workflows, `itoc` may help start from code, discover related code artifacts, and support semantic extraction from the resulting evidence. The same relationship model may eventually cover documents, source code, tests, decisions, and other development artifacts in both Forward and Reverse workflows.

## Principles

- Prefer reproducible, inspectable evidence for candidate discovery.
- Use AI as an assistant for uncertain or interpretive work, not as the sole discovery mechanism.
- Show why an artifact was selected and where uncertainty remains.
- Treat Git co-change and other signals as evidence, not automatic proof of semantic relationships.
- Use concrete workflows and evaluation to guide incremental expansion.

## Reference

This PRD is grounded in the ISE problem framing in [`02_artifact_relationship_traceability.md`](../../ise-methodology/ise/foundations/02_artifact_relationship_traceability.md).
