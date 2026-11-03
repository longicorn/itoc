# Analysis Product Requirements

## Purpose

Analysis turns discovery evidence into ranked, classified candidate artifacts. It separates evidence collection from interpretation so discovery adapters can focus on obtaining repository facts.

## Product direction

Interpret discovery evidence as support for candidates, not as proof of semantic relationships. Preserve inspectable evidence and avoid presenting candidates as confirmed links. The [product PRD](../prd.md) describes the broader goals for evidence and uncertainty.

Candidate preparation and rendering are addressed by [filter](../filter/prd.md) and [output](../output/prd.md).

## Current implementation

`discoverWithLister` obtains evidence from a `PathLister` and calls `analyzeCandidates`. The current implementation receives `cochangeEvidence` records containing a repository-relative path, the set of commits that changed the path with the target, and noise evidence derived from commit messages. `analyzeCandidates` creates a candidate for each record.

The score is the number of co-changing commits plus a filename similarity bonus. Noise score is separate: commit-message signals add small values, while built-in path rules such as benchmark/performance paths add a high value. Candidates above the default noise threshold are hidden; `--show-noisy` includes them, and `--max-noise-score` changes the threshold. Noise rules can be extended independently of user include/exclude filters.

These are heuristics over Git-derived evidence. The current implementation does not inspect source semantics or use an LLM.

## Future direction

Analysis may incorporate richer evidence while retaining inspectability and uncertainty. The evidence and approach remain open.
