package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type fun01Rule struct{}

const fun01Chapter = 5

// NewFUN01 returns the FUN01 rule implementation.
func NewFUN01() Rule {
	return fun01Rule{}
}

// ID returns the rule identifier.
func (fun01Rule) ID() string {
	return ruleFUN01
}

// Chapter returns the chapter number for this rule.
func (fun01Rule) Chapter() int {
	return fun01Chapter
}

// Run executes this rule against the provided context.
func (fun01Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !hasNamedReturns(fn.Type) {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				ret, ok := n.(*ast.ReturnStmt)
				if !ok || len(ret.Results) > 0 {
					return true
				}

				pos := pf.FSet.Position(ret.Return)
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleFUN01,
					Severity: diag.SeverityError,
					Message:  "naked return is forbidden",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "return explicit values",
				})
				return true
			})
		}
	}

	return diagnostics, nil
}

func hasNamedReturns(ft *ast.FuncType) bool {
	if ft == nil || ft.Results == nil {
		return false
	}

	for _, field := range ft.Results.List {
		if len(field.Names) > 0 {
			return true
		}
	}
	return false
}
