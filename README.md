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
--max-workers int          max analysis workers (default 4)
--timeout duration         command timeout (default 2m)
```

Notes:
- `--level` defines the default enabled rule set (cumulative from level 0 to chosen level).
- `--rule` overrides level-based selection (explicit include list).
- `--chapter` and `--disable` are applied on top of the active include set.
- `--max-workers` is parsed and carried in config for future/extended scheduling controls.
- `--timeout` applies to tool invocations run through the internal tools wrapper.

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
2. Discover Go files under `--root` (excluding `.git`, `vendor`, `node_modules`)
3. Build rule context (`root`, files, tool strictness)
4. Select rules via chapter/include/disable filters
5. Execute selected rules sequentially
6. Aggregate diagnostics + runtime errors
7. Sort diagnostics deterministically
8. Print report (`text` or `json`) and return exit code

## Key packages

- `cmd/goulinette`: CLI entry point
- `internal/app`: orchestration and run loop
- `internal/config`: flag parsing
- `internal/discovery`: repository file discovery
- `internal/rules`: rule implementations, helpers, registry
- `internal/tools`: external command execution + diagnostics parsing
- `internal/diag`: shared diagnostic/result model
- `internal/report`: text/json rendering

---

## Performance model

Goulinette currently runs rules sequentially, but avoids repeated expensive setup work via helper-level caches:

- **AST cache**: repeated `parseFiles(...)` calls over the same file set reuse parsed structures.
- **Typed package cache**: repeated `loadTypedPackages(...)` calls per root reuse type-loaded package graphs.
- **Per-run reset**: caches are reset at app run start, preventing stale data across independent runs.

This preserves simple rule files while significantly reducing duplicate parse/type-load overhead.

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

## Run one package

```bash
go test ./internal/rules
```

## Build binary

```bash
go build -o goulinette ./cmd/goulinette
```

## Typical feature branch flow

```bash
git checkout -b feat/my-change
# edit
# test
git add .
git commit -m "my change"
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
