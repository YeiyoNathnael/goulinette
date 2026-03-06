package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"goulinette/internal/diag"
)

type slc01Rule struct{}

func NewSLC01() Rule {
	return slc01Rule{}
}

func (slc01Rule) ID() string {
	return "SLC-01"
}

func (slc01Rule) Chapter() int {
	return 10
}

func (slc01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}

			if !isEmptySliceLiteral(lit) {
				return true
			}

			line := pf.FSet.Position(lit.Lbrace).Line
			if hasSliceLiteralJustification(pf.File, pf.FSet, line) {
				return true
			}

			pos := pf.FSet.Position(lit.Lbrace)
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "SLC-01",
				Severity: diag.SeverityWarning,
				Message:  "prefer nil slices over empty slice literals when initializing empty collections",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "use nil slice initialization (e.g., var s []T or s = nil) unless non-nil empty slice is explicitly required",
			})
			return true
		})
	}

	return diagnostics, nil
}

func isEmptySliceLiteral(lit *ast.CompositeLit) bool {
	if lit == nil || len(lit.Elts) != 0 {
		return false
	}

	arrayType, ok := lit.Type.(*ast.ArrayType)
	if !ok {
		return false
	}

	return arrayType.Len == nil
}

func hasSliceLiteralJustification(file *ast.File, fset *token.FileSet, line int) bool {
	for _, cg := range file.Comments {
		start := fset.Position(cg.Pos()).Line
		end := fset.Position(cg.End()).Line
		if start != line && end != line && end != line-1 {
			continue
		}

		text := strings.ToLower(cg.Text())
		if strings.Contains(text, "non-nil") || strings.Contains(text, "json") {
			return true
		}
	}

	return false
}
