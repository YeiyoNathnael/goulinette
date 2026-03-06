package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkParseFilesCacheMiss measures the cost of parsing a set of Go
// source files when the cache is cold (first call for that file set).
func BenchmarkParseFilesCacheMiss(b *testing.B) {
	dir := b.TempDir()
	paths := writeGoFiles(b, dir, 10)

	b.ResetTimer()
	for range b.N {
		clearParseFilesCache()
		if _, err := parseFiles(paths); err != nil {
			b.Fatalf("parseFiles: %v", err)
		}
	}
}

// BenchmarkParseFilesCacheHit measures the overhead of retrieving already-
// parsed files from the in-memory cache. This should be near-zero compared
// to the cold path.
func BenchmarkParseFilesCacheHit(b *testing.B) {
	dir := b.TempDir()
	paths := writeGoFiles(b, dir, 10)

	// Warm the cache once before the timed loop.
	clearParseFilesCache()
	if _, err := parseFiles(paths); err != nil {
		b.Fatalf("warm parseFiles: %v", err)
	}

	b.ResetTimer()
	for range b.N {
		if _, err := parseFiles(paths); err != nil {
			b.Fatalf("parseFiles: %v", err)
		}
	}
}

// BenchmarkCacheKeyGeneration measures the cost of computing the sorted
// cache key from a file path slice.
func BenchmarkCacheKeyGeneration(b *testing.B) {
	dir := b.TempDir()
	paths := writeGoFiles(b, dir, 50)

	b.ResetTimer()
	for range b.N {
		_ = parsedFilesCacheKey(paths)
	}
}

// writeGoFiles creates n minimal Go source files in dir and returns their
// paths. The files are valid Go packages so they can be parsed correctly.
func writeGoFiles(b *testing.B, dir string, n int) []string {
	b.Helper()
	paths := make([]string, n)
	for i := range paths {
		name := filepath.Join(dir, "file"+string(rune('a'+i%26))+".go")
		content := "package bench\n\nfunc f" + string(rune('A'+i%26)) + "() {}\n"
		if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
			b.Fatalf("write file: %v", err)
		}
		paths[i] = name
	}
	return paths
}
