package rules

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestIsZeroLiteralExpr(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want bool
	}{
		{name: "int zero", expr: &ast.BasicLit{Kind: token.INT, Value: "0"}, want: true},
		{name: "float zero", expr: &ast.BasicLit{Kind: token.FLOAT, Value: "0.0"}, want: true},
		{name: "empty string", expr: &ast.BasicLit{Kind: token.STRING, Value: "\"\""}, want: true},
		{name: "false bool", expr: &ast.Ident{Name: "false"}, want: true},
		{name: "non-zero int", expr: &ast.BasicLit{Kind: token.INT, Value: "1"}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isZeroLiteralExpr(tc.expr)
			if got != tc.want {
				t.Fatalf("isZeroLiteralExpr() = %v, want %v", got, tc.want)
			}
		})
	}
}
