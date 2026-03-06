package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

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
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "x.go", tc.src, parser.ParseComments)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			fn := file.Decls[0].(*ast.FuncDecl)
			if got := errorMustBeLast(fn.Type.Results); got != tc.want {
				t.Fatalf("errorMustBeLast() = %v, want %v", got, tc.want)
			}
		})
	}
}
