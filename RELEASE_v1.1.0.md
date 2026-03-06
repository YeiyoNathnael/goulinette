# Goulinette v1.1.0 — Cobalt Parallel

A major feature release adding parallel rule execution, inline suppression,
`--version`, `--explain`, a GitHub Actions CI pipeline, and cache benchmarks.

---

## What's new since v1.0.0

### Parallel rule execution (`--max-workers`)

Rules now execute concurrently through a bounded goroutine worker pool.

```bash
./goulinette --max-workers 8   # use 8 workers (default 4)
```

- Workers pull from a buffered work channel and write to a pre-sized results channel.
- The pool respects `context.Context` cancellation.
- Combined with the existing AST/type-load caches, this significantly reduces
  wall-clock time on large repositories.

---

### Inline suppression (`//goulinette:ignore`)

Any diagnostic can be silenced without touching analysis configuration:

```go
// Suppress one rule on the same line
result := riskyOp() //goulinette:ignore ERR-01

// Suppress multiple rules on the next line
//goulinette:ignore MAG-02 NAM-07
const appName = "myapp"

// Suppress everything on the next line
//goulinette:ignore
legacyHack()
```

- Placing the directive on the **same line** or on the **immediately preceding line** both work.
- Rule IDs are matched case-insensitively.
- A bare `//goulinette:ignore` (no IDs) silences all rules for that line.

---

### `--version` flag

```bash
./goulinette --version
# goulinette 1.1.0
```

Release binaries have the version string injected at build time via `-ldflags`.

---

### `--explain` flag

Print the rationale behind any rule without running analysis:

```bash
./goulinette --explain CTX-01
./goulinette --explain all     # print rationale for all 62 rules
```

---

### GitHub Actions CI

`.github/workflows/ci.yml` runs on every push to `main` / `fix/**` / `feat/**`
and on all pull requests:

| Job         | Command                                              |
|-------------|------------------------------------------------------|
| Build       | `go build ./...`                                     |
| Vet         | `go vet ./...`                                       |
| Test (race) | `go test -race -count=1 ./...`                       |
| Self-lint   | build binary → `--level=3 --format json` → 0 findings |

---

### Cache benchmarks

```bash
go test ./internal/rules -bench BenchmarkParseFiles -benchmem
go test ./internal/rules -bench BenchmarkCacheKey   -benchmem
```

| Benchmark                      | ns/op    |
|-------------------------------|----------|
| `BenchmarkParseFilesCacheMiss` | ~420 000 |
| `BenchmarkParseFilesCacheHit`  | ~5 000   |
| `BenchmarkCacheKeyGeneration`  | < 1 000  |

---

## Install

### Pre-built binaries (this release)

Download the archive for your platform from the release assets, extract, and
place `goulinette` (or `goulinette.exe` on Windows) somewhere on your `PATH`.

| OS      | Arch    | Asset                               |
|---------|---------|-------------------------------------|
| Linux   | amd64   | `goulinette_linux_amd64.tar.gz`     |
| Linux   | arm64   | `goulinette_linux_arm64.tar.gz`     |
| macOS   | amd64   | `goulinette_darwin_amd64.tar.gz`    |
| macOS   | arm64   | `goulinette_darwin_arm64.tar.gz`    |
| Windows | amd64   | `goulinette_windows_amd64.zip`      |
| Windows | arm64   | `goulinette_windows_arm64.zip`      |

### Build from source

```bash
git clone https://github.com/YeiyoNathnael/goulinette.git
cd goulinette
go build -ldflags "-X github.com/YeiyoNathnael/goulinette/internal/version.Current=1.1.0" \
  -o goulinette ./cmd/goulinette
```

---

## Upgrade from v1.0.0

No configuration file changes required. The new flags are additive and all
existing flags keep their meaning. Simply swap the binary.

---

## Commits since v1.0.0

- `e096aa5` feat: implement --max-workers parallel rule execution
- `8a23069` feat: CI, inline suppression, --version, --explain, cache benchmarks
- `ab7d99d` docs: update README for all new features
