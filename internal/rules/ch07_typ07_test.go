package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

const (
	typ07TestFileName = "x.go"
	typ07ParseErrFmt  = "parse failed: %v"
)

// TestContainsAnyType documents this exported function.
func TestContainsAnyType(t *testing.T) {
	src := `package p
// S documents this exported type.
type S struct {
  Value any
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, typ07TestFileName, src, parser.ParseComments)
	if err != nil {
		t.Fatalf(typ07ParseErrFmt, err)
	}

	gd, ok := file.Decls[0].(*ast.GenDecl)
	if !ok {
		t.Fatalf("expected gen decl")
	}
	ts, ok := gd.Specs[0].(*ast.TypeSpec)
	if !ok {
		t.Fatalf("expected type spec")
	}
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		t.Fatalf("expected struct type")
	}
	if !containsAnyType(st.Fields.List[0].Type) {
		t.Fatalf("expected containsAnyType to return true")
	}
}

// TestHasAnyJustification documents this exported function.
func TestHasAnyJustification(t *testing.T) {
	src := `package p
// because dynamic JSON payload from external API
type S struct {
  Value any
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, typ07TestFileName, src, parser.ParseComments)
	if err != nil {
		t.Fatalf(typ07ParseErrFmt, err)
	}

	gd, ok := file.Decls[0].(*ast.GenDecl)
	if !ok {
		t.Fatalf("expected gen decl")
	}
	if !hasAnyJustification(gd.Doc) {
		t.Fatalf("expected justification comment to be accepted")
	}
}
