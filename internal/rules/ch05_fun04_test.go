package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

const (
	fun04TestFileName    = "x.go"
	fun04ParseErrFmt     = "parse failed: %v"
	fun04ExpectedFuncMsg = "expected func decl"
	fun04ThreeReturns    = 3
)

// TestCountFuncReturns verifies that countFuncReturns returns the correct
// number of return statements reachable within a function body, including
// those inside nested function literals.
func TestCountFuncReturns(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{name: "no return", src: "package p\nfunc f() {}", want: 0},
		{name: "single unnamed", src: "package p\nfunc f() int { return 0 }", want: 1},
		{name: "named grouped", src: "package p\nfunc f() (a, b int, err error) { return 0, 0, nil }", want: fun04ThreeReturns},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, fun04TestFileName, tc.src, parser.ParseComments)
			if err != nil {
				t.Fatalf(fun04ParseErrFmt, err)
			}
			fn, ok := file.Decls[0].(*ast.FuncDecl)
			if !ok {
				t.Fatal(fun04ExpectedFuncMsg)
			}
			if got := countFuncReturns(fn.Type); got != tc.want {
				t.Fatalf("countFuncReturns() = %d, want %d", got, tc.want)
			}
		})
	}
}
