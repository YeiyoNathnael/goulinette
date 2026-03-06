package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type err04Rule struct{}

const err04Chapter = 6

// NewERR04 returns the ERR04 rule implementation.
func NewERR04() Rule {
	return err04Rule{}
}

// ID returns the rule identifier.
func (err04Rule) ID() string {
	return ruleERR04
}

// Chapter returns the chapter number for this rule.
func (err04Rule) Chapter() int {
	return err04Chapter
}

// Run executes this rule against the provided context.
func (err04Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleERR04,
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
