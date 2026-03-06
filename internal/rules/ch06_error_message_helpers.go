package rules

import (
	"go/ast"
	"go/token"
	"strconv"
)

type errorMessageLiteral struct {
	call *ast.CallExpr
	text string
}

func collectErrorMessageLiterals(file *ast.File) []errorMessageLiteral {
	out := make([]errorMessageLiteral, 0)
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if !isErrorsNewCall(call) && !isFmtErrorfCall(call) {
			return true
		}
		if len(call.Args) == 0 {
			return true
		}

		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		msg, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}

		out = append(out, errorMessageLiteral{call: call, text: msg})
		return true
	})
	return out
}

func isErrorsNewCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return x.Name == "errors" && sel.Sel.Name == "New"
}

func isFmtErrorfCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return x.Name == "fmt" && sel.Sel.Name == "Errorf"
}
