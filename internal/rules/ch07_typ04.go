package rules

import (
	"go/ast"
	"go/token"

	"goulinette/internal/diag"
)

type typ04Rule struct{}

func NewTYP04() Rule {
	return typ04Rule{}
}

func (typ04Rule) ID() string {
	return "TYP-04"
}

func (typ04Rule) Chapter() int {
	return 7
}

func (typ04Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, recv := range collectSingleValueReceives(pf.File) {
			pos := pf.FSet.Position(recv.OpPos)
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "TYP-04",
				Severity: diag.SeverityError,
				Message:  "channel reads should use comma-ok form when channels may be closed",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "use v, ok := <-ch and check ok",
			})
		}
	}

	return diagnostics, nil
}

func collectSingleValueReceives(file *ast.File) []*ast.UnaryExpr {
	out := make([]*ast.UnaryExpr, 0)
	stack := make([]ast.Node, 0, 32)

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			return false
		}

		if un, ok := n.(*ast.UnaryExpr); ok && un.Op == token.ARROW {
			if !isTwoValueReceive(un, stack) {
				out = append(out, un)
			}
		}

		stack = append(stack, n)
		return true
	})

	return out
}

func isTwoValueReceive(un *ast.UnaryExpr, ancestors []ast.Node) bool {
	for i := len(ancestors) - 1; i >= 0; i-- {
		as, ok := ancestors[i].(*ast.AssignStmt)
		if !ok {
			continue
		}
		for _, rhs := range as.Rhs {
			if rhs == un && len(as.Lhs) >= 2 {
				return true
			}
		}
		return false
	}
	return false
}
