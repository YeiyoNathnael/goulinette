# Goulinette (Go Static Analyzer)

Goulinette is a rule-based static analyzer for Go projects, built to enforce a strict, chaptered requirements spec.
It scans a target repository, runs selected rules, emits diagnostics in text or JSON, and returns CI-friendly exit codes.

## What it does

- Enforces **73 rules** across formatting, naming, structure, context, concurrency, testing, and more.
- Supports per-chapter and per-rule selection/disable.
- Produces deterministic diagnostics sorted by file/line/column/rule.
- Integrates external tools where required (for example `go vet`, optionally `staticcheck`).
- Uses process-local caching to avoid repeated AST/type-load work across rules in one run.

---

## Quick start

## 1) Build

```bash
go build -o goulinette ./cmd/goulinette
```

## 2) Scan current repo

```bash
./goulinette --root .
```

## 3) JSON output for CI

```bash
./goulinette --root . --format json
```

## 4) Run only one chapter

```bash
./goulinette --root . --chapter 14
```

## 5) Use strictness levels

```bash
./goulinette --root . --level 0   # bugs-only baseline
./goulinette --root . --level 1   # idiomatic default
./goulinette --root . --level 2   # opinionated
./goulinette --root . --level 3   # maximum strictness
```

## 6) Run a subset of rules

```bash
./goulinette --root . --rule CTX-01,CTX-04
```

## 7) Disable specific rules

```bash
./goulinette --root . --disable TST-03
```

## 8) Print the version

```bash
./goulinette --version
# goulinette dev
```

## 9) Explain a rule

```bash
./goulinette --explain CTX-01
# CTX-01
#
# context.Context must be the first parameter of any function that accepts
# one. This is a hard Go convention. Fix: move ctx to position 0.
```

Print rationale for every rule at once:

```bash
./goulinette --explain all
```

## 10) Suppress a specific finding inline

Add a `//goulinette:ignore` directive on the same line as, or on the
line immediately preceding, the offending code:

```go
// Suppress one rule for this line
x := doThing() //goulinette:ignore ERR-01

// Suppress multiple rules
//goulinette:ignore MAG-02 NAM-07
const appName = "myapp"

// Suppress all rules for the next line
//goulinette:ignore
blockingLegacyCall()
```

Rule IDs in the directive are case-insensitive.

---

## How invocation works (clear behavior)

### Running `./goulinette` with no flags

```bash
./goulinette
```

This is equivalent to:

```bash
./goulinette --root . --format text --level 1
```

Behavior:
- Scans the current directory recursively for Go files.
- Runs the default rule set for level `1`.
- Prints human-readable colored text diagnostics.
- Returns exit code `0`, `1`, or `2` (see Exit codes).

### `--root` behavior

- `--root .`: scan current directory.
- `--root /path/to/repo`: scan that repository/path instead.
- Discovery skips common non-source folders such as `.git`, `vendor`, and `node_modules`.

Example:

```bash
./goulinette --root /home/user/myservice
```

### What each flag does

- `--format text|json`
  - `text`: colored human-readable output.
  - `json`: machine-readable output for CI/reporting.

- `--level 0..3`
  - Selects the default strictness tier.
  - Default is `1`.

- `--chapter 1,2,14`
  - Filters active rules to the listed chapters.

- `--rule CTX-01,CTX-04`
  - Explicit rule include list.
  - Takes precedence over level-based selection.

- `--disable RULE1,RULE2`
  - Removes specific rules from the active selection after includes/filters.

- `--warnings-as-errors`
  - Treats warnings as failing diagnostics (affects exit behavior).

- `--strict-tools`
  - Fails hard when required external tools are missing instead of soft-degrading.

- `--version`
  - Prints the goulinette version string and exits.
  - The default value is `dev`; release builds set this via `-ldflags`.

- `--explain RULE-ID|all`
  - Prints the rationale for a single rule (e.g. `--explain CTX-01`) and exits.
  - Use `--explain all` to print rationale for every rule in order.

- `--max-workers N`
  - Number of goroutines used to execute rules concurrently (default 4).
  - Rules are dispatched over a buffered work channel; the pool is bounded by `N`.

- `--timeout 2m`
  - Timeout budget for external tool invocations.

### Flag precedence (important)

Selection is applied in this order:
1. Start from `--level` defaults.
2. If `--rule` is provided, use that explicit include set instead.
3. Apply `--chapter` filter.
4. Apply `--disable` exclusions.

---

## CLI reference

```text
--root string              root directory to scan (default ".")
--format string            output format: text|json (default "text")
--level int                strictness level: 0..3 (default 1)
--chapter string           comma-separated chapter numbers
--rule string              comma-separated rule IDs
--disable string           comma-separated rule IDs to disable
--warnings-as-errors       treat warnings as errors
--strict-tools             fail when required external tools are missing
--max-workers int          max concurrent rule workers (default 4)
--timeout duration         command timeout (default 2m)
--version                  print version and exit
--explain string           print rule rationale and exit (rule ID or "all")
```

Notes:
- `--level` defines the default enabled rule set (cumulative from level 0 to chosen level).
- `--rule` overrides level-based selection (explicit include list).
- `--chapter` and `--disable` are applied on top of the active include set.
- `--max-workers` controls the size of the concurrent worker pool used during rule execution.
- `--timeout` applies to tool invocations run through the internal tools wrapper.
- `--version` and `--explain` are query flags; they print and exit before any analysis runs.

### Strictness levels

- `0` (non-negotiable): compiler-invisible/runtime-risk checks (for example context misuse, copy-sensitive sync patterns, channel close ownership, comma-ok pitfalls).
- `1` (strict idiomatic): strong Go conventions and API hygiene checks; this is the default.
- `2` (opinionated): team-style preferences and stricter readability conventions.
- `3` (maximum strictness): pedagogical/library-grade structural constraints.

---

## Exit codes

- `0`: no errors (and no warnings if `--warnings-as-errors` is used)
- `1`: at least one error diagnostic, or warning promoted to error
- `2`: runtime/tooling failure (for example parse/type-load/tool execution failure)

---

## Output formats

## Text (default)

```text
path/to/file.go:12:5: error [CTX-03] nil must not be passed as context.Context
runtime: FMT-02: tool "go" not found in PATH
```

Text output is ANSI-colorized for readability:
- **Severity**: `error` in **bold red**, `warning` in yellow.
- **Broken rule ID**: highlighted with an actual color (not just bold).
- **File location**: rendered in dim gray to reduce visual noise.

Fixed chapter color mapping (not runtime-random):
- Rule IDs are color-grouped by chapter prefix and always use the same preselected color.
- Colors are enforced in code and do not change between runs.

Chapter palette (ANSI 256 colors):
- Chapter 1 (`FMT-*`) → `39`
- Chapter 2 (`NAM-*`) → `208`
- Chapter 3 (`VAR-*`) → `45`
- Chapter 4 (`CTL-*`) → `171`
- Chapter 5 (`FUN-*`) → `220`
- Chapter 6 (`ERR-*`) → `196`
- Chapter 7 (`TYP-*`) → `51`
- Chapter 8 (`STR-*`) → `99`
- Chapter 9 (`DOC-*`) → `34`
- Chapter 10 (`SLC-*`) → `214`
- Chapter 11 (`CON-*`) → `129`
- Chapter 12 (`CER-*`) → `44`
- Chapter 13 (`LIM-*`) → `177`
- Chapter 14 (`CTX-*`) → `81`
- Chapter 15 (`IMP-*`) → `203`
- Chapter 16 (`RES-*`) → `141`
- Chapter 17 (`SAF-*`) → `40`
- Chapter 18 (`MAG-*`) → `201`
- Chapter 19 (`TST-*`) → `93`

## JSON

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

---

## Inline suppression

Any diagnostic can be silenced without changing the analysis by placing a
`//goulinette:ignore` comment directive in the source file.

### Directive syntax

```
//goulinette:ignore [RULE-ID ...]
```

- **No rule IDs** — suppresses all findings on the associated line.
- **One or more rule IDs** (space-separated) — suppresses only those rules.
- Rule IDs are matched case-insensitively.

### Placement rules

The directive may appear either:
- **On the same line** as the flagged code (trailing comment), or
- **On the immediately preceding line** (standalone comment).

```go
// Same-line suppression
result := riskyOp() //goulinette:ignore ERR-01

// Preceding-line suppression
//goulinette:ignore MAG-02
const retries = 3

// Suppress multiple rules at once
//goulinette:ignore CON-02 CON-03
go func() { /* ... */ }()

// Suppress everything on the next line
//goulinette:ignore
legacyHack()
```

### When to use it

Use inline suppression sparingly for narrow cases where the rule does not apply in context
(e.g. a deliberate constant that is intentionally not named, a generated file pattern,
or a vendored snippet). Prefer fixing the root cause where possible.

---

## Rule catalog (implemented)

All requirement rules in `goulinette_requirements.md` are implemented and registered.

### 1) Formatting & Tooling
`FMT-01`, `FMT-02`, `FMT-03`

### 2) Naming Conventions
`NAM-01`, `NAM-02`, `NAM-03`, `NAM-04`, `NAM-05`, `NAM-06`, `NAM-07`

### 3) Variable Declarations
`VAR-01`, `VAR-02`, `VAR-03`, `VAR-04`

### 4) Control Structures
`CTL-01`, `CTL-02`, `CTL-03`, `CTL-04`

### 5) Functions & Return Values
`FUN-01`, `FUN-02`, `FUN-03`, `FUN-04`

### 6) Error Handling
`ERR-01`, `ERR-02`, `ERR-03`, `ERR-04`, `ERR-05`, `ERR-06`, `ERR-07`, `ERR-08`

### 7) Types, Interfaces & Pointers
`TYP-01`, `TYP-02`, `TYP-03`, `TYP-04`, `TYP-05`, `TYP-06`, `TYP-07`

### 8) Methods & Structs
`STR-01`, `STR-02`, `STR-03`, `STR-04`

### 9) Packages & Documentation
`DOC-01`, `DOC-02`, `DOC-03`, `DOC-04`, `DOC-05`

### 10) Slices & Collections
`SLC-01`

### 11) Concurrency
`CON-01`, `CON-02`, `CON-03`

### 12) Custom Errors
`CER-01`, `CER-02`, `CER-03`

### 13) Structural Limits
`LIM-01`, `LIM-02`, `LIM-03`, `LIM-04`

### 14) Context Handling
`CTX-01`, `CTX-02`, `CTX-03`, `CTX-04`

### 15) Import Organization
`IMP-01`, `IMP-02`, `IMP-03`

### 16) Resource Management
`RES-01`, `RES-02`

### 17) Concurrency Safety
`SAF-01`, `SAF-02`

### 18) Magic Values
`MAG-01`, `MAG-02`

### 19) Testing
`TST-01`, `TST-02`, `TST-03`

---

## Architecture

## Execution pipeline

1. Parse CLI flags into `internal/config.Config`
2. Handle query flags: `--version` prints and exits; `--explain` prints rule rationale and exits
3. Discover Go files under `--root` (excluding `.git`, `vendor`, `node_modules`)
4. Build rule context (`root`, files, tool strictness)
5. Select rules via chapter/include/disable filters
6. Dispatch selected rules to a bounded `--max-workers` goroutine pool via a buffered work channel
7. Aggregate diagnostics + runtime errors from the results channel
8. Apply inline suppression: remove any finding covered by a `//goulinette:ignore` directive
9. Sort diagnostics deterministically (file → line → column → rule ID)
10. Print report (`text` or `json`) and return exit code

## Key packages

- `cmd/goulinette`: CLI entry point; handles `--version` / `--explain` before delegating to `app.Runner`
- `internal/app`: orchestration and run loop; owns the worker-pool dispatch and suppress-filter step
- `internal/config`: flag parsing; defines `Settings` (includes `ExplainRule`, `PrintVersion`, `MaxWorkers`)
- `internal/discovery`: repository file discovery
- `internal/rules`: rule implementations, helpers, registry, and `explain.go` (rationale map for all 62 rules)
- `internal/suppress`: `Filter()` — removes diagnostics covered by `//goulinette:ignore` directives
- `internal/version`: build-time version string (`Current`); injectable via `-ldflags` at release time
- `internal/tools`: external command execution + diagnostics parsing
- `internal/diag`: shared diagnostic/result model
- `internal/report`: text/json rendering

---

## Performance model

Goulinette runs rules concurrently through a bounded worker pool and avoids repeated expensive work via per-run caches:

- **Worker pool**: `--max-workers` goroutines (default 4) pull from a buffered work channel; results are collected over a pre-sized results channel. The pool respects `context.Context` cancellation.
- **AST cache**: repeated `parseFiles(...)` calls over the same file set reuse parsed structures.
- **Typed package cache**: repeated `loadTypedPackages(...)` calls per root reuse type-loaded package graphs.
- **Per-run reset**: caches are reset at app run start, preventing stale data across independent runs.

This means parse/type-load work is shared across all concurrent rules, while `--max-workers` lets you trade off CPU saturation against wall-clock time.

### Cache benchmarks

Benchmarks for both cache paths live in `internal/rules/cache_bench_test.go`:

```bash
go test ./internal/rules -bench BenchmarkParseFiles -benchmem
go test ./internal/rules -bench BenchmarkCacheKey -benchmem
```

Typical results on a modern laptop:

| Benchmark | ns/op |
|---|---|
| `BenchmarkParseFilesCacheMiss` | ~420 000 |
| `BenchmarkParseFilesCacheHit` | ~5 000 |
| `BenchmarkCacheKeyGeneration` | < 1 000 |

---

## External tool behavior

Some rules call external tools through `internal/tools`.

- If a tool is missing or command fails, diagnostics/runtime behavior depends on rule implementation and `--strict-tools`.
- Commands are timeout-bound by config (`--timeout`).

Common tools involved:
- `go` (for `go vet`, `gofmt` checks)
- `staticcheck` (warning-level enrichment when available)

---

## Development workflow

## Run all tests

```bash
go test ./...
```

## Run tests with race detector

```bash
go test -race -count=1 ./...
```

## Run one package

```bash
go test ./internal/rules
```

## Run cache benchmarks

```bash
go test ./internal/rules -bench BenchmarkParseFiles -benchmem
go test ./internal/rules -bench BenchmarkCacheKey -benchmem
```

## Build binary

```bash
go build -o goulinette ./cmd/goulinette
```

## Build a versioned release binary

```bash
go build -ldflags "-X github.com/YeiyoNathnael/goulinette/internal/version.Current=1.2.3" \
  -o goulinette ./cmd/goulinette
./goulinette --version
# goulinette 1.2.3
```

## Self-lint gate (must pass before every commit)

```bash
./goulinette --level=3 --format json | python3 -c \
  "import sys,json; d=json.load(sys.stdin); assert len(d['diagnostics'])==0, d['diagnostics']"
```

## CI

GitHub Actions runs on every push to `main` / `fix/**` / `feat/**` and on all pull requests:

| Job | Command |
|---|---|
| Build | `go build ./...` |
| Vet | `go vet ./...` |
| Test (race) | `go test -race -count=1 ./...` |
| Self-lint | build binary → `./goulinette --level=3 --format json` → assert 0 findings |

See [`.github/workflows/ci.yml`](.github/workflows/ci.yml) for the full workflow definition.

## Typical feature branch flow

```bash
git checkout -b feat/my-change
# edit
go test -race ./...
./goulinette --level=3 --format json   # must report 0 findings
git add .
git commit -m "feat: my change"
git push -u origin feat/my-change
```

---

## Design notes and limits

- Rules are intentionally independent files for maintainability.
- Some checks use conservative heuristics to avoid over-claiming certainty in static analysis.
- Typed analysis requires successful package loading; unresolved build contexts can degrade specific typed rules.
- `TST-03` intentionally flags direct `time.Sleep` calls in `_test.go` files only.

---

## Troubleshooting

## "tool not found in PATH"
Install required tool (`go`, optionally `staticcheck`) and retry.

## "type loading failed"
Ensure target project module/build context is valid (`go mod tidy`, buildable package graph).

## Unexpected non-zero exit code
- Use `--format json` to inspect diagnostics and runtime errors clearly.
- Check whether `--warnings-as-errors` is enabled.

---

## Requirements source of truth

Rule intent and chapter definitions are specified in:

- `goulinette_requirements.md`

If a behavior question arises, this document takes precedence over inferred style.
