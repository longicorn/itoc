# /discovery

## Purpose

`/discovery` defines the shared boundary for obtaining related-path evidence and turning it into candidates, independent of the evidence source.

## Current behavior

The CLI accepts one or more comma-separated method names, or `all`. The method registry is the single source for available implementations; `all` selects every registered method. Selected methods run concurrently, and results are merged by repository-relative path with evidence retained. Each additional method finding the same candidate adds a provisional +1 agreement bonus. This is an initial heuristic to evaluate, not a calibrated probability. Errors identify which method failed.

`DocumentPathLister` dispatches by document extension to a format extractor. Markdown/MDX and reStructuredText are supported. Extractors recognize links, inline code/literals, and emphasis/roles. Explicit local paths resolve directly; emphasized or code-formatted terms are matched against code file names and source contents. This is deliberately literal and does not infer semantic equivalents. If the input extension is unsupported, document discovery reports itself as not applicable; in a multi-method run this skips only document discovery, while document-only mode reports the unsupported format.

Document search walks the repository once to build a source index. Source files are read by a bounded worker pool, then all extracted terms are matched against the in-memory index. This avoids rescanning and rereading the repository once per term.

## Related paths

- [`/discovery/git`](discovery/git.md) implements the current Git-backed adapter.
- [`/analysis`](analysis.md) converts evidence into scored candidates.
