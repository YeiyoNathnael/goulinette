package rules

import (
	"go/ast"
	"go/token"

	"goulinette/internal/diag"
)

type ctl03Rule struct{}

func NewCTL03() Rule {
	return ctl03Rule{}
}

func (ctl03Rule) ID() string {
	return "CTL-03"
}

func (ctl03Rule) Chapter() int {
	return 4
}

func (ctl03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			branch, ok := n.(*ast.BranchStmt)
			if !ok || branch.Tok != token.GOTO {
				return true
			}

			pos := pf.FSet.Position(branch.Pos())
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "CTL-03",
				Severity: diag.SeverityError,
				Message:  "goto is forbidden",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "use structured control flow with for/break/continue/return",
			})
			return true
		})
	}

	return diagnostics, nil
}
