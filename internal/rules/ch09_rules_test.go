package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeGoFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestDOC01ToDOC04(t *testing.T) {
	tests := []struct {
		name        string
		ruleFactory func() Rule
		source      string
		wantCount   int
	}{
		{
			name:        "DOC-01 blank line between comment and declaration",
			ruleFactory: NewDOC01,
			source: `package sample
// Exported does work.

func Exported() {}
`,
			wantCount: 1,
		},
		{
			name:        "DOC-02 block comment docs are forbidden",
			ruleFactory: NewDOC02,
			source: `package sample
/* Exported does work. */
func Exported() {}
`,
			wantCount: 1,
		},
		{
			name:        "DOC-03 comment must start with symbol name",
			ruleFactory: NewDOC03,
			source: `package sample
// This function does work.
func Exported() {}
`,
			wantCount: 1,
		},
		{
			name:        "DOC-04 exported symbol requires comment",
			ruleFactory: NewDOC04,
			source: `package sample
func Exported() {}
`,
			wantCount: 1,
		},
		{
			name:        "DOC-04 accepts valid comments on exported symbols",
			ruleFactory: NewDOC04,
			source: `package sample
// Exported does work.
func Exported() {}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			file := writeGoFile(t, dir, "sample.go", tc.source)

			diags, err := tc.ruleFactory().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run %s: %v", tc.ruleFactory().ID(), err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}

func TestDOC05InitRestrictions(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "immutable package var setup allowed",
			source: `package sample
var timeoutSeconds int
func init() {
	timeoutSeconds = 30
}
`,
			wantCount: 0,
		},
		{
			name: "call in init triggers warning",
			source: `package sample
import "os"
var envValue string
func init() {
	envValue = os.Getenv("APP_ENV")
}
`,
			wantCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			file := writeGoFile(t, dir, "sample.go", tc.source)

			diags, err := NewDOC05().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run DOC-05: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}
