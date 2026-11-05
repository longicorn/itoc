# ADR: Git-Based Relationship Discovery

## Status

Accepted. Relocation discovery is a first implementation to evaluate.

## Decision

For an existing tracked input, collect co-change evidence from current-path Git history. `--follow` enables Git rename-following history for primary and test discovery; it is disabled by default. Commit messages and changed paths provide inspectable clues, not proof of a relationship. A failed message retrieval leaves no message evidence; a failed diff skips that commit.

For an input missing from the working tree, retrieve its most recent historical content and compare tracked current code and document files. Record matching filenames, identical contents, and shared Ctags definitions as separate relocation clues, retaining multiple candidates. Symbol matches are limited to code with the same extension. Missing Ctags produces a warning. No match does not prove deletion; historical content must be available. The results are relocation candidates, not co-change history. Future tree-sitter use requires evaluation.

## Limitations

Default history does not follow earlier names. Ctags may miss symbols or match unrelated names. Matching is limited to tracked files and does not identify every move or split. Git execution remains required in this phase. Optional follow uses Git CLI and does not turn symbol matches into confirmed moves. Ranking and scoring are separate from discovery.
