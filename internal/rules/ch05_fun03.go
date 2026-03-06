package rules

import (
	"go/ast"

	"goulinette/internal/diag"
)

type fun03Rule struct{}

func NewFUN03() Rule {
	return fun03Rule{}
}

func (fun03Rule) ID() string {
	return "FUN-03"
}

func (fun03Rule) Chapter() int {
	return 5
}

func (fun03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Type == nil {
				continue
			}

			if shouldWarnConcreteParams(fn.Type) {
				pos := pf.FSet.Position(fn.Name.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "FUN-03",
					Severity: diag.SeverityWarning,
					Message:  "function accepts concrete types only; consider interface parameters for decoupling",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "accept interfaces where behavior contracts are sufficient",
				})
			}

			if shouldWarnInterfaceReturn(fn.Type) {
				pos := pf.FSet.Position(fn.Name.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "FUN-03",
					Severity: diag.SeverityWarning,
					Message:  "function returns interface-like type; consider returning concrete type",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "return concrete structs unless polymorphism is required",
				})
			}
		}
	}

	return diagnostics, nil
}

func shouldWarnConcreteParams(ft *ast.FuncType) bool {
	if ft == nil || ft.Params == nil || len(ft.Params.List) == 0 {
		return false
	}

	hasConcrete := false
	hasInterface := false
	for _, p := range ft.Params.List {
		if isInterfaceTypeExpr(p.Type) {
			hasInterface = true
			continue
		}
		if isConcreteNamedTypeExpr(p.Type) {
			hasConcrete = true
		}
	}

	return hasConcrete && !hasInterface
}

func shouldWarnInterfaceReturn(ft *ast.FuncType) bool {
	if ft == nil || ft.Results == nil {
		return false
	}

	for _, field := range ft.Results.List {
		if isErrorTypeExpr(field.Type) {
			continue
		}
		if isInterfaceTypeExpr(field.Type) {
			return true
		}
	}

	return false
}

func isConcreteNamedTypeExpr(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name != "any" && t.Name != "error"
	case *ast.StarExpr:
		_, ok := t.X.(*ast.Ident)
		return ok
	case *ast.SelectorExpr:
		return true
	default:
		return false
	}
}

func isInterfaceTypeExpr(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.InterfaceType:
		return true
	case *ast.Ident:
		return t.Name == "any"
	default:
		return false
	}
}
