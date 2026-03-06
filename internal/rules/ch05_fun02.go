package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type fun02Rule struct{}

func NewFUN02() Rule {
	return fun02Rule{}
}

func (fun02Rule) ID() string {
	return "FUN-02"
}

func (fun02Rule) Chapter() int {
	return 5
}

func (fun02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Type == nil || fn.Type.Results == nil {
				continue
			}

			if !errorMustBeLast(fn.Type.Results) {
				pos := pf.FSet.Position(fn.Name.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "FUN-02",
					Severity: diag.SeverityError,
					Message:  "error must be the last return value",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "reorder return values so error is rightmost",
				})
			}
		}
	}

	return diagnostics, nil
}

func errorMustBeLast(results *ast.FieldList) bool {
	returnTypes := flattenResultTypes(results)
	if len(returnTypes) == 0 {
		return true
	}

	for idx, typ := range returnTypes {
		if !isErrorTypeExpr(typ) {
			continue
		}
		if idx != len(returnTypes)-1 {
			return false
		}
	}

	return true
}

func flattenResultTypes(results *ast.FieldList) []ast.Expr {
	out := make([]ast.Expr, 0)
	if results == nil {
		return out
	}

	for _, field := range results.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			out = append(out, field.Type)
		}
	}

	return out
}

func isErrorTypeExpr(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "error"
}
