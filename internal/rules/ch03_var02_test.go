package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

const (
	var02TestFileName  = "x.go"
	var02ExpectedFnMsg = "expected func decl"
	var02BoolType      = "bool"
)

// TestDefaultLiteralType verifies that defaultLiteralType maps untyped
// literal expressions to their default Go types (int, string, bool, etc.).
func TestDefaultLiteralType(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want string
		ok   bool
	}{
		{name: "int", expr: &ast.BasicLit{Kind: token.INT, Value: "20"}, want: "int", ok: true},
		{name: "float", expr: &ast.BasicLit{Kind: token.FLOAT, Value: "2.0"}, want: "float64", ok: true},
		{name: var02BoolType, expr: &ast.Ident{Name: "false"}, want: var02BoolType, ok: true},
		{name: "non literal", expr: &ast.Ident{Name: "x"}, want: "", ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			got, ok := defaultLiteralType(tc.expr)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("defaultLiteralType() = (%q,%v), want (%q,%v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}

// TestFindPostDeclConversionNeed verifies that VAR-02 detects the pattern
// of a typed var declaration immediately followed by an assignment that
// converts the initialiser to the declared type.
func TestFindPostDeclConversionNeed(t *testing.T) {
	src := `package p
func f() {
	b := 20
	_ = byte(b)
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, var02TestFileName, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	fn, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		t.Fatal(var02ExpectedFnMsg)
	}

	assign, ok := fn.Body.List[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected assign stmt")
	}

	ident, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		t.Fatalf("expected assign lhs ident")
	}
	ctype, found := findPostDeclConversionNeed(fn.Body, assign.End(), ident, "int")
	if !found || ctype != "byte" {
		t.Fatalf("findPostDeclConversionNeed() = (%q,%v), want (byte,true)", ctype, found)
	}
}
