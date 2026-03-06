package rules

import (
	"go/ast"

	"goulinette/internal/diag"
)

type fun01Rule struct{}

func NewFUN01() Rule {
	return fun01Rule{}
}

func (fun01Rule) ID() string {
	return "FUN-01"
}

func (fun01Rule) Chapter() int {
	return 5
}

func (fun01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
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
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "FUN-01",
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
