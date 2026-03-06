package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type var03Rule struct{}

const var03Chapter = 3

// NewVAR03 returns the VAR03 rule implementation.
func NewVAR03() Rule {
	return var03Rule{}
}

// ID returns the rule identifier.
func (var03Rule) ID() string {
	return ruleVAR03
}

// Chapter returns the chapter number for this rule.
func (var03Rule) Chapter() int {
	return var03Chapter
}

// Run executes this rule against the provided context.
func (var03Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			declStmt, ok := n.(*ast.DeclStmt)
			if !ok {
				return true
			}

			gen, ok := declStmt.Decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				return true
			}

			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				if vs.Type != nil || len(vs.Values) == 0 {
					continue
				}

				if allZeroLiteralValues(vs.Values) {
					continue
				}

				for _, name := range vs.Names {
					if name == nil || name.Name == "_" {
						continue
					}
					pos := pf.FSet.Position(name.Pos())
					diagnostics = append(diagnostics, diag.Finding{
						RuleID:   ruleVAR03,
						Severity: diag.SeverityWarning,
						Message:  "prefer := for local declarations when var type is inferred",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "replace var declaration with short declaration :=",
					})
				}
			}

			return true
		})
	}

	return diagnostics, nil
}

func allZeroLiteralValues(values []ast.Expr) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if !isZeroLiteralExpr(value) {
			return false
		}
	}
	return true
}
