# itoc Product Requirements Document

## Overview

`itoc` is a dedicated tool for discovering and maintaining relationships between intent-bearing artifacts and implementation artifacts. Its name combines `ito` (意図, intent, evoking the Japanese word for thread) with code: “ito to code” and “ito + code.” The tool addresses a practical traceability problem that is especially visible in Nagi, while also applying to software development more broadly.

## Problem

Nagi resolves code from semantic paths, but the discovery process currently relies on AI-driven exploration. Relying on non-deterministic AI reasoning alone makes it difficult to reproduce, explain, and consistently reuse how relevant code was found. AI assistance remains valuable, but by itself it leaves too much uncertainty and repeated exploration.

The underlying cost predates Nagi. Developers often have to reconstruct which implementation, tests, and related artifacts correspond to a document or feature. This work is an implicit, recurring cost, and there has not been a broadly effective way to preserve and recover these relationships.

Nagi's planned Reverse Sync makes the need more immediate. Reverse Sync must reason from existing code toward intent. It may need to follow relationships from one code artifact to other related code and use those relationships to help extract semantic information. Reliable relationship discovery can support both Forward development and this future Reverse work.

## Product Direction

`itoc` will make artifact relationships easier to discover by combining reproducible evidence with AI assistance. Deterministic methods should identify and rank candidates using evidence such as Git history, changed paths, diffs, and repository structure. Results should explain their supporting evidence and expose uncertainty. AI may help interpret ambiguous evidence, expand candidate searches, or summarize relationships, but should supplement rather than replace reproducible methods.

An association is a candidate unless the evidence establishes more. For example, two artifacts appearing in the same commit is useful evidence, but does not by itself prove that they are semantically related.

## Initial Goal

Begin with Forward development: when a document is committed, use its content and Git history to find implementation code related to that document. Present candidate code artifacts with the historical and other available evidence that supports each relationship.

The initial scope is intentionally focused on Git-backed discovery. Git evidence may include not only document and code changes in the same commit, but also later commits that modify an established document, repeated co-changes, diffs, author and chronology patterns, and work-item context when available. These signals can help even when commit messages are uninformative. Repeated evidence may strengthen a candidate, while no single signal should be treated as proof.

Discovery should distinguish locating artifacts from understanding their intent. The initial goal is primarily to locate candidate implementation artifacts. Deeper interpretation of what the code implements and which intent it serves is a broader need, especially for Reverse Sync, and must retain uncertainty when repository evidence is insufficient.

The initial scope is intentionally focused on Git-backed discovery. It should establish a practical basis for later work on broader artifact types, finer-grained links, and Reverse Sync. Confirmed relationships may eventually be retained and reused across Forward and Reverse work; persistence, freshness, and handling moves, splits, or merges remain open design questions.

## Noise and Evidence Quality

Git history contains changes that can obscure semantic relationships, including bulk linter or formatter application, package updates, and refactoring that only renames or moves files. Distinguishing these from changes that express or implement intent is a key product challenge.

Potential signals include commit messages and comments, change shape, file moves, and other repository evidence. Jev may be explored as a way to classify whether a change is noise. The classification approach is not yet decided; the tool should preserve the evidence behind its decisions so users can review them.

## Future Direction

As Nagi adds Reverse Sync, `itoc` may help start from code, discover related code artifacts, and support semantic extraction from the resulting evidence. The same relationship model may eventually cover documents, source code, tests, decisions, and other development artifacts in both Forward and Reverse workflows.

## Principles

- Prefer reproducible, inspectable evidence for candidate discovery.
- Use AI as an assistant for uncertain or interpretive work, not as the sole discovery mechanism.
- Show why an artifact was selected and where uncertainty remains.
- Treat Git co-change and other signals as evidence, not automatic proof of semantic relationships.
- Start with a narrow Forward workflow and use its results to guide expansion.

## Reference

This PRD is grounded in the ISE problem framing in [`02_artifact_relationship_traceability.md`](../../ise-methodology/ise/foundations/02_artifact_relationship_traceability.md).
