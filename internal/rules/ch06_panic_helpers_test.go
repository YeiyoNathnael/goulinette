package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestIsRecoverInDeferredAnonymousFunc(t *testing.T) {
	src := `package p
func f() {
  defer func() {
    _ = recover()
  }()
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	calls := collectCalls(file, "recover")
	if len(calls) != 1 {
		t.Fatalf("expected 1 recover call, got %d", len(calls))
	}
	if !isRecoverInDeferredAnonymousFunc(calls[0]) {
		t.Fatalf("recover call in deferred anon func should be accepted")
	}
}

func TestIsOperationalPanicArg(t *testing.T) {
	info := &types.Info{}
	msg := &ast.BasicLit{Kind: token.STRING, Value: "\"failed to connect\""}
	if !isOperationalPanicArg(msg, info) {
		t.Fatalf("operational panic string should be detected")
	}
}
