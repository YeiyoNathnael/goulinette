package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type ctl01Rule struct{}

func NewCTL01() Rule {
	return ctl01Rule{}
}

func (ctl01Rule) ID() string {
	return "CTL-01"
}

func (ctl01Rule) Chapter() int {
	return 4
}

func (ctl01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			ifStmt, ok := n.(*ast.IfStmt)
			if !ok {
				return true
			}

			branches, subject, ok := extractComparableIfChain(ifStmt)
			if !ok {
				return true
			}

			if branches >= 3 {
				pos := pf.FSet.Position(ifStmt.If)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "CTL-01",
					Severity: diag.SeverityWarning,
					Message:  "if/else-if chain with shared subject should use a switch",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "replace this chain with switch on " + subject,
				})
			}

			return true
		})
	}

	return diagnostics, nil
}

func extractComparableIfChain(root *ast.IfStmt) (int, string, bool) {
	count := 0
	common := ""
	current := root

	for current != nil {
		x, ok := comparedSubject(current.Cond)
		if !ok {
			return 0, "", false
		}
		if common == "" {
			common = x
		}
		if x != common {
			return 0, "", false
		}

		count++
		next, ok := current.Else.(*ast.IfStmt)
		if !ok {
			break
		}
		current = next
	}

	if count == 0 || common == "" {
		return 0, "", false
	}

	return count, common, true
}

func comparedSubject(expr ast.Expr) (string, bool) {
	b, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return "", false
	}
	if b.Op != token.EQL && b.Op != token.NEQ {
		return "", false
	}

	id, ok := b.X.(*ast.Ident)
	if !ok || id.Name == "" {
		return "", false
	}

	return id.Name, true
}
