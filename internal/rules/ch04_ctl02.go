package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type ctl02Rule struct{}

func NewCTL02() Rule {
	return ctl02Rule{}
}

func (ctl02Rule) ID() string {
	return "CTL-02"
}

func (ctl02Rule) Chapter() int {
	return 4
}

func (ctl02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			branch, ok := n.(*ast.BranchStmt)
			if !ok || branch.Tok != token.FALLTHROUGH {
				return true
			}

			pos := pf.FSet.Position(branch.Pos())
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "CTL-02",
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
