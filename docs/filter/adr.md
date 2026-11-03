# ADR: Candidate Filtering Boundary

## Status

Current implementation direction; defaults and filtering rules remain subject to evaluation.

## Context and interface

The [filter PRD](prd.md) requires selection and ranking of discovered candidates for presentation. Discovery collects evidence; [analysis](../analysis/adr.md) scores and classifies it. After analysis, multi-method results may be merged and optional test discovery may expand them. Filtering receives the resulting candidates before output, not raw discovery evidence. See the [discovery ADR](../discovery/adr.md).

The current interface is `prepareCandidates([]candidate, candidateOptions) []candidate`. Each candidate carries a repository-relative path, score, classification, noise score and supporting evidence; options specify selection criteria and limits. The returned slice contains the selected and ordered candidates for rendering. Filtering uses these fields but does not calculate scores or classifications, interpret evidence, or determine whether a relationship is correct.

## Decision: filter responsibilities

Provide a common preparation stage independent of the discovery method. Apply user selection and exclusion criteria, reduce noise, rank the remaining candidates and bound the number shown per group. Preserve the selected candidates' evidence for output so users can assess them. The CLI owns option parsing and validation; output owns presentation.

## Current behavior

1. Apply extension and path inclusion and exclusion filters, then explicit candidate or group exclusions.
2. Exclude candidates above the maximum noise score unless `--show-noisy` is set; apply the minimum score when greater than zero.
3. Sort by score descending, breaking ties by path ascending, then apply the per-group limit. Test-like candidates form their own group; other candidates use their classification. Sorting first retains the highest-ranked candidates in each group.

The CLI defaults to a maximum noise score of 1, a minimum score of 0 and a limit of 10 per group. A limit of 0 is unlimited. Plain path patterns match substrings of repository-relative paths; patterns with glob metacharacters use `filepath.Match` against the full path, where `*` does not cross `/`. Extensions are compared in lower case. Explicit exclusions accept a group name or an exact normalized repository-relative path.

## Consequences and open questions

A shared pipeline gives discovery methods consistent selection and makes ordering reproducible. Thresholds and limits can hide relevant candidates, and their usefulness depends on heuristic scores and classifications. Users can adjust the thresholds, include noisy candidates and remove the limit, but still need to assess the evidence.

How should defaults, matching rules and group limits be evaluated against missed candidates and review effort? Should the candidate and options interface evolve as discovery sources and analysis change?
