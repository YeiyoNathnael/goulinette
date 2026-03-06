package rules

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type var01Rule struct{}

const (
	var01Chapter            = 3
	var01ParseBitSize       = 64
	var01FalseLiteral       = "false"
	var01ZeroInt            = 0
	var01ZeroFloat          = 0
	var01EmptyStringLiteral = ""
)

// NewVAR01 returns the VAR01 rule implementation.
func NewVAR01() Rule {
	return var01Rule{}
}

// ID returns the rule identifier.
func (var01Rule) ID() string {
	return ruleVAR01
}

// Chapter returns the chapter number for this rule.
func (var01Rule) Chapter() int {
	return var01Chapter
}

// Run executes this rule against the provided context.
func (var01Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		forInitDecls := map[token.Pos]struct{}{}
		ast.Inspect(pf.File, func(n ast.Node) bool {
			fs, ok := n.(*ast.ForStmt)
			if !ok || fs.Init == nil {
				return true
			}
			assign, ok := fs.Init.(*ast.AssignStmt)
			if !ok || assign.Tok != token.DEFINE {
				return true
			}
			forInitDecls[assign.Pos()] = struct{}{}
			return true
		})

		ast.Inspect(pf.File, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok || assign.Tok != token.DEFINE {
				return true
			}
			if _, ok := forInitDecls[assign.Pos()]; ok {
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
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleVAR01,
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
			return err == nil && value == var01EmptyStringLiteral
		}
		if e.Kind == token.INT {
			value, err := strconv.ParseInt(e.Value, var01ZeroInt, var01ParseBitSize)
			return err == nil && value == var01ZeroInt
		}
		if e.Kind == token.FLOAT {
			value, err := strconv.ParseFloat(e.Value, var01ParseBitSize)
			return err == nil && value == var01ZeroFloat
		}
	case *ast.Ident:
		return e.Name == var01FalseLiteral
	default:
		return false
	}

	return false
}
