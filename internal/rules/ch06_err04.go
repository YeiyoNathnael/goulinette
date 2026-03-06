package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"goulinette/internal/diag"
)

type err04Rule struct{}

func NewERR04() Rule {
	return err04Rule{}
}

func (err04Rule) ID() string {
	return "ERR-04"
}

func (err04Rule) Chapter() int {
	return 6
}

func (err04Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			ast.Inspect(syntaxFile, func(n ast.Node) bool {
				bin, ok := n.(*ast.BinaryExpr)
				if !ok {
					return true
				}
				if bin.Op != token.EQL && bin.Op != token.NEQ {
					return true
				}
				if isNilIdent(bin.X) || isNilIdent(bin.Y) {
					return true
				}

				tx := pkg.TypesInfo.TypeOf(bin.X)
				ty := pkg.TypesInfo.TypeOf(bin.Y)
				if !isErrorType(tx) || !isErrorType(ty) {
					return true
				}

				if !looksLikeSentinelError(bin.X) && !looksLikeSentinelError(bin.Y) {
					return true
				}

				pos := pkg.Fset.Position(bin.OpPos)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "ERR-04",
					Severity: diag.SeverityError,
					Message:  "specific error values must be checked with errors.Is, not ==/!=",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "replace comparison with errors.Is(err, target)",
				})
				return true
			})
		}
	}

	return diagnostics, nil
}

func isNilIdent(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "nil"
}

func looksLikeSentinelError(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return strings.HasPrefix(e.Name, "Err") || strings.HasPrefix(e.Name, "err")
	case *ast.SelectorExpr:
		return strings.HasPrefix(e.Sel.Name, "Err") || strings.HasPrefix(e.Sel.Name, "err")
	default:
		return false
	}
}
