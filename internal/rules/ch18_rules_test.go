package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeMAGFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestMAG01_NumberLiteralRule(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		context   func(string, []string) Context
		wantCount int
	}{
		{
			name: "repeated numeric literal across files fails",
			files: map[string]string{
				"go.mod":    "module example.com/mag01\n\ngo 1.22\n",
				"a.go":      "package sample\nfunc a() int { return 30 }\n",
				"b.go":      "package sample\nfunc b() int { return 30 }\n",
				"helper.go": "package sample\nfunc c() int { return 1 }\n",
			},
			context:   func(dir string, _ []string) Context { return Context{Root: dir} },
			wantCount: 2,
		},
		{
			name: "exempt 0 1 2 and -1",
			files: map[string]string{
				"sample.go": `package sample
func f() int {
	v := 0
	v += 1
	v += 2
	v += -1
	return v
}
`,
			},
			context:   func(_ string, files []string) Context { return Context{Files: files} },
			wantCount: 0,
		},
		{
			name: "const declaration values are skipped",
			files: map[string]string{
				"sample.go": `package sample
const timeout = 30
func f() int { return 30 }
`,
			},
			context:   func(_ string, files []string) Context { return Context{Files: files} },
			wantCount: 0,
		},
		{
			name: "test files require threshold three",
			files: map[string]string{
				"sample_test.go": `package sample
func a() int { return 30 }
func b() int { return 30 }
`,
			},
			context:   func(_ string, files []string) Context { return Context{Files: files} },
			wantCount: 0,
		},
		{
			name: "test files with three hits are flagged",
			files: map[string]string{
				"sample_test.go": `package sample
func a() int { return 30 }
func b() int { return 30 }
func c() int { return 30 }
`,
			},
			context:   func(_ string, files []string) Context { return Context{Files: files} },
			wantCount: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := make([]string, 0)
			for name, content := range tc.files {
				path := writeMAGFile(t, dir, name, content)
				if filepath.Ext(path) == ".go" {
					files = append(files, path)
				}
			}

			diags, err := NewMAG01().Run(tc.context(dir, files))
			if err != nil {
				t.Fatalf("run MAG-01: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d MAG-01 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}

func TestMAG02_StringLiteralRule(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		context   func(string, []string) Context
		wantCount int
	}{
		{
			name: "repeated key-like string across files fails",
			files: map[string]string{
				"go.mod": "module example.com/mag02\n\ngo 1.22\n",
				"a.go": `package sample
func a() map[string]int { return map[string]int{"user_id": 1} }
`,
				"b.go": `package sample
func b(m map[string]int) int { return m["user_id"] }
`,
			},
			context:   func(dir string, _ []string) Context { return Context{Root: dir} },
			wantCount: 2,
		},
		{
			name: "import path strings are skipped",
			files: map[string]string{
				"sample.go": `package sample
import "strings"
func f() string { return strings.ToUpper("abc") }
`,
			},
			context:   func(_ string, files []string) Context { return Context{Files: files} },
			wantCount: 0,
		},
		{
			name: "struct tags are skipped",
			files: map[string]string{
				"sample.go": "package sample\ntype User struct { A string `json:\"id\"`; B string `json:\"id\"` }\n",
			},
			context:   func(_ string, files []string) Context { return Context{Files: files} },
			wantCount: 0,
		},
		{
			name: "errors.New and fmt.Errorf literals are skipped",
			files: map[string]string{
				"sample.go": `package sample
import (
	"errors"
	"fmt"
)
func f() error {
	if true {
		return errors.New("failed to read")
	}
	return fmt.Errorf("failed to read")
}
`,
			},
			context:   func(_ string, files []string) Context { return Context{Files: files} },
			wantCount: 0,
		},
		{
			name: "short literals are skipped",
			files: map[string]string{
				"sample.go": `package sample
func f() string {
	a := ","
	b := ","
	return a + b
}
`,
			},
			context:   func(_ string, files []string) Context { return Context{Files: files} },
			wantCount: 0,
		},
		{
			name: "test files require threshold three",
			files: map[string]string{
				"sample_test.go": `package sample
func a() string { return "header_name" }
func b() string { return "header_name" }
`,
			},
			context:   func(_ string, files []string) Context { return Context{Files: files} },
			wantCount: 0,
		},
		{
			name: "test files with three hits are flagged",
			files: map[string]string{
				"sample_test.go": `package sample
func a() string { return "header_name" }
func b() string { return "header_name" }
func c() string { return "header_name" }
`,
			},
			context:   func(_ string, files []string) Context { return Context{Files: files} },
			wantCount: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := make([]string, 0)
			for name, content := range tc.files {
				path := writeMAGFile(t, dir, name, content)
				if filepath.Ext(path) == ".go" {
					files = append(files, path)
				}
			}

			diags, err := NewMAG02().Run(tc.context(dir, files))
			if err != nil {
				t.Fatalf("run MAG-02: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d MAG-02 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}
