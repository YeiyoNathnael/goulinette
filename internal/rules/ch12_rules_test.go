package rules

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	cerTestFilePerms = 0o644
	cerWriteErrFmt   = "write %s: %v"
	cerGoModFile     = "go.mod"
	cerSampleGoFile  = "sample.go"
	cerAGoFile       = "a.go"
)

func writeModuleFileCER(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), cerTestFilePerms); err != nil {
		t.Fatalf(cerWriteErrFmt, name, err)
	}
	return path
}

// TestCER01ReturnsErrorInterface verifies that CER-01 flags functions that
// return a concrete error type (e.g. *MyError) instead of the error
// interface, and does not fire on functions returning error directly.
func TestCER01ReturnsErrorInterface(t *testing.T) {
	dir := t.TempDir()
	_ = writeModuleFileCER(t, dir, cerGoModFile, "module example.com/cer01\n\ngo 1.22\n")
	_ = writeModuleFileCER(t, dir, cerSampleGoFile, `package sample

// ValidationError documents this exported type.
type ValidationError struct{}

// Error documents this exported method.
func (e *ValidationError) Error() string { return "bad" }

// Bad documents this exported function.
func Bad() *ValidationError { return nil }

// ParseErr documents this exported type.
type ParseErr struct{}

// Error documents this exported method.
func (e ParseErr) Error() string { return "parse" }

// AlsoBad documents this exported function.
func AlsoBad() ParseErr { return ParseErr{} }

// Good documents this exported function.
func Good() error { return nil }
`)

	diags, err := NewCER01().Run(Context{Root: dir})
	if err != nil {
		t.Fatalf("run CER-01: %v", err)
	}
	if len(diags) != 2 {
		t.Fatalf("expected 2 CER-01 diagnostics, got %d", len(diags))
	}
}

// TestCER02ConcreteErrorVarDeclarations verifies that CER-02 flags
// package-level variable declarations whose type is a concrete error struct
// rather than the error interface.
func TestCER02ConcreteErrorVarDeclarations(t *testing.T) {
	dir := t.TempDir()
	_ = writeModuleFileCER(t, dir, cerGoModFile, "module example.com/cer02\n\ngo 1.22\n")
	_ = writeModuleFileCER(t, dir, cerAGoFile, `package sample

func f() {
	var err *LaterError
	_ = err
}
`)
	_ = writeModuleFileCER(t, dir, "b.go", `package sample

// LaterError documents this exported type.
type LaterError struct{}

// Error documents this exported method.
func (e *LaterError) Error() string { return "later" }
`)

	diags, err := NewCER02().Run(Context{Root: dir})
	if err != nil {
		t.Fatalf("run CER-02: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected 1 CER-02 diagnostic, got %d", len(diags))
	}
}

// TestCER03UnassignedConcreteErrorReturn verifies that CER-03 detects
// concrete error variables that are declared but never populated before
// being returned, leaving the caller with an unexpected non-nil typed nil.
func TestCER03UnassignedConcreteErrorReturn(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "conditional assignment warns",
			source: `package sample
// ValidationError documents this exported type.
type ValidationError struct{}
// Error documents this exported method.
func (e *ValidationError) Error() string { return "bad" }
func f(bad bool) error {
	var err *ValidationError
	if bad {
		err = &ValidationError{}
	}
	return err
}
`,
			wantCount: 1,
		},
		{
			name: "all-path assignment passes",
			source: `package sample
// ValidationError documents this exported type.
type ValidationError struct{}
// Error documents this exported method.
func (e *ValidationError) Error() string { return "bad" }
func f(bad bool) error {
	var err *ValidationError
	if bad {
		err = &ValidationError{}
	} else {
		err = &ValidationError{}
	}
	return err
}
`,
			wantCount: 0,
		},
		{
			name: "immediate assignment passes",
			source: `package sample
// ValidationError documents this exported type.
type ValidationError struct{}
// Error documents this exported method.
func (e *ValidationError) Error() string { return "bad" }
func f() error {
	var err *ValidationError
	err = &ValidationError{}
	return err
}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			_ = writeModuleFileCER(t, dir, cerGoModFile, "module example.com/cer03\n\ngo 1.22\n")
			_ = writeModuleFileCER(t, dir, cerSampleGoFile, tc.source)

			diags, err := NewCER03().Run(Context{Root: dir})
			if err != nil {
				t.Fatalf("run CER-03: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d CER-03 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}
