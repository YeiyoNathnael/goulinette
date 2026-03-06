package rules

import (
	"go/parser"
	"go/token"
	"testing"
)

const (
	errMessageTestFileName  = "x.go"
	errMessageParseFailHint = "parse failed: %v"
)

// TestHasForbiddenErrorSuffix documents this exported function.
func TestHasForbiddenErrorSuffix(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{name: "period", msg: "failed.", want: true},
		{name: "exclamation", msg: "failed!", want: true},
		{name: "newline", msg: "failed\n", want: true},
		{name: "clean", msg: "failed to open file", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			if got := hasForbiddenErrorSuffix(tc.msg); got != tc.want {
				t.Fatalf("hasForbiddenErrorSuffix(%q) = %v, want %v", tc.msg, got, tc.want)
			}
		})
	}
}

// TestCollectErrorMessageLiterals documents this exported function.
func TestCollectErrorMessageLiterals(t *testing.T) {
	src := `package p
import (
  "errors"
  "fmt"
)
func f() error {
  _ = errors.New("bad start")
  _ = fmt.Errorf("still bad")
  return nil
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, errMessageTestFileName, src, parser.ParseComments)
	if err != nil {
		t.Fatalf(errMessageParseFailHint, err)
	}

	items := collectErrorMessageLiterals(file)
	if len(items) != 2 {
		t.Fatalf("expected 2 literals, got %d", len(items))
	}
}
