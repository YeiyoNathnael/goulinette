package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type typ04Rule struct{}

const (
	typ04Chapter      = 7
	typ04CommaOKArity = 2
	typ04AstStackCap  = 32
)

// NewTYP04 returns the TYP04 rule implementation.
func NewTYP04() Rule {
	return typ04Rule{}
}

// ID returns the rule identifier.
func (typ04Rule) ID() string {
	return ruleTYP04
}

// Chapter returns the chapter number for this rule.
func (typ04Rule) Chapter() int {
	return typ04Chapter
}

// Run executes this rule against the provided context.
func (typ04Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, recv := range collectSingleValueReceives(pf.File) {
			pos := pf.FSet.Position(recv.OpPos)
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleTYP04,
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
	stack := make([]ast.Node, 0, typ04AstStackCap)

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			return false
		}

		if un, ok := n.(*ast.UnaryExpr); ok && un.Op == token.ARROW {
			if !isTwoValueReceive(un, stack) && !isSelectExprReceive(stack) {
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
			if rhs == un && len(as.Lhs) >= typ04CommaOKArity {
				return true
			}
		}
		return false
	}
	return false
}

// isSelectExprReceive reports whether the innermost two ancestors show that
// this receive is a pure expression inside a select CommClause (i.e. the
// canonical `case <-ctx.Done():` signal pattern). These receives discard the
// value intentionally and do not benefit from comma-ok form.
func isSelectExprReceive(ancestors []ast.Node) bool {
	n := len(ancestors)
	if n < 2 {
		return false
	}
	_, parentIsExpr := ancestors[n-1].(*ast.ExprStmt)
	_, grandIsComm := ancestors[n-2].(*ast.CommClause)
	return parentIsExpr && grandIsComm
}
