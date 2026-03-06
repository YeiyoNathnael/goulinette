package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type ctl02Rule struct{}

const ctl02Chapter = 4

// NewCTL02 returns the CTL02 rule implementation.
func NewCTL02() Rule {
	return ctl02Rule{}
}

// ID returns the rule identifier.
func (ctl02Rule) ID() string {
	return ruleCTL02
}

// Chapter returns the chapter number for this rule.
func (ctl02Rule) Chapter() int {
	return ctl02Chapter
}

// Run executes this rule against the provided context.
func (ctl02Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			branch, ok := n.(*ast.BranchStmt)
			if !ok || branch.Tok != token.FALLTHROUGH {
				return true
			}

			pos := pf.FSet.Position(branch.Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleCTL02,
				Severity: diag.SeverityError,
				Message:  "fallthrough is forbidden",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "combine cases or extract shared logic",
			})
			return true
		})
	}

	return diagnostics, nil
}
