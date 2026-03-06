package rules

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
)

type callContext struct {
	call      *ast.CallExpr
	ancestors []ast.Node
}

const (
	ancestorStackCap = 32
	errorKeyword     = "error"
)

func collectCalls(file ast.Node, name string) []callContext {
	out := make([]callContext, 0)
	stack := make([]ast.Node, 0, ancestorStackCap)

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			return false
		}

		if call, ok := n.(*ast.CallExpr); ok && isBuiltinCall(call, name) {
			anc := make([]ast.Node, len(stack))
			copy(anc, stack)
			out = append(out, callContext{call: call, ancestors: anc})
		}

		stack = append(stack, n)
		return true
	})

	return out
}

func isBuiltinCall(call *ast.CallExpr, name string) bool {
	id, ok := call.Fun.(*ast.Ident)
	return ok && id.Name == name
}

func isRecoverInDeferredAnonymousFunc(ctx callContext) bool {
	if ctx.call == nil {
		return false
	}

	var target *ast.FuncLit
	for i := len(ctx.ancestors) - 1; i >= 0; i-- {
		if fl, ok := ctx.ancestors[i].(*ast.FuncLit); ok {
			target = fl
			break
		}
	}
	if target == nil {
		return false
	}

	for _, anc := range ctx.ancestors {
		d, ok := anc.(*ast.DeferStmt)
		if !ok || d.Call == nil {
			continue
		}
		if fl, ok := d.Call.Fun.(*ast.FuncLit); ok && fl == target {
			return true
		}
	}

	return false
}

func enclosingFuncName(ancestors []ast.Node) string {
	for i := len(ancestors) - 1; i >= 0; i-- {
		if fd, ok := ancestors[i].(*ast.FuncDecl); ok && fd.Name != nil {
			return fd.Name.Name
		}
	}
	return ""
}

func isOperationalPanicArg(expr ast.Expr, info *types.Info) bool {
	if expr == nil {
		return false
	}

	if call, ok := expr.(*ast.CallExpr); ok {
		if isErrorsNewCall(call) || isFmtErrorfCall(call) {
			return true
		}
	}

	if info != nil && isErrorType(info.TypeOf(expr)) {
		return true
	}

	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		v, err := strconv.Unquote(lit.Value)
		if err != nil {
			return false
		}
		lower := strings.ToLower(v)
		for _, kw := range []string{"failed", errorKeyword, "timeout", "not found", "invalid", "unable", "cannot"} {
			if strings.Contains(lower, kw) {
				return true
			}
		}
	}

	return false
}
