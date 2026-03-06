package rules

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestCollectSingleValueAssertions(t *testing.T) {
	src := `package p
func f(x any) {
  _ = x.(int)
  if v, ok := x.(string); ok {
    _ = v
  }
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	items := collectSingleValueAssertions(file)
	if len(items) != 1 {
		t.Fatalf("expected 1 single-value assertion violation candidate, got %d", len(items))
	}
}

func TestTypeSwitchCaseExemption(t *testing.T) {
	src := `package p
func f(x any) {
  switch x.(type) {
  case int:
    _ = x.(int)
  }
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	items := collectSingleValueAssertions(file)
	if len(items) != 1 {
		t.Fatalf("expected 1 assertion collected, got %d", len(items))
	}
	if !items[0].inTypeSwitchCase {
		t.Fatalf("assertion in type switch case should be marked exempt")
	}
}
