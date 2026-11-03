# itoc

`itoc` helps discover relationships between intent and code. Its name combines **ito** (意図, “intent,” while also evoking 糸, “thread”) and **code**: “ito to code” / “ito + code.”

`itoc` aims to make code discovery from intent-bearing documents more reproducible and explainable. This traceability problem affects software development broadly, so the tool is designed to be useful on its own.

The initial focus is Forward development: use Git history and related evidence to find implementation code for a committed document. Deterministic evidence is central; AI can assist with interpretation and ambiguous cases, with results showing their rationale and uncertainty.

See the [product requirements](docs/prd.md) for the current direction.

## Installation

Install Go 1.23 or later and Git, then build and install the CLI from the repository root:

```sh
go install .
```

Make sure the Go binary directory (`$(go env GOPATH)/bin`, unless `GOBIN` is set) is on your `PATH`. To build a binary in the current directory instead, run `go build -o itoc .`.

For additional test discovery clues, install [Universal Ctags](https://ctags.io/) with JSON output support and make `ctags` available on `PATH`. Without it, filename and Git co-change discovery still work, but identifier matching is skipped.

## Usage

Run `itoc` from a Git repository. Start with a repository file path, or search tracked file contents for a keyword:

```sh
itoc <path>
itoc --search <keyword>
```

With `--search`, if multiple files match, select the starting file from the numbered list. The default Git discovery uses the selected file's history to find related candidates. Run `itoc --help` for all options.

## More options

Results are grouped into implementation, test, and document candidates. The default limit is 10 candidates per group; use `--limit 5` to show fewer or `--limit 0` to show all. JSON output is available with `--format json`.
Test candidates are also discovered from implementation candidates by filename and Git co-changes. [Ctags identifier matches](docs/discovery/ctags/prd.md) provide additional clues. These mechanical matches can produce false positives and miss relevant files; a match does not prove test coverage. Use `--exclude-candidate test` to hide test candidates, or `--discover-tests=false` to disable this expansion.

Use `--min-score N` to require a minimum candidate score. When combined with `--limit`, the score filter is applied first, then the top N results per group are shown.

Use repeatable filters to narrow results:

```sh
itoc --include-ext .go --include-path internal docs/cli/adr.md
itoc --exclude-path vendor --exclude-ext .md docs/cli/adr.md
itoc --exclude-candidate internal/router_test.go docs/cli/adr.md
```

Select discovery methods with `--discovery git` (default), a comma-separated list such as `--discovery git,document`, or `--discovery all`. Selected methods run concurrently and merge candidates while retaining their evidence; each additional method that finds a candidate adds one provisional point. `all` runs every registered method. Document discovery supports Markdown/MDX and reStructuredText, using explicit local links, inline code, and emphasized terms as literal clues:

```sh
itoc --discovery document docs/cli/adr.md
```

Document mode does not infer semantic equivalents. Use `--verbose` to show the matching path, filename, or source-text evidence and its document line.
Built-in noise rules hide candidates with a noise score above 1 by default. Use `--show-noisy` to include them, or `--max-noise-score N` to set a different threshold.

Plain `--include-path` and `--exclude-path` values match anywhere in a repository-relative path. Glob patterns match the full path using Go's `filepath.Match` syntax (`*` does not cross `/`). `--exclude-candidate` hides a whole group (`implementation`, `document`, or `test`) or an exact repository-relative candidate path.

## Documentation

See [docs/](docs/) for product requirements and component documentation.
