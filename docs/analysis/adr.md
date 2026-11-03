# ADR: Evidence-Based Candidate Analysis

## Status

Current direction; heuristics and their evaluation remain open to change.

## Context

Analysis turns discovery evidence into ranked, classified candidates. Evidence collection and interpretation have different responsibilities: discovery adapters obtain repository facts, while analysis prepares candidates for inspection. A co-change or path match supports a candidate but does not establish a semantic relationship. See the [analysis PRD](prd.md).

## Decision

- Create a candidate for each supplied evidence record, retaining its path, commits, evidence, discovery sources and noise information. The current interface is shared by Git and document discovery.
- Score a candidate using the number of distinct co-changing commits plus a path similarity bonus. The bonus compares weighted tokens in repository-relative paths, gives filename tokens more weight than directory tokens, and ranges from zero to three. Expose the bonus separately so that the score can be understood.
- Classify document paths by extension and identify test-like paths using directory and language-specific filename conventions. These are best-effort heuristics, not semantic classifications.
- Keep noise separate from candidate relevance. Discovery computes noise evidence from available signals, including Git commit-message terms and path rules. Candidate filtering hides scores above a configurable threshold by default; users can include noisy candidates. Retain noise terms and score for inspection. This reflects the current implementation and supersedes the earlier Git-focused direction in the [product ADR](../adr.md) that deferred noise filtering.
- Leave candidate filtering and rendering to their respective stages; analysis does not establish which candidates a consumer should accept.

## Consequences

Simple, inspectable signals provide a reproducible starting point, but co-changes can be incidental and path similarity can miss related files or favor unrelated ones. Test and document classifications can be wrong. Noise filtering can hide relevant candidates, so users need access to the unfiltered results. Scores are heuristic and should not be interpreted as probabilities or proof. Current analysis does not inspect source semantics or use an LLM.

## Open questions

- How should ranking and noise rules be evaluated against concrete examples, including missed relevant candidates and review effort?
- What additional evidence should analysis incorporate, and how should its contribution and limitations be exposed?
- How should scoring and classification evolve across discovery sources with different evidence?
