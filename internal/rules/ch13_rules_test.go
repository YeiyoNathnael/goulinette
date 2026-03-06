package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	structuralSampleGo = "sample.go"
	structuralTestGo   = "sample_test.go"
	structuralFilePerm = 0o644
	lineLimitMax       = 50
	overLineLimit      = 51
	overFileLines      = 501
	testSkipLines      = 600
	nearFileLimit      = 499
	packageSampleLine  = "package sample\n"
	varOneLine         = "var _ = 1\n"
	writeFileErrFmt    = "write %s: %v"
	expectedDiagsFmt   = "expected %d diagnostics, got %d"
)

func writeStructuralFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), structuralFilePerm); err != nil {
		t.Fatalf(writeFileErrFmt, name, err)
	}
	return path
}

func bodyWithNLines(n int) string {
	if n <= 0 {
		return ""
	}
	var builder strings.Builder
	for i := 0; i < n; i++ {
		builder.WriteString("\t_ = 0\n")
	}
	return builder.String()
}

// TestLIM01FunctionLineLimits verifies that LIM-01 reports any function
// (including func literals) whose body exceeds 50 lines, and does not
// fire for functions at or below the limit.
func TestLIM01FunctionLineLimits(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name:      "long func decl fails",
			source:    "package sample\nfunc f() {\n" + bodyWithNLines(overLineLimit) + "}\n",
			wantCount: 1,
		},
		{
			name:      "long func literal fails independently",
			source:    "package sample\nfunc f() {\n\tg := func() {\n" + bodyWithNLines(overLineLimit) + "\t}\n\t_ = g\n}\n",
			wantCount: 1,
		},
		{
			name:      "exactly 50 lines passes",
			source:    "package sample\nfunc f() {\n" + bodyWithNLines(lineLimitMax) + "}\n",
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			file := writeStructuralFile(t, dir, structuralSampleGo, tc.source)

			diags, err := NewLIM01().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run LIM-01: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf(expectedDiagsFmt, tc.wantCount, len(diags))
			}
		})
	}
}

// TestLIM02ParameterCounting verifies that LIM-02 counts grouped parameters
// by name (not type), excludes the method receiver, and treats variadic
// parameters as a single argument.
func TestLIM02ParameterCounting(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "grouped parameters counted by names",
			source: `package sample
func f(a, b, c, d, e, f int) {}
`,
			wantCount: 1,
		},
		{
			name: "method receiver excluded",
			source: `package sample
// S documents this exported type.
type S struct{}
func (s S) f(a, b, c, d, e int) {}
`,
			wantCount: 0,
		},
		{
			name: "variadic counts as one",
			source: `package sample
func f(a, b, c, d int, rest ...string) {}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			file := writeStructuralFile(t, dir, structuralSampleGo, tc.source)

			diags, err := NewLIM02().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run LIM-02: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf(expectedDiagsFmt, tc.wantCount, len(diags))
			}
		})
	}
}

// TestLIM03NestingDepth verifies that LIM-03 reports function bodies whose
// control-flow nesting exceeds the allowed depth, and passes functions that
// stay within the limit.
func TestLIM03NestingDepth(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantAtMin int
	}{
		{
			name: "depth over four fails",
			source: `package sample
func f(a, b, c, d, e bool) {
	if a {
		for b {
			switch {
			case c:
				if d {
					if e {
						_ = 1
					}
				}
			}
		}
	}
}
`,
			wantAtMin: 1,
		},
		{
			name: "func literal resets depth",
			source: `package sample
func f(a, b, c, d bool) {
	if a {
		for b {
			g := func() {
				if c {
					if d {
						_ = 1
					}
				}
			}
			_ = g
		}
	}
}
`,
			wantAtMin: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			file := writeStructuralFile(t, dir, structuralSampleGo, tc.source)

			diags, err := NewLIM03().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run LIM-03: %v", err)
			}
			if len(diags) < tc.wantAtMin {
				t.Fatalf("expected at least %d diagnostics, got %d", tc.wantAtMin, len(diags))
			}
			if tc.wantAtMin == 0 && len(diags) != 0 {
				t.Fatalf("expected no diagnostics, got %d", len(diags))
			}
		})
	}
}

// TestLIM04FileLength verifies that LIM-04 flags source files exceeding
// 500 lines and exempts test files from the file-length limit.
func TestLIM04FileLength(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		source    string
		wantCount int
	}{
		{
			name:      "over 500 warns",
			filename:  structuralSampleGo,
			source:    packageSampleLine + strings.Repeat(varOneLine, overFileLines),
			wantCount: 1,
		},
		{
			name:      "generated file skipped",
			filename:  "generated.pb.go",
			source:    packageSampleLine + "// Code generated by tool. DO NOT EDIT.\n" + strings.Repeat(varOneLine, testSkipLines),
			wantCount: 0,
		},
		{
			name:      "test file skipped",
			filename:  structuralTestGo,
			source:    packageSampleLine + strings.Repeat(varOneLine, testSkipLines),
			wantCount: 0,
		},
		{
			name:      "trailing blanks ignored",
			filename:  "trim.go",
			source:    packageSampleLine + strings.Repeat(varOneLine, nearFileLimit) + "\n\n\n",
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			file := writeStructuralFile(t, dir, tc.filename, tc.source)

			diags, err := NewLIM04().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run LIM-04: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf(expectedDiagsFmt, tc.wantCount, len(diags))
			}
		})
	}
}
