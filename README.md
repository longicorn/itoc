# itoc

`itoc` helps discover relationships between intent and code. Its name combines **ito** (意図, “intent,” while also evoking 糸, “thread”) and **code**: “ito to code” / “ito + code.”

`itoc` aims to make code discovery from intent-bearing documents more reproducible and explainable. This traceability problem affects software development broadly, so the tool is designed to be useful on its own.

The initial focus is Forward development: use Git history and related evidence to find implementation code for a committed document. Deterministic evidence is central; AI can assist with interpretation and ambiguous cases, with results showing their rationale and uncertainty.

See the [product requirements](docs/prd.md) for the current direction.

## Component documentation

The component pages describe the current implementation behind these semantic paths:

- [`/analysis`](docs/analysis.md)
- [`/cli`](docs/cli.md)
- [`/discovery`](docs/discovery.md)
- [`/discovery/git`](docs/discovery/git.md)
- [`/filter`](docs/filter.md)
- [`/output`](docs/output.md)

They distinguish current behavior from broader product direction in the [PRD](docs/prd.md).

## Usage

Build the CLI with Go:

```sh
go build -o itoc .
```

Pass a committed repository file to discover files changed alongside it in Git history:

```sh
./itoc docs/cli/adr.md
```

Results are grouped into implementation, test, and document candidates. The default limit is 10 candidates per group; use `--limit 5` to show fewer or `--limit 0` to show all. JSON output is available with `--format json`.
Use `--min-score N` to require a minimum candidate score. When combined with `--limit`, the score filter is applied first, then the top N results per group are shown.

Use repeatable filters to narrow results:

```sh
./itoc --include-ext .go --include-path internal docs/cli/adr.md
./itoc --exclude-path vendor --exclude-ext .md docs/cli/adr.md
./itoc --exclude-candidate internal/router_test.go docs/cli/adr.md
```

Select discovery methods with `--discovery git` (default), a comma-separated list such as `--discovery git,document`, or `--discovery all`. Selected methods run concurrently and merge candidates while retaining their evidence; each additional method that finds a candidate adds one provisional point. `all` runs every registered method. Document discovery supports Markdown/MDX and reStructuredText, using explicit local links, inline code, and emphasized terms as literal clues:

```sh
./itoc --discovery document docs/cli/adr.md
```

Document mode does not infer semantic equivalents. Use `--verbose` to show the matching path, filename, or source-text evidence and its document line.
Built-in noise rules hide candidates with a noise score above 1 by default. Use `--show-noisy` to include them, or `--max-noise-score N` to set a different threshold.

Plain `--include-path` and `--exclude-path` values match anywhere in a repository-relative path. Glob patterns match the full path using Go's `filepath.Match` syntax (`*` does not cross `/`). `--exclude-candidate` hides a whole group (`implementation`, `document`, or `test`) or an exact repository-relative candidate path. Run `./itoc --help` for all options.
