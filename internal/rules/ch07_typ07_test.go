package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestContainsAnyType(t *testing.T) {
	src := `package p
type S struct {
  Value any
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	ts := file.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	st := ts.Type.(*ast.StructType)
	if !containsAnyType(st.Fields.List[0].Type) {
		t.Fatalf("expected containsAnyType to return true")
	}
}

func TestHasAnyJustification(t *testing.T) {
	src := `package p
// because dynamic JSON payload from external API
type S struct {
  Value any
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	gd := file.Decls[0].(*ast.GenDecl)
	if !hasAnyJustification(gd.Doc) {
		t.Fatalf("expected justification comment to be accepted")
	}
}
