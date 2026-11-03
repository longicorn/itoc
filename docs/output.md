# /output

## Purpose

`/output` renders prepared candidates for people or downstream tools. Rendering does not discover, score, or filter candidates.

## Current behavior

`renderCandidates` selects text or JSON rendering. Text output groups candidates under implementation, test, and document headings. Detailed evidence strings, including co-change commit IDs and document references, are hidden by default and shown with `--verbose`. Empty text output says that no code candidates were found.

JSON output is an array of candidate records with path, score, co-changed commit count, filename bonus, kind, noise score, and optional noise terms. Test-like candidates have kind `test`. The CLI writes rendered bytes to stdout; output-file support is not currently implemented.

## Related paths

- [`/cli`](cli/prd.md) selects the format and writes the result.
- [`/filter`](filter.md) prepares the candidate set before rendering.
- [`/analysis`](analysis/prd.md) defines the candidate fields and evidence.
