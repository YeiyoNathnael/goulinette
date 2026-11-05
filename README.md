![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)
![Go Report Card](https://goreportcard.com/badge/github.com/YeiyoNathnael/goulinette)
![CI](https://github.com/YeiyoNathnael/goulinette/actions/workflows/ci.yml/badge.svg)
![License](https://img.shields.io/github/license/YeiyoNathnael/goulinette)
![Release](https://img.shields.io/github/v/release/YeiyoNathnael/goulinette)
[![codecov](https://codecov.io/gh/YeiyoNathnael/goulinette/branch/main/graph/badge.svg)](https://codecov.io/gh/YeiyoNathnael/goulinette)
![License](https://img.shields.io/github/license/YeiyoNathnael/goulinette)

# Goulinette

Goulinette is a static analyzer for Go. It checks a repository against 73 rules covering formatting, naming, error handling, concurrency, context use, testing and project structure, then reports diagnostics as text or JSON with exit codes suited to CI.

## Features

- 73 rules grouped into 19 chapters
- Four strictness levels
- Selection by chapter or rule, with per-rule exclusions
- Inline suppression with `//goulinette:ignore`
- Rule explanations from the command line
- Deterministic output sorted by file, line, column and rule
- Concurrent rule execution with a configurable worker pool
- Integration with `go vet` and, when installed, `staticcheck`

## Installation

Requires Go 1.24 or later.

```bash
go install github.com/YeiyoNathnael/goulinette/cmd/goulinette@latest
```

To build from source:

```bash
git clone https://github.com/YeiyoNathnael/goulinette.git
cd goulinette
go build -o goulinette ./cmd/goulinette
```

## Usage

```bash
# Scan the current directory
goulinette

# Scan another repository
goulinette --root /path/to/repo

# Produce JSON output
goulinette --format json

# Run one chapter
goulinette --chapter 14

# Run specific rules
goulinette --rule CTX-01,CTX-04

# Exclude a rule
goulinette --disable TST-03

# Explain a rule
goulinette --explain CTX-01
```

Running `goulinette` with no flags is equivalent to:

```bash
goulinette --root . --format text --level 1
```

## Command-line reference

| Flag | Default | Description |
|---|---|---|
| `--root` | `.` | Directory to scan |
| `--format` | `text` | Output format: `text` or `json` |
| `--level` | `1` | Strictness level, from 0 to 3 |
| `--chapter` | | Comma-separated chapter numbers |
| `--rule` | | Comma-separated rule IDs |
| `--disable` | | Comma-separated rule IDs to exclude |
| `--warnings-as-errors` | `false` | Treats warnings as errors |
| `--strict-tools` | `false` | Fails when a required external tool is missing |
| `--max-workers` | `4` | Maximum concurrent rule workers |
| `--timeout` | `2m` | Timeout for external tools |
| `--version` | | Prints the version and exits |
| `--explain` | | Prints the rationale for a rule ID, or for `all`, and exits |

Rule selection is applied in this order:

1. `--level` sets the default rule set.
2. `--rule` replaces that set with an explicit list.
3. `--chapter` filters the active rules.
4. `--disable` removes rules.

## Strictness levels

Levels are cumulative.

| Level | Name | Scope |
|---:|---|---|
| 0 | Non-negotiable | Runtime risks such as context misuse, copied sync types and unsafe channel closes |
| 1 | Strict idiomatic | Go conventions and API hygiene. This is the default |
| 2 | Opinionated | Team style and readability conventions |
| 3 | Maximum strictness | Structural constraints for libraries and teaching |

## Rules

| Chapter | Category | Rules |
|---:|---|---|
| 1 | Formatting and tooling | `FMT-01` to `FMT-03` |
| 2 | Naming conventions | `NAM-01` to `NAM-07` |
| 3 | Variable declarations | `VAR-01` to `VAR-04` |
| 4 | Control structures | `CTL-01` to `CTL-04` |
| 5 | Functions and return values | `FUN-01` to `FUN-04` |
| 6 | Error handling | `ERR-01` to `ERR-08` |
| 7 | Types, interfaces and pointers | `TYP-01` to `TYP-07` |
| 8 | Methods and structs | `STR-01` to `STR-04` |
| 9 | Packages and documentation | `DOC-01` to `DOC-05` |
| 10 | Slices and collections | `SLC-01` |
| 11 | Concurrency | `CON-01` to `CON-03` |
| 12 | Custom errors | `CER-01` to `CER-03` |
| 13 | Structural limits | `LIM-01` to `LIM-04` |
| 14 | Context handling | `CTX-01` to `CTX-04` |
| 15 | Import organization | `IMP-01` to `IMP-03` |
| 16 | Resource management | `RES-01`, `RES-02` |
| 17 | Concurrency safety | `SAF-01`, `SAF-02` |
| 18 | Magic values | `MAG-01`, `MAG-02` |
| 19 | Testing | `TST-01` to `TST-03` |

The full specification is in [`goulinette_requirements.md`](goulinette_requirements.md).

## Output

### Text

```text
path/to/file.go:12:5: error [CTX-03] nil must not be passed as context.Context
```

### JSON

```json
{
  "diagnostics": [
    {
      "ruleId": "CTX-03",
      "severity": "error",
      "message": "nil must not be passed as context.Context",
      "position": {"file": "x.go", "line": 12, "col": 5},
      "hint": "use context.Background() or context.TODO()"
    }
  ],
  "runtimeErrors": []
}
```

## Exit codes

| Code | Meaning |
|---:|---|
| 0 | No errors |
| 1 | At least one error, or a warning when `--warnings-as-errors` is set |
| 2 | Runtime or tooling failure |

## Inline suppression

Place a directive on the flagged line or on the line above it. Rule IDs are case-insensitive.

```go
// Suppress one rule
result := riskyOp() //goulinette:ignore ERR-01

// Suppress several rules
//goulinette:ignore MAG-02 NAM-07
const appName = "myapp"

// Suppress all rules on the next line
//goulinette:ignore
legacyCall()
```

## Architecture

Each run follows these steps:

1. Parse flags into the configuration.
2. Discover Go files under the root, skipping `.git`, `vendor` and `node_modules`.
3. Build the shared rule context.
4. Select rules from the level, chapter, include and disable settings.
5. Dispatch rules to a bounded worker pool over a buffered channel.
6. Collect diagnostics and runtime errors.
7. Apply inline suppressions.
8. Sort the diagnostics and write the report.

A process-local cache shares parsed AST and type information across rules within a run.

```text
cmd/goulinette/       Command-line entry point
internal/app/         Run orchestration
internal/config/      Flag parsing and settings
internal/diag/        Diagnostic model
internal/discovery/   File discovery
internal/report/      Text and JSON reporters
internal/rules/       Rule implementations
```

## Development

```bash
go build ./...
go vet ./...
go test -race ./...
```

CI builds the project, runs `go vet`, runs the tests with the race detector, reports coverage to Codecov, and runs Goulinette against its own source.

## License

[MIT](LICENSE)
