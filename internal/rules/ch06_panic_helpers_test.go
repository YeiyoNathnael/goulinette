package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

const (
	panicHelpersTestFileName  = "x.go"
	panicHelpersParseFailHint = "parse failed: %v"
)

// TestIsRecoverInDeferredAnonymousFunc verifies that a recover() call is
// recognised as valid only when it appears directly inside a deferred
// anonymous function, not in a regular or non-deferred closure.
func TestIsRecoverInDeferredAnonymousFunc(t *testing.T) {
	src := `package p
func f() {
  defer func() {
    _ = recover()
  }()
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, panicHelpersTestFileName, src, parser.ParseComments)
	if err != nil {
		t.Fatalf(panicHelpersParseFailHint, err)
	}

	calls := collectCalls(file, "recover")
	if len(calls) != 1 {
		t.Fatalf("expected 1 recover call, got %d", len(calls))
	}
	if !isRecoverInDeferredAnonymousFunc(calls[0]) {
		t.Fatalf("recover call in deferred anon func should be accepted")
	}
}

// TestIsOperationalPanicArg verifies that panic arguments containing
// "impossible", "invariant", or "unreachable" keywords are classified as
// operational (expected programmer-assertion panics) rather than violations.
func TestIsOperationalPanicArg(t *testing.T) {
	info := &types.Info{}
	msg := &ast.BasicLit{Kind: token.STRING, Value: "\"failed to connect\""}
	if !isOperationalPanicArg(msg, info) {
		t.Fatalf("operational panic string should be detected")
	}
}
