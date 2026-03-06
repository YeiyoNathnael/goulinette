package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestHasNamedReturns(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "named return",
			src:  "package p\nfunc f() (value int, err error) { return }",
			want: true,
		},
		{
			name: "unnamed return",
			src:  "package p\nfunc f() (int, error) { return 0, nil }",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "x.go", tc.src, parser.ParseComments)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			fn := file.Decls[0].(*ast.FuncDecl)
			if got := hasNamedReturns(fn.Type); got != tc.want {
				t.Fatalf("hasNamedReturns() = %v, want %v", got, tc.want)
			}
		})
	}
}
