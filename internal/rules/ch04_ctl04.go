package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type ctl04Rule struct{}

func NewCTL04() Rule {
	return ctl04Rule{}
}

func (ctl04Rule) ID() string {
	return "CTL-04"
}

func (ctl04Rule) Chapter() int {
	return 4
}

func (ctl04Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSwitchStmt)
			if !ok {
				return true
			}

			if hasTypeSwitchDefault(ts) {
				return true
			}

			pos := pf.FSet.Position(ts.Switch)
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "CTL-04",
				Severity: diag.SeverityError,
				Message:  "type switch must include a default case",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "add a default branch to handle unexpected concrete types",
			})

			return true
		})
	}

	return diagnostics, nil
}

func hasTypeSwitchDefault(ts *ast.TypeSwitchStmt) bool {
	if ts == nil || ts.Body == nil {
		return false
	}
	for _, stmt := range ts.Body.List {
		cc, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		if cc.List == nil {
			return true
		}
	}
	return false
}
