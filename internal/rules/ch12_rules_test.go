package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeModuleFileCER(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestCER01_ReturnsErrorInterface(t *testing.T) {
	dir := t.TempDir()
	writeModuleFileCER(t, dir, "go.mod", "module example.com/cer01\n\ngo 1.22\n")
	writeModuleFileCER(t, dir, "sample.go", `package sample

type ValidationError struct{}

func (e *ValidationError) Error() string { return "bad" }

func Bad() *ValidationError { return nil }

type ParseErr struct{}

func (e ParseErr) Error() string { return "parse" }

func AlsoBad() ParseErr { return ParseErr{} }

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

func TestCER02_ConcreteErrorVarDeclarations(t *testing.T) {
	dir := t.TempDir()
	writeModuleFileCER(t, dir, "go.mod", "module example.com/cer02\n\ngo 1.22\n")
	writeModuleFileCER(t, dir, "a.go", `package sample

func f() {
	var err *LaterError
	_ = err
}
`)
	writeModuleFileCER(t, dir, "b.go", `package sample

type LaterError struct{}

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

func TestCER03_UnassignedConcreteErrorReturn(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "conditional assignment warns",
			source: `package sample
type ValidationError struct{}
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
type ValidationError struct{}
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
type ValidationError struct{}
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
			dir := t.TempDir()
			writeModuleFileCER(t, dir, "go.mod", "module example.com/cer03\n\ngo 1.22\n")
			writeModuleFileCER(t, dir, "sample.go", tc.source)

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
