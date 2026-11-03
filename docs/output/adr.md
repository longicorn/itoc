# ADR: Candidate Output Boundary

## Status

Current implementation direction; output formats and their evolution remain open to evaluation.

## Context and interface

[Filter](../filter/adr.md) selects, orders and limits analyzed candidates for presentation. Output receives the resulting `[]candidate`, rather than raw discovery evidence. Each candidate carries a path, score, classification, noise information and supporting evidence. These are clues for assessment, not proof of a relationship.

The current rendering interface is `renderCandidatesWithEvidence([]candidate, format, showEvidence) ([]byte, error)`. The CLI selects the format and whether to show detailed evidence, calls the renderer and writes the resulting bytes. `renderCandidates` uses the same renderer with evidence hidden. Output does not select, rank or score candidates; filter owns selection and ordering, and analysis owns scoring and classification.

## Decision and current behavior

Render prepared candidates as text for people or JSON for downstream tools. Preserve the information needed to inspect candidates and expose detailed evidence when requested. The CLI owns option parsing and writing; output owns presentation.

Text rendering groups candidates under implementation, test and document headings, in that order, preserving the input order within each group. Each entry shows its path, score, noise score, co-change commit count and path bonus, plus noise terms when present. Detailed evidence is shown when requested. Empty text output says `No code candidates found.`

JSON rendering preserves input order and emits an array of records containing `path`, `score`, `co_changed_commits`, `filename_bonus`, `kind` and `noise_score`. It includes `noise_terms` when present and `evidence` when requested and available. Test-like candidates use kind `test`. An empty result is an empty array. The CLI currently writes the rendered result to standard output; output-file support is not implemented.

## Consequences and open questions

A shared rendering boundary keeps selection rules independent of presentation. Text grouping changes the overall order across groups, while JSON retains the filter's order. The displayed scores, classifications and evidence remain heuristic and require human assessment.

How should JSON compatibility be managed as candidate fields and evidence evolve? What additional output formats or evidence representations would help consumers?
