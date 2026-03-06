package rules

import (
	"go/ast"

	"goulinette/internal/diag"
)

type err05Rule struct{}

func NewERR05() Rule {
	return err05Rule{}
}

func (err05Rule) ID() string {
	return "ERR-05"
}

func (err05Rule) Chapter() int {
	return 6
}

func (err05Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
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
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "ERR-05",
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
