package rules

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	namTestFilePerms = 0o644
	namWriteErrFmt   = "write %s: %v"
	namSampleGoFile  = "sample.go"
)

func writeNAMFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), namTestFilePerms); err != nil {
		t.Fatalf(namWriteErrFmt, name, err)
	}
	return path
}

// TestNAM03ScopeProportionality verifies that NAM-03 flags single-letter
// variables used across a long scope and overly long names used in a
// short scope.
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
			t.Helper()
			dir := t.TempDir()
			file := writeNAMFile(t, dir, namSampleGoFile, tc.source)

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

// TestNAM04PackageLevelDescriptiveNames verifies that NAM-04 flags
// short or cryptic package-level variable and constant names.
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
			t.Helper()
			dir := t.TempDir()
			file := writeNAMFile(t, dir, namSampleGoFile, tc.source)

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

// TestNAM05InterfaceErSuffix verifies that NAM-05 flags exported interface
// types whose names do not end in the conventional "-er" suffix.
func TestNAM05InterfaceErSuffix(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "interface without er suffix warns",
			source: `package sample
// Service documents this exported type.
type Service interface { Run() }
`,
			wantCount: 1,
		},
		{
			name: "interface with er suffix passes",
			source: `package sample
// Runner documents this exported type.
type Runner interface { Run() }
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			file := writeNAMFile(t, dir, namSampleGoFile, tc.source)

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
