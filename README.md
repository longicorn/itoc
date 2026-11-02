# itoc

`itoc` helps discover relationships between intent and code. Its name combines **ito** (意図, “intent,” while also evoking 糸, “thread”) and **code**: “ito to code” / “ito + code.”

The primary goal is to support [Nagi](https://github.com/longicorn/nagi), especially by making code discovery from semantic paths more reproducible and explainable. The underlying traceability problem also affects software development more broadly, so `itoc` aims to be useful as a standalone tool as well.

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

Plain `--include-path` and `--exclude-path` values match anywhere in a repository-relative path. Glob patterns match the full path using Go's `filepath.Match` syntax (`*` does not cross `/`). `--exclude-candidate` hides a whole group (`implementation`, `document`, or `test`) or an exact repository-relative candidate path. Run `./itoc --help` for all options.
