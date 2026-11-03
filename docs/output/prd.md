# Output Product Requirements

## Purpose

Output renders prepared candidates for people or downstream tools. It does not discover, score, or filter candidates.

## Requirements

- Present candidates in text and JSON formats.
- Group text results by implementation, test, and document, and allow users to request detailed evidence.
- Provide candidate fields in JSON for downstream tools.
- Communicate when no candidates are found.

[CLI](../cli/prd.md) selects the format and writes the result, [filter](../filter/prd.md) prepares the candidate set, and [analysis](../analysis/prd.md) defines candidate fields and evidence. Candidates are clues, not proof of a relationship.

## Current behavior

`renderCandidates` selects text or JSON rendering. Text output groups candidates under implementation, test, and document headings. Detailed evidence strings, including co-change commit IDs and document references, are hidden by default and shown with `--verbose`. Empty text output says that no code candidates were found.

JSON output is an array of candidate records with path, score, co-changed commit count, filename bonus, kind, noise score, and optional noise terms. Test-like candidates have kind `test`. The CLI writes rendered bytes to stdout; output-file support is not currently implemented.
