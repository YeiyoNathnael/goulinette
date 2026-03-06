package rules

import (
	"go/parser"
	"go/token"
	"testing"
)

const (
	typTestFileName = "x.go"
	typParseErrFmt  = "parse failed: %v"
)

// TestCollectSingleValueAssertions verifies that single-value type assertions
// of the form v := x.(T) are correctly collected from an AST, while
// comma-ok assertions v, ok := x.(T) are not included.
func TestCollectSingleValueAssertions(t *testing.T) {
	src := `package p
func f(x any) {
  _ = x.(int)
  if v, ok := x.(string); ok {
    _ = v
  }
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, typTestFileName, src, parser.ParseComments)
	if err != nil {
		t.Fatalf(typParseErrFmt, err)
	}

	items := collectSingleValueAssertions(file)
	if len(items) != 1 {
		t.Fatalf("expected 1 single-value assertion violation candidate, got %d", len(items))
	}
}

// TestTypeSwitchCaseExemption verifies that TYP-05 does not report a
// single-value type assertion whose expression is a case clause inside
// a type-switch statement.
func TestTypeSwitchCaseExemption(t *testing.T) {
	src := `package p
func f(x any) {
  switch x.(type) {
  case int:
    _ = x.(int)
  }
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, typTestFileName, src, parser.ParseComments)
	if err != nil {
		t.Fatalf(typParseErrFmt, err)
	}

	items := collectSingleValueAssertions(file)
	if len(items) != 1 {
		t.Fatalf("expected 1 assertion collected, got %d", len(items))
	}
	if !items[0].inTypeSwitchCase {
		t.Fatalf("assertion in type switch case should be marked exempt")
	}
}
