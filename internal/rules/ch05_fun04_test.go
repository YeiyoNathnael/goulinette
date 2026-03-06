package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestCountFuncReturns(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{name: "no return", src: "package p\nfunc f() {}", want: 0},
		{name: "single unnamed", src: "package p\nfunc f() int { return 0 }", want: 1},
		{name: "named grouped", src: "package p\nfunc f() (a, b int, err error) { return 0, 0, nil }", want: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "x.go", tc.src, parser.ParseComments)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			fn := file.Decls[0].(*ast.FuncDecl)
			if got := countFuncReturns(fn.Type); got != tc.want {
				t.Fatalf("countFuncReturns() = %d, want %d", got, tc.want)
			}
		})
	}
}
