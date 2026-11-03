# ADR: CLI Workflow Coordination

## Status

Current implementation direction; keyword search and multi-method scoring remain experimental or provisional.

## Context

The [CLI requirements](prd.md) call for a single entry point to discover and inspect repository artifacts. Discovery produces candidates and evidence rather than confirmed relationships. Users need to select methods, reduce noise, and choose how much detail to inspect without each discovery method implementing its own interface.

## Decision

- Make the CLI responsible for validating input and coordinating discovery, optional test-candidate expansion, candidate preparation, and rendering. Delegate their substantive behavior to [discovery](../discovery.md), [filtering](../filter.md), and [output](../output.md).
- Accept exactly one repository file path, or an explicitly supplied `--search` keyword, as the starting point. Do not infer a keyword from a failed path. The search connection is experimental; its selection behavior and open questions are recorded in the [keyword ADR](../keyword/adr.md).
- Select discovery methods explicitly, defaulting to Git. Run multiple selected methods concurrently and merge candidates by path while retaining their evidence. The current one-point bonus for each additional method is provisional and does not establish a relationship.
- Enable test-candidate expansion by default, with an option to disable it. Allow repeatable candidate filters and controls for noise, minimum score, and per-group result limits. Apply the minimum score before the limit.
- Render text by default or JSON on request, with detailed evidence available through verbose output. Write results to standard output; send warnings, prompts, and errors to standard error. Invalid input and workflow errors result in an unsuccessful exit.

## Consequences

A shared workflow gives methods consistent filtering and presentation while retaining evidence for human assessment. Parallel methods can provide complementary clues, but their agreement may reflect correlated evidence. The score bonus needs evaluation. Test expansion can depend on external tools; when ctags is unavailable, identifier-based expansion is skipped with a warning.

## Open questions

- How should agreement between discovery methods be evaluated and weighted?
- How should keyword search candidates be selected and passed into discovery as that experiment evolves?
- Which defaults and controls best balance missed candidates against review effort?
