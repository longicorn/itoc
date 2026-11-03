# ADR: Discovery boundary and orchestration

## Status

The separation of evidence collection from analysis is adopted. The shared evidence model and the experimental test expansion remain subject to evaluation.

## Context

The [discovery PRD](prd.md) calls for independently selectable methods that produce inspectable evidence for related path candidates. Evidence is a clue, not proof of a semantic relationship. The [analysis PRD](../analysis/prd.md) assigns interpretation, scoring, and classification to analysis, so adapters can focus on collecting evidence. Method-specific decisions are recorded in the [Git ADR](git/adr.md) and [Ctags ADR](ctags/adr.md).

## Decision

Each primary discovery method implements a common boundary: given an input, it returns evidence, the resolved target for analysis, or an error. After a successful collection, analysis converts that method's evidence into candidates. Evidence and its originating methods remain available for inspection; users decide whether a candidate is relevant. Scoring and classification rules belong to analysis.

Methods are selected by name from the registry; `all` selects registered methods. Duplicate selections are removed and unknown or empty names are rejected. When multiple methods are selected, they run concurrently and each result is analyzed before candidates are merged by repository-relative path. The merge preserves evidence and method identities. Its provisional agreement bonus and score handling are heuristics, not calibrated probabilities.

A failure in a selected method fails a multi-method run with the method name, except that an unsupported document format skips the document method when multiple methods run. A document-only run reports that condition as an error. Individual adapters may have more specific handling for incomplete evidence; their ADRs describe those limits.

After primary results are merged, optional test discovery expands implementation candidates by one step. This derived evidence is merged by path without an independent-method agreement bonus; new test candidates start at score zero and existing scores are unchanged. Candidate preparation, including filters and limits, follows expansion, and output follows preparation. The [Ctags ADR](ctags/adr.md) describes the experimental expansion and its failure behavior.

## Package organization direction

Refactor discovery toward `internal/discovery/`, separating the shared boundary and orchestration from responsibilities such as Git and document evidence collection. Discovery and analysis share the `Evidence` struct defined in `internal/discovery`; discovery does not depend on analysis. The struct is the boundary, while its fields may evolve as evidence needs change.

Consider separating Ctags-related code in light of its role in experimental test expansion. Do not fix a tree-sitter package structure until its intended use and responsibilities are clear. Introduce boundaries as needed rather than assuming each tool must have a separate package.

Migrate incrementally while preserving observable behavior, including evidence, candidate scores, selection, and error handling. Existing tests and focused regression checks should verify behavior across each move.

Git and document evidence collection live in `internal/discovery`; document format clue extraction lives in `internal/discovery/document`, and Ctags invocation and JSON parsing in `internal/discovery/ctags`. Test expansion passes implementation paths to discovery, which returns derived `Evidence` for test paths. The root package selects implementation candidates and merges those clues into candidates without changing existing scores or awarding a method agreement bonus. The registry and analysis orchestration remain in the root package.

## Discovery interface

Each primary method implements `PathLister.List(input string) ([]Evidence, string, error)`. It returns evidence and the resolved target, or an error. Target and candidate paths are repository-relative. The initial evidence fields carry the path, distinct Git commit IDs, noise score and terms, descriptive details, and method names. Commit IDs may be empty for non-Git methods. These fields preserve current behavior and are not a fixed schema for future evidence. Analysis consumes evidence and remains responsible for candidate scoring and classification.

An exported unsupported-document-format error lets orchestration preserve the existing distinction between a skipped method in a multi-method run and an error in a document-only run. Implementations and analysis depend on the shared discovery boundary, not on each other.

## Open questions

The provisional merge and test-expansion scoring rules require evaluation. Adding new methods or changing the processing order may require revisiting evidence representation, error handling, and scoring without treating an additional match as proof of relevance.
