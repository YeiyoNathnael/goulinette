package rules

import (
	"go/ast"
	"go/token"
	"strconv"

	"goulinette/internal/diag"
)

type var01Rule struct{}

func NewVAR01() Rule {
	return var01Rule{}
}

func (var01Rule) ID() string {
	return "VAR-01"
}

func (var01Rule) Chapter() int {
	return 3
}

func (var01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok || assign.Tok != token.DEFINE {
				return true
			}

			for i, rhs := range assign.Rhs {
				if i >= len(assign.Lhs) || !isZeroLiteralExpr(rhs) {
					continue
				}

				lhsIdent, ok := assign.Lhs[i].(*ast.Ident)
				if !ok || lhsIdent.Name == "_" {
					continue
				}

				pos := pf.FSet.Position(lhsIdent.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "VAR-01",
					Severity: diag.SeverityError,
					Message:  "use var declaration for zero-value initialization instead of := with zero literal",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "replace with: var " + lhsIdent.Name + " <type>",
				})
			}
			return true
		})
	}

	return diagnostics, nil
}

func isZeroLiteralExpr(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			value, err := strconv.Unquote(e.Value)
			return err == nil && value == ""
		}
		if e.Kind == token.INT {
			value, err := strconv.ParseInt(e.Value, 0, 64)
			return err == nil && value == 0
		}
		if e.Kind == token.FLOAT {
			value, err := strconv.ParseFloat(e.Value, 64)
			return err == nil && value == 0
		}
	case *ast.Ident:
		return e.Name == "false"
	}

	return false
}
