package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type err05Rule struct{}

const err05Chapter = 6

// NewERR05 returns the ERR05 rule implementation.
func NewERR05() Rule {
	return err05Rule{}
}

// ID returns the rule identifier.
func (err05Rule) ID() string {
	return ruleERR05
}

// Chapter returns the chapter number for this rule.
func (err05Rule) Chapter() int {
	return err05Chapter
}

// Run executes this rule against the provided context.
func (err05Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			ast.Inspect(syntaxFile, func(n ast.Node) bool {
				ta, ok := n.(*ast.TypeAssertExpr)
				if !ok {
					return true
				}

				typ := pkg.TypesInfo.TypeOf(ta.X)
				if !isErrorType(typ) {
					return true
				}

				pos := pkg.Fset.Position(ta.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleERR05,
					Severity: diag.SeverityError,
					Message:  "specific error types must be checked with errors.As, not type assertions",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "replace with errors.As(err, &target)",
				})

				return true
			})
		}
	}

	return diagnostics, nil
}
