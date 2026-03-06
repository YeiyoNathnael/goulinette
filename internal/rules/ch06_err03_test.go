package rules

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"
)

// TestIsNilableType documents this exported function.
func TestIsNilableType(t *testing.T) {
	tests := []struct {
		name string
		typ  types.Type
		want bool
	}{
		{name: "pointer", typ: types.NewPointer(types.Typ[types.Int]), want: true},
		{name: "slice", typ: types.NewSlice(types.Typ[types.Int]), want: true},
		{name: "map", typ: types.NewMap(types.Typ[types.String], types.Typ[types.Int]), want: true},
		{name: "struct", typ: types.NewStruct(nil, nil), want: false},
		{name: "int", typ: types.Typ[types.Int], want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			if got := isNilableType(tc.typ); got != tc.want {
				t.Fatalf("isNilableType() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestClassifyReturnedErrorExpr documents this exported function.
func TestClassifyReturnedErrorExpr(t *testing.T) {
	info := &types.Info{}

	if got := classifyReturnedErrorExpr(info, &ast.Ident{Name: "nil"}); got != returnedErrorNil {
		t.Fatalf("nil should classify as returnedErrorNil")
	}

	errorsNew := &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "errors"}, Sel: &ast.Ident{Name: "New"}},
		Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: "\"x\""}},
	}
	if got := classifyReturnedErrorExpr(info, errorsNew); got != returnedErrorNonNil {
		t.Fatalf("errors.New should classify as returnedErrorNonNil")
	}

	fmtErr := &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
		Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: "\"x\""}},
	}
	if got := classifyReturnedErrorExpr(info, fmtErr); got != returnedErrorNonNil {
		t.Fatalf("fmt.Errorf should classify as returnedErrorNonNil")
	}
}
