# /analysis

## Purpose

`/analysis` turns discovery evidence into ranked, classified candidate artifacts. It separates evidence collection from interpretation so discovery adapters can remain focused on obtaining repository facts.

## Current behavior

The current implementation receives `cochangeEvidence` records containing a repository-relative path, the set of commits that changed the path with the target, and noise evidence derived from commit messages. `analyzeCandidates` creates a candidate for each record.

The score is the number of co-changing commits plus a filename similarity bonus. The bonus compares path filename tokens and is exposed separately in output. Candidates are classified as code or document; test-like paths are identified using filename and directory conventions. Noise terms and their accumulated score are retained as evidence, but do not affect ranking or filtering.

These are heuristics over Git-derived evidence. A score does not establish a semantic relationship. The present implementation does not inspect source semantics or use an LLM.

## Flow

`discoverWithLister` obtains evidence from a `PathLister`, then calls `analyzeCandidates`. Candidate preparation and rendering are described by [`/filter`](filter.md) and [`/output`](output.md).

## Related direction

The [PRD](prd.md) describes broader evidence and uncertainty goals. Future analysis may incorporate richer evidence, but any expansion should preserve inspectable evidence and avoid presenting candidates as confirmed links.
