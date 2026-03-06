package rules

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	tstTestFilePerms = 0o644
	tstWriteErrFmt   = "write %s: %v"
	tstGoModFile     = "go.mod"
	tstSampleGoFile  = "sample.go"
	tstMinDiagCount  = 3
)

func writeTSTFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), tstTestFilePerms); err != nil {
		t.Fatalf(tstWriteErrFmt, name, err)
	}
	return path
}

// TestTST01TableDrivenHeuristics verifies that TST-01 detects test
// functions that manually call t.Run with repeated argument patterns
// where a table-driven approach would be more idiomatic.
func TestTST01TableDrivenHeuristics(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "manually unrolled subtests fail",
			source: `package sample
import "testing"
// TestParse documents this exported function.
func TestParse(t *testing.T) {
	t.Run("empty", func(t *testing.T) {})
	t.Run("valid", func(t *testing.T) {})
	t.Run("invalid", func(t *testing.T) {})
}
`,
			wantCount: 1,
		},
		{
			name: "range loop table-driven passes",
			source: `package sample
import "testing"
// TestParse documents this exported function.
func TestParse(t *testing.T) {
	tests := []struct{ name string }{{name: "a"}, {name: "b"}, {name: "c"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {})
	}
}
`,
			wantCount: 0,
		},
		{
			name: "three test variants sharing prefix fail",
			source: `package sample
import "testing"
// TestParseEmpty documents this exported function.
func TestParseEmpty(t *testing.T) {}
// TestParseValid documents this exported function.
func TestParseValid(t *testing.T) {}
// TestParseInvalid documents this exported function.
func TestParseInvalid(t *testing.T) {}
`,
			wantCount: 3,
		},
		{
			name: "benchmark and example excluded",
			source: `package sample
import "testing"
// BenchmarkParse documents this exported function.
func BenchmarkParse(b *testing.B) {}
// ExampleParse documents this exported function.
func ExampleParse() {}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			file := writeTSTFile(t, dir, "sample_test.go", tc.source)

			diags, err := NewTST01().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run TST-01: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d TST-01 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}

// TestTST02HelperFirstStatement verifies that TST-02 requires t.Helper()
// to be the first statement in any function that calls t.Helper(), and
// flags functions where it appears after other statements.
func TestTST02HelperFirstStatement(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "named helper missing helper call fails",
			source: `package sample
import "testing"
func requireEqual(t *testing.T) { t.Fatalf("bad") }
// TestX documents this exported function.
func TestX(t *testing.T) { requireEqual(t) }
`,
			wantCount: 1,
		},
		{
			name: "named helper with helper first passes",
			source: `package sample
import "testing"
func requireEqual(t *testing.T) { t.Helper(); t.Fatalf("bad") }
// TestX documents this exported function.
func TestX(t *testing.T) { requireEqual(t) }
`,
			wantCount: 0,
		},
		{
			name: "named helper with helper second fails",
			source: `package sample
import "testing"
func requireEqual(t *testing.T) { x := 1; t.Helper(); _ = x; t.Fatalf("bad") }
// TestX documents this exported function.
func TestX(t *testing.T) { requireEqual(t) }
`,
			wantCount: 1,
		},
		{
			name: "top-level test function is excluded",
			source: `package sample
import "testing"
// TestX documents this exported function.
func TestX(t *testing.T) { t.Fatalf("bad") }
`,
			wantCount: 0,
		},
		{
			name: "subtest anonymous function requires helper",
			source: `package sample
import "testing"
// TestX documents this exported function.
func TestX(t *testing.T) {
	t.Run("case", func(t *testing.T) { t.Fatalf("bad") })
}
`,
			wantCount: 1,
		},
		{
			name: "testing TB helper is recognized",
			source: `package sample
import "testing"
func requireTB(tb testing.TB) { tb.Fatalf("bad") }
// TestX documents this exported function.
func TestX(t *testing.T) { requireTB(t) }
`,
			wantCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			file := writeTSTFile(t, dir, "sample_test.go", tc.source)

			diags, err := NewTST02().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run TST-02: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d TST-02 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}

// TestTST03TimeSleepInTests verifies that TST-03 flags time.Sleep calls
// inside test files as errors, resolves aliased time imports, and does
// not fire on non-test production files.
func TestTST03TimeSleepInTests(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]string
		context      func(string, []string) Context
		wantCount    int
		wantSeverity string
	}{
		{
			name: "direct time sleep in test errors",
			files: map[string]string{
				tstGoModFile: "module example.com/tst03\n\ngo 1.22\n",
				"a_test.go": `package sample
import (
	"testing"
	"time"
)
// TestX documents this exported function.
func TestX(t *testing.T) { time.Sleep(time.Millisecond) }
`,
			},
			context:      func(dir string, _ []string) Context { return Context{Root: dir} },
			wantCount:    1,
			wantSeverity: "error",
		},
		{
			name: "aliased time import is resolved",
			files: map[string]string{
				tstGoModFile: "module example.com/tst03alias\n\ngo 1.22\n",
				"a_test.go": `package sample
import (
	"testing"
	tm "time"
)
// TestX documents this exported function.
func TestX(t *testing.T) { tm.Sleep(tm.Millisecond) }
`,
			},
			context:      func(dir string, _ []string) Context { return Context{Root: dir} },
			wantCount:    1,
			wantSeverity: "error",
		},
		{
			name: "non test file is ignored",
			files: map[string]string{
				tstSampleGoFile: `package sample
import "time"
func f() { time.Sleep(time.Second) }
`,
			},
			context:      func(_ string, files []string) Context { return Context{Files: files} },
			wantCount:    0,
			wantSeverity: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			files := make([]string, 0)
			for name, content := range tc.files {
				path := writeTSTFile(t, dir, name, content)
				if filepath.Ext(path) == ".go" {
					files = append(files, path)
				}
			}

			diags, err := NewTST03().Run(tc.context(dir, files))
			if err != nil {
				t.Fatalf("run TST-03: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d TST-03 diagnostics, got %d", tc.wantCount, len(diags))
			}
			if tc.wantCount > 0 && string(diags[0].Severity) != tc.wantSeverity {
				t.Fatalf("expected severity %s, got %s", tc.wantSeverity, diags[0].Severity)
			}
		})
	}
}
