package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestDefaultLiteralType(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want string
		ok   bool
	}{
		{name: "int", expr: &ast.BasicLit{Kind: token.INT, Value: "20"}, want: "int", ok: true},
		{name: "float", expr: &ast.BasicLit{Kind: token.FLOAT, Value: "2.0"}, want: "float64", ok: true},
		{name: "bool", expr: &ast.Ident{Name: "false"}, want: "bool", ok: true},
		{name: "non literal", expr: &ast.Ident{Name: "x"}, want: "", ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := defaultLiteralType(tc.expr)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("defaultLiteralType() = (%q,%v), want (%q,%v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestFindPostDeclConversionNeed(t *testing.T) {
	src := `package p
func f() {
	b := 20
	_ = byte(b)
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	fn, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected func decl")
	}

	assign, ok := fn.Body.List[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected assign stmt")
	}

	ident := assign.Lhs[0].(*ast.Ident)
	ctype, found := findPostDeclConversionNeed(fn.Body, assign.End(), ident, "int")
	if !found || ctype != "byte" {
		t.Fatalf("findPostDeclConversionNeed() = (%q,%v), want (byte,true)", ctype, found)
	}
}
