package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type ctl03Rule struct{}

const ctl03Chapter = 4

// NewCTL03 returns the CTL03 rule implementation.
func NewCTL03() Rule {
	return ctl03Rule{}
}

// ID returns the rule identifier.
func (ctl03Rule) ID() string {
	return ruleCTL03
}

// Chapter returns the chapter number for this rule.
func (ctl03Rule) Chapter() int {
	return ctl03Chapter
}

// Run executes this rule against the provided context.
func (ctl03Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			branch, ok := n.(*ast.BranchStmt)
			if !ok || branch.Tok != token.GOTO {
				return true
			}

			pos := pf.FSet.Position(branch.Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleCTL03,
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
