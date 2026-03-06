package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeNAMFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestNAM03ScopeProportionality(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "single letter long scope warns",
			source: `package sample
func f() {
	x := 0
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
	_ = x
}
`,
			wantCount: 1,
		},
		{
			name: "long name short scope warns",
			source: `package sample
func f() {
	veryLongTemporaryName := 1
	_ = veryLongTemporaryName
}
`,
			wantCount: 1,
		},
		{
			name: "well proportioned names pass",
			source: `package sample
func f() {
	value := 1
	_ = value
}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			file := writeNAMFile(t, dir, "sample.go", tc.source)

			diags, err := NewNAM03().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run NAM-03: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d NAM-03 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}

func TestNAM04PackageLevelDescriptiveNames(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "short package-level names warn",
			source: `package sample
var dt = 10
const id = "x"
`,
			wantCount: 2,
		},
		{
			name: "descriptive package-level names pass",
			source: `package sample
var defaultTimeout = 10
const userIdentifier = "x"
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			file := writeNAMFile(t, dir, "sample.go", tc.source)

			diags, err := NewNAM04().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run NAM-04: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d NAM-04 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}

func TestNAM05InterfaceErSuffix(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "interface without er suffix warns",
			source: `package sample
type Service interface { Run() }
`,
			wantCount: 1,
		},
		{
			name: "interface with er suffix passes",
			source: `package sample
type Runner interface { Run() }
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			file := writeNAMFile(t, dir, "sample.go", tc.source)

			diags, err := NewNAM05().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run NAM-05: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d NAM-05 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}
