package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

const (
	fun02TestFileName  = "x.go"
	fun02ParseFailHint = "parse failed: %v"
)

// TestErrorMustBeLast documents this exported function.
func TestErrorMustBeLast(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "error last",
			src:  "package p\nfunc f() (int, error) { return 0, nil }",
			want: true,
		},
		{
			name: "error first",
			src:  "package p\nfunc f() (error, int) { return nil, 0 }",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, fun02TestFileName, tc.src, parser.ParseComments)
			if err != nil {
				t.Fatalf(fun02ParseFailHint, err)
			}
			fn, ok := file.Decls[0].(*ast.FuncDecl)
			if !ok {
				t.Fatalf("expected func decl")
			}
			if got := errorMustBeLast(fn.Type.Results); got != tc.want {
				t.Fatalf("errorMustBeLast() = %v, want %v", got, tc.want)
			}
		})
	}
}
