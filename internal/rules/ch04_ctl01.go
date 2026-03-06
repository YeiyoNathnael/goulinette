package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type ctl01Rule struct{}

const (
	ctl01Chapter       = 4
	ctl01MinChainCount = 3
)

// NewCTL01 returns the CTL01 rule implementation.
func NewCTL01() Rule {
	return ctl01Rule{}
}

// ID returns the rule identifier.
func (ctl01Rule) ID() string {
	return ruleCTL01
}

// Chapter returns the chapter number for this rule.
func (ctl01Rule) Chapter() int {
	return ctl01Chapter
}

// Run executes this rule against the provided context.
func (ctl01Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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

			if branches >= ctl01MinChainCount {
				pos := pf.FSet.Position(ifStmt.If)
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleCTL01,
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
	var count int
	var common string
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
