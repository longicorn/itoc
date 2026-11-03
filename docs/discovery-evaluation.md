# Discovery improvement evaluation

## Purpose

Evaluate improvements to test discovery and document-clue ranking before committing to algorithms, weights, thresholds, or defaults. Results are candidates supported by evidence, not confirmed semantic relationships. Reproducibility does not guarantee correctness.

## Test candidates from implementation candidates

Start with discovered implementation files and compare three independent kinds of evidence: identifier matches, file naming and placement, and Git co-changes between implementation and test files. Keep the originating implementation and the evidence for each candidate. Merge duplicate candidates without losing their evidence; do not apply an independent-method agreement bonus automatically to a test derived from another candidate. Explore only one step from implementation candidates to avoid recursive expansion.

Try Universal Ctags to extract definitions from implementation files, then search test files for those identifiers. Record the definition and matching test location and distinguish textual matches from resolved references. Assess ambiguous or common names, comments, language-specific syntax, and indirect calls. Ctags availability and version, supported languages, and failure or missing-tool behavior must be recorded. GNU Global support is deferred for this trial. Evaluate naming conventions per language and Git history separately, including missing or renamed files and bulk changes.

Compare the approaches individually and combined. Include tests with matching and different names, shared tests, unrelated similar names, and implementations with no corresponding test. Consider how filters, noise thresholds, limits and scores affect both the originating implementation and derived tests; decide their ordering based on results rather than assuming inherited scores.

## Document-clue ranking

Record match type, source document location, matched candidate and term frequency across indexed files. Distinguish at least explicit existing paths, exact filenames, filename substrings and source-text matches. Compare lowering the contribution of widespread terms without excluding candidates solely for frequency. Repeated occurrences of the same clue should not yield unlimited credit; evaluate a cap for multiple distinct clues as well. A direct path is evidence of a reference, not proof of implementation.

First compare rankings within document discovery against the current behavior. Then evaluate interaction with Git scores, the cross-method agreement bonus, noise rules and filters. Do not fix numeric weights or thresholds until the results support them.

## Evaluation procedure

1. Select several repositories and logical targets, including documents with rich and sparse history and with explicit references. Candidate repositories include Discourse (Rails), Wagtail (Django), and Spinel (C); Nagi's Runner documents provide an additional initial case.
2. Record revisions, inputs, tool versions, settings, index scope and exclusions. `--limit 0` does not disable noise filtering. Document unavailable methods, errors, and index freshness.
3. Manually review relevant implementation and test candidates, recording reasons and uncertain cases. AI output can inform review but is not a ground truth. Include negative cases and known missed relationships.
4. Compare baseline and proposed methods separately and together: relevant candidates found, omissions, unrelated candidates, ranking, evidence, and execution time. Preserve examples and unresolved judgments. Use results to choose implementation scope and defaults.

The initial two-file feasibility checks found that the installed Universal Ctags could extract example Ruby, Python, and C definitions, while the installed GNU Global default configuration indexed the C example but not the Ruby or Python examples. This is not a repository-wide accuracy or performance evaluation.
