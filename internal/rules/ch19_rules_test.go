package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTSTFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestTST01_TableDrivenHeuristics(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "manually unrolled subtests fail",
			source: `package sample
import "testing"
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
func TestParseEmpty(t *testing.T) {}
func TestParseValid(t *testing.T) {}
func TestParseInvalid(t *testing.T) {}
`,
			wantCount: 3,
		},
		{
			name: "benchmark and example excluded",
			source: `package sample
import "testing"
func BenchmarkParse(b *testing.B) {}
func ExampleParse() {}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

func TestTST02_HelperFirstStatement(t *testing.T) {
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
func TestX(t *testing.T) { requireEqual(t) }
`,
			wantCount: 1,
		},
		{
			name: "named helper with helper first passes",
			source: `package sample
import "testing"
func requireEqual(t *testing.T) { t.Helper(); t.Fatalf("bad") }
func TestX(t *testing.T) { requireEqual(t) }
`,
			wantCount: 0,
		},
		{
			name: "named helper with helper second fails",
			source: `package sample
import "testing"
func requireEqual(t *testing.T) { x := 1; t.Helper(); _ = x; t.Fatalf("bad") }
func TestX(t *testing.T) { requireEqual(t) }
`,
			wantCount: 1,
		},
		{
			name: "top-level test function is excluded",
			source: `package sample
import "testing"
func TestX(t *testing.T) { t.Fatalf("bad") }
`,
			wantCount: 0,
		},
		{
			name: "subtest anonymous function requires helper",
			source: `package sample
import "testing"
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
func TestX(t *testing.T) { requireTB(t) }
`,
			wantCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

func TestTST03_TimeSleepInTests(t *testing.T) {
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
				"go.mod": "module example.com/tst03\n\ngo 1.22\n",
				"a_test.go": `package sample
import (
	"testing"
	"time"
)
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
				"go.mod": "module example.com/tst03alias\n\ngo 1.22\n",
				"a_test.go": `package sample
import (
	"testing"
	tm "time"
)
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
				"sample.go": `package sample
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
