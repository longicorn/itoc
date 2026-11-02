# /discovery

## Purpose

`/discovery` defines the shared boundary for obtaining related-path evidence and turning it into candidates, independent of the evidence source.

## Current behavior

The CLI supports `--discovery git`, `--discovery document`, and `--discovery both`. `both` runs the two independent methods concurrently, collects their results through a channel, and merges candidates by repository-relative path while retaining evidence from each method. It does not yet alter scores when methods agree; this mode is for comparison and evaluation. Errors identify which method failed.

`DocumentPathLister` dispatches by document extension to a format extractor. Markdown/MDX and reStructuredText are supported. Extractors recognize links, inline code/literals, and emphasis/roles. Explicit local paths resolve directly; emphasized or code-formatted terms are matched against code file names and source contents. This is deliberately literal and does not infer semantic equivalents.

Document search walks the repository once to build a source index. Source files are read by a bounded worker pool, then all extracted terms are matched against the in-memory index. This avoids rescanning and rereading the repository once per term.

## Related paths

- [`/discovery/git`](discovery/git.md) implements the current Git-backed adapter.
- [`/analysis`](analysis.md) converts evidence into scored candidates.
