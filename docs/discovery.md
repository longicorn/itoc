# /discovery

## Purpose

`/discovery` defines the shared boundary for obtaining related-path evidence and turning it into candidates, independent of the evidence source.

## Current behavior

`PathLister` is the adapter contract. Its `List` method accepts the input path and returns co-change evidence, the resolved target path, and an error. `discoverWithLister` invokes the adapter and passes its evidence and target to `/analysis` through `analyzeCandidates`.

The default `discover` currently chooses `GitPathLister`. The interface permits another evidence source to be supplied, but no non-Git lister is currently implemented. The contract currently returns Git-shaped co-change evidence; broadening it for other evidence types would require an explicit model decision.

## Related paths

- [`/discovery/git`](discovery/git.md) implements the current Git-backed adapter.
- [`/analysis`](analysis.md) converts evidence into scored candidates.
