package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeCTXFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestCTX01_ContextFirstParameter(t *testing.T) {
	dir := t.TempDir()
	writeCTXFile(t, dir, "go.mod", "module example.com/ctx01\n\ngo 1.22\n")
	writeCTXFile(t, dir, "sample.go", `package sample

import ctxpkg "context"

type S struct{}

func Bad(a int, ctx ctxpkg.Context) {}
func Good(ctx ctxpkg.Context, a int) {}
func (s S) Method(a int, ctx ctxpkg.Context) {}

type Runner interface {
	Run(a int, ctx ctxpkg.Context) error
	Okay(ctx ctxpkg.Context) error
}

var _ = func(a int, ctx ctxpkg.Context) {}
`)

	diags, err := NewCTX01().Run(Context{Root: dir})
	if err != nil {
		t.Fatalf("run CTX-01: %v", err)
	}
	if len(diags) != 4 {
		t.Fatalf("expected 4 CTX-01 diagnostics, got %d", len(diags))
	}
}

func TestCTX02_ContextStoredInStruct(t *testing.T) {
	dir := t.TempDir()
	writeCTXFile(t, dir, "go.mod", "module example.com/ctx02\n\ngo 1.22\n")
	writeCTXFile(t, dir, "sample.go", `package sample

import "context"

type A struct {
	context.Context
}

type B struct {
	Ctx *context.Context
}

type C struct {
	Value int
}
`)

	diags, err := NewCTX02().Run(Context{Root: dir})
	if err != nil {
		t.Fatalf("run CTX-02: %v", err)
	}
	if len(diags) != 2 {
		t.Fatalf("expected 2 CTX-02 diagnostics, got %d", len(diags))
	}
}

func TestCTX03_NilContextHeuristics(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "literal nil argument warns",
			source: `package sample
import "context"
func use(ctx context.Context) {}
func f() { use(nil) }
`,
			wantCount: 1,
		},
		{
			name: "unassigned local context var warns",
			source: `package sample
import "context"
func use(ctx context.Context) {}
func f() {
	var ctx context.Context
	use(ctx)
}
`,
			wantCount: 1,
		},
		{
			name: "assigned context var passes",
			source: `package sample
import "context"
func use(ctx context.Context) {}
func f() {
	var ctx context.Context
	ctx = context.Background()
	use(ctx)
}
`,
			wantCount: 0,
		},
		{
			name: "conditional assignment still warns",
			source: `package sample
import "context"
func use(ctx context.Context) {}
func f(b bool) {
	var ctx context.Context
	if b {
		ctx = context.Background()
	}
	use(ctx)
}
`,
			wantCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeCTXFile(t, dir, "go.mod", "module example.com/ctx03\n\ngo 1.22\n")
			writeCTXFile(t, dir, "sample.go", tc.source)

			diags, err := NewCTX03().Run(Context{Root: dir})
			if err != nil {
				t.Fatalf("run CTX-03: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d CTX-03 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}
