package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type fun02Rule struct{}

const (
	fun02Chapter = 5
	goErrorType  = "error"
)

// NewFUN02 returns the FUN02 rule implementation.
func NewFUN02() Rule {
	return fun02Rule{}
}

// ID returns the rule identifier.
func (fun02Rule) ID() string {
	return ruleFUN02
}

// Chapter returns the chapter number for this rule.
func (fun02Rule) Chapter() int {
	return fun02Chapter
}

// Run executes this rule against the provided context.
func (fun02Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Type == nil || fn.Type.Results == nil {
				continue
			}

			if !errorMustBeLast(fn.Type.Results) {
				pos := pf.FSet.Position(fn.Name.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleFUN02,
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
	return ok && id.Name == goErrorType
}
