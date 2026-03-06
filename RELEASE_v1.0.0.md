# Goulinette v1.0.0 — Cobalt Baseline

A strict, chaptered Go static analyzer with deterministic output and CI-ready behavior.

---

## Highlights

- Complete rule catalog implemented and registered from `goulinette_requirements.md`.
- Deterministic diagnostics sorted by file, line, column, and rule ID.
- Strictness tiers via `--level 0..3` for gradual adoption.
- Rule selection controls: `--chapter`, `--rule`, and `--disable`.
- Fixed chapter-based color mapping in text output for better readability.
- `error` severity rendered in **bold red** to increase visibility.
- Performance optimization via per-run caching for parse/type-load helpers.

---

## What’s included

- 73 rules across 19 chapters, including:
  - formatting/tooling
  - naming and declarations
  - control flow and functions
  - error handling and typing
  - context, concurrency, resource handling
  - imports, magic values, testing
- Output formats:
  - `text` (default)
  - `json`
- CI-friendly exits:
  - `0` no errors
  - `1` diagnostics failed policy
  - `2` runtime/tooling failure

---

## Install / Build

```bash
git clone https://github.com/YeiyoNathnael/goulinette.git
cd goulinette
go build -o goulinette ./cmd/goulinette
```

Quick run:

```bash
./goulinette --root .
```

---

## CLI quick reference

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

---

## Upgrade notes

- If you used earlier snapshots/dev commits:
  - prefer `--level 1` as the initial baseline for team rollout.
  - use `--rule` for focused migration or incremental adoption.
  - use `--disable` for temporary exceptions while burning down findings.
- Text output now applies fixed chapter colors and bold red `error` labels.
- If your CI parser expects plain text tokens only, use JSON output:

```bash
./goulinette --root . --format json
```

---

## Breaking changes

- None (first stable `v1.0.0` release).

---

## Known limitations

- Some rules rely on conservative static heuristics and may require local tuning/workflow adaptation.
- Typed checks depend on successful package loading in the target module context.
- External-tool-backed checks depend on tool availability unless handled via policy flags.
- ANSI color rendering depends on terminal support.

---

## Suggested CI usage

Fail on errors only:

```bash
./goulinette --root . --level 1
```

Fail on warnings too:

```bash
./goulinette --root . --level 1 --warnings-as-errors
```

Machine-readable output:

```bash
./goulinette --root . --format json
```

---

## Release assets (suggested)

Attach one or more of:

- `goulinette-linux-amd64`
- `goulinette-linux-arm64`
- `goulinette-darwin-amd64`
- `goulinette-darwin-arm64`
- `goulinette-windows-amd64.exe`

Checksums:

```text
SHA256 (goulinette-linux-amd64) = <fill>
SHA256 (goulinette-linux-arm64) = <fill>
SHA256 (goulinette-darwin-amd64) = <fill>
SHA256 (goulinette-darwin-arm64) = <fill>
SHA256 (goulinette-windows-amd64.exe) = <fill>
```

---

## Thanks

Thanks to everyone who tested rules, validated diagnostics, and tightened quality workflows during the build-up to `v1.0.0`.
