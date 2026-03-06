package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type typ05Rule struct{}

func NewTYP05() Rule {
	return typ05Rule{}
}

func (typ05Rule) ID() string {
	return "TYP-05"
}

func (typ05Rule) Chapter() int {
	return 7
}

func (typ05Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, tc := range collectSingleValueAssertions(pf.File) {
			if tc.inTypeSwitchCase {
				continue
			}

			pos := pf.FSet.Position(tc.assertion.Lparen)
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "TYP-05",
				Severity: diag.SeverityError,
				Message:  "type assertions must use comma-ok form",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "use v, ok := x.(T) and check ok",
			})
		}
	}

	return diagnostics, nil
}

type assertionContext struct {
	assertion        *ast.TypeAssertExpr
	inTypeSwitchCase bool
}

func collectSingleValueAssertions(file *ast.File) []assertionContext {
	out := make([]assertionContext, 0)
	stack := make([]ast.Node, 0, 32)

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			return false
		}

		if ta, ok := n.(*ast.TypeAssertExpr); ok && ta.Type != nil {
			if isCommaOKAssertion(ta, stack) {
				stack = append(stack, n)
				return true
			}

			out = append(out, assertionContext{
				assertion:        ta,
				inTypeSwitchCase: isInTypeSwitchCase(stack),
			})
		}

		stack = append(stack, n)
		return true
	})

	return out
}

func isCommaOKAssertion(ta *ast.TypeAssertExpr, ancestors []ast.Node) bool {
	for i := len(ancestors) - 1; i >= 0; i-- {
		as, ok := ancestors[i].(*ast.AssignStmt)
		if !ok {
			continue
		}
		for _, rhs := range as.Rhs {
			if rhs == ta && len(as.Lhs) >= 2 {
				return true
			}
		}
		return false
	}
	return false
}

func isInTypeSwitchCase(ancestors []ast.Node) bool {
	hasCase := false
	for i := len(ancestors) - 1; i >= 0; i-- {
		switch ancestors[i].(type) {
		case *ast.CaseClause:
			hasCase = true
		case *ast.TypeSwitchStmt:
			return hasCase
		}
	}
	return false
}
