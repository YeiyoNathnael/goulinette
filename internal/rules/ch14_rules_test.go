package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

const (
	ctxFileGoMod  = "go.mod"
	ctxFileSample = "sample.go"
	ctxFilePerm   = 0o644
	ctxModule01   = "module example.com/ctx01\n\ngo 1.22\n"
	ctxModule02   = "module example.com/ctx02\n\ngo 1.22\n"
	ctxModule03   = "module example.com/ctx03\n\ngo 1.22\n"
	ctxModule04   = "module example.com/ctx04\n\ngo 1.22\n"
)

func writeCTXFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), ctxFilePerm); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

// TestCTX01ContextFirstParameter verifies that CTX-01 flags functions,
// methods, and interface methods where a context.Context parameter appears
// in any position other than first.
func TestCTX01ContextFirstParameter(t *testing.T) {
	dir := t.TempDir()
	_ = writeCTXFile(t, dir, ctxFileGoMod, ctxModule01)
	_ = writeCTXFile(t, dir, ctxFileSample, `package sample

import ctxpkg "context"

// S documents this exported type.
type S struct{}

// Bad documents this exported function.
func Bad(a int, ctx ctxpkg.Context) {}
// Good documents this exported function.
func Good(ctx ctxpkg.Context, a int) {}
// Method documents this exported method.
func (s S) Method(a int, ctx ctxpkg.Context) {}

// Runner documents this exported type.
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

// TestCTX02ContextStoredInStruct verifies that CTX-02 flags struct fields
// whose declared type is context.Context (by value or pointer), and does
// not fire on struct fields of unrelated types.
func TestCTX02ContextStoredInStruct(t *testing.T) {
	dir := t.TempDir()
	_ = writeCTXFile(t, dir, ctxFileGoMod, ctxModule02)
	_ = writeCTXFile(t, dir, ctxFileSample, `package sample

import "context"

// A documents this exported type.
type A struct {
	context.Context
}

// B documents this exported type.
type B struct {
	Ctx *context.Context
}

// C documents this exported type.
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

// TestCTX03NilContextHeuristics verifies that CTX-03 detects call sites
// that pass an uninitialised context variable or an explicit nil where a
// context.Context argument is expected.
func TestCTX03NilContextHeuristics(t *testing.T) {
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
			t.Helper()
			dir := t.TempDir()
			_ = writeCTXFile(t, dir, ctxFileGoMod, ctxModule03)
			_ = writeCTXFile(t, dir, ctxFileSample, tc.source)

			diags, err := NewCTX03().Run(Context{Root: dir})
			if err != nil {
				t.Fatalf("run CTX-03: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d CTX-03 diagnostics, got %d", tc.wantCount, len(diags))
			}
			if tc.wantCount > 0 && diags[0].Severity != diag.SeverityError {
				t.Fatalf("expected CTX-03 severity error, got %s", diags[0].Severity)
			}
		})
	}
}

// TestCTX04CancelMustBeHandled documents this exported function.
func TestCTX04CancelMustBeHandled(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "missing cancel handling fails",
			source: `package sample
import "context"
func f(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	_ = ctx
	_ = cancel
}
`,
			wantCount: 1,
		},
		{
			name: "defer cancel immediately passes",
			source: `package sample
import "context"
func f(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	_ = ctx
}
`,
			wantCount: 0,
		},
		{
			name: "defer with wrapper argument accepted",
			source: `package sample
import "context"
func cleanup(fn func()) { fn() }
func f(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, 0)
	defer cleanup(cancel)
	_ = ctx
}
`,
			wantCount: 0,
		},
		{
			name: "conditional cancel call warns",
			source: `package sample
import "context"
func f(parent context.Context, b bool) {
	ctx, cancel := context.WithCancel(parent)
	if b {
		cancel()
	}
	_ = ctx
}
`,
			wantCount: 1,
		},
		{
			name: "early return before defer warns",
			source: `package sample
import "context"
func f(parent context.Context, b bool) {
	ctx, cancel := context.WithCancel(parent)
	if b {
		return
	}
	defer cancel()
	_ = ctx
}
`,
			wantCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			_ = writeCTXFile(t, dir, ctxFileGoMod, ctxModule04)
			_ = writeCTXFile(t, dir, ctxFileSample, tc.source)

			diags, err := NewCTX04().Run(Context{Root: dir})
			if err != nil {
				t.Fatalf("run CTX-04: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d CTX-04 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}
